package service

import (
	"context"
	"errors"
	"sort"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
)

var ErrScopeForbidden = errors.New("permission denied")

type DataAccessScope struct {
	TenantID      uint
	Level         model.DataScope
	DepartmentIDs []uint
	departmentSet map[uint]struct{}
}

func newDataAccessScope(tenantID uint, level model.DataScope, departmentIDs []uint) *DataAccessScope {
	idSet := make(map[uint]struct{}, len(departmentIDs))
	normalizedIDs := make([]uint, 0, len(departmentIDs))
	for _, id := range departmentIDs {
		if id == 0 {
			continue
		}
		if _, exists := idSet[id]; exists {
			continue
		}
		idSet[id] = struct{}{}
		normalizedIDs = append(normalizedIDs, id)
	}
	sort.Slice(normalizedIDs, func(i, j int) bool {
		return normalizedIDs[i] < normalizedIDs[j]
	})
	return &DataAccessScope{
		TenantID:      tenantID,
		Level:         model.NormalizeDataScope(string(level)),
		DepartmentIDs: normalizedIDs,
		departmentSet: idSet,
	}
}

func (s *DataAccessScope) AllowsAll() bool {
	return s != nil && s.Level == model.DataScopeTenant
}

func (s *DataAccessScope) AllowsDepartmentID(departmentID uint) bool {
	if s == nil || departmentID == 0 {
		return false
	}
	if s.AllowsAll() {
		return true
	}
	_, ok := s.departmentSet[departmentID]
	return ok
}

type AccessScopeService struct {
	userRepo *repository.UserRepo
	deptRepo *repository.DepartmentRepo
}

func NewAccessScopeService(userRepo *repository.UserRepo, deptRepo *repository.DepartmentRepo) *AccessScopeService {
	return &AccessScopeService{
		userRepo: userRepo,
		deptRepo: deptRepo,
	}
}

func (s *AccessScopeService) Resolve(ctx context.Context, tenantID uint, userID uint) (*DataAccessScope, error) {
	_ = ctx

	user, err := s.userRepo.GetByIDWithPermissionsInTenant(tenantID, userID)
	if err != nil {
		return nil, err
	}

	if user.IsAdmin {
		return newDataAccessScope(tenantID, model.DataScopeTenant, nil), nil
	}

	scopeLevel := model.DataScopeSelfDepartment
	for _, role := range user.Roles {
		scopeLevel = model.MaxDataScope(scopeLevel, model.NormalizeDataScope(string(role.DataScope)))
	}
	if user.Department != nil {
		for _, role := range user.Department.Roles {
			scopeLevel = model.MaxDataScope(scopeLevel, model.NormalizeDataScope(string(role.DataScope)))
		}
	}

	if scopeLevel == model.DataScopeTenant {
		return newDataAccessScope(tenantID, scopeLevel, nil), nil
	}
	if user.PrimaryDeptID == nil || *user.PrimaryDeptID == 0 {
		return newDataAccessScope(tenantID, scopeLevel, nil), nil
	}
	if scopeLevel == model.DataScopeSelfDepartment {
		return newDataAccessScope(tenantID, scopeLevel, []uint{*user.PrimaryDeptID}), nil
	}

	departments, err := s.deptRepo.ListHierarchyInTenant(tenantID)
	if err != nil {
		return nil, err
	}

	return newDataAccessScope(tenantID, scopeLevel, collectDepartmentTreeIDs(*user.PrimaryDeptID, departments)), nil
}

func (s *AccessScopeService) EnsureDepartmentAccess(ctx context.Context, tenantID uint, userID uint, departmentID uint) error {
	scope, err := s.Resolve(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if scope.AllowsDepartmentID(departmentID) {
		return nil
	}
	return ErrScopeForbidden
}

func (s *AccessScopeService) EnsureDepartmentsAccess(ctx context.Context, tenantID uint, userID uint, departmentIDs []uint) error {
	scope, err := s.Resolve(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if scope.AllowsAll() {
		return nil
	}
	for _, departmentID := range departmentIDs {
		if !scope.AllowsDepartmentID(departmentID) {
			return ErrScopeForbidden
		}
	}
	return nil
}

func (s *AccessScopeService) EnsureUserAccess(ctx context.Context, tenantID uint, userID uint, targetUserID uint) error {
	scope, err := s.Resolve(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if scope.AllowsAll() || userID == targetUserID {
		return nil
	}

	targetUser, err := s.userRepo.GetByIDInTenant(tenantID, targetUserID)
	if err != nil {
		return err
	}
	if targetUser.PrimaryDeptID != nil && scope.AllowsDepartmentID(*targetUser.PrimaryDeptID) {
		return nil
	}
	return ErrScopeForbidden
}

func (s *AccessScopeService) EnsureUsersAccess(ctx context.Context, tenantID uint, userID uint, targetUserIDs []uint) error {
	for _, targetUserID := range targetUserIDs {
		if err := s.EnsureUserAccess(ctx, tenantID, userID, targetUserID); err != nil {
			return err
		}
	}
	return nil
}

func collectDepartmentTreeIDs(rootID uint, departments []model.Department) []uint {
	childrenMap := make(map[uint][]uint)
	for _, department := range departments {
		if department.ParentID == nil {
			continue
		}
		childrenMap[*department.ParentID] = append(childrenMap[*department.ParentID], department.ID)
	}

	queue := []uint{rootID}
	visited := make(map[uint]struct{})
	result := make([]uint, 0, len(departments))

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, exists := visited[current]; exists {
			continue
		}
		visited[current] = struct{}{}
		result = append(result, current)
		queue = append(queue, childrenMap[current]...)
	}

	return result
}
