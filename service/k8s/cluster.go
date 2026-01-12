package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"devops/internal/database"
	k8smodels "devops/models/k8s"
	usermodels "devops/models/user"

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

	// 设置导入状态
	cluster.ImportStatus = "importing"
	cluster.ImportMethod = "kubeconfig"

	// 获取并验证K8s版本
	version, status, err := s.getClusterVersion(cluster.KubeConfig)
	if err != nil {
		cluster.ImportStatus = "failed"
		cluster.ClusterStatus = "unhealthy"
		// 仍然存储，但标记为失败
		dbErr := database.Db.Create(cluster).Error
		if dbErr != nil {
			return dbErr
		}
		return errors.New("集群连接失败: " + err.Error())
	}

	// 验证版本是否支持（要求 >= 1.17）
	if !isVersionSupported(version) {
		return fmt.Errorf("K8s版本不支持，要求 >= 1.17，当前版本: %s", version)
	}

	cluster.Version = version
	cluster.ClusterStatus = status
	cluster.ImportStatus = "success"

	// TODO: 加密KubeConfig存储
	return database.Db.Create(cluster).Error
}

// Update 更新集群
func (s *ClusterService) Update(id uint, cluster *k8smodels.Cluster) error {
	var existCluster k8smodels.Cluster
	if err := database.Db.First(&existCluster, id).Error; err != nil {
		return errors.New("集群不存在")
	}

	// 如果更新了KubeConfig，需要验证并重新获取版本
	if cluster.KubeConfig != "" && cluster.KubeConfig != existCluster.KubeConfig {
		if err := s.validateKubeConfig(cluster.KubeConfig); err != nil {
			return errors.New("KubeConfig验证失败: " + err.Error())
		}

		// 设置导入状态
		cluster.ImportStatus = "importing"

		// 获取并验证K8s版本
		version, status, err := s.getClusterVersion(cluster.KubeConfig)
		if err != nil {
			cluster.ImportStatus = "failed"
			cluster.ClusterStatus = "unhealthy"
			// 仍然更新，但标记为失败
			cluster.ID = id
			dbErr := database.Db.Model(&existCluster).Updates(cluster).Error
			if dbErr != nil {
				return dbErr
			}
			return errors.New("集群连接失败: " + err.Error())
		}

		// 验证版本是否支持（要求 >= 1.17）
		if !isVersionSupported(version) {
			return fmt.Errorf("K8s版本不支持，要求 >= 1.17，当前版本: %s", version)
		}

		cluster.Version = version
		cluster.ClusterStatus = status
		cluster.ImportStatus = "success"
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
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&clusters).Error; err != nil {
		return nil, 0, err
	}

	// 不返回敏感信息
	for i := range clusters {
		clusters[i].KubeConfig = ""
	}

	return clusters, total, nil
}

// GetListByUser 根据用户权限获取集群列表
func (s *ClusterService) GetListByUser(userID uint, page, pageSize int, name string, deptID uint) ([]k8smodels.Cluster, int64, error) {
	// 1. 检查是否为 admin 用户
	var user usermodels.User
	if err := database.Db.First(&user, userID).Error; err != nil {
		return nil, 0, errors.New("用户不存在")
	}

	// admin 用户可以查看所有集群
	if user.Username == "admin" {
		return s.GetList(page, pageSize, name, deptID)
	}

	// 2. 非 admin 用户，查询有权限的集群
	// 2.1 获取用户的角色
	if err := database.Db.Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, 0, errors.New("用户不存在")
	}

	if len(user.Roles) == 0 {
		// 没有角色，返回空列表
		return []k8smodels.Cluster{}, 0, nil
	}

	// 2.2 获取角色ID列表
	roleIDs := make([]uint, len(user.Roles))
	for i, role := range user.Roles {
		roleIDs[i] = role.ID
	}

	// 2.3 查询有权限访问的集群ID
	var accesses []k8smodels.ClusterAccess
	if err := database.Db.Where("role_id IN ?", roleIDs).Find(&accesses).Error; err != nil {
		return nil, 0, err
	}

	if len(accesses) == 0 {
		// 没有任何集群访问权限，返回空列表
		return []k8smodels.Cluster{}, 0, nil
	}

	// 2.4 提取可访问的集群ID
	clusterIDs := make([]uint, 0)
	clusterIDMap := make(map[uint]bool)
	for _, access := range accesses {
		if !clusterIDMap[access.ClusterID] {
			clusterIDs = append(clusterIDs, access.ClusterID)
			clusterIDMap[access.ClusterID] = true
		}
	}

	// 2.5 查询集群列表
	var clusters []k8smodels.Cluster
	var total int64

	query := database.Db.Model(&k8smodels.Cluster{}).Where("id IN ?", clusterIDs)

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
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&clusters).Error; err != nil {
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
		// 更新集群状态为不健康
		database.Db.Model(&k8smodels.Cluster{}).Where("id = ?", clusterID).Updates(map[string]interface{}{
			"cluster_status": "unhealthy",
		})
		return false, err.Error(), err
	}

	// 更新集群状态为健康
	database.Db.Model(&k8smodels.Cluster{}).Where("id = ?", clusterID).Updates(map[string]interface{}{
		"cluster_status": "healthy",
		"version":        version.GitVersion,
	})

	return true, version.String(), nil
}

// ReimportKubeConfig 重新导入KubeConfig
func (s *ClusterService) ReimportKubeConfig(id uint, kubeconfig string) error {
	var cluster k8smodels.Cluster
	if err := database.Db.First(&cluster, id).Error; err != nil {
		return errors.New("集群不存在")
	}

	// 验证KubeConfig
	if err := s.validateKubeConfig(kubeconfig); err != nil {
		return errors.New("KubeConfig验证失败: " + err.Error())
	}

	// 设置导入状态
	updates := map[string]interface{}{
		"kube_config":   kubeconfig,
		"import_status": "importing",
		"import_method": "kubeconfig",
	}

	// 先更新导入状态
	if err := database.Db.Model(&cluster).Updates(updates).Error; err != nil {
		return err
	}

	// 获取并验证K8s版本
	version, status, err := s.getClusterVersion(kubeconfig)
	if err != nil {
		// 更新为失败状态
		database.Db.Model(&cluster).Updates(map[string]interface{}{
			"import_status":  "failed",
			"cluster_status": "unhealthy",
		})
		return errors.New("集群连接失败: " + err.Error())
	}

	// 验证版本是否支持（要求 >= 1.17）
	if !isVersionSupported(version) {
		database.Db.Model(&cluster).Updates(map[string]interface{}{
			"import_status": "failed",
		})
		return fmt.Errorf("K8s版本不支持，要求 >= 1.17，当前版本: %s", version)
	}

	// 更新为成功状态
	return database.Db.Model(&cluster).Updates(map[string]interface{}{
		"version":        version,
		"cluster_status": status,
		"import_status":  "success",
	}).Error
}

// validateKubeConfig 验证KubeConfig格式
func (s *ClusterService) validateKubeConfig(kubeconfig string) error {
	_, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	return err
}

// getClusterVersion 获取集群版本和状态
func (s *ClusterService) getClusterVersion(kubeconfig string) (string, string, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return "", "unhealthy", fmt.Errorf("解析KubeConfig失败: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "unhealthy", fmt.Errorf("创建K8s客户端失败: %w", err)
	}

	// 获取版本信息
	versionInfo, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return "", "unhealthy", fmt.Errorf("获取集群版本失败: %w", err)
	}

	// 返回简化的版本号（如 v1.23.5）
	return versionInfo.GitVersion, "healthy", nil
}

// isVersionSupported 检查版本是否支持（>= 1.17，兼容所有高版本）
func isVersionSupported(version string) bool {
	// 移除 'v' 前缀
	version = strings.TrimPrefix(version, "v")

	// 提取主版本和次版本号
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false
	}

	// 移除次版本号中可能包含的额外信息（如 "23+k3s1" -> "23"）
	minorStr := parts[1]
	if idx := strings.IndexAny(minorStr, "+-"); idx != -1 {
		minorStr = minorStr[:idx]
	}

	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(minorStr)

	if err1 != nil || err2 != nil {
		return false
	}

	// 检查版本 >= 1.17
	// 支持 Kubernetes 1.17 及以上所有版本
	if major > 1 {
		return true // 支持 K8s 2.x 及更高版本
	}
	if major == 1 && minor >= 17 {
		return true // 支持 1.17, 1.18, 1.19, ..., 1.30+
	}

	return false
}

// PermissionService 权限服务
type PermissionService struct{}

// CheckAccess 检查用户对集群的访问权限
func (s *PermissionService) CheckAccess(userID, clusterID uint, operation string) (string, []string, error) {
	// 1. 获取用户信息
	var user usermodels.User

	if err := database.Db.Preload("Roles").First(&user, userID).Error; err != nil {
		return "", nil, errors.New("用户不存在")
	}

	// 2. 检查是否为 admin 用户（超级管理员，拥有所有权限）
	if user.Username == "admin" {
		// admin 用户拥有最高权限，可以访问所有集群和命名空间
		return "admin", nil, nil
	}

	// 3. 检查用户是否有角色
	if len(user.Roles) == 0 {
		return "", nil, errors.New("用户没有分配角色")
	}

	// 4. 查询角色对集群的访问权限
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

	// 5. 确定最高权限
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

	// 6. 检查操作权限
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
