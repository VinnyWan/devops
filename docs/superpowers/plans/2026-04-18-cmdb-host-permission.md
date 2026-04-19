# CMDB 主机权限配置 - 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 CMDB 添加按分组授权的主机权限 CRUD 管理，支持 view/terminal/admin 三级权限和分组继承。

**Architecture:** 新增 HostPermission 模型，按 (tenant_id, user_id, host_group_id, permission) 唯一约束存储权限规则。Service 层负责分组继承解析（查询时向上遍历分组树），不做物化冗余。API 层提供 CRUD + /my-hosts + /check 接口。前端 PermissionList.vue 左侧分组树 + 右侧权限表格。

**Tech Stack:** Go/Gin/GORM (backend), Vue 3 + Element Plus (frontend), MySQL (database)

**Spec:** `docs/superpowers/specs/2026-04-18-cmdb-host-permission-design.md`

---

## File Structure

**New files:**
- `backend/internal/modules/cmdb/model/permission.go` — HostPermission GORM 模型
- `backend/internal/modules/cmdb/repository/permission.go` — CRUD + 查询
- `backend/internal/modules/cmdb/service/permission.go` — 业务逻辑 + 继承解析
- `backend/internal/modules/cmdb/api/permission.go` — HTTP Handler
- `frontend/src/api/cmdb/permission.js` — 前端 API 客户端
- `frontend/src/views/Cmdb/PermissionList.vue` — 权限管理页面

**Modified files:**
- `backend/internal/bootstrap/db.go` — AutoMigrate + 权限种子
- `backend/internal/modules/cmdb/api/common.go` — permissionSvcInstance + getter
- `backend/routers/v1/cmdb.go` — 注册权限路由
- `frontend/src/router/index.js` — 添加 /cmdb/permissions 路由
- `frontend/src/components/Layout/MainLayout.vue` — 侧边栏菜单项

---

### Task 1: Model + Bootstrap

**Files:**
- Create: `backend/internal/modules/cmdb/model/permission.go`
- Modify: `backend/internal/bootstrap/db.go:64-77` (AutoMigrate) and `backend/internal/bootstrap/db.go:244-335` (seedPermissions)

- [ ] **Step 1: Create HostPermission model**

```go
// backend/internal/modules/cmdb/model/permission.go
package model

import (
	"time"

	"gorm.io/gorm"
)

type HostPermission struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"not null;uniqueIndex:uk_perm_user_group,priority:1" json:"tenantId"`
	UserID      uint           `gorm:"not null;uniqueIndex:uk_perm_user_group,priority:2;index:idx_perm_user" json:"userId"`
	HostGroupID uint           `gorm:"not null;uniqueIndex:uk_perm_user_group,priority:3;index:idx_perm_group" json:"hostGroupId"`
	Permission  string         `gorm:"size:20;not null;uniqueIndex:uk_perm_user_group,priority:4" json:"permission"`
	CreatedBy   uint           `gorm:"not null" json:"createdBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
```

- [ ] **Step 2: Register model in AutoMigrate**

Add `&cmdbModel.HostPermission{}` to the AutoMigrate call in `backend/internal/bootstrap/db.go` after the `&cmdbModel.TerminalSession{}` line (around line 76):

```go
err = db.AutoMigrate(
	&userModel.Tenant{},
	&userModel.Department{},
	&userModel.User{},
	&userModel.Role{},
	&userModel.Permission{},
	&userModel.AuditLog{},
	&userModel.LoginLog{},
	&k8sModel.Cluster{},
	&cmdbModel.Host{},
	&cmdbModel.HostGroup{},
	&cmdbModel.Credential{},
	&cmdbModel.TerminalSession{},
	&cmdbModel.HostPermission{}, // <-- add this line
)
```

- [ ] **Step 3: Add permission seeds**

Add these entries to the `permissions` slice in `seedPermissions()` (in `backend/internal/bootstrap/db.go`), after the existing CMDB terminal permissions (after line 316):

```go
// CMDB 权限配置
{Name: "查看权限配置", Resource: "cmdb:permission", Action: "list", Description: "查看主机权限配置"},
{Name: "授予权限", Resource: "cmdb:permission", Action: "create", Description: "授予主机权限"},
{Name: "更新权限", Resource: "cmdb:permission", Action: "update", Description: "更新主机权限"},
{Name: "删除权限", Resource: "cmdb:permission", Action: "delete", Description: "删除主机权限"},
```

- [ ] **Step 4: Build and verify migration**

Run: `cd backend && go build ./...`
Expected: compiles with no errors

- [ ] **Step 5: Commit**

```bash
git add backend/internal/modules/cmdb/model/permission.go backend/internal/bootstrap/db.go
git commit -m "feat(cmdb): add HostPermission model and seed data"
```

---

### Task 2: Repository

**Files:**
- Create: `backend/internal/modules/cmdb/repository/permission.go`

- [ ] **Step 1: Create permission repository**

```go
// backend/internal/modules/cmdb/repository/permission.go
package repository

import (
	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
)

type PermissionRepo struct {
	db *gorm.DB
}

func NewPermissionRepo(db *gorm.DB) *PermissionRepo {
	return &PermissionRepo{db: db}
}

func (r *PermissionRepo) scopeInTenant(query *gorm.DB, tenantID uint) *gorm.DB {
	return query.Where("tenant_id = ?", tenantID)
}

func (r *PermissionRepo) Create(perm *model.HostPermission) error {
	return r.db.Create(perm).Error
}

func (r *PermissionRepo) GetByIDInTenant(tenantID, id uint) (*model.HostPermission, error) {
	var perm model.HostPermission
	if err := r.scopeInTenant(r.db, tenantID).Where("id = ?", id).First(&perm).Error; err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *PermissionRepo) UpdateInTenant(tenantID uint, perm *model.HostPermission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.HostPermission
		if err := r.scopeInTenant(tx, tenantID).Where("id = ?", perm.ID).First(&existing).Error; err != nil {
			return err
		}
		return tx.Save(perm).Error
	})
}

func (r *PermissionRepo) DeleteInTenant(tenantID, id uint) error {
	return r.scopeInTenant(r.db, tenantID).Where("id = ?", id).Delete(&model.HostPermission{}).Error
}

func (r *PermissionRepo) ListInTenant(tenantID uint, page, pageSize int, userID, hostGroupID uint, permission string) ([]model.HostPermission, int64, error) {
	var perms []model.HostPermission
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.HostPermission{}), tenantID)

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if hostGroupID > 0 {
		query = query.Where("host_group_id = ?", hostGroupID)
	}
	if permission != "" {
		query = query.Where("permission = ?", permission)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&perms).Error; err != nil {
		return nil, 0, err
	}

	return perms, total, nil
}

func (r *PermissionRepo) GetByUserInTenant(tenantID, userID uint) ([]model.HostPermission, error) {
	var perms []model.HostPermission
	if err := r.scopeInTenant(r.db, tenantID).Where("user_id = ?", userID).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *PermissionRepo) ExistsByUserGroupPermission(tenantID, userID, hostGroupID uint, permission string) (bool, error) {
	var count int64
	if err := r.scopeInTenant(r.db.Model(&model.HostPermission{}), tenantID).
		Where("user_id = ? AND host_group_id = ? AND permission = ?", userID, hostGroupID, permission).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PermissionRepo) DeleteByUserGroupInTenant(tenantID, userID, hostGroupID uint) error {
	return r.scopeInTenant(r.db, tenantID).
		Where("user_id = ? AND host_group_id = ?", userID, hostGroupID).
		Delete(&model.HostPermission{}).Error
}
```

- [ ] **Step 2: Build**

Run: `cd backend && go build ./...`
Expected: compiles with no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/modules/cmdb/repository/permission.go
git commit -m "feat(cmdb): add permission repository with CRUD and queries"
```

---

### Task 3: Service

**Files:**
- Create: `backend/internal/modules/cmdb/service/permission.go`

- [ ] **Step 1: Create permission service with inheritance logic**

```go
// backend/internal/modules/cmdb/service/permission.go
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

var validPermissions = map[string]bool{"view": true, "terminal": true, "admin": true}

func (s *PermissionService) CreateInTenant(tenantID uint, userID uint, req *PermissionCreateRequest) ([]model.HostPermission, error) {
	// Validate group exists
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

	// Check for duplicate after update
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

// PermissionHostEntry /my-hosts 返回的单条主机权限信息
type PermissionHostEntry struct {
	HostID     uint   `json:"hostId"`
	Hostname   string `json:"hostname"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	Permission string `json:"permission"`
	GroupID    uint   `json:"groupId"`
	GroupName  string `json:"groupName"`
}

// MyHosts 返回当前用户有权限的主机列表（含继承）
func (s *PermissionService) MyHosts(tenantID, userID uint) ([]PermissionHostEntry, error) {
	perms, err := s.permRepo.GetByUserInTenant(tenantID, userID)
	if err != nil {
		return nil, err
	}
	if len(perms) == 0 {
		return []PermissionHostEntry{}, nil
	}

	// Load all groups to build parent→children map
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

	// For each permission, collect all descendant group IDs (including self)
	var collectDescendants func(groupID uint) []uint
	collectDescendants = func(groupID uint) []uint {
		result := []uint{groupID}
		for _, childID := range childrenMap[groupID] {
			result = append(result, collectDescendants(childID)...)
		}
		return result
	}

	// Map: groupID → set of expanded group IDs
	expandedGroups := make(map[uint][]uint)
	groupPermSet := make(map[uint]map[string]bool) // groupID → permissions
	for _, p := range perms {
		if _, ok := expandedGroups[p.HostGroupID]; !ok {
			expandedGroups[p.HostGroupID] = collectDescendants(p.HostGroupID)
		}
		if groupPermSet[p.HostGroupID] == nil {
			groupPermSet[p.HostGroupID] = make(map[string]bool)
		}
		groupPermSet[p.HostGroupID][p.Permission] = true
	}

	// Collect all group IDs that are covered
	allGroupIDs := make(map[uint]bool)
	for _, ids := range expandedGroups {
		for _, id := range ids {
			allGroupIDs[id] = true
		}
	}

	// For each covered group, determine effective permission from ancestors
	effectivePerm := make(map[uint]string) // groupID → highest permission
	for groupID := range allGroupIDs {
		highest := ""
		// Walk up from this group to root, checking each ancestor's permission set
		current := groupID
		for current != 0 {
			if perms2, ok := groupPermSet[current]; ok {
				for p := range perms2 {
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

	// Get hosts for all covered groups
	result := make([]PermissionHostEntry, 0)
	for groupID, perm := range effectivePerm {
		hosts, _, err := s.hostRepo.GetByGroupIDInTenant(tenantID, groupID, 1, 1000)
		if err != nil {
			continue
		}
		gn := ""
		if g, ok := groupMap[groupID]; ok {
			gn = g.Name
		}
		for _, h := range hosts {
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
	}

	return result, nil
}

// CheckPermission 检查用户对指定主机的权限
func (s *PermissionService) CheckPermission(tenantID, userID, hostID uint, action string) (bool, string, error) {
	if !validPermissions[action] {
		return false, "", errors.New("无效的权限类型: " + action)
	}

	// Get host to find its group
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

	// Load all groups for ancestor walk
	allGroups, err := s.groupRepo.ListInTenant(tenantID)
	if err != nil {
		return false, "", err
	}
	groupMap := make(map[uint]model.HostGroup)
	for _, g := range allGroups {
		groupMap[g.ID] = g
	}

	// Get user's permissions
	perms, err := s.permRepo.GetByUserInTenant(tenantID, userID)
	if err != nil {
		return false, "", err
	}
	permMap := make(map[uint]map[string]bool) // groupID → permissions
	for _, p := range perms {
		if permMap[p.HostGroupID] == nil {
			permMap[p.HostGroupID] = make(map[string]bool)
		}
		permMap[p.HostGroupID][p.Permission] = true
	}

	// Walk up from host's group to root
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

	// Check if highest permission satisfies the requested action
	if permRank(highest) >= permRank(action) {
		return true, highest, nil
	}
	return false, highest, nil
}

// GetGroupHostCount 返回分组及其子分组下的主机总数
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

	var collectDescendants func(id uint) []uint
	collectDescendants = func(id uint) []uint {
		result := []uint{id}
		for _, childID := range childrenMap[id] {
			result = append(result, collectDescendants(childID)...)
		}
		return result
	}

	ids := collectDescendants(groupID)
	var total int64
	for _, id := range ids {
		hosts, count, err := s.hostRepo.GetByGroupIDInTenant(tenantID, id, 1, 1)
		_ = hosts
		if err != nil {
			continue
		}
		total += count
	}
	return total, nil
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
```

- [ ] **Step 2: Build**

Run: `cd backend && go build ./...`
Expected: compiles with no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/modules/cmdb/service/permission.go
git commit -m "feat(cmdb): add permission service with group inheritance"
```

---

### Task 4: API Handlers + Routes

**Files:**
- Create: `backend/internal/modules/cmdb/api/permission.go`
- Modify: `backend/internal/modules/cmdb/api/common.go` — add permSvcInstance + getter
- Modify: `backend/routers/v1/cmdb.go` — register permission routes

- [ ] **Step 1: Add permission service to common.go**

Add `permSvcInstance` to the var block and add `getPermissionService()` function in `backend/internal/modules/cmdb/api/common.go`.

In the `var` block (after `terminalSvcInstance`), add:

```go
	permSvcInstance     *service.PermissionService
```

In `SetDB`, add nil reset for `permSvcInstance`:

```go
	permSvcInstance = nil
```

Add new getter function after `getTerminalService()`:

```go
func getPermissionService() *service.PermissionService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if permSvcInstance != nil {
		return permSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	permSvcInstance = service.NewPermissionService(cmdbDB)
	return permSvcInstance
}
```

- [ ] **Step 2: Create permission API handler**

```go
// backend/internal/modules/cmdb/api/permission.go
package api

import (
	"errors"
	"net/http"
	"strconv"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionList 权限规则列表
func PermissionList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("userId", "0"), 10, 32)
	hostGroupID, _ := strconv.ParseUint(c.DefaultQuery("hostGroupId", "0"), 10, 32)
	permission := c.Query("permission")

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	perms, total, err := svc.ListInTenant(tenantID, page, pageSize, uint(userID), uint(hostGroupID), permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取权限列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "获取成功",
		"data":     perms,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// PermissionCreate 授予权限
func PermissionCreate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	userIDValue, _ := c.Get("userID")
	currentUserID := userIDValue.(uint)

	var req service.PermissionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	created, err := svc.CreateInTenant(tenantID, currentUserID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "授予权限失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "授权成功",
		"data":    created,
	})
}

// PermissionUpdate 更新权限
func PermissionUpdate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.PermissionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	perm, err := svc.UpdateInTenant(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "更新权限失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    perm,
	})
}

// PermissionDelete 删除权限
func PermissionDelete(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	if err := svc.DeleteInTenant(tenantID, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "删除权限失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// PermissionMyHosts 当前用户可访问的主机
func PermissionMyHosts(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	userIDValue, _ := c.Get("userID")
	userID := userIDValue.(uint)

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	hosts, err := svc.MyHosts(tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取主机列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    hosts,
	})
}

// PermissionCheck 检查用户对主机的权限
func PermissionCheck(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	userIDValue, _ := c.Get("userID")
	userID := userIDValue.(uint)

	hostID, _ := strconv.ParseUint(c.DefaultQuery("hostId", "0"), 10, 32)
	action := c.DefaultQuery("action", "view")
	if hostID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少 hostId 参数"})
		return
	}

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	allowed, perm, err := svc.CheckPermission(tenantID, userID, uint(hostID), action)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查权限失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"allowed":    allowed,
			"permission": perm,
		},
	})
}

// PermissionGroupHostCount 返回分组及其子分组下的主机数量
func PermissionGroupHostCount(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	groupID, _ := strconv.ParseUint(c.DefaultQuery("groupId", "0"), 10, 32)
	if groupID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少 groupId 参数"})
		return
	}

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	count, err := svc.GetGroupHostCount(tenantID, uint(groupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取主机数量失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"hostCount": count,
		},
	})
}
```

- [ ] **Step 3: Register permission routes in cmdb.go**

Add permission middleware definitions and routes to `backend/routers/v1/cmdb.go`. After the existing `terminalReplayPerm` line (line 29), add:

```go
	permListPerm := middleware.RequirePermission("cmdb:permission", "list")
	permCreatePerm := middleware.RequirePermission("cmdb:permission", "create")
	permUpdatePerm := middleware.RequirePermission("cmdb:permission", "update")
	permDeletePerm := middleware.RequirePermission("cmdb:permission", "delete")
```

Add a new route group inside the main block (after the terminal routes, before the closing `}`):

```go
			// 权限配置
			g.GET("/permission/list", permListPerm, api.PermissionList)
			g.GET("/permission/group-host-count", permListPerm, api.PermissionGroupHostCount)
			g.POST("/permission/create", permCreatePerm, middleware.SetAuditOperation("授予权限"), api.PermissionCreate)
			g.POST("/permission/update", permUpdatePerm, middleware.SetAuditOperation("更新权限"), api.PermissionUpdate)
			g.POST("/permission/delete", permDeletePerm, middleware.SetAuditOperation("删除权限"), api.PermissionDelete)
			g.GET("/permission/my-hosts", api.PermissionMyHosts)
			g.GET("/permission/check", api.PermissionCheck)
```

Note: `/my-hosts` and `/check` are login-only (no specific permission required), so they don't have permission middleware.

- [ ] **Step 4: Build**

Run: `cd backend && go build ./...`
Expected: compiles with no errors

- [ ] **Step 5: Start server and verify migration**

Run: `cd backend && DEVOPS_SERVER_PORT=8001 go run cmd/server/main.go`

Wait for "服务启动" log message. Check that `host_permissions` table was created in MySQL. Stop the server with Ctrl+C.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/modules/cmdb/api/permission.go backend/internal/modules/cmdb/api/common.go backend/routers/v1/cmdb.go
git commit -m "feat(cmdb): add permission API handlers and routes"
```

---

### Task 5: Frontend API Client + Route + Sidebar

**Files:**
- Create: `frontend/src/api/cmdb/permission.js`
- Modify: `frontend/src/router/index.js:97-105` (add permission route)
- Modify: `frontend/src/components/Layout/MainLayout.vue:43-52` (add sidebar item)

- [ ] **Step 1: Create permission API client**

```javascript
// frontend/src/api/cmdb/permission.js
import request from '../request'

export const getPermissionList = (params) => request.get('/cmdb/permission/list', { params })
export const createPermission = (data) => request.post('/cmdb/permission/create', data)
export const updatePermission = (data) => request.post('/cmdb/permission/update', data)
export const deletePermission = (data) => request.post('/cmdb/permission/delete', data)
export const getMyHosts = () => request.get('/cmdb/permission/my-hosts')
export const checkPermission = (params) => request.get('/cmdb/permission/check', { params })
export const getGroupHostCount = (params) => request.get('/cmdb/permission/group-host-count', { params })
```

- [ ] **Step 2: Add route in router/index.js**

Add after the `cmdb/terminal/replay/:id` route entry (after line 104):

```javascript
      {
        path: 'cmdb/permissions',
        component: () => import('../views/Cmdb/PermissionList.vue')
      }
```

- [ ] **Step 3: Add sidebar menu item in MainLayout.vue**

Add after the "终端审计" menu item (after line 51):

```html
          <el-menu-item index="/cmdb/permissions">权限配置</el-menu-item>
```

- [ ] **Step 4: Build frontend**

Run: `cd frontend && npm run build`
Expected: build succeeds

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/cmdb/permission.js frontend/src/router/index.js frontend/src/components/Layout/MainLayout.vue
git commit -m "feat(cmdb): add permission frontend API, route, and sidebar"
```

---

### Task 6: Frontend PermissionList.vue

**Files:**
- Create: `frontend/src/views/Cmdb/PermissionList.vue`

- [ ] **Step 1: Create PermissionList.vue**

```vue
<template>
  <div class="permission-page">
    <div class="left-panel">
      <div class="panel-header">
        <h3>分组</h3>
      </div>
      <div class="tree-wrap">
        <el-tree
          ref="treeRef"
          :data="treeData"
          node-key="id"
          highlight-current
          default-expand-all
          :props="{ label: 'name', children: 'children' }"
          @node-click="handleNodeClick"
        />
      </div>
    </div>
    <div class="right-panel">
      <div class="page-container">
        <div class="page-header">
          <h3>权限配置</h3>
          <el-button type="primary" @click="showCreateDialog">授予权限</el-button>
        </div>
        <div class="toolbar">
          <el-select v-model="filterUserId" placeholder="全部用户" clearable filterable style="width: 180px" @change="fetchData">
            <el-option v-for="u in userList" :key="u.id" :label="u.username" :value="u.id" />
          </el-select>
          <el-select v-model="filterPermission" placeholder="全部权限" clearable style="width: 150px" @change="fetchData">
            <el-option label="查看" value="view" />
            <el-option label="终端" value="terminal" />
            <el-option label="管理" value="admin" />
          </el-select>
        </div>
        <el-table :data="tableData" stripe v-loading="loading">
          <el-table-column prop="userId" label="用户 ID" width="100" />
          <el-table-column label="用户名" width="140">
            <template #default="{ row }">{{ getUsername(row.userId) }}</template>
          </el-table-column>
          <el-table-column label="分组" min-width="180">
            <template #default="{ row }">{{ getGroupName(row.hostGroupId) }}</template>
          </el-table-column>
          <el-table-column label="权限" width="120">
            <template #default="{ row }">
              <el-tag :type="permTagType(row.permission)" size="small">{{ permLabel(row.permission) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180" />
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-wrap">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            @size-change="fetchData"
            @current-change="fetchData"
          />
        </div>
      </div>

      <!-- Create dialog -->
      <el-dialog v-model="createDialogVisible" title="授予权限" width="500px" destroy-on-close>
        <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="100px">
          <el-form-item label="用户" prop="userId">
            <el-select v-model="createForm.userId" placeholder="选择用户" filterable style="width: 100%">
              <el-option v-for="u in userList" :key="u.id" :label="u.username" :value="u.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="分组" prop="hostGroupId">
            <el-tree-select
              v-model="createForm.hostGroupId"
              :data="treeData"
              :props="{ label: 'name', children: 'children', value: 'id' }"
              placeholder="选择分组"
              check-strictly
              filterable
              style="width: 100%"
              @change="handleGroupSelectChange"
            />
          </el-form-item>
          <el-form-item label="权限" prop="permissions">
            <el-checkbox-group v-model="createForm.permissions">
              <el-checkbox label="view">查看 (view)</el-checkbox>
              <el-checkbox label="terminal">终端 (terminal)</el-checkbox>
              <el-checkbox label="admin">管理 (admin)</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item v-if="groupHostCount >= 0" label="">
            <span class="host-count-tip">此权限将影响 {{ groupHostCount }} 台主机</span>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleCreateSubmit" :loading="submitting">确定</el-button>
        </template>
      </el-dialog>

      <!-- Edit dialog -->
      <el-dialog v-model="editDialogVisible" title="编辑权限" width="400px" destroy-on-close>
        <el-form :model="editForm" :rules="editRules" ref="editFormRef" label-width="80px">
          <el-form-item label="权限" prop="permission">
            <el-select v-model="editForm.permission" style="width: 100%">
              <el-option label="查看 (view)" value="view" />
              <el-option label="终端 (terminal)" value="terminal" />
              <el-option label="管理 (admin)" value="admin" />
            </el-select>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleEditSubmit" :loading="submitting">确定</el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPermissionList, createPermission, updatePermission, deletePermission, getGroupHostCount } from '@/api/cmdb/permission'
import { getGroupTree } from '@/api/cmdb/group'
import { getUserList } from '@/api/system'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const filterUserId = ref('')
const filterPermission = ref('')
const selectedGroupId = ref(0)

const treeData = ref([])
const userList = ref([])
const groupFlatMap = ref({})

const createDialogVisible = ref(false)
const editDialogVisible = ref(false)
const submitting = ref(false)
const createFormRef = ref()
const editFormRef = ref()
const groupHostCount = ref(-1)

const createForm = reactive({
  userId: '',
  hostGroupId: '',
  permissions: []
})

const editForm = reactive({
  id: 0,
  permission: ''
})

const createRules = {
  userId: [{ required: true, message: '请选择用户', trigger: 'change' }],
  hostGroupId: [{ required: true, message: '请选择分组', trigger: 'change' }],
  permissions: [{ required: true, type: 'array', min: 1, message: '请至少选择一个权限', trigger: 'change' }]
}

const editRules = {
  permission: [{ required: true, message: '请选择权限', trigger: 'change' }]
}

const fetchTree = async () => {
  try {
    const res = await getGroupTree()
    treeData.value = res.data || []
    flattenTree(treeData.value)
  } catch (e) {
    console.error('fetch tree:', e)
  }
}

const flattenTree = (nodes, path = '') => {
  for (const node of nodes) {
    const currentPath = path ? `${path} / ${node.name}` : node.name
    groupFlatMap.value[node.id] = { ...node, path: currentPath }
    if (node.children && node.children.length) {
      flattenTree(node.children, currentPath)
    }
  }
}

const fetchUsers = async () => {
  try {
    const res = await getUserList({ page: 1, pageSize: 200 })
    userList.value = res.data?.list || res.data || []
  } catch (e) {
    console.error('fetch users:', e)
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (filterUserId.value) params.userId = filterUserId.value
    if (filterPermission.value) params.permission = filterPermission.value
    if (selectedGroupId.value) params.hostGroupId = selectedGroupId.value

    const res = await getPermissionList(params)
    tableData.value = res.data || []
    total.value = res.total || 0
  } catch (e) {
    ElMessage.error('获取权限列表失败')
  } finally {
    loading.value = false
  }
}

const handleNodeClick = (data) => {
  selectedGroupId.value = data.id
  page.value = 1
  fetchData()
}

const getUsername = (userId) => {
  const u = userList.value.find(u => u.id === userId)
  return u ? u.username : `用户${userId}`
}

const getGroupName = (groupId) => {
  const g = groupFlatMap.value[groupId]
  return g ? g.path : `分组${groupId}`
}

const permLabel = (p) => {
  const map = { view: '查看', terminal: '终端', admin: '管理' }
  return map[p] || p
}

const permTagType = (p) => {
  const map = { view: 'info', terminal: 'warning', admin: 'danger' }
  return map[p] || 'info'
}

const showCreateDialog = () => {
  createForm.userId = ''
  createForm.hostGroupId = selectedGroupId.value || ''
  createForm.permissions = []
  groupHostCount.value = -1
  createDialogVisible.value = true
}

const handleGroupSelectChange = async (val) => {
  if (!val) {
    groupHostCount.value = -1
    return
  }
  try {
    const res = await getGroupHostCount({ groupId: val })
    groupHostCount.value = res.data?.hostCount ?? -1
  } catch {
    groupHostCount.value = -1
  }
}

const handleCreateSubmit = async () => {
  try {
    await createFormRef.value.validate()
  } catch { return }

  submitting.value = true
  try {
    await createPermission({
      userId: createForm.userId,
      hostGroupId: createForm.hostGroupId,
      permissions: createForm.permissions
    })
    ElMessage.success('授权成功')
    createDialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '授权失败')
  } finally {
    submitting.value = false
  }
}

const handleEdit = (row) => {
  editForm.id = row.id
  editForm.permission = row.permission
  editDialogVisible.value = true
}

const handleEditSubmit = async () => {
  try {
    await editFormRef.value.validate()
  } catch { return }

  submitting.value = true
  try {
    await updatePermission({
      id: editForm.id,
      permission: editForm.permission
    })
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '更新失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该权限规则？', '确认', { type: 'warning' })
  } catch { return }

  try {
    await deletePermission({ id: row.id })
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '删除失败')
  }
}

onMounted(() => {
  fetchTree()
  fetchUsers()
  fetchData()
})
</script>

<style scoped>
.permission-page {
  display: flex;
  gap: 16px;
  height: calc(100vh - 120px);
}
.left-panel {
  width: 260px;
  min-width: 260px;
  background: #fff;
  border-radius: 4px;
  padding: 16px;
  overflow-y: auto;
}
.panel-header h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 500;
}
.tree-wrap {
  margin-top: 8px;
}
.right-panel {
  flex: 1;
  min-width: 0;
}
.page-container {
  background: #fff;
  border-radius: 4px;
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.host-count-tip {
  color: #909399;
  font-size: 13px;
}
</style>
```

- [ ] **Step 2: Build frontend**

Run: `cd frontend && npm run build`
Expected: build succeeds

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/Cmdb/PermissionList.vue
git commit -m "feat(cmdb): add PermissionList page with group tree and CRUD"
```

---

### Task 7: API Verification against Real Database

**Files:** None (verification only)

- [ ] **Step 1: Start backend**

Run: `cd backend && DEVOPS_SERVER_PORT=8001 go run cmd/server/main.go`
Wait for "服务启动" log.

- [ ] **Step 2: Start frontend dev server**

Run: `cd frontend && npm run dev`
Wait for "ready" message.

- [ ] **Step 3: Login and get token**

```bash
TOKEN=$(curl -s http://localhost:8001/api/v1/user/login -H 'Content-Type: application/json' -d '{"tenantCode":"default","username":"admin","password":"admin@2026"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "Token: $TOKEN"
```

- [ ] **Step 4: Test permission list (empty)**

```bash
curl -s http://localhost:8001/api/v1/cmdb/permission/list -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```
Expected: `{"code":200,"data":[],"total":0,...}`

- [ ] **Step 5: Test create permission**

```bash
curl -s http://localhost:8001/api/v1/cmdb/permission/create -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"userId":1,"hostGroupId":1,"permissions":["view","terminal"]}' | python3 -m json.tool
```
Expected: `{"code":200,"message":"授权成功","data":[...]}` with 2 permission records created.

- [ ] **Step 6: Test permission list (with data)**

```bash
curl -s "http://localhost:8001/api/v1/cmdb/permission/list?page=1&pageSize=10" -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```
Expected: `total: 2`, two records visible.

- [ ] **Step 7: Test update permission**

Use the ID from step 5 response:
```bash
curl -s http://localhost:8001/api/v1/cmdb/permission/update -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"id":<ID>,"permission":"admin"}' | python3 -m json.tool
```
Expected: `{"code":200,"message":"更新成功",...}`

- [ ] **Step 8: Test my-hosts**

```bash
curl -s http://localhost:8001/api/v1/cmdb/permission/my-hosts -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```
Expected: list of hosts with permission levels.

- [ ] **Step 9: Test check**

Use a host ID from step 8:
```bash
curl -s "http://localhost:8001/api/v1/cmdb/permission/check?hostId=<HOST_ID>&action=view" -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```
Expected: `{"data":{"allowed":true,"permission":"admin"}}`

- [ ] **Step 10: Test delete permission**

Use the IDs from step 5:
```bash
curl -s http://localhost:8001/api/v1/cmdb/permission/delete -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"id":<ID>}' | python3 -m json.tool
```
Expected: `{"code":200,"message":"删除成功"}`

- [ ] **Step 11: Test frontend page**

Open browser at `http://localhost:3002/cmdb/permissions` (or whatever port the dev server chose). Verify:
- Left panel shows group tree
- Right panel shows permission table
- Create dialog with user/group/permission selectors
- Edit and delete work
- Group tree click filters the table

- [ ] **Step 12: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix(cmdb): address permission verification findings"
```
