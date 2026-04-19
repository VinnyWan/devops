package service

import (
	"errors"

	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"

	"gorm.io/gorm"
)

type PermissionService struct {
	permRepo  *repository.PermissionRepo
	groupRepo *repository.GroupRepo
	hostRepo  *repository.HostRepo
}

func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{
		permRepo:  repository.NewPermissionRepo(db),
		groupRepo: repository.NewGroupRepo(db),
		hostRepo:  repository.NewHostRepo(db),
	}
}

type PermissionCreateRequest struct {
	UserID      uint     `json:"userId" binding:"required"`
	HostGroupID uint     `json:"hostGroupId" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

type PermissionUpdateRequest struct {
	ID         uint   `json:"id" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

func (s *PermissionService) normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func collectDescendantGroupIDs(childrenMap map[uint][]uint, groupID uint) []uint {
	result := []uint{groupID}
	for _, childID := range childrenMap[groupID] {
		result = append(result, collectDescendantGroupIDs(childrenMap, childID)...)
	}
	return result
}

var validPermissions = map[string]bool{"view": true, "terminal": true, "admin": true}

func (s *PermissionService) CreateInTenant(tenantID uint, userID uint, req *PermissionCreateRequest) ([]model.HostPermission, error) {
	_, err := s.groupRepo.GetByIDInTenant(tenantID, req.HostGroupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分组不存在")
		}
		return nil, err
	}

	created := make([]model.HostPermission, 0, len(req.Permissions))
	for _, perm := range req.Permissions {
		if !validPermissions[perm] {
			return nil, errors.New("无效的权限类型: " + perm)
		}
		exists, err := s.permRepo.ExistsByUserGroupPermission(tenantID, req.UserID, req.HostGroupID, perm)
		if err != nil {
			return nil, err
		}
		if exists {
			continue
		}
		record := &model.HostPermission{
			TenantID:    tenantID,
			UserID:      req.UserID,
			HostGroupID: req.HostGroupID,
			Permission:  perm,
			CreatedBy:   userID,
		}
		if err := s.permRepo.Create(record); err != nil {
			return nil, err
		}
		created = append(created, *record)
	}
	return created, nil
}

func (s *PermissionService) UpdateInTenant(tenantID uint, req *PermissionUpdateRequest) (*model.HostPermission, error) {
	if !validPermissions[req.Permission] {
		return nil, errors.New("无效的权限类型: " + req.Permission)
	}

	perm, err := s.permRepo.GetByIDInTenant(tenantID, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("权限规则不存在")
		}
		return nil, err
	}

	exists, err := s.permRepo.ExistsByUserGroupPermission(tenantID, perm.UserID, perm.HostGroupID, req.Permission)
	if err != nil {
		return nil, err
	}
	if exists && req.Permission != perm.Permission {
		return nil, errors.New("该用户在此分组上已有相同权限")
	}

	perm.Permission = req.Permission
	if err := s.permRepo.UpdateInTenant(tenantID, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *PermissionService) DeleteInTenant(tenantID, id uint) error {
	_, err := s.permRepo.GetByIDInTenant(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("权限规则不存在")
		}
		return err
	}
	return s.permRepo.DeleteInTenant(tenantID, id)
}

func (s *PermissionService) ListInTenant(tenantID uint, page, pageSize int, userID, hostGroupID uint, permission string) ([]model.HostPermission, int64, error) {
	page, pageSize = s.normalizePage(page, pageSize)
	return s.permRepo.ListInTenant(tenantID, page, pageSize, userID, hostGroupID, permission)
}

type PermissionHostEntry struct {
	HostID     uint   `json:"hostId"`
	Hostname   string `json:"hostname"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Permission string `json:"permission"`
	GroupID    uint   `json:"groupId"`
	GroupName  string `json:"groupName"`
}

func (s *PermissionService) MyHosts(tenantID, userID uint) ([]PermissionHostEntry, error) {
	perms, err := s.permRepo.GetByUserInTenant(tenantID, userID)
	if err != nil {
		return nil, err
	}
	if len(perms) == 0 {
		return []PermissionHostEntry{}, nil
	}

	allGroups, err := s.groupRepo.ListInTenant(tenantID)
	if err != nil {
		return nil, err
	}

	childrenMap := make(map[uint][]uint)
	groupMap := make(map[uint]model.HostGroup)
	for _, g := range allGroups {
		groupMap[g.ID] = g
		if g.ParentID != 0 {
			childrenMap[g.ParentID] = append(childrenMap[g.ParentID], g.ID)
		}
	}

	var collectDescendants func(groupID uint) []uint
	collectDescendants = func(groupID uint) []uint {
		result := []uint{groupID}
		for _, childID := range childrenMap[groupID] {
			result = append(result, collectDescendants(childID)...)
		}
		return result
	}

	expandedGroups := make(map[uint][]uint)
	groupPermSet := make(map[uint]map[string]bool)
	for _, p := range perms {
		if _, ok := expandedGroups[p.HostGroupID]; !ok {
			expandedGroups[p.HostGroupID] = collectDescendantGroupIDs(childrenMap, p.HostGroupID)
		}
		if groupPermSet[p.HostGroupID] == nil {
			groupPermSet[p.HostGroupID] = make(map[string]bool)
		}
		groupPermSet[p.HostGroupID][p.Permission] = true
	}

	allGroupIDs := make(map[uint]bool)
	for _, ids := range expandedGroups {
		for _, id := range ids {
			allGroupIDs[id] = true
		}
	}

	effectivePerm := make(map[uint]string)
	for groupID := range allGroupIDs {
		highest := ""
		current := groupID
		for current != 0 {
			if pSet, ok := groupPermSet[current]; ok {
				for p := range pSet {
					if permRank(p) > permRank(highest) {
						highest = p
					}
				}
			}
			g, ok := groupMap[current]
			if !ok {
				break
			}
			current = g.ParentID
		}
		if highest != "" {
			effectivePerm[groupID] = highest
		}
	}

	groupIDs := make([]uint, 0, len(effectivePerm))
	for groupID := range effectivePerm {
		groupIDs = append(groupIDs, groupID)
	}
	hosts, err := s.hostRepo.ListByGroupIDsInTenant(tenantID, groupIDs)
	if err != nil {
		return nil, err
	}

	result := make([]PermissionHostEntry, 0, len(hosts))
	for _, h := range hosts {
		if h.GroupID == nil {
			continue
		}
		groupID := *h.GroupID
		perm, ok := effectivePerm[groupID]
		if !ok {
			continue
		}
		gn := ""
		if g, ok := groupMap[groupID]; ok {
			gn = g.Name
		}
		result = append(result, PermissionHostEntry{
			HostID:     h.ID,
			Hostname:   h.Hostname,
			Ip:         h.Ip,
			Port:       h.Port,
			Permission: perm,
			GroupID:    groupID,
			GroupName:  gn,
		})
	}

	return result, nil
}

func (s *PermissionService) CheckPermission(tenantID, userID, hostID uint, action string) (bool, string, error) {
	if !validPermissions[action] {
		return false, "", errors.New("无效的权限类型: " + action)
	}

	host, err := s.hostRepo.GetByIDInTenant(tenantID, hostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", errors.New("主机不存在")
		}
		return false, "", err
	}
	if host.GroupID == nil {
		return false, "", nil
	}

	allGroups, err := s.groupRepo.ListInTenant(tenantID)
	if err != nil {
		return false, "", err
	}
	groupMap := make(map[uint]model.HostGroup)
	for _, g := range allGroups {
		groupMap[g.ID] = g
	}

	perms, err := s.permRepo.GetByUserInTenant(tenantID, userID)
	if err != nil {
		return false, "", err
	}
	permMap := make(map[uint]map[string]bool)
	for _, p := range perms {
		if permMap[p.HostGroupID] == nil {
			permMap[p.HostGroupID] = make(map[string]bool)
		}
		permMap[p.HostGroupID][p.Permission] = true
	}

	current := *host.GroupID
	highest := ""
	for current != 0 {
		if pSet, ok := permMap[current]; ok {
			for p := range pSet {
				if permRank(p) > permRank(highest) {
					highest = p
				}
			}
		}
		g, ok := groupMap[current]
		if !ok {
			break
		}
		current = g.ParentID
	}

	if highest == "" {
		return false, "", nil
	}

	if permRank(highest) >= permRank(action) {
		return true, highest, nil
	}
	return false, highest, nil
}

func (s *PermissionService) GetGroupHostCount(tenantID, groupID uint) (int64, error) {
	allGroups, err := s.groupRepo.ListInTenant(tenantID)
	if err != nil {
		return 0, err
	}

	childrenMap := make(map[uint][]uint)
	for _, g := range allGroups {
		if g.ParentID != 0 {
			childrenMap[g.ParentID] = append(childrenMap[g.ParentID], g.ID)
		}
	}

	ids := collectDescendantGroupIDs(childrenMap, groupID)
	return s.hostRepo.CountByGroupIDsInTenant(tenantID, ids)
}

func permRank(perm string) int {
	switch perm {
	case "admin":
		return 3
	case "terminal":
		return 2
	case "view":
		return 1
	default:
		return 0
	}
}
