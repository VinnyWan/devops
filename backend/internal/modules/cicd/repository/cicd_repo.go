package repository

import (
	"errors"
	"strings"
	"sync"
	"time"

	"devops-platform/internal/modules/cicd/model"
)

type CICDRepo struct {
	mu        sync.RWMutex
	pipelines []model.Pipeline
	logs      []model.PipelineLog
	templates []model.PipelineTemplate
	runs      []model.PipelineRun
	nextRunID uint
	config    model.JenkinsConfig
}

func NewCICDRepo() *CICDRepo {
	now := time.Now()
	return &CICDRepo{
		pipelines: []model.Pipeline{
			{
				ID:        1,
				Name:      "payments-release",
				Status:    "running",
				Branch:    "main",
				LastRunAt: now.Add(-3 * time.Minute),
				CreatedAt: now.Add(-14 * 24 * time.Hour),
				UpdatedAt: now.Add(-3 * time.Minute),
			},
			{
				ID:        2,
				Name:      "gateway-release",
				Status:    "success",
				Branch:    "release/v2.3.0",
				LastRunAt: now.Add(-45 * time.Minute),
				CreatedAt: now.Add(-20 * 24 * time.Hour),
				UpdatedAt: now.Add(-45 * time.Minute),
			},
			{
				ID:        3,
				Name:      "ops-tooling-build",
				Status:    "failed",
				Branch:    "feature/metrics",
				LastRunAt: now.Add(-90 * time.Minute),
				CreatedAt: now.Add(-30 * 24 * time.Hour),
				UpdatedAt: now.Add(-90 * time.Minute),
			},
		},
		logs: []model.PipelineLog{
			{ID: 101, PipelineID: 1, Stage: "build", Level: "info", Message: "开始构建镜像", CreatedAt: now.Add(-3 * time.Minute)},
			{ID: 102, PipelineID: 1, Stage: "deploy", Level: "info", Message: "发布到 staging 命名空间", CreatedAt: now.Add(-2 * time.Minute)},
			{ID: 103, PipelineID: 2, Stage: "test", Level: "info", Message: "单元测试通过", CreatedAt: now.Add(-46 * time.Minute)},
			{ID: 104, PipelineID: 3, Stage: "build", Level: "error", Message: "Docker build 失败: 缺少依赖", CreatedAt: now.Add(-89 * time.Minute)},
		},
		templates: []model.PipelineTemplate{
			{
				ID:          1,
				Name:        "标准发布模板",
				Description: "构建-测试-发布的标准编排",
				Source:      "manual",
				Stages: []model.TemplateStage{
					{Name: "build", Kind: "build", Order: 1, Parameters: map[string]string{"imageRepo": "registry.example.com/devops"}},
					{Name: "test", Kind: "test", Order: 2, Parameters: map[string]string{"suite": "unit,integration"}},
					{Name: "deploy", Kind: "deploy", Order: 3, Parameters: map[string]string{"strategy": "rolling"}},
				},
				CreatedAt: now.Add(-10 * 24 * time.Hour),
				UpdatedAt: now.Add(-2 * time.Hour),
			},
			{
				ID:          2,
				Name:        "快速热修复模板",
				Description: "快速修复分支直达发布",
				Source:      "git_event",
				Stages: []model.TemplateStage{
					{Name: "build", Kind: "build", Order: 1, Parameters: map[string]string{"cache": "true"}},
					{Name: "smoke-test", Kind: "test", Order: 2, Parameters: map[string]string{"suite": "smoke"}},
					{Name: "deploy", Kind: "deploy", Order: 3, Parameters: map[string]string{"strategy": "canary"}},
				},
				CreatedAt: now.Add(-7 * 24 * time.Hour),
				UpdatedAt: now.Add(-4 * time.Hour),
			},
		},
		runs: []model.PipelineRun{
			{
				ID:          1001,
				PipelineID:  1,
				Pipeline:    "payments-release",
				TemplateID:  1,
				Template:    "标准发布模板",
				Branch:      "main",
				Environment: "staging",
				TriggerType: "manual",
				CommitID:    "8b4af17",
				Operator:    "system",
				Status:      "running",
				Stages: []model.TemplateStage{
					{Name: "build", Kind: "build", Order: 1, Parameters: map[string]string{"imageRepo": "registry.example.com/devops"}},
					{Name: "test", Kind: "test", Order: 2, Parameters: map[string]string{"suite": "unit,integration"}},
					{Name: "deploy", Kind: "deploy", Order: 3, Parameters: map[string]string{"strategy": "rolling"}},
				},
				CreatedAt: now.Add(-3 * time.Minute),
			},
		},
		nextRunID: 1002,
		config: model.JenkinsConfig{
			Endpoint:       "http://jenkins.cicd.svc:8080",
			Username:       "admin",
			APIToken:       "token-demo",
			DefaultJob:     "payments-release",
			TimeoutSeconds: 10,
			UpdatedAt:      now,
		},
	}
}

func (r *CICDRepo) ListPipelines() []model.Pipeline {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.Pipeline(nil), r.pipelines...)
}

func (r *CICDRepo) ListLogs(pipelineID uint) []model.PipelineLog {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.PipelineLog, 0, len(r.logs))
	for _, item := range r.logs {
		if pipelineID > 0 && item.PipelineID != pipelineID {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (r *CICDRepo) ListTemplates() []model.PipelineTemplate {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.PipelineTemplate, 0, len(r.templates))
	for _, item := range r.templates {
		items = append(items, cloneTemplate(item))
	}
	return items
}

func (r *CICDRepo) SaveTemplate(template model.PipelineTemplate) model.PipelineTemplate {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	template = cloneTemplate(template)
	template.UpdatedAt = now
	for i, item := range r.templates {
		if item.ID != template.ID || template.ID == 0 {
			continue
		}
		template.CreatedAt = item.CreatedAt
		r.templates[i] = cloneTemplate(template)
		return cloneTemplate(r.templates[i])
	}
	var maxID uint
	for _, item := range r.templates {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	template.ID = maxID + 1
	template.CreatedAt = now
	r.templates = append(r.templates, cloneTemplate(template))
	return cloneTemplate(template)
}

func (r *CICDRepo) GetPipelineByID(id uint) (model.Pipeline, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.pipelines {
		if item.ID == id {
			return item, true
		}
	}
	return model.Pipeline{}, false
}

func (r *CICDRepo) GetTemplateByID(id uint) (model.PipelineTemplate, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.templates {
		if item.ID == id {
			return cloneTemplate(item), true
		}
	}
	return model.PipelineTemplate{}, false
}

func (r *CICDRepo) CreateRun(run model.PipelineRun) model.PipelineRun {
	r.mu.Lock()
	defer r.mu.Unlock()
	run = cloneRun(run)
	if run.ID == 0 {
		run.ID = r.nextRunID
		r.nextRunID++
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = time.Now()
	}
	r.runs = append([]model.PipelineRun{run}, r.runs...)
	return cloneRun(run)
}

func (r *CICDRepo) ListRuns(pipelineID uint, status string) []model.PipelineRun {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.PipelineRun, 0, len(r.runs))
	status = strings.TrimSpace(strings.ToLower(status))
	for _, item := range r.runs {
		if pipelineID > 0 && item.PipelineID != pipelineID {
			continue
		}
		if status != "" && strings.ToLower(item.Status) != status {
			continue
		}
		items = append(items, cloneRun(item))
	}
	return items
}

func cloneTemplate(in model.PipelineTemplate) model.PipelineTemplate {
	stages := make([]model.TemplateStage, 0, len(in.Stages))
	for _, stage := range in.Stages {
		params := make(map[string]string, len(stage.Parameters))
		for k, v := range stage.Parameters {
			params[k] = v
		}
		stages = append(stages, model.TemplateStage{
			Name:       stage.Name,
			Kind:       stage.Kind,
			Order:      stage.Order,
			Parameters: params,
		})
	}
	in.Stages = stages
	return in
}

func cloneRun(in model.PipelineRun) model.PipelineRun {
	stages := make([]model.TemplateStage, 0, len(in.Stages))
	for _, stage := range in.Stages {
		params := make(map[string]string, len(stage.Parameters))
		for k, v := range stage.Parameters {
			params[k] = v
		}
		stages = append(stages, model.TemplateStage{
			Name:       stage.Name,
			Kind:       stage.Kind,
			Order:      stage.Order,
			Parameters: params,
		})
	}
	in.Stages = stages
	return in
}

func (r *CICDRepo) GetConfig() model.JenkinsConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

func (r *CICDRepo) SaveConfig(cfg model.JenkinsConfig) model.JenkinsConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	cfg.UpdatedAt = time.Now()
	r.config = cfg
	return r.config
}

func (r *CICDRepo) ValidateConfigConnection(cfg model.JenkinsConfig) error {
	if strings.Contains(strings.ToLower(cfg.Endpoint), "invalid") {
		return errors.New("jenkins endpoint 不可达")
	}
	if strings.Contains(strings.ToLower(cfg.Endpoint), "timeout") {
		return errors.New("jenkins 请求超时")
	}
	return nil
}
