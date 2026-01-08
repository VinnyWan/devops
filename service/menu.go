package service

import (
	"devops/internal/database"
	"devops/models"
	"errors"
)

type MenuService struct{}

// Create 创建菜单
func (s *MenuService) Create(menu *models.Menu) error {
	return database.Db.Create(menu).Error
}

// Update 更新菜单
func (s *MenuService) Update(id uint, menu *models.Menu) error {
	var existMenu models.Menu
	if err := database.Db.First(&existMenu, id).Error; err != nil {
		return errors.New("菜单不存在")
	}

	menu.ID = id
	return database.Db.Model(&existMenu).Updates(menu).Error
}

// Delete 删除菜单
func (s *MenuService) Delete(id uint) error {
	return database.Db.Delete(&models.Menu{}, id).Error
}

// GetByID 根据ID获取菜单
func (s *MenuService) GetByID(id uint) (*models.Menu, error) {
	var menu models.Menu
	if err := database.Db.First(&menu, id).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

// GetList 获取菜单列表
func (s *MenuService) GetList(menuName, status string) ([]models.Menu, error) {
	var menus []models.Menu

	query := database.Db.Model(&models.Menu{})

	if menuName != "" {
		query = query.Where("menu_name LIKE ?", "%"+menuName+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("sort ASC").Find(&menus).Error; err != nil {
		return nil, err
	}

	return menus, nil
}

// GetTreeList 获取菜单树形结构
func (s *MenuService) GetTreeList() ([]models.Menu, error) {
	var menus []models.Menu
	if err := database.Db.Order("sort ASC").Find(&menus).Error; err != nil {
		return nil, err
	}

	// 构建树形结构
	return buildMenuTree(menus, 0), nil
}

// buildMenuTree 构建菜单树
func buildMenuTree(menus []models.Menu, parentID uint) []models.Menu {
	var tree []models.Menu
	for _, menu := range menus {
		if menu.ParentID == parentID {
			tree = append(tree, menu)
		}
	}
	return tree
}
