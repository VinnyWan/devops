package service

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
	cmdbterminal "devops-platform/internal/modules/cmdb/terminal"

	"gorm.io/gorm"
)

type FileService struct {
	auditRepo *repository.FileAuditRepo
	hostRepo  *repository.HostRepo
	credRepo  *repository.CredentialRepo
	db        *gorm.DB
}

func NewFileService(db *gorm.DB) *FileService {
	return &FileService{
		auditRepo: repository.NewFileAuditRepo(db),
		hostRepo:  repository.NewHostRepo(db),
		credRepo:  repository.NewCredentialRepo(db),
		db:        db,
	}
}

func normalizePageFile(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

type BrowseRequest struct {
	HostID uint   `form:"hostId" binding:"required"`
	Path   string `form:"path"`
}

type BrowseResponse struct {
	Entries []cmdbterminal.FileEntry `json:"entries"`
	Path    string                   `json:"path"`
}

func (s *FileService) Browse(tenantID, userID uint, req BrowseRequest) (*BrowseResponse, error) {
	host, credential, err := s.getConnectTarget(tenantID, req.HostID)
	if err != nil {
		return nil, err
	}
	if req.Path == "" {
		req.Path = "/"
	}
	client, err := cmdbterminal.NewSFTPClient(host.Ip, host.Port, credential)
	if err != nil {
		return nil, fmt.Errorf("建立 SFTP 连接失败: %w", err)
	}
	defer client.Close()

	entries, err := client.Browse(req.Path)
	if err != nil {
		return nil, err
	}
	return &BrowseResponse{Entries: entries, Path: req.Path}, nil
}

type DownloadRequest struct {
	HostID   uint   `form:"hostId" binding:"required"`
	FilePath string `form:"path" binding:"required"`
}

func (s *FileService) Download(tenantID, userID, hostID uint, filePath string, writer io.Writer) (int64, *model.Host, error) {
	host, credential, err := s.getConnectTarget(tenantID, hostID)
	if err != nil {
		return 0, nil, err
	}
	client, err := cmdbterminal.NewSFTPClient(host.Ip, host.Port, credential)
	if err != nil {
		return 0, nil, fmt.Errorf("建立 SFTP 连接失败: %w", err)
	}
	defer client.Close()

	n, err := client.Download(filePath, writer)
	if err != nil {
		return 0, host, err
	}
	return n, host, nil
}

type UploadRequest struct {
	HostID     uint   `form:"hostId" binding:"required"`
	RemotePath string `form:"path" binding:"required"`
}

func (s *FileService) Upload(tenantID, userID uint, req UploadRequest, reader io.Reader, size int64) error {
	host, credential, err := s.getConnectTarget(tenantID, req.HostID)
	if err != nil {
		return err
	}
	client, err := cmdbterminal.NewSFTPClient(host.Ip, host.Port, credential)
	if err != nil {
		return fmt.Errorf("建立 SFTP 连接失败: %w", err)
	}
	defer client.Close()

	return client.Upload(req.RemotePath, reader, size)
}

type DeleteRequest struct {
	HostID   uint   `json:"hostId" binding:"required"`
	FilePath string `json:"path" binding:"required"`
}

func (s *FileService) Delete(tenantID, userID, hostID uint, filePath string) error {
	host, credential, err := s.getConnectTarget(tenantID, hostID)
	if err != nil {
		return err
	}
	client, err := cmdbterminal.NewSFTPClient(host.Ip, host.Port, credential)
	if err != nil {
		return fmt.Errorf("建立 SFTP 连接失败: %w", err)
	}
	defer client.Close()

	return client.Delete(filePath)
}

type RenameRequest struct {
	HostID  uint   `json:"hostId" binding:"required"`
	OldPath string `json:"oldPath" binding:"required"`
	NewPath string `json:"newPath" binding:"required"`
}

func (s *FileService) Rename(tenantID, userID uint, req RenameRequest) error {
	host, credential, err := s.getConnectTarget(tenantID, req.HostID)
	if err != nil {
		return err
	}
	client, err := cmdbterminal.NewSFTPClient(host.Ip, host.Port, credential)
	if err != nil {
		return fmt.Errorf("建立 SFTP 连接失败: %w", err)
	}
	defer client.Close()

	return client.Rename(req.OldPath, req.NewPath)
}

type MkdirRequest struct {
	HostID uint   `json:"hostId" binding:"required"`
	Path   string `json:"path" binding:"required"`
}

func (s *FileService) Mkdir(tenantID, userID uint, hostID uint, path string) error {
	host, credential, err := s.getConnectTarget(tenantID, hostID)
	if err != nil {
		return err
	}
	client, err := cmdbterminal.NewSFTPClient(host.Ip, host.Port, credential)
	if err != nil {
		return fmt.Errorf("建立 SFTP 连接失败: %w", err)
	}
	defer client.Close()

	return client.Mkdir(path)
}

type ChmodRequest struct {
	HostID uint   `json:"hostId" binding:"required"`
	Path   string `json:"path" binding:"required"`
	Mode   uint32 `json:"mode" binding:"required"`
}

func (s *FileService) Chmod(tenantID, userID uint, req ChmodRequest) error {
	host, credential, err := s.getConnectTarget(tenantID, req.HostID)
	if err != nil {
		return err
	}
	client, err := cmdbterminal.NewSFTPClient(host.Ip, host.Port, credential)
	if err != nil {
		return fmt.Errorf("建立 SFTP 连接失败: %w", err)
	}
	defer client.Close()

	return client.Chmod(req.Path, os.FileMode(req.Mode))
}

type ReadFileRequest struct {
	HostID   uint   `form:"hostId" binding:"required"`
	FilePath string `form:"path" binding:"required"`
}

const maxPreviewSize = 1 * 1024 * 1024 // 1MB

func (s *FileService) ReadFile(tenantID, userID uint, hostID uint, filePath string) (string, error) {
	host, credential, err := s.getConnectTarget(tenantID, hostID)
	if err != nil {
		return "", err
	}
	client, err := cmdbterminal.NewSFTPClient(host.Ip, host.Port, credential)
	if err != nil {
		return "", fmt.Errorf("建立 SFTP 连接失败: %w", err)
	}
	defer client.Close()

	return client.ReadFile(filePath, maxPreviewSize)
}

type WriteFileRequest struct {
	HostID   uint   `json:"hostId" binding:"required"`
	FilePath string `json:"path" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

func (s *FileService) WriteFile(tenantID, userID uint, req WriteFileRequest) error {
	host, credential, err := s.getConnectTarget(tenantID, req.HostID)
	if err != nil {
		return err
	}
	client, err := cmdbterminal.NewSFTPClient(host.Ip, host.Port, credential)
	if err != nil {
		return fmt.Errorf("建立 SFTP 连接失败: %w", err)
	}
	defer client.Close()

	return client.WriteFile(req.FilePath, req.Content)
}

type DistributeRequest struct {
	HostIDs    []uint `json:"hostIds" binding:"required"`
	RemotePath string `json:"path" binding:"required"`
}

type DistributeResult struct {
	HostID   uint   `json:"hostId"`
	HostIP   string `json:"hostIp"`
	HostName string `json:"hostName"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

func (s *FileService) Distribute(tenantID, userID uint, req DistributeRequest, content []byte) []DistributeResult {
	results := make([]DistributeResult, 0, len(req.HostIDs))
	for _, hostID := range req.HostIDs {
		host, credential, err := s.getConnectTarget(tenantID, hostID)
		if err != nil {
			results = append(results, DistributeResult{HostID: hostID, Error: err.Error()})
			continue
		}
		client, err := cmdbterminal.NewSFTPClient(host.Ip, host.Port, credential)
		if err != nil {
			results = append(results, DistributeResult{
				HostID: hostID, HostIP: host.Ip, HostName: host.Hostname,
				Error: fmt.Sprintf("建立 SFTP 连接失败: %v", err),
			})
			continue
		}
		err = client.Upload(req.RemotePath, bytes.NewReader(content), int64(len(content)))
		client.Close()
		result := DistributeResult{
			HostID: hostID, HostIP: host.Ip, HostName: host.Hostname, Success: err == nil,
		}
		if err != nil {
			result.Error = err.Error()
		}
		results = append(results, result)
	}
	return results
}

type ListAuditRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Keyword  string `form:"keyword"`
	OpType   string `form:"opType"`
	Username string `form:"username"`
	HostIP   string `form:"hostIp"`
	StartAt  string `form:"startAt"`
	EndAt    string `form:"endAt"`
}

func (s *FileService) ListAudit(tenantID uint, req ListAuditRequest) ([]model.FileOperationLog, int64, error) {
	page, pageSize := normalizePageFile(req.Page, req.PageSize)
	return s.auditRepo.ListInTenant(tenantID, page, pageSize, req.Keyword, req.OpType, req.Username, req.HostIP, req.StartAt, req.EndAt)
}

func (s *FileService) RecordAudit(tenantID, userID uint, username, clientIP string, hostID uint, hostIP, hostName, opType, filePath string, fileSize int64, result, errMsg string) {
	log := &model.FileOperationLog{
		TenantID: tenantID, UserID: userID, Username: username,
		HostID: hostID, HostIP: hostIP, HostName: hostName,
		OpType: opType, FilePath: filePath, FileSize: fileSize,
		Result: result, ErrorMsg: errMsg, ClientIP: clientIP,
	}
	if err := s.auditRepo.Create(log); err != nil {
		fmt.Printf("记录文件操作审计日志失败: %v\n", err)
	}
}

// getConnectTarget returns host and credential for a file operation.
func (s *FileService) getConnectTarget(tenantID, hostID uint) (*model.Host, *model.Credential, error) {
	host, err := s.hostRepo.GetByIDInTenant(hostID, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("主机不存在或无权访问")
	}
	if host.CredentialID == nil {
		return nil, nil, fmt.Errorf("主机未绑定凭据")
	}
	cred, err := s.credRepo.GetByIDInTenant(*host.CredentialID, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("凭据不存在或无权访问")
	}
	return host, cred, nil
}
