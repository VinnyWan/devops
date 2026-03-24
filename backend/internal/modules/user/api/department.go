package api

import (
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	deptService *service.DepartmentService
	deptOnce    sync.Once
)

func getDeptService() *service.DepartmentService {
	deptOnce.Do(func() {
		deptRepo := repository.NewDepartmentRepo(db)
		userRepo := repository.NewUserRepo(db)
		deptService = service.NewDepartmentService(deptRepo, userRepo)
	})
	return deptService
}

// CreateDepartment godoc
// @Summary 创建部门
// @Description 创建新部门 (Admin Only)
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateDepartmentRequest true "部门信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 403 {object} map[string]interface{} "权限不足"
// @Router /department/create [post]
func CreateDepartment(c *gin.Context) {
	var req service.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	dept, err := getDeptService().Create(c.Request.Context(), &req)
	if err != nil {
		logger.Log.Error("Failed to create department", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": dept})
}

// UpdateDepartment godoc
// @Summary 更新部门
// @Description 更新部门信息 (Admin Only)
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.UpdateDepartmentRequest true "部门信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Router /department/update [post]
func UpdateDepartment(c *gin.Context) {
	var req service.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := getDeptService().Update(c.Request.Context(), &req); err != nil {
		logger.Log.Error("Failed to update department", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功"})
}

// DeleteDepartment godoc
// @Summary 删除部门
// @Description 删除部门 (Admin Only)
// @Tags 部门管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "部门ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Router /department/delete [post]
func DeleteDepartment(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	if err := getDeptService().Delete(c.Request.Context(), uint(id)); err != nil {
		logger.Log.Error("Failed to delete department", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功"})
}

// ListDepartments godoc
// @Summary 获取部门列表
// @Description 获取部门树状列表
// @Tags 部门管理
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "搜索关键词（长度>=3生效）"
// @Success 200 {object} map[string]interface{} "部门列表"
// @Router /department/list [get]
func ListDepartments(c *gin.Context) {
	tree, err := getDeptService().GetTree(c.Request.Context(), c.Query("keyword"))
	if err != nil {
		logger.Log.Error("Failed to list departments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": tree})
}

// AssignDeptRoles godoc
// @Summary 部门分配角色
// @Description 给部门绑定角色 (Admin Only)
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "分配信息 {deptId: 1, roleIds: [1,2]}"
// @Success 200 {object} map[string]interface{} "分配成功"
// @Router /department/assign-roles [post]
func AssignDeptRoles(c *gin.Context) {
	var req struct {
		DeptID  uint   `json:"deptId" binding:"required"`
		RoleIDs []uint `json:"roleIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := getDeptService().AssignRoles(c.Request.Context(), req.DeptID, req.RoleIDs); err != nil {
		logger.Log.Error("Failed to assign department roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功"})
}
