package service

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodVO struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	IP        string            `json:"ip"`
	Node      string            `json:"node"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"createdAt"`
}

type PodListVO struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Status    string    `json:"status"`
	IP        string    `json:"ip"`
	Node      string    `json:"node"`
	CreatedAt time.Time `json:"createdAt"`
}

// PodListResponse Pod 列表分页响应
type PodListResponse struct {
	Total int64       `json:"total"`
	Items []PodListVO `json:"items"`
}

func (s *K8sService) ListPods(clusterId uint, namespace string, page, pageSize int, keyword string) (*PodListResponse, error) {
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

	list, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.clientFactory.RemoveClient(clusterId)
		return nil, err
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item corev1.Pod) []string {
		return []string{
			item.Name,
			item.Namespace,
			string(item.Status.Phase),
			item.Spec.NodeName,
			item.Status.PodIP,
			flattenLabels(item.Labels),
		}
	})

	total := int64(len(filtered))

	// 分页
	page, pageSize = normalizePage(page, pageSize)
	start, end := paginateRange(len(filtered), page, pageSize)
	paged := filtered[start:end]

	result := make([]PodListVO, 0, len(paged))
	for _, item := range paged {
		result = append(result, PodListVO{
			Name:      item.Name,
			Namespace: item.Namespace,
			Status:    string(item.Status.Phase),
			IP:        item.Status.PodIP,
			Node:      item.Spec.NodeName,
			CreatedAt: item.CreationTimestamp.Time,
		})
	}
	return &PodListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) GetPodDetail(clusterId uint, namespace, name string) (*PodVO, error) {
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	item, err := client.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &PodVO{
		Name:      item.Name,
		Namespace: item.Namespace,
		Status:    string(item.Status.Phase),
		IP:        item.Status.PodIP,
		Node:      item.Spec.NodeName,
		Labels:    item.Labels,
		CreatedAt: item.CreationTimestamp.Time,
	}, nil
}

func (s *K8sService) DeletePod(clusterId uint, namespace, name string) error {
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return err
	}

	return client.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (s *K8sService) CreatePod(clusterId uint, namespace string, pod *corev1.Pod) (*corev1.Pod, error) {
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})
}

func (s *K8sService) UpdatePod(clusterId uint, namespace string, pod *corev1.Pod) (*corev1.Pod, error) {
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.CoreV1().Pods(namespace).Update(context.Background(), pod, metav1.UpdateOptions{})
}

// ListPodsByOwner 根据控制器类型和名称获取 Pod 列表
func (s *K8sService) ListPodsByOwner(clusterId uint, namespace string, ownerType string, ownerName string) ([]PodListVO, error) {
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

	var selector string
	switch ownerType {
	case "Deployment":
		deploy, err := client.AppsV1().Deployments(namespace).Get(context.Background(), ownerName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		selector = metav1.FormatLabelSelector(deploy.Spec.Selector)
	case "StatefulSet":
		sts, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), ownerName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		selector = metav1.FormatLabelSelector(sts.Spec.Selector)
	case "DaemonSet":
		ds, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), ownerName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		selector = metav1.FormatLabelSelector(ds.Spec.Selector)
	default:
		return nil, fmt.Errorf("不支持的控制器类型: %s", ownerType)
	}

	list, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}

	result := make([]PodListVO, 0, len(list.Items))
	for _, item := range list.Items {
		result = append(result, PodListVO{
			Name:      item.Name,
			Namespace: item.Namespace,
			Status:    string(item.Status.Phase),
			IP:        item.Status.PodIP,
			Node:      item.Spec.NodeName,
			CreatedAt: item.CreationTimestamp.Time,
		})
	}
	return result, nil
}
