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
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		userRepo:       repository.NewUserRepo(db),
		roleRepo:       repository.NewRoleRepo(db),
		permissionRepo: repository.NewPermissionRepo(db),
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
func (s *UserService) Register(req *RegisterRequest) (*model.User, error) {
	if err := utils.ValidatePasswordComplexity(req.Password); err != nil {
		return nil, err
	}

	existingUser, err := s.userRepo.GetByUsername(req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	existingUser, err = s.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &model.User{
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
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	// TODO: 实现用户信息缓存
	return s.userRepo.GetByID(id)
}

// GetUserByUsername 根据用户名获取用户信息
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	return s.userRepo.GetByUsername(username)
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int, keyword string) ([]model.User, int64, error) {
	return s.userRepo.List(page, pageSize, keyword)
}

func (s *UserService) UpdateUserByRequest(req *UpdateUserRequest) error {
	_, err := s.userRepo.GetByID(req.ID)
	if err != nil {
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

	if req.DepartmentID != nil {
		if *req.DepartmentID == 0 {
			updates["department_id"] = nil
		} else {
			updates["department_id"] = *req.DepartmentID
		}
	}

	if len(updates) == 0 {
		return nil
	}

	// 使得相关缓存失效
	s.invalidateUserCache(context.Background(), req.ID)
	s.invalidateUserPermsCache(context.Background(), req.ID)

	return s.userRepo.UpdateByID(req.ID, updates)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	s.invalidateUserCache(context.Background(), id)
	s.invalidateUserPermsCache(context.Background(), id)
	return s.userRepo.Delete(id)
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	if err := utils.ValidatePasswordComplexity(req.NewPassword); err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(userID)
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

	return s.userRepo.UpdatePassword(userID, hashedPassword)
}

// ResetPassword 重置密码（管理员操作）
func (s *UserService) ResetPassword(userID uint, newPassword string) error {
	if err := utils.ValidatePasswordComplexity(newPassword); err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(userID)
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

	return s.userRepo.UpdatePassword(userID, hashedPassword)
}

// AssignRoles 分配角色
func (s *UserService) AssignRoles(userID uint, roleIDs []uint) error {
	// 验证所有角色是否存在
	for _, roleID := range roleIDs {
		_, err := s.roleRepo.GetByID(roleID)
		if err != nil {
			return fmt.Errorf("role %d not found", roleID)
		}
	}

	// 失效权限缓存
	s.invalidateUserPermsCache(context.Background(), userID)

	return s.userRepo.AssignRoles(userID, roleIDs)
}

// LockUser 锁定用户
func (s *UserService) LockUser(userID uint) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.IsLocked = true
	s.invalidateUserCache(context.Background(), userID)
	return s.userRepo.Update(user)
}

// UnlockUser 解锁用户
func (s *UserService) UnlockUser(userID uint) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.IsLocked = false
	s.invalidateUserCache(context.Background(), userID)
	return s.userRepo.Update(user)
}

// GetUserPermissions 获取用户的所有权限（含完整权限链）
func (s *UserService) GetUserPermissions(userID uint) ([]model.Permission, error) {
	user, err := s.userRepo.GetByIDWithPermissions(userID)
	if err != nil {
		return nil, err
	}

	// 如果是超级管理员，返回所有权限
	if user.IsAdmin {
		return s.permissionRepo.ListAll()
	}

	// 1. 获取默认只读角色
	readOnlyRole, _ := s.roleRepo.GetByName("READ_ONLY")

	// 2. 收集所有权限
	permMap := make(map[uint]model.Permission)

	// Helper to add permissions
	addPerms := func(perms []model.Permission) {
		for _, p := range perms {
			permMap[p.ID] = p
		}
	}

	// 3. 添加默认只读角色的权限
	if readOnlyRole != nil {
		addPerms(readOnlyRole.Permissions)
	}

	// 4. 添加用户自身角色的权限
	for _, role := range user.Roles {
		addPerms(role.Permissions)
	}

	// 5. 添加部门角色的权限
	if user.Department != nil {
		for _, role := range user.Department.Roles {
			addPerms(role.Permissions)
		}
	}

	// 6. 转换为列表
	permissions := make([]model.Permission, 0, len(permMap))
	for _, p := range permMap {
		permissions = append(permissions, p)
	}

	return permissions, nil
}

// GetUserPermissionCodes 获取用户权限编码集合 (Cached)
func (s *UserService) GetUserPermissionCodes(ctx context.Context, userID uint) ([]string, error) {
	cacheKey := fmt.Sprintf("user:perms:%d", userID)

	// Try Cache
	cached, err := redis.SMembers(ctx, cacheKey)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	// DB Query
	perms, err := s.GetUserPermissions(userID)
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
func (s *UserService) CheckPermission(ctx context.Context, userID uint, resource, action string) (bool, error) {
	// 0. Check Admin shortcut (if we want to avoid loading all perms)
	// But we need to load user to know if admin.
	// We can cache "is_admin" too.
	// For now, let's rely on GetUserPermissionCodes which caches perms.
	// If admin, it returns ALL perms.

	perms, err := s.GetUserPermissionCodes(ctx, userID)
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

func (s *UserService) invalidateUserCache(ctx context.Context, userID uint) {
	redis.Del(ctx, fmt.Sprintf("user:info:%d", userID))
}

func (s *UserService) invalidateUserPermsCache(ctx context.Context, userID uint) {
	redis.Del(ctx, fmt.Sprintf("user:perms:%d", userID))
}
