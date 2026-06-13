package api

import (
	"fmt"
	"net/http"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var apiKeyService *service.ApiKeyService

func InitApiKeyService(db *gorm.DB) {
	apiKeyService = service.NewApiKeyService(db)
}

// CreateApiKey godoc
// @Summary 创建 API Key
// @Description 为当前用户创建一个新的 API Key，创建后仅展示一次完整 Key
// @Tags 用户认证
// @Produce json
// @Security BearerAuth
// @Param body body model.CreateApiKeyRequest true "API Key 信息"
// @Success 200 {object} model.ApiKeyResponse "成功，key 字段仅在此时返回"
// @Router /user/api-keys [post]
func CreateApiKey(c *gin.Context) {
	userID := c.GetUint("userID")
	tenantID := c.GetUint("tenantID")

	var req model.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	key, err := apiKeyService.Create(userID, tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": key})
}

// ListApiKeys godoc
// @Summary 获取 API Key 列表
// @Tags 用户认证
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /user/api-keys [get]
func ListApiKeys(c *gin.Context) {
	userID := c.GetUint("userID")
	tenantID := c.GetUint("tenantID")

	var req model.ListApiKeyRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	keys, total, err := apiKeyService.List(userID, tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"total": total,
			"items": keys,
		},
	})
}

// DeleteApiKey godoc
// @Summary 删除 API Key
// @Tags 用户认证
// @Produce json
// @Security BearerAuth
// @Param id path int true "API Key ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /user/api-keys/{id} [delete]
func DeleteApiKey(c *gin.Context) {
	userID := c.GetUint("userID")
	tenantID := c.GetUint("tenantID")

	id := c.Param("id")
	var keyID uint
	if _, err := fmt.Sscanf(id, "%d", &keyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的 ID"})
		return
	}

	if err := apiKeyService.Delete(keyID, userID, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已删除"})
}

// InstallApiKeyRoutes registers API Key endpoints on the given router group.
func InstallApiKeyRoutes(r *gin.RouterGroup) {
	r.POST("/api-keys", CreateApiKey)
	r.GET("/api-keys", ListApiKeys)
	r.DELETE("/api-keys/:id", DeleteApiKey)
}
