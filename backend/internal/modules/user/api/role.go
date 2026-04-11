package api

import (
	"net/http"
	"strconv"
	"sync"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	roleService *service.RoleService
	roleOnce    sync.Once
)

func getRoleService() *service.RoleService {
	roleOnce.Do(func() {
		roleService = service.NewRoleService(db)
	})
	return roleService
}

func requireTenantID(c *gin.Context) (uint, bool) {
	tenantID := GetCurrentTenantID(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证：租户上下文缺失",
		})
		return 0, false
	}
	return tenantID, true
}

// CreateRole godoc
// @Summary 创建角色
// @Description 创建新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateRoleRequest true "角色信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /role/create [post]
func CreateRole(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	newRole, err := getRoleService().CreateRole(tenantID, &req)
	if err != nil {
		logger.Log.Error("Failed to create role", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "角色创建成功",
		"data":    newRole,
	})
}

// GetRoleDetail godoc
// @Summary 获取角色详情
// @Description 根据ID获取角色详细信息
// @Tags 角色管理
// @Produce json
// @Security BearerAuthAuth
// @Param id query int true "角色ID"
// @Success 200 {object} map[string]interface{} "角色详情"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /role/detail [get]
func GetRoleDetail(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的角色ID",
		})
		return
	}

	roleDetail, err := getRoleService().GetRoleByID(tenantID, uint(id))
	if err != nil {
		logger.Log.Error("Failed to get role detail", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "角色不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    roleDetail,
	})
}

// ListRoles godoc
// @Summary 获取角色列表
// @Description 获取角色列表（支持分页和关键词搜索）
// @Tags 角色管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} map[string]interface{} "角色列表"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /role/list [get]
func ListRoles(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")

	roles, total, err := getRoleService().ListRoles(tenantID, page, pageSize, keyword)
	if err != nil {
		logger.Log.Error("Failed to list roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取角色列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  roles,
			"total": total,
			"page":  page,
		},
	})
}

// UpdateRole godoc
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.UpdateRoleRequest true "角色信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /role/update [post]
func UpdateRole(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := getRoleService().UpdateRole(tenantID, &req); err != nil {
		logger.Log.Error("Failed to update role", zap.Uint("roleID", req.ID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "角色更新成功",
	})
}

// DeleteRole godoc
// @Summary 删除角色
// @Description 删除指定角色
// @Tags 角色管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "角色ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /role/delete [post]
func DeleteRole(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的角色ID",
		})
		return
	}

	if err := getRoleService().DeleteRole(tenantID, uint(id)); err != nil {
		logger.Log.Error("Failed to delete role", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "角色删除成功",
	})
}

// AssignPermissions godoc
// @Summary 分配权限
// @Description 为角色分配权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "分配信息 {roleId: 1, permissionIds: [1,2,3]}"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /role/assign-permissions [post]
func AssignPermissions(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req struct {
		RoleID        uint   `json:"roleId" binding:"required"`
		PermissionIDs []uint `json:"permissionIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := getRoleService().AssignPermissions(tenantID, req.RoleID, req.PermissionIDs); err != nil {
		logger.Log.Error("Failed to assign permissions", zap.Uint("roleID", req.RoleID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "权限分配成功",
	})
}

// GetRoleUsers godoc
// @Summary 获取角色下的用户
// @Description 获取指定角色下的所有用户
// @Tags 角色管理
// @Produce json
// @Security BearerAuthAuth
// @Param id query int true "角色ID"
// @Success 200 {object} map[string]interface{} "用户列表"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /role/users [get]
func GetRoleUsers(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	operatorID := GetCurrentUserID(c)

	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的角色ID",
		})
		return
	}

	users, err := getRoleService().GetRoleUsers(c.Request.Context(), tenantID, operatorID, uint(id))
	if err != nil {
		logger.Log.Error("Failed to get role users", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    users,
	})
}

// GetRoleDepartments godoc
// @Summary 获取角色关联的部门
// @Description 获取指定角色关联的所有部门
// @Tags 角色管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "角色ID"
// @Success 200 {object} map[string]interface{} "部门列表"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /role/departments [get]
func GetRoleDepartments(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	operatorID := GetCurrentUserID(c)

	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的角色ID",
		})
		return
	}

	departments, err := getRoleService().GetRoleDepartments(c.Request.Context(), tenantID, operatorID, uint(id))
	if err != nil {
		logger.Log.Error("Failed to get role departments", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    departments,
	})
}

// AssignRoleUsers godoc
// @Summary 角色关联用户
// @Description 为角色绑定用户（会覆盖原有绑定）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "分配信息 {roleId: 1, userIds: [1,2,3]}"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /role/assign-users [post]
func AssignRoleUsers(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	operatorID := GetCurrentUserID(c)

	var req struct {
		RoleID  uint   `json:"roleId" binding:"required"`
		UserIDs []uint `json:"userIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := getRoleService().AssignUsers(c.Request.Context(), tenantID, operatorID, req.RoleID, req.UserIDs); err != nil {
		logger.Log.Error("Failed to assign role users", zap.Uint("roleID", req.RoleID), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户分配成功",
	})
}

// AssignRoleDepartments godoc
// @Summary 角色关联部门
// @Description 为角色绑定部门（会覆盖原有绑定）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "分配信息 {roleId: 1, departmentIds: [1,2,3]}"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /role/assign-departments [post]
func AssignRoleDepartments(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	operatorID := GetCurrentUserID(c)

	var req struct {
		RoleID        uint   `json:"roleId" binding:"required"`
		DepartmentIDs []uint `json:"departmentIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := getRoleService().AssignDepartments(c.Request.Context(), tenantID, operatorID, req.RoleID, req.DepartmentIDs); err != nil {
		logger.Log.Error("Failed to assign role departments", zap.Uint("roleID", req.RoleID), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "部门分配成功",
	})
}

// CreatePermission godoc
// @Summary 创建权限
// @Description 创建新权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreatePermissionRequest true "权限信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /permission/create [post]
func CreatePermission(c *gin.Context) {
	var req service.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	newPerm, err := getRoleService().CreatePermission(&req)
	if err != nil {
		logger.Log.Error("Failed to create permission", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "权限创建成功",
		"data":    newPerm,
	})
}

// GetPermissionDetail godoc
// @Summary 获取权限详情
// @Description 根据ID获取权限详细信息
// @Tags 权限管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "权限ID"
// @Success 200 {object} map[string]interface{} "权限详情"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /permission/detail [get]
func GetPermissionDetail(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的权限ID",
		})
		return
	}

	permission, err := getRoleService().GetPermissionByID(uint(id))
	if err != nil {
		logger.Log.Error("Failed to get permission detail", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "权限不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    permission,
	})
}

// ListPermissions godoc
// @Summary 获取权限列表
// @Description 获取权限列表（支持分页、按资源过滤与关键词检索）
// @Tags 权限管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param resource query string false "资源名称"
// @Param keyword query string false "搜索关键词（长度>=3生效）"
// @Success 200 {object} map[string]interface{} "权限列表"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /permission/list [get]
func ListPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	resource := c.Query("resource")
	keyword := c.Query("keyword")

	permissions, total, err := getRoleService().ListPermissions(page, pageSize, resource, keyword)
	if err != nil {
		logger.Log.Error("Failed to list permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  permissions,
			"total": total,
			"page":  page,
		},
	})
}

// ListAllPermissions godoc
// @Summary 获取所有权限
// @Description 获取所有权限（不分页）
// @Tags 权限管理
// @Produce json
// @Security BearerAuthAuth
// @Success 200 {object} map[string]interface{} "所有权限"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /permission/all [get]
func ListAllPermissions(c *gin.Context) {
	permissions, err := getRoleService().ListAllPermissions()
	if err != nil {
		logger.Log.Error("Failed to list all permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    permissions,
	})
}

// UpdatePermission godoc
// @Summary 更新权限
// @Description 更新权限信息
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.Permission true "权限信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /permission/update [post]
func UpdatePermission(c *gin.Context) {
	var permission model.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := getRoleService().UpdatePermission(&permission); err != nil {
		logger.Log.Error("Failed to update permission", zap.Uint("permissionID", permission.ID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "权限更新成功",
	})
}

// DeletePermission godoc
// @Summary 删除权限
// @Description 删除指定权限
// @Tags 权限管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "权限ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /permission/delete [post]
func DeletePermission(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的权限ID",
		})
		return
	}

	if err := getRoleService().DeletePermission(uint(id)); err != nil {
		logger.Log.Error("Failed to delete permission", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "权限删除成功",
	})
}
