package service

import (
	"bytes"
	"fmt"
	"time"

	"devops-platform/internal/modules/tool/model"
	"devops-platform/internal/modules/tool/repository"

	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

type ToolService struct {
	repo *repository.ToolRepo
	db   *gorm.DB
}

func NewToolService(db *gorm.DB) *ToolService {
	return &ToolService{repo: repository.NewToolRepo(db), db: db}
}

// SeedBuiltinScripts inserts built-in tool definitions on first run.
func (s *ToolService) SeedBuiltinScripts() error {
	for _, sc := range builtinScripts {
		tool := &model.Tool{
			Name:           sc.Name,
			DisplayName:    sc.DisplayName,
			Category:       sc.Category,
			Description:    sc.Description,
			InstallScript:  sc.InstallCmd,
			CheckScript:    sc.CheckCmd,
			VersionCmd:     sc.VersionCmd,
			DefaultVersion: "latest",
			Dependencies:   "",
		}
		if err := s.repo.Upsert(tool); err != nil {
			return fmt.Errorf("seed %s: %w", sc.Name, err)
		}
	}
	return nil
}

// ListTools returns all available tools, optionally filtered by category.
func (s *ToolService) ListTools(category string) ([]model.Tool, error) {
	return s.repo.ListTools(category)
}

// GetTool returns a single tool by ID.
func (s *ToolService) GetTool(id uint) (*model.Tool, error) {
	return s.repo.GetByID(id)
}

// InstallRequest carries the parameters for tool installation on a host.
type InstallRequest struct {
	ToolID       uint   `json:"toolId"`
	HostID       uint   `json:"hostId"`
	HostIP       string `json:"hostIp"`
	SSHPort      int    `json:"sshPort"`
	SSHUser      string `json:"sshUser"`
	SSHPassword  string `json:"sshPassword"`
	SSHKey       string `json:"sshKey"`
	Version      string `json:"version"`
}

// Install executes the tool's install script on the target host via SSH.
func (s *ToolService) Install(tenantID uint, req InstallRequest) (*model.ToolInstallation, error) {
	tool, err := s.repo.GetByID(req.ToolID)
	if err != nil {
		return nil, fmt.Errorf("tool not found: %w", err)
	}

	if req.SSHPort == 0 {
		req.SSHPort = 22
	}

	inst := &model.ToolInstallation{
		TenantID: tenantID,
		ToolID:   req.ToolID,
		HostID:   req.HostID,
		HostIP:   req.HostIP,
		Status:   model.ToolInstalling,
	}
	if err := s.repo.UpsertInstallation(inst); err != nil {
		return nil, err
	}

	output, err := s.runSSH(req.HostIP, req.SSHPort, req.SSHUser, req.SSHPassword, req.SSHKey, tool.InstallScript)
	now := time.Now()
	inst.Log = output
	if err != nil {
		inst.Status = model.ToolFailed
		inst.Log = fmt.Sprintf("%s\nError: %s", output, err.Error())
	} else {
		inst.Status = model.ToolInstalled
		inst.InstalledAt = &now
		if tool.VersionCmd != "" {
			ver, _ := s.runSSH(req.HostIP, req.SSHPort, req.SSHUser, req.SSHPassword, req.SSHKey, tool.VersionCmd)
			inst.Version = ver
		}
	}
	s.repo.UpsertInstallation(inst)
	return inst, nil
}

// CheckStatus runs the tool's check script on the target host.
func (s *ToolService) CheckStatus(tenantID, toolID, hostID uint, hostIP string, sshPort int, sshUser, sshPassword, sshKey string) (*model.ToolInstallation, error) {
	tool, err := s.repo.GetByID(toolID)
	if err != nil {
		return nil, err
	}
	if sshPort == 0 {
		sshPort = 22
	}
	output, err := s.runSSH(hostIP, sshPort, sshUser, sshPassword, sshKey, tool.CheckScript)
	inst, repoErr := s.repo.GetInstallation(tenantID, toolID, hostID)
	if repoErr != nil {
		inst = &model.ToolInstallation{
			TenantID: tenantID, ToolID: toolID, HostID: hostID, HostIP: hostIP,
		}
	}
	inst.Log = output
	if err != nil {
		inst.Status = model.ToolNotInstalled
	} else {
		inst.Status = model.ToolInstalled
	}
	s.repo.UpsertInstallation(inst)
	return inst, nil
}

// ListInstallations returns installation records for a tenant, optionally filtered by host.
func (s *ToolService) ListInstallations(tenantID, hostID uint) ([]model.ToolInstallation, error) {
	return s.repo.ListInstallations(tenantID, hostID)
}

func (s *ToolService) runSSH(host string, port int, user, password, key, script string) (string, error) {
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}
	if key != "" {
		signer, err := ssh.ParsePrivateKey([]byte(key))
		if err != nil {
			return "", fmt.Errorf("parse SSH key: %w", err)
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if password != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(password)}
	} else {
		return "", fmt.Errorf("no SSH auth provided")
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("SSH dial: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("SSH session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(script); err != nil {
		return stdout.String() + stderr.String(), err
	}
	return stdout.String(), nil
}
