package service

import (
	"devops/internal/database"
	"devops/models"
	"errors"
)

type RoleService struct{}

// Create 创建角色
func (s *RoleService) Create(role *models.Role) error {
	return database.Db.Create(role).Error
}

// Update 更新角色
func (s *RoleService) Update(id uint, role *models.Role) error {
	var existRole models.Role
	if err := database.Db.First(&existRole, id).Error; err != nil {
		return errors.New("角色不存在")
	}

	role.ID = id
	return database.Db.Model(&existRole).Updates(role).Error
}

// Delete 删除角色
func (s *RoleService) Delete(id uint) error {
	return database.Db.Delete(&models.Role{}, id).Error
}

// GetByID 根据ID获取角色
func (s *RoleService) GetByID(id uint) (*models.Role, error) {
	var role models.Role
	if err := database.Db.Preload("Menus").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// GetList 获取角色列表
func (s *RoleService) GetList(page, pageSize int, roleName, status string) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	query := database.Db.Model(&models.Role{})

	if roleName != "" {
		query = query.Where("role_name LIKE ?", "%"+roleName+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Menus").Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// AssignMenus 分配菜单权限
func (s *RoleService) AssignMenus(roleID uint, menuIDs []uint) error {
	var role models.Role
	if err := database.Db.First(&role, roleID).Error; err != nil {
		return errors.New("角色不存在")
	}

	var menus []models.Menu
	if err := database.Db.Find(&menus, menuIDs).Error; err != nil {
		return err
	}

	return database.Db.Model(&role).Association("Menus").Replace(menus)
}
