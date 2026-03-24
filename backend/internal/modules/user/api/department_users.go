package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	deptUserService *service.DepartmentUserService
	deptUserOnce    sync.Once
)

func getDeptUserService() *service.DepartmentUserService {
	deptUserOnce.Do(func() {
		userRepo := repository.NewUserRepo(db)
		deptRepo := repository.NewDepartmentRepo(db)
		deptUserService = service.NewDepartmentUserService(userRepo, deptRepo)
	})
	return deptUserService
}

// ListDepartmentUsers godoc
// @Summary 获取部门用户列表
// @Description 根据部门ID查询部门下用户（非管理员只能查询本部门）
// @Tags 部门管理
// @Produce json
// @Security BearerAuth
// @Param departmentId query int false "部门ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词（用户名、姓名、邮箱）"
// @Success 200 {object} map[string]interface{} "用户列表"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Router /department/users/list [get]
func ListDepartmentUsers(c *gin.Context) {
	operatorID := GetCurrentUserID(c)
	if operatorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")

	deptID := uint(0)
	if deptIDStr := c.Query("departmentId"); deptIDStr != "" {
		id, err := strconv.ParseUint(deptIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
			return
		}
		deptID = uint(id)
	}

	list, total, err := getDeptUserService().List(operatorID, deptID, page, pageSize, keyword)
	if err != nil {
		writeDepartmentUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// CreateDepartmentUser godoc
// @Summary 创建部门用户
// @Description 在指定部门创建用户（非管理员只能在本部门创建）
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateDeptUserRequest true "用户信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Router /department/users/create [post]
func CreateDepartmentUser(c *gin.Context) {
	operatorID := GetCurrentUserID(c)
	if operatorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return
	}

	var req service.CreateDeptUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	user, err := getDeptUserService().Create(operatorID, &req)
	if err != nil {
		logger.Log.Error("Failed to create department user", zap.Error(err))
		writeDepartmentUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": user})
}

// UpdateDepartmentUser godoc
// @Summary 更新部门用户
// @Description 更新本部门用户信息（非管理员只能更新本部门用户）
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.UpdateDeptUserRequest true "用户信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Router /department/users/update [post]
func UpdateDepartmentUser(c *gin.Context) {
	operatorID := GetCurrentUserID(c)
	if operatorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return
	}

	var req service.UpdateDeptUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := getDeptUserService().Update(operatorID, &req); err != nil {
		logger.Log.Error("Failed to update department user", zap.Error(err))
		writeDepartmentUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功"})
}

// DeleteDepartmentUser godoc
// @Summary 删除部门用户
// @Description 删除本部门用户（非管理员只能删除本部门用户）
// @Tags 部门管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "用户ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Router /department/users/delete [post]
func DeleteDepartmentUser(c *gin.Context) {
	operatorID := GetCurrentUserID(c)
	if operatorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return
	}

	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	if err := getDeptUserService().Delete(operatorID, uint(id)); err != nil {
		logger.Log.Error("Failed to delete department user", zap.Error(err))
		writeDepartmentUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功"})
}

// TransferDepartmentUser godoc
// @Summary 用户切换部门
// @Description 将用户从A部门切换到B部门，并清空用户原有角色，继承新部门权限
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.TransferUserDepartmentRequest true "切换信息"
// @Success 200 {object} map[string]interface{} "切换成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未认证"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Router /department/users/transfer [post]
func TransferDepartmentUser(c *gin.Context) {
	operatorID := GetCurrentUserID(c)
	if operatorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return
	}

	var req service.TransferUserDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := getDeptUserService().Transfer(operatorID, &req); err != nil {
		logger.Log.Error("Failed to transfer department user", zap.Error(err))
		writeDepartmentUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功"})
}

func writeDepartmentUserError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status = http.StatusNotFound
	} else if strings.Contains(err.Error(), "permission denied") {
		status = http.StatusForbidden
	}
	c.JSON(status, gin.H{"code": status, "message": err.Error()})
}
