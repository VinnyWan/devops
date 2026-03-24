package service

import (
	"context"
	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/pkg/logger"
	"time"

	"go.uber.org/zap"
)

type AuditService struct {
	repo   *repository.AuditRepo
	cancel context.CancelFunc
}

type AuditListRequest struct {
	UserID    *uint
	Username  string
	Operation string
	Resource  string
	Keyword   string
	StartAt   string
	EndAt     string
	Page      int
	PageSize  int
}

func NewAuditService(repo *repository.AuditRepo) *AuditService {
	return &AuditService{repo: repo}
}

// CleanExpiredAuditLogs 清理过期的审计日志
func (s *AuditService) CleanExpiredAuditLogs() {
	logger.Log.Info("开始清理过期审计日志")

	count, err := s.CleanExpiredNow()
	if err != nil {
		logger.Log.Error("清理过期审计日志失败", zap.Error(err))
	} else {
		logger.Log.Info("清理过期审计日志完成", zap.Int64("count", count))
	}
}

func (s *AuditService) CleanExpiredNow() (int64, error) {
	return s.repo.CleanExpired(time.Now())
}

// StartAuditCleanupTask 开启审计日志定期清理任务（支持优雅关闭）
func (s *AuditService) StartAuditCleanupTask() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		s.CleanExpiredAuditLogs()

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("审计日志清理任务已停止")
				return
			case <-ticker.C:
				s.CleanExpiredAuditLogs()
			}
		}
	}()
}

// StopAuditCleanupTask 停止审计日志清理任务
func (s *AuditService) StopAuditCleanupTask() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *AuditService) List(req AuditListRequest) ([]map[string]interface{}, int64, error) {
	query, err := buildAuditQuery(req)
	if err != nil {
		return nil, 0, err
	}
	query.Page = req.Page
	query.PageSize = req.PageSize

	logs, total, err := s.repo.List(query)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, 0, len(logs))
	for _, item := range logs {
		result = append(result, formatAuditLog(item))
	}

	return result, total, nil
}

func (s *AuditService) Export(req AuditListRequest, limit int) ([]map[string]interface{}, error) {
	query, err := buildAuditQuery(req)
	if err != nil {
		return nil, err
	}
	logs, err := s.repo.ListForExport(query, limit)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(logs))
	for _, item := range logs {
		result = append(result, formatAuditLog(item))
	}
	return result, nil
}

func buildAuditQuery(req AuditListRequest) (repository.AuditQuery, error) {
	query := repository.AuditQuery{
		UserID:    req.UserID,
		Username:  req.Username,
		Operation: req.Operation,
		Resource:  req.Resource,
		Keyword:   req.Keyword,
	}

	if req.StartAt != "" {
		startAt, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			return repository.AuditQuery{}, err
		}
		query.StartAt = &startAt
	}
	if req.EndAt != "" {
		endAt, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			return repository.AuditQuery{}, err
		}
		query.EndAt = &endAt
	}
	return query, nil
}

func formatAuditLog(item model.AuditLog) map[string]interface{} {
	return map[string]interface{}{
		"id":            item.ID,
		"userId":        item.UserID,
		"username":      item.Username,
		"operation":     item.Operation,
		"method":        item.Method,
		"path":          item.Path,
		"params":        item.Params,
		"result":        item.Result,
		"errorMessage":  item.ErrorMessage,
		"ip":            item.IP,
		"status":        item.Status,
		"latency":       item.Latency,
		"retentionDays": item.RetentionDays,
		"requestAt":     item.RequestAt,
		"createdAt":     item.CreatedAt,
	}
}
