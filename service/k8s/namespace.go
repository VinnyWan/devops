package k8s

import (
	"context"
	"encoding/json"

	"devops/internal/database"
	k8smodels "devops/models/k8s"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceService Namespace服务
type NamespaceService struct {
	clusterService *ClusterService
}

// List 获取命名空间列表（简化版，仅返回名称）
// 如需完整信息，使用 Get() 方法获取单个命名空间详情
func (s *NamespaceService) List(clusterID uint) ([]k8smodels.NamespaceDTO, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 同步到数据库
	go s.syncToDatabase(clusterID, namespaces.Items)

	// 转换为 DTO
	result := make([]k8smodels.NamespaceDTO, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		result = append(result, k8smodels.NamespaceDTO{
			Name: ns.Name,
		})
	}

	return result, nil
}

// Get 获取命名空间详情
func (s *NamespaceService) Get(clusterID uint, name string) (*corev1.Namespace, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	namespace, err := clientset.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return namespace, nil
}

// Create 创建命名空间
func (s *NamespaceService) Create(clusterID uint, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	ns, err := clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	// 同步到数据库
	go s.syncOneToDatabase(clusterID, ns)

	return ns, nil
}

// Delete 删除命名空间
func (s *NamespaceService) Delete(clusterID uint, name string) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	err = clientset.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	// 从数据库删除
	go s.deleteFromDatabase(clusterID, name)

	return nil
}

// syncToDatabase 同步命名空间列表到数据库
func (s *NamespaceService) syncToDatabase(clusterID uint, namespaces []corev1.Namespace) {
	for _, ns := range namespaces {
		s.syncOneToDatabase(clusterID, &ns)
	}
}

// syncOneToDatabase 同步单个命名空间到数据库
func (s *NamespaceService) syncOneToDatabase(clusterID uint, ns *corev1.Namespace) {
	labels, _ := json.Marshal(ns.Labels)
	annotations, _ := json.Marshal(ns.Annotations)

	dbNs := k8smodels.Namespace{
		ClusterID:   clusterID,
		Name:        ns.Name,
		Labels:      string(labels),
		Annotations: string(annotations),
		Status:      string(ns.Status.Phase),
	}

	// 先尝试更新，如果不存在则创建
	var existing k8smodels.Namespace
	if err := database.Db.Where("cluster_id = ? AND name = ?", clusterID, ns.Name).First(&existing).Error; err == nil {
		database.Db.Model(&existing).Updates(&dbNs)
	} else {
		database.Db.Create(&dbNs)
	}
}

// deleteFromDatabase 从数据库删除命名空间记录
func (s *NamespaceService) deleteFromDatabase(clusterID uint, name string) {
	database.Db.Where("cluster_id = ? AND name = ?", clusterID, name).Delete(&k8smodels.Namespace{})
}
