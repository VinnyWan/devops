package model

import "time"

// UserDepartment 用户-部门多对多关联表，支持多部门归属
type UserDepartment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uk_user_dept" json:"userId"`
	DeptID    uint      `gorm:"not null;uniqueIndex:uk_user_dept" json:"deptId"`
	IsPrimary bool      `gorm:"not null;default:false" json:"isPrimary"`
	RoleID    *uint     `gorm:"index" json:"roleId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
