package user

import (
	"devops/internal/database"
	usermodels "devops/models/user"
	"errors"
)

type PostService struct{}

// Create 创建岗位
func (s *PostService) Create(post *usermodels.Post) error {
	return database.Db.Create(post).Error
}

// Update 更新岗位
func (s *PostService) Update(id uint, post *usermodels.Post) error {
	var existPost usermodels.Post
	if err := database.Db.First(&existPost, id).Error; err != nil {
		return errors.New("岗位不存在")
	}

	post.ID = id
	return database.Db.Model(&existPost).Updates(post).Error
}

// Delete 删除岗位
func (s *PostService) Delete(id uint) error {
	return database.Db.Delete(&usermodels.Post{}, id).Error
}

// GetByID 根据ID获取岗位
func (s *PostService) GetByID(id uint) (*usermodels.Post, error) {
	var post usermodels.Post
	if err := database.Db.First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

// GetList 获取岗位列表
func (s *PostService) GetList(page, pageSize int, postName, status string) ([]usermodels.Post, int64, error) {
	var posts []usermodels.Post
	var total int64

	query := database.Db.Model(&usermodels.Post{})

	if postName != "" {
		query = query.Where("post_name LIKE ?", "%"+postName+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("sort ASC").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}
