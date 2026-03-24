package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	redisPkg "devops-platform/internal/pkg/redis"

	miniredis "github.com/alicebob/miniredis/v2"
	redisv9 "github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRoleAssignPermissions_InvalidateAffectedUserPermCache(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)

	dept := createDepartment(t, db, "dept-a")
	role := createRole(t, db, "role-a")
	oldPerm := createPermission(t, db, "perm-old", "user", "list")
	newPerm := createPermission(t, db, "perm-new", "user", "update")
	bindRolePermissions(t, db, role, []model.Permission{oldPerm})
	bindDepartmentRoles(t, db, dept, []model.Role{role})
	user := createUser(t, db, "user-a", "user-a@example.com", false, &dept.ID)

	cacheKey := fmt.Sprintf("user:perms:%d", user.ID)
	mustNoError(t, redisPkg.SAdd(context.Background(), cacheKey, "user:list"))

	svc := NewRoleService(db)
	mustNoError(t, svc.AssignPermissions(role.ID, []uint{newPerm.ID}))

	cached, err := redisPkg.SMembers(context.Background(), cacheKey)
	mustNoError(t, err)
	if len(cached) != 0 {
		t.Fatalf("expected permission cache invalidated, got %v", cached)
	}
}

func TestDepartmentAssignRoles_InvalidateDepartmentUserPermCache(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)

	dept := createDepartment(t, db, "dept-b")
	role1 := createRole(t, db, "role-b1")
	role2 := createRole(t, db, "role-b2")
	user := createUser(t, db, "user-b", "user-b@example.com", false, &dept.ID)
	bindDepartmentRoles(t, db, dept, []model.Role{role1})

	cacheKey := fmt.Sprintf("user:perms:%d", user.ID)
	mustNoError(t, redisPkg.SAdd(context.Background(), cacheKey, "user:list"))

	svc := NewDepartmentService(repository.NewDepartmentRepo(db), repository.NewUserRepo(db))
	mustNoError(t, svc.AssignRoles(context.Background(), dept.ID, []uint{role2.ID}))

	cached, err := redisPkg.SMembers(context.Background(), cacheKey)
	mustNoError(t, err)
	if len(cached) != 0 {
		t.Fatalf("expected permission cache invalidated, got %v", cached)
	}
}

func TestDepartmentUserTransfer_BoundaryChecks(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)

	deptA := createDepartment(t, db, "dept-c-a")
	deptB := createDepartment(t, db, "dept-c-b")
	deptAdminRole := createRole(t, db, "DEPT_ADMIN")
	operator := createUser(t, db, "operator-c", "operator-c@example.com", false, &deptA.ID)
	target := createUser(t, db, "target-c", "target-c@example.com", false, &deptA.ID)
	bindUserRoles(t, db, operator, []model.Role{deptAdminRole})

	svc := NewDepartmentUserService(repository.NewUserRepo(db), repository.NewDepartmentRepo(db))

	err := svc.Transfer(operator.ID, &TransferUserDepartmentRequest{
		UserID:         target.ID,
		ToDepartmentID: deptB.ID,
	})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission denied, got %v", err)
	}

	err = svc.Transfer(operator.ID, &TransferUserDepartmentRequest{
		UserID:         target.ID,
		ToDepartmentID: deptA.ID,
	})
	if err == nil || !strings.Contains(err.Error(), "target department must be different") {
		t.Fatalf("expected target department must be different, got %v", err)
	}
}

func TestDepartmentUserTransfer_InvalidateMovedUserPermCache(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)

	deptA := createDepartment(t, db, "dept-d-a")
	deptB := createDepartment(t, db, "dept-d-b")
	admin := createUser(t, db, "admin-d", "admin-d@example.com", true, &deptA.ID)
	target := createUser(t, db, "target-d", "target-d@example.com", false, &deptA.ID)

	cacheKey := fmt.Sprintf("user:perms:%d", target.ID)
	mustNoError(t, redisPkg.SAdd(context.Background(), cacheKey, "department:update"))

	svc := NewDepartmentUserService(repository.NewUserRepo(db), repository.NewDepartmentRepo(db))
	mustNoError(t, svc.Transfer(admin.ID, &TransferUserDepartmentRequest{
		UserID:         target.ID,
		ToDepartmentID: deptB.ID,
	}))

	cached, err := redisPkg.SMembers(context.Background(), cacheKey)
	mustNoError(t, err)
	if len(cached) != 0 {
		t.Fatalf("expected permission cache invalidated, got %v", cached)
	}
}

func TestDepartmentUserTransfer_PermissionConvergence(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)

	deptA := createDepartment(t, db, "dept-e-a")
	deptB := createDepartment(t, db, "dept-e-b")
	admin := createUser(t, db, "admin-e", "admin-e@example.com", true, &deptA.ID)
	target := createUser(t, db, "target-e", "target-e@example.com", false, &deptA.ID)

	legacyRole := createRole(t, db, "legacy-role")
	deptRole := createRole(t, db, "dept-role")
	readOnlyRole := createRole(t, db, "READ_ONLY")

	legacyPerm := createPermission(t, db, "legacy", "user", "delete")
	deptPerm := createPermission(t, db, "dept", "department", "list")
	readOnlyPerm := createPermission(t, db, "readonly", "cluster", "list")

	bindRolePermissions(t, db, legacyRole, []model.Permission{legacyPerm})
	bindRolePermissions(t, db, deptRole, []model.Permission{deptPerm})
	bindRolePermissions(t, db, readOnlyRole, []model.Permission{readOnlyPerm})
	bindDepartmentRoles(t, db, deptB, []model.Role{deptRole})
	bindUserRoles(t, db, target, []model.Role{legacyRole})

	userSvc := NewUserService(db)
	before, err := userSvc.GetUserPermissionCodes(context.Background(), target.ID)
	mustNoError(t, err)
	if !containsPermissionCode(before, "user:delete") {
		t.Fatalf("expected legacy permission before transfer, got %v", before)
	}

	deptUserSvc := NewDepartmentUserService(repository.NewUserRepo(db), repository.NewDepartmentRepo(db))
	mustNoError(t, deptUserSvc.Transfer(admin.ID, &TransferUserDepartmentRequest{
		UserID:         target.ID,
		ToDepartmentID: deptB.ID,
	}))

	reloaded, err := repository.NewUserRepo(db).GetByID(target.ID)
	mustNoError(t, err)
	if reloaded.DepartmentID == nil || *reloaded.DepartmentID != deptB.ID {
		t.Fatalf("expected department changed to %d, got %+v", deptB.ID, reloaded.DepartmentID)
	}
	if len(reloaded.Roles) != 0 {
		t.Fatalf("expected direct roles cleared after transfer, got %d", len(reloaded.Roles))
	}

	after, err := userSvc.GetUserPermissionCodes(context.Background(), target.ID)
	mustNoError(t, err)
	assertPermissionSet(t, after, []string{"department:list", "cluster:list"})
	if containsPermissionCode(after, "user:delete") {
		t.Fatalf("expected legacy permission removed after transfer, got %v", after)
	}
}

func setupUserModuleTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)), &gorm.Config{})
	mustNoError(t, err)
	mustNoError(t, db.AutoMigrate(&model.Permission{}, &model.Role{}, &model.Department{}, &model.User{}))
	return db
}

func setupRedisForTest(t *testing.T) {
	t.Helper()

	mini, err := miniredis.Run()
	mustNoError(t, err)
	redisPkg.Client = redisv9.NewClient(&redisv9.Options{Addr: mini.Addr()})

	t.Cleanup(func() {
		_ = redisPkg.Client.Close()
		mini.Close()
	})
}

func createDepartment(t *testing.T, db *gorm.DB, name string) model.Department {
	t.Helper()
	dept := model.Department{Name: name}
	mustNoError(t, db.Create(&dept).Error)
	return dept
}

func createRole(t *testing.T, db *gorm.DB, name string) model.Role {
	t.Helper()
	role := model.Role{Name: name, DisplayName: name, Type: "custom"}
	mustNoError(t, db.Create(&role).Error)
	return role
}

func createPermission(t *testing.T, db *gorm.DB, name, resource, action string) model.Permission {
	t.Helper()
	perm := model.Permission{Name: name, Resource: resource, Action: action}
	mustNoError(t, db.Create(&perm).Error)
	return perm
}

func createUser(t *testing.T, db *gorm.DB, username, email string, isAdmin bool, deptID *uint) model.User {
	t.Helper()
	user := model.User{
		Username:     username,
		Password:     "hashed",
		Email:        email,
		Name:         username,
		AuthType:     model.AuthTypeLocal,
		Status:       "active",
		IsAdmin:      isAdmin,
		DepartmentID: deptID,
	}
	mustNoError(t, db.Create(&user).Error)
	return user
}

func bindRolePermissions(t *testing.T, db *gorm.DB, role model.Role, perms []model.Permission) {
	t.Helper()
	mustNoError(t, db.Model(&role).Association("Permissions").Replace(perms))
}

func bindDepartmentRoles(t *testing.T, db *gorm.DB, dept model.Department, roles []model.Role) {
	t.Helper()
	mustNoError(t, db.Model(&dept).Association("Roles").Replace(roles))
}

func bindUserRoles(t *testing.T, db *gorm.DB, user model.User, roles []model.Role) {
	t.Helper()
	mustNoError(t, db.Model(&user).Association("Roles").Replace(roles))
}

func containsPermissionCode(codes []string, target string) bool {
	for _, code := range codes {
		if code == target {
			return true
		}
	}
	return false
}

func assertPermissionSet(t *testing.T, got []string, expected []string) {
	t.Helper()
	gotSet := make(map[string]struct{}, len(got))
	for _, item := range got {
		gotSet[item] = struct{}{}
	}
	if len(gotSet) != len(expected) {
		t.Fatalf("expected %d unique permissions, got %v", len(expected), got)
	}
	for _, item := range expected {
		if _, ok := gotSet[item]; !ok {
			t.Fatalf("expected permission %s in %v", item, got)
		}
	}
}

func mustNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}
