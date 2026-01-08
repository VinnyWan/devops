package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username" binding:"required"`
	Password string `gorm:"type:varchar(255);not null" json:"-"`
	Nickname string `gorm:"type:varchar(50)" json:"nickname"`
	Email    string `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	Phone    string `gorm:"type:varchar(20);uniqueIndex" json:"phone"`
	Avatar   string `gorm:"type:varchar(255)" json:"avatar"`
	Status   int    `gorm:"type:tinyint;default:1;comment:状态:1正常,2禁用" json:"status"`
	Gender   int    `gorm:"type:tinyint;default:0;comment:性别:0未知,1男,2女" json:"gender"`
	DeptID   uint   `gorm:"comment:部门ID" json:"deptId"`
	PostID   uint   `gorm:"comment:岗位ID" json:"postId"`
	Remark   string `gorm:"type:varchar(500)" json:"remark"`

	// 关联
	Dept  *Department `gorm:"foreignKey:DeptID" json:"dept,omitempty"`
	Post  *Post       `gorm:"foreignKey:PostID" json:"post,omitempty"`
	Roles []Role      `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "sys_user"
}
