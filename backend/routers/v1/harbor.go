package v1

import (
	"devops-platform/internal/middleware"
	harborAPI "devops-platform/internal/modules/harbor/api"

	"github.com/gin-gonic/gin"
)

func registerHarbor(r *gin.RouterGroup) {
	g := r.Group("/harbor")
	queryPermission := middleware.RequirePermission("harbor", "list")
	updatePermission := middleware.RequirePermission("harbor", "update")

	// Config
	g.GET("/configs", queryPermission, harborAPI.ListHarborConfigs)
	g.POST("/configs", updatePermission,
		middleware.SetAuditOperation("Harbor 配置保存"),
		harborAPI.SaveHarborConfig)
	g.PUT("/configs/:id", updatePermission, harborAPI.SaveHarborConfig)
	g.DELETE("/configs/:id", updatePermission,
		middleware.SetAuditOperation("Harbor 配置删除"),
		harborAPI.DeleteHarborConfig)
	g.POST("/configs/test", queryPermission, harborAPI.TestHarborConnection)

	// Projects
	g.GET("/projects", queryPermission, harborAPI.ListProjects)

	// Repositories
	g.GET("/projects/:projectName/repos", queryPermission, harborAPI.ListRepositories)

	// Artifacts
	g.GET("/projects/:projectName/repos/:repoName/artifacts", queryPermission, harborAPI.ListArtifacts)
	g.DELETE("/projects/:projectName/repos/:repoName/artifacts", updatePermission,
		middleware.SetAuditOperation("删除 Harbor artifact"),
		harborAPI.DeleteArtifact)
}
