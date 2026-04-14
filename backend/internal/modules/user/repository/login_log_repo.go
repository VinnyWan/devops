package repository

import (
	"devops-platform/internal/modules/user/model"
	"time"

	"gorm.io/gorm"
)

type LoginLogRepo struct {
	db *gorm.DB
}

type LoginLogQuery struct {
	Username string
	Status   string
	StartAt  *time.Time
	EndAt    *time.Time
	Page     int
	PageSize int
}

func NewLoginLogRepo(db *gorm.DB) *LoginLogRepo {
	return &LoginLogRepo{db: db}
}

// Create 创建登录日志
func (r *LoginLogRepo) Create(log *model.LoginLog) error {
	return r.db.Create(log).Error
}

// List 分页查询登录日志
func (r *LoginLogRepo) List(query LoginLogQuery) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
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
	if err := tx.Order("login_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *LoginLogRepo) buildListQuery(query LoginLogQuery) *gorm.DB {
	tx := r.db.Model(&model.LoginLog{})
	if query.Username != "" {
		tx = tx.Where("username LIKE ?", "%"+query.Username+"%")
	}
	if query.Status != "" {
		tx = tx.Where("status = ?", query.Status)
	}
	if query.StartAt != nil {
		tx = tx.Where("login_at >= ?", *query.StartAt)
	}
	if query.EndAt != nil {
		tx = tx.Where("login_at <= ?", *query.EndAt)
	}
	return tx
}
