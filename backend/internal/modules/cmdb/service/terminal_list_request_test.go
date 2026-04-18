package service

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestBuildTerminalListQuery_ParsesRFC3339TimeRange(t *testing.T) {
	start := "2026-04-17T08:30:00Z"
	end := "2026-04-17T10:45:00Z"

	query, err := buildTerminalListQuery(TerminalListRequest{
		Keyword:  " keyword ",
		Username: " user ",
		Status:   " active ",
		StartAt:  start,
		EndAt:    end,
		Page:     2,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("buildTerminalListQuery returned error: %v", err)
	}

	if query.Keyword != "keyword" {
		t.Fatalf("Keyword = %q, want %q", query.Keyword, "keyword")
	}
	if query.Username != "user" {
		t.Fatalf("Username = %q, want %q", query.Username, "user")
	}
	if query.Status != "active" {
		t.Fatalf("Status = %q, want %q", query.Status, "active")
	}
	if query.StartAt == nil {
		t.Fatal("StartAt = nil, want parsed time")
	}
	if query.EndAt == nil {
		t.Fatal("EndAt = nil, want parsed time")
	}

	wantStart, _ := time.Parse(time.RFC3339, start)
	wantEnd, _ := time.Parse(time.RFC3339, end)
	if !query.StartAt.Equal(wantStart) {
		t.Fatalf("StartAt = %v, want %v", *query.StartAt, wantStart)
	}
	if !query.EndAt.Equal(wantEnd) {
		t.Fatalf("EndAt = %v, want %v", *query.EndAt, wantEnd)
	}
}

func TestBuildTerminalListQuery_InvalidStartAt(t *testing.T) {
	_, err := buildTerminalListQuery(TerminalListRequest{StartAt: "not-a-time"})
	if err == nil {
		t.Fatal("expected error for invalid StartAt")
	}
	if !errors.Is(err, ErrInvalidTerminalListFilter) {
		t.Fatalf("errors.Is(err, ErrInvalidTerminalListFilter) = false, err = %v", err)
	}
	if !strings.Contains(err.Error(), "invalid startAt format") {
		t.Fatalf("error = %q, want substring %q", err.Error(), "invalid startAt format")
	}
}

func TestBuildTerminalListQuery_InvertedTimeRange(t *testing.T) {
	_, err := buildTerminalListQuery(TerminalListRequest{
		StartAt: "2026-04-17T10:45:00Z",
		EndAt:   "2026-04-17T08:30:00Z",
	})
	if err == nil {
		t.Fatal("expected error for inverted time range")
	}
	if !errors.Is(err, ErrInvalidTerminalListFilter) {
		t.Fatalf("errors.Is(err, ErrInvalidTerminalListFilter) = false, err = %v", err)
	}
	if !strings.Contains(err.Error(), "startAt must be before or equal to endAt") {
		t.Fatalf("error = %q, want substring %q", err.Error(), "startAt must be before or equal to endAt")
	}
}
