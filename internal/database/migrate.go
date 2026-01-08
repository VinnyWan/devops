package database

import (
	"devops/internal/logger"
	"devops/models"

	"go.uber.org/zap"
)

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	logger.Log.Info("开始数据库迁移...")

	err := Db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Menu{},
		&models.Department{},
		&models.Post{},
		&models.OperationLog{},
		&models.LoginLog{},
	)

	if err != nil {
		logger.Log.Error("数据库迁移失败", zap.Error(err))
		return err
	}

	logger.Log.Info("数据库迁移完成")
	return nil
}
