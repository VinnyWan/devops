package service

import (
	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type GroupService struct {
	repo *repository.GroupRepo
}

func NewGroupService(db *gorm.DB) *GroupService {
	return &GroupService{repo: repository.NewGroupRepo(db)}
}

// GroupTree 分组树节点
type GroupTree struct {
	ID        uint         `json:"id"`
	Name      string       `json:"name"`
	Level     int          `json:"level"`
	ParentID  uint         `json:"parentId"`
	SortOrder int          `json:"sortOrder"`
	Children  []GroupTree  `json:"children,omitempty"`
	HostCount int64        `json:"hostCount"`
}

type GroupCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	Level     int    `json:"level" binding:"required"`
	ParentID  uint   `json:"parentId"`
	SortOrder int    `json:"sortOrder"`
}

type GroupUpdateRequest struct {
	ID        uint   `json:"id" binding:"required"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
}

func (s *GroupService) ListInTenant(tenantID uint) ([]model.HostGroup, error) {
	return s.repo.ListInTenant(tenantID)
}

// GetTreeInTenant 构建三级树结构
func (s *GroupService) GetTreeInTenant(tenantID uint) ([]GroupTree, error) {
	groups, err := s.repo.ListInTenant(tenantID)
	if err != nil {
		return nil, err
	}

	// 构建 map
	groupMap := make(map[uint]*GroupTree)
	for _, g := range groups {
		groupMap[g.ID] = &GroupTree{
			ID:        g.ID,
			Name:      g.Name,
			Level:     g.Level,
			ParentID:  g.ParentID,
			SortOrder: g.SortOrder,
			Children:  []GroupTree{},
		}
	}

	// 构建树
	childrenMap := make(map[uint][]GroupTree)
	for _, g := range groups {
		node := *groupMap[g.ID]
		if g.ParentID == 0 || g.Level == 1 {
			continue
		}
		childrenMap[g.ParentID] = append(childrenMap[g.ParentID], node)
	}

	var buildChildren func(parentID uint) []GroupTree
	buildChildren = func(parentID uint) []GroupTree {
		children := childrenMap[parentID]
		for i := range children {
			children[i].Children = buildChildren(children[i].ID)
		}
		return children
	}

	var roots []GroupTree
	for _, g := range groups {
		node := *groupMap[g.ID]
		if g.ParentID == 0 || g.Level == 1 {
			node.Children = buildChildren(node.ID)
			roots = append(roots, node)
		}
	}

	return roots, nil
}

func (s *GroupService) GetByIDInTenant(tenantID uint, id uint) (*model.HostGroup, error) {
	return s.repo.GetByIDInTenant(tenantID, id)
}

func (s *GroupService) CreateInTenant(tenantID uint, req *GroupCreateRequest) (*model.HostGroup, error) {
	// 校验层级约束
	if req.Level < 1 || req.Level > 3 {
		return nil, errors.New("层级必须为 1-3")
	}

	if req.Level == 1 && req.ParentID != 0 {
		return nil, errors.New("一级分组的 parent_id 必须为 0")
	}

	if req.Level > 1 && req.ParentID == 0 {
		return nil, fmt.Errorf("第 %d 级分组必须指定父分组", req.Level)
	}

	if req.Level > 1 {
		parent, err := s.repo.GetByIDInTenant(tenantID, req.ParentID)
		if err != nil {
			return nil, errors.New("父分组不存在")
		}
		if parent.Level != req.Level-1 {
			return nil, fmt.Errorf("父分组层级必须为 %d", req.Level-1)
		}
	}

	group := &model.HostGroup{
		Name:      req.Name,
		Level:     req.Level,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
	}

	if err := s.repo.CreateInTenant(tenantID, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) UpdateInTenant(tenantID uint, req *GroupUpdateRequest) (*model.HostGroup, error) {
	group, err := s.repo.GetByIDInTenant(tenantID, req.ID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		group.Name = req.Name
	}
	group.SortOrder = req.SortOrder

	if err := s.repo.UpdateInTenant(tenantID, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) DeleteInTenant(tenantID uint, id uint) error {
	// 检查是否有子分组
	hasChildren, err := s.repo.HasChildren(id)
	if err != nil {
		return err
	}
	if hasChildren {
		return errors.New("该分组下有子分组，无法删除")
	}

	return s.repo.DeleteInTenant(tenantID, id)
}
