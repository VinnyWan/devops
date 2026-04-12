package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/pkg/redis"

	"gorm.io/gorm"
)

type RoleService struct {
	db             *gorm.DB
	roleRepo       *repository.RoleRepo
	permissionRepo *repository.PermissionRepo
	userRepo       *repository.UserRepo
	scopeSvc       *AccessScopeService
}

func NewRoleService(db *gorm.DB) *RoleService {
	userRepo := repository.NewUserRepo(db)
	deptRepo := repository.NewDepartmentRepo(db)
	return &RoleService{
		db:             db,
		roleRepo:       repository.NewRoleRepo(db),
		permissionRepo: repository.NewPermissionRepo(db),
		userRepo:       userRepo,
		scopeSvc:       NewAccessScopeService(userRepo, deptRepo),
	}
}

func (s *RoleService) CheckAdmin(tenantID uint, userID uint) error {
	user, err := s.userRepo.GetByIDInTenant(tenantID, userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New("permission denied: only admin can perform this action")
	}
	return nil
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	DataScope   string `json:"dataScope"`
}

func (s *RoleService) CreateRole(tenantID uint, req *CreateRoleRequest) (*model.Role, error) {
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}
	dataScope, err := normalizeRoleDataScope(req.DataScope)
	if err != nil {
		return nil, err
	}

	role := &model.Role{
		TenantID:    &tenantID,
		Name:        req.Name,
		DisplayName: displayName,
		Description: req.Description,
		Type:        "custom",
		DataScope:   dataScope,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

type UpdateRoleRequest struct {
	ID            uint   `json:"id" binding:"required"`
	Name          string `json:"name"`
	DisplayName   string `json:"displayName"`
	Description   string `json:"description"`
	DataScope     *string `json:"dataScope"`
	PermissionIDs []uint `json:"permissionIds"`
}

func (s *RoleService) UpdateRole(tenantID uint, req *UpdateRoleRequest) error {
	role, err := s.roleRepo.GetByIDInTenant(tenantID, req.ID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if req.Name != "" {
		if role.Type == "system" && req.Name != role.Name {
			return errors.New("system role name cannot be modified")
		}
		role.Name = req.Name
	}
	if req.DisplayName != "" {
		role.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.DataScope != nil {
		if role.Type == "system" {
			return errors.New("system role data scope cannot be modified")
		}
		dataScope, err := normalizeRoleDataScope(*req.DataScope)
		if err != nil {
			return err
		}
		role.DataScope = dataScope
	}

	if err := s.roleRepo.Update(role); err != nil {
		return err
	}

	if req.PermissionIDs != nil {
		if role.Type == "system" {
			return errors.New("system role permissions cannot be modified")
		}
		affectedUserIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleIDInTenant(tenantID, req.ID)
		if err != nil {
			return err
		}
		if err := s.roleRepo.AssignPermissionsInTenant(tenantID, req.ID, req.PermissionIDs); err != nil {
			return err
		}
		s.invalidateUserPermsByUserIDs(context.Background(), tenantID, affectedUserIDs)
		return nil
	}

	return nil
}

func (s *RoleService) DeleteRole(tenantID uint, id uint) error {
	role, err := s.roleRepo.GetByIDInTenant(tenantID, id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if role.Type == "system" {
		return errors.New("system role cannot be deleted")
	}

	affectedUserIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleIDInTenant(tenantID, id)
	if err != nil {
		return err
	}
	if err := s.roleRepo.DeleteInTenant(tenantID, id); err != nil {
		return err
	}
	s.invalidateUserPermsByUserIDs(context.Background(), tenantID, affectedUserIDs)
	return nil
}

func (s *RoleService) ListRoles(tenantID uint, page, pageSize int, keyword string) ([]model.Role, int64, error) {
	return s.roleRepo.ListInTenant(tenantID, page, pageSize, keyword)
}

func (s *RoleService) GetRoleByID(tenantID uint, id uint) (*model.Role, error) {
	return s.roleRepo.GetByIDInTenant(tenantID, id)
}

func (s *RoleService) AssignPermissions(tenantID uint, roleID uint, permissionIDs []uint) error {
	role, err := s.roleRepo.GetByIDInTenant(tenantID, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	if role.Type == "system" {
		return errors.New("system role permissions cannot be modified")
	}

	affectedUserIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleIDInTenant(tenantID, roleID)
	if err != nil {
		return err
	}
	if err := s.roleRepo.AssignPermissionsInTenant(tenantID, roleID, permissionIDs); err != nil {
		return err
	}
	s.invalidateUserPermsByUserIDs(context.Background(), tenantID, affectedUserIDs)
	return nil
}

func (s *RoleService) GetRoleUsers(ctx context.Context, tenantID uint, operatorID uint, roleID uint) ([]model.User, error) {
	users, err := s.roleRepo.GetRoleUsersInTenant(tenantID, roleID)
	if err != nil {
		return nil, err
	}
	scope, err := s.scopeSvc.Resolve(ctx, tenantID, operatorID)
	if err != nil {
		return nil, err
	}
	if scope.AllowsAll() {
		return users, nil
	}

	filtered := make([]model.User, 0, len(users))
	for _, user := range users {
		if user.ID == operatorID {
			filtered = append(filtered, user)
			continue
		}
		if user.PrimaryDeptID != nil && scope.AllowsDepartmentID(*user.PrimaryDeptID) {
			filtered = append(filtered, user)
		}
	}
	return filtered, nil
}

func (s *RoleService) GetRoleDepartments(ctx context.Context, tenantID uint, operatorID uint, roleID uint) ([]model.Department, error) {
	departments, err := s.roleRepo.GetRoleDepartmentsInTenant(tenantID, roleID)
	if err != nil {
		return nil, err
	}
	scope, err := s.scopeSvc.Resolve(ctx, tenantID, operatorID)
	if err != nil {
		return nil, err
	}
	if scope.AllowsAll() {
		return departments, nil
	}

	filtered := make([]model.Department, 0, len(departments))
	for _, department := range departments {
		if scope.AllowsDepartmentID(department.ID) {
			filtered = append(filtered, department)
		}
	}
	return filtered, nil
}

func (s *RoleService) AssignUsers(ctx context.Context, tenantID uint, operatorID uint, roleID uint, userIDs []uint) error {
	_, err := s.roleRepo.GetByIDInTenant(tenantID, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	if err := s.scopeSvc.EnsureUsersAccess(ctx, tenantID, operatorID, userIDs); err != nil {
		return err
	}
	beforeIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleIDInTenant(tenantID, roleID)
	if err != nil {
		return err
	}
	if err := s.roleRepo.AssignUsersInTenant(tenantID, roleID, userIDs); err != nil {
		return err
	}
	afterIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleIDInTenant(tenantID, roleID)
	if err != nil {
		return err
	}
	s.invalidateUserPermsByUserIDs(context.Background(), tenantID, mergeUserIDs(beforeIDs, afterIDs))
	return nil
}

func (s *RoleService) AssignDepartments(ctx context.Context, tenantID uint, operatorID uint, roleID uint, departmentIDs []uint) error {
	_, err := s.roleRepo.GetByIDInTenant(tenantID, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	if err := s.scopeSvc.EnsureDepartmentsAccess(ctx, tenantID, operatorID, departmentIDs); err != nil {
		return err
	}
	beforeIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleIDInTenant(tenantID, roleID)
	if err != nil {
		return err
	}
	if err := s.roleRepo.AssignDepartmentsInTenant(tenantID, roleID, departmentIDs); err != nil {
		return err
	}
	afterIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleIDInTenant(tenantID, roleID)
	if err != nil {
		return err
	}
	s.invalidateUserPermsByUserIDs(context.Background(), tenantID, mergeUserIDs(beforeIDs, afterIDs))
	return nil
}

// -------------------------------------------------------------------
// Permission Management
// -------------------------------------------------------------------

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

// CreatePermission 创建权限
func (s *RoleService) CreatePermission(req *CreatePermissionRequest) (*model.Permission, error) {
	perm := &model.Permission{
		Name:        req.Name,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	}

	if err := s.permissionRepo.Create(perm); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return perm, nil
}

func (s *RoleService) ListPermissions(page, pageSize int, resource, keyword string) ([]model.Permission, int64, error) {
	return s.permissionRepo.List(page, pageSize, resource, keyword)
}

func (s *RoleService) ListAllPermissions() ([]model.Permission, error) {
	return s.permissionRepo.ListAll()
}

func (s *RoleService) GetPermissionByID(id uint) (*model.Permission, error) {
	return s.permissionRepo.GetByID(id)
}

func (s *RoleService) UpdatePermission(permission *model.Permission) error {
	_, err := s.permissionRepo.GetByID(permission.ID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	return s.permissionRepo.Update(permission)
}

func (s *RoleService) DeletePermission(id uint) error {
	// 检查权限是否存在
	_, err := s.permissionRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	return s.permissionRepo.Delete(id)
}

func (s *RoleService) invalidateUserPermsByUserIDs(ctx context.Context, tenantID uint, userIDs []uint) {
	for _, userID := range userIDs {
		redis.Del(ctx, fmt.Sprintf("tenant:%d:user:perms:%d", tenantID, userID))
	}
}

func mergeUserIDs(a []uint, b []uint) []uint {
	idSet := make(map[uint]struct{}, len(a)+len(b))
	for _, id := range a {
		idSet[id] = struct{}{}
	}
	for _, id := range b {
		idSet[id] = struct{}{}
	}
	merged := make([]uint, 0, len(idSet))
	for id := range idSet {
		merged = append(merged, id)
	}
	return merged
}

func normalizeRoleDataScope(scope string) (model.DataScope, error) {
	trimmed := strings.TrimSpace(scope)
	if trimmed == "" {
		return model.DataScopeSelfDepartment, nil
	}
	if !model.IsValidDataScope(trimmed) {
		return "", errors.New("invalid dataScope")
	}
	return model.NormalizeDataScope(trimmed), nil
}
