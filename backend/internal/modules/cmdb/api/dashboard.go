package api

import (
	"github.com/gin-gonic/gin"
)

func DashboardData(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	svc := getDashboardService()
	if svc == nil {
		c.JSON(500, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	data, err := svc.GetDashboard(tenantID, userID)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "获取仪表盘数据失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}
