package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"devops-platform/internal/modules/k8s/model"
	"devops-platform/internal/modules/k8s/repository"
	"devops-platform/internal/pkg/k8s"
	"devops-platform/internal/pkg/utils"

	"gorm.io/gorm"
)

type ClusterService struct {
	repo *repository.ClusterRepo
}

func NewClusterService(db *gorm.DB) *ClusterService {
	return &ClusterService{
		repo: repository.NewClusterRepo(db),
	}
}

// CreateRequest 创建集群请求
type CreateRequest struct {
	Name       string `json:"name" example:"k8s-prod-01"`                                                                        // 集群名称
	AuthType   string `json:"authType" example:"kubeconfig"`                                                                     // "kubeconfig" 或 "token"
	Kubeconfig string `json:"kubeconfig" example:"apiVersion: v1\nclusters:\n- cluster:\n    server: https://1.2.3.4:6443\n..."` // kubeconfig YAML（当 authType=kubeconfig 时）
	Url        string `json:"url" example:"https://1.2.3.4:6443"`                                                                // API Server 地址（当 authType=token 时）
	Token      string `json:"token" example:"eyJhbGciOiJSUzI1NiIsImtpZCI6In..."`                                                 // Token（当 authType=token 时）
	CaData     string `json:"caData" example:"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t..."`                                          // CA 证书（当 authType=token 时，可选）
	Remark     string `json:"remark" example:"生产环境核心集群"`                                                                         // 备注
	Labels     string `json:"labels" example:"{\"region\":\"shanghai\",\"dept\":\"it\"}"`                                        // 标签 (JSON 格式)
	Env        string `json:"env" example:"prod"`                                                                                // 环境 (dev, test, prod)
}

// UpdateRequest 更新集群请求
type UpdateRequest struct {
	ID         uint   `json:"id" example:"1"`                                             // 集群ID
	Name       string `json:"name" example:"k8s-prod-01-new"`                             // 集群名称
	Kubeconfig string `json:"kubeconfig"`                                                 // 新的 kubeconfig
	Url        string `json:"url" example:"https://1.2.3.4:6443"`                         // 新的 URL
	Token      string `json:"token" example:"eyJhbGciOiJSUzI1NiIsImtpZCI6In..."`          // 新的 Token
	CaData     string `json:"caData" example:"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t..."`   // 新的 CA
	Remark     string `json:"remark" example:"生产环境核心集群-已迁移"`                              // 备注
	Labels     string `json:"labels" example:"{\"region\":\"shanghai\",\"dept\":\"it\"}"` // 标签
	Env        string `json:"env" example:"prod"`                                         // 环境
}

// Create 创建集群
func (s *ClusterService) Create(req *CreateRequest) (*model.Cluster, error) {
	return s.CreateInTenant(0, req)
}

// CreateInTenant 在指定租户下创建集群
func (s *ClusterService) CreateInTenant(tenantID uint, req *CreateRequest) (*model.Cluster, error) {
	// 1. 参数校验
	if req.Name == "" {
		return nil, errors.New("集群名称不能为空")
	}
	if req.AuthType != "kubeconfig" && req.AuthType != "token" {
		return nil, errors.New("authType 必须是 kubeconfig 或 token")
	}

	// 检查集群名称是否已存在
	existing, err := s.repo.GetByExactNameInTenant(tenantID, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("集群名称 %s 已存在", req.Name)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询集群名称失败: %w", err)
	}

	var server, caData, certData, keyData, token string

	// 2. 解析 kubeconfig 或 token
	if req.AuthType == "kubeconfig" {
		if req.Kubeconfig == "" {
			return nil, errors.New("kubeconfig 不能为空")
		}

		// 解析 kubeconfig
		kubeconfigData, err := utils.ParseKubeconfig(req.Kubeconfig)
		if err != nil {
			return nil, err // 直接返回 ParseKubeconfig 的详细错误
		}

		server = kubeconfigData.Server
		caData = kubeconfigData.CaData
		if kubeconfigData.AuthType == "cert" {
			certData = kubeconfigData.CertData
			keyData = kubeconfigData.KeyData
		} else {
			token = kubeconfigData.Token
		}
	} else {
		// Token 模式
		if req.Url == "" {
			return nil, errors.New("API Server 地址不能为空")
		}
		if req.Token == "" {
			return nil, errors.New("Token 不能为空")
		}

		// 定义通用清洗函数
		cleanStr := func(s string) string {
			s = strings.TrimSpace(s)
			s = strings.Trim(s, "`")
			s = strings.ReplaceAll(s, "\\n", "")
			s = strings.ReplaceAll(s, "\\r", "")
			s = strings.TrimRight(s, "\n\r")
			return strings.TrimSpace(s)
		}

		server = cleanStr(req.Url)
		token = cleanStr(req.Token)
		caData = cleanStr(req.CaData)

		// 进一步处理 server 地址中可能的拼接
		if idx := strings.Index(server, " "); idx != -1 {
			server = server[:idx]
		}

		// 验证 base64 编码
		if err := utils.ValidateBase64(caData); err != nil {
			return nil, fmt.Errorf("CA 证书格式错误: %w", err)
		}
	}

	// 3. 尝试连接 k8s（健康校验）
	clientCfg := &k8s.ClientConfig{
		Server:   server,
		CaData:   caData,
		CertData: certData,
		KeyData:  keyData,
		Token:    token,
	}

	client, err := k8s.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("创建 K8s 客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.HealthCheck(ctx)
	if err != nil {
		return nil, fmt.Errorf("K8s 集群连接失败: %w", err)
	}

	// 4. 加密敏感信息
	var encryptedKubeconfig, encryptedToken, encryptedCaData string

	if req.AuthType == "kubeconfig" {
		// 重新构建 kubeconfig 并加密
		kubeconfigYAML := utils.BuildKubeconfig(server, caData, certData, keyData, token)
		encryptedKubeconfig, err = utils.Encrypt(kubeconfigYAML)
		if err != nil {
			return nil, fmt.Errorf("加密 kubeconfig 失败: %w", err)
		}
	} else {
		// 加密 token
		encryptedToken, err = utils.Encrypt(token)
		if err != nil {
			return nil, fmt.Errorf("加密 token 失败: %w", err)
		}

		// 加密 CA（如果有）
		if caData != "" {
			encryptedCaData, err = utils.Encrypt(caData)
			if err != nil {
				return nil, fmt.Errorf("加密 CA 证书失败: %w", err)
			}
		}
	}

	// 5. 数据入库
	cluster := &model.Cluster{
		TenantID:   nil,
		Name:       req.Name,
		Url:        server,
		AuthType:   req.AuthType,
		Kubeconfig: encryptedKubeconfig,
		Token:      encryptedToken,
		CaData:     encryptedCaData,
		Status:     "healthy",
		Remark:     req.Remark,
		Labels:     req.Labels,
		Env:        req.Env,
	}

	err = s.repo.CreateInTenant(tenantID, cluster)
	if err != nil {
		return nil, fmt.Errorf("保存集群失败: %w", err)
	}

	_, derr := s.repo.GetDefaultInTenant(tenantID)
	if derr != nil {
		if errors.Is(derr, gorm.ErrRecordNotFound) {
			if err := s.repo.SetDefaultInTenant(tenantID, cluster.ID); err != nil {
				return nil, fmt.Errorf("设置默认集群失败: %w", err)
			}
			cluster.IsDefault = true
		} else {
			return nil, fmt.Errorf("查询默认集群失败: %w", derr)
		}
	}

	return cluster, nil
}

// GetByID 根据ID获取集群
func (s *ClusterService) GetByID(id uint) (*model.Cluster, error) {
	return s.GetByIDInTenant(0, id)
}

func (s *ClusterService) GetByIDInTenant(tenantID uint, id uint) (*model.Cluster, error) {
	return s.repo.GetByIDInTenant(tenantID, id)
}

// GetByName 根据Name获取集群
func (s *ClusterService) GetByName(name string) ([]model.Cluster, error) {
	return s.GetByNameInTenant(0, name)
}

func (s *ClusterService) GetByNameInTenant(tenantID uint, name string) ([]model.Cluster, error) {
	return s.repo.GetByNameInTenant(tenantID, name)
}

func (s *ClusterService) GetByExactName(name string) (*model.Cluster, error) {
	return s.GetByExactNameInTenant(0, name)
}

func (s *ClusterService) GetByExactNameInTenant(tenantID uint, name string) (*model.Cluster, error) {
	return s.repo.GetByExactNameInTenant(tenantID, name)
}

// GetByEnv 根据Env获取集群
func (s *ClusterService) GetByEnv(env string) (*model.Cluster, error) {
	return s.GetByEnvInTenant(0, env)
}

func (s *ClusterService) GetByEnvInTenant(tenantID uint, env string) (*model.Cluster, error) {
	return s.repo.GetByEnvInTenant(tenantID, env)
}

func (s *ClusterService) GetDefault() (*model.Cluster, error) {
	return s.GetDefaultInTenant(0)
}

func (s *ClusterService) GetDefaultInTenant(tenantID uint) (*model.Cluster, error) {
	return s.repo.GetDefaultInTenant(tenantID)
}

func (s *ClusterService) GetDefaultOrFirst() (*model.Cluster, error) {
	return s.GetDefaultOrFirstInTenant(0)
}

func (s *ClusterService) GetDefaultOrFirstInTenant(tenantID uint) (*model.Cluster, error) {
	cluster, err := s.repo.GetDefaultInTenant(tenantID)
	if err == nil {
		return cluster, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	clusters, _, err := s.repo.ListInTenant(tenantID, 1, 1, "", "")
	if err != nil {
		return nil, err
	}
	if len(clusters) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &clusters[0], nil
}

func (s *ClusterService) SetDefault(id uint) error {
	return s.SetDefaultInTenant(0, id)
}

func (s *ClusterService) SetDefaultInTenant(tenantID, id uint) error {
	if _, err := s.repo.GetByIDInTenant(tenantID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("集群不存在")
		}
		return err
	}
	return s.repo.SetDefaultInTenant(tenantID, id)
}

// List 获取集群列表
func (s *ClusterService) List(page, pageSize int, env, keyword string) ([]model.Cluster, int64, error) {
	return s.ListInTenant(0, page, pageSize, env, keyword)
}

func (s *ClusterService) ListInTenant(tenantID uint, page, pageSize int, env, keyword string) ([]model.Cluster, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return s.repo.ListInTenant(tenantID, page, pageSize, env, keyword)
}

// Update 更新集群
func (s *ClusterService) Update(req *UpdateRequest) (*model.Cluster, error) {
	return s.UpdateInTenant(0, req)
}

func (s *ClusterService) UpdateInTenant(tenantID uint, req *UpdateRequest) (*model.Cluster, error) {
	// 1. 获取现有集群
	cluster, err := s.repo.GetByIDInTenant(tenantID, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("集群不存在")
		}
		return nil, err
	}

	// 2. 更新基本信息
	if req.Name != "" {
		// 检查名称是否与其他集群冲突
		existing, err := s.repo.GetByExactNameInTenant(tenantID, req.Name)
		if err == nil && existing != nil && existing.ID != cluster.ID {
			return nil, fmt.Errorf("集群名称 %s 已被使用", req.Name)
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("查询集群名称失败: %w", err)
		}
		cluster.Name = req.Name
	}

	cluster.Remark = req.Remark
	cluster.Labels = req.Labels
	if req.Env != "" {
		cluster.Env = req.Env
	}

	// 3. 更新认证信息（如果提供）
	needReconnect := false
	var newServer, newCaData, newCertData, newKeyData, newToken string

	if cluster.AuthType == "kubeconfig" && req.Kubeconfig != "" {
		// 更新 kubeconfig
		kubeconfigData, err := utils.ParseKubeconfig(req.Kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("解析新的 kubeconfig 失败: %w", err)
		}

		newServer = kubeconfigData.Server
		newCaData = kubeconfigData.CaData
		if kubeconfigData.AuthType == "cert" {
			newCertData = kubeconfigData.CertData
			newKeyData = kubeconfigData.KeyData
		} else {
			newToken = kubeconfigData.Token
		}
		needReconnect = true
	} else if cluster.AuthType == "token" && (req.Token != "" || req.Url != "") {
		// 更新 token 模式
		if req.Url != "" {
			newServer = req.Url
		} else {
			newServer = cluster.Url
		}

		if req.Token != "" {
			newToken = req.Token
		} else {
			// 需要解密旧 token
			decryptedToken, err := utils.Decrypt(cluster.Token)
			if err != nil {
				return nil, fmt.Errorf("解密旧 token 失败: %w", err)
			}
			newToken = decryptedToken
		}

		if req.CaData != "" {
			newCaData = req.CaData
			if err := utils.ValidateBase64(newCaData); err != nil {
				return nil, fmt.Errorf("新的 CA 证书格式错误: %w", err)
			}
		} else if cluster.CaData != "" {
			// 解密旧 CA
			decryptedCa, err := utils.Decrypt(cluster.CaData)
			if err == nil {
				newCaData = decryptedCa
			}
		}
		needReconnect = true
	}

	// 4. 如果需要重新连接，进行健康检查
	if needReconnect {
		clientCfg := &k8s.ClientConfig{
			Server:   newServer,
			CaData:   newCaData,
			CertData: newCertData,
			KeyData:  newKeyData,
			Token:    newToken,
		}

		client, err := k8s.NewClient(clientCfg)
		if err != nil {
			return nil, fmt.Errorf("创建 K8s 客户端失败: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.HealthCheck(ctx)
		if err != nil {
			return nil, fmt.Errorf("新配置连接 K8s 失败: %w", err)
		}

		// 同步集群信息
		version, _ := client.GetServerVersion(ctx)
		nodeCount, _ := client.GetNodeCount(ctx)
		cluster.K8sVersion = version
		cluster.NodeCount = nodeCount

		// 5. 加密并更新
		if cluster.AuthType == "kubeconfig" {
			kubeconfigYAML := utils.BuildKubeconfig(newServer, newCaData, newCertData, newKeyData, newToken)
			encryptedKubeconfig, err := utils.Encrypt(kubeconfigYAML)
			if err != nil {
				return nil, fmt.Errorf("加密新 kubeconfig 失败: %w", err)
			}
			cluster.Kubeconfig = encryptedKubeconfig
			cluster.Url = newServer
		} else {
			encryptedToken, err := utils.Encrypt(newToken)
			if err != nil {
				return nil, fmt.Errorf("加密新 token 失败: %w", err)
			}
			cluster.Token = encryptedToken
			cluster.Url = newServer

			if newCaData != "" {
				encryptedCa, err := utils.Encrypt(newCaData)
				if err != nil {
					return nil, fmt.Errorf("加密新 CA 失败: %w", err)
				}
				cluster.CaData = encryptedCa
			}
		}

		cluster.Status = "healthy"
	}

	// 6. 保存更新
	err = s.repo.UpdateInTenant(tenantID, cluster)
	if err != nil {
		return nil, fmt.Errorf("更新集群失败: %w", err)
	}

	return cluster, nil
}

// Delete 删除集群
func (s *ClusterService) Delete(id uint) error {
	return s.DeleteInTenant(0, id)
}

func (s *ClusterService) DeleteInTenant(tenantID uint, id uint) error {
	// 检查集群是否存在
	_, err := s.repo.GetByIDInTenant(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("集群不存在")
		}
		return err
	}

	return s.repo.DeleteInTenant(tenantID, id)
}

// HealthCheck 健康检查
func (s *ClusterService) HealthCheck(id uint) (string, error) {
	return s.HealthCheckInTenant(0, id)
}

func (s *ClusterService) HealthCheckInTenant(tenantID uint, id uint) (string, error) {
	cluster, err := s.repo.GetByIDInTenant(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("集群不存在")
		}
		return "", err
	}

	// 解密认证信息
	var server, caData, certData, keyData, token string
	server = cluster.Url

	if cluster.AuthType == "kubeconfig" {
		decryptedKubeconfig, err := utils.Decrypt(cluster.Kubeconfig)
		if err != nil {
			s.repo.UpdateStatusInTenant(tenantID, id, "unhealthy")
			return "unhealthy", fmt.Errorf("解密 kubeconfig 失败: %w", err)
		}

		kubeconfigData, err := utils.ParseKubeconfig(decryptedKubeconfig)
		if err != nil {
			s.repo.UpdateStatusInTenant(tenantID, id, "unhealthy")
			return "unhealthy", fmt.Errorf("解析 kubeconfig 失败: %w", err)
		}

		server = kubeconfigData.Server
		caData = kubeconfigData.CaData
		if kubeconfigData.AuthType == "cert" {
			certData = kubeconfigData.CertData
			keyData = kubeconfigData.KeyData
		} else {
			token = kubeconfigData.Token
		}
	} else {
		decryptedToken, err := utils.Decrypt(cluster.Token)
		if err != nil {
			s.repo.UpdateStatusInTenant(tenantID, id, "unhealthy")
			return "unhealthy", fmt.Errorf("解密 token 失败: %w", err)
		}
		token = decryptedToken

		if cluster.CaData != "" {
			decryptedCa, err := utils.Decrypt(cluster.CaData)
			if err == nil {
				caData = decryptedCa
			}
		}
	}

	// 创建客户端并测试连接
	clientCfg := &k8s.ClientConfig{
		Server:   server,
		CaData:   caData,
		CertData: certData,
		KeyData:  keyData,
		Token:    token,
	}

	client, err := k8s.NewClient(clientCfg)
	if err != nil {
		s.repo.UpdateStatusInTenant(tenantID, id, "unhealthy")
		return "unhealthy", fmt.Errorf("创建客户端失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.HealthCheck(ctx)
	if err != nil {
		s.repo.UpdateStatusInTenant(tenantID, id, "unhealthy")
		return "unhealthy", err
	}

	// 同步集群信息
	version, _ := client.GetServerVersion(ctx)
	nodeCount, _ := client.GetNodeCount(ctx)

	// 更新状态为健康并保存信息
	cluster.Status = "healthy"
	cluster.K8sVersion = version
	cluster.NodeCount = nodeCount
	if err := s.repo.UpdateInTenant(tenantID, cluster); err != nil {
		return "healthy", fmt.Errorf("更新集群信息失败: %w", err)
	}

	return "healthy", nil
}

func (s *ClusterService) Search(
	name string,
	env string,
	page int,
	pageSize int,
) ([]model.Cluster, int64, error) {
	return s.SearchInTenant(0, name, env, page, pageSize)
}

func (s *ClusterService) SearchInTenant(
	tenantID uint,
	name string,
	env string,
	page int,
	pageSize int,
) ([]model.Cluster, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return s.repo.SearchInTenant(tenantID, name, env, page, pageSize)
}
