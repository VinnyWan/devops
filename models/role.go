package models

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色表
type Role struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	RoleName string `gorm:"type:varchar(50);uniqueIndex;not null" json:"roleName" binding:"required"`
	RoleKey  string `gorm:"type:varchar(50);uniqueIndex;not null" json:"roleKey" binding:"required"`
	Sort     int    `gorm:"type:int;default:0" json:"sort"`
	Status   int    `gorm:"type:tinyint;default:1;comment:状态:1正常,2禁用" json:"status"`
	Remark   string `gorm:"type:varchar(500)" json:"remark"`

	// 关联
	Menus []Menu `gorm:"many2many:role_menus;" json:"menus,omitempty"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "sys_role"
}
