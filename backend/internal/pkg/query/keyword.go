package query

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"gorm.io/gorm"
)

const MinKeywordLength = 3

func NormalizeKeyword(keyword string) string {
	normalized := strings.TrimSpace(keyword)
	if utf8.RuneCountInString(normalized) < MinKeywordLength {
		return ""
	}
	return normalized
}

func EscapeLike(keyword string) string {
	escaped := strings.ReplaceAll(keyword, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `%`, `\%`)
	escaped = strings.ReplaceAll(escaped, `_`, `\_`)
	return escaped
}

func ApplyKeywordLike(db *gorm.DB, keyword string, columns ...string) *gorm.DB {
	normalized := NormalizeKeyword(keyword)
	if normalized == "" || len(columns) == 0 {
		return db
	}

	if db.Dialector != nil && db.Dialector.Name() == "mysql" {
		booleanQuery := BuildMySQLBooleanQuery(normalized)
		if booleanQuery != "" {
			return db.Where("MATCH ("+strings.Join(columns, ", ")+") AGAINST (? IN BOOLEAN MODE)", booleanQuery)
		}
	}

	pattern := "%" + EscapeLike(strings.ToLower(normalized)) + "%"
	conditions := make([]string, 0, len(columns))
	args := make([]interface{}, 0, len(columns))
	for _, column := range columns {
		conditions = append(conditions, "LOWER("+column+") LIKE ? ESCAPE '\\'")
		args = append(args, pattern)
	}

	return db.Where(strings.Join(conditions, " OR "), args...)
}

func BuildMySQLBooleanQuery(keyword string) string {
	parts := strings.Fields(keyword)
	if len(parts) == 0 {
		parts = []string{keyword}
	}

	terms := make([]string, 0, len(parts))
	for _, part := range parts {
		token := sanitizeBooleanToken(part)
		if utf8.RuneCountInString(token) < MinKeywordLength {
			continue
		}
		terms = append(terms, "+"+token+"*")
	}

	if len(terms) == 0 {
		token := sanitizeBooleanToken(keyword)
		if utf8.RuneCountInString(token) < MinKeywordLength {
			return ""
		}
		return "+" + token + "*"
	}

	return strings.Join(terms, " ")
}

func MatchKeywordAny(keyword string, fields ...string) bool {
	normalized := NormalizeKeyword(keyword)
	if normalized == "" {
		return true
	}
	normalized = strings.ToLower(normalized)
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), normalized) {
			return true
		}
	}
	return false
}

func sanitizeBooleanToken(token string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' {
			return r
		}
		return -1
	}, strings.TrimSpace(token))
}
