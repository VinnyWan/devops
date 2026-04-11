package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	queryutil "devops-platform/internal/pkg/query"
	"devops-platform/internal/pkg/redis"
)

type DepartmentService struct {
	deptRepo *repository.DepartmentRepo
	userRepo *repository.UserRepo
	scopeSvc *AccessScopeService
}

func NewDepartmentService(deptRepo *repository.DepartmentRepo, userRepo *repository.UserRepo) *DepartmentService {
	return &DepartmentService{
		deptRepo: deptRepo,
		userRepo: userRepo,
		scopeSvc: NewAccessScopeService(userRepo, deptRepo),
	}
}

type CreateDepartmentRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID *uint  `json:"parentId"`
}

func (s *DepartmentService) Create(ctx context.Context, tenantID uint, operatorID uint, req *CreateDepartmentRequest) (*model.Department, error) {
	scope, err := s.scopeSvc.Resolve(ctx, tenantID, operatorID)
	if err != nil {
		return nil, err
	}
	if req.ParentID != nil {
		if _, err := s.deptRepo.GetByIDInTenant(tenantID, *req.ParentID); err != nil {
			return nil, fmt.Errorf("parent department not found: %w", err)
		}
		if !scope.AllowsDepartmentID(*req.ParentID) {
			return nil, ErrScopeForbidden
		}
	} else if !scope.AllowsAll() {
		return nil, ErrScopeForbidden
	}
	dept := &model.Department{
		TenantID: &tenantID,
		Name:     req.Name,
		ParentID: req.ParentID,
	}
	if err := s.deptRepo.Create(dept); err != nil {
		return nil, err
	}
	s.invalidateCache(ctx, tenantID)
	return dept, nil
}

type UpdateDepartmentRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Name     string `json:"name"`
	ParentID *uint  `json:"parentId"`
}

func (s *DepartmentService) Update(ctx context.Context, tenantID uint, operatorID uint, req *UpdateDepartmentRequest) error {
	if err := s.scopeSvc.EnsureDepartmentAccess(ctx, tenantID, operatorID, req.ID); err != nil {
		return err
	}
	dept, err := s.deptRepo.GetByIDInTenant(tenantID, req.ID)
	if err != nil {
		return err
	}

	if req.Name != "" {
		dept.Name = req.Name
	}
	if req.ParentID != nil {
		if *req.ParentID == req.ID {
			return fmt.Errorf("parent department cannot be self")
		}
		if _, err := s.deptRepo.GetByIDInTenant(tenantID, *req.ParentID); err != nil {
			return fmt.Errorf("parent department not found: %w", err)
		}
		if err := s.scopeSvc.EnsureDepartmentAccess(ctx, tenantID, operatorID, *req.ParentID); err != nil {
			return err
		}
		dept.ParentID = req.ParentID
	}

	if err := s.deptRepo.Update(dept); err != nil {
		return err
	}
	s.invalidateCache(ctx, tenantID)
	return nil
}

func (s *DepartmentService) Delete(ctx context.Context, tenantID uint, operatorID uint, id uint) error {
	if err := s.scopeSvc.EnsureDepartmentAccess(ctx, tenantID, operatorID, id); err != nil {
		return err
	}
	// 检查是否有子部门或关联用户，这里简化处理，直接删除
	if err := s.deptRepo.DeleteInTenant(tenantID, id); err != nil {
		return err
	}
	s.invalidateCache(ctx, tenantID)
	return nil
}

func (s *DepartmentService) GetTree(ctx context.Context, tenantID uint, operatorID uint, keyword string) ([]*model.Department, error) {
	normalizedKeyword := queryutil.NormalizeKeyword(keyword)
	cacheKey := fmt.Sprintf("tenant:%d:dept:tree", tenantID)
	scope, err := s.scopeSvc.Resolve(ctx, tenantID, operatorID)
	if err != nil {
		return nil, err
	}

	var roots []*model.Department
	if normalizedKeyword == "" {
		val, err := redis.Get(ctx, cacheKey)
		if err == nil && val != "" {
			var cachedRoots []*model.Department
			if err := json.Unmarshal([]byte(val), &cachedRoots); err == nil {
				if scope.AllowsAll() {
					return cachedRoots, nil
				}
				return filterDepartmentTree(cachedRoots, scope), nil
			}
		}
	}

	list, err := s.deptRepo.ListInTenant(tenantID, normalizedKeyword)
	if err != nil {
		return nil, err
	}

	// 构建树状结构
	deptMap := make(map[uint]*model.Department)
	for i := range list {
		deptMap[list[i].ID] = &list[i]
	}

	for i := range list {
		dept := &list[i]
		if dept.ParentID == nil {
			roots = append(roots, dept)
		} else {
			if parent, ok := deptMap[*dept.ParentID]; ok {
				parent.Children = append(parent.Children, dept)
			} else {
				roots = append(roots, dept)
			}
		}
	}

	// Set Cache
	if normalizedKeyword == "" {
		if data, err := json.Marshal(roots); err == nil {
			redis.Set(ctx, cacheKey, string(data), 5*time.Minute)
		}
	}

	if scope.AllowsAll() {
		return roots, nil
	}

	return filterDepartmentTree(roots, scope), nil
}

func (s *DepartmentService) AssignRoles(ctx context.Context, tenantID uint, operatorID uint, deptID uint, roleIDs []uint) error {
	if err := s.scopeSvc.EnsureDepartmentAccess(ctx, tenantID, operatorID, deptID); err != nil {
		return err
	}
	affectedUserIDs, err := s.userRepo.ListUserIDsByDepartmentIDInTenant(tenantID, deptID)
	if err != nil {
		return err
	}
	if err := s.deptRepo.AssignRolesInTenant(tenantID, deptID, roleIDs); err != nil {
		return err
	}
	for _, userID := range affectedUserIDs {
		redis.Del(ctx, fmt.Sprintf("tenant:%d:user:perms:%d", tenantID, userID))
	}
	s.invalidateCache(ctx, tenantID)
	return nil
}

func filterDepartmentTree(nodes []*model.Department, scope *DataAccessScope) []*model.Department {
	if scope == nil || scope.AllowsAll() {
		return nodes
	}

	filtered := make([]*model.Department, 0, len(nodes))
	for _, node := range nodes {
		filteredChildren := filterDepartmentTree(node.Children, scope)
		if scope.AllowsDepartmentID(node.ID) {
			copyNode := *node
			copyNode.Children = filteredChildren
			filtered = append(filtered, &copyNode)
			continue
		}
		filtered = append(filtered, filteredChildren...)
	}
	return filtered
}

func (s *DepartmentService) invalidateCache(ctx context.Context, tenantID uint) {
	redis.Del(ctx, fmt.Sprintf("tenant:%d:dept:tree", tenantID))
}
