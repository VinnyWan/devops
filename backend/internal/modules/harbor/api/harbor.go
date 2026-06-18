package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/harbor/model"
	"devops-platform/internal/modules/harbor/service"
	"devops-platform/internal/pkg/obserr"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var harborSvc *service.HarborService

// SetHarborDB sets the database instance for the harbor module.
func SetHarborDB(db *gorm.DB) {
	harborSvc = service.NewHarborService(db)
}

// Configs

func ListHarborConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	configs, total, err := harborSvc.ListConfigs(page, pageSize)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": configs, "total": total})
}

func SaveHarborConfig(c *gin.Context) {
	var cfg model.HarborConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := harborSvc.SaveConfig(&cfg); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": cfg})
}

func DeleteHarborConfig(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := harborSvc.DeleteConfig(uint(id)); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

func TestHarborConnection(c *gin.Context) {
	var req struct {
		URL      string `json:"url" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := harborSvc.TestConnection(req.URL, req.Username, req.Password); err != nil {
		writeObservableError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "connection successful"})
}

// Projects

func ListProjects(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.DefaultQuery("configId", "0"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	projects, total, err := harborSvc.ListProjects(uint(configID), keyword, page, pageSize)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": projects, "total": total})
}

// Repositories

func ListRepositories(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.DefaultQuery("configId", "0"), 10, 64)
	projectName := c.Param("projectName")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	repos, total, err := harborSvc.ListRepositories(uint(configID), projectName, keyword, page, pageSize)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": repos, "total": total})
}

// Artifacts

func ListArtifacts(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.DefaultQuery("configId", "0"), 10, 64)
	projectName := c.Param("projectName")
	repoName := c.Param("repoName")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	artifacts, total, err := harborSvc.ListArtifacts(uint(configID), projectName, repoName, page, pageSize)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": artifacts, "total": total})
}

// Delete Artifact

func DeleteArtifact(c *gin.Context) {
	configID, _ := strconv.ParseUint(c.DefaultQuery("configId", "0"), 10, 64)
	projectName := c.Param("projectName")
	repoName := c.Param("repoName")
	var req struct {
		Reference string `json:"reference" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := harborSvc.DeleteArtifact(uint(configID), projectName, repoName, req.Reference); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

func writeObservableError(c *gin.Context, status int, err error) {
	details := obserr.Details(err)
	msg, _ := details["message"].(string)
	code, _ := details["code"].(string)
	c.JSON(status, gin.H{
		"code":    status,
		"message": msg,
		"error": gin.H{
			"code":  code,
			"chain": details["chain"],
		},
	})
}
