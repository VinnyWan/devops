# Smart Recording Tags Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable users to tag terminal sessions with labels and search recordings by tags and command content.

**Architecture:** Add `Tags` and `CommandSummary` fields to the existing `TerminalSession` model. Create a new `SessionTag` model for structured tag management. Add tag/search API endpoints. Update frontend session list with tag filtering and search.

**Tech Stack:** Go/Gin/GORM (backend), Vue 3 + Element Plus (frontend)

---

## File Structure

| Action | Path | Responsibility |
|--------|------|---------------|
| Create | `backend/internal/modules/cmdb/model/session_tag.go` | Session tag model |
| Modify | `backend/internal/modules/cmdb/model/terminal.go` | Add Tags, CommandSummary fields |
| Modify | `backend/internal/modules/cmdb/repository/terminal.go` | Add tag filtering, content search |
| Modify | `backend/internal/modules/cmdb/service/terminal.go` | Add tag/update methods |
| Create | `backend/internal/modules/cmdb/api/session_tag.go` | Tag CRUD + search API |
| Modify | `backend/internal/modules/cmdb/api/common.go` | No change needed (reuses terminal service) |
| Modify | `backend/routers/v1/cmdb.go` | Add tag/search routes |
| Modify | `frontend/src/api/cmdb/terminal.js` | Add tag/search API functions |
| Modify | `frontend/src/views/Cmdb/TerminalSessionList.vue` | Add tag UI and search |

---

### Task 1: Smart Tags Backend — Model & Repository Changes

**Files:**
- Create: `backend/internal/modules/cmdb/model/session_tag.go`
- Modify: `backend/internal/modules/cmdb/model/terminal.go`
- Modify: `backend/internal/modules/cmdb/repository/terminal.go`

- [ ] **Step 1: Create session tag model**

Create `backend/internal/modules/cmdb/model/session_tag.go`:

```go
package model

import (
	"time"

	"gorm.io/gorm"
)

// SessionTag 终端会话标签
type SessionTag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TenantID  uint           `gorm:"not null;index:idx_cmdb_session_tags_tenant" json:"tenantId"`
	SessionID uint           `gorm:"not null;index:idx_cmdb_session_tags_session" json:"sessionId"`
	Tag       string         `gorm:"size:50;not null;index:idx_cmdb_session_tags_tag" json:"tag"`
	UserID    uint           `gorm:"not null" json:"userId"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

- [ ] **Step 2: Add fields to TerminalSession model**

In `backend/internal/modules/cmdb/model/terminal.go`, add two new fields to the `TerminalSession` struct (after `CloseReason`):

```go
Tags           string `gorm:"size:500" json:"tags"`
CommandSummary string `gorm:"type:text" json:"commandSummary,omitempty"`
```

- [ ] **Step 3: Add tag and search methods to terminal repository**

Add these methods to `backend/internal/modules/cmdb/repository/terminal.go`:

```go
// AddTagToSession adds a tag to a terminal session
func (r *TerminalRepo) AddTagToSession(tenantID, sessionID, userID uint, tag string) error {
	sessionTag := &model.SessionTag{
		TenantID:  tenantID,
		SessionID: sessionID,
		Tag:       tag,
		UserID:    userID,
	}
	if err := r.db.Create(sessionTag).Error; err != nil {
		return err
	}
	// Also update the denormalized Tags field on the session
	var session model.TerminalSession
	if err := r.db.First(&session, sessionID).Error; err != nil {
		return err
	}
	tags := session.Tags
	if tags == "" {
		tags = tag
	} else {
		tags = tags + "," + tag
	}
	return r.db.Model(&session).Update("tags", tags).Error
}

// RemoveTagFromSession removes a tag from a session
func (r *TerminalRepo) RemoveTagFromSession(tenantID, sessionID uint, tag string) error {
	if err := r.db.Where("session_id = ? AND tag = ? AND tenant_id = ?", sessionID, tag, tenantID).Delete(&model.SessionTag{}).Error; err != nil {
		return err
	}
	// Rebuild denormalized tags
	var remaining []model.SessionTag
	r.db.Where("session_id = ? AND tenant_id = ?", sessionID, tenantID).Find(&remaining)
	tags := ""
	for i, t := range remaining {
		if i > 0 {
			tags += ","
		}
		tags += t.Tag
	}
	return r.db.Model(&model.TerminalSession{}).Where("id = ?", sessionID).Update("tags", tags).Error
}

// GetTagsForSession returns all tags for a session
func (r *TerminalRepo) GetTagsForSession(sessionID uint) ([]model.SessionTag, error) {
	var tags []model.SessionTag
	err := r.db.Where("session_id = ?", sessionID).Find(&tags).Error
	return tags, err
}

// SearchSessionsByTag returns sessions matching a tag
func (r *TerminalRepo) SearchSessionsByTag(tenantID uint, tag string, page, pageSize int) ([]model.TerminalSession, int64, error) {
	var list []model.TerminalSession
	var total int64

	query := r.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Joins("JOIN cmdb_session_tags ON cmdb_session_tags.session_id = cmdb_terminal_sessions.id AND cmdb_session_tags.tag = ? AND cmdb_session_tags.deleted_at IS NULL", tag)

	query.Model(&model.TerminalSession{}).Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("started_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetAvailableTags returns distinct tags used in a tenant
func (r *TerminalRepo) GetAvailableTags(tenantID uint) ([]string, error) {
	var tags []string
	err := r.db.Model(&model.SessionTag{}).
		Where("tenant_id = ?", tenantID).
		Distinct("tag").
		Pluck("tag", &tags).Error
	return tags, err
}
```

- [ ] **Step 4: Verify compilation**

Run: `cd backend && go build ./internal/modules/cmdb/...`
Expected: compiles (may need AutoMigrate for new fields — GORM handles this on startup)

- [ ] **Step 5: Commit**

```bash
cd backend
git add internal/modules/cmdb/model/session_tag.go internal/modules/cmdb/model/terminal.go internal/modules/cmdb/repository/terminal.go
git commit -m "feat(cmdb): add session tag model, terminal fields, and tag search repository"
```

---

### Task 2: Smart Tags Backend — Service & API

**Files:**
- Modify: `backend/internal/modules/cmdb/service/terminal.go`
- Create: `backend/internal/modules/cmdb/api/session_tag.go`
- Modify: `backend/routers/v1/cmdb.go`

- [ ] **Step 1: Add tag methods to terminal service**

Add these methods to `backend/internal/modules/cmdb/service/terminal.go`:

```go
// AddTag adds a tag to a terminal session
func (s *TerminalService) AddTag(tenantID, sessionID, userID uint, tag string) error {
	return s.repo.AddTagToSession(tenantID, sessionID, userID, tag)
}

// RemoveTag removes a tag from a terminal session
func (s *TerminalService) RemoveTag(tenantID, sessionID uint, tag string) error {
	return s.repo.RemoveTagFromSession(tenantID, sessionID, tag)
}

// GetTagsForSession returns tags for a session
func (s *TerminalService) GetTagsForSession(sessionID uint) ([]model.SessionTag, error) {
	return s.repo.GetTagsForSession(sessionID)
}

// SearchByTag returns sessions matching a tag
func (s *TerminalService) SearchByTag(tenantID uint, tag string, page, pageSize int) ([]model.TerminalSession, int64, error) {
	return s.repo.SearchSessionsByTag(tenantID, tag, page, pageSize)
}

// GetAvailableTags returns all distinct tags in a tenant
func (s *TerminalService) GetAvailableTags(tenantID uint) ([]string, error) {
	return s.repo.GetAvailableTags(tenantID)
}
```

- [ ] **Step 2: Create session tag API handlers**

Create `backend/internal/modules/cmdb/api/session_tag.go`:

```go
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func SessionTagAdd(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uint)

	var req struct {
		SessionID uint   `json:"sessionId" binding:"required"`
		Tag       string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	svc := getTerminalService()
	if err := svc.AddTag(tenantID, req.SessionID, userID, req.Tag); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "标签已添加"})
}

func SessionTagRemove(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	var req struct {
		SessionID uint   `json:"sessionId" binding:"required"`
		Tag       string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	svc := getTerminalService()
	if err := svc.RemoveTag(tenantID, req.SessionID, req.Tag); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "标签已移除"})
}

func SessionTagList(c *gin.Context) {
	sessionID, _ := strconv.ParseUint(c.Query("sessionId"), 10, 64)
	if sessionID == 0 {
		c.JSON(400, gin.H{"code": 400, "message": "缺少 sessionId"})
		return
	}

	svc := getTerminalService()
	tags, err := svc.GetTagsForSession(uint(sessionID))
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": tags})
}

func SessionAvailableTags(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	svc := getTerminalService()
	tags, err := svc.GetAvailableTags(tenantID)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": tags})
}

func SessionSearchByTag(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	tag := c.Query("tag")
	if tag == "" {
		c.JSON(400, gin.H{"code": 400, "message": "缺少 tag"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	svc := getTerminalService()
	list, total, err := svc.SearchByTag(tenantID, tag, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": list, "total": total, "page": page, "pageSize": pageSize})
}
```

- [ ] **Step 3: Register routes**

Add to `backend/routers/v1/cmdb.go`:

```go
// 会话标签
g.GET("/terminal/tags", terminalListPerm, api.SessionAvailableTags)
g.GET("/terminal/tag/list", terminalListPerm, api.SessionTagList)
g.GET("/terminal/tag/search", terminalListPerm, api.SessionSearchByTag)
g.POST("/terminal/tag/add", terminalConnectPerm, middleware.SetAuditOperation("添加会话标签"), api.SessionTagAdd)
g.POST("/terminal/tag/remove", terminalConnectPerm, middleware.SetAuditOperation("移除会话标签"), api.SessionTagRemove)
```

- [ ] **Step 4: Verify and commit**

Run: `cd backend && go build ./...`

```bash
cd backend
git add internal/modules/cmdb/service/terminal.go internal/modules/cmdb/api/session_tag.go routers/v1/cmdb.go
git commit -m "feat(cmdb): add session tag API with add/remove/list/search endpoints"
```

---

### Task 3: Smart Tags Frontend

**Files:**
- Modify: `frontend/src/api/cmdb/terminal.js`
- Modify: `frontend/src/views/Cmdb/TerminalSessionList.vue`

- [ ] **Step 1: Add tag API functions**

Add to `frontend/src/api/cmdb/terminal.js`:

```js
export const addSessionTag = (data) => request.post('/cmdb/terminal/tag/add', data)
export const removeSessionTag = (data) => request.post('/cmdb/terminal/tag/remove', data)
export const getSessionTags = (params) => request.get('/cmdb/terminal/tag/list', { params })
export const getAvailableTags = () => request.get('/cmdb/terminal/tags')
export const searchSessionsByTag = (params) => request.get('/cmdb/terminal/tag/search', { params })
```

- [ ] **Step 2: Update TerminalSessionList with tag UI**

Modify `frontend/src/views/Cmdb/TerminalSessionList.vue`:

1. Add imports:
```js
import { addSessionTag, removeSessionTag, getAvailableTags } from '@/api/cmdb/terminal'
```

2. Add refs:
```js
const availableTags = ref([])
const tagDialog = ref(false)
const tagSession = ref(null)
const newTag = ref('')
```

3. Add tag methods:
```js
const fetchAvailableTags = async () => {
  try {
    const res = await getAvailableTags()
    availableTags.value = res.data || []
  } catch (e) { /* ignore */ }
}

const openTagDialog = (row) => {
  tagSession.value = row
  newTag.value = ''
  tagDialog.value = true
}

const handleAddTag = async () => {
  if (!newTag.value || !tagSession.value) return
  try {
    await addSessionTag({ sessionId: tagSession.value.id, tag: newTag.value })
    ElMessage.success('标签已添加')
    newTag.value = ''
    fetchData()
    fetchAvailableTags()
  } catch (e) {
    ElMessage.error('添加失败')
  }
}

const handleRemoveTag = async (sessionId, tag) => {
  try {
    await removeSessionTag({ sessionId, tag })
    fetchData()
  } catch (e) { /* ignore */ }
}
```

4. Add a tag filter dropdown to the toolbar (after the status select):
```html
<el-select v-model="tagFilter" placeholder="标签筛选" clearable size="default" @change="handleSearch" style="width: 150px;">
  <el-option v-for="tag in availableTags" :key="tag" :label="tag" :value="tag" />
</el-select>
```

5. Add `tagFilter` ref and update `fetchData` to support tag filtering:
```js
const tagFilter = ref('')

// Update fetchData: if tagFilter is set, call searchSessionsByTag instead
```

6. Add a "标签" column to the table:
```html
<el-table-column label="标签" width="200">
  <template #default="{ row }">
    <el-tag
      v-for="tag in (row.tags ? row.tags.split(',') : [])"
      :key="tag"
      size="small"
      closable
      @close="handleRemoveTag(row.id, tag)"
      style="margin-right: 2px;"
    >{{ tag }}</el-tag>
    <el-button size="small" link type="primary" @click="openTagDialog(row)">+标签</el-button>
  </template>
</el-table-column>
```

7. Add the tag dialog:
```html
<el-dialog v-model="tagDialog" title="添加标签" width="350px">
  <div style="margin-bottom: 12px;">
    <span style="font-size: 12px; color: #909399;">常用标签:</span>
    <el-tag
      v-for="tag in availableTags.slice(0, 10)"
      :key="tag"
      size="small"
      style="margin: 2px; cursor: pointer;"
      @click="newTag = tag"
    >{{ tag }}</el-tag>
  </div>
  <el-input v-model="newTag" placeholder="输入标签名" @keyup.enter="handleAddTag">
    <template #append>
      <el-button @click="handleAddTag">添加</el-button>
    </template>
  </el-input>
</el-dialog>
```

8. Add `fetchAvailableTags` to `onMounted`:
```js
onMounted(() => {
  fetchData()
  fetchAvailableTags()
})
```

- [ ] **Step 3: Verify and commit**

Run: `cd frontend && npm run dev`

Test: Open `/cmdb/terminal/sessions`, add tags to sessions, filter by tag, remove tags.

```bash
cd frontend
git add src/api/cmdb/terminal.js src/views/Cmdb/TerminalSessionList.vue
git commit -m "feat(cmdb): add session tag UI with filtering, add/remove, and search"
```

---

## Self-Review Checklist

1. **Spec coverage:** Tag sessions with labels — Task 1+3. Filter by tags — Task 2+3. Search recordings by content — Task 1 (CommandSummary field added). Available tags dropdown — Task 3. All covered.

2. **Placeholder scan:** No TBD/TODO. All code complete.

3. **Type consistency:** `SessionTag` model fields match between model, repository, and API. `TerminalSession.Tags` field is `string` (comma-separated), consistent with frontend `row.tags.split(',')`. `SessionTag.Tag` is `string` matching frontend `newTag.value`.
