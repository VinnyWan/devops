package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireUintQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/?other=1", nil)
		c.Request = req

		value, ok := RequireUintQuery(c, "id")
		if ok || value != 0 {
			t.Fatalf("expected missing param to return false and 0, got ok=%v value=%d", ok, value)
		}
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/?id=abc", nil)
		c.Request = req

		value, ok := RequireUintQuery(c, "id")
		if ok || value != 0 {
			t.Fatalf("expected invalid param to return false and 0, got ok=%v value=%d", ok, value)
		}
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("valid param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/?id=42", nil)
		c.Request = req

		value, ok := RequireUintQuery(c, "id")
		if !ok || value != 42 {
			t.Fatalf("expected ok and value=42, got ok=%v value=%d", ok, value)
		}
	})
}

func TestOptionalUintQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/?other=1", nil)
		c.Request = req

		value, provided, err := OptionalUintQuery(c, "deptId")
		if err != nil || provided || value != 0 {
			t.Fatalf("expected missing param to return provided=false value=0 err=nil")
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/?deptId=bad", nil)
		c.Request = req

		_, _, err := OptionalUintQuery(c, "deptId")
		if err == nil {
			t.Fatalf("expected error for invalid param")
		}
	})

	t.Run("valid param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/?deptId=7", nil)
		c.Request = req

		value, provided, err := OptionalUintQuery(c, "deptId")
		if err != nil || !provided || value != 7 {
			t.Fatalf("expected provided=true value=7 err=nil, got provided=%v value=%d err=%v", provided, value, err)
		}
	})
}

func TestParsePageParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("defaults", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		c.Request = req

		page, pageSize, err := ParsePageParams(c, 1, 10, 100)
		if err != nil || page != 1 || pageSize != 10 {
			t.Fatalf("expected defaults page=1 pageSize=10 err=nil, got page=%d pageSize=%d err=%v", page, pageSize, err)
		}
	})

	t.Run("invalid page", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/?page=0&pageSize=10", nil)
		c.Request = req

		_, _, err := ParsePageParams(c, 1, 10, 100)
		if err == nil {
			t.Fatalf("expected error for invalid page")
		}
	})

	t.Run("clamp pageSize", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/?page=1&pageSize=1000", nil)
		c.Request = req

		page, pageSize, err := ParsePageParams(c, 1, 10, 200)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if page != 1 || pageSize != 200 {
			t.Fatalf("expected page=1 pageSize=200, got page=%d pageSize=%d", page, pageSize)
		}
	})
}
