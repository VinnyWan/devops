package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"devops-platform/internal/modules/task/model"
	"devops-platform/internal/modules/task/repository"

	"gorm.io/gorm"
)

type TaskService struct {
	taskRepo      *repository.TaskRepo
	execRepo      *repository.ExecutionRepo
	scheduleRepo  *repository.ScheduleRepo
	scheduler     *Scheduler
	db            *gorm.DB
}

func NewTaskService(db *gorm.DB, scheduler *Scheduler) *TaskService {
	return &TaskService{
		taskRepo:     repository.NewTaskRepo(db),
		execRepo:     repository.NewExecutionRepo(db),
		scheduleRepo: repository.NewScheduleRepo(db),
		scheduler:    scheduler,
		db:           db,
	}
}

func (s *TaskService) Create(tenantID, userID uint, name, desc string, taskType model.TaskType, content string, timeout int) (*model.Task, error) {
	if timeout <= 0 {
		timeout = 300
	}
	t := &model.Task{
		TenantID:    tenantID,
		Name:        name,
		Description: desc,
		Type:        taskType,
		Content:     content,
		Timeout:     timeout,
		CreatedBy:   userID,
	}
	if err := s.taskRepo.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) Update(id, tenantID uint, name, desc, content string, timeout int) error {
	t, err := s.taskRepo.GetByID(id, tenantID)
	if err != nil {
		return err
	}
	if name != "" {
		t.Name = name
	}
	if desc != "" {
		t.Description = desc
	}
	if content != "" {
		t.Content = content
	}
	if timeout > 0 {
		t.Timeout = timeout
	}
	return s.taskRepo.Update(t)
}

func (s *TaskService) Delete(id, tenantID uint) error {
	_ = s.scheduleRepo.Delete(id, tenantID)
	return s.taskRepo.Delete(id, tenantID)
}

func (s *TaskService) Get(id, tenantID uint) (*model.Task, error) {
	return s.taskRepo.GetByID(id, tenantID)
}

func (s *TaskService) List(tenantID uint, keyword string, page, pageSize int) ([]model.Task, int64, error) {
	return s.taskRepo.List(tenantID, keyword, page, pageSize)
}

func (s *TaskService) Execute(ctx context.Context, taskID, tenantID uint, targets []string) (*model.TaskExecution, error) {
	task, err := s.taskRepo.GetByID(taskID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	targetsStr := ""
	for i, t := range targets {
		if i > 0 {
			targetsStr += ","
		}
		targetsStr += t
	}

	exec := &model.TaskExecution{
		TaskID:   taskID,
		TenantID: tenantID,
		Status:   model.TaskStatusPending,
		Targets:  targetsStr,
	}
	if err := s.execRepo.Create(exec); err != nil {
		return nil, err
	}

	go s.runExecution(ctx, exec, task)
	return exec, nil
}

func (s *TaskService) GetExecution(id, tenantID uint) (*model.TaskExecution, error) {
	return s.execRepo.GetByID(id, tenantID)
}

func (s *TaskService) ListExecutions(taskID, tenantID uint, page, pageSize int) ([]model.TaskExecution, int64, error) {
	return s.execRepo.ListByTask(taskID, tenantID, page, pageSize)
}

func (s *TaskService) SetSchedule(taskID, tenantID uint, cronExpr string) error {
	existing, err := s.scheduleRepo.GetByTaskID(taskID, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			schedule := &model.TaskSchedule{
				TaskID:   taskID,
				TenantID: tenantID,
				CronExpr: cronExpr,
				Enabled:  true,
			}
			if err := s.scheduleRepo.Create(schedule); err != nil {
				return err
			}
			if s.scheduler != nil {
				s.scheduler.Add(taskID, tenantID, cronExpr)
			}
			return nil
		}
		return err
	}
	existing.CronExpr = cronExpr
	return s.scheduleRepo.Update(existing)
}

func (s *TaskService) DeleteSchedule(taskID, tenantID uint) error {
	if s.scheduler != nil {
		s.scheduler.Remove(taskID)
	}
	return s.scheduleRepo.Delete(taskID, tenantID)
}

func (s *TaskService) StartCleanupTask(retentionDays int) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			if err := s.execRepo.DeleteOlderThan(retentionDays); err != nil {
				// Log cleanup error - non-fatal
				_ = err
			}
		}
	}()
}
