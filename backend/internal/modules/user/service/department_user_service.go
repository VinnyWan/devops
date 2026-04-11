package service

import (
	"context"
	"errors"
	"fmt"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/pkg/redis"
	"devops-platform/internal/pkg/utils"
)

type DepartmentUserService struct {
	userRepo *repository.UserRepo
	deptRepo *repository.DepartmentRepo
	scopeSvc *AccessScopeService
}

func NewDepartmentUserService(userRepo *repository.UserRepo, deptRepo *repository.DepartmentRepo) *DepartmentUserService {
	return &DepartmentUserService{
		userRepo: userRepo,
		deptRepo: deptRepo,
		scopeSvc: NewAccessScopeService(userRepo, deptRepo),
	}
}

type CreateDeptUserRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required,min=8"`
	Email        string `json:"email" binding:"required,email"`
	Name         string `json:"name"`
	DepartmentID uint   `json:"departmentId"`
	Status       string `json:"status"`
}

type UpdateDeptUserRequest struct {
	ID     uint    `json:"id" binding:"required"`
	Email  *string `json:"email"`
	Name   *string `json:"name"`
	Status *string `json:"status"`
}

type TransferUserDepartmentRequest struct {
	UserID         uint `json:"userId" binding:"required"`
	ToDepartmentID uint `json:"toDepartmentId" binding:"required"`
}

func (s *DepartmentUserService) List(tenantID uint, operatorID uint, deptID uint, page, pageSize int, keyword string) ([]model.User, int64, error) {
	scope, err := s.scopeSvc.Resolve(context.Background(), tenantID, operatorID)
	if err != nil {
		return nil, 0, err
	}
	if deptID != 0 {
		if !scope.AllowsAll() && !scope.AllowsDepartmentID(deptID) {
			return nil, 0, ErrScopeForbidden
		}
		if _, err := s.deptRepo.GetByIDInTenant(tenantID, deptID); err != nil {
			return nil, 0, err
		}
		targetDepartmentIDs, err := s.resolveListDepartmentIDs(tenantID, deptID, scope)
		if err != nil {
			return nil, 0, err
		}
		return s.userRepo.ListByDepartmentIDsInTenant(tenantID, targetDepartmentIDs, page, pageSize, keyword)
	}
	if scope.AllowsAll() {
		return s.userRepo.ListInTenant(tenantID, page, pageSize, keyword)
	}
	return s.userRepo.ListByDepartmentIDsInTenant(tenantID, scope.DepartmentIDs, page, pageSize, keyword)
}

func (s *DepartmentUserService) resolveListDepartmentIDs(tenantID uint, deptID uint, scope *DataAccessScope) ([]uint, error) {
	departments, err := s.deptRepo.ListHierarchyInTenant(tenantID)
	if err != nil {
		return nil, err
	}

	targetDepartmentIDs := collectDepartmentTreeIDs(deptID, departments)
	if scope == nil || scope.AllowsAll() {
		return targetDepartmentIDs, nil
	}

	allowed := make(map[uint]struct{}, len(scope.DepartmentIDs))
	for _, departmentID := range scope.DepartmentIDs {
		allowed[departmentID] = struct{}{}
	}

	filtered := make([]uint, 0, len(targetDepartmentIDs))
	for _, departmentID := range targetDepartmentIDs {
		if _, ok := allowed[departmentID]; ok {
			filtered = append(filtered, departmentID)
		}
	}
	return filtered, nil
}

func (s *DepartmentUserService) Create(tenantID uint, operatorID uint, req *CreateDeptUserRequest) (*model.User, error) {
	targetDeptID := req.DepartmentID
	if targetDeptID == 0 {
		operator, err := s.userRepo.GetByIDInTenant(tenantID, operatorID)
		if err != nil {
			return nil, err
		}
		if operator.DepartmentID == nil {
			return nil, errors.New("departmentId is required")
		}
		targetDeptID = *operator.DepartmentID
	}

	if err := s.scopeSvc.EnsureDepartmentAccess(context.Background(), tenantID, operatorID, targetDeptID); err != nil {
		return nil, err
	}

	if err := utils.ValidatePasswordComplexity(req.Password); err != nil {
		return nil, err
	}

	if existing, err := s.userRepo.GetByUsernameInTenant(tenantID, req.Username); err == nil && existing != nil {
		return nil, errors.New("username already exists")
	}
	if existing, err := s.userRepo.GetByEmailInTenant(tenantID, req.Email); err == nil && existing != nil {
		return nil, errors.New("email already exists")
	}

	if _, err := s.deptRepo.GetByIDInTenant(tenantID, targetDeptID); err != nil {
		return nil, fmt.Errorf("department not found: %w", err)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	deptID := targetDeptID
	user := &model.User{
		TenantID:     &tenantID,
		Username:     req.Username,
		Password:     hashedPassword,
		Email:        req.Email,
		Name:         req.Name,
		AuthType:     model.AuthTypeLocal,
		Status:       status,
		IsAdmin:      false,
		IsLocked:     false,
		DepartmentID: &deptID,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *DepartmentUserService) Update(tenantID uint, operatorID uint, req *UpdateDeptUserRequest) error {
	existing, err := s.userRepo.GetByIDInTenant(tenantID, req.ID)
	if err != nil {
		return err
	}

	if existing.DepartmentID == nil {
		return errors.New("user has no department")
	}

	if err := s.scopeSvc.EnsureUserAccess(context.Background(), tenantID, operatorID, existing.ID); err != nil {
		return err
	}

	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}

	if err := s.userRepo.Update(existing); err != nil {
		return err
	}
	redis.Del(context.Background(), fmt.Sprintf("tenant:%d:user:perms:%d", tenantID, existing.ID))
	return nil
}

func (s *DepartmentUserService) Delete(tenantID uint, operatorID uint, userID uint) error {
	existing, err := s.userRepo.GetByIDInTenant(tenantID, userID)
	if err != nil {
		return err
	}
	if existing.DepartmentID == nil {
		return errors.New("user has no department")
	}
	if err := s.scopeSvc.EnsureUserAccess(context.Background(), tenantID, operatorID, existing.ID); err != nil {
		return err
	}
	if err := s.userRepo.DeleteInTenant(tenantID, userID); err != nil {
		return err
	}
	redis.Del(context.Background(), fmt.Sprintf("tenant:%d:user:perms:%d", tenantID, userID))
	return nil
}

func (s *DepartmentUserService) Transfer(tenantID uint, operatorID uint, req *TransferUserDepartmentRequest) error {
	if req.ToDepartmentID == 0 {
		return errors.New("toDepartmentId is required")
	}
	user, err := s.userRepo.GetByIDInTenant(tenantID, req.UserID)
	if err != nil {
		return err
	}
	if user.DepartmentID == nil {
		return errors.New("user has no department")
	}
	if *user.DepartmentID == req.ToDepartmentID {
		return errors.New("target department must be different")
	}

	if err := s.scopeSvc.EnsureUserAccess(context.Background(), tenantID, operatorID, user.ID); err != nil {
		return err
	}
	if err := s.scopeSvc.EnsureDepartmentAccess(context.Background(), tenantID, operatorID, req.ToDepartmentID); err != nil {
		return err
	}
	if _, err := s.deptRepo.GetByIDInTenant(tenantID, req.ToDepartmentID); err != nil {
		return fmt.Errorf("target department not found: %w", err)
	}

	if err := s.userRepo.AssignRolesInTenant(tenantID, user.ID, []uint{}); err != nil {
		return err
	}

	toDeptID := req.ToDepartmentID
	user.DepartmentID = &toDeptID
	user.Department = nil
	user.Roles = nil

	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	redis.Del(context.Background(), fmt.Sprintf("tenant:%d:user:perms:%d", tenantID, user.ID))
	return nil
}
