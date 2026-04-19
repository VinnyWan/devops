package model

import "time"

type FieldAction string

const (
	FieldActionVisible  FieldAction = "visible"
	FieldActionReadonly FieldAction = "readonly"
	FieldActionHidden   FieldAction = "hidden"
)

// FieldPermission 字段级权限，控制角色对资源字段的可见性
type FieldPermission struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	RoleID    uint        `gorm:"not null;uniqueIndex:uk_field_perm" json:"roleId"`
	Resource  string      `gorm:"size:100;not null;uniqueIndex:uk_field_perm" json:"resource"`
	FieldName string      `gorm:"size:100;not null;uniqueIndex:uk_field_perm" json:"fieldName"`
	Action    FieldAction `gorm:"size:20;not null;default:'visible'" json:"action"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}
