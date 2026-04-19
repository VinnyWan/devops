package api

import (
	"errors"
	"sync"

	"devops-platform/internal/modules/cmdb/repository"
	"devops-platform/internal/modules/cmdb/service"
	userservice "devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	cmdbDB              *gorm.DB
	hostSvcInstance     *service.HostService
	groupSvcInstance    *service.GroupService
	credSvcInstance     *service.CredentialService
	terminalSvcInstance *service.TerminalService
	permSvcInstance     *service.PermissionService
	cloudSvcInstance    *service.CloudAccountService
	fileSvcInstance     *service.FileService
	dashboardSvcInstance *service.DashboardService
	cmdbOnce            sync.Once
	cmdbMu              sync.Mutex
)

func SetDB(database *gorm.DB) {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	cmdbDB = database
	hostSvcInstance = nil
	groupSvcInstance = nil
	credSvcInstance = nil
	terminalSvcInstance = nil
	permSvcInstance = nil
	cloudSvcInstance = nil
	fileSvcInstance = nil
	dashboardSvcInstance = nil
	cmdbOnce = sync.Once{}
}

func getHostService() *service.HostService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if hostSvcInstance != nil {
		return hostSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	hostSvcInstance = service.NewHostService(cmdbDB)
	return hostSvcInstance
}

func getGroupService() *service.GroupService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if groupSvcInstance != nil {
		return groupSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	groupSvcInstance = service.NewGroupService(cmdbDB)
	return groupSvcInstance
}

func getCredentialService() *service.CredentialService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if credSvcInstance != nil {
		return credSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	credSvcInstance = service.NewCredentialService(cmdbDB)
	return credSvcInstance
}

func getTerminalService() *service.TerminalService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if terminalSvcInstance != nil {
		return terminalSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	terminalSvcInstance = service.NewTerminalService(cmdbDB)
	return terminalSvcInstance
}

func getPermissionService() *service.PermissionService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if permSvcInstance != nil {
		return permSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	permSvcInstance = service.NewPermissionService(cmdbDB)
	return permSvcInstance
}

func getCloudAccountService() *service.CloudAccountService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if cloudSvcInstance != nil {
		return cloudSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	cloudSvcInstance = service.NewCloudAccountService(cmdbDB)
	return cloudSvcInstance
}

func getFileService() *service.FileService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if fileSvcInstance != nil {
		return fileSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	fileSvcInstance = service.NewFileService(cmdbDB)
	return fileSvcInstance
}

func getDashboardService() *service.DashboardService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if dashboardSvcInstance != nil {
		return dashboardSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	repo := repository.NewDashboardRepo(cmdbDB)
	dashboardSvcInstance = service.NewDashboardService(repo)
	return dashboardSvcInstance
}

func getCurrentTenantID(c *gin.Context) (uint, error) {
	tenantIDValue, exists := c.Get("tenantID")
	if !exists {
		return 0, errors.New("未认证：无法获取租户信息")
	}
	tenantID, ok := tenantIDValue.(uint)
	if !ok || tenantID == 0 {
		return 0, errors.New("未认证：租户信息无效")
	}
	return tenantID, nil
}

// isCmdbAdmin checks if user has cmdb:host:admin permission (admin bypasses host-level filtering)
func isCmdbAdmin(c *gin.Context, tenantID, userID uint) bool {
	userSvc := userservice.NewUserService(cmdbDB)
	isAdmin, err := userSvc.CheckPermission(c.Request.Context(), tenantID, userID, "cmdb:host", "admin")
	if err != nil {
		logger.Log.Error("检查 CMDB 管理员权限失败", zap.Error(err))
		return false
	}
	return isAdmin
}

// StartCloudSync starts the scheduled cloud sync task (called from main.go)
func StartCloudSync() {
	svc := getCloudAccountService()
	svc.StartScheduledSync()
	logger.Log.Info("定时云同步已启动")
}

// StartRecordingCleanup starts the recording cleanup scheduler (called from main.go)
func StartRecordingCleanup() {
	svc := service.NewRecordingCleanupService(cmdbDB)
	svc.StartCleanupScheduler()
}
