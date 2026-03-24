package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/pkg/redis"
	"devops-platform/internal/pkg/utils"
)

type DepartmentUserService struct {
	userRepo *repository.UserRepo
	deptRepo *repository.DepartmentRepo
}

func NewDepartmentUserService(userRepo *repository.UserRepo, deptRepo *repository.DepartmentRepo) *DepartmentUserService {
	return &DepartmentUserService{userRepo: userRepo, deptRepo: deptRepo}
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

func (s *DepartmentUserService) List(operatorID uint, deptID uint, page, pageSize int, keyword string) ([]model.User, int64, error) {
	targetDeptID, err := s.resolveTargetDeptID(operatorID, deptID)
	if err != nil {
		return nil, 0, err
	}
	return s.userRepo.ListByDepartment(targetDeptID, page, pageSize, keyword)
}

func (s *DepartmentUserService) Create(operatorID uint, req *CreateDeptUserRequest) (*model.User, error) {
	targetDeptID := req.DepartmentID
	if targetDeptID == 0 {
		operator, err := s.userRepo.GetByID(operatorID)
		if err != nil {
			return nil, err
		}
		if operator.DepartmentID == nil {
			return nil, errors.New("departmentId is required")
		}
		targetDeptID = *operator.DepartmentID
	}

	if err := s.checkDeptManager(operatorID, targetDeptID); err != nil {
		return nil, err
	}

	if err := utils.ValidatePasswordComplexity(req.Password); err != nil {
		return nil, err
	}

	if existing, err := s.userRepo.GetByUsername(req.Username); err == nil && existing != nil {
		return nil, errors.New("username already exists")
	}
	if existing, err := s.userRepo.GetByEmail(req.Email); err == nil && existing != nil {
		return nil, errors.New("email already exists")
	}

	if _, err := s.deptRepo.GetByID(targetDeptID); err != nil {
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

func (s *DepartmentUserService) Update(operatorID uint, req *UpdateDeptUserRequest) error {
	existing, err := s.userRepo.GetByID(req.ID)
	if err != nil {
		return err
	}

	if existing.DepartmentID == nil {
		return errors.New("user has no department")
	}

	if err := s.checkDeptManager(operatorID, *existing.DepartmentID); err != nil {
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
	redis.Del(context.Background(), fmt.Sprintf("user:perms:%d", existing.ID))
	return nil
}

func (s *DepartmentUserService) Delete(operatorID uint, userID uint) error {
	existing, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if existing.DepartmentID == nil {
		return errors.New("user has no department")
	}
	if err := s.checkDeptManager(operatorID, *existing.DepartmentID); err != nil {
		return err
	}
	if err := s.userRepo.Delete(userID); err != nil {
		return err
	}
	redis.Del(context.Background(), fmt.Sprintf("user:perms:%d", userID))
	return nil
}

func (s *DepartmentUserService) Transfer(operatorID uint, req *TransferUserDepartmentRequest) error {
	if req.ToDepartmentID == 0 {
		return errors.New("toDepartmentId is required")
	}
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		return err
	}
	if user.DepartmentID == nil {
		return errors.New("user has no department")
	}
	if *user.DepartmentID == req.ToDepartmentID {
		return errors.New("target department must be different")
	}

	if err := s.checkDeptManager(operatorID, *user.DepartmentID); err != nil {
		return err
	}
	if err := s.checkDeptManager(operatorID, req.ToDepartmentID); err != nil {
		return err
	}
	if _, err := s.deptRepo.GetByID(req.ToDepartmentID); err != nil {
		return fmt.Errorf("target department not found: %w", err)
	}

	if err := s.userRepo.AssignRoles(user.ID, []uint{}); err != nil {
		return err
	}

	toDeptID := req.ToDepartmentID
	user.DepartmentID = &toDeptID
	user.Department = nil
	user.Roles = nil

	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	redis.Del(context.Background(), fmt.Sprintf("user:perms:%d", user.ID))
	return nil
}

func (s *DepartmentUserService) resolveTargetDeptID(operatorID uint, deptID uint) (uint, error) {
	operator, err := s.userRepo.GetByID(operatorID)
	if err != nil {
		return 0, err
	}

	if operator.IsAdmin {
		if deptID == 0 {
			if operator.DepartmentID == nil {
				return 0, errors.New("departmentId is required")
			}
			return *operator.DepartmentID, nil
		}
		return deptID, nil
	}

	if operator.DepartmentID == nil {
		return 0, errors.New("permission denied")
	}

	if deptID != 0 && deptID != *operator.DepartmentID {
		return 0, errors.New("permission denied")
	}

	return *operator.DepartmentID, nil
}

func (s *DepartmentUserService) checkDeptManager(operatorID uint, deptID uint) error {
	operator, err := s.userRepo.GetByID(operatorID)
	if err != nil {
		return err
	}

	if operator.IsAdmin {
		return nil
	}

	if operator.DepartmentID == nil || *operator.DepartmentID != deptID {
		return errors.New("permission denied")
	}

	if hasRole(operator.Roles, "DEPT_ADMIN", "dept_admin", "department_admin") {
		return nil
	}

	return errors.New("permission denied")
}

func hasRole(roles []model.Role, names ...string) bool {
	if len(roles) == 0 || len(names) == 0 {
		return false
	}
	nameSet := make(map[string]struct{}, len(names))
	for _, n := range names {
		nameSet[strings.ToLower(n)] = struct{}{}
	}
	for _, r := range roles {
		if _, ok := nameSet[strings.ToLower(r.Name)]; ok {
			return true
		}
	}
	return false
}
