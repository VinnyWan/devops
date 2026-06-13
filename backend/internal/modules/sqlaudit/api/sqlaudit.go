package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/sqlaudit/service"

	"github.com/gin-gonic/gin"
)

var sqlAuditService *service.SqlAuditService

func InitSqlAuditService(svc *service.SqlAuditService) {
	sqlAuditService = svc
}

// Connection management

type createConnReq struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Host        string `json:"host" binding:"required"`
	Port        int    `json:"port"`
	Database    string `json:"database"`
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Mode        string `json:"mode"`
	Description string `json:"description"`
}

func CreateConnection(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	var req createConnReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	conn, err := sqlAuditService.CreateConnection(tenantID, req.Name, req.Type, req.Host, req.Port, req.Database, req.Username, req.Password, req.Mode, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": conn})
}

func ListConnections(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	conns, err := sqlAuditService.ListConnections(tenantID, c.Query("type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": conns})
}

func TestConnection(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := sqlAuditService.TestConnection(uint(id), tenantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "连接正常"})
}

func DeleteConnection(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := sqlAuditService.DeleteConnection(uint(id), tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已删除"})
}

// SQL execution

func ExecuteSQL(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	userID := c.GetUint("userID")
	clientIP := c.ClientIP()
	var req service.ExecuteSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	result, err := sqlAuditService.ExecuteSQL(tenantID, userID, clientIP, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// Audit records

func ListSqlRecords(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	connID, _ := strconv.ParseUint(c.Query("connectionId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	records, total, err := sqlAuditService.ListRecords(tenantID, uint(connID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"total": total, "items": records}})
}

func InstallSqlAuditRoutes(r *gin.RouterGroup) {
	r.POST("/db-connections", CreateConnection)
	r.GET("/db-connections", ListConnections)
	r.POST("/db-connections/:id/test", TestConnection)
	r.DELETE("/db-connections/:id", DeleteConnection)
	r.POST("/sql/execute", ExecuteSQL)
	r.GET("/sql/records", ListSqlRecords)
}
