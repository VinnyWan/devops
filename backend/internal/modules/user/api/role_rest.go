package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetRoleByID godoc
// @Summary 获取角色详情 (RESTful)
// @Description 根据路径 ID 获取角色详细信息
// @Tags 系统管理-角色
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/roles/{id} [get]
func GetRoleByID(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}

	roleDetail, err := getRoleService().GetRoleByID(tenantID, uint(id))
	if err != nil {
		logger.Log.Error("Failed to get role detail", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "角色不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": roleDetail})
}

// UpdateRoleREST godoc
// @Summary 更新角色 (RESTful)
// @Description 根据路径 ID 更新角色信息
// @Tags 系统管理-角色
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Param request body map[string]interface{} true "角色信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/roles/{id} [put]
func UpdateRoleREST(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	req["id"] = uint(id)

	if err := getRoleService().UpdateRoleByMap(tenantID, uint(id), req); err != nil {
		logger.Log.Error("Failed to update role", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "角色更新成功"})
}

// DeleteRoleREST godoc
// @Summary 删除角色 (RESTful)
// @Description 根据路径 ID 删除角色
// @Tags 系统管理-角色
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/roles/{id} [delete]
func DeleteRoleREST(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}

	if err := getRoleService().DeleteRole(tenantID, uint(id)); err != nil {
		logger.Log.Error("Failed to delete role", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "角色删除成功"})
}

// AssignPermissionsREST godoc
// @Summary 分配权限 (RESTful)
// @Description 为指定角色分配权限
// @Tags 系统管理-角色
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Param request body map[string]interface{} true "{permissionIds: [1,2,3]}"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/roles/{id}/permissions [put]
func AssignPermissionsREST(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}

	var req struct {
		PermissionIDs []uint `json:"permissionIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := getRoleService().AssignPermissions(tenantID, uint(id), req.PermissionIDs); err != nil {
		logger.Log.Error("Failed to assign permissions", zap.Uint64("roleID", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "权限分配成功"})
}

// AssignRoleUsersREST godoc
// @Summary 角色关联用户 (RESTful)
// @Description 为指定角色绑定用户
// @Tags 系统管理-角色
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Param request body map[string]interface{} true "{userIds: [1,2,3]}"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/roles/{id}/users [put]
func AssignRoleUsersREST(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	operatorID := GetCurrentUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}

	var req struct {
		UserIDs []uint `json:"userIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := getRoleService().AssignUsers(c.Request.Context(), tenantID, operatorID, uint(id), req.UserIDs); err != nil {
		logger.Log.Error("Failed to assign role users", zap.Uint64("roleID", id), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "用户分配成功"})
}

// AssignRoleDepartmentsREST godoc
// @Summary 角色关联部门 (RESTful)
// @Description 为指定角色绑定部门
// @Tags 系统管理-角色
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Param request body map[string]interface{} true "{departmentIds: [1,2,3]}"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/roles/{id}/departments [put]
func AssignRoleDepartmentsREST(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	operatorID := GetCurrentUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}

	var req struct {
		DepartmentIDs []uint `json:"departmentIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := getRoleService().AssignDepartments(c.Request.Context(), tenantID, operatorID, uint(id), req.DepartmentIDs); err != nil {
		logger.Log.Error("Failed to assign role departments", zap.Uint64("roleID", id), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "部门分配成功"})
}

// GetRoleUsersREST godoc
// @Summary 获取角色下的用户 (RESTful)
// @Description 获取指定角色下的所有用户
// @Tags 系统管理-角色
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/roles/{id}/users [get]
func GetRoleUsersREST(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	operatorID := GetCurrentUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}

	users, err := getRoleService().GetRoleUsers(c.Request.Context(), tenantID, operatorID, uint(id))
	if err != nil {
		logger.Log.Error("Failed to get role users", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": users})
}

// GetRoleDepartmentsREST godoc
// @Summary 获取角色关联的部门 (RESTful)
// @Description 获取指定角色关联的所有部门
// @Tags 系统管理-角色
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/roles/{id}/departments [get]
func GetRoleDepartmentsREST(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}
	operatorID := GetCurrentUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}

	departments, err := getRoleService().GetRoleDepartments(c.Request.Context(), tenantID, operatorID, uint(id))
	if err != nil {
		logger.Log.Error("Failed to get role departments", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": departments})
}

// GetPermissionByID godoc
// @Summary 获取权限详情 (RESTful)
// @Description 根据路径 ID 获取权限详细信息
// @Tags 系统管理-权限
// @Produce json
// @Security BearerAuth
// @Param id path int true "权限ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/permissions/{id} [get]
func GetPermissionByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的权限ID"})
		return
	}

	permission, err := getRoleService().GetPermissionByID(uint(id))
	if err != nil {
		logger.Log.Error("Failed to get permission detail", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "权限不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": permission})
}

// UpdatePermissionREST godoc
// @Summary 更新权限 (RESTful)
// @Description 根据路径 ID 更新权限信息
// @Tags 系统管理-权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "权限ID"
// @Param request body model.Permission true "权限信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/permissions/{id} [put]
func UpdatePermissionREST(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的权限ID"})
		return
	}

	var permission model.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	permission.ID = uint(id)

	if err := getRoleService().UpdatePermission(&permission); err != nil {
		logger.Log.Error("Failed to update permission", zap.Uint("permissionID", permission.ID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "权限更新成功"})
}

// DeletePermissionREST godoc
// @Summary 删除权限 (RESTful)
// @Description 根据路径 ID 删除权限
// @Tags 系统管理-权限
// @Produce json
// @Security BearerAuth
// @Param id path int true "权限ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/permissions/{id} [delete]
func DeletePermissionREST(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的权限ID"})
		return
	}

	if err := getRoleService().DeletePermission(uint(id)); err != nil {
		logger.Log.Error("Failed to delete permission", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "权限删除成功"})
}
