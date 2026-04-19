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
	permListPerm := middleware.RequirePermission("cmdb:permission", "list")
	permCreatePerm := middleware.RequirePermission("cmdb:permission", "create")
	permUpdatePerm := middleware.RequirePermission("cmdb:permission", "update")
	permDeletePerm := middleware.RequirePermission("cmdb:permission", "delete")
	cloudListPerm := middleware.RequirePermission("cmdb:cloud", "list")
	cloudGetPerm := middleware.RequirePermission("cmdb:cloud", "get")
	cloudCreatePerm := middleware.RequirePermission("cmdb:cloud", "create")
	cloudUpdatePerm := middleware.RequirePermission("cmdb:cloud", "update")
	cloudDeletePerm := middleware.RequirePermission("cmdb:cloud", "delete")
	cloudSyncPerm := middleware.RequirePermission("cmdb:cloud", "sync")
	fileBrowsePerm := middleware.RequirePermission("cmdb:file", "browse")
	fileUploadPerm := middleware.RequirePermission("cmdb:file", "upload")
	fileDeletePerm := middleware.RequirePermission("cmdb:file", "delete")
	fileAuditPerm := middleware.RequirePermission("cmdb:file", "audit")

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

			// 权限配置
			g.GET("/permission/list", permListPerm, api.PermissionList)
			g.GET("/permission/group-host-count", permListPerm, api.PermissionGroupHostCount)
			g.POST("/permission/create", permCreatePerm, middleware.SetAuditOperation("授予权限"), api.PermissionCreate)
			g.POST("/permission/update", permUpdatePerm, middleware.SetAuditOperation("更新权限"), api.PermissionUpdate)
			g.POST("/permission/delete", permDeletePerm, middleware.SetAuditOperation("删除权限"), api.PermissionDelete)
			g.GET("/permission/my-hosts", api.PermissionMyHosts)
			g.GET("/permission/check", api.PermissionCheck)

			// 云账号管理
			g.GET("/cloud-account/list", cloudListPerm, api.CloudAccountList)
			g.GET("/cloud-account/detail", cloudGetPerm, api.CloudAccountDetail)
			g.POST("/cloud-account/create", cloudCreatePerm, middleware.SetAuditOperation("创建云账号"), api.CloudAccountCreate)
			g.POST("/cloud-account/update", cloudUpdatePerm, middleware.SetAuditOperation("更新云账号"), api.CloudAccountUpdate)
			g.POST("/cloud-account/delete", cloudDeletePerm, middleware.SetAuditOperation("删除云账号"), api.CloudAccountDelete)
			g.POST("/cloud-account/sync", cloudSyncPerm, middleware.SetAuditOperation("同步云资源"), api.CloudAccountSync)
			g.GET("/cloud-account/resources", cloudListPerm, api.CloudResourceList)

			// 文件管理
			g.GET("/file/browse", fileBrowsePerm, api.FileBrowse)
			g.GET("/file/download", fileBrowsePerm, api.FileDownload)
			g.POST("/file/upload/:hostId", fileUploadPerm, middleware.SetAuditOperation("上传文件"), api.FileUpload)
			g.POST("/file/delete", fileDeletePerm, middleware.SetAuditOperation("删除文件"), api.FileDelete)
			g.POST("/file/rename", fileDeletePerm, middleware.SetAuditOperation("重命名文件"), api.FileRename)
			g.POST("/file/mkdir", fileUploadPerm, middleware.SetAuditOperation("创建目录"), api.FileMkdir)
			g.POST("/file/chmod", fileDeletePerm, middleware.SetAuditOperation("修改文件权限"), api.FileChmod)
			g.GET("/file/preview", fileBrowsePerm, api.FilePreview)
			g.POST("/file/edit", fileDeletePerm, middleware.SetAuditOperation("编辑文件"), api.FileEdit)
			g.POST("/file/distribute", fileUploadPerm, middleware.SetAuditOperation("批量分发文件"), api.FileDistribute)
			g.GET("/file/audit", fileAuditPerm, api.FileAuditList)
	}
}
