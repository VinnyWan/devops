package repository

import (
	"devops-platform/internal/modules/tool/model"

	"gorm.io/gorm"
)

type ToolRepo struct {
	db *gorm.DB
}

func NewToolRepo(db *gorm.DB) *ToolRepo {
	return &ToolRepo{db: db}
}

func (r *ToolRepo) ListTools(category string) ([]model.Tool, error) {
	var tools []model.Tool
	q := r.db.Model(&model.Tool{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	err := q.Order("name ASC").Find(&tools).Error
	return tools, err
}

func (r *ToolRepo) GetByID(id uint) (*model.Tool, error) {
	var t model.Tool
	err := r.db.First(&t, id).Error
	return &t, err
}

func (r *ToolRepo) Upsert(tool *model.Tool) error {
	return r.db.Where("name = ?", tool.Name).Assign(tool).FirstOrCreate(tool).Error
}

func (r *ToolRepo) Delete(id uint) error {
	return r.db.Delete(&model.Tool{}, id).Error
}

func (r *ToolRepo) ListInstallations(tenantID, hostID uint) ([]model.ToolInstallation, error) {
	var installs []model.ToolInstallation
	q := r.db.Where("tenant_id = ?", tenantID)
	if hostID > 0 {
		q = q.Where("host_id = ?", hostID)
	}
	err := q.Order("created_at DESC").Find(&installs).Error
	return installs, err
}

func (r *ToolRepo) GetInstallation(tenantID, toolID, hostID uint) (*model.ToolInstallation, error) {
	var inst model.ToolInstallation
	err := r.db.Where("tenant_id = ? AND tool_id = ? AND host_id = ?", tenantID, toolID, hostID).First(&inst).Error
	return &inst, err
}

func (r *ToolRepo) UpsertInstallation(inst *model.ToolInstallation) error {
	return r.db.Where("tenant_id = ? AND tool_id = ? AND host_id = ?", inst.TenantID, inst.ToolID, inst.HostID).Assign(inst).FirstOrCreate(inst).Error
}
