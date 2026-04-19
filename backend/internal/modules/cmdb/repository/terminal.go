package repository

import (
	"strings"
	"time"

	"devops-platform/internal/modules/cmdb/model"
	queryutil "devops-platform/internal/pkg/query"

	"gorm.io/gorm"
)

type TerminalRepo struct {
	db *gorm.DB
}

func NewTerminalRepo(db *gorm.DB) *TerminalRepo {
	return &TerminalRepo{db: db}
}

func (r *TerminalRepo) scopeInTenant(query *gorm.DB, tenantID uint) *gorm.DB {
	return query.Where("tenant_id = ?", tenantID)
}

func (r *TerminalRepo) Create(session *model.TerminalSession) error {
	return r.db.Create(session).Error
}

func (r *TerminalRepo) GetByIDInTenant(tenantID, id uint) (*model.TerminalSession, error) {
	var session model.TerminalSession
	if err := r.scopeInTenant(r.db, tenantID).Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *TerminalRepo) UpdateInTenant(tenantID uint, session *model.TerminalSession) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.TerminalSession
		if err := r.scopeInTenant(tx, tenantID).Where("id = ?", session.ID).First(&existing).Error; err != nil {
			return err
		}
		session.TenantID = existing.TenantID
		return tx.Save(session).Error
	})
}

func (r *TerminalRepo) ListInTenant(tenantID uint, page, pageSize int, keyword, username, status string, startAt, endAt *time.Time) ([]model.TerminalSession, int64, error) {
	var sessions []model.TerminalSession
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.TerminalSession{}), tenantID)
	query = applyTerminalListFilters(query, keyword, username, status, startAt, endAt)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("started_at DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

func applyTerminalListFilters(query *gorm.DB, keyword, username, status string, startAt, endAt *time.Time) *gorm.DB {
	query = applyTerminalListKeywordLike(query, keyword, "host_name", "host_ip")
	query = applyTerminalListKeywordLike(query, username, "username")

	if status = strings.TrimSpace(status); status != "" {
		query = query.Where("status = ?", status)
	}
	if startAt != nil {
		query = query.Where("started_at >= ?", *startAt)
	}
	if endAt != nil {
		query = query.Where("started_at <= ?", *endAt)
	}

	return query
}

func applyTerminalListKeywordLike(query *gorm.DB, keyword string, columns ...string) *gorm.DB {
	normalized := queryutil.NormalizeKeyword(keyword)
	if normalized == "" || len(columns) == 0 {
		return query
	}

	pattern := "%" + queryutil.EscapeLike(strings.ToLower(normalized)) + "%"
	conditions := make([]string, 0, len(columns))
	args := make([]interface{}, 0, len(columns))
	for _, column := range columns {
		conditions = append(conditions, "LOWER("+column+") LIKE ?")
		args = append(args, pattern)
	}

	return query.Where(strings.Join(conditions, " OR "), args...)
}

// AddTagToSession adds a tag to a terminal session
func (r *TerminalRepo) AddTagToSession(tenantID, sessionID, userID uint, tag string) error {
	sessionTag := &model.SessionTag{
		TenantID:  tenantID,
		SessionID: sessionID,
		Tag:       tag,
		UserID:    userID,
	}
	if err := r.db.Create(sessionTag).Error; err != nil {
		return err
	}
	// Also update the denormalized Tags field on the session
	var session model.TerminalSession
	if err := r.db.First(&session, sessionID).Error; err != nil {
		return err
	}
	tags := session.Tags
	if tags == "" {
		tags = tag
	} else {
		tags = tags + "," + tag
	}
	return r.db.Model(&session).Update("tags", tags).Error
}

// RemoveTagFromSession removes a tag from a session
func (r *TerminalRepo) RemoveTagFromSession(tenantID, sessionID uint, tag string) error {
	if err := r.db.Where("session_id = ? AND tag = ? AND tenant_id = ?", sessionID, tag, tenantID).Delete(&model.SessionTag{}).Error; err != nil {
		return err
	}
	// Rebuild denormalized tags
	var remaining []model.SessionTag
	r.db.Where("session_id = ? AND tenant_id = ?", sessionID, tenantID).Find(&remaining)
	tags := ""
	for i, t := range remaining {
		if i > 0 {
			tags += ","
		}
		tags += t.Tag
	}
	return r.db.Model(&model.TerminalSession{}).Where("id = ?", sessionID).Update("tags", tags).Error
}

// GetTagsForSession returns all tags for a session
func (r *TerminalRepo) GetTagsForSession(sessionID uint) ([]model.SessionTag, error) {
	var tags []model.SessionTag
	err := r.db.Where("session_id = ?", sessionID).Find(&tags).Error
	return tags, err
}

// SearchSessionsByTag returns sessions matching a tag
func (r *TerminalRepo) SearchSessionsByTag(tenantID uint, tag string, page, pageSize int) ([]model.TerminalSession, int64, error) {
	var list []model.TerminalSession
	var total int64

	query := r.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Joins("JOIN cmdb_session_tags ON cmdb_session_tags.session_id = cmdb_terminal_sessions.id AND cmdb_session_tags.tag = ? AND cmdb_session_tags.deleted_at IS NULL", tag)

	query.Model(&model.TerminalSession{}).Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("started_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetAvailableTags returns distinct tags used in a tenant
func (r *TerminalRepo) GetAvailableTags(tenantID uint) ([]string, error) {
	var tags []string
	err := r.db.Model(&model.SessionTag{}).
		Where("tenant_id = ?", tenantID).
		Distinct("tag").
		Pluck("tag", &tags).Error
	return tags, err
}
