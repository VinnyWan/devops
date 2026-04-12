package repository

import (
	"devops-platform/internal/modules/user/model"

	"gorm.io/gorm"
)

type UserDepartmentRepo struct {
	db *gorm.DB
}

func NewUserDepartmentRepo(db *gorm.DB) *UserDepartmentRepo {
	return &UserDepartmentRepo{db: db}
}

// AssignUserToDepartments 将用户分配到多个部门
func (r *UserDepartmentRepo) AssignUserToDepartments(userID uint, deptIDs []uint, primaryDeptID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserDepartment{}).Error; err != nil {
			return err
		}
		for _, deptID := range deptIDs {
			ud := model.UserDepartment{
				UserID:    userID,
				DeptID:    deptID,
				IsPrimary: deptID == primaryDeptID,
			}
			if err := tx.Create(&ud).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetUserDepartments 获取用户的所有部门关联
func (r *UserDepartmentRepo) GetUserDepartments(userID uint) ([]model.UserDepartment, error) {
	var uds []model.UserDepartment
	err := r.db.Where("user_id = ?", userID).Find(&uds).Error
	return uds, err
}

// GetPrimaryDepartment 获取用户的主部门
func (r *UserDepartmentRepo) GetPrimaryDepartment(userID uint) (*model.UserDepartment, error) {
	var ud model.UserDepartment
	err := r.db.Where("user_id = ? AND is_primary = ?", userID, true).First(&ud).Error
	if err != nil {
		return nil, err
	}
	return &ud, nil
}

// GetDepartmentUsers 获取部门下的所有用户关联
func (r *UserDepartmentRepo) GetDepartmentUsers(deptID uint) ([]model.UserDepartment, error) {
	var uds []model.UserDepartment
	err := r.db.Where("dept_id = ?", deptID).Find(&uds).Error
	return uds, err
}

// TransferDepartment 用户跨部门调动
func (r *UserDepartmentRepo) TransferDepartment(userID uint, fromDeptID, toDeptID uint, isPrimary bool) error {
	return r.db.Model(&model.UserDepartment{}).
		Where("user_id = ? AND dept_id = ?", userID, fromDeptID).
		Updates(map[string]interface{}{
			"dept_id":    toDeptID,
			"is_primary": isPrimary,
		}).Error
}
