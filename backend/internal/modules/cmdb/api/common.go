package api

import (
	"errors"
	"sync"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	cmdbDB              *gorm.DB
	hostSvcInstance     *service.HostService
	groupSvcInstance    *service.GroupService
	credSvcInstance     *service.CredentialService
	terminalSvcInstance *service.TerminalService
	permSvcInstance     *service.PermissionService
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
