package user

import (
	"time"

	"gorm.io/gorm"
)

// LoginLog 登录日志表
type LoginLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username  string    `gorm:"type:varchar(50);comment:用户名" json:"username"`
	IP        string    `gorm:"type:varchar(50);comment:登录IP" json:"ip"`
	Location  string    `gorm:"type:varchar(100);comment:登录地点" json:"location"`
	Browser   string    `gorm:"type:varchar(50);comment:浏览器" json:"browser"`
	OS        string    `gorm:"type:varchar(50);comment:操作系统" json:"os"`
	Status    int       `gorm:"type:tinyint;comment:登录状态:1成功,2失败" json:"status"`
	Message   string    `gorm:"type:varchar(255);comment:提示信息" json:"message"`
	LoginTime time.Time `json:"loginTime"`
}

// TableName 指定表名
func (LoginLog) TableName() string {
	return "sys_login_log"
}
