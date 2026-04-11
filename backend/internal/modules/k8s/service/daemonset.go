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

type DaemonSetListVO struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	DesiredNumber   int32             `json:"desiredNumber"`
	CurrentNumber   int32             `json:"currentNumber"`
	ReadyNumber     int32             `json:"readyNumber"`
	Labels          map[string]string `json:"labels"`
	Containers      []ContainerInfo   `json:"containers"`
	ResourceSummary ResourceSummary   `json:"resourceSummary"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"createdAt"`
}

type DaemonSetListResponse struct {
	Total int64             `json:"total"`
	Items []DaemonSetListVO `json:"items"`
}

func (s *K8sService) ListDaemonSets(clusterName string, namespace string, page, pageSize int, keyword string) (*DaemonSetListResponse, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	list, err := client.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.clientFactory.RemoveClient(cluster.Name)
		return nil, err
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item appsv1.DaemonSet) []string {
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

	result := make([]DaemonSetListVO, 0, len(paged))
	for _, item := range paged {
		containers := buildContainerInfos(item.Spec.Template.Spec.Containers)
		resSummary := aggregateResources(item.Spec.Template.Spec.Containers)

		result = append(result, DaemonSetListVO{
			Name:            item.Name,
			Namespace:       item.Namespace,
			DesiredNumber:   item.Status.DesiredNumberScheduled,
			CurrentNumber:   item.Status.CurrentNumberScheduled,
			ReadyNumber:     item.Status.NumberReady,
			Labels:          item.Labels,
			Containers:      containers,
			ResourceSummary: resSummary,
			Status:          computeDaemonSetStatus(&item),
			CreatedAt:       item.CreationTimestamp.Time,
		})
	}

	return &DaemonSetListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) GetDaemonSetDetail(clusterName string, namespace, name string) (*DaemonSetListVO, error) {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	item, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := buildContainerInfos(item.Spec.Template.Spec.Containers)
	resSummary := aggregateResources(item.Spec.Template.Spec.Containers)

	return &DaemonSetListVO{
		Name:            item.Name,
		Namespace:       item.Namespace,
		DesiredNumber:   item.Status.DesiredNumberScheduled,
		CurrentNumber:   item.Status.CurrentNumberScheduled,
		ReadyNumber:     item.Status.NumberReady,
		Labels:          item.Labels,
		Containers:      containers,
		ResourceSummary: resSummary,
		Status:          computeDaemonSetStatus(item),
		CreatedAt:       item.CreationTimestamp.Time,
	}, nil
}

func (s *K8sService) GetDaemonSetYAML(clusterName string, namespace, name string) (string, error) {
	obj, err := s.GetDaemonSetObject(clusterName, namespace, name)
	if err != nil {
		return "", err
	}
	b, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *K8sService) GetDaemonSetObject(clusterName string, namespace, name string) (*appsv1.DaemonSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}
	return client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) UpdateDaemonSetByYAML(clusterName string, namespace, name, rawYAML string) (*appsv1.DaemonSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	current, err := s.GetDaemonSetObject(clusterName, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current daemonset: %w", err)
	}

	var desired appsv1.DaemonSet
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	desired.Namespace = namespace
	desired.Name = name
	desired.ResourceVersion = current.ResourceVersion
	desired.Status = appsv1.DaemonSetStatus{}
	desired.ManagedFields = nil

	return s.UpdateDaemonSet(clusterName, namespace, &desired)
}

func (s *K8sService) RestartDaemonSet(clusterName string, namespace, name string) (*appsv1.DaemonSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	patch := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":%q}}}}}`, now)
	return client.AppsV1().DaemonSets(namespace).Patch(context.Background(), name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
}

func (s *K8sService) DeleteDaemonSet(clusterName string, namespace, name string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return err
	}

	return client.AppsV1().DaemonSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (s *K8sService) CreateDaemonSet(clusterName string, namespace string, ds *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.AppsV1().DaemonSets(namespace).Create(context.Background(), ds, metav1.CreateOptions{})
}

func (s *K8sService) UpdateDaemonSet(clusterName string, namespace string, ds *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.AppsV1().DaemonSets(namespace).Update(context.Background(), ds, metav1.UpdateOptions{})
}

func computeDaemonSetStatus(ds *appsv1.DaemonSet) string {
	if ds.Status.NumberReady == 0 {
		return "Unavailable"
	}
	if ds.Status.NumberReady == ds.Status.DesiredNumberScheduled {
		return "Available"
	}
	return "Progressing"
}
