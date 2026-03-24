package service

import (
	"testing"
	"time"

	"devops-platform/internal/modules/log/model"
)

func TestLogServiceSearch_FilterByKeyword(t *testing.T) {
	svc := NewLogService()
	result := svc.Search(model.SearchRequest{
		Keyword:  "timeout",
		Page:     1,
		PageSize: 20,
	})
	if result.Total != 1 {
		t.Fatalf("expected 1 log matched keyword, got %d", result.Total)
	}
	if result.Items[0].Level != "error" {
		t.Fatalf("expected matched log level error, got %s", result.Items[0].Level)
	}

	shortKeyword := svc.Search(model.SearchRequest{
		Keyword: "io",
	})
	if shortKeyword.Total != 3 {
		t.Fatalf("expected short keyword ignored, got %d", shortKeyword.Total)
	}
}

func TestLogServiceSearch_PaginationAndTimeRange(t *testing.T) {
	svc := NewLogService()
	start := time.Now().Add(-3 * time.Minute)
	end := time.Now()
	result := svc.Search(model.SearchRequest{
		Start:    start,
		End:      end,
		Page:     1,
		PageSize: 1,
	})
	if result.Total == 0 {
		t.Fatalf("expected logs in recent range")
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected one item by pagination, got %d", len(result.Items))
	}
}

func TestLogServiceSearch_InvalidPaginationFallback(t *testing.T) {
	svc := NewLogService()
	result := svc.Search(model.SearchRequest{
		Page:     -1,
		PageSize: 999,
	})
	if result.Page != 1 {
		t.Fatalf("expected fallback page 1, got %d", result.Page)
	}
	if result.PageSize != 100 {
		t.Fatalf("expected capped page size 100, got %d", result.PageSize)
	}
}

func TestLogServiceSearch_EndBeforeStart(t *testing.T) {
	svc := NewLogService()
	start := time.Now()
	end := start.Add(-5 * time.Minute)
	result := svc.Search(model.SearchRequest{
		Start: start,
		End:   end,
	})
	if result.Total == 0 {
		t.Fatalf("expected non-empty result after swapping invalid time range")
	}
}

func TestParseTime_InvalidFormat(t *testing.T) {
	parsed := ParseTime("not-a-time")
	if !parsed.IsZero() {
		t.Fatalf("expected zero time for invalid format")
	}
}
