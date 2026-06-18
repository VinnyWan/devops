package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"devops-platform/internal/modules/cicd/model"
	"devops-platform/internal/pkg/obserr"

	"gorm.io/gorm"
)

const op = "cicd/repository"

type CICDRepo struct {
	db         *gorm.DB
	httpClient *http.Client
}

func NewCICDRepo(db *gorm.DB) *CICDRepo {
	return &CICDRepo{
		db:         db,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// --- Config CRUD ---

func (r *CICDRepo) ListConfigs(page, pageSize int) ([]model.JenkinsConfig, int64, error) {
	var configs []model.JenkinsConfig
	var total int64
	q := r.db.Model(&model.JenkinsConfig{})
	q.Count(&total)
	if err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&configs).Error; err != nil {
		return nil, 0, obserr.Wrap("DB_ERROR", op, "list jenkins configs failed", err)
	}
	return configs, total, nil
}

func (r *CICDRepo) GetConfig(id uint) (*model.JenkinsConfig, error) {
	var cfg model.JenkinsConfig
	if err := r.db.First(&cfg, id).Error; err != nil {
		return nil, obserr.Wrap("DB_ERROR", op, "get jenkins config failed", err)
	}
	return &cfg, nil
}

func (r *CICDRepo) SaveConfig(cfg *model.JenkinsConfig) error {
	if err := r.db.Save(cfg).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "save jenkins config failed", err)
	}
	return nil
}

func (r *CICDRepo) DeleteConfig(id uint) error {
	if err := r.db.Delete(&model.JenkinsConfig{}, id).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "delete jenkins config failed", err)
	}
	return nil
}

// --- Jenkins API helpers ---

func (r *CICDRepo) jenkinsGet(cfg *model.JenkinsConfig, path string, result interface{}) error {
	u := strings.TrimRight(cfg.URL, "/") + path
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return obserr.Wrap("JENKINS_REQUEST_FAILED", op, "failed to build request", err)
	}
	req.SetBasicAuth(cfg.Username, cfg.APIToken)
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return obserr.Wrap("JENKINS_CONNECT_FAILED", op, "cannot reach jenkins server", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return obserr.New("JENKINS_NOT_FOUND", op, "resource not found on jenkins")
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return obserr.New("JENKINS_REQUEST_FAILED", op, fmt.Sprintf("jenkins returned %d: %s", resp.StatusCode, string(body)))
	}
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return obserr.Wrap("JENKINS_PARSE_FAILED", op, "failed to parse jenkins response", err)
		}
	}
	return nil
}

func (r *CICDRepo) jenkinsPost(cfg *model.JenkinsConfig, path string) error {
	u := strings.TrimRight(cfg.URL, "/") + path
	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return obserr.Wrap("JENKINS_REQUEST_FAILED", op, "failed to build request", err)
	}
	req.SetBasicAuth(cfg.Username, cfg.APIToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return obserr.Wrap("JENKINS_CONNECT_FAILED", op, "cannot reach jenkins server", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return obserr.New("JENKINS_REQUEST_FAILED", op, fmt.Sprintf("jenkins returned %d: %s", resp.StatusCode, string(body)))
	}
	return nil
}

// --- Connection test ---

func (r *CICDRepo) TestConnection(urlStr, username, apiToken string) error {
	cfg := &model.JenkinsConfig{URL: urlStr, Username: username, APIToken: apiToken}
	var result map[string]interface{}
	if err := r.jenkinsGet(cfg, "/api/json", &result); err != nil {
		return err
	}
	return nil
}

// --- Job browsing ---

type jenkinsJob struct {
	Name  string       `json:"name"`
	URL   string       `json:"url"`
	Color string       `json:"color"`
	Jobs  []jenkinsJob `json:"jobs,omitempty"`
}

func (r *CICDRepo) ListJobs(configID uint, keyword string) ([]model.JobInfo, error) {
	cfg, err := r.GetConfig(configID)
	if err != nil {
		return nil, obserr.Wrap("JENKINS_CONFIG_NOT_FOUND", op, "config not found", err)
	}

	var rootJobs []jenkinsJob
	if err := r.jenkinsGet(cfg, "/api/json?tree=jobs[name,url,color,jobs[name,url,color]]", &struct {
		Jobs *[]jenkinsJob `json:"jobs"`
	}{Jobs: &rootJobs}); err != nil {
		return nil, err
	}

	var result []model.JobInfo
	keyword = strings.ToLower(keyword)
	for _, j := range rootJobs {
		if keyword == "" || strings.Contains(strings.ToLower(j.Name), keyword) || strings.Contains(strings.ToLower(j.URL), keyword) {
			result = append(result, model.JobInfo{
				Name: j.Name, DisplayName: j.Name, URL: j.URL, Color: j.Color, Buildable: j.Color != "disabled",
			})
		}
	}
	return result, nil
}

// --- Build management ---

func (r *CICDRepo) TriggerBuild(configID uint, jobName string) error {
	cfg, err := r.GetConfig(configID)
	if err != nil {
		return obserr.Wrap("JENKINS_CONFIG_NOT_FOUND", op, "config not found", err)
	}
	path := fmt.Sprintf("/job/%s/build", url.PathEscape(jobName))
	return r.jenkinsPost(cfg, path)
}

func (r *CICDRepo) ListBuilds(configID uint, jobName string) ([]model.BuildInfo, error) {
	cfg, err := r.GetConfig(configID)
	if err != nil {
		return nil, obserr.Wrap("JENKINS_CONFIG_NOT_FOUND", op, "config not found", err)
	}

	path := fmt.Sprintf("/job/%s/api/json?tree=builds[number,url,result,duration,timestamp,building]", url.PathEscape(jobName))
	var resp struct {
		Builds []struct {
			Number    int    `json:"number"`
			URL       string `json:"url"`
			Result    string `json:"result"`
			Duration  int64  `json:"duration"`
			Timestamp int64  `json:"timestamp"`
			Building  bool   `json:"building"`
		} `json:"builds"`
	}
	if err := r.jenkinsGet(cfg, path, &resp); err != nil {
		return nil, err
	}

	var builds []model.BuildInfo
	for _, b := range resp.Builds {
		builds = append(builds, model.BuildInfo{
			Number: b.Number, URL: b.URL, Result: b.Result,
			Duration: b.Duration, Timestamp: b.Timestamp, Building: b.Building,
		})
	}
	return builds, nil
}

func (r *CICDRepo) GetBuildLog(configID uint, jobName string, buildNumber int) (*model.BuildLogEntry, error) {
	cfg, err := r.GetConfig(configID)
	if err != nil {
		return nil, obserr.Wrap("JENKINS_CONFIG_NOT_FOUND", op, "config not found", err)
	}

	path := fmt.Sprintf("/job/%s/%d/consoleText", url.PathEscape(jobName), buildNumber)
	u := strings.TrimRight(cfg.URL, "/") + path
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, obserr.Wrap("JENKINS_REQUEST_FAILED", op, "failed to build request", err)
	}
	req.SetBasicAuth(cfg.Username, cfg.APIToken)
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, obserr.Wrap("JENKINS_CONNECT_FAILED", op, "cannot reach jenkins", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // limit to 1MB
	if err != nil {
		return nil, obserr.Wrap("JENKINS_LOG_FAILED", op, "failed to read build log", err)
	}

	text := string(body)
	hasMore := len(text) >= 1<<20
	return &model.BuildLogEntry{Text: text, HasMore: hasMore}, nil
}

// Pipeline CRUD (DB-backed)
func (r *CICDRepo) ListPipelines(page, pageSize int) ([]model.Pipeline, int64, error) {
	var pipelines []model.Pipeline
	var total int64
	q := r.db.Model(&model.Pipeline{})
	q.Count(&total)
	if err := q.Offset((page-1)*pageSize).Limit(pageSize).Order("created_at DESC").Find(&pipelines).Error; err != nil {
		return nil, 0, obserr.Wrap("DB_ERROR", op, "list pipelines failed", err)
	}
	return pipelines, total, nil
}

func (r *CICDRepo) SavePipeline(p *model.Pipeline) error {
	if err := r.db.Save(p).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "save pipeline failed", err)
	}
	return nil
}

func (r *CICDRepo) DeletePipeline(id uint) error {
	if err := r.db.Delete(&model.Pipeline{}, id).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "delete pipeline failed", err)
	}
	return nil
}
