package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"

	"gorm.io/gorm"
)

type TenantService struct {
	repo *repository.TenantRepo
}

type CreateTenantRequest struct {
	Name           string     `json:"name" binding:"required"`
	Code           string     `json:"code" binding:"required"`
	Description    string     `json:"description"`
	Logo           string     `json:"logo"`
	MaxUsers       int        `json:"maxUsers"`
	MaxDepartments int        `json:"maxDepartments"`
	MaxRoles       int        `json:"maxRoles"`
	Modules        string     `json:"modules"`
	ContactName    string     `json:"contactName"`
	ContactEmail   string     `json:"contactEmail"`
	ContactPhone   string     `json:"contactPhone"`
	ExpiresAt      *time.Time `json:"expiresAt"`
}

type UpdateTenantRequest struct {
	ID             uint       `json:"id" binding:"required"`
	Name           *string    `json:"name"`
	Description    *string    `json:"description"`
	Logo           *string    `json:"logo"`
	Status         *string    `json:"status"`
	MaxUsers       *int       `json:"maxUsers"`
	MaxDepartments *int       `json:"maxDepartments"`
	MaxRoles       *int       `json:"maxRoles"`
	Modules        *string    `json:"modules"`
	ContactName    *string    `json:"contactName"`
	ContactEmail   *string    `json:"contactEmail"`
	ContactPhone   *string    `json:"contactPhone"`
	ExpiresAt      *time.Time `json:"expiresAt"`
}

func NewTenantService(repo *repository.TenantRepo) *TenantService {
	return &TenantService{repo: repo}
}

func (s *TenantService) Create(req *CreateTenantRequest) (*model.Tenant, error) {
	code := strings.ToLower(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, errors.New("tenant code is required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("tenant name is required")
	}

	tenant := &model.Tenant{
		Name:           name,
		Code:           code,
		Description:    strings.TrimSpace(req.Description),
		Logo:           strings.TrimSpace(req.Logo),
		Status:         "active",
		MaxUsers:       normalizePositive(req.MaxUsers, 100),
		MaxDepartments: normalizePositive(req.MaxDepartments, 20),
		MaxRoles:       normalizePositive(req.MaxRoles, 50),
		Modules:        strings.TrimSpace(req.Modules),
		ContactName:    strings.TrimSpace(req.ContactName),
		ContactEmail:   strings.TrimSpace(req.ContactEmail),
		ContactPhone:   strings.TrimSpace(req.ContactPhone),
		ExpiresAt:      req.ExpiresAt,
	}

	if err := s.repo.Create(tenant); err != nil {
		if isDuplicateErr(err) {
			return nil, errors.New("tenant code or name already exists")
		}
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	return tenant, nil
}

func (s *TenantService) GetByID(id uint) (*model.Tenant, error) {
	return s.repo.GetByID(id)
}

func (s *TenantService) GetByCode(code string) (*model.Tenant, error) {
	return s.repo.GetByCode(strings.ToLower(strings.TrimSpace(code)))
}

func (s *TenantService) List(page, pageSize int, keyword, status string) ([]model.Tenant, int64, error) {
	return s.repo.List(page, pageSize, keyword, status)
}

func (s *TenantService) Update(req *UpdateTenantRequest) error {
	_, err := s.repo.GetByID(req.ID)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return errors.New("tenant name cannot be empty")
		}
		updates["name"] = name
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.Logo != nil {
		updates["logo"] = strings.TrimSpace(*req.Logo)
	}
	if req.Status != nil {
		status := strings.TrimSpace(strings.ToLower(*req.Status))
		if status != "active" && status != "inactive" && status != "suspended" {
			return errors.New("invalid tenant status")
		}
		updates["status"] = status
	}
	if req.MaxUsers != nil {
		updates["max_users"] = normalizePositive(*req.MaxUsers, 100)
	}
	if req.MaxDepartments != nil {
		updates["max_departments"] = normalizePositive(*req.MaxDepartments, 20)
	}
	if req.MaxRoles != nil {
		updates["max_roles"] = normalizePositive(*req.MaxRoles, 50)
	}
	if req.Modules != nil {
		updates["modules"] = strings.TrimSpace(*req.Modules)
	}
	if req.ContactName != nil {
		updates["contact_name"] = strings.TrimSpace(*req.ContactName)
	}
	if req.ContactEmail != nil {
		updates["contact_email"] = strings.TrimSpace(*req.ContactEmail)
	}
	if req.ContactPhone != nil {
		updates["contact_phone"] = strings.TrimSpace(*req.ContactPhone)
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = req.ExpiresAt
	}

	if len(updates) == 0 {
		return nil
	}
	if err := s.repo.UpdateByID(req.ID, updates); err != nil {
		if isDuplicateErr(err) {
			return errors.New("tenant name already exists")
		}
		return err
	}
	return nil
}

func (s *TenantService) Disable(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	return s.repo.UpdateByID(id, map[string]interface{}{"status": "inactive"})
}

func normalizePositive(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}

func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
