package model

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus represents the state of a change order.
type OrderStatus string

const (
	StatusDraft         OrderStatus = "draft"
	StatusPendingReview OrderStatus = "pending_review"
	StatusApproved      OrderStatus = "approved"
	StatusExecuting     OrderStatus = "executing"
	StatusCompleted     OrderStatus = "completed"
	StatusRejected      OrderStatus = "rejected"
	StatusFailed        OrderStatus = "failed"
)

// ValidTransitions maps each status to allowed next statuses.
var ValidTransitions = map[OrderStatus][]OrderStatus{
	StatusDraft:         {StatusPendingReview},
	StatusPendingReview: {StatusApproved, StatusRejected},
	StatusApproved:      {StatusExecuting},
	StatusExecuting:     {StatusCompleted, StatusFailed},
}

// ChangeOrder represents an approval-tracked change request.
type ChangeOrder struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	TenantID        uint           `gorm:"index;not null" json:"tenant_id"`
	Title           string         `gorm:"size:255;not null" json:"title"`
	Description     string         `gorm:"type:text" json:"description"`
	Type            string         `gorm:"size:32;not null" json:"type"`
	Status          OrderStatus    `gorm:"size:32;not null;default:draft" json:"status"`
	ApprovalLevels  int            `gorm:"default:1" json:"approval_levels"`
	CurrentLevel    int            `gorm:"default:0" json:"current_level"`
	SubmittedBy     uint           `json:"submitted_by"`
	CallbackModule  string         `gorm:"size:64" json:"callback_module"`
	CallbackAction  string         `gorm:"size:64" json:"callback_action"`
	CallbackPayload string         `gorm:"type:text" json:"callback_payload"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChangeOrder) TableName() string { return "change_orders" }

// Approval records each approval step in a change order.
type Approval struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	OrderID       uint        `gorm:"index;not null" json:"order_id"`
	Level         int         `gorm:"not null" json:"level"`
	ApproverID    uint        `json:"approver_id"`
	Status        OrderStatus `gorm:"size:32" json:"status"`
	Comment       string      `gorm:"size:1024" json:"comment"`
	ApprovedAt    *time.Time  `json:"approved_at"`
	CreatedAt     time.Time   `json:"created_at"`
}

func (Approval) TableName() string { return "change_approvals" }
