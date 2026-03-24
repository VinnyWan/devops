package v1

import (
	"devops-platform/internal/middleware"
	alertAPI "devops-platform/internal/modules/alert/api"

	"github.com/gin-gonic/gin"
)

func registerAlert(r *gin.RouterGroup) {
	g := r.Group("/alert")
	listPermission := middleware.RequirePermission("alert", "list")
	updatePermission := middleware.RequirePermission("alert", "update")
	createPermission := middleware.RequirePermission("alert", "create")
	{
		g.GET("/rules", listPermission, alertAPI.ListAlertRules)
		g.GET("/history", listPermission, alertAPI.ListAlertHistory)
		g.POST("/rule/toggle", updatePermission, middleware.SetAuditOperation("告警规则启停"), alertAPI.ToggleAlertRule)
		g.GET("/silences", listPermission, alertAPI.ListAlertSilences)
		g.POST("/silence/upsert", createPermission, middleware.SetAuditOperation("告警静默配置"), alertAPI.UpsertAlertSilence)
		g.GET("/channels", listPermission, alertAPI.ListAlertChannels)
		g.POST("/channel/upsert", createPermission, middleware.SetAuditOperation("告警通知渠道配置"), alertAPI.UpsertAlertChannel)
		g.GET("/config", listPermission, alertAPI.GetAlertmanagerConfig)
		g.POST("/config/upsert", updatePermission, middleware.SetAuditOperation("Alertmanager 配置更新"), alertAPI.SaveAlertmanagerConfig)
	}
}
