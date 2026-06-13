package service

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"devops-platform/internal/modules/sqlaudit/model"
	"devops-platform/internal/modules/sqlaudit/repository"
	"devops-platform/internal/pkg/utils"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/gorm"
)

type SqlAuditService struct {
	repo *repository.SqlAuditRepo
	db   *gorm.DB
}

func NewSqlAuditService(db *gorm.DB) *SqlAuditService {
	return &SqlAuditService{repo: repository.NewSqlAuditRepo(db), db: db}
}

// Connection management

func (s *SqlAuditService) CreateConnection(tenantID uint, name, dbType, host string, port int, database, username, password, mode, desc string) (*model.DbConnection, error) {
	if mode == "" {
		mode = "read_write"
	}
	encPwd, err := utils.Encrypt(password)
	if err != nil {
		return nil, fmt.Errorf("encrypt password: %w", err)
	}
	conn := &model.DbConnection{
		TenantID: tenantID, Name: name, Type: dbType, Host: host, Port: port,
		Database: database, Username: username, Password: encPwd, Mode: mode,
		Description: desc, Status: "active",
	}
	if err := s.repo.CreateConnection(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *SqlAuditService) TestConnection(id, tenantID uint) error {
	conn, err := s.repo.GetConnection(id, tenantID)
	if err != nil {
		return err
	}
	db, err := s.openDB(conn)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("Ping 失败: %w", err)
	}
	return nil
}

func (s *SqlAuditService) ListConnections(tenantID uint, connType string) ([]model.DbConnection, error) {
	return s.repo.ListConnections(tenantID, connType)
}

func (s *SqlAuditService) DeleteConnection(id, tenantID uint) error {
	return s.repo.DeleteConnection(id, tenantID)
}

// SQL execution

type ExecuteSQLRequest struct {
	ConnectionID uint   `json:"connectionId"`
	SQL          string `json:"sql"`
}

type ExecuteSQLResult struct {
	Columns      []string        `json:"columns"`
	Rows         [][]interface{} `json:"rows"`
	RowsAffected int64           `json:"rowsAffected"`
	Duration     int64           `json:"duration"`
}

// ExecuteSQL runs a SQL statement and records an audit log.
func (s *SqlAuditService) ExecuteSQL(tenantID, userID uint, clientIP string, req ExecuteSQLRequest) (*ExecuteSQLResult, error) {
	conn, err := s.repo.GetConnection(req.ConnectionID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("数据库连接不存在: %w", err)
	}

	stmt := strings.TrimSpace(req.SQL)
	if stmt == "" {
		return nil, fmt.Errorf("SQL 不能为空")
	}

	// Sensitive SQL detection
	sensitive, risk := detectSensitive(stmt)

	// Read-only mode blocks write operations
	if conn.Mode == "read_only" && isWriteSQL(stmt) {
		s.recordAudit(tenantID, userID, clientIP, conn, stmt, "read_only", sensitive, risk, 0, 0, "只读模式禁止写操作")
		return nil, fmt.Errorf("只读模式禁止写操作: %s", firstWord(stmt))
	}

	db, err := s.openDB(conn)
	if err != nil {
		s.recordAudit(tenantID, userID, clientIP, conn, stmt, conn.Mode, sensitive, risk, 0, 0, err.Error())
		return nil, err
	}
	defer db.Close()

	start := time.Now()
	var result ExecuteSQLResult

	if isQuerySQL(stmt) {
		rows, err := db.Query(stmt)
		if err != nil {
			elapsed := time.Since(start).Milliseconds()
			s.recordAudit(tenantID, userID, clientIP, conn, stmt, conn.Mode, sensitive, risk, elapsed, 0, err.Error())
			return nil, err
		}
		defer rows.Close()
		columns, _ := rows.Columns()
		result.Columns = columns
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			rows.Scan(valuePtrs...)
			result.Rows = append(result.Rows, values)
		}
		result.RowsAffected = int64(len(result.Rows))
	} else {
		res, err := db.Exec(stmt)
		if err != nil {
			elapsed := time.Since(start).Milliseconds()
			s.recordAudit(tenantID, userID, clientIP, conn, stmt, conn.Mode, sensitive, risk, elapsed, 0, err.Error())
			return nil, err
		}
		result.RowsAffected, _ = res.RowsAffected()
	}

	result.Duration = time.Since(start).Milliseconds()
	s.recordAudit(tenantID, userID, clientIP, conn, stmt, conn.Mode, sensitive, risk, result.Duration, result.RowsAffected, "")
	return &result, nil
}

// ListRecords returns paginated SQL audit records.
func (s *SqlAuditService) ListRecords(tenantID, connectionID uint, page, pageSize int) ([]model.SqlRecord, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListRecords(tenantID, connectionID, page, pageSize)
}

func (s *SqlAuditService) openDB(conn *model.DbConnection) (*sql.DB, error) {
	pwd, err := utils.Decrypt(conn.Password)
	if err != nil {
		return nil, fmt.Errorf("decrypt password: %w", err)
	}
	driver := conn.Type
	if driver == "" {
		driver = "mysql"
	}
	port := conn.Port
	if port == 0 {
		if driver == "postgresql" || driver == "postgres" {
			port = 5432
		} else {
			port = 3306
		}
	}
	var dsn string
	driverName := driver
	switch driver {
	case "postgresql", "postgres":
		driverName = "pgx"
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			conn.Username, pwd, conn.Host, port, conn.Database)
	default:
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
			conn.Username, pwd, conn.Host, port, conn.Database)
	}
	return sql.Open(driverName, dsn)
}

func (s *SqlAuditService) recordAudit(tenantID, userID uint, clientIP string, conn *model.DbConnection, sqlStmt, mode string, sensitive bool, risk string, duration, rowsAffected int64, errMsg string) {
	record := &model.SqlRecord{
		TenantID:     tenantID,
		ConnectionID: conn.ID,
		UserID:       userID,
		Database:     conn.Database,
		SQL:          sqlStmt,
		Mode:         mode,
		Sensitive:    sensitive,
		RiskLevel:    risk,
		Duration:     duration,
		RowsAffected: rowsAffected,
		Error:        errMsg,
		ClientIP:     clientIP,
		ExecutedAt:   time.Now(),
	}
	s.repo.CreateRecord(record)
}

func detectSensitive(stmt string) (bool, string) {
	for _, p := range model.SensitivePatterns {
		if matched, _ := regexp.MatchString(p.Pattern, stmt); matched {
			return true, p.Risk
		}
	}
	return false, RiskNone
}

const RiskNone = "none"

func isQuerySQL(stmt string) bool {
	s := strings.ToUpper(strings.TrimSpace(stmt))
	return strings.HasPrefix(s, "SELECT") || strings.HasPrefix(s, "SHOW") ||
		strings.HasPrefix(s, "DESCRIBE") || strings.HasPrefix(s, "EXPLAIN")
}

func isWriteSQL(stmt string) bool {
	s := strings.ToUpper(strings.TrimSpace(stmt))
	writeOps := []string{"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "CREATE", "TRUNCATE", "RENAME", "GRANT", "REVOKE"}
	for _, op := range writeOps {
		if strings.HasPrefix(s, op) {
			return true
		}
	}
	return false
}

func firstWord(s string) string {
	parts := strings.Fields(s)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
