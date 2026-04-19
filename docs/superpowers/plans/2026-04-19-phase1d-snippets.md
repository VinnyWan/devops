# Command Snippet Library Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow users to save, share, and quick-paste common shell commands from a library accessible in the terminal view.

**Architecture:** Standard CRUD — new model, repository, service, API for command snippets. Frontend adds a collapsible snippet panel alongside the terminal.

**Tech Stack:** Go/Gin/GORM (backend), Vue 3 + Element Plus (frontend)

---

## File Structure

| Action | Path | Responsibility |
|--------|------|---------------|
| Create | `backend/internal/modules/cmdb/model/snippet.go` | Snippet model |
| Create | `backend/internal/modules/cmdb/repository/snippet.go` | DB queries |
| Create | `backend/internal/modules/cmdb/service/snippet.go` | Business logic |
| Create | `backend/internal/modules/cmdb/api/snippet.go` | HTTP handlers |
| Modify | `backend/internal/modules/cmdb/api/common.go` | Add snippet service singleton |
| Modify | `backend/routers/v1/cmdb.go` | Register snippet routes |
| Create | `frontend/src/api/cmdb/snippet.js` | Frontend API client |
| Create | `frontend/src/components/Cmdb/SnippetPanel.vue` | Collapsible snippet panel |
| Modify | `frontend/src/components/Cmdb/MultiTabTerminal.vue` | Integrate snippet panel |

---

### Task 1: Snippet Backend — Model, Repository, Service

**Files:**
- Create: `backend/internal/modules/cmdb/model/snippet.go`
- Create: `backend/internal/modules/cmdb/repository/snippet.go`
- Create: `backend/internal/modules/cmdb/service/snippet.go`

- [ ] **Step 1: Create snippet model**

Create `backend/internal/modules/cmdb/model/snippet.go`:

```go
package model

import (
	"time"

	"gorm.io/gorm"
)

// CommandSnippet 命令片段
type CommandSnippet struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	TenantID   uint           `gorm:"not null;index:idx_cmdb_snippets_tenant" json:"tenantId"`
	UserID     uint           `gorm:"not null;index:idx_cmdb_snippets_user" json:"userId"`
	Name       string         `gorm:"size:100;not null" json:"name"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Tags       string         `gorm:"size:500" json:"tags"`
	Visibility string         `gorm:"size:20;not null;default:'personal'" json:"visibility"` // personal, team, public
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
```

- [ ] **Step 2: Create snippet repository**

Create `backend/internal/modules/cmdb/repository/snippet.go`:

```go
package repository

import (
	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
)

type SnippetRepo struct {
	db *gorm.DB
}

func NewSnippetRepo(db *gorm.DB) *SnippetRepo {
	return &SnippetRepo{db: db}
}

func (r *SnippetRepo) ListInTenant(tenantID uint, userID uint, page, pageSize int) ([]model.CommandSnippet, int64, error) {
	var list []model.CommandSnippet
	var total int64

	query := r.db.Where("tenant_id = ? AND (visibility = 'public' OR visibility = 'team' OR user_id = ?)", tenantID, userID)
	query.Model(&model.CommandSnippet{}).Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *SnippetRepo) GetByID(id uint) (*model.CommandSnippet, error) {
	var s model.CommandSnippet
	err := r.db.First(&s, id).Error
	return &s, err
}

func (r *SnippetRepo) Create(snippet *model.CommandSnippet) error {
	return r.db.Create(snippet).Error
}

func (r *SnippetRepo) Update(snippet *model.CommandSnippet) error {
	return r.db.Save(snippet).Error
}

func (r *SnippetRepo) Delete(id uint) error {
	return r.db.Delete(&model.CommandSnippet{}, id).Error
}

func (r *SnippetRepo) Search(tenantID uint, userID uint, keyword string, limit int) ([]model.CommandSnippet, error) {
	var list []model.CommandSnippet
	pattern := "%" + keyword + "%"
	err := r.db.Where(
		"tenant_id = ? AND (visibility = 'public' OR visibility = 'team' OR user_id = ?)", tenantID, userID,
	).Where(
		"name LIKE ? OR tags LIKE ? OR content LIKE ?", pattern, pattern, pattern,
	).Limit(limit).Find(&list).Error
	return list, err
}
```

- [ ] **Step 3: Create snippet service**

Create `backend/internal/modules/cmdb/service/snippet.go`:

```go
package service

import (
	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
)

type SnippetService struct {
	repo *repository.SnippetRepo
}

func NewSnippetService(repo *repository.SnippetRepo) *SnippetService {
	return &SnippetService{repo: repo}
}

type SnippetListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

func (s *SnippetService) List(tenantID, userID uint, page, pageSize int) ([]model.CommandSnippet, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.ListInTenant(tenantID, userID, page, pageSize)
}

func (s *SnippetService) GetByID(id uint) (*model.CommandSnippet, error) {
	return s.repo.GetByID(id)
}

type SnippetCreateRequest struct {
	Name       string `json:"name" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Tags       string `json:"tags"`
	Visibility string `json:"visibility"`
}

func (s *SnippetService) Create(tenantID, userID uint, req SnippetCreateRequest) (*model.CommandSnippet, error) {
	visibility := req.Visibility
	if visibility == "" {
		visibility = "personal"
	}
	snippet := &model.CommandSnippet{
		TenantID:   tenantID,
		UserID:     userID,
		Name:       req.Name,
		Content:    req.Content,
		Tags:       req.Tags,
		Visibility: visibility,
	}
	if err := s.repo.Create(snippet); err != nil {
		return nil, err
	}
	return snippet, nil
}

type SnippetUpdateRequest struct {
	ID         uint   `json:"id" binding:"required"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	Tags       string `json:"tags"`
	Visibility string `json:"visibility"`
}

func (s *SnippetService) Update(tenantID, userID uint, req SnippetUpdateRequest) (*model.CommandSnippet, error) {
	snippet, err := s.repo.GetByID(req.ID)
	if err != nil {
		return nil, err
	}
	if snippet.TenantID != tenantID || (snippet.UserID != userID && snippet.Visibility == "personal") {
		return nil, ErrPermissionDenied
	}
	if req.Name != "" {
		snippet.Name = req.Name
	}
	if req.Content != "" {
		snippet.Content = req.Content
	}
	snippet.Tags = req.Tags
	if req.Visibility != "" {
		snippet.Visibility = req.Visibility
	}
	if err := s.repo.Update(snippet); err != nil {
		return nil, err
	}
	return snippet, nil
}

func (s *SnippetService) Delete(tenantID, userID uint, id uint) error {
	snippet, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if snippet.TenantID != tenantID || (snippet.UserID != userID && snippet.Visibility == "personal") {
		return ErrPermissionDenied
	}
	return s.repo.Delete(id)
}

func (s *SnippetService) Search(tenantID, userID uint, keyword string) ([]model.CommandSnippet, error) {
	return s.repo.Search(tenantID, userID, keyword, 20)
}

var ErrPermissionDenied = &permissionError{}

type permissionError struct{}

func (e *permissionError) Error() string { return "权限不足" }
```

- [ ] **Step 4: Verify compilation**

Run: `cd backend && go build ./internal/modules/cmdb/...`
Expected: compiles

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/modules/cmdb/model/snippet.go internal/modules/cmdb/repository/snippet.go internal/modules/cmdb/service/snippet.go
git commit -m "feat(cmdb): add command snippet model, repository, and service"
```

---

### Task 2: Snippet Backend — API & Routes

**Files:**
- Create: `backend/internal/modules/cmdb/api/snippet.go`
- Modify: `backend/internal/modules/cmdb/api/common.go`
- Modify: `backend/routers/v1/cmdb.go`

- [ ] **Step 1: Create snippet API handlers**

Create `backend/internal/modules/cmdb/api/snippet.go`:

```go
package api

import (
	"strconv"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
)

func SnippetList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")

	svc := getSnippetService()
	if svc == nil {
		c.JSON(500, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	if keyword != "" {
		list, err := svc.Search(tenantID, userID, keyword)
		if err != nil {
			c.JSON(500, gin.H{"code": 500, "message": err.Error()})
			return
		}
		c.JSON(200, gin.H{"code": 200, "data": list})
		return
	}

	list, total, err := svc.List(tenantID, userID, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": list, "total": total, "page": page, "pageSize": pageSize})
}

func SnippetCreate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	var req service.SnippetCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	svc := getSnippetService()
	snippet, err := svc.Create(tenantID, userID, req)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": snippet})
}

func SnippetUpdate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	var req service.SnippetUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "请求参数错误: " + err.Error()})
		return
	}

	svc := getSnippetService()
	snippet, err := svc.Update(tenantID, userID, req)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": snippet})
}

func SnippetDelete(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		c.JSON(400, gin.H{"code": 400, "message": "缺少 id"})
		return
	}

	svc := getSnippetService()
	if err := svc.Delete(tenantID, userID, uint(id)); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "删除成功"})
}
```

- [ ] **Step 2: Add singleton to common.go**

Add to var block: `snippetSvcInstance *service.SnippetService`

Add to SetDB: `snippetSvcInstance = nil`

Add getter:
```go
func getSnippetService() *service.SnippetService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if snippetSvcInstance != nil {
		return snippetSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	snippetSvcInstance = service.NewSnippetService(repository.NewSnippetRepo(cmdbDB))
	return snippetSvcInstance
}
```

- [ ] **Step 3: Register routes**

Add to `backend/routers/v1/cmdb.go`:

```go
snippetListPerm := middleware.RequirePermission("cmdb:terminal", "connect")
snippetCreatePerm := middleware.RequirePermission("cmdb:terminal", "connect")
snippetUpdatePerm := middleware.RequirePermission("cmdb:terminal", "connect")
snippetDeletePerm := middleware.RequirePermission("cmdb:terminal", "connect")
```

Add routes:
```go
// 命令片段
g.GET("/snippet/list", snippetListPerm, api.SnippetList)
g.POST("/snippet/create", snippetCreatePerm, middleware.SetAuditOperation("创建命令片段"), api.SnippetCreate)
g.POST("/snippet/update", snippetUpdatePerm, middleware.SetAuditOperation("更新命令片段"), api.SnippetUpdate)
g.POST("/snippet/delete", snippetDeletePerm, middleware.SetAuditOperation("删除命令片段"), api.SnippetDelete)
```

- [ ] **Step 4: Verify and commit**

Run: `cd backend && go build ./...`

```bash
cd backend
git add internal/modules/cmdb/api/snippet.go internal/modules/cmdb/api/common.go routers/v1/cmdb.go
git commit -m "feat(cmdb): add command snippet CRUD API and routes"
```

---

### Task 3: Snippet Frontend

**Files:**
- Create: `frontend/src/api/cmdb/snippet.js`
- Create: `frontend/src/components/Cmdb/SnippetPanel.vue`
- Modify: `frontend/src/components/Cmdb/MultiTabTerminal.vue`

- [ ] **Step 1: Create snippet API client**

Create `frontend/src/api/cmdb/snippet.js`:

```js
import request from '../request'

export const getSnippetList = (params) => request.get('/cmdb/snippet/list', { params })
export const searchSnippets = (keyword) => request.get('/cmdb/snippet/list', { params: { keyword } })
export const createSnippet = (data) => request.post('/cmdb/snippet/create', data)
export const updateSnippet = (data) => request.post('/cmdb/snippet/update', data)
export const deleteSnippet = (params) => request.post('/cmdb/snippet/delete', null, { params })
```

- [ ] **Step 2: Create SnippetPanel component**

Create `frontend/src/components/Cmdb/SnippetPanel.vue`:

```vue
<template>
  <div class="snippet-panel" :class="{ collapsed: !expanded }">
    <div class="panel-toggle" @click="expanded = !expanded">
      <el-icon><component :is="expanded ? 'ArrowRight' : 'ArrowLeft'" /></el-icon>
      <span v-if="!expanded">片段</span>
    </div>
    <div v-if="expanded" class="panel-content">
      <div class="panel-header">
        <el-input v-model="keyword" placeholder="搜索..." size="small" clearable @input="handleSearch" style="margin-bottom: 8px;" />
        <el-button type="primary" size="small" @click="showCreate = true" style="width: 100%;">+ 新建</el-button>
      </div>
      <div class="snippet-list">
        <div v-for="s in snippets" :key="s.id" class="snippet-item" @click="insertSnippet(s)">
          <div class="snippet-name">{{ s.name }}</div>
          <div class="snippet-tags" v-if="s.tags">
            <el-tag v-for="tag in s.tags.split(',')" :key="tag" size="small" type="info" style="margin-right: 2px;">{{ tag.trim() }}</el-tag>
          </div>
          <pre class="snippet-preview">{{ s.content.substring(0, 80) }}{{ s.content.length > 80 ? '...' : '' }}</pre>
        </div>
        <el-empty v-if="!snippets.length" description="暂无片段" :image-size="40" />
      </div>
    </div>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreate" title="新建命令片段" width="400px" append-to-body>
      <el-form :model="form" label-width="60px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="命令"><el-input v-model="form.content" type="textarea" :rows="4" style="font-family: monospace;" /></el-form-item>
        <el-form-item label="标签"><el-input v-model="form.tags" placeholder="逗号分隔" /></el-form-item>
        <el-form-item label="可见性">
          <el-select v-model="form.visibility">
            <el-option label="仅自己" value="personal" />
            <el-option label="团队" value="team" />
            <el-option label="公开" value="public" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ArrowRight, ArrowLeft } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getSnippetList, searchSnippets, createSnippet } from '@/api/cmdb/snippet'

const emit = defineEmits(['insert'])

const expanded = ref(false)
const snippets = ref([])
const keyword = ref('')
const showCreate = ref(false)
const form = ref({ name: '', content: '', tags: '', visibility: 'personal' })

const fetchSnippets = async () => {
  try {
    const res = await getSnippetList({ page: 1, pageSize: 50 })
    snippets.value = res.data || []
  } catch (e) { /* ignore */ }
}

const handleSearch = async () => {
  if (!keyword.value) {
    await fetchSnippets()
    return
  }
  try {
    const res = await searchSnippets(keyword.value)
    snippets.value = res.data || []
  } catch (e) { /* ignore */ }
}

const insertSnippet = (snippet) => {
  emit('insert', snippet.content)
}

const handleCreate = async () => {
  if (!form.value.name || !form.value.content) {
    ElMessage.warning('名称和命令不能为空')
    return
  }
  try {
    await createSnippet(form.value)
    ElMessage.success('创建成功')
    showCreate.value = false
    form.value = { name: '', content: '', tags: '', visibility: 'personal' }
    await fetchSnippets()
  } catch (e) {
    ElMessage.error('创建失败')
  }
}

onMounted(fetchSnippets)
</script>

<style scoped>
.snippet-panel {
  display: flex;
  height: 100%;
  border-left: 1px solid #3c3c3c;
  background: #252526;
}

.snippet-panel.collapsed {
  width: 32px;
}

.panel-toggle {
  width: 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 12px;
  cursor: pointer;
  color: #888;
  font-size: 11px;
  writing-mode: vertical-rl;
}

.panel-toggle:hover { color: #fff; }

.panel-content {
  width: 220px;
  padding: 8px;
  overflow-y: auto;
}

.snippet-list { max-height: calc(100% - 80px); overflow-y: auto; }

.snippet-item {
  padding: 8px;
  border-radius: 4px;
  cursor: pointer;
  margin-bottom: 4px;
  background: #1e1e1e;
  transition: background 0.15s;
}
.snippet-item:hover { background: #2a2d2e; }

.snippet-name { font-size: 12px; font-weight: 600; color: #e0e0e0; margin-bottom: 2px; }
.snippet-tags { margin-bottom: 4px; }
.snippet-preview { margin: 0; font-size: 10px; color: #888; font-family: monospace; white-space: pre-wrap; }
</style>
```

- [ ] **Step 3: Integrate into MultiTabTerminal**

In `frontend/src/components/Cmdb/MultiTabTerminal.vue`, add the snippet panel:

1. Add import:
```js
import SnippetPanel from '@/components/Cmdb/SnippetPanel.vue'
```

2. Add insert handler:
```js
const handleSnippetInsert = (content) => {
  const ref = terminalRefs.value[activeIndex.value]
  if (ref && typeof ref.write === 'function') {
    ref.write(content)
  }
}
```

3. Add SnippetPanel to the template, after the `.terminal-panels` div:
```html
<SnippetPanel @insert="handleSnippetInsert" />
```

4. Update the flex layout — wrap terminal-panels and snippet-panel in a flex row:
```html
<div style="display: flex; flex: 1; overflow: hidden;">
  <div class="terminal-panels" style="flex: 1;">
    <!-- existing terminal panels -->
  </div>
  <SnippetPanel @insert="handleSnippetInsert" />
</div>
```

- [ ] **Step 4: Verify and commit**

Run: `cd frontend && npm run dev` — verify snippet panel appears alongside terminal

```bash
cd frontend
git add src/api/cmdb/snippet.js src/components/Cmdb/SnippetPanel.vue src/components/Cmdb/MultiTabTerminal.vue
git commit -m "feat(cmdb): add command snippet panel integrated with multi-tab terminal"
```

---

## Self-Review Checklist

1. **Spec coverage:** Save and share commands — Task 1+2. Quick-paste into terminal — Task 3. Search/filter — Task 1+3. Visibility (personal/team/public) — Task 1. All covered.

2. **Placeholder scan:** No TBD/TODO. All code complete.

3. **Type consistency:** `CommandSnippet` model fields match service DTOs and API responses. Frontend `snippets` array uses matching JSON field names from model tags.
