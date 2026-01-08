package models

import (
	"time"

	"gorm.io/gorm"
)

// Post 岗位表
type Post struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PostName string `gorm:"type:varchar(50);uniqueIndex;not null" json:"postName" binding:"required"`
	PostCode string `gorm:"type:varchar(50);uniqueIndex;not null" json:"postCode" binding:"required"`
	Sort     int    `gorm:"type:int;default:0" json:"sort"`
	Status   int    `gorm:"type:tinyint;default:1;comment:状态:1正常,2禁用" json:"status"`
	Remark   string `gorm:"type:varchar(500)" json:"remark"`
}

// TableName 指定表名
func (Post) TableName() string {
	return "sys_post"
}
