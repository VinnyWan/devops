package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/tool/model"
	"devops-platform/internal/modules/tool/service"
	"devops-platform/internal/pkg/obserr"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var toolService *service.ToolService

func InitToolService(svc *service.ToolService) {
	toolService = svc
}

func SetToolDB(db *gorm.DB) {
	toolService = service.NewToolService(db)
	toolService.SeedDefaultTemplates()
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

// Templates
func ListTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	category := c.Query("category")
	templates, total, err := toolService.ListTemplates(category, page, pageSize)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": templates, "total": total})
}

func GetTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	t, err := toolService.GetTemplate(uint(id))
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": t})
}

func SaveTemplate(c *gin.Context) {
	var t model.ToolTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := toolService.SaveTemplate(&t); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": t})
}

func DeleteTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := toolService.DeleteTemplate(uint(id)); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

// Versions
func ListTemplateVersions(c *gin.Context) {
	templateID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	versions, err := toolService.ListVersions(uint(templateID))
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": versions})
}

func SaveTemplateVersion(c *gin.Context) {
	templateID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var v model.ToolTemplateVersion
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	v.TemplateID = uint(templateID)
	if err := toolService.SaveVersion(&v); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": v})
}

func DeleteTemplateVersion(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("versionId"), 10, 64)
	if err := toolService.DeleteVersion(uint(versionID)); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

func writeObservableError(c *gin.Context, status int, err error) {
	if details := obserr.Details(err); details != nil {
		c.JSON(status, gin.H{"code": 500, "message": details["message"], "error": details})
		return
	}
	c.JSON(status, gin.H{"code": 500, "message": err.Error()})
}
