package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"

	"gorm.io/gorm"
)

type TerminalService struct {
	terminalRepo   *repository.TerminalRepo
	hostRepo       *repository.HostRepo
	credentialRepo *repository.CredentialRepo
}

type TerminalListRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
	Keyword  string `form:"keyword" json:"keyword"`
	Username string `form:"username" json:"username"`
	Status   string `form:"status" json:"status"`
	StartAt  string `form:"startAt" json:"startAt"`
	EndAt    string `form:"endAt" json:"endAt"`
}

type TerminalListQuery struct {
	Keyword  string
	Username string
	Status   string
	StartAt  *time.Time
	EndAt    *time.Time
}

var ErrInvalidTerminalListFilter = errors.New("invalid terminal list filter")

func newInvalidTerminalListFilterError(message string, err error) error {
	if err == nil {
		return fmt.Errorf("%w: %s", ErrInvalidTerminalListFilter, message)
	}
	return fmt.Errorf("%w: %s: %w", ErrInvalidTerminalListFilter, message, err)
}

func NewTerminalService(db *gorm.DB) *TerminalService {
	return &TerminalService{
		terminalRepo:   repository.NewTerminalRepo(db),
		hostRepo:       repository.NewHostRepo(db),
		credentialRepo: repository.NewCredentialRepo(db),
	}
}

func (s *TerminalService) normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func (s *TerminalService) GetConnectTarget(tenantID, hostID uint) (*model.Host, *model.Credential, error) {
	host, err := s.hostRepo.GetByIDInTenant(tenantID, hostID)
	if err != nil {
		return nil, nil, err
	}
	if host.CredentialID == nil || *host.CredentialID == 0 {
		return nil, nil, errors.New("主机未绑定凭据")
	}

	credential, err := s.credentialRepo.GetByIDInTenant(tenantID, *host.CredentialID)
	if err != nil {
		return nil, nil, err
	}

	return host, credential, nil
}

func (s *TerminalService) CreateSession(tenantID, userID uint, username string, host *model.Host, credentialID uint, clientIP, recordingPath string) (*model.TerminalSession, error) {
	startedAt := time.Now()
	session := &model.TerminalSession{
		TenantID:      tenantID,
		UserID:        userID,
		Username:      username,
		HostID:        host.ID,
		HostIP:        host.Ip,
		HostName:      host.Hostname,
		CredentialID:  credentialID,
		ClientIP:      clientIP,
		StartedAt:     startedAt,
		Duration:      0,
		RecordingPath: recordingPath,
		FileSize:      0,
		Status:        "active",
	}

	if err := s.terminalRepo.Create(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *TerminalService) CloseSession(tenantID, sessionID uint, finishedAt time.Time, fileSize int64, status string, closeReason string) error {
	session, err := s.terminalRepo.GetByIDInTenant(tenantID, sessionID)
	if err != nil {
		return err
	}

	duration := int(finishedAt.Sub(session.StartedAt).Seconds())
	if duration < 0 {
		duration = 0
	}
	if fileSize < 0 {
		fileSize = 0
	}

	session.FinishedAt = &finishedAt
	session.Duration = duration
	session.FileSize = fileSize
	session.Status = status
	session.CloseReason = closeReason

	return s.terminalRepo.UpdateInTenant(tenantID, session)
}

func buildTerminalListQuery(req TerminalListRequest) (TerminalListQuery, error) {
	query := TerminalListQuery{
		Keyword:  strings.TrimSpace(req.Keyword),
		Username: strings.TrimSpace(req.Username),
		Status:   strings.TrimSpace(req.Status),
	}

	if req.StartAt != "" {
		startAt, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			return TerminalListQuery{}, newInvalidTerminalListFilterError("invalid startAt format", err)
		}
		query.StartAt = &startAt
	}

	if req.EndAt != "" {
		endAt, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			return TerminalListQuery{}, newInvalidTerminalListFilterError("invalid endAt format", err)
		}
		query.EndAt = &endAt
	}

	if query.StartAt != nil && query.EndAt != nil && query.StartAt.After(*query.EndAt) {
		return TerminalListQuery{}, newInvalidTerminalListFilterError("startAt must be before or equal to endAt", nil)
	}

	return query, nil
}

func (s *TerminalService) ListInTenant(tenantID uint, req TerminalListRequest) ([]model.TerminalSession, int64, error) {
	page, pageSize := s.normalizePage(req.Page, req.PageSize)
	query, err := buildTerminalListQuery(req)
	if err != nil {
		return nil, 0, err
	}
	return s.terminalRepo.ListInTenant(tenantID, page, pageSize, query.Keyword, query.Username, query.Status, query.StartAt, query.EndAt)
}

func (s *TerminalService) DetailInTenant(tenantID, id uint) (*model.TerminalSession, error) {
	return s.terminalRepo.GetByIDInTenant(tenantID, id)
}

func (s *TerminalService) BuildRecordingPath(baseDir string, startedAt time.Time, sessionID uint) string {
	dayDir := startedAt.Format("2006-01-02")
	fileName := fmt.Sprintf("%d.cast", sessionID)
	return filepath.Join(baseDir, dayDir, fileName)
}

// AddTag adds a tag to a terminal session
func (s *TerminalService) AddTag(tenantID, sessionID, userID uint, tag string) error {
	return s.terminalRepo.AddTagToSession(tenantID, sessionID, userID, tag)
}

// RemoveTag removes a tag from a terminal session
func (s *TerminalService) RemoveTag(tenantID, sessionID uint, tag string) error {
	return s.terminalRepo.RemoveTagFromSession(tenantID, sessionID, tag)
}

// GetTagsForSession returns tags for a session
func (s *TerminalService) GetTagsForSession(sessionID uint) ([]model.SessionTag, error) {
	return s.terminalRepo.GetTagsForSession(sessionID)
}

// SearchByTag returns sessions matching a tag
func (s *TerminalService) SearchByTag(tenantID uint, tag string, page, pageSize int) ([]model.TerminalSession, int64, error) {
	return s.terminalRepo.SearchSessionsByTag(tenantID, tag, page, pageSize)
}

// GetAvailableTags returns all distinct tags in a tenant
func (s *TerminalService) GetAvailableTags(tenantID uint) ([]string, error) {
	return s.terminalRepo.GetAvailableTags(tenantID)
}
