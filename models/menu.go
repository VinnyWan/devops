package models

import (
	"time"

	"gorm.io/gorm"
)

// Menu 菜单表
type Menu struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	MenuName  string `gorm:"type:varchar(50);not null" json:"menuName" binding:"required"`
	ParentID  uint   `gorm:"default:0;comment:父菜单ID" json:"parentId"`
	Sort      int    `gorm:"type:int;default:0" json:"sort"`
	Path      string `gorm:"type:varchar(200)" json:"path"`
	Component string `gorm:"type:varchar(255)" json:"component"`
	MenuType  string `gorm:"type:char(1);comment:菜单类型:M目录,C菜单,B按钮" json:"menuType"`
	Visible   int    `gorm:"type:tinyint;default:1;comment:是否显示:1显示,2隐藏" json:"visible"`
	Status    int    `gorm:"type:tinyint;default:1;comment:状态:1正常,2禁用" json:"status"`
	Perms     string `gorm:"type:varchar(100);comment:权限标识" json:"perms"`
	Icon      string `gorm:"type:varchar(100)" json:"icon"`
	Remark    string `gorm:"type:varchar(500)" json:"remark"`
}

// TableName 指定表名
func (Menu) TableName() string {
	return "sys_menu"
}
