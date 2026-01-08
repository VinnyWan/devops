package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"devops/internal/database"
	k8smodels "devops/models/k8s"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterService K8s集群服务
type ClusterService struct{}

// Create 创建集群
func (s *ClusterService) Create(cluster *k8smodels.Cluster) error {
	// 验证KubeConfig
	if err := s.validateKubeConfig(cluster.KubeConfig); err != nil {
		return errors.New("KubeConfig验证失败: " + err.Error())
	}

	// TODO: 加密KubeConfig存储
	return database.Db.Create(cluster).Error
}

// Update 更新集群
func (s *ClusterService) Update(id uint, cluster *k8smodels.Cluster) error {
	var existCluster k8smodels.Cluster
	if err := database.Db.First(&existCluster, id).Error; err != nil {
		return errors.New("集群不存在")
	}

	// 如果更新了KubeConfig，需要验证
	if cluster.KubeConfig != "" && cluster.KubeConfig != existCluster.KubeConfig {
		if err := s.validateKubeConfig(cluster.KubeConfig); err != nil {
			return errors.New("KubeConfig验证失败: " + err.Error())
		}
	}

	cluster.ID = id
	return database.Db.Model(&existCluster).Updates(cluster).Error
}

// Delete 删除集群
func (s *ClusterService) Delete(id uint) error {
	return database.Db.Delete(&k8smodels.Cluster{}, id).Error
}

// GetByID 根据ID获取集群
func (s *ClusterService) GetByID(id uint) (*k8smodels.Cluster, error) {
	var cluster k8smodels.Cluster
	if err := database.Db.First(&cluster, id).Error; err != nil {
		return nil, err
	}
	// 不返回敏感信息
	cluster.KubeConfig = ""
	return &cluster, nil
}

// GetList 获取集群列表
func (s *ClusterService) GetList(page, pageSize int, name string, deptID uint) ([]k8smodels.Cluster, int64, error) {
	var clusters []k8smodels.Cluster
	var total int64

	query := database.Db.Model(&k8smodels.Cluster{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if deptID > 0 {
		query = query.Where("dept_id = ?", deptID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&clusters).Error; err != nil {
		return nil, 0, err
	}

	// 不返回敏感信息
	for i := range clusters {
		clusters[i].KubeConfig = ""
	}

	return clusters, total, nil
}

// GetClient 获取K8s客户端
func (s *ClusterService) GetClient(clusterID uint) (*kubernetes.Clientset, error) {
	var cluster k8smodels.Cluster
	if err := database.Db.First(&cluster, clusterID).Error; err != nil {
		return nil, errors.New("集群不存在")
	}

	if cluster.Status != 1 {
		return nil, errors.New("集群已禁用")
	}

	// TODO: 解密KubeConfig
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(cluster.KubeConfig))
	if err != nil {
		return nil, fmt.Errorf("解析KubeConfig失败: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建K8s客户端失败: %w", err)
	}

	return clientset, nil
}

// HealthCheck 健康检查
func (s *ClusterService) HealthCheck(clusterID uint) (bool, string, error) {
	clientset, err := s.GetClient(clusterID)
	if err != nil {
		return false, err.Error(), err
	}

	// 尝试获取版本信息
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return false, err.Error(), err
	}

	return true, version.String(), nil
}

// validateKubeConfig 验证KubeConfig格式
func (s *ClusterService) validateKubeConfig(kubeconfig string) error {
	_, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	return err
}

// PermissionService 权限服务
type PermissionService struct{}

// CheckAccess 检查用户对集群的访问权限
func (s *PermissionService) CheckAccess(userID, clusterID uint, operation string) (string, []string, error) {
	// 1. 获取用户的所有角色
	var user struct {
		Roles []struct {
			ID uint `json:"id"`
		} `json:"roles"`
	}

	if err := database.Db.Table("users").
		Select("users.*").
		Preload("Roles").
		Where("users.id = ?", userID).
		First(&user).Error; err != nil {
		return "", nil, errors.New("用户不存在")
	}

	if len(user.Roles) == 0 {
		return "", nil, errors.New("用户没有分配角色")
	}

	// 2. 查询角色对集群的访问权限
	roleIDs := make([]uint, len(user.Roles))
	for i, role := range user.Roles {
		roleIDs[i] = role.ID
	}

	var accesses []k8smodels.ClusterAccess
	if err := database.Db.Where("cluster_id = ? AND role_id IN ?", clusterID, roleIDs).
		Find(&accesses).Error; err != nil {
		return "", nil, err
	}

	if len(accesses) == 0 {
		return "", nil, errors.New("无权访问该集群")
	}

	// 3. 确定最高权限
	accessType := "readonly"
	var namespaces []string

	for _, access := range accesses {
		if access.AccessType == "admin" {
			accessType = "admin"
			// admin权限可以访问所有namespace
			namespaces = nil
			break
		}

		// 合并readonly权限的namespace
		if access.Namespaces != "" {
			var ns []string
			if err := json.Unmarshal([]byte(access.Namespaces), &ns); err == nil {
				namespaces = append(namespaces, ns...)
			}
		}
	}

	// 4. 检查操作权限
	if accessType == "readonly" && isWriteOperation(operation) {
		return "", nil, errors.New("只读权限，无法执行写操作")
	}

	return accessType, namespaces, nil
}

// isWriteOperation 判断是否为写操作
func isWriteOperation(operation string) bool {
	writeOps := []string{"create", "update", "delete", "patch", "scale", "restart"}
	for _, op := range writeOps {
		if op == operation {
			return true
		}
	}
	return false
}

// CreateAccess 创建集群访问权限
func (s *PermissionService) CreateAccess(access *k8smodels.ClusterAccess) error {
	// 验证访问类型
	if access.AccessType != "readonly" && access.AccessType != "admin" {
		return errors.New("访问类型必须是 readonly 或 admin")
	}

	// 验证namespaces是否为有效JSON
	if access.Namespaces != "" {
		var ns []string
		if err := json.Unmarshal([]byte(access.Namespaces), &ns); err != nil {
			return errors.New("namespaces必须是有效的JSON数组")
		}
	}

	return database.Db.Create(access).Error
}

// GetAccessList 获取集群访问权限列表
func (s *PermissionService) GetAccessList(clusterID uint) ([]k8smodels.ClusterAccess, error) {
	var accesses []k8smodels.ClusterAccess
	if err := database.Db.Where("cluster_id = ?", clusterID).Find(&accesses).Error; err != nil {
		return nil, err
	}
	return accesses, nil
}

// DeleteAccess 删除集群访问权限
func (s *PermissionService) DeleteAccess(id uint) error {
	return database.Db.Delete(&k8smodels.ClusterAccess{}, id).Error
}

// LogOperation 记录操作日志
func LogOperation(ctx context.Context, log *k8smodels.OperationLog) {
	database.Db.Create(log)
}
