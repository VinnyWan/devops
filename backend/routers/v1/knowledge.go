package v1

import (
	kbAPI "devops-platform/internal/modules/knowledge/api"
	"devops-platform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerKnowledgeBase(r *gin.RouterGroup) {
	g := r.Group("/kb")
	queryPermission := middleware.RequirePermission("knowledge", "list")
	updatePermission := middleware.RequirePermission("knowledge", "update")

	// Categories
	g.GET("/categories", queryPermission, kbAPI.ListCategories)
	g.POST("/categories", updatePermission,
		middleware.SetAuditOperation("创建知识库分类"),
		kbAPI.CreateCategory)
	g.PUT("/categories/:id", updatePermission, kbAPI.UpdateCategory)
	g.DELETE("/categories/:id", updatePermission,
		middleware.SetAuditOperation("删除知识库分类"),
		kbAPI.DeleteCategory)

	// Articles
	g.GET("/articles", queryPermission, kbAPI.ListArticles)
	g.GET("/articles/:id", queryPermission, kbAPI.GetArticle)
	g.POST("/articles", updatePermission,
		middleware.SetAuditOperation("创建知识库文章"),
		kbAPI.CreateArticle)
	g.PUT("/articles/:id", updatePermission, kbAPI.UpdateArticle)
	g.DELETE("/articles/:id", updatePermission,
		middleware.SetAuditOperation("删除知识库文章"),
		kbAPI.DeleteArticle)
}
