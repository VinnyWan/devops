package service

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceVO struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"clusterIP"`
	Ports     []int32           `json:"ports"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"createdAt"`
}

type ServiceListResponse struct {
	Total int64       `json:"total"`
	Items []ServiceVO `json:"items"`
}

func (s *K8sService) ListServices(clusterId uint, namespace string, page, pageSize int, keyword string) (*ServiceListResponse, error) {
	cc, err := s.getClusterClient(clusterId)
	if err != nil {
		return nil, err
	}

	list, err := cc.Client.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, s.handleClientError(clusterId, err)
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item corev1.Service) []string {
		return []string{
			item.Name,
			item.Namespace,
			string(item.Spec.Type),
			item.Spec.ClusterIP,
			flattenLabels(item.Labels),
		}
	})

	paged, total := paginateItems(filtered, page, pageSize)

	result := make([]ServiceVO, 0, len(paged))
	for _, item := range paged {
		ports := make([]int32, 0, len(item.Spec.Ports))
		for _, p := range item.Spec.Ports {
			ports = append(ports, p.Port)
		}
		result = append(result, ServiceVO{
			Name:      item.Name,
			Namespace: item.Namespace,
			Type:      string(item.Spec.Type),
			ClusterIP: item.Spec.ClusterIP,
			Ports:     ports,
			Labels:    item.Labels,
			CreatedAt: item.CreationTimestamp.Time,
		})
	}
	return &ServiceListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) GetServiceDetail(clusterId uint, namespace, name string) (*ServiceVO, error) {
	cc, err := s.getClusterClient(clusterId)
	if err != nil {
		return nil, err
	}

	item, err := cc.Client.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	ports := make([]int32, 0, len(item.Spec.Ports))
	for _, p := range item.Spec.Ports {
		ports = append(ports, p.Port)
	}

	return &ServiceVO{
		Name:      item.Name,
		Namespace: item.Namespace,
		Type:      string(item.Spec.Type),
		ClusterIP: item.Spec.ClusterIP,
		Ports:     ports,
		Labels:    item.Labels,
		CreatedAt: item.CreationTimestamp.Time,
	}, nil
}

func (s *K8sService) DeleteService(clusterId uint, namespace, name string) error {
	cc, err := s.getClusterClient(clusterId)
	if err != nil {
		return err
	}
	return cc.Client.CoreV1().Services(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (s *K8sService) CreateService(clusterId uint, namespace string, svc *corev1.Service) (*corev1.Service, error) {
	cc, err := s.getClusterClient(clusterId)
	if err != nil {
		return nil, err
	}
	return cc.Client.CoreV1().Services(namespace).Create(context.Background(), svc, metav1.CreateOptions{})
}

func (s *K8sService) UpdateService(clusterId uint, namespace string, svc *corev1.Service) (*corev1.Service, error) {
	cc, err := s.getClusterClient(clusterId)
	if err != nil {
		return nil, err
	}
	return cc.Client.CoreV1().Services(namespace).Update(context.Background(), svc, metav1.UpdateOptions{})
}
