package service

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// --- StorageClass ---

type StorageClassVO struct {
	Name                string `json:"name"`
	Provisioner         string `json:"provisioner"`
	ReclaimPolicy       string `json:"reclaimPolicy"`
	VolumeBindingMode   string `json:"volumeBindingMode"`
	AllowVolumeExpansion *bool `json:"allowVolumeExpansion"`
	IsDefault           bool   `json:"isDefault"`
	CreatedAt           time.Time `json:"createdAt"`
}

func (s *K8sService) ListStorageClasses(clusterName string) ([]StorageClassVO, error) {
	client, err := s.getClient(clusterName)
	if err != nil {
		return nil, err
	}

	list, err := client.StorageV1().StorageClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]StorageClassVO, 0, len(list.Items))
	for _, item := range list.Items {
		isDefault := false
		if item.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			isDefault = true
		}

		result = append(result, StorageClassVO{
			Name:                item.Name,
			Provisioner:         item.Provisioner,
			ReclaimPolicy:       string(*item.ReclaimPolicy),
			VolumeBindingMode:   string(*item.VolumeBindingMode),
			AllowVolumeExpansion: item.AllowVolumeExpansion,
			IsDefault:           isDefault,
			CreatedAt:           item.CreationTimestamp.Time,
		})
	}
	return result, nil
}

// --- PV ---

type PVVO struct {
	Name          string            `json:"name"`
	Capacity      string            `json:"capacity"`
	AccessModes   []string          `json:"accessModes"`
	ReclaimPolicy string            `json:"reclaimPolicy"`
	Status        string            `json:"status"`
	Claim         string            `json:"claim"`
	StorageClass  string            `json:"storageClass"`
	Reason        string            `json:"reason"`
	CreatedAt     time.Time         `json:"createdAt"`
}

func (s *K8sService) ListPersistentVolumes(clusterName string) ([]PVVO, error) {
	client, err := s.getClient(clusterName)
	if err != nil {
		return nil, err
	}

	list, err := client.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]PVVO, 0, len(list.Items))
	for _, item := range list.Items {
		capacity := ""
		if item.Spec.Capacity != nil {
			if qty, ok := item.Spec.Capacity[corev1.ResourceStorage]; ok {
				capacity = qty.String()
			}
		}

		claim := ""
		if item.Spec.ClaimRef != nil {
			claim = item.Spec.ClaimRef.Namespace + "/" + item.Spec.ClaimRef.Name
		}

		accessModes := make([]string, 0, len(item.Spec.AccessModes))
		for _, am := range item.Spec.AccessModes {
			accessModes = append(accessModes, string(am))
		}

		result = append(result, PVVO{
			Name:          item.Name,
			Capacity:      capacity,
			AccessModes:   accessModes,
			ReclaimPolicy: string(item.Spec.PersistentVolumeReclaimPolicy),
			Status:        string(item.Status.Phase),
			Claim:         claim,
			StorageClass:  item.Spec.StorageClassName,
			Reason:        item.Status.Reason,
			CreatedAt:     item.CreationTimestamp.Time,
		})
	}
	return result, nil
}

// --- PVC ---

type PVCVO struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Status       string    `json:"status"`
	Volume       string    `json:"volume"`
	Capacity     string    `json:"capacity"`
	AccessModes  []string  `json:"accessModes"`
	StorageClass string    `json:"storageClass"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (s *K8sService) ListPersistentVolumeClaims(clusterName, namespace string) ([]PVCVO, error) {
	client, err := s.getClient(clusterName)
	if err != nil {
		return nil, err
	}

	list, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]PVCVO, 0, len(list.Items))
	for _, item := range list.Items {
		capacity := ""
		if item.Status.Capacity != nil {
			if qty, ok := item.Status.Capacity[corev1.ResourceStorage]; ok {
				capacity = qty.String()
			}
		}

		accessModes := make([]string, 0, len(item.Spec.AccessModes))
		for _, am := range item.Spec.AccessModes {
			accessModes = append(accessModes, string(am))
		}

		result = append(result, PVCVO{
			Name:         item.Name,
			Namespace:    item.Namespace,
			Status:       string(item.Status.Phase),
			Volume:       item.Spec.VolumeName,
			Capacity:     capacity,
			AccessModes:  accessModes,
			StorageClass: func() string {
				if item.Spec.StorageClassName != nil {
					return *item.Spec.StorageClassName
				}
				return ""
			}(),
			CreatedAt: item.CreationTimestamp.Time,
		})
	}
	return result, nil
}

// --- YAML Update ---

func (s *K8sService) UpdateStorageClassByYAML(clusterName, name, rawYAML string) (*storagev1.StorageClass, error) {
	client, err := s.getClient(clusterName)
	if err != nil {
		return nil, err
	}

	current, err := client.StorageV1().StorageClasses().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 StorageClass 失败: %w", err)
	}

	var desired storagev1.StorageClass
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("无效的 YAML: %w", err)
	}

	if desired.Name != "" && desired.Name != name {
		return nil, fmt.Errorf("不允许修改 metadata.name")
	}

	desired.Name = name
	desired.ResourceVersion = current.ResourceVersion

	return client.StorageV1().StorageClasses().Update(context.Background(), &desired, metav1.UpdateOptions{})
}

func (s *K8sService) UpdatePVByYAML(clusterName, name, rawYAML string) (*corev1.PersistentVolume, error) {
	client, err := s.getClient(clusterName)
	if err != nil {
		return nil, err
	}

	current, err := client.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PV 失败: %w", err)
	}

	var desired corev1.PersistentVolume
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("无效的 YAML: %w", err)
	}

	if desired.Name != "" && desired.Name != name {
		return nil, fmt.Errorf("不允许修改 metadata.name")
	}

	desired.Name = name
	desired.ResourceVersion = current.ResourceVersion
	desired.Status = corev1.PersistentVolumeStatus{}

	return client.CoreV1().PersistentVolumes().Update(context.Background(), &desired, metav1.UpdateOptions{})
}

func (s *K8sService) UpdatePVCByYAML(clusterName, namespace, name, rawYAML string) (*corev1.PersistentVolumeClaim, error) {
	client, err := s.getClient(clusterName)
	if err != nil {
		return nil, err
	}

	current, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PVC 失败: %w", err)
	}

	var desired corev1.PersistentVolumeClaim
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("无效的 YAML: %w", err)
	}

	if desired.Name != "" && desired.Name != name {
		return nil, fmt.Errorf("不允许修改 metadata.name")
	}
	if desired.Namespace != "" && desired.Namespace != namespace {
		return nil, fmt.Errorf("不允许修改 metadata.namespace")
	}

	desired.Namespace = namespace
	desired.Name = name
	desired.ResourceVersion = current.ResourceVersion
	desired.Status = corev1.PersistentVolumeClaimStatus{}

	return client.CoreV1().PersistentVolumeClaims(namespace).Update(context.Background(), &desired, metav1.UpdateOptions{})
}

func (s *K8sService) DeletePVC(clusterName, namespace, name string) error {
	client, err := s.getClient(clusterName)
	if err != nil {
		return err
	}
	return client.CoreV1().PersistentVolumeClaims(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
