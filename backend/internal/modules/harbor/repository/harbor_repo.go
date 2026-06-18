package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"devops-platform/internal/modules/harbor/model"
	"devops-platform/internal/pkg/obserr"

	"gorm.io/gorm"
)

const op = "harbor/repository"

type HarborRepo struct {
	db         *gorm.DB
	httpClient *http.Client
}

func NewHarborRepo(db *gorm.DB) *HarborRepo {
	return &HarborRepo{
		db:         db,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// --- Config CRUD ---

func (r *HarborRepo) ListConfigs(page, pageSize int) ([]model.HarborConfig, int64, error) {
	var configs []model.HarborConfig
	var total int64
	q := r.db.Model(&model.HarborConfig{})
	q.Count(&total)
	if err := q.Offset((page-1)*pageSize).Limit(pageSize).Order("created_at DESC").Find(&configs).Error; err != nil {
		return nil, 0, obserr.Wrap("DB_ERROR", op, "list harbor configs failed", err)
	}
	return configs, total, nil
}

func (r *HarborRepo) GetConfig(id uint) (*model.HarborConfig, error) {
	var cfg model.HarborConfig
	if err := r.db.First(&cfg, id).Error; err != nil {
		return nil, obserr.Wrap("DB_ERROR", op, "get harbor config failed", err)
	}
	return &cfg, nil
}

func (r *HarborRepo) SaveConfig(cfg *model.HarborConfig) error {
	if err := r.db.Save(cfg).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "save harbor config failed", err)
	}
	return nil
}

func (r *HarborRepo) DeleteConfig(id uint) error {
	if err := r.db.Delete(&model.HarborConfig{}, id).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "delete harbor config failed", err)
	}
	return nil
}

// --- Harbor API helpers ---

func (r *HarborRepo) harborRequest(cfg *model.HarborConfig, method, path string, result interface{}) error {
	u := strings.TrimRight(cfg.URL, "/") + "/api/v2.0" + path
	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return obserr.Wrap("HARBOR_REQUEST_FAILED", op, "failed to build request", err)
	}
	req.SetBasicAuth(cfg.Username, cfg.Password)
	req.Header.Set("Accept", "application/json")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return obserr.Wrap("HARBOR_CONNECT_FAILED", op, "cannot reach harbor server", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return obserr.New("HARBOR_AUTH_FAILED", op, "harbor authentication failed")
	}
	if resp.StatusCode == 404 {
		return obserr.New("HARBOR_NOT_FOUND", op, "resource not found on harbor")
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return obserr.New("HARBOR_REQUEST_FAILED", op, fmt.Sprintf("harbor returned %d: %s", resp.StatusCode, string(body)))
	}
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return obserr.Wrap("HARBOR_PARSE_FAILED", op, "failed to parse harbor response", err)
		}
	}
	return nil
}

// --- Connection test ---

func (r *HarborRepo) TestConnection(url, username, password string) error {
	cfg := &model.HarborConfig{URL: url, Username: username, Password: password}
	var info map[string]interface{}
	if err := r.harborRequest(cfg, "GET", "/systeminfo", &info); err != nil {
		return err
	}
	return nil
}

// --- Projects ---

func (r *HarborRepo) ListProjects(configID uint, keyword string, page, pageSize int) ([]model.Project, int64, error) {
	cfg, err := r.GetConfig(configID)
	if err != nil {
		return nil, 0, obserr.Wrap("HARBOR_CONFIG_NOT_FOUND", op, "config not found", err)
	}

	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("page_size", fmt.Sprintf("%d", pageSize))
	if keyword != "" {
		params.Set("name", keyword)
	}

	var projects []model.Project
	path := "/projects?" + params.Encode()
	if err := r.harborRequest(cfg, "GET", path, &projects); err != nil {
		return nil, 0, err
	}

	// Harbor v2.0 API returns projects directly; total count from headers or estimate
	return projects, int64(len(projects)), nil
}

// --- Repositories ---

func (r *HarborRepo) ListRepositories(configID uint, projectName string, keyword string, page, pageSize int) ([]model.Repository, int64, error) {
	cfg, err := r.GetConfig(configID)
	if err != nil {
		return nil, 0, obserr.Wrap("HARBOR_CONFIG_NOT_FOUND", op, "config not found", err)
	}

	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("page_size", fmt.Sprintf("%d", pageSize))
	if keyword != "" {
		params.Set("q", keyword)
	}

	var repos []model.Repository
	path := fmt.Sprintf("/projects/%s/repositories?%s", url.PathEscape(projectName), params.Encode())
	if err := r.harborRequest(cfg, "GET", path, &repos); err != nil {
		return nil, 0, err
	}
	return repos, int64(len(repos)), nil
}

// --- Artifacts (images with tags) ---

func (r *HarborRepo) ListArtifacts(configID uint, projectName, repoName string, page, pageSize int) ([]model.Artifact, int64, error) {
	cfg, err := r.GetConfig(configID)
	if err != nil {
		return nil, 0, obserr.Wrap("HARBOR_CONFIG_NOT_FOUND", op, "config not found", err)
	}

	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("page_size", fmt.Sprintf("%d", pageSize))

	var artifacts []model.Artifact
	path := fmt.Sprintf("/projects/%s/repositories/%s/artifacts?%s",
		url.PathEscape(projectName), url.PathEscape(repoName), params.Encode())
	if err := r.harborRequest(cfg, "GET", path, &artifacts); err != nil {
		return nil, 0, err
	}
	return artifacts, int64(len(artifacts)), nil
}

// --- Delete artifact tag ---

func (r *HarborRepo) DeleteArtifact(configID uint, projectName, repoName, reference string) error {
	cfg, err := r.GetConfig(configID)
	if err != nil {
		return obserr.Wrap("HARBOR_CONFIG_NOT_FOUND", op, "config not found", err)
	}

	path := fmt.Sprintf("/projects/%s/repositories/%s/artifacts/%s",
		url.PathEscape(projectName), url.PathEscape(repoName), url.PathEscape(reference))
	return r.harborRequest(cfg, "DELETE", path, nil)
}
