package repository

import (
	"errors"
	"strings"
	"sync"
	"time"

	"devops-platform/internal/modules/harbor/model"
)

type HarborRepo struct {
	mu       sync.RWMutex
	projects []model.HarborProject
	images   []model.RepositoryImage
	config   model.HarborConfig
}

func NewHarborRepo() *HarborRepo {
	now := time.Now()
	return &HarborRepo{
		projects: []model.HarborProject{
			{ID: 1, Name: "platform", Public: false, CreatedAt: now.Add(-60 * 24 * time.Hour), UpdatedAt: now.Add(-3 * time.Hour)},
			{ID: 2, Name: "payments", Public: false, CreatedAt: now.Add(-45 * 24 * time.Hour), UpdatedAt: now.Add(-50 * time.Minute)},
			{ID: 3, Name: "shared", Public: true, CreatedAt: now.Add(-90 * 24 * time.Hour), UpdatedAt: now.Add(-5 * time.Hour)},
		},
		images: []model.RepositoryImage{
			{ID: 1001, ProjectName: "platform", Repository: "gateway", Tag: "v1.9.0", Digest: "sha256:abc001", Size: 214748364, PushedAt: now.Add(-4 * time.Hour)},
			{ID: 1002, ProjectName: "platform", Repository: "gateway", Tag: "v1.8.9", Digest: "sha256:abc000", Size: 214000000, PushedAt: now.Add(-72 * time.Hour)},
			{ID: 1003, ProjectName: "payments", Repository: "payments-api", Tag: "v2.3.4", Digest: "sha256:def123", Size: 198765432, PushedAt: now.Add(-55 * time.Minute)},
			{ID: 1004, ProjectName: "shared", Repository: "busybox-tools", Tag: "1.0.0", Digest: "sha256:xyz900", Size: 73400320, PushedAt: now.Add(-12 * time.Hour)},
		},
		config: model.HarborConfig{
			Endpoint:       "https://harbor.example.com",
			Project:        "platform",
			Username:       "admin",
			Password:       "password",
			TimeoutSeconds: 10,
			UpdatedAt:      now,
		},
	}
}

func (r *HarborRepo) ListProjects() []model.HarborProject {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.HarborProject(nil), r.projects...)
}

func (r *HarborRepo) ListImages(projectName string) []model.RepositoryImage {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.RepositoryImage, 0, len(r.images))
	for _, item := range r.images {
		if projectName != "" && item.ProjectName != projectName {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (r *HarborRepo) GetConfig() model.HarborConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

func (r *HarborRepo) SaveConfig(cfg model.HarborConfig) model.HarborConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	cfg.UpdatedAt = time.Now()
	r.config = cfg
	return r.config
}

func (r *HarborRepo) ValidateConfigConnection(cfg model.HarborConfig) error {
	if strings.Contains(strings.ToLower(cfg.Endpoint), "invalid") {
		return errors.New("harbor endpoint 不可达")
	}
	if strings.Contains(strings.ToLower(cfg.Endpoint), "timeout") {
		return errors.New("harbor 请求超时")
	}
	return nil
}
