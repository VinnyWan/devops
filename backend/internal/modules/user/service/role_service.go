package service

import (
	"context"
	"errors"
	"fmt"

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
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{
		db:             db,
		roleRepo:       repository.NewRoleRepo(db),
		permissionRepo: repository.NewPermissionRepo(db),
		userRepo:       repository.NewUserRepo(db),
	}
}

func (s *RoleService) CheckAdmin(userID uint) error {
	user, err := s.userRepo.GetByID(userID)
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
}

func (s *RoleService) CreateRole(req *CreateRoleRequest) (*model.Role, error) {
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}

	role := &model.Role{
		Name:        req.Name,
		DisplayName: displayName,
		Description: req.Description,
		Type:        "custom",
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
	PermissionIDs []uint `json:"permissionIds"`
}

func (s *RoleService) UpdateRole(req *UpdateRoleRequest) error {
	role, err := s.roleRepo.GetByID(req.ID)
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

	if err := s.roleRepo.Update(role); err != nil {
		return err
	}

	if req.PermissionIDs != nil {
		affectedUserIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleID(req.ID)
		if err != nil {
			return err
		}
		if err := s.roleRepo.AssignPermissions(req.ID, req.PermissionIDs); err != nil {
			return err
		}
		s.invalidateUserPermsByUserIDs(context.Background(), affectedUserIDs)
		return nil
	}

	return nil
}

func (s *RoleService) DeleteRole(id uint) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if role.Type == "system" {
		return errors.New("system role cannot be deleted")
	}

	affectedUserIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleID(id)
	if err != nil {
		return err
	}
	if err := s.roleRepo.Delete(id); err != nil {
		return err
	}
	s.invalidateUserPermsByUserIDs(context.Background(), affectedUserIDs)
	return nil
}

func (s *RoleService) ListRoles(page, pageSize int, keyword string) ([]model.Role, int64, error) {
	return s.roleRepo.List(page, pageSize, keyword)
}

func (s *RoleService) GetRoleByID(id uint) (*model.Role, error) {
	return s.roleRepo.GetByID(id)
}

func (s *RoleService) AssignPermissions(roleID uint, permissionIDs []uint) error {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	affectedUserIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleID(roleID)
	if err != nil {
		return err
	}
	if err := s.roleRepo.AssignPermissions(roleID, permissionIDs); err != nil {
		return err
	}
	s.invalidateUserPermsByUserIDs(context.Background(), affectedUserIDs)
	return nil
}

func (s *RoleService) GetRoleUsers(roleID uint) ([]model.User, error) {
	return s.roleRepo.GetRoleUsers(roleID)
}

func (s *RoleService) GetRoleDepartments(roleID uint) ([]model.Department, error) {
	return s.roleRepo.GetRoleDepartments(roleID)
}

func (s *RoleService) AssignUsers(roleID uint, userIDs []uint) error {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	beforeIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleID(roleID)
	if err != nil {
		return err
	}
	if err := s.roleRepo.AssignUsers(roleID, userIDs); err != nil {
		return err
	}
	afterIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleID(roleID)
	if err != nil {
		return err
	}
	s.invalidateUserPermsByUserIDs(context.Background(), mergeUserIDs(beforeIDs, afterIDs))
	return nil
}

func (s *RoleService) AssignDepartments(roleID uint, departmentIDs []uint) error {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}
	beforeIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleID(roleID)
	if err != nil {
		return err
	}
	if err := s.roleRepo.AssignDepartments(roleID, departmentIDs); err != nil {
		return err
	}
	afterIDs, err := s.userRepo.ListPermissionAffectedUserIDsByRoleID(roleID)
	if err != nil {
		return err
	}
	s.invalidateUserPermsByUserIDs(context.Background(), mergeUserIDs(beforeIDs, afterIDs))
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

func (s *RoleService) invalidateUserPermsByUserIDs(ctx context.Context, userIDs []uint) {
	for _, userID := range userIDs {
		redis.Del(ctx, fmt.Sprintf("user:perms:%d", userID))
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
