package model

import (
	"time"

	"gorm.io/gorm"
)

// Department 部门模型
type Department struct {
	ID        uint          `gorm:"primarykey" json:"id"`
	Name      string        `gorm:"size:100;not null" json:"name"`
	ParentID  *uint         `json:"parentId"` // 上级部门ID，根部门为NULL
	Roles     []Role        `gorm:"many2many:department_roles;" json:"roles,omitempty"`
	Users     []User        `gorm:"foreignKey:DepartmentID" json:"users,omitempty"`
	Children  []*Department `gorm:"-" json:"children,omitempty"` // 用于树状结构展示
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Department) TableName() string {
	return "departments"
}
