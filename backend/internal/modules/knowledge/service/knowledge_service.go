package service

import (
	"bytes"
	"html"

	"devops-platform/internal/modules/knowledge/model"
	"devops-platform/internal/modules/knowledge/repository"
	"devops-platform/internal/pkg/obserr"

	"github.com/yuin/goldmark"
	"gorm.io/gorm"
)

const op = "knowledge/service"

type KnowledgeService struct {
	repo *repository.KnowledgeRepo
	md   goldmark.Markdown
}

func NewKnowledgeService(db *gorm.DB) *KnowledgeService {
	return &KnowledgeService{
		repo: repository.NewKnowledgeRepo(db),
		md:   goldmark.New(),
	}
}

// --- Categories ---

func (s *KnowledgeService) ListCategories() ([]model.Category, error) {
	return s.repo.ListCategories()
}

func (s *KnowledgeService) BuildCategoryTree() ([]model.CategoryTreeItem, error) {
	cats, err := s.repo.ListCategories()
	if err != nil {
		return nil, err
	}

	// Build tree from flat list
	catMap := make(map[uint]*model.CategoryTreeItem)
	var roots []model.CategoryTreeItem

	for _, c := range cats {
		item := &model.CategoryTreeItem{ID: c.ID, Name: c.Name, ParentID: c.ParentID}
		catMap[c.ID] = item
	}
	for _, c := range cats {
		item := catMap[c.ID]
		if c.ParentID != nil && *c.ParentID > 0 {
			if parent, ok := catMap[*c.ParentID]; ok {
				parent.Children = append(parent.Children, *item)
			}
		} else {
			roots = append(roots, *item)
		}
	}
	if roots == nil {
		roots = []model.CategoryTreeItem{}
	}
	return roots, nil
}

func (s *KnowledgeService) SaveCategory(cat *model.Category) error {
	if cat.Name == "" {
		return obserr.New("INVALID_PARAM", op, "category name is required")
	}
	return s.repo.SaveCategory(cat)
}

func (s *KnowledgeService) DeleteCategory(id uint) error {
	return s.repo.DeleteCategory(id)
}

// --- Articles ---

func (s *KnowledgeService) ListArticles(categoryID *uint, keyword string, page, pageSize int) ([]model.ArticleListResponse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListArticles(categoryID, keyword, page, pageSize)
}

func (s *KnowledgeService) GetArticle(id uint) (*model.Article, error) {
	article, err := s.repo.GetArticle(id)
	if err != nil {
		return nil, err
	}
	s.repo.IncrementViewCount(id)
	return article, nil
}

func (s *KnowledgeService) CreateArticle(req *model.ArticleCreateRequest) (*model.Article, error) {
	if req.Title == "" {
		return nil, obserr.New("INVALID_PARAM", op, "title is required")
	}
	if req.Content == "" {
		return nil, obserr.New("INVALID_PARAM", op, "content is required")
	}

	htmlContent := s.renderMarkdown(req.Content)

	article := &model.Article{
		Title:       req.Title,
		CategoryID:  req.CategoryID,
		Content:     req.Content,
		ContentHTML: htmlContent,
	}
	if err := s.repo.CreateArticle(article); err != nil {
		return nil, err
	}
	return article, nil
}

func (s *KnowledgeService) UpdateArticle(id uint, req *model.ArticleUpdateRequest) (*model.Article, error) {
	article, err := s.repo.GetArticle(id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		article.Title = req.Title
	}
	if req.Content != "" {
		article.Content = req.Content
		article.ContentHTML = s.renderMarkdown(req.Content)
	}
	if req.CategoryID != nil {
		article.CategoryID = req.CategoryID
	}

	if err := s.repo.UpdateArticle(article); err != nil {
		return nil, err
	}
	return article, nil
}

func (s *KnowledgeService) DeleteArticle(id uint) error {
	return s.repo.DeleteArticle(id)
}

// renderMarkdown converts markdown to HTML using goldmark
func (s *KnowledgeService) renderMarkdown(content string) string {
	var buf bytes.Buffer
	if err := s.md.Convert([]byte(content), &buf); err != nil {
		// Fallback: escape HTML in raw content
		return html.EscapeString(content)
	}
	return buf.String()
}
