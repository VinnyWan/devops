package model

import (
	"time"

	"gorm.io/gorm"
)

// LoginLog 登录日志模型
type LoginLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"size:50;index" json:"username"`
	IP        string         `gorm:"size:50;index" json:"ip"`
	Location  string         `gorm:"size:100" json:"location"`
	Browser   string         `gorm:"size:100" json:"browser"`
	OS        string         `gorm:"size:100" json:"os"`
	Status    string         `gorm:"size:20;index" json:"status"`
	Message   string         `gorm:"size:200" json:"message"`
	UserAgent string         `gorm:"size:500" json:"userAgent"`
	LoginAt   time.Time      `gorm:"index" json:"loginAt"`
	CreatedAt time.Time      `gorm:"index" json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (LoginLog) TableName() string {
	return "login_logs"
}
