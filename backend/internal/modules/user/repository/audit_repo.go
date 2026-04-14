package repository

import (
	"devops-platform/internal/modules/user/model"
	queryutil "devops-platform/internal/pkg/query"
	"time"

	"gorm.io/gorm"
)

type AuditRepo struct {
	db *gorm.DB
}

type AuditQuery struct {
	UserID    *uint
	Username  string
	Method    string
	Operation string
	Resource  string
	Keyword   string
	StartAt   *time.Time
	EndAt     *time.Time
	Page      int
	PageSize  int
}

func NewAuditRepo(db *gorm.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

// Create 创建审计日志
func (r *AuditRepo) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

// CleanExpired 清理过期日志
func (r *AuditRepo) CleanExpired(now time.Time) (int64, error) {
	var logs []model.AuditLog
	if err := r.db.Select("id", "created_at", "retention_days").Find(&logs).Error; err != nil {
		return 0, err
	}
	expiredIDs := make([]uint, 0, len(logs))
	for _, log := range logs {
		if log.CreatedAt.AddDate(0, 0, log.RetentionDays).Before(now) {
			expiredIDs = append(expiredIDs, log.ID)
		}
	}
	if len(expiredIDs) == 0 {
		return 0, nil
	}
	result := r.db.Where("id IN ?", expiredIDs).Delete(&model.AuditLog{})
	return result.RowsAffected, result.Error
}

func (r *AuditRepo) List(query AuditQuery) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 200 {
		query.PageSize = 20
	}

	tx := r.buildListQuery(query)

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := tx.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *AuditRepo) ListForExport(query AuditQuery, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	if limit <= 0 || limit > 50000 {
		limit = 10000
	}
	tx := r.buildListQuery(query)
	if err := tx.Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *AuditRepo) buildListQuery(query AuditQuery) *gorm.DB {
	tx := r.db.Model(&model.AuditLog{})
	if query.UserID != nil {
		tx = tx.Where("user_id = ?", *query.UserID)
	}
	if query.Username != "" {
		tx = tx.Where("username LIKE ?", "%"+query.Username+"%")
	}
	if query.Method != "" {
		tx = tx.Where("method = ?", query.Method)
	}
	if query.Operation != "" {
		tx = tx.Where("operation LIKE ?", "%"+query.Operation+"%")
	}
	if query.Resource != "" {
		tx = tx.Where("path LIKE ?", "%"+query.Resource+"%")
	}
	tx = queryutil.ApplyKeywordLike(tx, query.Keyword, "username", "operation", "method", "path", "ip", "params", "result", "error_message")
	if query.StartAt != nil {
		tx = tx.Where("created_at >= ?", *query.StartAt)
	}
	if query.EndAt != nil {
		tx = tx.Where("created_at <= ?", *query.EndAt)
	}
	return tx
}
