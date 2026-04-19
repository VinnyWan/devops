package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/pkg/redis"
	"devops-platform/internal/pkg/utils"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo       *repository.UserRepo
	roleRepo       *repository.RoleRepo
	permissionRepo *repository.PermissionRepo
	userDeptRepo   *repository.UserDepartmentRepo
	deptRepo       *repository.DepartmentRepo
	fieldPermRepo  *repository.FieldPermissionRepo
	scopeSvc       *AccessScopeService
}

func NewUserService(db *gorm.DB) *UserService {
	userRepo := repository.NewUserRepo(db)
	deptRepo := repository.NewDepartmentRepo(db)
	userDeptRepo := repository.NewUserDepartmentRepo(db)
	roleRepo := repository.NewRoleRepo(db)
	return &UserService{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: repository.NewPermissionRepo(db),
		userDeptRepo:   userDeptRepo,
		deptRepo:       deptRepo,
		fieldPermRepo:  repository.NewFieldPermissionRepo(db),
		scopeSvc:       NewAccessScopeService(userRepo, deptRepo, userDeptRepo).WithRoleRepo(roleRepo),
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

// Register 注册新用户（仅限本地认证）
func (s *UserService) Register(tenantID uint, req *RegisterRequest) (*model.User, error) {
	if err := utils.ValidatePasswordComplexity(req.Password); err != nil {
		return nil, err
	}

	existingUser, err := s.userRepo.GetByUsernameInTenant(tenantID, req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	existingUser, err = s.userRepo.GetByEmailInTenant(tenantID, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &model.User{
		TenantID: &tenantID,
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Name:     req.Name,
		AuthType: model.AuthTypeLocal,
		Status:   "active",
		IsAdmin:  false,
		IsLocked: false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID 根据ID获取用户信息 (带缓存)
func (s *UserService) GetUserByID(ctx context.Context, tenantID uint, id uint) (*model.User, error) {
	// TODO: 实现用户信息缓存
	return s.userRepo.GetByIDInTenant(tenantID, id)
}

// GetUserByUsername 根据用户名获取用户信息
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	return s.userRepo.GetByUsername(username)
}

// ListUsers 获取用户列表
func (s *UserService) GetAccessibleUserByID(ctx context.Context, tenantID uint, operatorID uint, id uint) (*model.User, error) {
	if err := s.scopeSvc.EnsureUserAccess(ctx, tenantID, operatorID, id); err != nil {
		return nil, err
	}
	return s.userRepo.GetByIDInTenant(tenantID, id)
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, tenantID uint, operatorID uint, page, pageSize int, keyword string, departmentID *uint) ([]model.User, int64, error) {
	// 按部门筛选：收集该部门及所有子部门ID，查询这些部门下的所有用户
	if departmentID != nil {
		deptIDs, err := s.deptRepo.GetDescendantIDsInTenant(tenantID, *departmentID)
		if err != nil {
			return nil, 0, err
		}
		return s.userRepo.ListByDepartmentIDsInTenant(tenantID, deptIDs, page, pageSize, keyword)
	}

	scope, err := s.scopeSvc.Resolve(ctx, tenantID, operatorID)
	if err != nil {
		return nil, 0, err
	}
	if scope.AllowsAll() {
		return s.userRepo.ListInTenant(tenantID, page, pageSize, keyword)
	}
	return s.userRepo.ListByDepartmentIDsInTenant(tenantID, scope.DepartmentIDs, page, pageSize, keyword)
}

func (s *UserService) UpdateUserByRequest(ctx context.Context, tenantID uint, operatorID uint, req *UpdateUserRequest) error {
	_, err := s.userRepo.GetByIDInTenant(tenantID, req.ID)
	if err != nil {
		return err
	}
	if err := s.scopeSvc.EnsureUserAccess(ctx, tenantID, operatorID, req.ID); err != nil {
		return err
	}

	updates := make(map[string]interface{})

	if req.Username != nil {
		if *req.Username == "" {
			return errors.New("username cannot be empty")
		}
		updates["username"] = *req.Username
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		if *req.Email == "" {
			return errors.New("email cannot be empty")
		}
		updates["email"] = *req.Email
	}
	if req.Status != nil {
		if *req.Status == "" {
			return errors.New("status cannot be empty")
		}
		updates["status"] = *req.Status
	}

	if req.PrimaryDeptID != nil {
		if *req.PrimaryDeptID == 0 {
			updates["primary_dept_id"] = nil
		} else {
			if err := s.scopeSvc.EnsureDepartmentAccess(ctx, tenantID, operatorID, *req.PrimaryDeptID); err != nil {
				return err
			}
			updates["primary_dept_id"] = *req.PrimaryDeptID
		}
	}

	if len(updates) == 0 {
		return nil
	}

	// 使得相关缓存失效
	s.invalidateUserCache(context.Background(), tenantID, req.ID)
	s.invalidateUserPermsCache(context.Background(), tenantID, req.ID)

	return s.userRepo.UpdateByIDInTenant(tenantID, req.ID, updates)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, tenantID uint, operatorID uint, id uint) error {
	if err := s.scopeSvc.EnsureUserAccess(ctx, tenantID, operatorID, id); err != nil {
		return err
	}
	s.invalidateUserCache(context.Background(), tenantID, id)
	s.invalidateUserPermsCache(context.Background(), tenantID, id)
	return s.userRepo.DeleteInTenant(tenantID, id)
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(tenantID uint, userID uint, req *ChangePasswordRequest) error {
	if err := utils.ValidatePasswordComplexity(req.NewPassword); err != nil {
		return err
	}

	user, err := s.userRepo.GetByIDInTenant(tenantID, userID)
	if err != nil {
		return err
	}

	if user.AuthType != model.AuthTypeLocal {
		return errors.New("only local users can change password")
	}

	if !utils.CheckPassword(req.OldPassword, user.Password) {
		return errors.New("old password is incorrect")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePasswordInTenant(tenantID, userID, hashedPassword); err != nil {
		return err
	}
	// 改密后强制全端登出
	if err := NewSessionService().RevokeAllUserSessions(context.Background(), tenantID, userID); err != nil {
		return fmt.Errorf("revoke sessions after password change failed: %w", err)
	}
	return nil
}

// ResetPassword 重置密码（管理员操作）
func (s *UserService) ResetPassword(ctx context.Context, tenantID uint, operatorID uint, userID uint, newPassword string) error {
	if err := utils.ValidatePasswordComplexity(newPassword); err != nil {
		return err
	}
	if err := s.scopeSvc.EnsureUserAccess(ctx, tenantID, operatorID, userID); err != nil {
		return err
	}

	user, err := s.userRepo.GetByIDInTenant(tenantID, userID)
	if err != nil {
		return err
	}

	if user.AuthType != model.AuthTypeLocal {
		return errors.New("only local users can reset password")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePasswordInTenant(tenantID, userID, hashedPassword); err != nil {
		return err
	}
	// 重置密码后强制全端登出
	if err := NewSessionService().RevokeAllUserSessions(context.Background(), tenantID, userID); err != nil {
		return fmt.Errorf("revoke sessions after password reset failed: %w", err)
	}
	return nil
}

// AssignRoles 分配角色
func (s *UserService) AssignRoles(ctx context.Context, tenantID uint, operatorID uint, userID uint, roleIDs []uint) error {
	if err := s.scopeSvc.EnsureUserAccess(ctx, tenantID, operatorID, userID); err != nil {
		return err
	}
	// 验证所有角色是否存在
	for _, roleID := range roleIDs {
		_, err := s.roleRepo.GetByIDInTenant(tenantID, roleID)
		if err != nil {
			return fmt.Errorf("role %d not found", roleID)
		}
	}

	// 失效权限缓存
	s.invalidateUserPermsCache(context.Background(), tenantID, userID)

	return s.userRepo.AssignRolesInTenant(tenantID, userID, roleIDs)
}

// LockUser 锁定用户
func (s *UserService) LockUser(ctx context.Context, tenantID uint, operatorID uint, userID uint) error {
	if err := s.scopeSvc.EnsureUserAccess(ctx, tenantID, operatorID, userID); err != nil {
		return err
	}
	user, err := s.userRepo.GetByIDInTenant(tenantID, userID)
	if err != nil {
		return err
	}

	user.IsLocked = true
	s.invalidateUserCache(context.Background(), tenantID, userID)
	return s.userRepo.Update(user)
}

// UnlockUser 解锁用户
func (s *UserService) UnlockUser(ctx context.Context, tenantID uint, operatorID uint, userID uint) error {
	if err := s.scopeSvc.EnsureUserAccess(ctx, tenantID, operatorID, userID); err != nil {
		return err
	}
	user, err := s.userRepo.GetByIDInTenant(tenantID, userID)
	if err != nil {
		return err
	}

	user.IsLocked = false
	s.invalidateUserCache(context.Background(), tenantID, userID)
	return s.userRepo.Update(user)
}

// GetUserPermissions 获取用户的所有权限（含完整权限链，支持多部门）
func (s *UserService) GetUserPermissions(tenantID uint, userID uint) ([]model.Permission, error) {
	user, err := s.userRepo.GetByIDWithPermissionsInTenant(tenantID, userID)
	if err != nil {
		return nil, err
	}

	// 如果是超级管理员，返回所有权限
	if user.IsAdmin {
		return s.permissionRepo.ListAll()
	}

	// 1. 获取默认只读角色
	readOnlyRole, _ := s.roleRepo.GetByNameInTenant(tenantID, "READ_ONLY")

	// 2. 收集所有权限（用 map 去重）
	permMap := make(map[uint]model.Permission)

	addPerms := func(perms []model.Permission) {
		for _, p := range perms {
			permMap[p.ID] = p
		}
	}

	// 3. 添加默认只读角色的权限
	if readOnlyRole != nil {
		addPerms(readOnlyRole.Permissions)
	}

	// 4. 添加用户自身角色的权限（user_roles）
	for _, role := range user.Roles {
		addPerms(role.Permissions)
	}

	// 5. 通过 user_departments 中间表获取所有部门的角色权限
	userDepts, err := s.userDeptRepo.GetUserDepartments(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user departments: %w", err)
	}

	// 收集需要查询的角色 ID（来自 UserDepartment.RoleID 指定的部门专属角色）
	var extraRoleIDs []uint
	for _, ud := range userDepts {
		if ud.RoleID != nil && *ud.RoleID != 0 {
			extraRoleIDs = append(extraRoleIDs, *ud.RoleID)
		}
	}

	// 收集所有部门 ID，批量获取部门绑定的角色
	deptIDs := make([]uint, 0, len(userDepts))
	for _, ud := range userDepts {
		deptIDs = append(deptIDs, ud.DeptID)
	}

	// 6. 批量获取每个部门绑定的角色权限（department_roles）
	if len(deptIDs) > 0 {
		depts, err := s.deptRepo.GetByIDsInTenant(tenantID, deptIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get departments: %w", err)
		}
		for _, dept := range depts {
			for _, role := range dept.Roles {
				addPerms(role.Permissions)
			}
		}
	}

	// 7. 批量获取 UserDepartment.RoleID 指定的部门专属角色权限
	if len(extraRoleIDs) > 0 {
		roles, err := s.roleRepo.GetByIDsInTenant(tenantID, extraRoleIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get extra roles: %w", err)
		}
		for _, role := range roles {
			addPerms(role.Permissions)
		}
	}

	// 8. 转换为列表
	permissions := make([]model.Permission, 0, len(permMap))
	for _, p := range permMap {
		permissions = append(permissions, p)
	}

	return permissions, nil
}

// GetUserPermissionCodes 获取用户权限编码集合 (Cached)
func (s *UserService) GetUserPermissionCodes(ctx context.Context, tenantID uint, userID uint) ([]string, error) {
	cacheKey := fmt.Sprintf("tenant:%d:user:perms:%d", tenantID, userID)

	// Try Cache
	cached, err := redis.SMembers(ctx, cacheKey)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	// DB Query
	perms, err := s.GetUserPermissions(tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Check if admin (GetUserPermissions returns all perms for admin, but we might want a shortcut)
	// Actually GetUserPermissions logic already handles admin by returning ALL perms.
	// But for efficiency, if admin, we might want to return a wildcard or handle it in CheckPermission.
	// However, existing logic returns ListAll().

	codes := make([]string, len(perms))
	for i, p := range perms {
		codes[i] = fmt.Sprintf("%s:%s", p.Resource, p.Action)
	}

	// Set Cache
	if len(codes) > 0 {
		args := make([]interface{}, len(codes))
		for i, c := range codes {
			args[i] = c
		}
		redis.SAdd(ctx, cacheKey, args...)
		redis.Expire(ctx, cacheKey, time.Hour)
	} else {
		// Cache empty set? Redis set cannot be empty.
		// Maybe store a placeholder or just don't cache (but then cache miss every time).
		// For simplicity, skip caching empty.
	}

	return codes, nil
}

// CheckPermission 检查用户是否有指定权限 (Cached)
func (s *UserService) CheckPermission(ctx context.Context, tenantID uint, userID uint, resource, action string) (bool, error) {
	// 0. Check Admin shortcut (if we want to avoid loading all perms)
	// But we need to load user to know if admin.
	// We can cache "is_admin" too.
	// For now, let's rely on GetUserPermissionCodes which caches perms.
	// If admin, it returns ALL perms.

	perms, err := s.GetUserPermissionCodes(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	target := fmt.Sprintf("%s:%s", resource, action)
	for _, p := range perms {
		if p == target {
			return true, nil
		}
		// Handle wildcards if needed, but current perm system seems exact match based on existing code.
		// Existing code: if perm.Resource == resource && perm.Action == action
		// So exact match.
	}

	return false, nil
}

func (s *UserService) invalidateUserCache(ctx context.Context, tenantID uint, userID uint) {
	redis.Del(ctx, fmt.Sprintf("tenant:%d:user:info:%d", tenantID, userID))
}

func (s *UserService) invalidateUserPermsCache(ctx context.Context, tenantID uint, userID uint) {
	redis.Del(ctx, fmt.Sprintf("tenant:%d:user:perms:%d", tenantID, userID))
}

// UserAllPermissions 用户全部权限（菜单/按钮/字段/API 四合一）
type UserAllPermissions struct {
	Menus      []MenuPermission             `json:"menus"`
	Buttons    map[string][]string          `json:"buttons"`
	FieldRules map[string]map[string]string `json:"fieldRules"`
	APIs       []string                     `json:"apis"`
}

// MenuPermission 菜单权限
type MenuPermission struct {
	Name     string           `json:"name"`
	Path     string           `json:"path"`
	Icon     string           `json:"icon"`
	Children []MenuPermission `json:"children,omitempty"`
}

// GetUserAllPermissions 获取用户全部权限（含菜单/按钮/字段/API）
func (s *UserService) GetUserAllPermissions(ctx context.Context, tenantID, userID uint) (*UserAllPermissions, error) {
	// 1. 获取所有权限
	perms, err := s.GetUserPermissions(tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	result := &UserAllPermissions{
		Buttons:    make(map[string][]string),
		FieldRules: make(map[string]map[string]string),
	}

	// 2. 按 Type 分类
	var menuPerms []model.Permission
	for _, p := range perms {
		switch p.Type {
		case model.PermissionTypeMenu:
			menuPerms = append(menuPerms, p)
		case model.PermissionTypeButton:
			result.Buttons[p.Resource] = append(result.Buttons[p.Resource], p.Action)
		case model.PermissionTypeAPI:
			result.APIs = append(result.APIs, fmt.Sprintf("%s:%s", p.Resource, p.Action))
		}
	}

	// 3. 构建菜单树
	result.Menus = buildMenuTree(menuPerms, nil)

	// 4. 获取用户角色ID列表，查询字段权限
	user, err := s.userRepo.GetByIDWithPermissionsInTenant(tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	var roleIDs []uint
	for _, role := range user.Roles {
		roleIDs = append(roleIDs, role.ID)
	}

	if len(roleIDs) > 0 {
		fieldPerms, err := s.fieldPermRepo.GetByRoleIDs(roleIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get field permissions: %w", err)
		}
		for _, fp := range fieldPerms {
			if _, ok := result.FieldRules[fp.Resource]; !ok {
				result.FieldRules[fp.Resource] = make(map[string]string)
			}
			result.FieldRules[fp.Resource][fp.FieldName] = string(fp.Action)
		}
	}

	return result, nil
}

// buildMenuTree 递归构建菜单树
func buildMenuTree(perms []model.Permission, parentID *uint) []MenuPermission {
	var trees []MenuPermission
	for _, p := range perms {
		if p.ParentID == nil && parentID == nil || p.ParentID != nil && parentID != nil && *p.ParentID == *parentID {
			node := MenuPermission{
				Name: p.Name,
				Path: p.Path,
				Icon: p.Icon,
			}
			childParentID := p.ID
			node.Children = buildMenuTree(perms, &childParentID)
			trees = append(trees, node)
		}
	}
	return trees
}
