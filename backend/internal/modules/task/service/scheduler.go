package service

import (
	"context"
	"sync"

	"devops-platform/internal/modules/task/model"

	"github.com/robfig/cron/v3"
)

type ScheduledTask struct {
	TaskID   uint
	TenantID uint
	CronExpr string
}

type Scheduler struct {
	cron     *cron.Cron
	entries  map[uint]cron.EntryID
	mu       sync.Mutex
	executor func(ctx context.Context, taskID, tenantID uint) (*model.TaskExecution, error)
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		entries: make(map[uint]cron.EntryID),
	}
}

func (s *Scheduler) SetExecutor(fn func(ctx context.Context, taskID, tenantID uint) (*model.TaskExecution, error)) {
	s.executor = fn
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) Add(taskID, tenantID uint, cronExpr string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, exists := s.entries[taskID]; exists {
		s.cron.Remove(id)
	}
	tid := taskID
	t := tenantID
	id, err := s.cron.AddFunc(cronExpr, func() {
		if s.executor != nil {
			s.executor(context.Background(), tid, t)
		}
	})
	if err != nil {
		return
	}
	s.entries[taskID] = id
}

func (s *Scheduler) Remove(taskID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, exists := s.entries[taskID]; exists {
		s.cron.Remove(id)
		delete(s.entries, taskID)
	}
}

func (s *Scheduler) LoadSchedules(schedules []model.TaskSchedule) {
	for _, sch := range schedules {
		s.Add(sch.TaskID, sch.TenantID, sch.CronExpr)
	}
}
