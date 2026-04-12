package service

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceVO struct {
	Name              string    `json:"name"`
	Status            string    `json:"status"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
}

type NamespaceListResponse struct {
	Total int64         `json:"total"`
	Items []NamespaceVO `json:"items"`
}

func (s *K8sService) ListNamespaces(clusterName string, keyword string, page, pageSize int) (*NamespaceListResponse, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}

	list, err := cc.Client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, s.handleClientError(clusterName, err)
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item corev1.Namespace) []string {
		return []string{item.Name}
	})

	paged, total := paginateItems(filtered, page, pageSize)

	result := make([]NamespaceVO, 0, len(paged))
	for _, item := range paged {
		result = append(result, NamespaceVO{
			Name:              item.Name,
			Status:            string(item.Status.Phase),
			CreationTimestamp: item.CreationTimestamp.Time,
		})
	}
	return &NamespaceListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) CreateNamespace(clusterName string, name string) (*NamespaceVO, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	created, err := cc.Client.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	if err != nil {
		return nil, s.handleClientError(clusterName, err)
	}

	return &NamespaceVO{
		Name:              created.Name,
		Status:            string(created.Status.Phase),
		CreationTimestamp: created.CreationTimestamp.Time,
	}, nil
}

func (s *K8sService) DeleteNamespace(clusterName string, name string) error {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return err
	}
	return cc.Client.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
}
