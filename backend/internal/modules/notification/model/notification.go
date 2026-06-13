package model

import (
	"time"

	"gorm.io/gorm"
)

type ChannelType string

const (
	ChannelFeishu   ChannelType = "feishu"
	ChannelDingTalk ChannelType = "dingtalk"
	ChannelWeCom    ChannelType = "wecom"
	ChannelEmail    ChannelType = "email"
)

type Template struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TenantID  uint           `gorm:"index;not null" json:"tenant_id"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	Channel   ChannelType    `gorm:"size:32;not null" json:"channel"`
	Subject   string         `gorm:"size:255" json:"subject"`
	Body      string         `gorm:"type:text;not null" json:"body"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Template) TableName() string { return "notification_templates" }

type ChannelConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TenantID  uint           `gorm:"index;not null" json:"tenant_id"`
	Channel   ChannelType    `gorm:"size:32;not null;uniqueIndex:uk_chan_tenant" json:"channel"`
	Enabled   bool           `gorm:"default:true" json:"enabled"`
	Priority  int            `gorm:"default:0" json:"priority"`
	Config    string         `gorm:"type:text" json:"config"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChannelConfig) TableName() string { return "notification_channels" }

type SendLog struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	TenantID  uint        `gorm:"index;not null" json:"tenant_id"`
	Channel   ChannelType `gorm:"size:32;not null" json:"channel"`
	Recipient string      `gorm:"size:255" json:"recipient"`
	Content   string      `gorm:"type:text" json:"content"`
	Success   bool        `json:"success"`
	Error     string      `gorm:"size:1024" json:"error"`
	CreatedAt time.Time   `json:"created_at"`
}

func (SendLog) TableName() string { return "notification_logs" }
