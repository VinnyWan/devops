package workflow

import (
	"encoding/json"
	"fmt"

	"devops-platform/internal/modules/workflow/model"
	"devops-platform/internal/pkg/logger"

	"go.uber.org/zap"
)

// CallbackHandler processes a callback for an approved change order.
// Each handler is registered for a specific module.
type CallbackHandler interface {
	Handle(order *model.ChangeOrder) error
}

// CallbackExecutor dispatches approved orders to registered handlers.
type CallbackExecutor struct {
	handlers map[string]CallbackHandler
}

func NewCallbackExecutor() *CallbackExecutor {
	return &CallbackExecutor{handlers: make(map[string]CallbackHandler)}
}

func (e *CallbackExecutor) Register(module string, handler CallbackHandler) {
	e.handlers[module] = handler
}

// Execute runs the callback for an approved order and updates its status.
// Returns nil if the callback succeeded, or an error if it failed.
func (e *CallbackExecutor) Execute(s *WorkflowService, order *model.ChangeOrder) error {
	if err := s.transition(order, model.StatusExecuting); err != nil {
		return fmt.Errorf("transition to executing: %w", err)
	}

	if order.CallbackModule == "" {
		if err := s.transition(order, model.StatusCompleted); err != nil {
			return fmt.Errorf("transition to completed: %w", err)
		}
		return nil
	}

	handler, ok := e.handlers[order.CallbackModule]
	if !ok {
		logger.Log.Warn("no callback handler registered for module", zap.String("module", order.CallbackModule))
		if err := s.transition(order, model.StatusCompleted); err != nil {
			return fmt.Errorf("transition to completed: %w", err)
		}
		return nil
	}

	if err := handler.Handle(order); err != nil {
		logger.Log.Error("callback execution failed",
			zap.Uint("orderID", order.ID),
			zap.String("module", order.CallbackModule),
			zap.String("action", order.CallbackAction),
			zap.Error(err))
		if transErr := s.transition(order, model.StatusFailed); transErr != nil {
			logger.Log.Error("transition to failed state error", zap.Error(transErr))
		}
		return err
	}

	if err := s.transition(order, model.StatusCompleted); err != nil {
		return fmt.Errorf("transition to completed: %w", err)
	}
	return nil
}

// ExecuteOrder is a convenience method that loads the order and executes its callback.
func (e *CallbackExecutor) ExecuteOrder(s *WorkflowService, orderID, tenantID uint) error {
	order, err := s.getOrder(orderID, tenantID)
	if err != nil {
		return err
	}
	if order.Status != model.StatusApproved {
		return fmt.Errorf("order %d is not in approved state (current: %s)", orderID, order.Status)
	}
	return e.Execute(s, order)
}

// TaskCallback handles callback for the "task" module.
type TaskCallback struct {
	ExecuteTask func(taskID uint, tenantID uint) error
}

func (h *TaskCallback) Handle(order *model.ChangeOrder) error {
	var payload struct {
		TaskID uint `json:"taskId"`
	}
	if err := json.Unmarshal([]byte(order.CallbackPayload), &payload); err != nil {
		return fmt.Errorf("invalid task callback payload: %w", err)
	}
	if payload.TaskID == 0 {
		return fmt.Errorf("task callback payload missing taskId")
	}
	return h.ExecuteTask(payload.TaskID, order.TenantID)
}

// NotificationCallback handles callback for the "notification" module.
type NotificationCallback struct {
	SendNotification func(tenantID uint, channel, recipients, subject, body string) error
}

func (h *NotificationCallback) Handle(order *model.ChangeOrder) error {
	var payload struct {
		Channel    string `json:"channel"`
		Recipients string `json:"recipients"`
		Subject    string `json:"subject"`
		Body       string `json:"body"`
	}
	if err := json.Unmarshal([]byte(order.CallbackPayload), &payload); err != nil {
		return fmt.Errorf("invalid notification callback payload: %w", err)
	}
	if payload.Subject == "" {
		payload.Subject = fmt.Sprintf("变更工单 #%d 已审批通过", order.ID)
	}
	if payload.Body == "" {
		payload.Body = fmt.Sprintf("工单「%s」已通过审批，请关注执行结果。", order.Title)
	}
	return h.SendNotification(order.TenantID, payload.Channel, payload.Recipients, payload.Subject, payload.Body)
}
