package bootstrap

import (
	"path/filepath"

	"devops-platform/internal/pkg/logger"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Enforcer *casbin.SyncedEnforcer

// InitCasbin 初始化 Casbin Enforcer，使用 gorm-adapter 连接数据库
func InitCasbin(db *gorm.DB) error {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		logger.Log.Error("Casbin gorm-adapter 创建失败", zap.Error(err))
		return err
	}

	// 读取 RBAC 模型配置
	configPath := filepath.Join("config", "rbac_model.conf")
	m, err := model.NewModelFromFile(configPath)
	if err != nil {
		logger.Log.Error("Casbin 模型加载失败", zap.Error(err))
		return err
	}

	syncedEnforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		logger.Log.Error("Casbin Enforcer 创建失败", zap.Error(err))
		return err
	}

	// 加载已有策略
	if err := syncedEnforcer.LoadPolicy(); err != nil {
		logger.Log.Error("Casbin 策略加载失败", zap.Error(err))
		return err
	}

	Enforcer = syncedEnforcer
	logger.Log.Info("Casbin 初始化成功")
	return nil
}

// GetEnforcer 获取全局 Casbin Enforcer 实例
func GetEnforcer() *casbin.SyncedEnforcer {
	return Enforcer
}
