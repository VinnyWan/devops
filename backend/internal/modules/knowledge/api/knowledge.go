package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/knowledge/model"
	"devops-platform/internal/modules/knowledge/service"
	"devops-platform/internal/pkg/obserr"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var kbSvc *service.KnowledgeService

func SetKnowledgeDB(db *gorm.DB) {
	kbSvc = service.NewKnowledgeService(db)
}

// --- Categories ---

func ListCategories(c *gin.Context) {
	tree, err := kbSvc.BuildCategoryTree()
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": tree})
}

func CreateCategory(c *gin.Context) {
	var cat model.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := kbSvc.SaveCategory(&cat); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": cat})
}

func UpdateCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var cat model.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	cat.ID = uint(id)
	if err := kbSvc.SaveCategory(&cat); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": cat})
}

func DeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := kbSvc.DeleteCategory(uint(id)); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

// --- Articles ---

func ListArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	categoryStr := c.Query("categoryId")
	var categoryID *uint
	if categoryStr != "" {
		if id, err := strconv.ParseUint(categoryStr, 10, 64); err == nil {
			cid := uint(id)
			categoryID = &cid
		}
	}
	articles, total, err := kbSvc.ListArticles(categoryID, keyword, page, pageSize)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": articles, "total": total, "page": page, "pageSize": pageSize})
}

func GetArticle(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	article, err := kbSvc.GetArticle(uint(id))
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": article})
}

func CreateArticle(c *gin.Context) {
	var req model.ArticleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	article, err := kbSvc.CreateArticle(&req)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": article})
}

func UpdateArticle(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req model.ArticleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	article, err := kbSvc.UpdateArticle(uint(id), &req)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": article})
}

func DeleteArticle(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := kbSvc.DeleteArticle(uint(id)); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

func writeObservableError(c *gin.Context, status int, err error) {
	details := obserr.Details(err)
	c.JSON(status, gin.H{"code": 500, "message": details["message"], "error": details})
}
