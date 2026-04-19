package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetUserByID godoc
// @Summary 获取用户详情 (RESTful)
// @Description 根据路径 ID 获取用户详细信息
// @Tags 系统管理-用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/users/{id} [get]
func GetUserByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	user, err := getService().GetAccessibleUserByID(c.Request.Context(), tenantID, operatorID, uint(id))
	if err != nil {
		logger.Log.Error("Failed to get user detail", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": user})
}

// CreateUserREST godoc
// @Summary 创建用户 (RESTful)
// @Description 创建新用户（由管理员在系统管理中创建）
// @Tags 系统管理-用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/users [post]
func CreateUserREST(c *gin.Context) {
	// 复用 Register 逻辑，但不需要 tenantCode（从上下文获取）
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	user, err := getService().AdminCreateUser(c.Request.Context(), tenantID, operatorID, &req.Username, &req.Password, &req.Nickname, &req.Email, &req.Phone)
	if err != nil {
		logger.Log.Error("Failed to create user", zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": user})
}

// UpdateUserREST godoc
// @Summary 更新用户 (RESTful)
// @Description 根据路径 ID 更新用户信息
// @Tags 系统管理-用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body map[string]interface{} true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/users/{id} [put]
func UpdateUserREST(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	req["id"] = uint(id)

	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	if err := getService().UpdateUserByMap(c.Request.Context(), tenantID, operatorID, uint(id), req); err != nil {
		logger.Log.Error("Failed to update user", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteUserREST godoc
// @Summary 删除用户 (RESTful)
// @Description 根据路径 ID 删除用户
// @Tags 系统管理-用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/users/{id} [delete]
func DeleteUserREST(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	if err := getService().DeleteUser(c.Request.Context(), tenantID, operatorID, uint(id)); err != nil {
		logger.Log.Error("Failed to delete user", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// AssignRolesREST godoc
// @Summary 分配角色 (RESTful)
// @Description 为指定用户分配角色
// @Tags 系统管理-用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body map[string]interface{} true "{roleIds: [1,2,3]}"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/users/{id}/roles [put]
func AssignRolesREST(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	var req struct {
		RoleIDs []uint `json:"roleIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	if err := getService().AssignRoles(c.Request.Context(), tenantID, operatorID, uint(id), req.RoleIDs); err != nil {
		logger.Log.Error("Failed to assign roles", zap.Uint64("userID", id), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "角色分配成功"})
}

// ChangePasswordByID godoc
// @Summary 修改用户密码 (RESTful)
// @Description 修改指定用户的密码（用户自身操作）
// @Tags 系统管理-用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/users/{id}/password [put]
func ChangePasswordByID(c *gin.Context) {
	// 直接复用现有 ChangePassword，它内部通过 GetCurrentUserID 获取当前用户
	ChangePassword(c)
}

// ResetPasswordREST godoc
// @Summary 重置用户密码 (RESTful)
// @Description 管理员重置指定用户密码
// @Tags 系统管理-用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body map[string]interface{} true "{newPassword: 'xxx'}"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/users/{id}/reset-password [put]
func ResetPasswordREST(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	var req struct {
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := getService().ResetPassword(c.Request.Context(), tenantID, operatorID, uint(id), req.NewPassword); err != nil {
		logger.Log.Error("Failed to reset password", zap.Uint64("userID", id), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "密码重置成功"})
}

// LockUserREST godoc
// @Summary 锁定用户 (RESTful)
// @Description 锁定指定用户
// @Tags 系统管理-用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/users/{id}/lock [put]
func LockUserREST(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	if err := getService().LockUser(c.Request.Context(), tenantID, operatorID, uint(id)); err != nil {
		logger.Log.Error("Failed to lock user", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "用户锁定成功"})
}

// UnlockUserREST godoc
// @Summary 解锁用户 (RESTful)
// @Description 解锁指定用户
// @Tags 系统管理-用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/users/{id}/unlock [put]
func UnlockUserREST(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	if err := getService().UnlockUser(c.Request.Context(), tenantID, operatorID, uint(id)); err != nil {
		logger.Log.Error("Failed to unlock user", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "用户解锁成功"})
}
