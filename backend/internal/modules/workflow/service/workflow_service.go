package workflow

import (
	"errors"
	"fmt"
	"time"

	"devops-platform/internal/modules/workflow/model"

	"gorm.io/gorm"
)

type WorkflowService struct {
	db               *gorm.DB
	callbackExecutor *CallbackExecutor
}

func NewWorkflowService(db *gorm.DB) *WorkflowService {
	return &WorkflowService{db: db}
}

func (s *WorkflowService) SetCallbackExecutor(executor *CallbackExecutor) {
	s.callbackExecutor = executor
}

func (s *WorkflowService) CreateOrder(tenantID, userID uint, title, desc, orderType string, approvalLevels int) (*model.ChangeOrder, error) {
	if approvalLevels < 1 {
		approvalLevels = 1
	}
	if approvalLevels > 3 {
		approvalLevels = 3
	}
	order := &model.ChangeOrder{
		TenantID:       tenantID,
		Title:          title,
		Description:    desc,
		Type:           orderType,
		Status:         model.StatusDraft,
		ApprovalLevels: approvalLevels,
		SubmittedBy:    userID,
	}
	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}
	return order, nil
}

func (s *WorkflowService) SubmitForReview(orderID, tenantID uint) error {
	order, err := s.getOrder(orderID, tenantID)
	if err != nil {
		return err
	}
	return s.transition(order, model.StatusPendingReview)
}

func (s *WorkflowService) Approve(orderID, tenantID, approverID uint, comment string) error {
	order, err := s.getOrder(orderID, tenantID)
	if err != nil {
		return err
	}
	if order.Status != model.StatusPendingReview {
		return errors.New("order is not pending review")
	}

	level := order.CurrentLevel + 1
	approval := &model.Approval{
		OrderID:    orderID,
		Level:      level,
		ApproverID: approverID,
		Status:     model.StatusApproved,
		Comment:    comment,
		ApprovedAt: timePtr(time.Now()),
	}
	if err := s.db.Create(approval).Error; err != nil {
		return err
	}

	order.CurrentLevel = level
	if level >= order.ApprovalLevels {
		if err := s.transition(order, model.StatusApproved); err != nil {
			return err
		}
		if s.callbackExecutor != nil {
			go func() {
				if execErr := s.callbackExecutor.Execute(s, order); execErr != nil {
					// order status already set to failed in Execute
				}
			}()
		}
		return nil
	}
	return s.db.Save(order).Error
}

func (s *WorkflowService) Reject(orderID, tenantID, approverID uint, comment string) error {
	order, err := s.getOrder(orderID, tenantID)
	if err != nil {
		return err
	}
	if order.Status != model.StatusPendingReview {
		return errors.New("order is not pending review")
	}
	approval := &model.Approval{
		OrderID:    orderID,
		Level:      order.CurrentLevel + 1,
		ApproverID: approverID,
		Status:     model.StatusRejected,
		Comment:    comment,
		ApprovedAt: timePtr(time.Now()),
	}
	s.db.Create(approval)
	return s.transition(order, model.StatusRejected)
}

func (wf *WorkflowService) transition(order *model.ChangeOrder, target model.OrderStatus) error {
	allowed, ok := model.ValidTransitions[order.Status]
	if !ok {
		return fmt.Errorf("invalid current status: %s", order.Status)
	}
	for _, s := range allowed {
		if s == target {
			order.Status = target
			return wf.db.Save(order).Error
		}
	}
	return fmt.Errorf("cannot transition from %s to %s", order.Status, target)
}

type OrderDetail struct {
	*model.ChangeOrder
	Approvals []model.Approval `json:"approvals"`
}

func (s *WorkflowService) GetOrder(id, tenantID uint) (*OrderDetail, error) {
	order, err := s.getOrder(id, tenantID)
	if err != nil {
		return nil, err
	}
	var approvals []model.Approval
	s.db.Where("order_id = ?", id).Order("level ASC").Find(&approvals)
	return &OrderDetail{ChangeOrder: order, Approvals: approvals}, nil
}

func (s *WorkflowService) GetApprovals(orderID uint) ([]model.Approval, error) {
	var approvals []model.Approval
	err := s.db.Where("order_id = ?", orderID).Order("level ASC").Find(&approvals).Error
	return approvals, err
}

func (s *WorkflowService) ExecuteOrder(id, tenantID uint) error {
	if s.callbackExecutor == nil {
		return errors.New("callback executor not configured")
	}
	return s.callbackExecutor.ExecuteOrder(s, id, tenantID)
}

func (s *WorkflowService) getOrder(id, tenantID uint) (*model.ChangeOrder, error) {
	var order model.ChangeOrder
	err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&order).Error
	return &order, err
}

func (s *WorkflowService) ListOrders(tenantID uint, page, pageSize int, status string) ([]model.ChangeOrder, int64, error) {
	var orders []model.ChangeOrder
	var total int64
	q := s.db.Model(&model.ChangeOrder{}).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	err := q.Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

func timePtr(t time.Time) *time.Time { return &t }
