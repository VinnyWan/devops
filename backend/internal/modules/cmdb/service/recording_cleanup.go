package service

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"devops-platform/config"
	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
	"devops-platform/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RecordingCleanupService struct {
	repo   *repository.TerminalRepo
	db     *gorm.DB
	cancel context.CancelFunc
}

func NewRecordingCleanupService(db *gorm.DB) *RecordingCleanupService {
	return &RecordingCleanupService{
		repo: repository.NewTerminalRepo(db),
		db:   db,
	}
}

func (s *RecordingCleanupService) StartCleanupScheduler() {
	maxAge := config.Cfg.GetInt("terminal.recording.max_age_days")
	if maxAge <= 0 {
		logger.Log.Info("录像清理已禁用（max_age_days = 0）")
		return
	}

	cleanupHour := config.Cfg.GetInt("terminal.recording.cleanup_hour")
	if cleanupHour < 0 || cleanupHour > 23 {
		cleanupHour = 3
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), cleanupHour, 0, 0, 0, now.Location())
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}
			timer := time.NewTimer(next.Sub(now))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				s.CleanupOldRecordings(maxAge)
			}
		}
	}()

	logger.Log.Info("录像清理定时任务已启动", zap.Int("max_age_days", maxAge), zap.Int("cleanup_hour", cleanupHour))
}

func (s *RecordingCleanupService) StopCleanupScheduler() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RecordingCleanupService) CleanupOldRecordings(maxAgeDays int) {
	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)

	var sessions []model.TerminalSession
	if err := s.db.Where("status = ? AND created_at < ?", "closed", cutoff).
		Find(&sessions).Error; err != nil {
		logger.Log.Error("查询过期录像记录失败", zap.Error(err))
		return
	}

	if len(sessions) == 0 {
		return
	}

	recordingDir := config.Cfg.GetString("terminal.recording_dir")
	cleaned := 0

	for _, session := range sessions {
		if session.RecordingPath == "" {
			continue
		}
		fullPath := filepath.Join(recordingDir, session.RecordingPath)
		if _, err := os.Stat(fullPath); err == nil {
			if err := os.Remove(fullPath); err != nil {
				logger.Log.Error("删除录像文件失败", zap.String("path", fullPath), zap.Error(err))
				continue
			}
		}
		s.db.Model(&session).Updates(map[string]interface{}{
			"status":         "archived",
			"recording_path": "",
		})
		cleaned++
	}

	s.cleanEmptyDirs(recordingDir)

	logger.Log.Info("录像清理完成",
		zap.Int("cleaned", cleaned),
		zap.Int("total", len(sessions)))
}

func (s *RecordingCleanupService) cleanEmptyDirs(baseDir string) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(baseDir, entry.Name())
		os.Remove(dirPath)
	}
}
