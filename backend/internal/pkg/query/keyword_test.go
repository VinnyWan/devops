package query

import "testing"

func TestBuildMySQLBooleanQuery(t *testing.T) {
	tests := []struct {
		name     string
		keyword  string
		expected string
	}{
		{name: "single token", keyword: "cluster", expected: "+cluster*"},
		{name: "multi token", keyword: "ops prod", expected: "+ops* +prod*"},
		{name: "strip symbols", keyword: "ops@prod!", expected: "+opsprod*"},
		{name: "too short", keyword: "ab", expected: ""},
		{name: "mixed short and valid", keyword: "ab cluster", expected: "+cluster*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildMySQLBooleanQuery(tt.keyword)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestMatchKeywordAny(t *testing.T) {
	tests := []struct {
		name     string
		keyword  string
		fields   []string
		expected bool
	}{
		{name: "blank keyword always true", keyword: " ", fields: []string{"abc"}, expected: true},
		{name: "short keyword ignored", keyword: "ab", fields: []string{"none"}, expected: true},
		{name: "cross fields match", keyword: "核心集群", fields: []string{"prod", "核心集群平台"}, expected: true},
		{name: "case insensitive", keyword: "ALICE", fields: []string{"alice@example.com"}, expected: true},
		{name: "no field match", keyword: "payment", fields: []string{"gateway", "cluster"}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchKeywordAny(tt.keyword, tt.fields...)
			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
