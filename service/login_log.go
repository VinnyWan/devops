package service

import (
	"devops/internal/database"
	"devops/models"
	"time"
)

type LoginLogService struct{}

// Create 创建登录日志
func (s *LoginLogService) Create(log *models.LoginLog) error {
	log.LoginTime = time.Now()
	return database.Db.Create(log).Error
}

// GetList 获取登录日志列表
func (s *LoginLogService) GetList(page, pageSize int, username, ip string) ([]models.LoginLog, int64, error) {
	var logs []models.LoginLog
	var total int64

	query := database.Db.Model(&models.LoginLog{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if ip != "" {
		query = query.Where("ip LIKE ?", "%"+ip+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("login_time DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// Delete 删除登录日志
func (s *LoginLogService) Delete(id uint) error {
	return database.Db.Delete(&models.LoginLog{}, id).Error
}

// Clear 清空登录日志
func (s *LoginLogService) Clear() error {
	return database.Db.Exec("TRUNCATE TABLE sys_login_log").Error
}
