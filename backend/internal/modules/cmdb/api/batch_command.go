package api

import (
	"encoding/json"
	"time"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var batchUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func BatchCommandConnect(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	usernameValue, _ := c.Get("username")
	username, _ := usernameValue.(string)

	ws, err := batchUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	// Read the batch command request from WebSocket
	_, msg, err := ws.ReadMessage()
	if err != nil {
		return
	}

	var req service.BatchCommandRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		ws.WriteJSON(gin.H{"type": "error", "message": "请求格式错误"})
		return
	}

	svc := getBatchCommandService()
	if svc == nil {
		ws.WriteJSON(gin.H{"type": "error", "message": "服务未初始化"})
		return
	}

	// Execute commands and stream results
	resultCh := make(chan service.HostResult, len(req.HostIDs))
	ctx := c.Request.Context()

	svc.ExecuteOnHosts(ctx, tenantID, req, resultCh)

	var allResults []service.HostResult
	for result := range resultCh {
		allResults = append(allResults, result)
		ws.WriteJSON(gin.H{
			"type": "host_result",
			"data": result,
		})
	}

	// Send completion signal
	ws.WriteJSON(gin.H{
		"type":    "complete",
		"total":   len(req.HostIDs),
		"success": countSuccess(allResults),
		"failed":  len(allResults) - countSuccess(allResults),
	})

	// Create audit records in background
	go svc.CreateBatchAuditRecords(tenantID, userID, username, req, allResults)

	// Keep connection alive briefly for client to read final message
	time.Sleep(100 * time.Millisecond)
}

func countSuccess(results []service.HostResult) int {
	count := 0
	for _, r := range results {
		if r.Status == "success" {
			count++
		}
	}
	return count
}
