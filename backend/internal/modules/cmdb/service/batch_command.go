package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
	"devops-platform/internal/modules/cmdb/terminal"
	"devops-platform/internal/pkg/logger"
	"devops-platform/internal/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BatchCommandRequest 批量命令请求
type BatchCommandRequest struct {
	HostIDs []uint `json:"hostIds" binding:"required"`
	Command string `json:"command" binding:"required"`
	Timeout int    `json:"timeout"` // seconds, default 30
}

// HostResult 单个主机的执行结果
type HostResult struct {
	HostID   uint   `json:"hostId"`
	HostName string `json:"hostName"`
	HostIP   string `json:"hostIp"`
	Status   string `json:"status"` // running, success, failed, timeout
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
}

// BatchCommandService 批量命令服务
type BatchCommandService struct {
	db       *gorm.DB
	hostRepo *repository.HostRepo
	credRepo *repository.CredentialRepo
	termSvc  *TerminalService
}

func NewBatchCommandService(db *gorm.DB) *BatchCommandService {
	return &BatchCommandService{
		db:       db,
		hostRepo: repository.NewHostRepo(db),
		credRepo: repository.NewCredentialRepo(db),
		termSvc:  NewTerminalService(db),
	}
}

// ExecuteOnHosts 在多个主机上执行命令，结果通过 channel 流式返回
func (s *BatchCommandService) ExecuteOnHosts(ctx context.Context, tenantID uint, req BatchCommandRequest, resultCh chan<- HostResult) {
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	var wg sync.WaitGroup
	// Limit concurrency to 10
	sem := make(chan struct{}, 10)

	for _, hostID := range req.HostIDs {
		wg.Add(1)
		go func(hid uint) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := s.executeOnSingleHost(ctx, tenantID, hid, req.Command, timeout)
			resultCh <- result
		}(hostID)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()
}

func (s *BatchCommandService) executeOnSingleHost(ctx context.Context, tenantID, hostID uint, command string, timeoutSeconds int) HostResult {
	result := HostResult{
		HostID: hostID,
		Status: "running",
	}

	// Get host
	host, err := s.hostRepo.GetByIDInTenant(tenantID, hostID)
	if err != nil {
		result.Status = "failed"
		result.Error = "主机不存在"
		return result
	}
	result.HostName = host.Hostname
	result.HostIP = host.Ip

	// Get credential
	if host.CredentialID == nil {
		result.Status = "failed"
		result.Error = "主机未绑定凭据"
		return result
	}
	cred, err := s.credRepo.GetByIDInTenant(tenantID, *host.CredentialID)
	if err != nil {
		result.Status = "failed"
		result.Error = "凭据不存在"
		return result
	}

	// Decrypt credential fields
	credPassword, err := utils.Decrypt(cred.Password)
	if err != nil {
		credPassword = ""
	}
	credPrivateKey, err := utils.Decrypt(cred.PrivateKey)
	if err != nil {
		credPrivateKey = ""
	}

	// Create SSH client
	sshClient, err := terminal.NewSSHClient(host.Ip, host.Port, cred.Username, credPassword, credPrivateKey)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("SSH 连接失败: %v", err)
		return result
	}
	defer sshClient.Close()

	// Execute command with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	output, err := terminal.ExecuteCommand(execCtx, sshClient, command)
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Status = "timeout"
			result.Error = "执行超时"
		} else {
			result.Status = "failed"
			result.Error = fmt.Sprintf("执行失败: %v", err)
		}
		result.Output = output
		return result
	}

	result.Status = "success"
	result.Output = output
	return result
}

// CreateBatchAuditRecords 为批量命令创建审计记录
func (s *BatchCommandService) CreateBatchAuditRecords(tenantID, userID uint, username string, req BatchCommandRequest, results []HostResult) {
	for _, r := range results {
		session := model.TerminalSession{
			TenantID:     tenantID,
			UserID:       userID,
			Username:     username,
			HostID:       r.HostID,
			HostIP:       r.HostIP,
			HostName:     r.HostName,
			CredentialID: 0, // batch command doesn't use a specific credential session
			Status:       "closed",
			CloseReason:  "batch_command",
			Duration:     0,
		}
		if err := s.db.Create(&session).Error; err != nil {
			logger.Log.Error("创建批量命令审计记录失败", zap.Error(err))
		}
	}
}
