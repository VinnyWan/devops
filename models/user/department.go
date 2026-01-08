package user

import (
	"time"

	"gorm.io/gorm"
)

// Department 部门表
type Department struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	DeptName string `gorm:"type:varchar(50);not null" json:"deptName" binding:"required"`
	ParentID uint   `gorm:"default:0;comment:父部门ID" json:"parentId"`
	Sort     int    `gorm:"type:int;default:0" json:"sort"`
	Leader   string `gorm:"type:varchar(50)" json:"leader"`
	Phone    string `gorm:"type:varchar(20)" json:"phone"`
	Email    string `gorm:"type:varchar(100)" json:"email"`
	Status   int    `gorm:"type:tinyint;default:1;comment:状态:1正常,2禁用" json:"status"`
	Remark   string `gorm:"type:varchar(500)" json:"remark"`
}

// TableName 指定表名
func (Department) TableName() string {
	return "sys_department"
}
