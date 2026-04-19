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

func (r *DepartmentRepo) GetByIDInTenant(tenantID uint, id uint) (*model.Department, error) {
	if err := requireTenantScope(tenantID); err != nil {
		return nil, err
	}
	var dept model.Department
	err := r.db.
		Preload("Roles", "(tenant_id = ? OR tenant_id IS NULL)", tenantID).
		Preload("Roles.Permissions").
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&dept).Error
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

func (r *DepartmentRepo) ListInTenant(tenantID uint, keyword string) ([]model.Department, error) {
	if err := requireTenantScope(tenantID); err != nil {
		return nil, err
	}
	var depts []model.Department
	query := r.db.Model(&model.Department{}).
		Preload("Roles", "(tenant_id = ? OR tenant_id IS NULL)", tenantID).
		Preload("Roles.Permissions").
		Where("tenant_id = ?", tenantID)
	query = queryutil.ApplyKeywordLike(query, keyword, "name")
	err := query.Find(&depts).Error
	return depts, err
}

func (r *DepartmentRepo) ListHierarchyInTenant(tenantID uint) ([]model.Department, error) {
	if err := requireTenantScope(tenantID); err != nil {
		return nil, err
	}
	var depts []model.Department
	err := r.db.Model(&model.Department{}).
		Select("id", "tenant_id", "name", "parent_id", "created_at", "updated_at").
		Where("tenant_id = ?", tenantID).
		Order("id ASC").
		Find(&depts).Error
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

func (r *DepartmentRepo) AssignRolesInTenant(tenantID uint, deptID uint, roleIDs []uint) error {
	if err := requireTenantScope(tenantID); err != nil {
		return err
	}
	var dept model.Department
	if err := r.db.Where("tenant_id = ? AND id = ?", tenantID, deptID).First(&dept).Error; err != nil {
		return err
	}

	var roles []model.Role
	if len(roleIDs) > 0 {
		if err := r.db.Where("(tenant_id = ? OR tenant_id IS NULL) AND id IN ?", tenantID, roleIDs).Find(&roles).Error; err != nil {
			return err
		}
	}

	return r.db.Model(&dept).Association("Roles").Replace(roles)
}

func (r *DepartmentRepo) DeleteInTenant(tenantID uint, id uint) error {
	if err := requireTenantScope(tenantID); err != nil {
		return err
	}
	return r.db.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&model.Department{}).Error
}

// GetDescendantIDsInTenant returns the IDs of a department and all its descendants within the tenant
func (r *DepartmentRepo) GetDescendantIDsInTenant(tenantID uint, deptID uint) ([]uint, error) {
	if err := requireTenantScope(tenantID); err != nil {
		return nil, err
	}
	// Load all departments in tenant (lightweight, only IDs and parent_id)
	var depts []struct {
		ID       uint
		ParentID *uint
	}
	if err := r.db.Model(&model.Department{}).
		Select("id, parent_id").
		Where("tenant_id = ?", tenantID).
		Find(&depts).Error; err != nil {
		return nil, err
	}

	// Build parent -> children map
	childrenMap := make(map[uint][]uint)
	for _, d := range depts {
		if d.ParentID != nil {
			childrenMap[*d.ParentID] = append(childrenMap[*d.ParentID], d.ID)
		}
	}

	// BFS to collect all descendant IDs
	var result []uint
	queue := []uint{deptID}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		result = append(result, curr)
		queue = append(queue, childrenMap[curr]...)
	}
	return result, nil
}
