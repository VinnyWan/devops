package api

import (
	"strconv"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
)

func SnippetList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")

	svc := getSnippetService()
	if svc == nil {
		c.JSON(500, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	if keyword != "" {
		list, err := svc.Search(tenantID, userID, keyword)
		if err != nil {
			c.JSON(500, gin.H{"code": 500, "message": err.Error()})
			return
		}
		c.JSON(200, gin.H{"code": 200, "data": list})
		return
	}

	list, total, err := svc.List(tenantID, userID, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": list, "total": total, "page": page, "pageSize": pageSize})
}

func SnippetCreate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	var req service.SnippetCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	svc := getSnippetService()
	snippet, err := svc.Create(tenantID, userID, req)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": snippet})
}

func SnippetUpdate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	var req service.SnippetUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	svc := getSnippetService()
	snippet, err := svc.Update(tenantID, userID, req)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": snippet})
}

func SnippetDelete(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		c.JSON(400, gin.H{"code": 400, "message": "缺少 id"})
		return
	}

	svc := getSnippetService()
	if err := svc.Delete(tenantID, userID, uint(id)); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "删除成功"})
}
