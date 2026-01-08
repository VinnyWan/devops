package database

import (
	"devops/internal/logger"
	k8smodels "devops/models/k8s"
	usermodels "devops/models/user"

	"go.uber.org/zap"
)

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	logger.Log.Info("开始数据库迁移...")

	err := Db.AutoMigrate(
		// 用户相关表
		&usermodels.User{},
		&usermodels.Role{},
		&usermodels.Menu{},
		&usermodels.Department{},
		&usermodels.Post{},
		&usermodels.OperationLog{},
		&usermodels.LoginLog{},

		// K8s相关表
		&k8smodels.Cluster{},
		&k8smodels.ClusterAccess{},
		&k8smodels.Namespace{},
		&k8smodels.OperationLog{},
	)

	if err != nil {
		logger.Log.Error("数据库迁移失败", zap.Error(err))
		return err
	}

	logger.Log.Info("数据库迁移完成")
	return nil
}
