package api

import (
	"net/http"
	"strconv"
	"sync"

	"devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	userService *service.UserService
	userOnce    sync.Once
)

func getService() *service.UserService {
	userOnce.Do(func() {
		userService = service.NewUserService(db)
	})
	return userService
}

// Register godoc
// @Summary 用户注册
// @Description 注册新用户（仅限本地认证）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "注册信息"
// @Success 200 {object} map[string]interface{} "注册成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /user/register [post]
func Register(c *gin.Context) {
	var req struct {
		TenantCode string `json:"tenantCode" binding:"required"`
		service.RegisterRequest
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	tenant, err := getTenantService().GetByCode(req.TenantCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的租户编码",
		})
		return
	}
	if tenant.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "租户不可用",
		})
		return
	}

	newUser, err := getService().Register(tenant.ID, &req.RegisterRequest)
	if err != nil {
		logger.Log.Error("Failed to register user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "注册成功",
		"data":    newUser,
	})
}

// GetUserInfo godoc
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "用户信息"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Router /user/info [get]
func GetUserInfo(c *gin.Context) {
	// 从认证中间件获取用户ID
	userID := GetCurrentUserID(c)
	tenantID := GetCurrentTenantID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
		})
		return
	}

	user, err := getService().GetUserByID(c.Request.Context(), tenantID, userID)
	if err != nil {
		logger.Log.Error("Failed to get user info", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    user,
	})
}

// List godoc
// @Summary 获取用户列表
// @Description 获取用户列表（支持分页、关键词搜索和部门筛选）
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词（用户名、姓名、邮箱）"
// @Param departmentId query int false "部门ID（可选，按部门筛选用户）"
// @Success 200 {object} map[string]interface{} "用户列表"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /user/list [get]
func List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	// 解析可选的部门ID筛选参数
	var departmentID *uint
	if deptIDStr := c.Query("departmentId"); deptIDStr != "" {
		if id, err := strconv.ParseUint(deptIDStr, 10, 32); err == nil {
			deptID := uint(id)
			departmentID = &deptID
		}
	}

	users, total, err := getService().ListUsers(c.Request.Context(), tenantID, operatorID, page, pageSize, keyword, departmentID)
	if err != nil {
		logger.Log.Error("Failed to list users", zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  users,
			"total": total,
			"page":  page,
		},
	})
}

// GetDetail godoc
// @Summary 获取用户详情
// @Description 根据ID获取用户详细信息
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "用户ID"
// @Success 200 {object} map[string]interface{} "用户详情"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Router /user/detail [get]
func GetDetail(c *gin.Context) {
	idStr := c.Query("id")
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	user, err := getService().GetAccessibleUserByID(c.Request.Context(), tenantID, operatorID, uint(id))
	if err != nil {
		logger.Log.Error("Failed to get user detail", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    user,
	})
}

// Update godoc
// @Summary 更新用户信息
// @Description 更新用户的基本信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.UpdateUserRequest true "用户信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /user/update [post]
func Update(c *gin.Context) {
	var req service.UpdateUserRequest
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := getService().UpdateUserByRequest(c.Request.Context(), tenantID, operatorID, &req); err != nil {
		logger.Log.Error("Failed to update user", zap.Uint("userID", req.ID), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// Delete godoc
// @Summary 删除用户
// @Description 删除指定用户
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "用户ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /user/delete [post]
func Delete(c *gin.Context) {
	idStr := c.Query("id")
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	if err := getService().DeleteUser(c.Request.Context(), tenantID, operatorID, uint(id)); err != nil {
		logger.Log.Error("Failed to delete user", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// ChangePassword godoc
// @Summary 修改密码
// @Description 修改当前用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.ChangePasswordRequest true "密码修改信息"
// @Success 200 {object} map[string]interface{} "修改成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Router /user/change-password [post]
func ChangePassword(c *gin.Context) {
	// 从认证中间件获取用户ID
	userID := GetCurrentUserID(c)
	tenantID := GetCurrentTenantID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
		})
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := getService().ChangePassword(tenantID, userID, &req); err != nil {
		logger.Log.Error("Failed to change password", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码修改成功",
	})
}

// ResetPassword godoc
// @Summary 重置密码（管理员）
// @Description 管理员重置用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "重置信息 {userId: 1, newPassword: 'xxx'}"
// @Success 200 {object} map[string]interface{} "重置成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /user/reset-password [post]
func ResetPassword(c *gin.Context) {
	var req struct {
		UserID      uint   `json:"userId" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := getService().ResetPassword(c.Request.Context(), tenantID, operatorID, req.UserID, req.NewPassword); err != nil {
		logger.Log.Error("Failed to reset password", zap.Uint("userID", req.UserID), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码重置成功",
	})
}

// AssignRoles godoc
// @Summary 分配角色
// @Description 为用户分配角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "分配信息 {userId: 1, roleIds: [1,2,3]}"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /user/assign-roles [post]
func AssignRoles(c *gin.Context) {
	var req struct {
		UserID  uint   `json:"userId" binding:"required"`
		RoleIDs []uint `json:"roleIds" binding:"required"`
	}
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := getService().AssignRoles(c.Request.Context(), tenantID, operatorID, req.UserID, req.RoleIDs); err != nil {
		logger.Log.Error("Failed to assign roles", zap.Uint("userID", req.UserID), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "角色分配成功",
	})
}

// LockUser godoc
// @Summary 锁定用户
// @Description 锁定指定用户
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "用户ID"
// @Success 200 {object} map[string]interface{} "锁定成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /user/lock [post]
func LockUser(c *gin.Context) {
	idStr := c.Query("id")
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	if err := getService().LockUser(c.Request.Context(), tenantID, operatorID, uint(id)); err != nil {
		logger.Log.Error("Failed to lock user", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户锁定成功",
	})
}

// UnlockUser godoc
// @Summary 解锁用户
// @Description 解锁指定用户
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "用户ID"
// @Success 200 {object} map[string]interface{} "解锁成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /user/unlock [post]
func UnlockUser(c *gin.Context) {
	idStr := c.Query("id")
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	if err := getService().UnlockUser(c.Request.Context(), tenantID, operatorID, uint(id)); err != nil {
		logger.Log.Error("Failed to unlock user", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户解锁成功",
	})
}

// GetUserPermissions godoc
// @Summary 获取当前用户权限
// @Description 获取当前登录用户的权限编码集合
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "用户权限"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /user/permissions [get]
func GetUserPermissions(c *gin.Context) {
	userID := GetCurrentUserID(c)
	tenantID := GetCurrentTenantID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
		})
		return
	}

	permissions, err := getService().GetUserPermissionCodes(c.Request.Context(), tenantID, userID)
	if err != nil {
		logger.Log.Error("Failed to get user permissions", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    permissions,
	})
}

// GetAllPermissions godoc
// @Summary 获取当前用户全部权限（聚合）
// @Description 获取当前登录用户的菜单/按钮/字段/API 四合一权限
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "全部权限"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /user/all-permissions [get]
func GetAllPermissions(c *gin.Context) {
	userID := GetCurrentUserID(c)
	tenantID := GetCurrentTenantID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证",
		})
		return
	}

	permissions, err := getService().GetUserAllPermissions(c.Request.Context(), tenantID, userID)
	if err != nil {
		logger.Log.Error("Failed to get all permissions", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    permissions,
	})
}
