package repository

import (
	"devops-platform/internal/modules/user/model"
	queryutil "devops-platform/internal/pkg/query"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

// Create 创建用户
func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID 根据ID获取用户（含角色和部门基本信息）
func (r *UserRepo) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Roles").Preload("Department").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByIDWithPermissions 根据ID获取用户（含完整权限链，用于权限校验场景）
func (r *UserRepo) GetByIDWithPermissions(id uint) (*model.User, error) {
	var user model.User
	err := r.db.
		Preload("Roles").
		Preload("Roles.Permissions").
		Preload("Department").
		Preload("Department.Roles").
		Preload("Department.Roles.Permissions").
		First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Roles").Preload("Department").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepo) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表
func (r *UserRepo) List(page, pageSize int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{}).Preload("Roles")

	// 关键词搜索
	query = queryutil.ApplyKeywordLike(query, keyword, "username", "name", "email")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ListByDepartment 获取部门下的用户列表
func (r *UserRepo) ListByDepartment(deptID uint, page, pageSize int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{}).Preload("Roles").Where("department_id = ?", deptID)

	query = queryutil.ApplyKeywordLike(query, keyword, "username", "name", "email")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update 更新用户
func (r *UserRepo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateByID 根据ID更新用户部分字段（不会覆盖 created_at 等字段）
func (r *UserRepo) UpdateByID(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除用户（事务保护）
func (r *UserRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		user := &model.User{ID: id}
		if err := tx.Model(user).Association("Roles").Clear(); err != nil {
			return err
		}
		return tx.Delete(user).Error
	})
}

// UpdatePassword 更新密码
func (r *UserRepo) UpdatePassword(userID uint, hashedPassword string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

// UpdateLastLoginTime 更新最后登录时间
func (r *UserRepo) UpdateLastLoginTime(userID uint) error {
	now := gorm.Expr("NOW()")
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("last_login_at", now).Error
}

// AssignRoles 分配角色（事务保护）
func (r *UserRepo) AssignRoles(userID uint, roleIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		user := &model.User{ID: userID}
		if len(roleIDs) > 0 {
			var roles []model.Role
			if err := tx.Find(&roles, roleIDs).Error; err != nil {
				return err
			}
			return tx.Model(user).Association("Roles").Replace(roles)
		}
		return tx.Model(user).Association("Roles").Clear()
	})
}

func (r *UserRepo) ListUserIDsByDepartmentID(deptID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Model(&model.User{}).
		Where("department_id = ?", deptID).
		Pluck("id", &userIDs).Error
	return userIDs, err
}

func (r *UserRepo) ListPermissionAffectedUserIDsByRoleID(roleID uint) ([]uint, error) {
	var userIDs []uint

	directUserSubQuery := r.db.Table("user_roles").Select("user_id").Where("role_id = ?", roleID)
	deptSubQuery := r.db.Table("department_roles").Select("department_id").Where("role_id = ?", roleID)

	err := r.db.Model(&model.User{}).
		Distinct("users.id").
		Where("users.id IN (?) OR users.department_id IN (?)", directUserSubQuery, deptSubQuery).
		Pluck("users.id", &userIDs).Error
	return userIDs, err
}
