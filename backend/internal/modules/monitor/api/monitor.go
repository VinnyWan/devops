package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/monitor/model"
	"devops-platform/internal/modules/monitor/service"
	"devops-platform/internal/pkg/obserr"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var monitorSvc *service.MonitorService

// SetMonitorDB initializes the monitor service with a DB connection and seeds defaults
func SetMonitorDB(db *gorm.DB) {
	monitorSvc = service.NewMonitorService(db)
	monitorSvc.EnsureDefaults()
}

// ListPrometheusConfigs GET /api/v1/monitor/prometheus
func ListPrometheusConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	configs, total, err := monitorSvc.ListConfigs(page, pageSize)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": configs, "total": total, "page": page, "pageSize": pageSize})
}

// GetPrometheusConfig GET /api/v1/monitor/prometheus/:id
func GetPrometheusConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}
	cfg, err := monitorSvc.GetConfig(uint(id))
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": cfg})
}

// SavePrometheusConfig POST/PUT /api/v1/monitor/prometheus
func SavePrometheusConfig(c *gin.Context) {
	var cfg model.PrometheusConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}
	if err := monitorSvc.SaveConfig(&cfg); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": cfg})
}

// DeletePrometheusConfig DELETE /api/v1/monitor/prometheus/:id
func DeletePrometheusConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}
	if err := monitorSvc.DeleteConfig(uint(id)); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

// TestPrometheusConnection POST /api/v1/monitor/prometheus/test
func TestPrometheusConnection(c *gin.Context) {
	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := monitorSvc.TestConnection(req.Endpoint, req.Username, req.Password); err != nil {
		writeObservableError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "connection successful"})
}

// QueryHostMetrics GET /api/v1/monitor/host/metrics
func QueryHostMetrics(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.DefaultQuery("configId", "0"), 10, 64)
	hostIP := c.Query("hostIp")
	metric := c.Query("metric")
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")

	if hostIP == "" || metric == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "hostIp and metric are required"})
		return
	}
	result, err := monitorSvc.QueryHostMetrics(uint(configID), hostIP, metric, startTime, endTime)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// QueryPortStatus GET /api/v1/monitor/host/ports
func QueryPortStatus(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.DefaultQuery("configId", "0"), 10, 64)
	hostIP := c.Query("hostIp")
	ports := c.QueryArray("ports")
	if hostIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "hostIp is required"})
		return
	}
	result, err := monitorSvc.QueryPortStatus(uint(configID), hostIP, ports)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// QueryAgentStatus GET /api/v1/monitor/agent/status
func QueryAgentStatus(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.DefaultQuery("configId", "0"), 10, 64)
	hostIPs := c.QueryArray("hostIps")
	result, err := monitorSvc.QueryAgentStatus(uint(configID), hostIPs)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

func writeObservableError(c *gin.Context, status int, err error) {
	details := obserr.Details(err)
	c.JSON(status, gin.H{"code": 500, "message": details["message"], "error": details})
}
