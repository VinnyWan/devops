package user

import (
	"devops/internal/database"
	usermodels "devops/models/user"
)

type OperationLogService struct{}

// Create 创建操作日志
func (s *OperationLogService) Create(log *usermodels.OperationLog) error {
	return database.Db.Create(log).Error
}

// GetList 获取操作日志列表
func (s *OperationLogService) GetList(page, pageSize int, module, operatorName string) ([]usermodels.OperationLog, int64, error) {
	var logs []usermodels.OperationLog
	var total int64

	query := database.Db.Model(&usermodels.OperationLog{})

	if module != "" {
		query = query.Where("module LIKE ?", "%"+module+"%")
	}
	if operatorName != "" {
		query = query.Where("operator_name LIKE ?", "%"+operatorName+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// Delete 删除操作日志
func (s *OperationLogService) Delete(id uint) error {
	return database.Db.Delete(&usermodels.OperationLog{}, id).Error
}

// Clear 清空操作日志
func (s *OperationLogService) Clear() error {
	return database.Db.Exec("TRUNCATE TABLE sys_operation_log").Error
}
