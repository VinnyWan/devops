package api

import (
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/modules/user/service"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	loginLogService *service.LoginLogService
	loginLogOnce    sync.Once
)

func getLoginLogService() *service.LoginLogService {
	loginLogOnce.Do(func() {
		loginLogService = service.NewLoginLogService(repository.NewLoginLogRepo(db), nil)
	})
	return loginLogService
}

// ListLoginLogs godoc
// @Summary 获取登录日志列表
// @Description 按条件分页查询登录日志
// @Tags 登录日志
// @Produce json
// @Security BearerAuth
// @Param username query string false "用户名"
// @Param status query string false "登录状态 (success/failed)"
// @Param startAt query string false "开始时间 (RFC3339)"
// @Param endAt query string false "结束时间 (RFC3339)"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "查询失败"
// @Router /login-log/list [get]
func ListLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	req := service.LoginLogListRequest{
		Username: c.Query("username"),
		Status:   c.Query("status"),
		StartAt:  c.Query("startAt"),
		EndAt:    c.Query("endAt"),
		Page:     page,
		PageSize: pageSize,
	}

	logs, total, err := getLoginLogService().List(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "查询登录日志失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":     logs,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}
