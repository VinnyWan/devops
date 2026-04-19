package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func SessionTagAdd(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	var req struct {
		SessionID uint   `json:"sessionId" binding:"required"`
		Tag       string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	svc := getTerminalService()
	if err := svc.AddTag(tenantID, req.SessionID, userID, req.Tag); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "标签已添加"})
}

func SessionTagRemove(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	var req struct {
		SessionID uint   `json:"sessionId" binding:"required"`
		Tag       string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	svc := getTerminalService()
	if err := svc.RemoveTag(tenantID, req.SessionID, req.Tag); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "标签已移除"})
}

func SessionTagList(c *gin.Context) {
	sessionID, _ := strconv.ParseUint(c.Query("sessionId"), 10, 64)
	if sessionID == 0 {
		c.JSON(400, gin.H{"code": 400, "message": "缺少 sessionId"})
		return
	}

	svc := getTerminalService()
	tags, err := svc.GetTagsForSession(uint(sessionID))
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": tags})
}

func SessionAvailableTags(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	svc := getTerminalService()
	tags, err := svc.GetAvailableTags(tenantID)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": tags})
}

func SessionSearchByTag(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	tag := c.Query("tag")
	if tag == "" {
		c.JSON(400, gin.H{"code": 400, "message": "缺少 tag"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	svc := getTerminalService()
	list, total, err := svc.SearchByTag(tenantID, tag, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": list, "total": total, "page": page, "pageSize": pageSize})
}
