package model

import (
	"time"

	"gorm.io/gorm"
)

// Category for knowledge base articles, tree structure
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	ParentID  *uint          `gorm:"index;default:null" json:"parentId"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Category) TableName() string { return "kb_categories" }

// Article represents a knowledge base article in Markdown
type Article struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:256;not null;index" json:"title"`
	CategoryID  *uint          `gorm:"index" json:"categoryId"`
	Content     string         `gorm:"type:longtext" json:"content"`
	ContentHTML string         `gorm:"type:longtext" json:"contentHtml"`
	ViewCount   int            `gorm:"default:0" json:"viewCount"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Article) TableName() string { return "kb_articles" }

// ArticleCreateRequest for article creation
type ArticleCreateRequest struct {
	Title      string `json:"title" binding:"required"`
	CategoryID *uint  `json:"categoryId"`
	Content    string `json:"content" binding:"required"`
}

// ArticleUpdateRequest for updating
type ArticleUpdateRequest struct {
	Title      string `json:"title"`
	CategoryID *uint  `json:"categoryId"`
	Content    string `json:"content"`
}

// ArticleListResponse with category name
type ArticleListResponse struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	CategoryID   *uint     `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	ViewCount    int       `json:"viewCount"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// CategoryTreeItem for tree structure response
type CategoryTreeItem struct {
	ID       uint               `json:"id"`
	Name     string             `json:"name"`
	ParentID *uint              `json:"parentId"`
	Children []CategoryTreeItem `json:"children,omitempty"`
}
