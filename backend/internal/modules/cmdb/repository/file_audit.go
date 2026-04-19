package repository

import (
	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
)

type FileAuditRepo struct {
	db *gorm.DB
}

func NewFileAuditRepo(db *gorm.DB) *FileAuditRepo {
	return &FileAuditRepo{db: db}
}

func (r *FileAuditRepo) Create(log *model.FileOperationLog) error {
	return r.db.Create(log).Error
}

func (r *FileAuditRepo) ListInTenant(tenantID uint, page, pageSize int, keyword, opType, username, hostIP string, startAt, endAt string) ([]model.FileOperationLog, int64, error) {
	query := r.db.Where("tenant_id = ?", tenantID)

	if keyword != "" {
		query = query.Where("file_path LIKE ?", "%"+keyword+"%")
	}
	if opType != "" {
		query = query.Where("op_type = ?", opType)
	}
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if hostIP != "" {
		query = query.Where("host_ip LIKE ?", "%"+hostIP+"%")
	}
	if startAt != "" {
		query = query.Where("created_at >= ?", startAt)
	}
	if endAt != "" {
		query = query.Where("created_at <= ?", endAt)
	}

	var total int64
	if err := query.Model(&model.FileOperationLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []model.FileOperationLog
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
