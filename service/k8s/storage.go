package k8s

import (
"context"

corev1 "k8s.io/api/core/v1"
storagev1 "k8s.io/api/storage/v1"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StorageService 存储服务
type StorageService struct {
	clusterService *ClusterService
}

// PV相关

// ListPVs 获取PV列表
func (s *StorageService) ListPVs(clusterID uint) ([]corev1.PersistentVolume, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	pvs, err := clientset.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pvs.Items, nil
}

// GetPV 获取PV详情
func (s *StorageService) GetPV(clusterID uint, name string) (*corev1.PersistentVolume, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	pv, err := clientset.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pv, nil
}

// DeletePV 删除PV
func (s *StorageService) DeletePV(clusterID uint, name string) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	return clientset.CoreV1().PersistentVolumes().Delete(context.Background(), name, metav1.DeleteOptions{})
}

// PVC相关

// ListPVCs 获取PVC列表
func (s *StorageService) ListPVCs(clusterID uint, namespace string) ([]corev1.PersistentVolumeClaim, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pvcs.Items, nil
}

// GetPVC 获取PVC详情
func (s *StorageService) GetPVC(clusterID uint, namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	pvc, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pvc, nil
}

// DeletePVC 删除PVC
func (s *StorageService) DeletePVC(clusterID uint, namespace, name string) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	return clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

// StorageClass相关

// ListStorageClasses 获取StorageClass列表
func (s *StorageService) ListStorageClasses(clusterID uint) ([]storagev1.StorageClass, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	scs, err := clientset.StorageV1().StorageClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return scs.Items, nil
}

// GetStorageClass 获取StorageClass详情
func (s *StorageService) GetStorageClass(clusterID uint, name string) (*storagev1.StorageClass, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	sc, err := clientset.StorageV1().StorageClasses().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return sc, nil
}
