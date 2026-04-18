package repository

import (
	"strings"
	"testing"
	"time"

	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type terminalRepoTestDialector struct{}

func (terminalRepoTestDialector) Name() string {
	return "terminal-repo-test"
}

func (terminalRepoTestDialector) Initialize(db *gorm.DB) error {
	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{
		CreateClauses: []string{"INSERT", "VALUES", "ON CONFLICT"},
		QueryClauses:  []string{},
		UpdateClauses: []string{"UPDATE", "SET", "WHERE"},
		DeleteClauses: []string{"DELETE", "FROM", "WHERE"},
	})
	return nil
}

func (terminalRepoTestDialector) Migrator(*gorm.DB) gorm.Migrator {
	return nil
}

func (terminalRepoTestDialector) DataTypeOf(*schema.Field) string {
	return ""
}

func (terminalRepoTestDialector) DefaultValueOf(*schema.Field) clause.Expression {
	return clause.Expr{SQL: "DEFAULT"}
}

func (terminalRepoTestDialector) BindVarTo(writer clause.Writer, _ *gorm.Statement, _ interface{}) {
	writer.WriteByte('?')
}

func (terminalRepoTestDialector) QuoteTo(writer clause.Writer, value string) {
	writer.WriteString(value)
}

func (terminalRepoTestDialector) Explain(sql string, _ ...interface{}) string {
	return sql
}

func openTerminalRepoDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(terminalRepoTestDialector{}, &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open dry-run db: %v", err)
	}
	return db
}

func TestTerminalRepoListInTenant_UsesLikeForKeywordFilters(t *testing.T) {
	db := openTerminalRepoDryRunDB(t)
	repo := NewTerminalRepo(db)
	startAt := time.Date(2026, 4, 17, 8, 30, 0, 0, time.UTC)
	endAt := time.Date(2026, 4, 17, 10, 45, 0, 0, time.UTC)

	tx := repo.scopeInTenant(db.Session(&gorm.Session{DryRun: true}).Model(&model.TerminalSession{}), 7)
	tx = applyTerminalListFilters(tx, "prod_%", "ops_%", "active", &startAt, &endAt)
	tx = tx.Order("started_at DESC").Offset(0).Limit(20).Find(&[]model.TerminalSession{})

	sql := tx.Statement.SQL.String()
	if strings.Contains(strings.ToUpper(sql), "MATCH (") {
		t.Fatalf("expected terminal list query to avoid MATCH AGAINST, got SQL: %s", sql)
	}
	if !strings.Contains(sql, "LOWER(host_name) LIKE ?") {
		t.Fatalf("expected host_name LIKE filter in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "LOWER(host_ip) LIKE ?") {
		t.Fatalf("expected host_ip LIKE filter in SQL, got: %s", sql)
	}
	if !strings.Contains(sql, "LOWER(username) LIKE ?") {
		t.Fatalf("expected username LIKE filter in SQL, got: %s", sql)
	}
	if strings.Contains(sql, "ESCAPE") {
		t.Fatalf("expected terminal list query to avoid ESCAPE clause, got SQL: %s", sql)
	}
	if !strings.Contains(sql, "status = ?") {
		t.Fatalf("expected status filter preserved, got: %s", sql)
	}
	if !strings.Contains(sql, "started_at >= ?") {
		t.Fatalf("expected startAt filter preserved, got: %s", sql)
	}
	if !strings.Contains(sql, "started_at <= ?") {
		t.Fatalf("expected endAt filter preserved, got: %s", sql)
	}
}
