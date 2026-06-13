package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/tool/service"

	"github.com/gin-gonic/gin"
)

var toolService *service.ToolService

func InitToolService(svc *service.ToolService) {
	toolService = svc
}

// ListTools returns all available tools, optionally filtered by category.
func ListTools(c *gin.Context) {
	tools, err := toolService.ListTools(c.Query("category"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": tools})
}

// GetTool returns a single tool by ID.
func GetTool(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	tool, err := toolService.GetTool(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "工具不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": tool})
}

// InstallTool installs a tool on the specified host via SSH.
func InstallTool(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	var req service.InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	inst, err := toolService.Install(tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": inst})
}

// CheckToolStatus checks installation status of a tool on a host.
func CheckToolStatus(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	toolID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		HostID       uint   `json:"hostId"`
		HostIP       string `json:"hostIp"`
		SSHPort      int    `json:"sshPort"`
		SSHUser      string `json:"sshUser"`
		SSHPassword  string `json:"sshPassword"`
		SSHKey       string `json:"sshKey"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	inst, err := toolService.CheckStatus(tenantID, uint(toolID), req.HostID, req.HostIP, req.SSHPort, req.SSHUser, req.SSHPassword, req.SSHKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": inst})
}

// ListInstallations returns installation records for a tenant.
func ListInstallations(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	hostID, _ := strconv.ParseUint(c.Query("hostId"), 10, 64)
	installs, err := toolService.ListInstallations(tenantID, uint(hostID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": installs})
}

func InstallToolRoutes(r *gin.RouterGroup) {
	r.GET("/tools", ListTools)
	r.GET("/tools/:id", GetTool)
	r.POST("/tools/:id/install", InstallTool)
	r.POST("/tools/:id/check", CheckToolStatus)
	r.GET("/tools/installations", ListInstallations)
}
