# Operations Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the K8s-only dashboard with a rich ops dashboard showing CMDB stats, host health, activity feed, and personalized "My Hosts" quick access.

**Architecture:** Backend adds a single `/api/v1/cmdb/dashboard` endpoint that aggregates stats from existing host, terminal, file audit, and cloud resource tables. Frontend rewrites `Dashboard/index.vue` to merge K8s data (existing APIs) with CMDB data (new API), using ECharts for charts.

**Tech Stack:** Go/Gin/GORM (backend), Vue 3 + Element Plus + ECharts (frontend)

---

## File Structure

| Action | Path | Responsibility |
|--------|------|---------------|
| Create | `backend/internal/modules/cmdb/model/dashboard.go` | Response DTOs |
| Create | `backend/internal/modules/cmdb/repository/dashboard.go` | Aggregation queries |
| Create | `backend/internal/modules/cmdb/service/dashboard.go` | Business logic, data assembly |
| Create | `backend/internal/modules/cmdb/api/dashboard.go` | HTTP handler |
| Modify | `backend/internal/modules/cmdb/api/common.go` | Add dashboard service singleton |
| Modify | `backend/routers/v1/cmdb.go` | Register dashboard route |
| Create | `frontend/src/api/cmdb/dashboard.js` | Frontend API client |
| Rewrite | `frontend/src/views/Dashboard/index.vue` | Full dashboard page |

---

### Task 1: Dashboard Backend — Model & Repository

**Files:**
- Create: `backend/internal/modules/cmdb/model/dashboard.go`
- Create: `backend/internal/modules/cmdb/repository/dashboard.go`

- [ ] **Step 1: Create dashboard response DTOs**

Create `backend/internal/modules/cmdb/model/dashboard.go`:

```go
package model

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	Hosts     HostStats     `json:"hosts"`
	Terminals TerminalStats `json:"terminals"`
	Cloud     CloudStats    `json:"cloud"`
	Files     FileStats     `json:"files"`
}

// HostStats 主机统计
type HostStats struct {
	Total   int64        `json:"total"`
	Online  int64        `json:"online"`
	Warning int64        `json:"warning"`
	Offline int64        `json:"offline"`
	Unknown int64        `json:"unknown"`
	ByGroup []GroupCount `json:"byGroup"`
}

// GroupCount 分组主机计数
type GroupCount struct {
	GroupID   uint   `json:"groupId"`
	GroupName string `json:"groupName"`
	Count     int64  `json:"count"`
}

// TerminalStats 终端统计
type TerminalStats struct {
	ActiveCount int64 `json:"activeCount"`
	TodayCount  int64 `json:"todayCount"`
	OnlineUsers int64 `json:"onlineUsers"`
}

// CloudStats 云资源统计
type CloudStats struct {
	InstanceCount int64  `json:"instanceCount"`
	LastSyncAt    string `json:"lastSyncAt"`
}

// FileStats 文件操作统计
type FileStats struct {
	TodayOps int64 `json:"todayOps"`
}

// ActivityEvent 活动事件
type ActivityEvent struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	User      string `json:"user"`
	Timestamp string `json:"timestamp"`
}

// MyHostInfo 我的常用主机
type MyHostInfo struct {
	ID           uint   `json:"id"`
	Hostname     string `json:"hostname"`
	IP           string `json:"ip"`
	Status       string `json:"status"`
	LastActiveAt string `json:"lastActiveAt"`
}

// DashboardResponse 仪表盘完整响应
type DashboardResponse struct {
	Stats      DashboardStats `json:"stats"`
	Activity   []ActivityEvent `json:"activity"`
	MyHosts    []MyHostInfo    `json:"myHosts"`
}
```

- [ ] **Step 2: Create dashboard repository**

Create `backend/internal/modules/cmdb/repository/dashboard.go`:

```go
package repository

import (
	"time"

	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
)

type DashboardRepo struct {
	db *gorm.DB
}

func NewDashboardRepo(db *gorm.DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

// GetHostStatusCounts 按状态统计主机数量
func (r *DashboardRepo) GetHostStatusCounts(tenantID uint) (map[string]int64, error) {
	type result struct {
		Status string
		Count  int64
	}
	var results []result
	err := r.db.Model(&model.Host{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Select("status, count(*) as count").
		Group("status").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}

// GetHostCountByGroup 按分组统计主机数量
func (r *DashboardRepo) GetHostCountByGroup(tenantID uint, limit int) ([]model.GroupCount, error) {
	var counts []model.GroupCount
	err := r.db.Model(&model.Host{}).
		Where("cmdb_hosts.tenant_id = ? AND cmdb_hosts.deleted_at IS NULL", tenantID).
		Joins("LEFT JOIN cmdb_groups ON cmdb_hosts.group_id = cmdb_groups.id").
		Select("COALESCE(cmdb_groups.id, 0) as group_id, COALESCE(cmdb_groups.name, '未分组') as group_name, count(*) as count").
		Group("cmdb_groups.id, cmdb_groups.name").
		Order("count DESC").
		Limit(limit).
		Find(&counts).Error
	return counts, err
}

// GetActiveTerminalCount 获取当前活跃终端会话数
func (r *DashboardRepo) GetActiveTerminalCount(tenantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.TerminalSession{}).
		Where("tenant_id = ? AND status = 'active' AND deleted_at IS NULL", tenantID).
		Count(&count).Error
	return count, err
}

// GetTodayTerminalCount 获取今日终端会话数
func (r *DashboardRepo) GetTodayTerminalCount(tenantID uint) (int64, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var count int64
	err := r.db.Model(&model.TerminalSession{}).
		Where("tenant_id = ? AND started_at >= ? AND deleted_at IS NULL", tenantID, today).
		Count(&count).Error
	return count, err
}

// GetOnlineTerminalUsers 获取当前在线终端用户数
func (r *DashboardRepo) GetOnlineTerminalUsers(tenantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.TerminalSession{}).
		Where("tenant_id = ? AND status = 'active' AND deleted_at IS NULL", tenantID).
		Distinct("user_id").
		Count(&count).Error
	return count, err
}

// GetCloudInstanceCount 获取云资源实例数
func (r *DashboardRepo) GetCloudInstanceCount(tenantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.CloudResource{}).
		Where("tenant_id = ? AND resource_type = 'cvm' AND deleted_at IS NULL", tenantID).
		Count(&count).Error
	return count, err
}

// GetLastCloudSyncAt 获取最后一次云同步时间
func (r *DashboardRepo) GetLastCloudSyncAt(tenantID uint) (string, error) {
	var account model.CloudAccount
	err := r.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Select("last_sync_at").
		Order("last_sync_at DESC").
		First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	if account.LastSyncAt == nil {
		return "", nil
	}
	return account.LastSyncAt.Format("2006-01-02 15:04:05"), nil
}

// GetTodayFileOps 获取今日文件操作数
func (r *DashboardRepo) GetTodayFileOps(tenantID uint) (int64, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var count int64
	err := r.db.Model(&model.FileOperationLog{}).
		Where("tenant_id = ? AND created_at >= ? AND deleted_at IS NULL", tenantID, today).
		Count(&count).Error
	return count, err
}

// GetRecentTerminalActivity 获取最近终端活动
func (r *DashboardRepo) GetRecentTerminalActivity(tenantID uint, limit int) ([]model.ActivityEvent, error) {
	var sessions []model.TerminalSession
	err := r.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("started_at DESC").
		Limit(limit).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	events := make([]model.ActivityEvent, 0, len(sessions))
	for _, s := range sessions {
		msg := s.Username + " 连接到 " + s.HostName + " (" + s.HostIP + ")"
		if s.Status != "active" {
			msg = s.Username + " 断开 " + s.HostName + " (" + s.HostIP + ")"
		}
		events = append(events, model.ActivityEvent{
			ID:        s.ID,
			Type:      "terminal",
			Message:   msg,
			User:      s.Username,
			Timestamp: s.StartedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return events, nil
}

// GetRecentFileActivity 获取最近文件操作活动
func (r *DashboardRepo) GetRecentFileActivity(tenantID uint, limit int) ([]model.ActivityEvent, error) {
	var logs []model.FileOperationLog
	err := r.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	events := make([]model.ActivityEvent, 0, len(logs))
	for _, l := range logs {
		events = append(events, model.ActivityEvent{
			ID:        l.ID,
			Type:      "file",
			Message:   l.Username + " " + l.OpType + " " + l.FilePath + " on " + l.HostName,
			User:      l.Username,
			Timestamp: l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return events, nil
}

// GetMyHosts 获取用户最近访问的主机
func (r *DashboardRepo) GetMyHosts(tenantID, userID uint, limit int) ([]model.MyHostInfo, error) {
	type recentSession struct {
		HostID    uint
		StartedAt time.Time
	}
	var recent []recentSession
	err := r.db.Model(&model.TerminalSession{}).
		Select("host_id, MAX(started_at) as started_at").
		Where("tenant_id = ? AND user_id = ? AND deleted_at IS NULL", tenantID, userID).
		Group("host_id").
		Order("started_at DESC").
		Limit(limit).
		Find(&recent).Error
	if err != nil {
		return nil, err
	}
	if len(recent) == 0 {
		return []model.MyHostInfo{}, nil
	}

	hostIDs := make([]uint, len(recent))
	sessionMap := make(map[uint]time.Time)
	for i, rs := range recent {
		hostIDs[i] = rs.HostID
		sessionMap[rs.HostID] = rs.StartedAt
	}

	var hosts []model.Host
	err = r.db.Where("id IN ? AND deleted_at IS NULL", hostIDs).Find(&hosts).Error
	if err != nil {
		return nil, err
	}

	result := make([]model.MyHostInfo, 0, len(hosts))
	for _, h := range hosts {
		lastActive := ""
		if t, ok := sessionMap[h.ID]; ok {
			lastActive = t.Format("2006-01-02 15:04:05")
		}
		result = append(result, model.MyHostInfo{
			ID:           h.ID,
			Hostname:     h.Hostname,
			IP:           h.Ip,
			Status:       h.Status,
			LastActiveAt: lastActive,
		})
	}
	return result, nil
}
```

- [ ] **Step 3: Verify compilation**

Run: `cd backend && go build ./internal/modules/cmdb/...`
Expected: compiles with no errors

- [ ] **Step 4: Commit**

```bash
cd backend
git add internal/modules/cmdb/model/dashboard.go internal/modules/cmdb/repository/dashboard.go
git commit -m "feat(cmdb): add dashboard model and repository for stats aggregation"
```

---

### Task 2: Dashboard Backend — Service & API

**Files:**
- Create: `backend/internal/modules/cmdb/service/dashboard.go`
- Create: `backend/internal/modules/cmdb/api/dashboard.go`
- Modify: `backend/internal/modules/cmdb/api/common.go`
- Modify: `backend/routers/v1/cmdb.go`

- [ ] **Step 1: Create dashboard service**

Create `backend/internal/modules/cmdb/service/dashboard.go`:

```go
package service

import (
	"sort"

	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
)

type DashboardService struct {
	repo *repository.DashboardRepo
}

func NewDashboardService(repo *repository.DashboardRepo) *DashboardService {
	return &DashboardService{repo: repo}
}

const dashboardActivityLimit = 10
const dashboardMyHostsLimit = 8

func (s *DashboardService) GetDashboard(tenantID, userID uint) (*model.DashboardResponse, error) {
	stats, err := s.getStats(tenantID)
	if err != nil {
		return nil, err
	}

	activity, err := s.getActivity(tenantID)
	if err != nil {
		return nil, err
	}

	myHosts, err := s.repo.GetMyHosts(tenantID, userID, dashboardMyHostsLimit)
	if err != nil {
		return nil, err
	}

	return &model.DashboardResponse{
		Stats:    *stats,
		Activity: activity,
		MyHosts:  myHosts,
	}, nil
}

func (s *DashboardService) getStats(tenantID uint) (*model.DashboardStats, error) {
	hostCounts, err := s.repo.GetHostStatusCounts(tenantID)
	if err != nil {
		return nil, err
	}

	hostByGroup, err := s.repo.GetHostCountByGroup(tenantID, 10)
	if err != nil {
		return nil, err
	}

	activeTerminals, err := s.repo.GetActiveTerminalCount(tenantID)
	if err != nil {
		return nil, err
	}

	todayTerminals, err := s.repo.GetTodayTerminalCount(tenantID)
	if err != nil {
		return nil, err
	}

	onlineUsers, err := s.repo.GetOnlineTerminalUsers(tenantID)
	if err != nil {
		return nil, err
	}

	cloudInstances, err := s.repo.GetCloudInstanceCount(tenantID)
	if err != nil {
		return nil, err
	}

	lastSyncAt, err := s.repo.GetLastCloudSyncAt(tenantID)
	if err != nil {
		return nil, err
	}

	todayFileOps, err := s.repo.GetTodayFileOps(tenantID)
	if err != nil {
		return nil, err
	}

	return &model.DashboardStats{
		Hosts: model.HostStats{
			Total:   hostCounts["online"] + hostCounts["offline"] + hostCounts["warning"] + hostCounts["unknown"],
			Online:  hostCounts["online"],
			Warning: hostCounts["warning"],
			Offline: hostCounts["offline"],
			Unknown: hostCounts["unknown"],
			ByGroup: hostByGroup,
		},
		Terminals: model.TerminalStats{
			ActiveCount: activeTerminals,
			TodayCount:  todayTerminals,
			OnlineUsers: onlineUsers,
		},
		Cloud: model.CloudStats{
			InstanceCount: cloudInstances,
			LastSyncAt:    lastSyncAt,
		},
		Files: model.FileStats{
			TodayOps: todayFileOps,
		},
	}, nil
}

func (s *DashboardService) getActivity(tenantID uint) ([]model.ActivityEvent, error) {
	terminalEvents, err := s.repo.GetRecentTerminalActivity(tenantID, dashboardActivityLimit)
	if err != nil {
		return nil, err
	}

	fileEvents, err := s.repo.GetRecentFileActivity(tenantID, dashboardActivityLimit)
	if err != nil {
		return nil, err
	}

	all := append(terminalEvents, fileEvents...)
	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp > all[j].Timestamp
	})
	if len(all) > dashboardActivityLimit {
		all = all[:dashboardActivityLimit]
	}
	return all, nil
}
```

- [ ] **Step 2: Create dashboard API handler**

Create `backend/internal/modules/cmdb/api/dashboard.go`:

```go
package api

import (
	"github.com/gin-gonic/gin"
)

func DashboardData(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(401, gin.H{"code": 401, "message": "未授权"})
		return
	}

	svc := getDashboardService()
	if svc == nil {
		c.JSON(500, gin.H{"code": 500, "message": "服务未初始化"})
		return
	}

	data, err := svc.GetDashboard(tenantID, userID)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "获取仪表盘数据失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}
```

- [ ] **Step 3: Add dashboard service singleton to common.go**

Add to the `var` block in `backend/internal/modules/cmdb/api/common.go`:

```go
dashboardSvcInstance *service.DashboardService
```

Add `dashboardSvcInstance = nil` to the `SetDB` function's body (after `fileSvcInstance = nil`).

Add getter function after `getFileService()`:

```go
func getDashboardService() *service.DashboardService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if dashboardSvcInstance != nil {
		return dashboardSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	repo := repository.NewDashboardRepo(cmdbDB)
	dashboardSvcInstance = service.NewDashboardService(repo)
	return dashboardSvcInstance
}
```

Add import for `"devops-platform/internal/modules/cmdb/repository"`.

- [ ] **Step 4: Register dashboard route in cmdb.go**

Add to `backend/routers/v1/cmdb.go`, inside the `registerCMDB` function, before the main route block:

```go
dashboardPerm := middleware.RequirePermission("cmdb:host", "list")
```

Add inside the route block:

```go
// 仪表盘
g.GET("/dashboard", dashboardPerm, api.DashboardData)
```

- [ ] **Step 5: Verify compilation and test API**

Run: `cd backend && go build ./...`
Expected: compiles with no errors

Start the backend server and test:
```bash
curl -H "Authorization: Bearer <session_id>" http://localhost:8000/api/v1/cmdb/dashboard
```
Expected: JSON response with `stats`, `activity`, `myHosts` fields

- [ ] **Step 6: Commit**

```bash
cd backend
git add internal/modules/cmdb/service/dashboard.go internal/modules/cmdb/api/dashboard.go internal/modules/cmdb/api/common.go routers/v1/cmdb.go
git commit -m "feat(cmdb): add dashboard API endpoint with stats, activity, and my-hosts"
```

---

### Task 3: Dashboard Frontend — API Service & ECharts Setup

**Files:**
- Create: `frontend/src/api/cmdb/dashboard.js`

- [ ] **Step 1: Install ECharts**

Run: `cd frontend && npm install echarts`

- [ ] **Step 2: Create dashboard API service**

Create `frontend/src/api/cmdb/dashboard.js`:

```js
import request from '../request'

export const getCmdbDashboard = () => request.get('/cmdb/dashboard')
```

- [ ] **Step 3: Commit**

```bash
cd frontend
git add src/api/cmdb/dashboard.js package.json package-lock.json
git commit -m "feat(cmdb): add dashboard frontend API service and echarts dependency"
```

---

### Task 4: Dashboard Frontend — Full Page Rewrite

**Files:**
- Rewrite: `frontend/src/views/Dashboard/index.vue`

- [ ] **Step 1: Rewrite the Dashboard page**

Write `frontend/src/views/Dashboard/index.vue`:

```vue
<template>
  <div class="page-container">
    <div class="page-header">
      <h3>运维仪表盘</h3>
      <el-button @click="fetchAllData" :loading="loading"><el-icon><Refresh /></el-icon>刷新</el-button>
    </div>

    <!-- Stats Cards -->
    <el-row :gutter="12" style="margin-bottom: 20px;">
      <el-col :span="4" v-for="card in statCards" :key="card.key">
        <el-card shadow="hover" class="stat-card" :body-style="{ padding: '16px' }">
          <div class="stat-label">{{ card.title }}</div>
          <div class="stat-value" :style="{ color: card.color }">{{ card.value }}</div>
          <div class="stat-sub">{{ card.sub }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Charts Row -->
    <el-row :gutter="16" style="margin-bottom: 20px;">
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>主机状态分布</span></template>
          <div ref="hostChartRef" style="height: 240px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>分组主机分布</span></template>
          <div ref="groupChartRef" style="height: 240px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>最近活动</span></template>
          <div class="activity-feed">
            <div v-for="event in cmdbData?.activity || []" :key="event.id + event.type" class="activity-item">
              <el-tag :type="activityTagType(event.type)" size="small" class="activity-tag">
                {{ activityLabel(event.type) }}
              </el-tag>
              <span class="activity-text">{{ event.message }}</span>
              <span class="activity-time">{{ relativeTime(event.timestamp) }}</span>
            </div>
            <el-empty v-if="!cmdbData?.activity?.length" description="暂无活动" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Bottom Row: K8s Clusters + My Hosts + Quick Actions -->
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card shadow="never">
          <template #header><span>集群状态</span></template>
          <el-table :data="clusters" stripe v-loading="k8sLoading" style="width: 100%" size="small">
            <el-table-column prop="name" label="集群" width="160">
              <template #default="{ row }">
                <router-link :to="`/k8s/cluster/${row.name}`" class="link">{{ row.name }}</router-link>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 'healthy' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'healthy' ? '健康' : '异常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="节点" width="60"><template #default="{ row }">{{ row.nodeCount || 0 }}</template></el-table-column>
            <el-table-column label="Pod" width="60"><template #default="{ row }">{{ row.podCount || 0 }}</template></el-table-column>
            <el-table-column label="CPU" width="120">
              <template #default="{ row }"><el-progress :percentage="row.cpuUsage || 0" :color="progressColor(row.cpuUsage)" :stroke-width="8" /></template>
            </el-table-column>
            <el-table-column label="内存" width="120">
              <template #default="{ row }"><el-progress :percentage="row.memoryUsage || 0" :color="progressColor(row.memoryUsage)" :stroke-width="8" /></template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>我的主机</span></template>
          <div class="my-hosts-grid" v-if="cmdbData?.myHosts?.length">
            <div v-for="host in cmdbData.myHosts" :key="host.id" class="host-card" @click="openTerminal(host)">
              <div class="host-status" :class="'status-' + host.status"></div>
              <div class="host-name">{{ host.hostname }}</div>
              <div class="host-ip">{{ host.ip }}</div>
            </div>
          </div>
          <el-empty v-else description="暂无访问记录" :image-size="60" />
        </el-card>
      </el-col>

      <el-col :span="4">
        <el-card shadow="never">
          <template #header><span>快捷操作</span></template>
          <div class="quick-actions">
            <router-link to="/cmdb/hosts" class="action-btn primary">
              <el-icon><Monitor /></el-icon>
              <span>主机列表</span>
            </router-link>
            <router-link to="/cmdb/terminal/sessions" class="action-btn warning">
              <el-icon><VideoCamera /></el-icon>
              <span>终端审计</span>
            </router-link>
            <router-link to="/cmdb/files" class="action-btn success">
              <el-icon><FolderOpened /></el-icon>
              <span>文件管理</span>
            </router-link>
            <router-link to="/cmdb/cloud-accounts" class="action-btn info">
              <el-icon><Cloudy /></el-icon>
              <span>云资源</span>
            </router-link>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh, Monitor, VideoCamera, FolderOpened, Cloudy } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { getClusterList } from '@/api/cluster'
import { getCmdbDashboard } from '@/api/cmdb/dashboard'
import { getTerminalConnectWsUrl } from '@/api/cmdb/terminal'

const router = useRouter()
const loading = ref(false)
const k8sLoading = ref(false)
const clusters = ref([])
const cmdbData = ref(null)

const hostChartRef = ref(null)
const groupChartRef = ref(null)
let hostChart = null
let groupChart = null

const statCards = computed(() => {
  const d = cmdbData.value
  if (!d) return []
  return [
    { key: 'hosts', title: '主机总数', value: d.stats.hosts.total, sub: d.stats.hosts.online + ' 在线', color: '#67C23A' },
    { key: 'terminals', title: '活跃终端', value: d.stats.terminals.activeCount, sub: d.stats.terminals.onlineUsers + ' 用户在线', color: '#E6A23C' },
    { key: 'todaySessions', title: '今日会话', value: d.stats.terminals.todayCount, sub: '', color: '#409EFF' },
    { key: 'cloud', title: '云实例', value: d.stats.cloud.instanceCount, sub: d.stats.cloud.lastSyncAt ? '同步于 ' + d.stats.cloud.lastSyncAt : '未同步', color: '#9B59B6' },
    { key: 'fileOps', title: '今日文件操作', value: d.stats.files.todayOps, sub: '', color: '#3498DB' },
    { key: 'clusters', title: 'K8s 集群', value: clusters.value.length, sub: clusters.value.reduce((s, c) => s + (c.podCount || 0), 0) + ' pods', color: '#F56C6C' }
  ]
})

const progressColor = (p) => { if (p >= 90) return '#F56C6C'; if (p >= 70) return '#E6A23C'; return '#67C23A' }

const activityTagType = (type) => {
  const map = { terminal: 'danger', file: 'warning', sync: 'success', host: '' }
  return map[type] || 'info'
}

const activityLabel = (type) => {
  const map = { terminal: '终端', file: '文件', sync: '同步', host: '主机' }
  return map[type] || type
}

const relativeTime = (ts) => {
  if (!ts) return ''
  const diff = Date.now() - new Date(ts).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return mins + '分钟前'
  const hours = Math.floor(mins / 60)
  if (hours < 24) return hours + '小时前'
  return Math.floor(hours / 24) + '天前'
}

const openTerminal = (host) => {
  router.push({ path: '/cmdb/hosts', query: { terminalHostId: host.id } })
}

const renderHostChart = () => {
  if (!hostChartRef.value || !cmdbData.value) return
  if (!hostChart) {
    hostChart = echarts.init(hostChartRef.value)
  }
  const h = cmdbData.value.stats.hosts
  hostChart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, textStyle: { fontSize: 11 } },
    series: [{
      type: 'pie',
      radius: ['40%', '65%'],
      center: ['50%', '45%'],
      label: { show: false },
      data: [
        { value: h.online, name: '在线', itemStyle: { color: '#67C23A' } },
        { value: h.warning, name: '告警', itemStyle: { color: '#E6A23C' } },
        { value: h.offline, name: '离线', itemStyle: { color: '#F56C6C' } },
        { value: h.unknown, name: '未知', itemStyle: { color: '#909399' } }
      ].filter(d => d.value > 0)
    }]
  })
}

const renderGroupChart = () => {
  if (!groupChartRef.value || !cmdbData.value) return
  if (!groupChart) {
    groupChart = echarts.init(groupChartRef.value)
  }
  const groups = cmdbData.value.stats.hosts.byGroup || []
  groupChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: 80, right: 20, top: 10, bottom: 30 },
    xAxis: { type: 'value' },
    yAxis: {
      type: 'category',
      data: groups.map(g => g.groupName).reverse(),
      axisLabel: { fontSize: 11, width: 70, overflow: 'truncate' }
    },
    series: [{
      type: 'bar',
      data: groups.map(g => g.count).reverse(),
      itemStyle: { color: '#409EFF', borderRadius: [0, 4, 4, 0] },
      barWidth: 16,
      label: { show: true, position: 'right', fontSize: 11 }
    }]
  })
}

const fetchCmdbData = async () => {
  try {
    const res = await getCmdbDashboard()
    cmdbData.value = res.data
    await nextTick()
    renderHostChart()
    renderGroupChart()
  } catch (e) {
    ElMessage.error('获取仪表盘数据失败')
  }
}

const fetchK8sData = async () => {
  k8sLoading.value = true
  try {
    const res = await getClusterList()
    clusters.value = res.data?.list || res.data || []
  } finally {
    k8sLoading.value = false
  }
}

const fetchAllData = async () => {
  loading.value = true
  try {
    await Promise.all([fetchCmdbData(), fetchK8sData()])
  } finally {
    loading.value = false
  }
}

const handleResize = () => {
  hostChart?.resize()
  groupChart?.resize()
}

onMounted(fetchAllData)
onMounted(() => window.addEventListener('resize', handleResize))
onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  hostChart?.dispose()
  groupChart?.dispose()
})
</script>

<style scoped>
.page-container { padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }

.stat-card { text-align: center; }
.stat-label { font-size: 12px; color: #909399; margin-bottom: 6px; }
.stat-value { font-size: 28px; font-weight: 700; line-height: 1.2; }
.stat-sub { font-size: 11px; color: #b0b5bd; margin-top: 4px; }

.activity-feed { max-height: 240px; overflow-y: auto; }
.activity-item { display: flex; align-items: center; gap: 8px; padding: 6px 0; border-bottom: 1px solid #f0f0f0; font-size: 12px; }
.activity-tag { flex-shrink: 0; width: 36px; text-align: center; }
.activity-text { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.activity-time { color: #909399; flex-shrink: 0; font-size: 11px; }

.my-hosts-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.host-card { background: #f5f7fa; border-radius: 6px; padding: 10px; cursor: pointer; transition: background 0.2s; }
.host-card:hover { background: #ecf5ff; }
.host-status { display: inline-block; width: 6px; height: 6px; border-radius: 50%; margin-right: 4px; }
.status-online { background: #67C23A; }
.status-offline { background: #F56C6C; }
.status-warning { background: #E6A23C; }
.status-unknown { background: #909399; }
.host-name { font-size: 13px; font-weight: 600; }
.host-ip { font-size: 11px; color: #909399; }

.quick-actions { display: flex; flex-direction: column; gap: 8px; }
.action-btn { display: flex; align-items: center; gap: 6px; padding: 10px 14px; border-radius: 6px; color: #fff; font-size: 13px; text-decoration: none; transition: opacity 0.2s; }
.action-btn:hover { opacity: 0.85; color: #fff; }
.action-btn.primary { background: #409EFF; }
.action-btn.warning { background: #E6A23C; }
.action-btn.success { background: #67C23A; }
.action-btn.info { background: #909399; }

.link { color: var(--el-color-primary); text-decoration: none; }
.link:hover { text-decoration: underline; }
</style>
```

- [ ] **Step 2: Start dev server and verify visually**

Run: `cd frontend && npm run dev`
Open: `http://localhost:5173/dashboard`

Verify:
- Six stat cards across top row
- Host status donut chart renders with data
- Group distribution bar chart renders with data
- Activity feed shows recent terminal and file events
- K8s cluster table shows with existing data
- "My Hosts" cards show recently accessed hosts
- Quick action buttons link to correct pages
- Responsive resize on window resize

- [ ] **Step 3: Commit**

```bash
cd frontend
git add src/views/Dashboard/index.vue
git commit -m "feat(cmdb): rewrite dashboard with ops stats, charts, activity feed, and my-hosts"
```

---

## Self-Review Checklist

1. **Spec coverage:** Stats cards (hosts/terminals/cloud/files/K8s) — Task 2 + 4. Host status distribution chart — Task 4. Group distribution — Task 4. Activity feed — Task 2 + 4. My Hosts — Task 2 + 4. Quick actions — Task 4. All covered.

2. **Placeholder scan:** No TBD/TODO found. All code blocks contain complete implementations.

3. **Type consistency:** `DashboardResponse` in model matches `DashboardService.GetDashboard()` return type and API handler serialization. Frontend `statCards` computed property references `d.stats.hosts.total`, `d.stats.hosts.online`, etc. matching the Go DTO field names via JSON tags.
