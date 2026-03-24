package service

import (
	"sort"
	"strings"
	"time"

	"devops-platform/internal/modules/cicd/model"
	"devops-platform/internal/modules/cicd/repository"
	"devops-platform/internal/pkg/obserr"
	queryutil "devops-platform/internal/pkg/query"
)

type CICDService struct {
	repo *repository.CICDRepo
}

type ListPipelineStatusResponse struct {
	Total int              `json:"total"`
	Items []model.Pipeline `json:"items"`
}

type ListPipelineLogsResponse struct {
	Total int                 `json:"total"`
	Items []model.PipelineLog `json:"items"`
}

type ListTemplateResponse struct {
	Total int                      `json:"total"`
	Items []model.PipelineTemplate `json:"items"`
}

type SaveTemplateRequest struct {
	ID          uint                  `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Source      string                `json:"source"`
	Stages      []model.TemplateStage `json:"stages"`
}

type TriggerPipelineRequest struct {
	PipelineID  uint              `json:"pipelineId"`
	TemplateID  uint              `json:"templateId"`
	Branch      string            `json:"branch"`
	Environment string            `json:"environment"`
	TriggerType string            `json:"triggerType"`
	CommitID    string            `json:"commitId"`
	Operator    string            `json:"operator"`
	Parameters  map[string]string `json:"parameters"`
}

type ListPipelineRunsResponse struct {
	Total int                 `json:"total"`
	Items []model.PipelineRun `json:"items"`
}

type SaveJenkinsConfigRequest struct {
	Endpoint              string `json:"endpoint"`
	Username              string `json:"username"`
	APIToken              string `json:"apiToken"`
	DefaultJob            string `json:"defaultJob"`
	TimeoutSeconds        int    `json:"timeoutSeconds"`
	TLSInsecureSkipVerify bool   `json:"tlsInsecureSkipVerify"`
}

func NewCICDService() *CICDService {
	return &CICDService{repo: repository.NewCICDRepo()}
}

func (s *CICDService) ListPipelineStatus(status, keyword string) ListPipelineStatusResponse {
	status = strings.TrimSpace(strings.ToLower(status))
	pipelines := s.repo.ListPipelines()
	items := make([]model.Pipeline, 0, len(pipelines))
	for _, pipeline := range pipelines {
		if status != "" && strings.ToLower(pipeline.Status) != status {
			continue
		}
		if !queryutil.MatchKeywordAny(keyword, pipeline.Name, pipeline.Branch, pipeline.Status) {
			continue
		}
		items = append(items, pipeline)
	}
	return ListPipelineStatusResponse{Total: len(items), Items: items}
}

func (s *CICDService) ListPipelineLogs(pipelineID uint, stage string, limit int) ListPipelineLogsResponse {
	stage = strings.TrimSpace(strings.ToLower(stage))
	if limit <= 0 {
		limit = 100
	}
	logs := s.repo.ListLogs(pipelineID)
	items := make([]model.PipelineLog, 0, len(logs))
	for _, entry := range logs {
		if stage != "" && strings.ToLower(entry.Stage) != stage {
			continue
		}
		items = append(items, entry)
		if len(items) == limit {
			break
		}
	}
	return ListPipelineLogsResponse{Total: len(items), Items: items}
}

func (s *CICDService) ListTemplates(keyword string) ListTemplateResponse {
	templates := s.repo.ListTemplates()
	items := make([]model.PipelineTemplate, 0, len(templates))
	for _, item := range templates {
		if !queryutil.MatchKeywordAny(keyword, item.Name, item.Description, item.Source) {
			continue
		}
		items = append(items, item)
	}
	return ListTemplateResponse{Total: len(items), Items: items}
}

func (s *CICDService) SaveTemplate(req SaveTemplateRequest) (model.PipelineTemplate, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return model.PipelineTemplate{}, obserr.New("JENKINS_TEMPLATE_NAME_REQUIRED", "cicd.SaveTemplate", "模板名称不能为空")
	}
	if len(req.Stages) == 0 {
		return model.PipelineTemplate{}, obserr.New("JENKINS_TEMPLATE_STAGES_REQUIRED", "cicd.SaveTemplate", "模板至少包含一个阶段")
	}
	stages := make([]model.TemplateStage, 0, len(req.Stages))
	for idx, stage := range req.Stages {
		stageName := strings.TrimSpace(stage.Name)
		stageKind := strings.TrimSpace(strings.ToLower(stage.Kind))
		if stageName == "" || stageKind == "" {
			return model.PipelineTemplate{}, obserr.New("JENKINS_TEMPLATE_STAGE_INVALID", "cicd.SaveTemplate", "阶段名称和类型不能为空")
		}
		if stage.Order <= 0 {
			stage.Order = idx + 1
		}
		params := make(map[string]string, len(stage.Parameters))
		for k, v := range stage.Parameters {
			key := strings.TrimSpace(k)
			if key == "" {
				continue
			}
			params[key] = strings.TrimSpace(v)
		}
		stages = append(stages, model.TemplateStage{
			Name:       stageName,
			Kind:       stageKind,
			Order:      stage.Order,
			Parameters: params,
		})
	}
	sort.Slice(stages, func(i, j int) bool {
		if stages[i].Order == stages[j].Order {
			return stages[i].Name < stages[j].Name
		}
		return stages[i].Order < stages[j].Order
	})
	template := s.repo.SaveTemplate(model.PipelineTemplate{
		ID:          req.ID,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Source:      strings.TrimSpace(req.Source),
		Stages:      stages,
	})
	return template, nil
}

func (s *CICDService) PreviewOrchestration(pipelineID, templateID uint, environment string, parameters map[string]string) (model.PipelineRun, error) {
	if err := s.ValidateCurrentConfig(); err != nil {
		return model.PipelineRun{}, err
	}
	pipeline, ok := s.repo.GetPipelineByID(pipelineID)
	if !ok {
		return model.PipelineRun{}, obserr.New("JENKINS_PIPELINE_NOT_FOUND", "cicd.PreviewOrchestration", "流水线不存在")
	}
	template, ok := s.repo.GetTemplateByID(templateID)
	if !ok {
		return model.PipelineRun{}, obserr.New("JENKINS_TEMPLATE_NOT_FOUND", "cicd.PreviewOrchestration", "模板不存在")
	}
	branch := pipeline.Branch
	if branch == "" {
		branch = "main"
	}
	if strings.TrimSpace(environment) == "" {
		environment = "staging"
	}
	stages := mergeTemplateStages(template.Stages, parameters)
	return model.PipelineRun{
		PipelineID:  pipeline.ID,
		Pipeline:    pipeline.Name,
		TemplateID:  template.ID,
		Template:    template.Name,
		Branch:      branch,
		Environment: strings.TrimSpace(environment),
		TriggerType: "preview",
		Status:      "planned",
		Stages:      stages,
	}, nil
}

func (s *CICDService) TriggerPipeline(req TriggerPipelineRequest) (model.PipelineRun, error) {
	if err := s.ValidateCurrentConfig(); err != nil {
		return model.PipelineRun{}, err
	}
	pipeline, ok := s.repo.GetPipelineByID(req.PipelineID)
	if !ok {
		return model.PipelineRun{}, obserr.New("JENKINS_PIPELINE_NOT_FOUND", "cicd.TriggerPipeline", "流水线不存在")
	}
	template, ok := s.repo.GetTemplateByID(req.TemplateID)
	if !ok {
		return model.PipelineRun{}, obserr.New("JENKINS_TEMPLATE_NOT_FOUND", "cicd.TriggerPipeline", "模板不存在")
	}
	branch := strings.TrimSpace(req.Branch)
	if branch == "" {
		branch = pipeline.Branch
	}
	if branch == "" {
		branch = "main"
	}
	triggerType := strings.TrimSpace(strings.ToLower(req.TriggerType))
	if triggerType == "" {
		triggerType = "manual"
	}
	environment := strings.TrimSpace(req.Environment)
	if environment == "" {
		environment = "staging"
	}
	operator := strings.TrimSpace(req.Operator)
	if operator == "" {
		operator = "system"
	}
	status := "running"
	if triggerType == "git_event" {
		status = "queued"
	}
	run := s.repo.CreateRun(model.PipelineRun{
		PipelineID:  pipeline.ID,
		Pipeline:    pipeline.Name,
		TemplateID:  template.ID,
		Template:    template.Name,
		Branch:      branch,
		Environment: environment,
		TriggerType: triggerType,
		CommitID:    strings.TrimSpace(req.CommitID),
		Operator:    operator,
		Status:      status,
		Stages:      mergeTemplateStages(template.Stages, req.Parameters),
		CreatedAt:   time.Now(),
	})
	return run, nil
}

func (s *CICDService) GetConfig() model.JenkinsConfig {
	return s.repo.GetConfig()
}

func (s *CICDService) SaveConfig(req SaveJenkinsConfigRequest) (model.JenkinsConfig, error) {
	endpoint := strings.TrimSpace(req.Endpoint)
	if endpoint == "" {
		return model.JenkinsConfig{}, obserr.New("JENKINS_ENDPOINT_REQUIRED", "cicd.SaveConfig", "Jenkins endpoint 不能为空")
	}
	defaultJob := strings.TrimSpace(req.DefaultJob)
	if defaultJob == "" {
		defaultJob = "default"
	}
	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 10
	}
	config := model.JenkinsConfig{
		Endpoint:              endpoint,
		Username:              strings.TrimSpace(req.Username),
		APIToken:              strings.TrimSpace(req.APIToken),
		DefaultJob:            defaultJob,
		TimeoutSeconds:        timeout,
		TLSInsecureSkipVerify: req.TLSInsecureSkipVerify,
	}
	if err := s.repo.ValidateConfigConnection(config); err != nil {
		return model.JenkinsConfig{}, obserr.Wrap("JENKINS_CONNECT_FAILED", "cicd.SaveConfig", "Jenkins 配置连接失败", err)
	}
	saved := s.repo.SaveConfig(config)
	return saved, nil
}

func (s *CICDService) ValidateCurrentConfig() error {
	config := s.repo.GetConfig()
	if err := s.repo.ValidateConfigConnection(config); err != nil {
		return obserr.Wrap("JENKINS_CONNECT_FAILED", "cicd.ValidateCurrentConfig", "Jenkins 配置连接失败", err)
	}
	return nil
}

func (s *CICDService) ListPipelineRuns(pipelineID uint, status string, limit int) ListPipelineRunsResponse {
	if limit <= 0 {
		limit = 20
	}
	items := s.repo.ListRuns(pipelineID, status)
	if len(items) > limit {
		items = items[:limit]
	}
	return ListPipelineRunsResponse{Total: len(items), Items: items}
}

func mergeTemplateStages(stages []model.TemplateStage, parameters map[string]string) []model.TemplateStage {
	result := make([]model.TemplateStage, 0, len(stages))
	for _, stage := range stages {
		params := make(map[string]string, len(stage.Parameters)+len(parameters))
		for k, v := range stage.Parameters {
			params[k] = v
		}
		for k, v := range parameters {
			trimmedKey := strings.TrimSpace(k)
			if trimmedKey == "" {
				continue
			}
			params[trimmedKey] = strings.TrimSpace(v)
		}
		result = append(result, model.TemplateStage{
			Name:       stage.Name,
			Kind:       stage.Kind,
			Order:      stage.Order,
			Parameters: params,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Order == result[j].Order {
			return result[i].Name < result[j].Name
		}
		return result[i].Order < result[j].Order
	})
	return result
}
