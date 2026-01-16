package database

import (
	"devops/internal/logger"
	k8smodels "devops/models/k8s"
	usermodels "devops/models/user"
	"devops/utils"

	"go.uber.org/zap"
)

// InitData 初始化基础数据
func InitData() error {
	logger.Log.Info("开始初始化基础数据...")

	// 检查是否已有管理员用户
	var count int64
	if err := Db.Model(&usermodels.User{}).Count(&count).Error; err != nil {
		logger.Log.Error("检查管理员用户失败", zap.Error(err))
		return err
	}
	if count > 0 {
		logger.Log.Info("数据已存在，跳过初始化")
		return nil
	}

	// 创建默认管理员角色
	adminRole := usermodels.Role{
		RoleName: "超级管理员",
		RoleKey:  "admin",
		Sort:     1,
		Status:   1,
		Remark:   "系统内置超级管理员角色",
	}
	if err := Db.Create(&adminRole).Error; err != nil {
		logger.Log.Error("创建管理员角色失败", zap.Error(err))
		return err
	}

	// 创建默认管理员用户
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		logger.Log.Error("管理员密码加密失败", zap.Error(err))
		return err
	}
	adminUser := usermodels.User{
		Username: "admin",
		Password: hashedPassword,
		Nickname: "超级管理员",
		Email:    "admin@example.com",
		Phone:    "13800138000",
		Status:   1,
		Gender:   1,
		Remark:   "系统内置超级管理员",
	}
	if err := Db.Create(&adminUser).Error; err != nil {
		logger.Log.Error("创建管理员用户失败", zap.Error(err))
		return err
	}

	// 关联角色
	if err := Db.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
		logger.Log.Error("关联角色失败", zap.Error(err))
		return err
	}

	// 创建默认部门
	dept := usermodels.Department{
		DeptName: "总公司",
		ParentID: 0,
		Sort:     0,
		Leader:   "超级管理员",
		Phone:    "13800138000",
		Email:    "admin@example.com",
		Status:   1,
		Remark:   "顶级部门",
	}
	if err := Db.Create(&dept).Error; err != nil {
		logger.Log.Error("创建默认部门失败", zap.Error(err))
		return err
	}

	// 创建默认岗位
	post := usermodels.Post{
		PostName: "董事长",
		PostCode: "ceo",
		Sort:     1,
		Status:   1,
		Remark:   "公司最高管理者",
	}
	if err := Db.Create(&post).Error; err != nil {
		logger.Log.Error("创建默认岗位失败", zap.Error(err))
		return err
	}

	// 创建系统菜单
	menus := []usermodels.Menu{
		{
			MenuName:  "系统管理",
			ParentID:  0,
			Sort:      1,
			Path:      "/system",
			Component: "Layout",
			MenuType:  "M",
			Visible:   1,
			Status:    1,
			Icon:      "system",
			Remark:    "系统管理目录",
		},
		{
			MenuName:  "用户管理",
			ParentID:  1,
			Sort:      1,
			Path:      "user",
			Component: "system/user/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:user:list",
			Icon:      "user",
			Remark:    "用户管理菜单",
		},
		{
			MenuName:  "角色管理",
			ParentID:  1,
			Sort:      2,
			Path:      "role",
			Component: "system/role/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:role:list",
			Icon:      "peoples",
			Remark:    "角色管理菜单",
		},
		{
			MenuName:  "菜单管理",
			ParentID:  1,
			Sort:      3,
			Path:      "menu",
			Component: "system/menu/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:menu:list",
			Icon:      "tree-table",
			Remark:    "菜单管理菜单",
		},
		{
			MenuName:  "部门管理",
			ParentID:  1,
			Sort:      4,
			Path:      "dept",
			Component: "system/dept/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:dept:list",
			Icon:      "tree",
			Remark:    "部门管理菜单",
		},
		{
			MenuName:  "岗位管理",
			ParentID:  1,
			Sort:      5,
			Path:      "post",
			Component: "system/post/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:post:list",
			Icon:      "post",
			Remark:    "岗位管理菜单",
		},
		{
			MenuName:  "操作日志",
			ParentID:  1,
			Sort:      6,
			Path:      "operlog",
			Component: "system/operlog/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:operlog:list",
			Icon:      "form",
			Remark:    "操作日志菜单",
		},
		{
			MenuName:  "登录日志",
			ParentID:  1,
			Sort:      7,
			Path:      "loginlog",
			Component: "system/loginlog/index",
			MenuType:  "C",
			Visible:   1,
			Status:    1,
			Perms:     "system:loginlog:list",
			Icon:      "logininfor",
			Remark:    "登录日志菜单",
		},
	}

	for _, menu := range menus {
		if err := Db.Create(&menu).Error; err != nil {
			logger.Log.Error("创建菜单失败", zap.Error(err))
			return err
		}
	}

	// 创建 K8s 测试集群数据（使用默认部门ID）
	if err := initK8sTestClusters(dept.ID); err != nil {
		logger.Log.Error("创建 K8s 测试集群失败", zap.Error(err))
		// 不影响主流程，继续执行
	}

	logger.Log.Info("基础数据初始化完成")
	logger.Log.Info("默认管理员账号: admin, 密码: admin123")
	return nil
}

// InitK8sTestData 独立初始化K8s测试数据（可多次调用）
func InitK8sTestData() error {
	logger.Log.Info("检查K8s测试数据...")

	// 检查是否已有K8s集群数据
	var count int64
	if err := Db.Model(&k8smodels.Cluster{}).Count(&count).Error; err != nil {
		logger.Log.Error("检查K8s集群数据失败", zap.Error(err))
		return err
	}
	if count > 0 {
		logger.Log.Info("K8s集群数据已存在，跳过初始化", zap.Int64("集群数量", count))
		return nil
	}

	// 获取默认部门ID（如果存在）
	var dept usermodels.Department
	if err := Db.First(&dept).Error; err != nil {
		logger.Log.Warn("未找到默认部门，使用部门ID=1", zap.Error(err))
		return initK8sTestClusters(1) // 使用默认ID
	}

	return initK8sTestClusters(dept.ID)
}

// initK8sTestClusters 初始化K8s测试集群数据
func initK8sTestClusters(deptID uint) error {
	logger.Log.Info("检查 K8s 测试集群数据...")

	// 先检查是否已有K8s集群数据
	var existingCount int64
	if err := Db.Model(&k8smodels.Cluster{}).Count(&existingCount).Error; err != nil {
		logger.Log.Error("检查K8s集群数据失败", zap.Error(err))
		return err
	}
	if existingCount > 0 {
		logger.Log.Info("K8s 集群数据已存在，取消初始化",
			zap.Int64("已有集群数量", existingCount))
		return nil
	}

	logger.Log.Info("开始初始化 K8s 测试集群数据...")

	// 模拟KubeConfig（示例格式）
	testKubeConfig := `apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkakNDQVIyZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdGMyVnkKZG1WeUxXTmhRREUzTXpZMU5EQTRNemN3SGhjTk1qWXdNVEF4TURFd05qRTNXaGNOTXpZd01EQTVNREV3TmpFMwpXakFqTVNFd0h3WURWUVFEREJock0zTXRjMlZ5ZG1WeUxXTmhRREUzTXpZMU5EQTRNemN3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFUTmVHbGtKZ3RXbVJOUlVycHU3M3lrcko4RlhVMlZFTGo2cS9TMHVBbkoKbFpDa1dXOVJJcTVMMXh2TktLV0lGUGprcW54UVRCZHRybWl1UE1DQzdXSU5vMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVVo4SXRNTlRVT25qZ3M2MnFBdmhrCnFxcW5uRXd3Q2dZSUtvWkl6ajBFQXdJRFJ3QXdSQUlnV3pTQmRYUzJnVm1hZlZKSjRnRkJBZkNWOXphRFVzYkoKOUVhWDcwL0NsQkVDSUY4dVo2TWRLSjBQcFppbzE5VFpFU3VSYXNBV1FiSDRKZVdCZmh5L2ZzMgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0t
    server: https://127.0.0.1:6443
  name: k3s-default
contexts:
- context:
    cluster: k3s-default
    user: k3s-default
  name: k3s-default
current-context: k3s-default
users:
- name: k3s-default
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJrakNDQVRlZ0F3SUJBZ0lJZFNHc1Q0RzFMclF3Q2dZSUtvWkl6ajBFQXdJd0l6RWhNQjhHQTFVRUF3d1kKYXpOekxXTnNhV1Z1ZEMxallVQXhOek0yTlRRd09ETTNNQTR4RGpBTUJnTlZCQU1NQldGa2JXbHVNQjRYRFRJMgpNREV3TVRBeE1EWTVOME1YRFRJM01ERXdNVEF4TURZNU4xb3dNREVYTUJVR0ExVUVDaE1PYzNsemRHVnRPbTFoCmMzUmxjbk14RlRBVEJnTlZCQU1UREhONWMzUmxiVHBoWkdsdGFUQlpNQk1HQnlxR1NNNDlBZ0VHQ0NxR1NNNDkKQXdFSEEwSUFCRmxpc1hNbVRqbFFZV29FNTBFUTloR09LZnJMRFFPMFlmYWZHZ3Uyd080cWNsNnM1cXBJWUhaYgpYdnBmcGdIcjVGOUhOOWRMS0FWelduZnErbU9mS0JxalNEQkdNQTRHQTFVZER3RUIvd1FFQXdJRm9EQVRCZ05WCkhTVUVEREFLQmdnckJnRUZCUWNEQWpBZkJnTlZIU01FR0RBV2dCUnFuVk5QV1FkQUxYb25HbDZYa3RldUlwNGIKU1RBS0JnZ3Foa2pPUFFRREFnTkpBREJHQWlFQXhoV2Y5NjBPRHhQWlpIUHI5dkdFa0FYQldzWmUrVGRWL29NNApSeElNZ0FJRFFVWHRTY0VVTUg5NUNmU3YycUoxRzNiQ1Z4Y3UwRHI1cUUrcWc4UkRjQkU9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkVENDQVJ5Z0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdFkyeHAKWlc1MExXTmhRREUzTXpZMU5EQTRNemN3SGhjTk1qWXdNVEF4TURFd05qRTNXaGNOTXpZd01EQTVNREV3TmpFMwpXakFqTVNFd0h3WURWUVFEREJock0zTXRZMnhwWlc1MExXTmhRREUzTXpZMU5EQTRNemN3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFTNlhOdERjUStNUlZsZ0lVY1piZEw3UkJOckJVMEpTT3pTZjdFd1p0bnUKYktQYmV4QWEyaFEvL1FoeE56cit4S2pUWFd0Yk5xU2xIL1JIT1dKRjNyS1hvMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVWFwMVRUMWtIUUMxNkp4cGVsNUxYCnJpS2VHMGt3Q2dZSUtvWkl6ajBFQXdJRFJ3QXdSQUlnZXhWNU5tSU9YS3k4aU5rUWZsWHlXaHdVRCtxVU1Dc2oKaUg5aG9GRnZQMGdDSUNsMFdBMHZBK01XQ3U1MEs1cDFKQ2poaVFvczVqYlBqdEFLM2EvWnJ2YXMKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1IY0NBUUVFSUdJOW5xNThWaFF3UUQ5MXBkOUZlOEkzU2VsQmFVWXhEZ21BdC9BL2J4YU5vQW9HQ0NxR1NNNDkKQXdFSG9VUURRZ0FFV1dLeGN5Wk9PVkJoYWdUblFSRDJFWTRwK3NzTkE3Umg5cDhhQzdiWTdpcHlYcXptcWtoZwpkbHRlK2wrbUFldmtYMGMzMTBzb0JYTmFkK3I2WTU4b0dnPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo=`

	// 创建测试集群数据
	testClusters := []k8smodels.Cluster{
		{
			Name:          "本地开发集群",
			Description:   "本地K3s开发测试集群，用于功能测试和开发",
			ApiServer:     "https://127.0.0.1:6443",
			KubeConfig:    testKubeConfig,
			Version:       "v1.28.5+k3s1",
			ImportMethod:  "kubeconfig",
			ImportStatus:  "success",
			ClusterStatus: "unknown", // 初始状态为未知，需要实际健康检查
			Status:        1,
			DeptID:        deptID,
			Remark:        "系统自动初始化的测试集群，可直接用于API测试",
		},
		{
			Name:          "测试集群-1.27",
			Description:   "Kubernetes 1.27版本测试集群",
			ApiServer:     "https://test-k8s-1-27.example.com:6443",
			KubeConfig:    testKubeConfig, // 使用相同的示例配置
			Version:       "v1.27.8",
			ImportMethod:  "kubeconfig",
			ImportStatus:  "failed", // 模拟失败状态
			ClusterStatus: "unhealthy",
			Status:        0, // 禁用状态
			DeptID:        deptID,
			Remark:        "模拟不可访问的集群，用于测试失败场景",
		},
		{
			Name:          "生产集群-示例",
			Description:   "生产环境K8s集群示例（仅供展示）",
			ApiServer:     "https://prod-k8s.example.com:6443",
			KubeConfig:    testKubeConfig,
			Version:       "v1.29.0",
			ImportMethod:  "kubeconfig",
			ImportStatus:  "success",
			ClusterStatus: "healthy",
			Status:        1,
			DeptID:        deptID,
			Remark:        "示例集群，展示完整的字段信息",
		},
		{
			Name:          "待导入集群",
			Description:   "正在导入中的集群",
			ApiServer:     "https://pending-k8s.example.com:6443",
			KubeConfig:    testKubeConfig,
			Version:       "",
			ImportMethod:  "kubeconfig",
			ImportStatus:  "importing", // 模拟导入中状态
			ClusterStatus: "unknown",
			Status:        1,
			DeptID:        deptID,
			Remark:        "模拟导入过程中的集群状态",
		},
	}

	// 批量创建集群
	for i, cluster := range testClusters {
		if err := Db.Create(&cluster).Error; err != nil {
			logger.Log.Warn("创建测试集群失败",
				zap.Int("索引", i),
				zap.String("集群名称", cluster.Name),
				zap.Error(err))
			continue
		}
		logger.Log.Info("创建测试集群成功",
			zap.Uint("ID", cluster.ID),
			zap.String("名称", cluster.Name),
			zap.String("版本", cluster.Version),
			zap.String("导入状态", cluster.ImportStatus),
			zap.String("集群状态", cluster.ClusterStatus))
	}

	logger.Log.Info("K8s 测试集群初始化完成",
		zap.Int("集群数量", len(testClusters)))

	return nil
}
