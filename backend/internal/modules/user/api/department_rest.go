package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UpdateDepartmentREST godoc
// @Summary 更新部门 (RESTful)
// @Description 根据路径 ID 更新部门信息
// @Tags 系统管理-部门
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Param request body map[string]interface{} true "部门信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/departments/{id} [put]
func UpdateDepartmentREST(c *gin.Context) {
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	req["id"] = uint(id)

	if err := getDeptService().UpdateByMap(c.Request.Context(), tenantID, operatorID, uint(id), req); err != nil {
		logger.Log.Error("Failed to update department", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteDepartmentREST godoc
// @Summary 删除部门 (RESTful)
// @Description 根据路径 ID 删除部门
// @Tags 系统管理-部门
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/departments/{id} [delete]
func DeleteDepartmentREST(c *gin.Context) {
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	if err := getDeptService().Delete(c.Request.Context(), tenantID, operatorID, uint(id)); err != nil {
		logger.Log.Error("Failed to delete department", zap.Uint64("id", id), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// AssignDeptRolesREST godoc
// @Summary 部门分配角色 (RESTful)
// @Description 为指定部门绑定角色
// @Tags 系统管理-部门
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Param request body map[string]interface{} true "{roleIds: [1,2]}"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/departments/{id}/roles [put]
func AssignDeptRolesREST(c *gin.Context) {
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	deptID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	var req struct {
		RoleIDs []uint `json:"roleIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := getDeptService().AssignRoles(c.Request.Context(), tenantID, operatorID, uint(deptID), req.RoleIDs); err != nil {
		logger.Log.Error("Failed to assign department roles", zap.Uint64("deptID", deptID), zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "角色分配成功"})
}

// ListDepartmentUsersREST godoc
// @Summary 获取部门用户列表 (RESTful)
// @Description 根据路径部门 ID 查询该部门用户
// @Tags 系统管理-部门
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/departments/{id}/users [get]
func ListDepartmentUsersREST(c *gin.Context) {
	operatorID := GetCurrentUserID(c)
	tenantID := GetCurrentTenantID(c)
	if operatorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return
	}

	deptID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")

	list, total, err := getDeptUserService().List(tenantID, operatorID, uint(deptID), page, pageSize, keyword)
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

// CreateDepartmentUserREST godoc
// @Summary 创建部门用户 (RESTful)
// @Description 在指定部门创建用户
// @Tags 系统管理-部门
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Param request body map[string]interface{} true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/departments/{id}/users [post]
func CreateDepartmentUserREST(c *gin.Context) {
	operatorID := GetCurrentUserID(c)
	tenantID := GetCurrentTenantID(c)
	if operatorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return
	}

	deptID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	req["departmentId"] = uint(deptID)

	user, err := getDeptUserService().CreateByMap(tenantID, operatorID, uint(deptID), req)
	if err != nil {
		logger.Log.Error("Failed to create department user", zap.Error(err))
		writeDepartmentUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": user})
}

// UpdateDepartmentUserREST godoc
// @Summary 更新部门用户 (RESTful)
// @Description 更新指定部门的指定用户信息
// @Tags 系统管理-部门
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Param uid path int true "用户ID"
// @Param request body map[string]interface{} true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/departments/{id}/users/{uid} [put]
func UpdateDepartmentUserREST(c *gin.Context) {
	operatorID := GetCurrentUserID(c)
	tenantID := GetCurrentTenantID(c)
	if operatorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return
	}

	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	uid, err := strconv.ParseUint(c.Param("uid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	req["id"] = uint(uid)

	if err := getDeptUserService().UpdateByMap(tenantID, operatorID, uint(uid), req); err != nil {
		logger.Log.Error("Failed to update department user", zap.Uint64("uid", uid), zap.Error(err))
		writeDepartmentUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteDepartmentUserREST godoc
// @Summary 删除部门用户 (RESTful)
// @Description 删除指定部门的指定用户
// @Tags 系统管理-部门
// @Produce json
// @Security BearerAuth
// @Param id path int true "部门ID"
// @Param uid path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /system/departments/{id}/users/{uid} [delete]
func DeleteDepartmentUserREST(c *gin.Context) {
	operatorID := GetCurrentUserID(c)
	tenantID := GetCurrentTenantID(c)
	if operatorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return
	}

	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的部门ID"})
		return
	}

	uid, err := strconv.ParseUint(c.Param("uid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	if err := getDeptUserService().Delete(tenantID, operatorID, uint(uid)); err != nil {
		logger.Log.Error("Failed to delete department user", zap.Uint64("uid", uid), zap.Error(err))
		writeDepartmentUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
