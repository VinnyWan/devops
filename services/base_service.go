package services

import (
	"devops/internal/database"

	"gorm.io/gorm"
)

// BaseService 基础服务结构，包含数据库连接
type BaseService struct {
	DB *gorm.DB
}

// NewBaseService 创建基础服务实例
func NewBaseService() *BaseService {
	return &BaseService{
		DB: database.Db,
	}
}

// GetDB 获取数据库实例
func (s *BaseService) GetDB() *gorm.DB {
	return s.DB
}
