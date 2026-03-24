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
}

func NewDepartmentService(deptRepo *repository.DepartmentRepo, userRepo *repository.UserRepo) *DepartmentService {
	return &DepartmentService{
		deptRepo: deptRepo,
		userRepo: userRepo,
	}
}

type CreateDepartmentRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID *uint  `json:"parentId"`
}

func (s *DepartmentService) Create(ctx context.Context, req *CreateDepartmentRequest) (*model.Department, error) {
	dept := &model.Department{
		Name:     req.Name,
		ParentID: req.ParentID,
	}
	if err := s.deptRepo.Create(dept); err != nil {
		return nil, err
	}
	s.invalidateCache(ctx)
	return dept, nil
}

type UpdateDepartmentRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Name     string `json:"name"`
	ParentID *uint  `json:"parentId"`
}

func (s *DepartmentService) Update(ctx context.Context, req *UpdateDepartmentRequest) error {
	dept, err := s.deptRepo.GetByID(req.ID)
	if err != nil {
		return err
	}

	if req.Name != "" {
		dept.Name = req.Name
	}
	if req.ParentID != nil {
		dept.ParentID = req.ParentID
	}

	if err := s.deptRepo.Update(dept); err != nil {
		return err
	}
	s.invalidateCache(ctx)
	return nil
}

func (s *DepartmentService) Delete(ctx context.Context, id uint) error {
	// 检查是否有子部门或关联用户，这里简化处理，直接删除
	if err := s.deptRepo.Delete(id); err != nil {
		return err
	}
	s.invalidateCache(ctx)
	return nil
}

func (s *DepartmentService) GetTree(ctx context.Context, keyword string) ([]*model.Department, error) {
	normalizedKeyword := queryutil.NormalizeKeyword(keyword)
	cacheKey := "dept:tree"
	if normalizedKeyword == "" {
		val, err := redis.Get(ctx, cacheKey)
		if err == nil && val != "" {
			var roots []*model.Department
			if err := json.Unmarshal([]byte(val), &roots); err == nil {
				return roots, nil
			}
		}
	}

	list, err := s.deptRepo.List(normalizedKeyword)
	if err != nil {
		return nil, err
	}

	// 构建树状结构
	deptMap := make(map[uint]*model.Department)
	var roots []*model.Department

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

	return roots, nil
}

func (s *DepartmentService) AssignRoles(ctx context.Context, deptID uint, roleIDs []uint) error {
	affectedUserIDs, err := s.userRepo.ListUserIDsByDepartmentID(deptID)
	if err != nil {
		return err
	}
	if err := s.deptRepo.AssignRoles(deptID, roleIDs); err != nil {
		return err
	}
	for _, userID := range affectedUserIDs {
		redis.Del(ctx, fmt.Sprintf("user:perms:%d", userID))
	}
	s.invalidateCache(ctx)
	return nil
}

func (s *DepartmentService) invalidateCache(ctx context.Context) {
	redis.Del(ctx, "dept:tree")
}
