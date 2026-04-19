# CMDB 核心补全 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补全 CMDB 核心功能：云同步分页拉取 + 启动接入、录像清理轮转、主机级权限接入终端和主机列表。

**Architecture:** 自底向上实施。先完善独立模块（云同步分页、录像清理），再做跨模块的权限接入。云同步提取公共分页函数；录像清理遵循审计日志清理的后台任务模式；权限接入通过 Casbin 的 `cmdb:host:admin` 权限判断管理员身份，非管理员走 HostPermission 表查询。

**Tech Stack:** Go 1.24+ / Gin / GORM / MySQL / Vue 3 / Element Plus / 腾讯云 SDK

---

## File Structure

| Operation | Path | Responsibility |
|-----------|------|----------------|
| Create | `backend/internal/modules/cmdb/service/cloud_sync_pagination.go` | 公共分页拉取逻辑 |
| Create | `backend/internal/modules/cmdb/service/cloud_sync_pagination_test.go` | 分页逻辑测试 |
| Create | `backend/internal/modules/cmdb/service/recording_cleanup.go` | 录像文件清理定时任务 |
| Modify | `backend/internal/modules/cmdb/service/cloud_sync.go` | 5 个 sync 方法改用分页 + syncRegion 容错 |
| Modify | `backend/cmd/server/main.go` | 接入云同步启动 + 录像清理启动 |
| Modify | `backend/config/defaults.go` | 新增录像清理默认配置 |
| Modify | `backend/config/config.yaml` | 新增录像清理配置项 |
| Modify | `backend/internal/bootstrap/db.go` | 新增 `cmdb:host:admin` 权限种子 |
| Modify | `backend/internal/modules/cmdb/api/common.go` | 新增 `isCmdbAdmin` 辅助函数 |
| Modify | `backend/internal/modules/cmdb/api/terminal.go` | 终端连接前权限校验 |
| Modify | `backend/internal/modules/cmdb/api/host.go` | 主机列表过滤 + 详情校验 |
| Modify | `backend/internal/modules/cmdb/repository/host.go` | ListInTenant 增加 hostIDs 过滤 |
| Modify | `backend/internal/modules/cmdb/service/host.go` | ListInTenant 透传 hostIDs |
| Modify | `frontend/src/views/Cmdb/HostList.vue` | 终端连接 403 处理 |

---

### Task 1: 云同步分页拉取

**Files:**
- Create: `backend/internal/modules/cmdb/service/cloud_sync_pagination.go`
- Create: `backend/internal/modules/cmdb/service/cloud_sync_pagination_test.go`
- Modify: `backend/internal/modules/cmdb/service/cloud_sync.go`

- [ ] **Step 1: 创建分页辅助函数**

创建 `backend/internal/modules/cmdb/service/cloud_sync_pagination.go`：

```go
package service

import (
	"devops-platform/internal/pkg/logger"

	"go.uber.org/zap"
)

const cloudSyncPageSize = 100

// maxPaginationPages 安全上限：单次同步最多拉取 200 页（20000 条）
const maxPaginationPages = 200

// paginateSync 通用分页拉取。fetchPage 回调传入 offset，返回本页条数和错误。
// 单页失败时记录日志并跳过该页继续拉取。
func (s *CloudAccountService) paginateSync(fetchPage func(offset int64) (int, error)) error {
	var offset int64
	for i := 0; i < maxPaginationPages; i++ {
		count, err := fetchPage(offset)
		if err != nil {
			logger.Log.Error("云同步分页拉取失败，跳过该页", zap.Int64("offset", offset), zap.Error(err))
			offset += cloudSyncPageSize
			continue
		}
		if count < cloudSyncPageSize {
			return nil
		}
		offset += cloudSyncPageSize
	}
	return nil
}
```

- [ ] **Step 2: 编写分页逻辑测试**

创建 `backend/internal/modules/cmdb/service/cloud_sync_pagination_test.go`：

```go
package service

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaginateSync_SinglePage(t *testing.T) {
	svc := &CloudAccountService{}
	called := int32(0)
	err := svc.paginateSync(func(offset int64) (int, error) {
		atomic.AddInt32(&called, 1)
		return 50, nil // 少于 pageSize，结束
	})
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&called))
}

func TestPaginateSync_MultiplePages(t *testing.T) {
	svc := &CloudAccountService{}
	pages := []int{100, 100, 100, 30}
	called := int32(0)
	err := svc.paginateSync(func(offset int64) (int, error) {
		idx := int(atomic.LoadInt32(&called))
		atomic.AddInt32(&called, 1)
		return pages[idx], nil
	})
	require.NoError(t, err)
	require.Equal(t, int32(4), atomic.LoadInt32(&called))
}

func TestPaginateSync_PageErrorSkipped(t *testing.T) {
	svc := &CloudAccountService{}
	called := int32(0)
	err := svc.paginateSync(func(offset int64) (int, error) {
		idx := atomic.AddInt32(&called, 1)
		if idx == 1 {
			return 0, errors.New("API error")
		}
		return 50, nil // 第二页结束
	})
	require.NoError(t, err)
	require.Equal(t, int32(2), atomic.LoadInt32(&called))
}

func TestPaginateSync_MaxPagesSafety(t *testing.T) {
	svc := &CloudAccountService{}
	called := int32(0)
	err := svc.paginateSync(func(offset int64) (int, error) {
		atomic.AddInt32(&called, 1)
		return 100, nil // 每页都满，触发 maxPaginationPages 上限
	})
	require.NoError(t, err)
	require.Equal(t, int32(maxPaginationPages), atomic.LoadInt32(&called))
}
```

- [ ] **Step 3: 运行测试确认通过**

Run: `cd backend && go test ./internal/modules/cmdb/service/ -run TestPaginateSync -v`
Expected: 4 tests PASS

- [ ] **Step 4: 重构 syncCVM 使用分页**

修改 `backend/internal/modules/cmdb/service/cloud_sync.go` 中 `syncCVM` 方法。替换原方法为：

```go
func (s *CloudAccountService) syncCVM(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := cvm.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	return s.paginateSync(func(offset int64) (int, error) {
		request := cvm.NewDescribeInstancesRequest()
		request.Limit = common.Int64Ptr(int64(cloudSyncPageSize))
		request.Offset = common.Int64Ptr(offset)

		response, err := client.DescribeInstances(request)
		if err != nil {
			return 0, err
		}
		if response == nil || response.Response == nil {
			return 0, nil
		}

		for _, instance := range response.Response.InstanceSet {
			instanceID := safeDereferenceString(instance.InstanceId)
			instanceName := safeDereferenceString(instance.InstanceName)
			privateIP := ""
			if len(instance.PrivateIpAddresses) > 0 && instance.PrivateIpAddresses[0] != nil {
				privateIP = *instance.PrivateIpAddresses[0]
			}
			state := safeDereferenceString(instance.InstanceState)
			zone := ""
			if instance.Placement != nil {
				zone = safeDereferenceString(instance.Placement.Zone)
			}
			osName := safeDereferenceString(instance.OsName)
			cpu := int(safeDereferenceInt64(instance.CPU))
			memory := int(safeDereferenceInt64(instance.Memory))

			specJSON, _ := json.Marshal(map[string]interface{}{
				"cpu": cpu, "memory": memory, "zone": zone,
			})
			resource := &model.CloudResource{
				TenantID:       account.TenantID,
				CloudAccountID: account.ID,
				ResourceType:   "cvm",
				ResourceID:     instanceID,
				Region:         region,
				Zone:           zone,
				Name:           instanceName,
				State:          state,
				Spec:           string(specJSON),
				SyncedAt:       time.Now(),
			}
			_ = s.repo.UpsertResource(resource)

			existing, err := s.repo.GetHostByCloudInstanceID(account.TenantID, instanceID)
			if err == nil && existing != nil {
				existing.Hostname = instanceName
				if privateIP != "" {
					existing.Ip = privateIP
				}
				existing.OsName = osName
				existing.CloudAccountID = &account.ID
				_ = s.repo.UpdateHost(existing)
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				tenantID := account.TenantID
				accountID := account.ID
				host := &model.Host{
					TenantID:        &tenantID,
					Hostname:        instanceName,
					Ip:              privateIP,
					Port:            22,
					OsName:          osName,
					Status:          state,
					CloudAccountID:  &accountID,
					CloudInstanceID: instanceID,
				}
				_ = s.repo.CreateHost(host)
			}
		}
		return len(response.Response.InstanceSet), nil
	})
}
```

- [ ] **Step 5: 重构 syncVPC 使用分页**

替换 `syncVPC` 方法为：

```go
func (s *CloudAccountService) syncVPC(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := vpc.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	return s.paginateSync(func(offset int64) (int, error) {
		request := vpc.NewDescribeVpcsRequest()
		request.Limit = common.StringPtr("100")
		request.Offset = common.StringPtr(fmt.Sprintf("%d", offset))

		response, err := client.DescribeVpcs(request)
		if err != nil {
			return 0, err
		}
		if response == nil || response.Response == nil {
			return 0, nil
		}

		for _, v := range response.Response.VpcSet {
			vpcID := safeDereferenceString(v.VpcId)
			name := safeDereferenceString(v.VpcName)
			cidr := safeDereferenceString(v.CidrBlock)

			specJSON, _ := json.Marshal(map[string]string{"cidr": cidr})
			resource := &model.CloudResource{
				TenantID:       account.TenantID,
				CloudAccountID: account.ID,
				ResourceType:   "vpc",
				ResourceID:     vpcID,
				Region:         region,
				Name:           name,
				State:          "available",
				Spec:           string(specJSON),
				SyncedAt:       time.Now(),
			}
			_ = s.repo.UpsertResource(resource)
		}
		return len(response.Response.VpcSet), nil
	})
}
```

- [ ] **Step 6: 重构 syncSubnets 使用分页**

替换 `syncSubnets` 方法为：

```go
func (s *CloudAccountService) syncSubnets(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := vpc.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	return s.paginateSync(func(offset int64) (int, error) {
		request := vpc.NewDescribeSubnetsRequest()
		request.Limit = common.StringPtr("100")
		request.Offset = common.StringPtr(fmt.Sprintf("%d", offset))

		response, err := client.DescribeSubnets(request)
		if err != nil {
			return 0, err
		}
		if response == nil || response.Response == nil {
			return 0, nil
		}

		for _, sub := range response.Response.SubnetSet {
			subnetID := safeDereferenceString(sub.SubnetId)
			name := safeDereferenceString(sub.SubnetName)
			cidr := safeDereferenceString(sub.CidrBlock)
			vpcID := safeDereferenceString(sub.VpcId)
			zone := safeDereferenceString(sub.Zone)

			specJSON, _ := json.Marshal(map[string]string{"cidr": cidr, "vpc_id": vpcID})
			resource := &model.CloudResource{
				TenantID:       account.TenantID,
				CloudAccountID: account.ID,
				ResourceType:   "subnet",
				ResourceID:     subnetID,
				Region:         region,
				Zone:           zone,
				Name:           name,
				State:          "available",
				Spec:           string(specJSON),
				SyncedAt:       time.Now(),
			}
			_ = s.repo.UpsertResource(resource)
		}
		return len(response.Response.SubnetSet), nil
	})
}
```

- [ ] **Step 7: 重构 syncSecurityGroups 使用分页**

替换 `syncSecurityGroups` 方法为：

```go
func (s *CloudAccountService) syncSecurityGroups(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := vpc.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	return s.paginateSync(func(offset int64) (int, error) {
		request := vpc.NewDescribeSecurityGroupsRequest()
		request.Limit = common.StringPtr("100")
		request.Offset = common.StringPtr(fmt.Sprintf("%d", offset))

		response, err := client.DescribeSecurityGroups(request)
		if err != nil {
			return 0, err
		}
		if response == nil || response.Response == nil {
			return 0, nil
		}

		for _, sg := range response.Response.SecurityGroupSet {
			sgID := safeDereferenceString(sg.SecurityGroupId)
			name := safeDereferenceString(sg.SecurityGroupName)
			desc := safeDereferenceString(sg.SecurityGroupDesc)

			specJSON, _ := json.Marshal(map[string]string{"description": desc})
			resource := &model.CloudResource{
				TenantID:       account.TenantID,
				CloudAccountID: account.ID,
				ResourceType:   "security_group",
				ResourceID:     sgID,
				Region:         region,
				Name:           name,
				Spec:           string(specJSON),
				SyncedAt:       time.Now(),
			}
			_ = s.repo.UpsertResource(resource)
		}
		return len(response.Response.SecurityGroupSet), nil
	})
}
```

- [ ] **Step 8: 重构 syncCBS 使用分页**

替换 `syncCBS` 方法为：

```go
func (s *CloudAccountService) syncCBS(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := cbs.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	return s.paginateSync(func(offset int64) (int, error) {
		request := cbs.NewDescribeDisksRequest()
		request.Limit = common.Uint64Ptr(uint64(cloudSyncPageSize))
		request.Offset = common.Uint64Ptr(uint64(offset))

		response, err := client.DescribeDisks(request)
		if err != nil {
			return 0, err
		}
		if response == nil || response.Response == nil {
			return 0, nil
		}

		for _, disk := range response.Response.DiskSet {
			diskID := safeDereferenceString(disk.DiskId)
			name := safeDereferenceString(disk.DiskName)
			state := safeDereferenceString(disk.DiskState)
			diskType := safeDereferenceString(disk.DiskType)
			diskSize := safeDereferenceUint64(disk.DiskSize)
			zone := ""
			if disk.Placement != nil {
				zone = safeDereferenceString(disk.Placement.Zone)
			}

			specJSON, _ := json.Marshal(map[string]interface{}{
				"size": diskSize, "type": diskType,
			})
			resource := &model.CloudResource{
				TenantID:       account.TenantID,
				CloudAccountID: account.ID,
				ResourceType:   "cbs",
				ResourceID:     diskID,
				Region:         region,
				Zone:           zone,
				Name:           name,
				State:          state,
				Spec:           string(specJSON),
				SyncedAt:       time.Now(),
			}
			_ = s.repo.UpsertResource(resource)
		}
		return len(response.Response.DiskSet), nil
	})
}
```

- [ ] **Step 9: 改造 syncRegion 为容错模式**

替换 `syncRegion` 方法，使单个资源类型失败不阻塞其他类型：

```go
func (s *CloudAccountService) syncRegion(account *model.CloudAccount, secretID, secretKey, region string) error {
	credential := common.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "GET"

	var errs []string

	if err := s.syncCVM(credential, cpf, account, region); err != nil {
		errs = append(errs, fmt.Sprintf("CVM: %s", err.Error()))
	}
	if err := s.syncVPC(credential, cpf, account, region); err != nil {
		errs = append(errs, fmt.Sprintf("VPC: %s", err.Error()))
	}
	if err := s.syncSubnets(credential, cpf, account, region); err != nil {
		errs = append(errs, fmt.Sprintf("子网: %s", err.Error()))
	}
	if err := s.syncSecurityGroups(credential, cpf, account, region); err != nil {
		errs = append(errs, fmt.Sprintf("安全组: %s", err.Error()))
	}
	if err := s.syncCBS(credential, cpf, account, region); err != nil {
		errs = append(errs, fmt.Sprintf("云硬盘: %s", err.Error()))
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分资源同步失败: %s", strings.Join(errs, "; "))
	}
	return nil
}
```

- [ ] **Step 10: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功，无错误

- [ ] **Step 11: 运行分页测试**

Run: `cd backend && go test ./internal/modules/cmdb/service/ -run TestPaginateSync -v`
Expected: 4 tests PASS

- [ ] **Step 12: 提交**

```bash
git add backend/internal/modules/cmdb/service/cloud_sync_pagination.go \
        backend/internal/modules/cmdb/service/cloud_sync_pagination_test.go \
        backend/internal/modules/cmdb/service/cloud_sync.go
git commit -m "feat(cmdb): cloud sync pagination with error resilience"
```

---

### Task 2: 云同步启动接入

**Files:**
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: 在 main.go 中启动定时云同步**

在 `main.go` 的 `cmdbAPI.SetDB(bootstrap.DB)` 之后（约第 99 行后），添加：

```go
	// 启动定时云同步
	cmdbAPI.StartCloudSync()
```

- [ ] **Step 2: 在 common.go 中暴露启动方法**

在 `backend/internal/modules/cmdb/api/common.go` 中添加：

```go
// StartCloudSync 启动定时云同步（由 main.go 调用）
func StartCloudSync() {
	svc := getCloudAccountService()
	svc.StartScheduledSync()
	logger.Log.Info("定时云同步已启动")
}
```

需要添加 import：`"devops-platform/internal/pkg/logger"`

- [ ] **Step 3: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add backend/cmd/server/main.go backend/internal/modules/cmdb/api/common.go
git commit -m "feat(cmdb): wire up cloud sync scheduled task at startup"
```

---

### Task 3: 录像文件清理轮转

**Files:**
- Modify: `backend/config/defaults.go`
- Modify: `backend/config/config.yaml`
- Create: `backend/internal/modules/cmdb/service/recording_cleanup.go`
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/internal/modules/cmdb/api/common.go`

- [ ] **Step 1: 添加录像清理配置默认值**

在 `backend/config/defaults.go` 的 Terminal 默认配置段末尾添加：

```go
		// 录像清理默认配置
		v.SetDefault("terminal.recording.max_age_days", 90)
		v.SetDefault("terminal.recording.cleanup_hour", 3)
```

- [ ] **Step 2: 在 config.yaml 中添加配置**

在 `backend/config/config.yaml` 的 `terminal:` 段末尾添加：

```yaml
  recording:
    max_age_days: 90    # 录像保留天数，0 表示不清理
    cleanup_hour: 3     # 每天凌晨 3 点执行清理（0-23）
```

- [ ] **Step 3: 创建录像清理服务**

创建 `backend/internal/modules/cmdb/service/recording_cleanup.go`：

```go
package service

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"devops-platform/config"
	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
	"devops-platform/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RecordingCleanupService struct {
	repo   *repository.TerminalRepo
	db     *gorm.DB
	cancel context.CancelFunc
}

func NewRecordingCleanupService(db *gorm.DB) *RecordingCleanupService {
	return &RecordingCleanupService{
		repo: repository.NewTerminalRepo(db),
		db:   db,
	}
}

func (s *RecordingCleanupService) StartCleanupScheduler() {
	maxAge := config.Cfg.GetInt("terminal.recording.max_age_days")
	if maxAge <= 0 {
		logger.Log.Info("录像清理已禁用（max_age_days = 0）")
		return
	}

	cleanupHour := config.Cfg.GetInt("terminal.recording.cleanup_hour")
	if cleanupHour < 0 || cleanupHour > 23 {
		cleanupHour = 3
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		// 计算下次清理时间
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), cleanupHour, 0, 0, 0, now.Location())
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}
			timer := time.NewTimer(next.Sub(now))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				s.CleanupOldRecordings(maxAge)
			}
		}
	}()

	logger.Log.Info("录像清理定时任务已启动", zap.Int("max_age_days", maxAge), zap.Int("cleanup_hour", cleanupHour))
}

func (s *RecordingCleanupService) StopCleanupScheduler() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RecordingCleanupService) CleanupOldRecordings(maxAgeDays int) {
	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)

	var sessions []model.TerminalSession
	if err := s.db.Where("status = ? AND created_at < ?", "closed", cutoff).
		Find(&sessions).Error; err != nil {
		logger.Log.Error("查询过期录像记录失败", zap.Error(err))
		return
	}

	if len(sessions) == 0 {
		return
	}

	recordingDir := config.Cfg.GetString("terminal.recording_dir")
	cleaned := 0

	for _, session := range sessions {
		if session.RecordingPath == "" {
			continue
		}
		fullPath := filepath.Join(recordingDir, session.RecordingPath)
		if _, err := os.Stat(fullPath); err == nil {
			if err := os.Remove(fullPath); err != nil {
				logger.Log.Error("删除录像文件失败", zap.String("path", fullPath), zap.Error(err))
				continue
			}
		}
		// 标记为已归档
		s.db.Model(&session).Updates(map[string]interface{}{
			"status":        "archived",
			"recording_path": "",
		})
		cleaned++
	}

	// 清理空日期目录
	s.cleanEmptyDirs(recordingDir)

	logger.Log.Info("录像清理完成",
		zap.Int("cleaned", cleaned),
		zap.Int("total", len(sessions)))
}

func (s *RecordingCleanupService) cleanEmptyDirs(baseDir string) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(baseDir, entry.Name())
		// 尝试删除空目录
		os.Remove(dirPath)
	}
}
```

- [ ] **Step 4: 在 main.go 中启动录像清理**

在 `main.go` 的云同步启动代码之后添加：

```go
	// 启动录像清理定时任务
	cmdbAPI.StartRecordingCleanup()
```

- [ ] **Step 5: 在 common.go 中暴露启动方法**

在 `backend/internal/modules/cmdb/api/common.go` 中添加：

```go
// StartRecordingCleanup 启动录像清理定时任务（由 main.go 调用）
func StartRecordingCleanup() {
	svc := service.NewRecordingCleanupService(cmdbDB)
	svc.StartCleanupScheduler()
}
```

需要在 common.go 的 import 中添加 `"devops-platform/internal/modules/cmdb/service"`（如果尚未存在）。

注意：`common.go` 已通过 lazy getter 间接使用 service 包（`getHostService` 等调用 `service.NewXxxService`），所以 import 已存在。但 `StartRecordingCleanup` 直接引用了 `service.NewRecordingCleanupService`，确认 import 中有 service 包即可。

- [ ] **Step 6: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功

- [ ] **Step 7: 提交**

```bash
git add backend/config/defaults.go \
        backend/config/config.yaml \
        backend/internal/modules/cmdb/service/recording_cleanup.go \
        backend/cmd/server/main.go \
        backend/internal/modules/cmdb/api/common.go
git commit -m "feat(cmdb): recording cleanup rotation with configurable retention"
```

---

### Task 4: 新增 cmdb:host:admin 权限种子

**Files:**
- Modify: `backend/internal/bootstrap/db.go`

- [ ] **Step 1: 添加权限种子**

在 `backend/internal/bootstrap/db.go` 的 `seedPermissions()` 函数中，找到 CMDB 主机权限段（包含 `cmdb:host:list` 等的位置），在 `cmdb:host:test` 之后添加：

```go
		{Name: "主机管理（管理员）", Resource: "cmdb:host", Action: "admin", Description: "CMDB 主机管理员权限，跳过主机级权限过滤"},
```

- [ ] **Step 2: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add backend/internal/bootstrap/db.go
git commit -m "feat(cmdb): add cmdb:host:admin permission seed for bypass"
```

---

### Task 5: 终端连接权限校验

**Files:**
- Modify: `backend/internal/modules/cmdb/api/common.go`
- Modify: `backend/internal/modules/cmdb/api/terminal.go`

- [ ] **Step 1: 在 common.go 中添加 isCmdbAdmin 辅助函数**

在 `backend/internal/modules/cmdb/api/common.go` 中添加 import `"devops-platform/internal/modules/user/service"` 和 `"devops-platform/internal/pkg/logger"`（如果尚未存在），然后添加：

```go
// isCmdbAdmin 检查用户是否拥有 cmdb:host:admin 权限（管理员跳过主机级过滤）
func isCmdbAdmin(c *gin.Context, tenantID, userID uint) bool {
	userSvc := userservice.NewUserService(cmdbDB)
	isAdmin, err := userSvc.CheckPermission(c.Request.Context(), tenantID, userID, "cmdb:host", "admin")
	if err != nil {
		logger.Log.Error("检查 CMDB 管理员权限失败", zap.Error(err))
		return false
	}
	return isAdmin
}
```

注意：import 别名为 `userservice "devops-platform/internal/modules/user/service"` 以避免与本地 `service` 包冲突。

- [ ] **Step 2: 修改 TerminalConnect 添加权限校验**

在 `backend/internal/modules/cmdb/api/terminal.go` 的 `TerminalConnect` 函数中，在 `svc.GetConnectTarget(tenantID, hostID)` 调用之前（约第 207 行），添加权限校验：

```go
	// 主机级权限校验
	if !isCmdbAdmin(c, tenantID, userID) {
		permSvc := getPermissionService()
		allowed, _, err := permSvc.CheckPermission(tenantID, userID, hostID, "terminal")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "权限校验失败"})
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无该主机终端访问权限"})
			return
		}
	}
```

这段代码放在获取 `hostID` 之后、调用 `svc.GetConnectTarget` 之前。

- [ ] **Step 3: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add backend/internal/modules/cmdb/api/common.go \
        backend/internal/modules/cmdb/api/terminal.go
git commit -m "feat(cmdb): enforce host-level permission on terminal connect"
```

---

### Task 6: 主机列表权限过滤

**Files:**
- Modify: `backend/internal/modules/cmdb/repository/host.go`
- Modify: `backend/internal/modules/cmdb/service/host.go`
- Modify: `backend/internal/modules/cmdb/api/host.go`

- [ ] **Step 1: 修改 HostRepo.ListInTenant 支持 hostIDs 过滤**

在 `backend/internal/modules/cmdb/repository/host.go` 中修改 `ListInTenant` 方法签名，增加 `allowedHostIDs []uint` 参数：

```go
func (r *HostRepo) ListInTenant(tenantID uint, page, pageSize int, groupID uint, status, keyword string, allowedHostIDs []uint) ([]model.Host, int64, error) {
	var hosts []model.Host
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.Host{}), tenantID)

	if len(allowedHostIDs) > 0 {
		query = query.Where("id IN ?", allowedHostIDs)
	}
	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query = queryutil.ApplyKeywordLike(query, keyword, "hostname", "ip", "os_name", "description")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&hosts).Error; err != nil {
		return nil, 0, err
	}

	return hosts, total, nil
}
```

- [ ] **Step 2: 修改 HostService.ListInTenant 透传参数**

在 `backend/internal/modules/cmdb/service/host.go` 中修改 `ListInTenant`：

```go
func (s *HostService) ListInTenant(tenantID uint, page, pageSize int, groupID uint, status, keyword string, allowedHostIDs []uint) ([]model.Host, int64, error) {
	page, pageSize = s.normalizePage(page, pageSize)
	return s.repo.ListInTenant(tenantID, page, pageSize, groupID, status, keyword, allowedHostIDs)
}
```

- [ ] **Step 3: 修改 HostList API handler 添加权限过滤**

在 `backend/internal/modules/cmdb/api/host.go` 的 `HostList` 函数中，在调用 `svc.ListInTenant` 之前，添加权限过滤逻辑：

```go
func HostList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	groupID, _ := strconv.ParseUint(c.DefaultQuery("groupId", "0"), 10, 64)
	status := c.Query("status")
	keyword := c.Query("keyword")

	// 主机级权限过滤
	var allowedHostIDs []uint
	userID := c.GetUint("userID")
	if !isCmdbAdmin(c, tenantID, userID) {
		permSvc := getPermissionService()
		hosts, err := permSvc.MyHosts(tenantID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取主机权限失败"})
			return
		}
		allowedHostIDs = make([]uint, 0, len(hosts))
		for _, h := range hosts {
			allowedHostIDs = append(allowedHostIDs, h.HostID)
		}
		if len(allowedHostIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"list": []interface{}{}, "total": 0}})
			return
		}
	}

	svc := getHostService()
	hosts, total, err := svc.ListInTenant(tenantID, page, pageSize, uint(groupID), status, keyword, allowedHostIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"list": hosts, "total": total}})
}
```

注意：此处完整重写了 `HostList` 函数，因为需要重新组织逻辑。参考原函数中的参数解析逻辑保持不变。

- [ ] **Step 4: 修复其他调用 ListInTenant 的地方**

由于 `ListInTenant` 签名变更，其他调用处（如果有的话）需要传入 `nil` 作为 `allowedHostIDs`。搜索代码中的 `ListInTenant` 调用，全部加上 `, nil` 参数。

可能的调用位置：
- `backend/internal/modules/cmdb/service/cloud_sync.go` 中的 `MyHosts` 方法调用 `s.hostRepo.GetByGroupIDInTenant`，不涉及 `ListInTenant`，无需修改。
- 权限服务的 `MyHosts` 调用 `s.hostRepo.GetByGroupIDInTenant`，不涉及 `ListInTenant`，无需修改。

- [ ] **Step 5: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功，无错误

- [ ] **Step 6: 提交**

```bash
git add backend/internal/modules/cmdb/repository/host.go \
        backend/internal/modules/cmdb/service/host.go \
        backend/internal/modules/cmdb/api/host.go
git commit -m "feat(cmdb): host list filtered by user host-level permissions"
```

---

### Task 7: 主机详情权限校验

**Files:**
- Modify: `backend/internal/modules/cmdb/api/host.go`

- [ ] **Step 1: 修改 HostDetail 添加权限校验**

在 `backend/internal/modules/cmdb/api/host.go` 的 `HostDetail` 函数中，在获取主机详情之前添加权限检查：

```go
func HostDetail(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		idStr = c.Query("id")
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的主机 ID"})
		return
	}

	// 主机级权限校验
	userID := c.GetUint("userID")
	if !isCmdbAdmin(c, tenantID, userID) {
		permSvc := getPermissionService()
		allowed, _, err := permSvc.CheckPermission(tenantID, userID, uint(id), "view")
		if err != nil || !allowed {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "主机不存在或无访问权限"})
			return
		}
	}

	svc := getHostService()
	host, err := svc.GetByIDInTenant(tenantID, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "主机不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": host})
}
```

- [ ] **Step 2: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add backend/internal/modules/cmdb/api/host.go
git commit -m "feat(cmdb): enforce host-level view permission on host detail"
```

---

### Task 8: 前端终端连接 403 处理

**Files:**
- Modify: `frontend/src/views/Cmdb/HostList.vue`

- [ ] **Step 1: 在 handleTerminal 中添加 403 处理**

在 `frontend/src/views/Cmdb/HostList.vue` 中，找到 `handleTerminal` 函数（约第 233 行），在建立连接前先发一个 HTTP 请求检查权限。但考虑到终端连接是 WebSocket，更简洁的方式是监听 Terminal 组件的 `@error` 事件。

修改 `handleTerminal` 函数，在调用前先通过 API 检查权限：

```javascript
import { checkHostPermission } from '@/api/cmdb/host'
```

然后修改 `handleTerminal`：

```javascript
const handleTerminal = async (row) => {
  if (!row.credentialId) {
    ElMessage.warning('该主机未绑定凭据，无法建立终端连接')
    return
  }
  try {
    const res = await checkHostPermission(row.id, 'terminal')
    if (res.data?.code !== 200 || !res.data?.data?.allowed) {
      ElMessage.error('无该主机终端访问权限')
      return
    }
  } catch (e) {
    ElMessage.error('权限校验失败')
    return
  }
  terminalTitle.value = `主机终端 - ${row.hostname || row.ip}`
  terminalWsUrl.value = getTerminalConnectWsUrl(row.id)
  terminalDialogVisible.value = true
  nextTick(() => {
    terminalRef.value?.fit()
  })
}
```

- [ ] **Step 2: 添加权限检查 API 函数**

在 `frontend/src/api/cmdb/host.js` 中添加：

```javascript
export function checkHostPermission(hostId, action) {
  return request({
    url: '/cmdb/permissions/check',
    method: 'get',
    params: { host_id: hostId, action }
  })
}
```

注意：这里复用了已有的 `/cmdb/permissions/check` API 端点（见权限服务的 API 设计），该端点返回 `{allowed: true/false}`。

- [ ] **Step 3: 验证前端编译**

Run: `cd frontend && npm run build`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add frontend/src/views/Cmdb/HostList.vue \
        frontend/src/api/cmdb/host.js
git commit -m "feat(cmdb): frontend terminal permission check before connect"
```

---

## Self-Review Checklist

- [x] **Spec coverage:** 每个设计需求都有对应 Task
  - 云同步分页 → Task 1
  - 云同步启动 → Task 2
  - 录像清理 → Task 3
  - 权限种子 → Task 4
  - 终端连接权限 → Task 5
  - 主机列表过滤 → Task 6
  - 主机详情权限 → Task 7
  - 前端 403 → Task 8
- [x] **Placeholder scan:** 无 TBD/TODO/placeholder
- [x] **Type consistency:** 方法签名在各 Task 间一致（`ListInTenant` 新增参数、`CheckPermission` 四参数版本、`isCmdbAdmin` 两参数版本）
