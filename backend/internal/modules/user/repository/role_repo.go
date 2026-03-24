package repository

import (
	"devops-platform/internal/modules/user/model"
	queryutil "devops-platform/internal/pkg/query"

	"gorm.io/gorm"
)

type RoleRepo struct {
	db *gorm.DB
}

func NewRoleRepo(db *gorm.DB) *RoleRepo {
	return &RoleRepo{
		db: db,
	}
}

// Create 创建角色
func (r *RoleRepo) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

// GetByID 根据ID获取角色
func (r *RoleRepo) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByName 根据名称获取角色
func (r *RoleRepo) GetByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	return &role, err
}

// List 获取角色列表
func (r *RoleRepo) List(page, pageSize int, keyword string) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := r.db.Model(&model.Role{}).Preload("Permissions")

	query = queryutil.ApplyKeywordLike(query, keyword, "name", "display_name", "description")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// Update 更新角色
func (r *RoleRepo) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

// Delete 删除角色（事务保护）
func (r *RoleRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		role := &model.Role{ID: id}
		if err := tx.Model(role).Association("Permissions").Clear(); err != nil {
			return err
		}
		if err := tx.Model(role).Association("Users").Clear(); err != nil {
			return err
		}
		if err := tx.Model(role).Association("Departments").Clear(); err != nil {
			return err
		}
		return tx.Delete(role).Error
	})
}

// AssignPermissions 分配权限（事务保护）
func (r *RoleRepo) AssignPermissions(roleID uint, permissionIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		role := &model.Role{ID: roleID}
		if len(permissionIDs) > 0 {
			var permissions []model.Permission
			if err := tx.Find(&permissions, permissionIDs).Error; err != nil {
				return err
			}
			return tx.Model(role).Association("Permissions").Replace(permissions)
		}
		return tx.Model(role).Association("Permissions").Clear()
	})
}

// AssignUsers 分配用户（事务保护）
func (r *RoleRepo) AssignUsers(roleID uint, userIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		role := &model.Role{ID: roleID}
		if len(userIDs) > 0 {
			var users []model.User
			if err := tx.Find(&users, userIDs).Error; err != nil {
				return err
			}
			return tx.Model(role).Association("Users").Replace(users)
		}
		return tx.Model(role).Association("Users").Clear()
	})
}

// AssignDepartments 分配部门（事务保护）
func (r *RoleRepo) AssignDepartments(roleID uint, departmentIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		role := &model.Role{ID: roleID}
		if len(departmentIDs) > 0 {
			var departments []model.Department
			if err := tx.Find(&departments, departmentIDs).Error; err != nil {
				return err
			}
			return tx.Model(role).Association("Departments").Replace(departments)
		}
		return tx.Model(role).Association("Departments").Clear()
	})
}

// GetRoleUsers 获取角色下的用户
func (r *RoleRepo) GetRoleUsers(roleID uint) ([]model.User, error) {
	var users []model.User
	err := r.db.Model(&model.Role{ID: roleID}).Association("Users").Find(&users)
	return users, err
}

// GetRoleDepartments 获取角色关联的部门
func (r *RoleRepo) GetRoleDepartments(roleID uint) ([]model.Department, error) {
	var departments []model.Department
	err := r.db.Model(&model.Role{ID: roleID}).Association("Departments").Find(&departments)
	return departments, err
}
