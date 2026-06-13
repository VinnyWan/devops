package bootstrap

import (
	"context"

	"devops-platform/config"
	"devops-platform/internal/middleware"
	"devops-platform/internal/pkg/logger"
	alertService "devops-platform/internal/modules/alert/service"
	sqlAuditAPI "devops-platform/internal/modules/sqlaudit/api"
	sqlAuditService "devops-platform/internal/modules/sqlaudit/service"
	cmdbAPI "devops-platform/internal/modules/cmdb/api"
	k8sAPI "devops-platform/internal/modules/k8s/api"
	notifModel "devops-platform/internal/modules/notification/model"
	notifService "devops-platform/internal/modules/notification/service"
	taskAPI "devops-platform/internal/modules/task/api"
	taskModel "devops-platform/internal/modules/task/model"
	taskService "devops-platform/internal/modules/task/service"
	toolAPI "devops-platform/internal/modules/tool/api"
	toolService "devops-platform/internal/modules/tool/service"
	userAPI "devops-platform/internal/modules/user/api"
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/modules/user/service"
	workflowAPI "devops-platform/internal/modules/workflow/api"
	workflowService "devops-platform/internal/modules/workflow/service"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitModules initializes all business modules after DB and dependencies are ready.
// This centralizes module wiring that would otherwise clutter main.go.
func InitModules(db *gorm.DB) {
	// User module
	userAPI.SetDB(db)
	userAPI.InitApiKeyService(db)

	// K8s module (requires DB + K8sFactory)
	k8sAPI.SetK8sDB(db, K8sFactory)

	// CMDB module
	cmdbAPI.SetDB(db)

	// Middleware
	middleware.SetDB(db)
	middleware.SetAuthDB(db)
	middleware.SetCasbinEnforcer(Enforcer)

	// Casbin 策略同步：将业务表（roles/user_roles/role_permissions）中的权限数据同步到 casbin_rule 表
	casbinSyncSvc := service.NewCasbinSyncService(db, Enforcer)
	var tenantIDs []uint
	if err := db.Table("tenants").Select("id").Pluck("id", &tenantIDs).Error; err != nil {
		logger.Log.Warn("查询租户列表失败，跳过 Casbin 策略同步", zap.Error(err))
	}
	for _, tid := range tenantIDs {
		if err := casbinSyncSvc.SyncTenantPolicies(tid); err != nil {
			logger.Log.Warn("Casbin 策略同步失败", zap.Uint("tenantID", tid), zap.Error(err))
		} else {
			logger.Log.Info("Casbin 策略同步完成", zap.Uint("tenantID", tid))
		}
	}

	// Audit log cleanup
	auditRepo := repository.NewAuditRepo(db)
	auditService := service.NewAuditService(auditRepo)
	auditService.StartAuditCleanupTask()

	// Task engine: scheduler + service
	scheduler := taskService.NewScheduler()
	ts := taskService.NewTaskService(db, scheduler)
	scheduler.SetExecutor(func(ctx context.Context, taskID, tenantID uint) (*taskModel.TaskExecution, error) {
		return ts.Execute(ctx, taskID, tenantID, nil)
	})
	scheduler.Start()
	ts.StartCleanupTask(30)
	taskAPI.InitTaskService(ts)

	// Notification hub: service + notifiers
	ns := notifService.NewNotificationService(db)
	feishuURL := config.Cfg.GetString("notification.feishu_webhook")
	if feishuURL != "" {
		ns.RegisterNotifier(notifModel.ChannelFeishu, notifService.NewFeishuNotifier(feishuURL))
	}
	dingtalkURL := config.Cfg.GetString("notification.dingtalk_webhook")
	if dingtalkURL != "" {
		ns.RegisterNotifier(notifModel.ChannelDingTalk, notifService.NewDingTalkNotifier(dingtalkURL))
	}

	// Alert-notification bridge
	alertBridge := alertService.NewAlertNotificationBridge(ns)

	// Workflow engine: service + callback executor
	ws := workflowService.NewWorkflowService(db)
	callbackExecutor := workflowService.NewCallbackExecutor()
	callbackExecutor.Register("task", &workflowService.TaskCallback{
		ExecuteTask: func(taskID, tenantID uint) error {
			_, err := ts.Execute(context.Background(), taskID, tenantID, nil)
			return err
		},
	})
	callbackExecutor.Register("notification", &workflowService.NotificationCallback{
		SendNotification: func(tenantID uint, channel, recipients, subject, body string) error {
			ch := notifModel.ChannelType(channel)
			return ns.Send(tenantID, ch, []string{recipients}, subject, body)
		},
	})
	ws.SetCallbackExecutor(callbackExecutor)
	workflowAPI.InitWorkflowService(ws)

	// Tool marketplace: service + builtin scripts seed
	toolSvc := toolService.NewToolService(db)
	if err := toolSvc.SeedBuiltinScripts(); err != nil {
		if logger.Log != nil {
			logger.Log.Warn("seed tool scripts failed", zap.Error(err))
		}
	}
	toolAPI.InitToolService(toolSvc)

	// SQL audit: service
	sqlAuditSvc := sqlAuditService.NewSqlAuditService(db)
	sqlAuditAPI.InitSqlAuditService(sqlAuditSvc)

	// Store references for background tasks
	_ = alertBridge
	_ = toolSvc
	_ = sqlAuditSvc
}

// StartModuleBackgroundTasks starts background services (schedulers, cleanup, etc.)
func StartModuleBackgroundTasks() {
	cmdbAPI.StartCloudSync()
	cmdbAPI.StartRecordingCleanup()
}
