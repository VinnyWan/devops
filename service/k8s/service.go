package k8s

import (
"context"

corev1 "k8s.io/api/core/v1"
networkingv1 "k8s.io/api/networking/v1"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K8sServiceService Service和Ingress服务
type K8sServiceService struct {
	clusterService *ClusterService
}

// ListServices 获取Service列表
func (s *K8sServiceService) ListServices(clusterID uint, namespace string) ([]corev1.Service, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	services, err := clientset.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return services.Items, nil
}

// GetService 获取Service详情
func (s *K8sServiceService) GetService(clusterID uint, namespace, name string) (*corev1.Service, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	service, err := clientset.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return service, nil
}

// CreateService 创建Service
func (s *K8sServiceService) CreateService(clusterID uint, namespace string, service *corev1.Service) (*corev1.Service, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	svc, err := clientset.CoreV1().Services(namespace).Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return svc, nil
}

// DeleteService 删除Service
func (s *K8sServiceService) DeleteService(clusterID uint, namespace, name string) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	return clientset.CoreV1().Services(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

// Ingress相关

// ListIngresses 获取Ingress列表
func (s *K8sServiceService) ListIngresses(clusterID uint, namespace string) ([]networkingv1.Ingress, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ingresses, err := clientset.NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return ingresses.Items, nil
}

// GetIngress 获取Ingress详情
func (s *K8sServiceService) GetIngress(clusterID uint, namespace, name string) (*networkingv1.Ingress, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ingress, err := clientset.NetworkingV1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ingress, nil
}

// CreateIngress 创建Ingress
func (s *K8sServiceService) CreateIngress(clusterID uint, namespace string, ingress *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ing, err := clientset.NetworkingV1().Ingresses(namespace).Create(context.Background(), ingress, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return ing, nil
}

// DeleteIngress 删除Ingress
func (s *K8sServiceService) DeleteIngress(clusterID uint, namespace, name string) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	return clientset.NetworkingV1().Ingresses(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
