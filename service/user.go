package service

import (
	"devops/internal/database"
	"devops/models"
	"devops/utils"
	"errors"
)

type UserService struct{}

// Login 用户登录
func (s *UserService) Login(username, password string) (*models.User, string, error) {
	var user models.User
	if err := database.Db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 验证密码
	if !utils.CheckPassword(password, user.Password) {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, "", errors.New("用户已被禁用")
	}

	// 生成Token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", errors.New("生成Token失败")
	}

	return &user, token, nil
}

// Create 创建用户
func (s *UserService) Create(user *models.User) error {
	// 密码加密
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	return database.Db.Create(user).Error
}

// Update 更新用户
func (s *UserService) Update(id uint, user *models.User) error {
	var existUser models.User
	if err := database.Db.First(&existUser, id).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 如果修改了密码，需要加密
	if user.Password != "" {
		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	} else {
		user.Password = existUser.Password
	}

	user.ID = id
	return database.Db.Model(&existUser).Updates(user).Error
}

// UpdateByUsername 根据用户名更新用户
func (s *UserService) UpdateByUsername(username string, user *models.User) error {
	var existUser models.User
	if err := database.Db.Where("username = ?", username).First(&existUser).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 如果修改了密码，需要加密
	if user.Password != "" {
		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	} else {
		user.Password = existUser.Password
	}

	user.ID = existUser.ID
	return database.Db.Model(&existUser).Updates(user).Error
}

// Delete 删除用户
func (s *UserService) Delete(id uint) error {
	return database.Db.Delete(&models.User{}, id).Error
}

// DeleteByUsername 根据用户名删除用户
func (s *UserService) DeleteByUsername(username string) error {
	return database.Db.Where("username = ?", username).Delete(&models.User{}).Error
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := database.Db.Preload("Dept").Preload("Post").Preload("Roles").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (s *UserService) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := database.Db.Preload("Dept").Preload("Post").Preload("Roles").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetList 获取用户列表
func (s *UserService) GetList(page, pageSize int, username, phone, status string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := database.Db.Model(&models.User{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Dept").Preload("Post").Preload("Roles").
		Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// AssignRoles 分配角色
func (s *UserService) AssignRoles(userID uint, roleIDs []uint) error {
	var user models.User
	if err := database.Db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	var roles []models.Role
	if err := database.Db.Find(&roles, roleIDs).Error; err != nil {
		return err
	}

	return database.Db.Model(&user).Association("Roles").Replace(roles)
}
