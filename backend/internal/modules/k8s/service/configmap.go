package service

import (
	"context"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapVO struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	DataKeys  []string          `json:"dataKeys"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"createdAt"`
}

type ConfigMapListVO struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	CreatedAt time.Time `json:"createdAt"`
}

type ConfigMapListResponse struct {
	Total int64             `json:"total"`
	Items []ConfigMapListVO `json:"items"`
}

func (s *K8sService) ListConfigMaps(clusterName string, namespace string, page, pageSize int, keyword string) (*ConfigMapListResponse, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}

	list, err := cc.Client.CoreV1().ConfigMaps(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, s.handleClientError(clusterName, err)
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item corev1.ConfigMap) []string {
		dataKeys := make([]string, 0, len(item.Data))
		for key := range item.Data {
			dataKeys = append(dataKeys, key)
		}
		return []string{
			item.Name,
			item.Namespace,
			flattenLabels(item.Labels),
			strings.Join(dataKeys, ","),
		}
	})

	paged, total := paginateItems(filtered, page, pageSize)

	result := make([]ConfigMapListVO, 0, len(paged))
	for _, item := range paged {
		result = append(result, ConfigMapListVO{
			Name:      item.Name,
			Namespace: item.Namespace,
			CreatedAt: item.CreationTimestamp.Time,
		})
	}
	return &ConfigMapListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) GetConfigMapDetail(clusterName string, namespace, name string) (*ConfigMapVO, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}

	item, err := cc.Client.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(item.Data))
	for k := range item.Data {
		keys = append(keys, k)
	}

	return &ConfigMapVO{
		Name:      item.Name,
		Namespace: item.Namespace,
		DataKeys:  keys,
		Labels:    item.Labels,
		CreatedAt: item.CreationTimestamp.Time,
	}, nil
}

func (s *K8sService) DeleteConfigMap(clusterName string, namespace, name string) error {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return err
	}
	return cc.Client.CoreV1().ConfigMaps(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (s *K8sService) CreateConfigMap(clusterName string, namespace string, cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}
	return cc.Client.CoreV1().ConfigMaps(namespace).Create(context.Background(), cm, metav1.CreateOptions{})
}

func (s *K8sService) UpdateConfigMap(clusterName string, namespace string, cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}
	return cc.Client.CoreV1().ConfigMaps(namespace).Update(context.Background(), cm, metav1.UpdateOptions{})
}
