package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/cmdb/api"

	"github.com/gin-gonic/gin"
)

func registerCMDB(r *gin.RouterGroup) {
	g := r.Group("/cmdb")

	hostListPerm := middleware.RequirePermission("cmdb:host", "list")
	hostCreatePerm := middleware.RequirePermission("cmdb:host", "create")
	hostUpdatePerm := middleware.RequirePermission("cmdb:host", "update")
	hostDeletePerm := middleware.RequirePermission("cmdb:host", "delete")
	hostTestPerm := middleware.RequirePermission("cmdb:host", "test")
	groupListPerm := middleware.RequirePermission("cmdb:group", "list")
	groupCreatePerm := middleware.RequirePermission("cmdb:group", "create")
	groupUpdatePerm := middleware.RequirePermission("cmdb:group", "update")
	groupDeletePerm := middleware.RequirePermission("cmdb:group", "delete")
	credListPerm := middleware.RequirePermission("cmdb:credential", "list")
	credCreatePerm := middleware.RequirePermission("cmdb:credential", "create")
	credUpdatePerm := middleware.RequirePermission("cmdb:credential", "update")
	credDeletePerm := middleware.RequirePermission("cmdb:credential", "delete")
	terminalConnectPerm := middleware.RequirePermission("cmdb:terminal", "connect")
	terminalListPerm := middleware.RequirePermission("cmdb:terminal", "list")
	terminalGetPerm := middleware.RequirePermission("cmdb:terminal", "get")
	terminalReplayPerm := middleware.RequirePermission("cmdb:terminal", "replay")

	{
		// 主机管理
		g.GET("/host/list", hostListPerm, api.HostList)
		g.GET("/host/stats", hostListPerm, api.HostStats)
		g.GET("/host/detail", hostListPerm, api.HostDetail)
		g.POST("/host/create", hostCreatePerm, middleware.SetAuditOperation("创建主机"), api.HostCreate)
		g.POST("/host/batch", hostCreatePerm, middleware.SetAuditOperation("批量导入主机"), api.HostBatchCreate)
		g.POST("/host/update", hostUpdatePerm, middleware.SetAuditOperation("更新主机"), api.HostUpdate)
		g.POST("/host/delete", hostDeletePerm, middleware.SetAuditOperation("删除主机"), api.HostDelete)
		g.POST("/host/test", hostTestPerm, api.HostTest)

		// 分组管理
		g.GET("/group/tree", groupListPerm, api.GroupTreeAPI)
		g.GET("/group/detail", groupListPerm, api.GroupDetail)
		g.POST("/group/create", groupCreatePerm, middleware.SetAuditOperation("创建分组"), api.GroupCreate)
		g.POST("/group/update", groupUpdatePerm, middleware.SetAuditOperation("更新分组"), api.GroupUpdate)
		g.POST("/group/delete", groupDeletePerm, middleware.SetAuditOperation("删除分组"), api.GroupDelete)

		// 凭据管理
		g.GET("/credential/list", credListPerm, api.CredentialList)
		g.GET("/credential/detail", credListPerm, api.CredentialDetail)
		g.POST("/credential/create", credCreatePerm, middleware.SetAuditOperation("创建凭据"), api.CredentialCreate)
		g.POST("/credential/update", credUpdatePerm, middleware.SetAuditOperation("更新凭据"), api.CredentialUpdate)
		g.POST("/credential/delete", credDeletePerm, middleware.SetAuditOperation("删除凭据"), api.CredentialDelete)

		// 终端审计
		g.GET("/terminal/connect", terminalConnectPerm, api.TerminalConnect)
		g.GET("/terminal/list", terminalListPerm, api.TerminalList)
		g.GET("/terminal/detail", terminalGetPerm, api.TerminalDetail)
		g.GET("/terminal/recording", terminalReplayPerm, api.TerminalRecording)
	}
}
