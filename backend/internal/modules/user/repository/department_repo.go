package repository

import (
	"devops-platform/internal/modules/user/model"
	queryutil "devops-platform/internal/pkg/query"
	"gorm.io/gorm"
)

type DepartmentRepo struct {
	db *gorm.DB
}

func NewDepartmentRepo(db *gorm.DB) *DepartmentRepo {
	return &DepartmentRepo{db: db}
}

// Create 创建部门
func (r *DepartmentRepo) Create(dept *model.Department) error {
	return r.db.Create(dept).Error
}

// Update 更新部门
func (r *DepartmentRepo) Update(dept *model.Department) error {
	return r.db.Save(dept).Error
}

// Delete 删除部门
func (r *DepartmentRepo) Delete(id uint) error {
	return r.db.Delete(&model.Department{}, id).Error
}

// GetByID 根据ID获取部门
func (r *DepartmentRepo) GetByID(id uint) (*model.Department, error) {
	var dept model.Department
	err := r.db.Preload("Roles").Preload("Roles.Permissions").First(&dept, id).Error
	return &dept, err
}

// List 获取部门列表（扁平结构，需Service层组装树）
func (r *DepartmentRepo) List(keyword string) ([]model.Department, error) {
	var depts []model.Department
	query := r.db.Model(&model.Department{}).Preload("Roles").Preload("Roles.Permissions")
	query = queryutil.ApplyKeywordLike(query, keyword, "name")
	err := query.Find(&depts).Error
	return depts, err
}

// AssignRoles 部门分配角色
func (r *DepartmentRepo) AssignRoles(deptID uint, roleIDs []uint) error {
	var dept model.Department
	if err := r.db.First(&dept, deptID).Error; err != nil {
		return err
	}

	var roles []model.Role
	if err := r.db.Find(&roles, roleIDs).Error; err != nil {
		return err
	}

	return r.db.Model(&dept).Association("Roles").Replace(roles)
}
