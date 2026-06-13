package repository

import (
	"devops-platform/internal/modules/sqlaudit/model"

	"gorm.io/gorm"
)

type SqlAuditRepo struct {
	db *gorm.DB
}

func NewSqlAuditRepo(db *gorm.DB) *SqlAuditRepo {
	return &SqlAuditRepo{db: db}
}

// DbConnection CRUD

func (r *SqlAuditRepo) CreateConnection(conn *model.DbConnection) error {
	return r.db.Create(conn).Error
}

func (r *SqlAuditRepo) GetConnection(id, tenantID uint) (*model.DbConnection, error) {
	var conn model.DbConnection
	err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&conn).Error
	return &conn, err
}

func (r *SqlAuditRepo) UpdateConnection(conn *model.DbConnection) error {
	return r.db.Save(conn).Error
}

func (r *SqlAuditRepo) DeleteConnection(id, tenantID uint) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.DbConnection{}).Error
}

func (r *SqlAuditRepo) ListConnections(tenantID uint, connType string) ([]model.DbConnection, error) {
	var conns []model.DbConnection
	q := r.db.Where("tenant_id = ?", tenantID)
	if connType != "" {
		q = q.Where("type = ?", connType)
	}
	err := q.Order("created_at DESC").Find(&conns).Error
	return conns, err
}

// SqlRecord CRUD

func (r *SqlAuditRepo) CreateRecord(record *model.SqlRecord) error {
	return r.db.Create(record).Error
}

func (r *SqlAuditRepo) ListRecords(tenantID uint, connectionID uint, page, pageSize int) ([]model.SqlRecord, int64, error) {
	var records []model.SqlRecord
	var total int64
	q := r.db.Model(&model.SqlRecord{}).Where("tenant_id = ?", tenantID)
	if connectionID > 0 {
		q = q.Where("connection_id = ?", connectionID)
	}
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("executed_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}
