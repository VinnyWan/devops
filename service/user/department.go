package user

import (
	"devops/internal/database"
	usermodels "devops/models/user"
	"errors"
)

type DepartmentService struct{}

// Create 创建部门
func (s *DepartmentService) Create(dept *usermodels.Department) error {
	return database.Db.Create(dept).Error
}

// Update 更新部门
func (s *DepartmentService) Update(id uint, dept *usermodels.Department) error {
	var existDept usermodels.Department
	if err := database.Db.First(&existDept, id).Error; err != nil {
		return errors.New("部门不存在")
	}

	dept.ID = id
	return database.Db.Model(&existDept).Updates(dept).Error
}

// Delete 删除部门
func (s *DepartmentService) Delete(id uint) error {
	return database.Db.Delete(&usermodels.Department{}, id).Error
}

// GetByID 根据ID获取部门
func (s *DepartmentService) GetByID(id uint) (*usermodels.Department, error) {
	var dept usermodels.Department
	if err := database.Db.First(&dept, id).Error; err != nil {
		return nil, err
	}
	return &dept, nil
}

// GetList 获取部门列表
func (s *DepartmentService) GetList(deptName, status string) ([]usermodels.Department, error) {
	var depts []usermodels.Department

	query := database.Db.Model(&usermodels.Department{})

	if deptName != "" {
		query = query.Where("dept_name LIKE ?", "%"+deptName+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("sort ASC").Find(&depts).Error; err != nil {
		return nil, err
	}

	return depts, nil
}

// GetTreeList 获取部门树形结构
func (s *DepartmentService) GetTreeList() ([]usermodels.Department, error) {
	var depts []usermodels.Department
	if err := database.Db.Order("sort ASC").Find(&depts).Error; err != nil {
		return nil, err
	}

	return buildDeptTree(depts, 0), nil
}

// buildDeptTree 构建部门树
func buildDeptTree(depts []usermodels.Department, parentID uint) []usermodels.Department {
	var tree []usermodels.Department
	for _, dept := range depts {
		if dept.ParentID == parentID {
			tree = append(tree, dept)
		}
	}
	return tree
}
