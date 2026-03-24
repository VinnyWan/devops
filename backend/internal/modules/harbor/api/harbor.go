package api

import (
	"net/http"

	"devops-platform/internal/modules/harbor/service"
	"devops-platform/internal/pkg/obserr"

	"github.com/gin-gonic/gin"
)

var harborService = service.NewHarborService()

// ListHarborProjects godoc
// @Summary 获取Harbor项目列表
// @Description 获取Harbor项目列表
// @Tags Harbor管理
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "查询失败"
// @Router /harbor/list [get]
func ListHarborProjects(c *gin.Context) {
	data, err := harborService.ListProjects(c.Query("keyword"))
	if err != nil {
		writeObservableError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

// ListHarborImages godoc
// @Summary 获取Harbor镜像列表
// @Description 获取Harbor项目下的镜像与标签
// @Tags Harbor管理
// @Produce json
// @Security BearerAuth
// @Param projectName query string false "项目名称"
// @Param keyword query string false "关键词（长度>=3生效，匹配仓库/tag/digest）"
// @Param repository query string false "兼容参数：仓库名称"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "查询失败"
// @Router /harbor/images [get]
func ListHarborImages(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		keyword = c.Query("repository")
	}
	data, err := harborService.ListImages(c.Query("projectName"), keyword)
	if err != nil {
		writeObservableError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

// GetHarborConfig godoc
// @Summary 获取Harbor配置
// @Description 获取Harbor模块配置
// @Tags Harbor管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /harbor/config [get]
func GetHarborConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    harborService.GetConfig(),
	})
}

// SaveHarborConfig godoc
// @Summary 保存Harbor配置
// @Description 创建或更新Harbor配置
// @Tags Harbor管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Harbor配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /harbor/config/upsert [post]
func SaveHarborConfig(c *gin.Context) {
	var req service.SaveHarborConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeObservableError(c, http.StatusBadRequest, obserr.Wrap("HARBOR_INVALID_REQUEST", "harbor.SaveHarborConfig", "参数错误", err))
		return
	}
	data, err := harborService.SaveConfig(req)
	if err != nil {
		writeObservableError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

func writeObservableError(c *gin.Context, status int, err error) {
	details := obserr.Details(err)
	msg, _ := details["message"].(string)
	code, _ := details["code"].(string)
	c.JSON(status, gin.H{
		"code":    status,
		"message": msg,
		"error": gin.H{
			"code":  code,
			"chain": details["chain"],
		},
	})
}
