package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/cicd/model"
	"devops-platform/internal/modules/cicd/service"
	"devops-platform/internal/pkg/obserr"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var cicdSvc *service.CICDService

func SetCICDDB(db *gorm.DB) {
	cicdSvc = service.NewCICDService(db)
}

// Jenkins Config
func ListJenkinsConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	configs, total, err := cicdSvc.ListConfigs(page, pageSize)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": configs, "total": total})
}

func SaveJenkinsConfig(c *gin.Context) {
	var cfg model.JenkinsConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := cicdSvc.SaveConfig(&cfg); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": cfg})
}

func DeleteJenkinsConfig(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := cicdSvc.DeleteConfig(uint(id)); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

func TestJenkinsConnection(c *gin.Context) {
	var req struct {
		URL      string `json:"url" binding:"required"`
		Username string `json:"username" binding:"required"`
		APIToken string `json:"apiToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := cicdSvc.TestConnection(req.URL, req.Username, req.APIToken); err != nil {
		writeObservableError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "connection successful"})
}

// Jobs
func ListJobs(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.Param("configId"), 10, 64)
	keyword := c.Query("keyword")
	jobs, err := cicdSvc.ListJobs(uint(configID), keyword)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": jobs})
}

func TriggerBuild(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.Param("configId"), 10, 64)
	var req struct {
		JobName string `json:"jobName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := cicdSvc.TriggerBuild(uint(configID), req.JobName); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "build triggered"})
}

func ListBuilds(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.Param("configId"), 10, 64)
	jobName := c.Query("jobName")
	builds, err := cicdSvc.ListBuilds(uint(configID), jobName)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": builds})
}

func GetBuildLog(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.Param("configId"), 10, 64)
	jobName := c.Query("jobName")
	buildNumber, _ := strconv.Atoi(c.Query("buildNumber"))
	logEntry, err := cicdSvc.GetBuildLog(uint(configID), jobName, buildNumber)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": logEntry})
}

// Pipelines
func ListPipelines(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	pipelines, total, err := cicdSvc.ListPipelines(page, pageSize)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": pipelines, "total": total})
}

func SavePipeline(c *gin.Context) {
	var p model.Pipeline
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := cicdSvc.SavePipeline(&p); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": p})
}

func DeletePipeline(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := cicdSvc.DeletePipeline(uint(id)); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

func writeObservableError(c *gin.Context, status int, err error) {
	details := obserr.Details(err)
	c.JSON(status, gin.H{"code": 500, "message": details["message"], "error": details})
}
