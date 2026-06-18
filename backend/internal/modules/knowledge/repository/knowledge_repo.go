package repository

import (
	"devops-platform/internal/modules/knowledge/model"
	"devops-platform/internal/pkg/obserr"
	"devops-platform/internal/pkg/query"

	"gorm.io/gorm"
)

const op = "knowledge/repository"

type KnowledgeRepo struct {
	db *gorm.DB
}

func NewKnowledgeRepo(db *gorm.DB) *KnowledgeRepo {
	return &KnowledgeRepo{db: db}
}

// --- Categories ---

func (r *KnowledgeRepo) ListCategories() ([]model.Category, error) {
	var categories []model.Category
	if err := r.db.Order("sort_order ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, obserr.Wrap("DB_ERROR", op, "list categories failed", err)
	}
	return categories, nil
}

func (r *KnowledgeRepo) GetCategory(id uint) (*model.Category, error) {
	var cat model.Category
	if err := r.db.First(&cat, id).Error; err != nil {
		return nil, obserr.Wrap("DB_ERROR", op, "get category failed", err)
	}
	return &cat, nil
}

func (r *KnowledgeRepo) SaveCategory(cat *model.Category) error {
	if err := r.db.Save(cat).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "save category failed", err)
	}
	return nil
}

func (r *KnowledgeRepo) DeleteCategory(id uint) error {
	// Check for children
	var childCount int64
	r.db.Model(&model.Category{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		return obserr.New("HAS_CHILDREN", op, "category has child categories, delete them first")
	}
	var articleCount int64
	r.db.Model(&model.Article{}).Where("category_id = ?", id).Count(&articleCount)
	if articleCount > 0 {
		return obserr.New("HAS_ARTICLES", op, "category contains articles, move or delete them first")
	}
	if err := r.db.Delete(&model.Category{}, id).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "delete category failed", err)
	}
	return nil
}

// --- Articles ---

func (r *KnowledgeRepo) ListArticles(categoryID *uint, keyword string, page, pageSize int) ([]model.ArticleListResponse, int64, error) {
	q := r.db.Model(&model.Article{}).Select("kb_articles.*, kb_categories.name as category_name").
		Joins("LEFT JOIN kb_categories ON kb_articles.category_id = kb_categories.id")

	if categoryID != nil && *categoryID > 0 {
		q = q.Where("kb_articles.category_id = ?", *categoryID)
	}
	if keyword != "" {
		q = query.ApplyKeywordLike(q, keyword, "kb_articles.title", "kb_articles.content")
	}

	var total int64
	q.Count(&total)

	var results []model.ArticleListResponse
	if err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("kb_articles.updated_at DESC").Find(&results).Error; err != nil {
		return nil, 0, obserr.Wrap("DB_ERROR", op, "list articles failed", err)
	}
	return results, total, nil
}

func (r *KnowledgeRepo) GetArticle(id uint) (*model.Article, error) {
	var article model.Article
	if err := r.db.First(&article, id).Error; err != nil {
		return nil, obserr.Wrap("DB_ERROR", op, "get article failed", err)
	}
	return &article, nil
}

func (r *KnowledgeRepo) CreateArticle(article *model.Article) error {
	if err := r.db.Create(article).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "create article failed", err)
	}
	return nil
}

func (r *KnowledgeRepo) UpdateArticle(article *model.Article) error {
	if err := r.db.Save(article).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "update article failed", err)
	}
	return nil
}

func (r *KnowledgeRepo) DeleteArticle(id uint) error {
	if err := r.db.Delete(&model.Article{}, id).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "delete article failed", err)
	}
	return nil
}

func (r *KnowledgeRepo) IncrementViewCount(id uint) error {
	if err := r.db.Model(&model.Article{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "increment view count failed", err)
	}
	return nil
}
