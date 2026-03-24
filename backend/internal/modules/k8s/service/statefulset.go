package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

type StatefulSetListVO struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"readyReplicas"`
	Labels            map[string]string `json:"labels"`
	Containers        []ContainerInfo   `json:"containers"`
	ResourceSummary   ResourceSummary   `json:"resourceSummary"`
	Status            string            `json:"status"`
	CreatedAt         time.Time         `json:"createdAt"`
}

type StatefulSetListResponse struct {
	Total int64               `json:"total"`
	Items []StatefulSetListVO `json:"items"`
}

func (s *K8sService) ListStatefulSets(clusterId uint, namespace string, page, pageSize int, keyword string) (*StatefulSetListResponse, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	list, err := client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.clientFactory.RemoveClient(clusterId)
		return nil, err
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item appsv1.StatefulSet) []string {
		images := make([]string, 0, len(item.Spec.Template.Spec.Containers))
		for _, container := range item.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
		return []string{
			item.Name,
			item.Namespace,
			strings.Join(images, ","),
			flattenLabels(item.Labels),
		}
	})

	total := int64(len(filtered))
	page, pageSize = normalizePage(page, pageSize)
	start, end := paginateRange(len(filtered), page, pageSize)
	paged := filtered[start:end]

	result := make([]StatefulSetListVO, 0, len(paged))
	for _, item := range paged {
		containers := buildContainerInfos(item.Spec.Template.Spec.Containers)
		resSummary := aggregateResources(item.Spec.Template.Spec.Containers)

		var replicas int32
		if item.Spec.Replicas != nil {
			replicas = *item.Spec.Replicas
		}

		result = append(result, StatefulSetListVO{
			Name:            item.Name,
			Namespace:       item.Namespace,
			Replicas:        replicas,
			ReadyReplicas:   item.Status.ReadyReplicas,
			Labels:          item.Labels,
			Containers:      containers,
			ResourceSummary: resSummary,
			Status:          computeStatefulSetStatus(&item),
			CreatedAt:       item.CreationTimestamp.Time,
		})
	}

	return &StatefulSetListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) GetStatefulSetDetail(clusterId uint, namespace, name string) (*StatefulSetListVO, error) {
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	item, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := buildContainerInfos(item.Spec.Template.Spec.Containers)
	resSummary := aggregateResources(item.Spec.Template.Spec.Containers)

	var replicas int32
	if item.Spec.Replicas != nil {
		replicas = *item.Spec.Replicas
	}

	return &StatefulSetListVO{
		Name:            item.Name,
		Namespace:       item.Namespace,
		Replicas:        replicas,
		ReadyReplicas:   item.Status.ReadyReplicas,
		Labels:          item.Labels,
		Containers:      containers,
		ResourceSummary: resSummary,
		Status:          computeStatefulSetStatus(item),
		CreatedAt:       item.CreationTimestamp.Time,
	}, nil
}

func (s *K8sService) GetStatefulSetYAML(clusterId uint, namespace, name string) (string, error) {
	obj, err := s.GetStatefulSetObject(clusterId, namespace, name)
	if err != nil {
		return "", err
	}
	b, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *K8sService) GetStatefulSetObject(clusterId uint, namespace, name string) (*appsv1.StatefulSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}
	return client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) UpdateStatefulSetByYAML(clusterId uint, namespace, name, rawYAML string) (*appsv1.StatefulSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	current, err := s.GetStatefulSetObject(clusterId, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current statefulset: %w", err)
	}

	var desired appsv1.StatefulSet
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	desired.Namespace = namespace
	desired.Name = name
	desired.ResourceVersion = current.ResourceVersion
	desired.Status = appsv1.StatefulSetStatus{}
	desired.ManagedFields = nil

	return s.UpdateStatefulSet(clusterId, namespace, &desired)
}

func (s *K8sService) RestartStatefulSet(clusterId uint, namespace, name string) (*appsv1.StatefulSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	patch := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":%q}}}}}`, now)
	return client.AppsV1().StatefulSets(namespace).Patch(context.Background(), name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
}

func (s *K8sService) ScaleStatefulSet(clusterId uint, namespace, name string, replicas int32) (*appsv1.StatefulSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	scale, err := client.AppsV1().StatefulSets(namespace).GetScale(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	scale.Spec.Replicas = replicas
	_, err = client.AppsV1().StatefulSets(namespace).UpdateScale(context.Background(), name, scale, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) DeleteStatefulSet(clusterId uint, namespace, name string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return err
	}

	return client.AppsV1().StatefulSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (s *K8sService) CreateStatefulSet(clusterId uint, namespace string, sts *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.AppsV1().StatefulSets(namespace).Create(context.Background(), sts, metav1.CreateOptions{})
}

func (s *K8sService) UpdateStatefulSet(clusterId uint, namespace string, sts *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.AppsV1().StatefulSets(namespace).Update(context.Background(), sts, metav1.UpdateOptions{})
}

func computeStatefulSetStatus(sts *appsv1.StatefulSet) string {
	if sts.Status.ReadyReplicas == 0 {
		return "Unavailable"
	}
	if sts.Spec.Replicas != nil && sts.Status.ReadyReplicas == *sts.Spec.Replicas {
		return "Available"
	}
	return "Progressing"
}
