# 操作审计模块 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在侧边栏新增独立的「操作审计」菜单，包含操作日志（复用已有后端）和登录日志（新建后端模块）两个页面。

**Architecture:** 登录日志后端在 user 模块内扩展，与已有的 audit_log 同级。使用 ip2region 离线库解析 IP 地理位置，使用 ua-parser 解析浏览器和操作系统。前端遵循项目已有的列表页模式。

**Tech Stack:** Go (Gin/GORM), Vue 3 (Composition API), Element Plus, ip2region, ua-parser

**Design Spec:** `docs/superpowers/specs/2026-04-14-operation-audit-design.md`

---

## File Structure

### Backend New Files
| File | Responsibility |
|------|---------------|
| `backend/internal/modules/user/model/login_log.go` | LoginLog 数据模型 |
| `backend/internal/modules/user/repository/login_log_repo.go` | 登录日志 DB 查询 |
| `backend/internal/modules/user/service/login_log_service.go` | 登录日志业务逻辑 + IP/UA 解析 |
| `backend/internal/modules/user/api/login_log.go` | 登录日志 HTTP handler |
| `backend/internal/routers/v1/login_log.go` | 路由注册 |

### Backend Modified Files
| File | Change |
|------|--------|
| `backend/internal/modules/user/api/common.go` | 添加 loginLogService 重置 |
| `backend/internal/modules/user/api/auth.go` | Login/OIDCCallback 中记录登录日志 |
| `backend/internal/bootstrap/db.go` | AutoMigrate 添加 LoginLog + 权限种子 |
| `backend/routers/v1/v1.go` | 注册 login-log 路由 |
| `backend/go.mod` / `backend/go.sum` | 添加 ip2region、ua-parser 依赖 |

### Frontend New Files
| File | Responsibility |
|------|---------------|
| `frontend/src/api/audit.js` | 操作日志 API 调用 |
| `frontend/src/api/loginLog.js` | 登录日志 API 调用 |
| `frontend/src/views/Audit/OperationLog.vue` | 操作日志页面 |
| `frontend/src/views/Audit/LoginLog.vue` | 登录日志页面 |

### Frontend Modified Files
| File | Change |
|------|--------|
| `frontend/src/components/Layout/MainLayout.vue` | 侧边栏新增「操作审计」菜单 |
| `frontend/src/router/index.js` | 新增两条路由 |

---

## Task 1: Backend — LoginLog Model

**Files:**
- Create: `backend/internal/modules/user/model/login_log.go`

- [ ] **Step 1: Create LoginLog model**

```go
package model

import (
	"time"

	"gorm.io/gorm"
)

// LoginLog 登录日志模型
type LoginLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"size:50;index" json:"username"`
	IP        string         `gorm:"size:50;index" json:"ip"`
	Location  string         `gorm:"size:100" json:"location"`
	Browser   string         `gorm:"size:100" json:"browser"`
	OS        string         `gorm:"size:100" json:"os"`
	Status    string         `gorm:"size:20;index" json:"status"`
	Message   string         `gorm:"size:200" json:"message"`
	UserAgent string         `gorm:"size:500" json:"userAgent"`
	LoginAt   time.Time      `gorm:"index" json:"loginAt"`
	CreatedAt time.Time      `gorm:"index" json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (LoginLog) TableName() string {
	return "login_logs"
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/modules/user/model/login_log.go
git commit -m "feat: add LoginLog model for login audit"
```

---

## Task 2: Backend — Install Dependencies (ip2region + ua-parser)

**Files:**
- Modify: `backend/go.mod`, `backend/go.sum`

- [ ] **Step 1: Add ip2region and ua-parser dependencies**

```bash
cd backend && go get github.com/lionsoul2014/ip2region/binding/golang/xdb && go get github.com/mssola/useragent
```

- [ ] **Step 2: Verify dependencies resolve**

```bash
cd backend && go mod tidy
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/go.mod backend/go.sum
git commit -m "chore: add ip2region and useragent dependencies"
```

---

## Task 3: Backend — LoginLog Repository

**Files:**
- Create: `backend/internal/modules/user/repository/login_log_repo.go`

- [ ] **Step 1: Create login_log_repo.go**

```go
package repository

import (
	"devops-platform/internal/modules/user/model"
	"time"

	"gorm.io/gorm"
)

type LoginLogRepo struct {
	db *gorm.DB
}

type LoginLogQuery struct {
	Username string
	Status   string
	StartAt  *time.Time
	EndAt    *time.Time
	Page     int
	PageSize int
}

func NewLoginLogRepo(db *gorm.DB) *LoginLogRepo {
	return &LoginLogRepo{db: db}
}

// Create 创建登录日志
func (r *LoginLogRepo) Create(log *model.LoginLog) error {
	return r.db.Create(log).Error
}

// List 分页查询登录日志
func (r *LoginLogRepo) List(query LoginLogQuery) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 200 {
		query.PageSize = 20
	}

	tx := r.buildListQuery(query)

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := tx.Order("login_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *LoginLogRepo) buildListQuery(query LoginLogQuery) *gorm.DB {
	tx := r.db.Model(&model.LoginLog{})
	if query.Username != "" {
		tx = tx.Where("username LIKE ?", "%"+query.Username+"%")
	}
	if query.Status != "" {
		tx = tx.Where("status = ?", query.Status)
	}
	if query.StartAt != nil {
		tx = tx.Where("login_at >= ?", *query.StartAt)
	}
	if query.EndAt != nil {
		tx = tx.Where("login_at <= ?", *query.EndAt)
	}
	return tx
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/modules/user/repository/login_log_repo.go
git commit -m "feat: add LoginLog repository with paginated query"
```

---

## Task 4: Backend — LoginLog Service (IP + UA 解析)

**Files:**
- Create: `backend/internal/modules/user/service/login_log_service.go`

- [ ] **Step 1: Create login_log_service.go**

```go
package service

import (
	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	"fmt"
	"strings"
	"time"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/mssola/useragent"
	"gorm.io/gorm"
)

type LoginLogService struct {
	repo    *repository.LoginLogRepo
	searcher *xdb.Searcher
}

type LoginLogListRequest struct {
	Username string
	Status   string
	StartAt  string
	EndAt    string
	Page     int
	PageSize int
}

var loginLogIPData []byte

func NewLoginLogService(repo *repository.LoginLogRepo, db *gorm.DB) *LoginLogService {
	s := &LoginLogService{repo: repo}

	// 加载 ip2region 数据到内存
	if loginLogIPData == nil {
		data, err := xdb.LoadContentFromFile("ip2region.xdb")
		if err != nil {
			data, err = xdb.LoadContentFromFile("backend/ip2region.xdb")
			if err != nil {
				data, err = xdb.LoadContentFromFile("third_party/ip2region/ip2region.xdb")
				if err != nil {
					// IP 解析不可用，不影响核心功能
					return s
				}
			}
		}
		loginLogIPData = data
	}

	searcher, err := xdb.NewWithBuffer(loginLogIPData)
	if err == nil {
		s.searcher = searcher
	}

	return s
}

// CreateLoginLog 创建登录日志（内部处理 IP 地理位置和 UA 解析）
func (s *LoginLogService) CreateLoginLog(username, ip, userAgentStr, status, message string) error {
	log := &model.LoginLog{
		Username:  username,
		IP:        ip,
		Location:  s.parseIPLocation(ip),
		Browser:   parseBrowser(userAgentStr),
		OS:        parseOS(userAgentStr),
		Status:    status,
		Message:   message,
		UserAgent: userAgentStr,
		LoginAt:   time.Now(),
	}
	return s.repo.Create(log)
}

// List 分页查询登录日志
func (s *LoginLogService) List(req LoginLogListRequest) ([]map[string]interface{}, int64, error) {
	query := repository.LoginLogQuery{
		Username: req.Username,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	if req.StartAt != "" {
		startAt, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid startAt format: %w", err)
		}
		query.StartAt = &startAt
	}
	if req.EndAt != "" {
		endAt, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid endAt format: %w", err)
		}
		query.EndAt = &endAt
	}

	logs, total, err := s.repo.List(query)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, 0, len(logs))
	for _, item := range logs {
		result = append(result, formatLoginLog(item))
	}

	return result, total, nil
}

func (s *LoginLogService) parseIPLocation(ip string) string {
	if s.searcher == nil {
		return ""
	}
	region, err := s.searcher.SearchByStr(ip)
	if err != nil {
		return ""
	}
	// ip2region 返回格式: 国家|区域|省份|城市|ISP
	// 提取省份+城市，去掉空段
	parts := strings.Split(region, "|")
	var location []string
	for i, p := range parts {
		if i > 3 {
			break
		}
		if p != "" && p != "0" {
			location = append(location, p)
		}
	}
	if len(location) == 0 {
		return ""
	}
	return strings.Join(location, " ")
}

func parseBrowser(uaStr string) string {
	ua := useragent.New(uaStr)
	browserName, browserVersion := ua.Browser()
	if browserName == "" {
		return "Unknown"
	}
	if browserVersion != "" {
		return browserName + " " + browserVersion
	}
	return browserName
}

func parseOS(uaStr string) string {
	ua := useragent.New(uaStr)
	osInfo := ua.OS()
	if osInfo == "" {
		return "Unknown"
	}
	return osInfo
}

func formatLoginLog(item model.LoginLog) map[string]interface{} {
	return map[string]interface{}{
		"id":        item.ID,
		"username":  item.Username,
		"ip":        item.IP,
		"location":  item.Location,
		"browser":   item.Browser,
		"os":        item.OS,
		"status":    item.Status,
		"message":   item.Message,
		"userAgent": item.UserAgent,
		"loginAt":   item.LoginAt,
		"createdAt": item.CreatedAt,
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/modules/user/service/login_log_service.go
git commit -m "feat: add LoginLog service with IP geolocation and UA parsing"
```

---

## Task 5: Backend — LoginLog API Handler

**Files:**
- Create: `backend/internal/modules/user/api/login_log.go`
- Modify: `backend/internal/modules/user/api/common.go`

- [ ] **Step 1: Create login_log.go API handler**

```go
package api

import (
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/modules/user/service"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	loginLogService *service.LoginLogService
	loginLogOnce    sync.Once
)

func getLoginLogService() *service.LoginLogService {
	loginLogOnce.Do(func() {
		loginLogService = service.NewLoginLogService(repository.NewLoginLogRepo(db), nil)
	})
	return loginLogService
}

// ListLoginLogs godoc
// @Summary 获取登录日志列表
// @Description 按条件分页查询登录日志
// @Tags 登录日志
// @Produce json
// @Security BearerAuth
// @Param username query string false "用户名"
// @Param status query string false "登录状态 (success/failed)"
// @Param startAt query string false "开始时间 (RFC3339)"
// @Param endAt query string false "结束时间 (RFC3339)"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "查询失败"
// @Router /login-log/list [get]
func ListLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	req := service.LoginLogListRequest{
		Username: c.Query("username"),
		Status:   c.Query("status"),
		StartAt:  c.Query("startAt"),
		EndAt:    c.Query("endAt"),
		Page:     page,
		PageSize: pageSize,
	}

	logs, total, err := getLoginLogService().List(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "查询登录日志失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":     logs,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}
```

- [ ] **Step 2: Update common.go — add loginLogService reset in SetDB**

In `backend/internal/modules/user/api/common.go`, add to the `SetDB` function body (after `tenantOnce = sync.Once{}`):

```go
	loginLogService = nil
	loginLogOnce = sync.Once{}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/modules/user/api/login_log.go backend/internal/modules/user/api/common.go
git commit -m "feat: add LoginLog API handler and register in SetDB"
```

---

## Task 6: Backend — LoginLog Route Registration

**Files:**
- Create: `backend/internal/routers/v1/login_log.go`
- Modify: `backend/routers/v1/v1.go`

- [ ] **Step 1: Create login_log.go route file**

```go
package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func registerLoginLog(r *gin.RouterGroup) {
	g := r.Group("/login-log")
	permission := middleware.RequirePermission("login-log", "list")

	g.GET("/list", permission, api.ListLoginLogs)
}
```

- [ ] **Step 2: Register in v1.go**

In `backend/routers/v1/v1.go`, add inside `Register` function after `registerAudit(auth)`:

```go
		registerLoginLog(auth)
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/routers/v1/login_log.go backend/routers/v1/v1.go
git commit -m "feat: register login-log routes with permission guard"
```

---

## Task 7: Backend — Bootstrap (AutoMigrate + Permission Seed)

**Files:**
- Modify: `backend/internal/bootstrap/db.go`

- [ ] **Step 1: Add LoginLog to AutoMigrate**

In `backend/internal/bootstrap/db.go`, update the `AutoMigrate` call (around line 63). Add `&userModel.LoginLog{}` after `&userModel.AuditLog{}`:

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
		)
```

- [ ] **Step 2: Add login-log permission seed**

In the `seedPermissions` function's `permissions` slice (around line 264, after the audit permission), add:

```go
			{Name: "查看登录日志", Resource: "login-log", Action: "list", Description: "查看登录日志"},
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/bootstrap/db.go
git commit -m "feat: add LoginLog to AutoMigrate and permission seed"
```

---

## Task 8: Backend — Integrate Login Log Capture in Auth

**Files:**
- Modify: `backend/internal/modules/user/api/auth.go`

- [ ] **Step 1: Add async login log recording to Login() function**

In `backend/internal/modules/user/api/auth.go`, in the `Login` function, after `getAuthService().Login(...)` call (line 71), wrap the existing success/error handling with login log recording. Replace the section from line 71 to line 93:

```go
	resp, err := getAuthService().Login(c.Request.Context(), &req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		logger.Log.Warn("Login failed", zap.String("username", req.Username), zap.Error(err))
		// 记录登录失败日志
		go func() {
			_ = getLoginLogService().CreateLoginLog(req.Username, c.ClientIP(), c.Request.UserAgent(), "failed", err.Error())
		}()
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": err.Error(),
		})
		return
	}

	// 记录登录成功日志
	go func() {
		_ = getLoginLogService().CreateLoginLog(req.Username, c.ClientIP(), c.Request.UserAgent(), "success", "登录成功")
	}()

	setSessionCookie(c, resp.SessionID, sessionCookieMaxAge())
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"token": resp.SessionID,
			"tenant": gin.H{
				"code": req.TenantCode,
			},
			"user":  resp.User,
		},
	})
```

- [ ] **Step 2: Add async login log recording to OIDCCallback() function**

In the same file, in `OIDCCallback`, after `getAuthService().LoginOIDC(...)` (line 179). Replace lines 179-198:

```go
	resp, err := getAuthService().LoginOIDC(c.Request.Context(), code, state, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		logger.Log.Error("OIDC Login failed", zap.Error(err))
		// 记录 OIDC 登录失败日志
		go func() {
			_ = getLoginLogService().CreateLoginLog("", c.ClientIP(), c.Request.UserAgent(), "failed", "OIDC登录失败: "+err.Error())
		}()
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": err.Error(),
		})
		return
	}

	// 记录 OIDC 登录成功日志
	username := ""
	if resp.User != nil {
		username = resp.User.Username
	}
	go func() {
		_ = getLoginLogService().CreateLoginLog(username, c.ClientIP(), c.Request.UserAgent(), "success", "OIDC登录成功")
	}()

	setSessionCookie(c, resp.SessionID, sessionCookieMaxAge())
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"token": resp.SessionID,
			"user":  resp.User,
		},
	})
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/modules/user/api/auth.go
git commit -m "feat: integrate login log capture in Login and OIDCCallback handlers"
```

---

## Task 9: Backend — Verify Build

- [ ] **Step 1: Run go build**

```bash
cd backend && go build ./...
```

Expected: no compilation errors

- [ ] **Step 2: Commit if any fixes needed**

---

## Task 10: Frontend — API Files

**Files:**
- Create: `frontend/src/api/audit.js`
- Create: `frontend/src/api/loginLog.js`

- [ ] **Step 1: Create audit.js**

```javascript
import request from './request'

export const getAuditList = (params) => request.get('/audit/list', { params })
```

- [ ] **Step 2: Create loginLog.js**

```javascript
import request from './request'

export const getLoginLogList = (params) => request.get('/login-log/list', { params })
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api/audit.js frontend/src/api/loginLog.js
git commit -m "feat: add audit and loginLog API modules"
```

---

## Task 11: Frontend — Sidebar + Routes

**Files:**
- Modify: `frontend/src/components/Layout/MainLayout.vue`
- Modify: `frontend/src/router/index.js`

- [ ] **Step 1: Update MainLayout.vue sidebar**

In `frontend/src/components/Layout/MainLayout.vue`:

Add `Notebook` to the icon import on line 59:

```javascript
import { HomeFilled, Grid, Setting, Expand, Fold, Notebook } from '@element-plus/icons-vue'
```

Add a new `el-sub-menu` after the system `el-sub-menu` closing tag (after line 35, before `</el-menu>`):

```html
        <el-sub-menu index="audit">
          <template #title>
            <el-icon><Notebook /></el-icon>
            <span>操作审计</span>
          </template>
          <el-menu-item index="/audit/operation">操作日志</el-menu-item>
          <el-menu-item index="/audit/login">登录日志</el-menu-item>
        </el-sub-menu>
```

- [ ] **Step 2: Update router/index.js**

In `frontend/src/router/index.js`, add two new child routes after the `system/permission` route (after line 73):

```javascript
          {
            path: 'audit/operation',
            component: () => import('../views/Audit/OperationLog.vue')
          },
          {
            path: 'audit/login',
            component: () => import('../views/Audit/LoginLog.vue')
          }
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/Layout/MainLayout.vue frontend/src/router/index.js
git commit -m "feat: add audit sidebar menu and routes"
```

---

## Task 12: Frontend — Operation Log Page

**Files:**
- Create: `frontend/src/views/Audit/OperationLog.vue`

- [ ] **Step 1: Create OperationLog.vue**

```vue
<template>
  <div class="page-container">
    <div class="page-header">
      <h3>操作日志</h3>
    </div>

    <div style="margin-bottom: 16px; display: flex; gap: 12px; flex-wrap: wrap;">
      <el-input v-model="filters.username" placeholder="用户账号" clearable style="width: 160px;" @keyup.enter="handleSearch" @clear="handleSearch" />
      <el-select v-model="filters.method" placeholder="请求方式" clearable style="width: 120px;" @change="handleSearch">
        <el-option label="GET" value="GET" />
        <el-option label="POST" value="POST" />
        <el-option label="PUT" value="PUT" />
        <el-option label="DELETE" value="DELETE" />
      </el-select>
      <el-input v-model="filters.operation" placeholder="操作描述" clearable style="width: 200px;" @keyup.enter="handleSearch" @clear="handleSearch" />
      <el-date-picker
        v-model="dateRange"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        value-format="YYYY-MM-DDTHH:mm:ssZ"
        style="width: 360px;"
        @change="handleSearch"
      />
      <el-button type="primary" @click="handleSearch">搜索</el-button>
    </div>

    <el-table :data="tableData" stripe v-loading="loading" style="width: 100%">
      <el-table-column prop="username" label="用户账号" width="120" />
      <el-table-column label="请求方式" width="100">
        <template #default="{ row }">
          <el-tag :type="methodTagType(row.method)" size="small">{{ row.method }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="ip" label="登录IP" width="140" />
      <el-table-column prop="path" label="请求URL" min-width="200" show-overflow-tooltip />
      <el-table-column prop="operation" label="操作描述" min-width="150" show-overflow-tooltip />
      <el-table-column label="操作时间" width="170">
        <template #default="{ row }">{{ formatTime(row.requestAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="showDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="fetchData"
        @size-change="fetchData"
      />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="操作详情" width="600px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="用户账号">{{ detail.username }}</el-descriptions-item>
        <el-descriptions-item label="请求方式">{{ detail.method }}</el-descriptions-item>
        <el-descriptions-item label="请求IP">{{ detail.ip }}</el-descriptions-item>
        <el-descriptions-item label="HTTP状态">{{ detail.status }}</el-descriptions-item>
        <el-descriptions-item label="响应耗时">{{ detail.latency }} ms</el-descriptions-item>
        <el-descriptions-item label="操作时间">{{ formatTime(detail.requestAt) }}</el-descriptions-item>
        <el-descriptions-item label="请求路径" :span="2">{{ detail.path }}</el-descriptions-item>
      </el-descriptions>
      <div v-if="detail.params" style="margin-top: 16px;">
        <div style="font-weight: 500; margin-bottom: 8px;">请求参数</div>
        <el-input type="textarea" :model-value="formatJSON(detail.params)" :rows="4" readonly />
      </div>
      <div v-if="detail.result" style="margin-top: 16px;">
        <div style="font-weight: 500; margin-bottom: 8px;">返回结果</div>
        <el-input type="textarea" :model-value="formatJSON(detail.result)" :rows="4" readonly />
      </div>
      <div v-if="detail.errorMessage" style="margin-top: 16px;">
        <div style="font-weight: 500; margin-bottom: 8px; color: #f56c6c;">错误信息</div>
        <el-input type="textarea" :model-value="detail.errorMessage" :rows="2" readonly />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getAuditList } from '@/api/audit'
import dayjs from 'dayjs'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const dateRange = ref(null)

const filters = reactive({
  username: '',
  method: '',
  operation: ''
})

const detailVisible = ref(false)
const detail = ref({})

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (filters.username) params.username = filters.username
    if (filters.method) params.keyword = filters.method
    if (filters.operation) params.operation = filters.operation
    if (dateRange.value && dateRange.value.length === 2) {
      params.startAt = dateRange.value[0]
      params.endAt = dateRange.value[1]
    }
    const res = await getAuditList(params)
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  fetchData()
}

const showDetail = (row) => {
  detail.value = row
  detailVisible.value = true
}

const methodTagType = (method) => {
  const map = { GET: 'success', POST: 'warning', DELETE: 'danger', PUT: '', PATCH: 'info' }
  return map[method] || 'info'
}

const formatTime = (val) => {
  if (!val) return '-'
  return dayjs(val).format('YYYY-MM-DD HH:mm:ss')
}

const formatJSON = (str) => {
  if (!str) return ''
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

onMounted(fetchData)
</script>

<style scoped>
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
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/Audit/OperationLog.vue
git commit -m "feat: add OperationLog page with filters and detail dialog"
```

---

## Task 13: Frontend — Login Log Page

**Files:**
- Create: `frontend/src/views/Audit/LoginLog.vue`

- [ ] **Step 1: Create LoginLog.vue**

```vue
<template>
  <div class="page-container">
    <div class="page-header">
      <h3>登录日志</h3>
    </div>

    <div style="margin-bottom: 16px; display: flex; gap: 12px; flex-wrap: wrap;">
      <el-input v-model="filters.username" placeholder="用户账号" clearable style="width: 200px;" @keyup.enter="handleSearch" @clear="handleSearch" />
      <el-select v-model="filters.status" placeholder="登录状态" clearable style="width: 120px;" @change="handleSearch">
        <el-option label="成功" value="success" />
        <el-option label="失败" value="failed" />
      </el-select>
      <el-date-picker
        v-model="dateRange"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        value-format="YYYY-MM-DDTHH:mm:ssZ"
        style="width: 360px;"
        @change="handleSearch"
      />
      <el-button type="primary" @click="handleSearch">搜索</el-button>
    </div>

    <el-table :data="tableData" stripe v-loading="loading" style="width: 100%">
      <el-table-column prop="username" label="用户账号" width="120" />
      <el-table-column prop="ip" label="登录IP" width="140" />
      <el-table-column prop="location" label="登录地点" width="150" />
      <el-table-column prop="browser" label="浏览器" width="140" />
      <el-table-column prop="os" label="操作系统" width="140" />
      <el-table-column label="登录状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
            {{ row.status === 'success' ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="message" label="提示消息" min-width="150" show-overflow-tooltip />
      <el-table-column label="访问时间" width="170">
        <template #default="{ row }">{{ formatTime(row.loginAt) }}</template>
      </el-table-column>
    </el-table>

    <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="fetchData"
        @size-change="fetchData"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getLoginLogList } from '@/api/loginLog'
import dayjs from 'dayjs'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const dateRange = ref(null)

const filters = reactive({
  username: '',
  status: ''
})

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (filters.username) params.username = filters.username
    if (filters.status) params.status = filters.status
    if (dateRange.value && dateRange.value.length === 2) {
      params.startAt = dateRange.value[0]
      params.endAt = dateRange.value[1]
    }
    const res = await getLoginLogList(params)
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  fetchData()
}

const formatTime = (val) => {
  if (!val) return '-'
  return dayjs(val).format('YYYY-MM-DD HH:mm:ss')
}

onMounted(fetchData)
</script>

<style scoped>
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
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/Audit/LoginLog.vue
git commit -m "feat: add LoginLog page with filters and status tags"
```

---

## Task 14: Download ip2region.xdb data file

- [ ] **Step 1: Download ip2region.xdb to backend directory**

```bash
cd backend && curl -L -o ip2region.xdb "https://raw.githubusercontent.com/lionsoul2014/ip2region/master/data/ip2region.xdb"
```

Expected: file downloaded, ~11MB

- [ ] **Step 2: Verify file exists**

```bash
ls -la backend/ip2region.xdb
```

- [ ] **Step 3: Commit**

```bash
git add -f backend/ip2region.xdb
git commit -m "chore: add ip2region offline IP database"
```

---

## Task 15: Final Verification

- [ ] **Step 1: Backend build check**

```bash
cd backend && go build ./...
```

Expected: no errors

- [ ] **Step 2: Frontend build check**

```bash
cd frontend && npm run build
```

Expected: no errors

- [ ] **Step 3: Start backend and verify endpoints**

```bash
cd backend && go run cmd/server/main.go
```

Verify in another terminal:
- GET `/api/v1/login-log/list` returns 200 with pagination response
- Login attempt records appear in `login_logs` table

- [ ] **Step 4: Start frontend and verify pages**

```bash
cd frontend && npm run dev
```

Verify:
- Sidebar shows 「操作审计」 menu with two sub-items
- Clicking 操作日志 navigates to `/audit/operation` and shows table
- Clicking 登录日志 navigates to `/audit/login` and shows table
- Filters and pagination work on both pages
