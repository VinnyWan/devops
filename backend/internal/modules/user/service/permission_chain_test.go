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
	tenant := createTenant(t, db, "default-a")

	dept := createDepartment(t, db, tenant.ID, "dept-a")
	role := createRole(t, db, tenant.ID, "role-a")
	oldPerm := createPermission(t, db, "perm-old", "user", "list")
	newPerm := createPermission(t, db, "perm-new", "user", "update")
	bindRolePermissions(t, db, role, []model.Permission{oldPerm})
	bindDepartmentRoles(t, db, dept, []model.Role{role})
	user := createUser(t, db, tenant.ID, "user-a", "user-a@example.com", false, &dept.ID)

	cacheKey := fmt.Sprintf("tenant:%d:user:perms:%d", tenant.ID, user.ID)
	mustNoError(t, redisPkg.SAdd(context.Background(), cacheKey, "user:list"))

	svc := NewRoleService(db)
	mustNoError(t, svc.AssignPermissions(tenant.ID, role.ID, []uint{newPerm.ID}))

	cached, err := redisPkg.SMembers(context.Background(), cacheKey)
	mustNoError(t, err)
	if len(cached) != 0 {
		t.Fatalf("expected permission cache invalidated, got %v", cached)
	}
}

func TestDepartmentAssignRoles_InvalidateDepartmentUserPermCache(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)
	tenant := createTenant(t, db, "default-b")

	dept := createDepartment(t, db, tenant.ID, "dept-b")
	operator := createUser(t, db, tenant.ID, "admin-b", "admin-b@example.com", true, &dept.ID)
	role1 := createRole(t, db, tenant.ID, "role-b1")
	role2 := createRole(t, db, tenant.ID, "role-b2")
	user := createUser(t, db, tenant.ID, "user-b", "user-b@example.com", false, &dept.ID)
	bindDepartmentRoles(t, db, dept, []model.Role{role1})

	cacheKey := fmt.Sprintf("tenant:%d:user:perms:%d", tenant.ID, user.ID)
	mustNoError(t, redisPkg.SAdd(context.Background(), cacheKey, "user:list"))

	svc := NewDepartmentService(repository.NewDepartmentRepo(db), repository.NewUserRepo(db))
	mustNoError(t, svc.AssignRoles(context.Background(), tenant.ID, operator.ID, dept.ID, []uint{role2.ID}))

	cached, err := redisPkg.SMembers(context.Background(), cacheKey)
	mustNoError(t, err)
	if len(cached) != 0 {
		t.Fatalf("expected permission cache invalidated, got %v", cached)
	}
}

func TestDepartmentUserTransfer_BoundaryChecks(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)
	tenant := createTenant(t, db, "default-c")

	deptA := createDepartment(t, db, tenant.ID, "dept-c-a")
	deptB := createDepartment(t, db, tenant.ID, "dept-c-b")
	deptAdminRole := createRole(t, db, tenant.ID, "DEPT_ADMIN")
	operator := createUser(t, db, tenant.ID, "operator-c", "operator-c@example.com", false, &deptA.ID)
	target := createUser(t, db, tenant.ID, "target-c", "target-c@example.com", false, &deptA.ID)
	bindUserRoles(t, db, operator, []model.Role{deptAdminRole})

	svc := NewDepartmentUserService(repository.NewUserRepo(db), repository.NewDepartmentRepo(db))

	err := svc.Transfer(tenant.ID, operator.ID, &TransferUserDepartmentRequest{
		UserID:         target.ID,
		ToDepartmentID: deptB.ID,
	})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission denied, got %v", err)
	}

	err = svc.Transfer(tenant.ID, operator.ID, &TransferUserDepartmentRequest{
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
	tenant := createTenant(t, db, "default-d")

	deptA := createDepartment(t, db, tenant.ID, "dept-d-a")
	deptB := createDepartment(t, db, tenant.ID, "dept-d-b")
	admin := createUser(t, db, tenant.ID, "admin-d", "admin-d@example.com", true, &deptA.ID)
	target := createUser(t, db, tenant.ID, "target-d", "target-d@example.com", false, &deptA.ID)

	cacheKey := fmt.Sprintf("tenant:%d:user:perms:%d", tenant.ID, target.ID)
	mustNoError(t, redisPkg.SAdd(context.Background(), cacheKey, "department:update"))

	svc := NewDepartmentUserService(repository.NewUserRepo(db), repository.NewDepartmentRepo(db))
	mustNoError(t, svc.Transfer(tenant.ID, admin.ID, &TransferUserDepartmentRequest{
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
	tenant := createTenant(t, db, "default-e")

	deptA := createDepartment(t, db, tenant.ID, "dept-e-a")
	deptB := createDepartment(t, db, tenant.ID, "dept-e-b")
	admin := createUser(t, db, tenant.ID, "admin-e", "admin-e@example.com", true, &deptA.ID)
	target := createUser(t, db, tenant.ID, "target-e", "target-e@example.com", false, &deptA.ID)

	legacyRole := createRole(t, db, tenant.ID, "legacy-role")
	deptRole := createRole(t, db, tenant.ID, "dept-role")
	readOnlyRole := createRole(t, db, tenant.ID, "READ_ONLY")

	legacyPerm := createPermission(t, db, "legacy", "user", "delete")
	deptPerm := createPermission(t, db, "dept", "department", "list")
	readOnlyPerm := createPermission(t, db, "readonly", "cluster", "list")

	bindRolePermissions(t, db, legacyRole, []model.Permission{legacyPerm})
	bindRolePermissions(t, db, deptRole, []model.Permission{deptPerm})
	bindRolePermissions(t, db, readOnlyRole, []model.Permission{readOnlyPerm})
	bindDepartmentRoles(t, db, deptB, []model.Role{deptRole})
	bindUserRoles(t, db, target, []model.Role{legacyRole})

	userSvc := NewUserService(db)
	before, err := userSvc.GetUserPermissionCodes(context.Background(), tenant.ID, target.ID)
	mustNoError(t, err)
	if !containsPermissionCode(before, "user:delete") {
		t.Fatalf("expected legacy permission before transfer, got %v", before)
	}

	deptUserSvc := NewDepartmentUserService(repository.NewUserRepo(db), repository.NewDepartmentRepo(db))
	mustNoError(t, deptUserSvc.Transfer(tenant.ID, admin.ID, &TransferUserDepartmentRequest{
		UserID:         target.ID,
		ToDepartmentID: deptB.ID,
	}))

	reloaded, err := repository.NewUserRepo(db).GetByID(target.ID)
	mustNoError(t, err)
	if reloaded.PrimaryDeptID == nil || *reloaded.PrimaryDeptID != deptB.ID {
		t.Fatalf("expected department changed to %d, got %+v", deptB.ID, reloaded.PrimaryDeptID)
	}
	if len(reloaded.Roles) != 0 {
		t.Fatalf("expected direct roles cleared after transfer, got %d", len(reloaded.Roles))
	}

	after, err := userSvc.GetUserPermissionCodes(context.Background(), tenant.ID, target.ID)
	mustNoError(t, err)
	assertPermissionSet(t, after, []string{"department:list", "cluster:list"})
	if containsPermissionCode(after, "user:delete") {
		t.Fatalf("expected legacy permission removed after transfer, got %v", after)
	}
}

func TestGetUserPermissionCodes_IncludeGlobalReadOnlyRole(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)
	tenant := createTenant(t, db, "default-f")

	readOnlyRole := model.Role{
		Name:        "READ_ONLY",
		DisplayName: "只读用户",
		Type:        "system",
	}
	mustNoError(t, db.Create(&readOnlyRole).Error)

	readOnlyPerm := createPermission(t, db, "readonly-global", "cluster", "list")
	bindRolePermissions(t, db, readOnlyRole, []model.Permission{readOnlyPerm})

	user := createUser(t, db, tenant.ID, "user-f", "user-f@example.com", false, nil)

	userSvc := NewUserService(db)
	codes, err := userSvc.GetUserPermissionCodes(context.Background(), tenant.ID, user.ID)
	mustNoError(t, err)
	assertPermissionSet(t, codes, []string{"cluster:list"})
}

func TestAssignRolesInTenant_AllowsGlobalRole(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)
	tenant := createTenant(t, db, "default-g")

	globalRole := model.Role{
		Name:        "TENANT_ADMIN",
		DisplayName: "租户管理员",
		Type:        "system",
	}
	mustNoError(t, db.Create(&globalRole).Error)

	user := createUser(t, db, tenant.ID, "user-g", "user-g@example.com", false, nil)

	userSvc := NewUserService(db)
	mustNoError(t, userSvc.AssignRoles(context.Background(), tenant.ID, user.ID, user.ID, []uint{globalRole.ID}))

	reloaded, err := repository.NewUserRepo(db).GetByIDInTenant(tenant.ID, user.ID)
	mustNoError(t, err)
	if len(reloaded.Roles) != 1 || reloaded.Roles[0].ID != globalRole.ID {
		t.Fatalf("expected global role assigned, got %+v", reloaded.Roles)
	}
}

func TestUserService_ListUsers_RespectsDepartmentTreeScope(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)
	tenant := createTenant(t, db, "default-h")

	rootDept := createDepartment(t, db, tenant.ID, "dept-h-root")
	childDept := createDepartmentWithParent(t, db, tenant.ID, "dept-h-child", rootDept.ID)
	otherDept := createDepartment(t, db, tenant.ID, "dept-h-other")

	scopeRole := createRoleWithScope(t, db, tenant.ID, "scope-tree", model.DataScopeDepartmentTree)
	operator := createUser(t, db, tenant.ID, "operator-h", "operator-h@example.com", false, &rootDept.ID)
	rootUser := createUser(t, db, tenant.ID, "root-h", "root-h@example.com", false, &rootDept.ID)
	childUser := createUser(t, db, tenant.ID, "child-h", "child-h@example.com", false, &childDept.ID)
	otherUser := createUser(t, db, tenant.ID, "other-h", "other-h@example.com", false, &otherDept.ID)
	bindUserRoles(t, db, operator, []model.Role{scopeRole})

	svc := NewUserService(db)
	items, total, err := svc.ListUsers(context.Background(), tenant.ID, operator.ID, 1, 20, "")
	mustNoError(t, err)
	if total != 3 {
		t.Fatalf("expected 3 accessible users, got %d", total)
	}
	if !containsUserID(items, operator.ID) || !containsUserID(items, rootUser.ID) || !containsUserID(items, childUser.ID) {
		t.Fatalf("expected root subtree users visible, got %+v", extractUserIDs(items))
	}
	if containsUserID(items, otherUser.ID) {
		t.Fatalf("expected sibling department user hidden, got %+v", extractUserIDs(items))
	}
}

func TestDepartmentUserService_Create_UsesScopeNotRoleName(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)
	tenant := createTenant(t, db, "default-i")

	dept := createDepartment(t, db, tenant.ID, "dept-i")
	scopeRole := createRoleWithScope(t, db, tenant.ID, "custom_scope_role", model.DataScopeDepartmentTree)
	operator := createUser(t, db, tenant.ID, "operator-i", "operator-i@example.com", false, &dept.ID)
	bindUserRoles(t, db, operator, []model.Role{scopeRole})

	svc := NewDepartmentUserService(repository.NewUserRepo(db), repository.NewDepartmentRepo(db))
	created, err := svc.Create(tenant.ID, operator.ID, &CreateDeptUserRequest{
		Username:     "new-i",
		Password:     "Passw0rd!",
		Email:        "new-i@example.com",
		Name:         "new-i",
		DepartmentID: dept.ID,
	})
	mustNoError(t, err)
	if created.PrimaryDeptID == nil || *created.PrimaryDeptID != dept.ID {
		t.Fatalf("expected created user in dept %d, got %+v", dept.ID, created.PrimaryDeptID)
	}
}

func TestDepartmentUserService_List_SelectedDepartmentIncludesDescendants(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)
	tenant := createTenant(t, db, "default-i-list-tree")

	rootDept := createDepartment(t, db, tenant.ID, "dept-i-root")
	childDept := createDepartmentWithParent(t, db, tenant.ID, "dept-i-child", rootDept.ID)
	grandChildDept := createDepartmentWithParent(t, db, tenant.ID, "dept-i-grand-child", childDept.ID)
	otherDept := createDepartment(t, db, tenant.ID, "dept-i-other")

	scopeRole := createRoleWithScope(t, db, tenant.ID, "scope-tree-list", model.DataScopeDepartmentTree)
	operator := createUser(t, db, tenant.ID, "operator-i-list", "operator-i-list@example.com", false, &rootDept.ID)
	rootUser := createUser(t, db, tenant.ID, "root-i-list", "root-i-list@example.com", false, &rootDept.ID)
	childUser := createUser(t, db, tenant.ID, "child-i-list", "child-i-list@example.com", false, &childDept.ID)
	grandChildUser := createUser(t, db, tenant.ID, "grand-i-list", "grand-i-list@example.com", false, &grandChildDept.ID)
	otherUser := createUser(t, db, tenant.ID, "other-i-list", "other-i-list@example.com", false, &otherDept.ID)
	bindUserRoles(t, db, operator, []model.Role{scopeRole})

	svc := NewDepartmentUserService(repository.NewUserRepo(db), repository.NewDepartmentRepo(db))
	items, total, err := svc.List(tenant.ID, operator.ID, rootDept.ID, 1, 20, "")
	mustNoError(t, err)
	if total != 4 {
		t.Fatalf("expected 4 users in root subtree, got %d", total)
	}
	if !containsUserID(items, operator.ID) || !containsUserID(items, rootUser.ID) || !containsUserID(items, childUser.ID) || !containsUserID(items, grandChildUser.ID) {
		t.Fatalf("expected root subtree users visible, got %+v", extractUserIDs(items))
	}
	if containsUserID(items, otherUser.ID) {
		t.Fatalf("expected unrelated department user hidden, got %+v", extractUserIDs(items))
	}
	for _, item := range items {
		if item.Department == nil || item.Department.Name == "" {
			t.Fatalf("expected department preloaded for user %d, got %+v", item.ID, item.Department)
		}
	}
}

func TestDepartmentUserService_List_SelectedDepartmentStillRespectsScope(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)
	tenant := createTenant(t, db, "default-i-list-self")

	rootDept := createDepartment(t, db, tenant.ID, "dept-i-self-root")
	childDept := createDepartmentWithParent(t, db, tenant.ID, "dept-i-self-child", rootDept.ID)

	operator := createUser(t, db, tenant.ID, "operator-i-self", "operator-i-self@example.com", false, &rootDept.ID)
	rootUser := createUser(t, db, tenant.ID, "root-i-self", "root-i-self@example.com", false, &rootDept.ID)
	childUser := createUser(t, db, tenant.ID, "child-i-self", "child-i-self@example.com", false, &childDept.ID)

	svc := NewDepartmentUserService(repository.NewUserRepo(db), repository.NewDepartmentRepo(db))
	items, total, err := svc.List(tenant.ID, operator.ID, rootDept.ID, 1, 20, "")
	mustNoError(t, err)
	if total != 2 {
		t.Fatalf("expected only self department users, got %d", total)
	}
	if !containsUserID(items, operator.ID) || !containsUserID(items, rootUser.ID) {
		t.Fatalf("expected self department users visible, got %+v", extractUserIDs(items))
	}
	if containsUserID(items, childUser.ID) {
		t.Fatalf("expected child department user hidden under self_department scope, got %+v", extractUserIDs(items))
	}
}

func TestRoleRepo_GetRoleUsersInTenant_FiltersGlobalRoleAssociations(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)
	tenantA := createTenant(t, db, "default-j-a")
	tenantB := createTenant(t, db, "default-j-b")

	globalRole := model.Role{
		Name:        "GLOBAL_VIEWER",
		DisplayName: "GLOBAL_VIEWER",
		Type:        "system",
		DataScope:   model.DataScopeTenant,
	}
	mustNoError(t, db.Create(&globalRole).Error)

	userA := createUser(t, db, tenantA.ID, "user-j-a", "user-j-a@example.com", false, nil)
	userB := createUser(t, db, tenantB.ID, "user-j-b", "user-j-b@example.com", false, nil)
	bindUserRoles(t, db, userA, []model.Role{globalRole})
	bindUserRoles(t, db, userB, []model.Role{globalRole})

	items, err := repository.NewRoleRepo(db).GetRoleUsersInTenant(tenantA.ID, globalRole.ID)
	mustNoError(t, err)
	if len(items) != 1 || items[0].ID != userA.ID {
		t.Fatalf("expected only tenant A user, got %+v", extractUserIDs(items))
	}
}

func TestUserRepo_InTenant_RejectsZeroTenant(t *testing.T) {
	db := setupUserModuleTestDB(t)
	setupRedisForTest(t)
	tenant := createTenant(t, db, "default-k")
	user := createUser(t, db, tenant.ID, "user-k", "user-k@example.com", false, nil)

	_, err := repository.NewUserRepo(db).GetByIDInTenant(0, user.ID)
	if !errors.Is(err, repository.ErrTenantScopeRequired) {
		t.Fatalf("expected ErrTenantScopeRequired, got %v", err)
	}
}

func setupUserModuleTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)), &gorm.Config{})
	mustNoError(t, err)
	mustNoError(t, db.AutoMigrate(&model.Tenant{}, &model.Permission{}, &model.Role{}, &model.Department{}, &model.User{}))
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

func createDepartment(t *testing.T, db *gorm.DB, tenantID uint, name string) model.Department {
	t.Helper()
	dept := model.Department{Name: name, TenantID: &tenantID}
	mustNoError(t, db.Create(&dept).Error)
	return dept
}

func createDepartmentWithParent(t *testing.T, db *gorm.DB, tenantID uint, name string, parentID uint) model.Department {
	t.Helper()
	dept := model.Department{Name: name, TenantID: &tenantID, ParentID: &parentID}
	mustNoError(t, db.Create(&dept).Error)
	return dept
}

func createRole(t *testing.T, db *gorm.DB, tenantID uint, name string) model.Role {
	return createRoleWithScope(t, db, tenantID, name, model.DataScopeSelfDepartment)
}

func createRoleWithScope(t *testing.T, db *gorm.DB, tenantID uint, name string, dataScope model.DataScope) model.Role {
	t.Helper()
	role := model.Role{
		Name:        name,
		DisplayName: name,
		Type:        "custom",
		TenantID:    &tenantID,
		DataScope:   dataScope,
	}
	mustNoError(t, db.Create(&role).Error)
	return role
}

func createPermission(t *testing.T, db *gorm.DB, name, resource, action string) model.Permission {
	t.Helper()
	perm := model.Permission{Name: name, Resource: resource, Action: action}
	mustNoError(t, db.Create(&perm).Error)
	return perm
}

func createUser(t *testing.T, db *gorm.DB, tenantID uint, username, email string, isAdmin bool, deptID *uint) model.User {
	t.Helper()
	user := model.User{
		TenantID:     &tenantID,
		Username:     username,
		Password:     "hashed",
		Email:        email,
		Name:         username,
		AuthType:     model.AuthTypeLocal,
		Status:       "active",
		IsAdmin:      isAdmin,
		PrimaryDeptID: deptID,
	}
	mustNoError(t, db.Create(&user).Error)
	return user
}

func createTenant(t *testing.T, db *gorm.DB, code string) model.Tenant {
	t.Helper()
	tenant := model.Tenant{
		Name:   "Tenant-" + code,
		Code:   code,
		Status: "active",
	}
	mustNoError(t, db.Create(&tenant).Error)
	return tenant
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

func containsUserID(users []model.User, target uint) bool {
	for _, user := range users {
		if user.ID == target {
			return true
		}
	}
	return false
}

func extractUserIDs(users []model.User) []uint {
	ids := make([]uint, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
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
