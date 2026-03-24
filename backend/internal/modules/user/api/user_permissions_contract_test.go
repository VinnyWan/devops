package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"devops-platform/internal/modules/user/model"
	redisPkg "devops-platform/internal/pkg/redis"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redisv9 "github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetUserPermissions_CurrentUserContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupUserAPITestDB(t)
	setupUserAPITestRedis(t)
	resetUserServiceState(db)

	role := model.Role{Name: "api-role", DisplayName: "api-role", Type: "custom"}
	mustNoErrorAPI(t, db.Create(&role).Error)
	perm := model.Permission{Name: "查看用户", Resource: "user", Action: "list"}
	mustNoErrorAPI(t, db.Create(&perm).Error)
	mustNoErrorAPI(t, db.Model(&role).Association("Permissions").Replace([]model.Permission{perm}))

	user := model.User{
		Username: "api-user",
		Password: "hashed",
		Email:    "api-user@example.com",
		Name:     "api-user",
		AuthType: model.AuthTypeLocal,
		Status:   "active",
	}
	mustNoErrorAPI(t, db.Create(&user).Error)
	mustNoErrorAPI(t, db.Model(&user).Association("Roles").Replace([]model.Role{role}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/permissions", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", user.ID)

	GetUserPermissions(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    []string `json:"data"`
	}
	mustNoErrorAPI(t, json.Unmarshal(w.Body.Bytes(), &resp))
	if resp.Code != 200 {
		t.Fatalf("expected code 200, got %d", resp.Code)
	}
	if len(resp.Data) != 1 || resp.Data[0] != "user:list" {
		t.Fatalf("expected [user:list], got %v", resp.Data)
	}
}

func TestGetUserPermissions_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupUserAPITestDB(t)
	setupUserAPITestRedis(t)
	resetUserServiceState(db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/permissions", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	GetUserPermissions(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "未认证") {
		t.Fatalf("expected unauthorized message, got %s", w.Body.String())
	}
}

func setupUserAPITestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)), &gorm.Config{})
	mustNoErrorAPI(t, err)
	mustNoErrorAPI(t, db.AutoMigrate(&model.Permission{}, &model.Role{}, &model.Department{}, &model.User{}))
	return db
}

func setupUserAPITestRedis(t *testing.T) {
	t.Helper()
	mini, err := miniredis.Run()
	mustNoErrorAPI(t, err)
	redisPkg.Client = redisv9.NewClient(&redisv9.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisPkg.Client.Close()
		mini.Close()
	})
}

func resetUserServiceState(database *gorm.DB) {
	SetDB(database)
	userService = nil
	userOnce = sync.Once{}
	deptUserService = nil
	deptUserOnce = sync.Once{}
}

func mustNoErrorAPI(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
