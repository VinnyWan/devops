package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetTenantByID godoc
// @Summary 获取租户详情 (RESTful)
// @Description 根据路径 ID 获取租户详细信息
// @Tags 平台管理-租户
// @Produce json
// @Security BearerAuth
// @Param id path int true "租户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /platform/tenants/{id} [get]
func GetTenantByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的租户ID"})
		return
	}

	item, err := getTenantService().GetByID(uint(id))
	if err != nil {
		logger.Log.Error("Failed to get tenant detail", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "租户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": item})
}

// UpdateTenantREST godoc
// @Summary 更新租户 (RESTful)
// @Description 根据路径 ID 更新租户信息
// @Tags 平台管理-租户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "租户ID"
// @Param request body map[string]interface{} true "租户信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /platform/tenants/{id} [put]
func UpdateTenantREST(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的租户ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	req["id"] = uint(id)

	if err := getTenantService().UpdateByMap(uint(id), req); err != nil {
		logger.Log.Error("Failed to update tenant", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DisableTenantREST godoc
// @Summary 停用租户 (RESTful)
// @Description 根据路径 ID 停用租户
// @Tags 平台管理-租户
// @Produce json
// @Security BearerAuth
// @Param id path int true "租户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /platform/tenants/{id} [delete]
func DisableTenantREST(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的租户ID"})
		return
	}

	if err := getTenantService().Disable(uint(id)); err != nil {
		logger.Log.Error("Failed to disable tenant", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "停用成功"})
}
