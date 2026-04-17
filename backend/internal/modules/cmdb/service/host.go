package service

import (
	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
	cmdbterminal "devops-platform/internal/modules/cmdb/terminal"
	"devops-platform/internal/pkg/utils"

	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
	"net"
	"strconv"
	"time"
)

type HostService struct {
	repo *repository.HostRepo
}

func NewHostService(db *gorm.DB) *HostService {
	return &HostService{repo: repository.NewHostRepo(db)}
}

type HostCreateRequest struct {
	Hostname     string `json:"hostname" binding:"required"`
	Ip           string `json:"ip" binding:"required"`
	Port         int    `json:"port"`
	OsType       string `json:"osType"`
	OsName       string `json:"osName"`
	CpuCores     int    `json:"cpuCores"`
	MemoryTotal  int    `json:"memoryTotal"`
	DiskTotal    int    `json:"diskTotal"`
	CredentialID uint   `json:"credentialId"`
	GroupID      uint   `json:"groupId"`
	Labels       string `json:"labels"`
	Description  string `json:"description"`
}

type HostUpdateRequest struct {
	ID           uint   `json:"id" binding:"required"`
	Hostname     string `json:"hostname"`
	Ip           string `json:"ip"`
	Port         int    `json:"port"`
	OsType       string `json:"osType"`
	OsName       string `json:"osName"`
	CpuCores     int    `json:"cpuCores"`
	MemoryTotal  int    `json:"memoryTotal"`
	DiskTotal    int    `json:"diskTotal"`
	CredentialID uint   `json:"credentialId"`
	GroupID      uint   `json:"groupId"`
	Labels       string `json:"labels"`
	Description  string `json:"description"`
}

func (s *HostService) normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func (s *HostService) ListInTenant(tenantID uint, page, pageSize int, groupID uint, status, keyword string) ([]model.Host, int64, error) {
	page, pageSize = s.normalizePage(page, pageSize)
	return s.repo.ListInTenant(tenantID, page, pageSize, groupID, status, keyword)
}

func (s *HostService) GetByIDInTenant(tenantID uint, id uint) (*model.Host, error) {
	return s.repo.GetByIDInTenant(tenantID, id)
}

func (s *HostService) CreateInTenant(tenantID uint, req *HostCreateRequest) (*model.Host, error) {
	port := req.Port
	if port == 0 {
		port = 22
	}

	host := &model.Host{
		Hostname:    req.Hostname,
		Ip:          req.Ip,
		Port:        port,
		OsType:      req.OsType,
		OsName:      req.OsName,
		CpuCores:    req.CpuCores,
		MemoryTotal: req.MemoryTotal,
		DiskTotal:   req.DiskTotal,
		Labels:      req.Labels,
		Description: req.Description,
		Status:      "unknown",
	}
	if req.CredentialID > 0 {
		host.CredentialID = &req.CredentialID
	}
	if req.GroupID > 0 {
		host.GroupID = &req.GroupID
	}

	if err := s.repo.CreateInTenant(tenantID, host); err != nil {
		return nil, err
	}
	return host, nil
}

func (s *HostService) UpdateInTenant(tenantID uint, req *HostUpdateRequest) (*model.Host, error) {
	host, err := s.repo.GetByIDInTenant(tenantID, req.ID)
	if err != nil {
		return nil, err
	}

	if req.Hostname != "" {
		host.Hostname = req.Hostname
	}
	if req.Ip != "" {
		host.Ip = req.Ip
	}
	if req.Port > 0 {
		host.Port = req.Port
	}
	if req.OsType != "" {
		host.OsType = req.OsType
	}
	if req.OsName != "" {
		host.OsName = req.OsName
	}
	host.CpuCores = req.CpuCores
	host.MemoryTotal = req.MemoryTotal
	host.DiskTotal = req.DiskTotal
	if req.CredentialID > 0 {
		host.CredentialID = &req.CredentialID
	} else {
		host.CredentialID = nil
	}
	if req.GroupID > 0 {
		host.GroupID = &req.GroupID
	} else {
		host.GroupID = nil
	}
	if req.Labels != "" {
		host.Labels = req.Labels
	}
	if req.Description != "" {
		host.Description = req.Description
	}

	if err := s.repo.UpdateInTenant(tenantID, host); err != nil {
		return nil, err
	}
	return host, nil
}

func (s *HostService) DeleteInTenant(tenantID uint, id uint) error {
	return s.repo.DeleteInTenant(tenantID, id)
}

func (s *HostService) BatchCreateInTenant(tenantID uint, reqs []HostCreateRequest) ([]model.Host, error) {
	hosts := make([]model.Host, 0, len(reqs))
	for _, req := range reqs {
		port := req.Port
		if port == 0 {
			port = 22
		}
		host := model.Host{
			Hostname:    req.Hostname,
			Ip:          req.Ip,
			Port:        port,
			OsType:      req.OsType,
			OsName:      req.OsName,
			CpuCores:    req.CpuCores,
			MemoryTotal: req.MemoryTotal,
			DiskTotal:   req.DiskTotal,
			Labels:      req.Labels,
			Description: req.Description,
			Status:      "unknown",
		}
		if req.CredentialID > 0 {
			host.CredentialID = &req.CredentialID
		}
		if req.GroupID > 0 {
			host.GroupID = &req.GroupID
		}
		hosts = append(hosts, host)
	}

	if err := s.repo.BatchCreateInTenant(tenantID, hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

// TestConnection 测试主机 SSH 连接（使用指定凭据或已关联凭据）
func (s *HostService) TestConnection(ip string, port int, cred *model.Credential) (bool, string) {
	if port == 0 {
		port = 22
	}

	var authMethods []ssh.AuthMethod

	if cred.Type == "password" {
		password, err := utils.Decrypt(cred.Password)
		if err != nil {
			return false, "凭据解密失败: " + err.Error()
		}
		authMethods = append(authMethods, ssh.Password(password))
	} else if cred.Type == "key" {
		privateKey, err := utils.Decrypt(cred.PrivateKey)
		if err != nil {
			return false, "凭据解密失败: " + err.Error()
		}
		var signer ssh.Signer
		if cred.Passphrase != "" {
			passphrase, err := utils.Decrypt(cred.Passphrase)
			if err != nil {
				return false, "密钥密码解密失败: " + err.Error()
			}
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(passphrase))
			if err != nil {
				return false, "密钥解析失败: " + err.Error()
			}
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(privateKey))
			if err != nil {
				return false, "密钥解析失败: " + err.Error()
			}
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	hostKeyCallback, err := cmdbterminal.BuildHostKeyCallback()
	if err != nil {
		return false, "加载 known_hosts 失败: " + err.Error()
	}
	config := &ssh.ClientConfig{
		User:            cred.Username,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, strconv.Itoa(port)), 10*time.Second)
	if err != nil {
		return false, "连接失败: " + err.Error()
	}
	defer conn.Close()

	sshConn, _, _, err := ssh.NewClientConn(conn, net.JoinHostPort(ip, strconv.Itoa(port)), config)
	if err != nil {
		return false, "SSH 认证失败: " + err.Error()
	}
	defer sshConn.Close()

	return true, "连接成功"
}

type HostStats struct {
	Total    int64            `json:"total"`
	ByStatus map[string]int64 `json:"byStatus"`
	ByGroup  map[uint]int64   `json:"byGroup"`
}

func (s *HostService) StatsInTenant(tenantID uint) (*HostStats, error) {
	total, err := s.repo.CountInTenant(tenantID)
	if err != nil {
		return nil, err
	}

	byStatus, err := s.repo.CountByStatusInTenant(tenantID)
	if err != nil {
		return nil, err
	}

	byGroup, err := s.repo.CountByGroupInTenant(tenantID)
	if err != nil {
		return nil, err
	}

	return &HostStats{
		Total:    total,
		ByStatus: byStatus,
		ByGroup:  byGroup,
	}, nil
}
