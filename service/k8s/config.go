package k8s

import (
"context"

corev1 "k8s.io/api/core/v1"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigService ConfigMap和Secret服务
type ConfigService struct {
	clusterService *ClusterService
}

// ConfigMap相关

// ListConfigMaps 获取ConfigMap列表
func (s *ConfigService) ListConfigMaps(clusterID uint, namespace string) ([]corev1.ConfigMap, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	configMaps, err := clientset.CoreV1().ConfigMaps(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return configMaps.Items, nil
}

// GetConfigMap 获取ConfigMap详情
func (s *ConfigService) GetConfigMap(clusterID uint, namespace, name string) (*corev1.ConfigMap, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return configMap, nil
}

// CreateConfigMap 创建ConfigMap
func (s *ConfigService) CreateConfigMap(clusterID uint, namespace string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	cm, err := clientset.CoreV1().ConfigMaps(namespace).Create(context.Background(), configMap, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return cm, nil
}

// UpdateConfigMap 更新ConfigMap
func (s *ConfigService) UpdateConfigMap(clusterID uint, namespace string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	cm, err := clientset.CoreV1().ConfigMaps(namespace).Update(context.Background(), configMap, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return cm, nil
}

// DeleteConfigMap 删除ConfigMap
func (s *ConfigService) DeleteConfigMap(clusterID uint, namespace, name string) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	return clientset.CoreV1().ConfigMaps(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

// Secret相关

// ListSecrets 获取Secret列表
func (s *ConfigService) ListSecrets(clusterID uint, namespace string) ([]corev1.Secret, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	secrets, err := clientset.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return secrets.Items, nil
}

// GetSecret 获取Secret详情
func (s *ConfigService) GetSecret(clusterID uint, namespace, name string) (*corev1.Secret, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// CreateSecret 创建Secret
func (s *ConfigService) CreateSecret(clusterID uint, namespace string, secret *corev1.Secret) (*corev1.Secret, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	sec, err := clientset.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return sec, nil
}

// UpdateSecret 更新Secret
func (s *ConfigService) UpdateSecret(clusterID uint, namespace string, secret *corev1.Secret) (*corev1.Secret, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	sec, err := clientset.CoreV1().Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return sec, nil
}

// DeleteSecret 删除Secret
func (s *ConfigService) DeleteSecret(clusterID uint, namespace, name string) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	return clientset.CoreV1().Secrets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
