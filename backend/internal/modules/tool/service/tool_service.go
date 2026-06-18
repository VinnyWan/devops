package service

import (
	"bytes"
	"fmt"
	"time"

	"devops-platform/internal/modules/tool/model"
	"devops-platform/internal/modules/tool/repository"
	"devops-platform/internal/pkg/obserr"

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

// --- Templates ---

func (s *ToolService) ListTemplates(category string, page, pageSize int) ([]model.ToolTemplate, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListTemplates(category, page, pageSize)
}

func (s *ToolService) GetTemplate(id uint) (*model.ToolTemplate, error) {
	return s.repo.GetTemplate(id)
}

func (s *ToolService) SaveTemplate(t *model.ToolTemplate) error {
	if t.Name == "" {
		return obserr.New("INVALID_PARAM", "tool/service", "template name is required")
	}
	if t.Category == "" {
		t.Category = "other"
	}
	return s.repo.SaveTemplate(t)
}

func (s *ToolService) DeleteTemplate(id uint) error {
	// Delete associated versions first
	versions, _ := s.repo.ListVersions(id)
	for _, v := range versions {
		s.repo.DeleteVersion(v.ID)
	}
	return s.repo.DeleteTemplate(id)
}

// --- Versions ---

func (s *ToolService) ListVersions(templateID uint) ([]model.ToolTemplateVersion, error) {
	return s.repo.ListVersions(templateID)
}

func (s *ToolService) SaveVersion(v *model.ToolTemplateVersion) error {
	if v.Version == "" {
		return obserr.New("INVALID_PARAM", "tool/service", "version is required")
	}
	if v.InstallScript == "" {
		return obserr.New("INVALID_PARAM", "tool/service", "install script is required")
	}
	return s.repo.SaveVersion(v)
}

func (s *ToolService) DeleteVersion(id uint) error {
	return s.repo.DeleteVersion(id)
}

func (s *ToolService) GetVersion(id uint) (*model.ToolTemplateVersion, error) {
	return s.repo.GetVersion(id)
}

// SeedDefaultTemplates creates the initial set of 15+ templates
func (s *ToolService) SeedDefaultTemplates() error {
	var count int64
	s.db.Model(&model.ToolTemplate{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	templates := []struct {
		t model.ToolTemplate
		v model.ToolTemplateVersion
	}{
		// Database
		{model.ToolTemplate{Name: "MySQL", Category: "database", Description: "MySQL 数据库服务", Icon: "Notebook"},
			model.ToolTemplateVersion{Version: "8.0.36", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root123 mysql:8.0.36`}},
		{model.ToolTemplate{Name: "MySQL", Category: "database", Description: "MySQL 数据库服务", Icon: "Notebook"},
			model.ToolTemplateVersion{Version: "8.4.0", InstallScript: `#!/bin/bash
docker run -d --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root123 mysql:8.4.0`}},
		{model.ToolTemplate{Name: "PostgreSQL", Category: "database", Description: "PostgreSQL 数据库服务", Icon: "Notebook"},
			model.ToolTemplateVersion{Version: "16.2", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres123 postgres:16.2`}},
		{model.ToolTemplate{Name: "MongoDB", Category: "database", Description: "MongoDB 文档数据库", Icon: "Notebook"},
			model.ToolTemplateVersion{Version: "7.0", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name mongodb -p 27017:27017 mongo:7.0`}},
		// Middleware
		{model.ToolTemplate{Name: "Redis", Category: "middleware", Description: "Redis 缓存服务", Icon: "Monitor"},
			model.ToolTemplateVersion{Version: "7.2", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name redis -p 6379:6379 redis:7.2-alpine redis-server --requirepass redis123`}},
		{model.ToolTemplate{Name: "Kafka", Category: "middleware", Description: "Apache Kafka 消息队列", Icon: "Monitor"},
			model.ToolTemplateVersion{Version: "3.7", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name kafka -p 9092:9092 apache/kafka:3.7.0`}},
		{model.ToolTemplate{Name: "Elasticsearch", Category: "middleware", Description: "Elasticsearch 搜索引擎", Icon: "Monitor"},
			model.ToolTemplateVersion{Version: "8.12", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name elasticsearch -p 9200:9200 -e "discovery.type=single-node" elasticsearch:8.12.0`}},
		// Monitoring
		{model.ToolTemplate{Name: "Prometheus", Category: "monitoring", Description: "Prometheus 监控系统", Icon: "Monitor"},
			model.ToolTemplateVersion{Version: "2.51", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name prometheus -p 9090:9090 prom/prometheus:v2.51.0`}},
		{model.ToolTemplate{Name: "Grafana", Category: "monitoring", Description: "Grafana 可视化仪表盘", Icon: "Monitor"},
			model.ToolTemplateVersion{Version: "10.4", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name grafana -p 3000:3000 grafana/grafana:10.4.0`}},
		{model.ToolTemplate{Name: "Node Exporter", Category: "monitoring", Description: "Prometheus Node Exporter", Icon: "Monitor"},
			model.ToolTemplateVersion{Version: "1.7", IsRecommended: true, InstallScript: `#!/bin/bash
wget -qO /tmp/node_exporter.tar.gz https://github.com/prometheus/node_exporter/releases/download/v1.7.0/node_exporter-1.7.0.linux-amd64.tar.gz
tar xzf /tmp/node_exporter.tar.gz -C /usr/local/bin/ --strip-components=1
nohup /usr/local/bin/node_exporter &
echo "Node Exporter started on port 9100"`}},
		// Web
		{model.ToolTemplate{Name: "Nginx", Category: "web", Description: "Nginx Web 服务器", Icon: "Monitor"},
			model.ToolTemplateVersion{Version: "1.26", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name nginx -p 80:80 nginx:1.26-alpine`}},
		{model.ToolTemplate{Name: "Nginx", Category: "web", Description: "Nginx Web 服务器", Icon: "Monitor"},
			model.ToolTemplateVersion{Version: "1.24", InstallScript: `#!/bin/bash
apt-get update && apt-get install -y nginx && systemctl start nginx`}},
		// CI/CD
		{model.ToolTemplate{Name: "Jenkins", Category: "cicd", Description: "Jenkins CI/CD 服务", Icon: "Setting"},
			model.ToolTemplateVersion{Version: "2.452", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name jenkins -p 8080:8080 -p 50000:50000 jenkins/jenkins:lts`}},
		{model.ToolTemplate{Name: "GitLab Runner", Category: "cicd", Description: "GitLab CI Runner", Icon: "Setting"},
			model.ToolTemplateVersion{Version: "16.10", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name gitlab-runner -v /var/run/docker.sock:/var/run/docker.sock gitlab/gitlab-runner:alpine`}},
		// Logging
		{model.ToolTemplate{Name: "Filebeat", Category: "logging", Description: "Elastic Filebeat 日志采集", Icon: "Notebook"},
			model.ToolTemplateVersion{Version: "8.12", IsRecommended: true, InstallScript: `#!/bin/bash
curl -L -O https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-8.12.0-amd64.deb
dpkg -i filebeat-8.12.0-amd64.deb
systemctl enable filebeat
systemctl start filebeat`}},
		{model.ToolTemplate{Name: "Loki", Category: "logging", Description: "Grafana Loki 日志聚合", Icon: "Notebook"},
			model.ToolTemplateVersion{Version: "2.9", IsRecommended: true, InstallScript: `#!/bin/bash
docker run -d --name loki -p 3100:3100 grafana/loki:2.9.0`}},
	}

	for _, item := range templates {
		if err := s.db.Create(&item.t).Error; err != nil {
			return obserr.Wrap("DB_ERROR", "tool/service", "seed template failed", err)
		}
		item.v.TemplateID = item.t.ID
		if err := s.db.Create(&item.v).Error; err != nil {
			return obserr.Wrap("DB_ERROR", "tool/service", "seed version failed", err)
		}
	}
	return nil
}
