# CMDB 云账号管理 - 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 CMDB 添加腾讯云账号管理，支持手动/定时同步 CVM/VPC/子网/安全组/CBS 资源，CVM 自动关联到 Host 记录。

**Architecture:** 新增 CloudAccount + CloudResource 模型。Service 层封装腾讯云 SDK 调用，实现资源同步逻辑。SecretId/SecretKey 使用 utils.Encrypt/Decrypt 加密存储。同步时 CVM 按 cloud_instance_id 匹配 Host 表。Host 模型已有 CloudAccountID 和 CloudInstanceID 字段，无需修改。

**Tech Stack:** Go/Gin/GORM (backend), Vue 3 + Element Plus (frontend), tencentcloud-sdk-go (Tencent Cloud API), utils.Encrypt/Decrypt (AES-256)

**Spec:** `docs/superpowers/specs/2026-04-18-cmdb-cloud-account-design.md`

---

## File Structure

**New files:**
- `backend/internal/modules/cmdb/model/cloud_account.go` — CloudAccount 模型
- `backend/internal/modules/cmdb/model/cloud_resource.go` — CloudResource 模型
- `backend/internal/modules/cmdb/repository/cloud.go` — 云账号/资源 CRUD + upsert
- `backend/internal/modules/cmdb/service/cloud_sync.go` — CRUD + 腾讯云同步 + 定时任务
- `backend/internal/modules/cmdb/api/cloud.go` — HTTP Handler
- `frontend/src/api/cmdb/cloud.js` — 前端 API 客户端
- `frontend/src/views/Cmdb/CloudAccountList.vue` — 云账号管理页面

**Modified files:**
- `backend/internal/bootstrap/db.go` — AutoMigrate + 权限种子
- `backend/internal/modules/cmdb/api/common.go` — cloudSvcInstance + getter
- `backend/routers/v1/cmdb.go` — 注册云账号路由
- `backend/config/defaults.go` — cloud 配置默认值
- `frontend/src/router/index.js` — 添加 /cmdb/cloud-accounts 路由
- `frontend/src/components/Layout/MainLayout.vue` — 侧边栏菜单项

---

### Task 1: Install SDK + Models + Bootstrap

**Files:**
- Modify: `backend/go.mod` — add tencentcloud SDK
- Create: `backend/internal/modules/cmdb/model/cloud_account.go`
- Create: `backend/internal/modules/cmdb/model/cloud_resource.go`
- Modify: `backend/internal/bootstrap/db.go`
- Modify: `backend/config/defaults.go`

- [ ] **Step 1: Install Tencent Cloud SDK**

```bash
cd backend
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312
```

Run: `cd backend && go mod tidy`
Expected: dependencies downloaded, go.sum updated

- [ ] **Step 2: Create CloudAccount model**

```go
// backend/internal/modules/cmdb/model/cloud_account.go
package model

import (
	"time"

	"gorm.io/gorm"
)

type CloudAccount struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TenantID      uint           `gorm:"not null;uniqueIndex:uk_cloud_tenant_provider_name,priority:1" json:"tenantId"`
	Name          string         `gorm:"size:100;not null;uniqueIndex:uk_cloud_tenant_provider_name,priority:2" json:"name"`
	Provider      string         `gorm:"size:20;not null;uniqueIndex:uk_cloud_tenant_provider_name,priority:3" json:"provider"`
	SecretID      string         `gorm:"size:500;not null" json:"-"`
	SecretKey     string         `gorm:"size:500;not null" json:"-"`
	Status        string         `gorm:"size:20;default:active" json:"status"`
	LastSyncAt    *time.Time     `json:"lastSyncAt"`
	LastSyncError string         `gorm:"type:text" json:"lastSyncError"`
	SyncInterval  int            `gorm:"default:60" json:"syncInterval"`
	Description   string         `gorm:"size:500" json:"description"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
```

Note: `SecretID` and `SecretKey` use `json:"-"` so they are never returned by API.

- [ ] **Step 3: Create CloudResource model**

```go
// backend/internal/modules/cmdb/model/cloud_resource.go
package model

import (
	"time"

	"gorm.io/gorm"
)

type CloudResource struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	TenantID       uint           `gorm:"not null;index" json:"tenantId"`
	CloudAccountID uint           `gorm:"not null;uniqueIndex:uk_cloud_res,priority:1;index" json:"cloudAccountId"`
	ResourceType   string         `gorm:"size:30;not null;uniqueIndex:uk_cloud_res,priority:2;index:idx_cloud_res_type" json:"resourceType"`
	ResourceID     string         `gorm:"size:100;not null;uniqueIndex:uk_cloud_res,priority:3" json:"resourceId"`
	Region         string         `gorm:"size:50;index:idx_cloud_res_region" json:"region"`
	Zone           string         `gorm:"size:50" json:"zone"`
	Name           string         `gorm:"size:200" json:"name"`
	State          string         `gorm:"size:30" json:"state"`
	Spec           string         `gorm:"type:text" json:"spec"`
	SyncedAt       time.Time      `json:"syncedAt"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
```

Note: `Spec` is stored as JSON string, not gorm datatypes.JSON, to keep it simple.

- [ ] **Step 4: Register in AutoMigrate + seed permissions**

In `backend/internal/bootstrap/db.go`, add to the AutoMigrate call (after `&cmdbModel.HostPermission{}`):

```go
		&cmdbModel.CloudAccount{},
		&cmdbModel.CloudResource{},
```

In the same file, add to the `permissions` slice in `seedPermissions()` (after the cmdb:permission entries):

```go
		// 云账号管理
		{Name: "查看云账号", Resource: "cmdb:cloud", Action: "list", Description: "查看云账号列表"},
		{Name: "查看云账号详情", Resource: "cmdb:cloud", Action: "get", Description: "查看云账号详情"},
		{Name: "添加云账号", Resource: "cmdb:cloud", Action: "create", Description: "添加云账号"},
		{Name: "更新云账号", Resource: "cmdb:cloud", Action: "update", Description: "更新云账号"},
		{Name: "删除云账号", Resource: "cmdb:cloud", Action: "delete", Description: "删除云账号"},
		{Name: "同步云资源", Resource: "cmdb:cloud", Action: "sync", Description: "手动触发云资源同步"},
```

- [ ] **Step 5: Add config defaults**

In `backend/config/defaults.go`, find the `SetDefaults` function and add before the closing `}`:

```go
	v.SetDefault("cloud.sync_concurrency", 5)
	v.SetDefault("cloud.sync_timeout", 300)
	v.SetDefault("cloud.default_regions", "ap-guangzhou,ap-shanghai,ap-beijing")
```

- [ ] **Step 6: Build**

Run: `cd backend && go build ./...`
Expected: compiles with no errors

- [ ] **Step 7: Commit**

```bash
git add backend/go.mod backend/go.sum backend/internal/modules/cmdb/model/cloud_account.go backend/internal/modules/cmdb/model/cloud_resource.go backend/internal/bootstrap/db.go backend/config/defaults.go
git commit -m "feat(cmdb): add CloudAccount/CloudResource models and Tencent Cloud SDK

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 2: Repository

**Files:**
- Create: `backend/internal/modules/cmdb/repository/cloud.go`

- [ ] **Step 1: Create cloud repository**

```go
// backend/internal/modules/cmdb/repository/cloud.go
package repository

import (
	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CloudRepo struct {
	db *gorm.DB
}

func NewCloudRepo(db *gorm.DB) *CloudRepo {
	return &CloudRepo{db: db}
}

func (r *CloudRepo) scopeInTenant(query *gorm.DB, tenantID uint) *gorm.DB {
	return query.Where("tenant_id = ?", tenantID)
}

// CloudAccount CRUD

func (r *CloudRepo) CreateAccount(account *model.CloudAccount) error {
	return r.db.Create(account).Error
}

func (r *CloudRepo) GetAccountByIDInTenant(tenantID, id uint) (*model.CloudAccount, error) {
	var account model.CloudAccount
	if err := r.scopeInTenant(r.db, tenantID).Where("id = ?", id).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *CloudRepo) UpdateAccount(account *model.CloudAccount) error {
	return r.db.Save(account).Error
}

func (r *CloudRepo) DeleteAccountInTenant(tenantID, id uint) error {
	return r.scopeInTenant(r.db, tenantID).Where("id = ?", id).Delete(&model.CloudAccount{}).Error
}

func (r *CloudRepo) ListAccountsInTenant(tenantID uint, page, pageSize int, status string) ([]model.CloudAccount, int64, error) {
	var accounts []model.CloudAccount
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.CloudAccount{}), tenantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&accounts).Error; err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

func (r *CloudRepo) ListAllActiveAccounts() ([]model.CloudAccount, error) {
	var accounts []model.CloudAccount
	if err := r.db.Where("status = ?", "active").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// CloudResource CRUD

func (r *CloudRepo) UpsertResource(resource *model.CloudResource) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "cloud_account_id"},
			{Name: "resource_type"},
			{Name: "resource_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "region", "zone", "state", "spec", "synced_at", "updated_at",
		}),
	}).Create(resource).Error
}

func (r *CloudRepo) ListResourcesByAccountInTenant(tenantID, accountID uint, resourceType string, page, pageSize int) ([]model.CloudResource, int64, error) {
	var resources []model.CloudResource
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.CloudResource{}), tenantID).
		Where("cloud_account_id = ?", accountID)
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("synced_at DESC").Offset(offset).Limit(pageSize).Find(&resources).Error; err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

func (r *CloudRepo) DeleteResourcesByAccount(accountID uint) error {
	return r.db.Where("cloud_account_id = ?", accountID).Delete(&model.CloudResource{}).Error
}

func (r *CloudRepo) GetHostByCloudInstanceID(tenantID uint, instanceID string) (*model.Host, error) {
	var host model.Host
	query := r.scopeInTenant(r.db, tenantID).Where("cloud_instance_id = ?", instanceID)
	if err := query.First(&host).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *CloudRepo) CreateHost(host *model.Host) error {
	return r.db.Create(host).Error
}

func (r *CloudRepo) UpdateHost(host *model.Host) error {
	return r.db.Save(host).Error
}
```

- [ ] **Step 2: Build**

Run: `cd backend && go build ./...`
Expected: compiles with no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/modules/cmdb/repository/cloud.go
git commit -m "feat(cmdb): add cloud repository with CRUD and resource upsert

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 3: Service (CRUD + Tencent Cloud Sync)

**Files:**
- Create: `backend/internal/modules/cmdb/service/cloud_sync.go`

This is the most complex task. The service handles CRUD operations AND the full Tencent Cloud sync logic.

- [ ] **Step 1: Create cloud sync service**

```go
// backend/internal/modules/cmdb/service/cloud_sync.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"devops-platform/config"
	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
	"devops-platform/internal/pkg/utils"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	"gorm.io/gorm"
)

type CloudAccountService struct {
	repo *repository.CloudRepo
	db   *gorm.DB
}

func NewCloudAccountService(db *gorm.DB) *CloudAccountService {
	return &CloudAccountService{repo: repository.NewCloudRepo(db), db: db}
}

func (s *CloudAccountService) normalizePage(page, pageSize int) (int, int) {
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

// Request structs

type CloudAccountCreateRequest struct {
	Name         string `json:"name" binding:"required"`
	Provider     string `json:"provider"`
	SecretID     string `json:"secretId" binding:"required"`
	SecretKey    string `json:"secretKey" binding:"required"`
	SyncInterval int    `json:"syncInterval"`
	Description  string `json:"description"`
}

type CloudAccountUpdateRequest struct {
	ID           uint   `json:"id" binding:"required"`
	Name         string `json:"name"`
	SecretID     string `json:"secretId"`
	SecretKey    string `json:"secretKey"`
	SyncInterval int    `json:"syncInterval"`
	Description  string `json:"description"`
}

// CRUD methods

func (s *CloudAccountService) ListInTenant(tenantID uint, page, pageSize int, status string) ([]model.CloudAccount, int64, error) {
	page, pageSize = s.normalizePage(page, pageSize)
	return s.repo.ListAccountsInTenant(tenantID, page, pageSize, status)
}

func (s *CloudAccountService) GetByIDInTenant(tenantID, id uint) (*model.CloudAccount, error) {
	return s.repo.GetAccountByIDInTenant(tenantID, id)
}

func (s *CloudAccountService) CreateInTenant(tenantID uint, req *CloudAccountCreateRequest) (*model.CloudAccount, error) {
	provider := req.Provider
	if provider == "" {
		provider = "tencent"
	}

	encryptedSecretID, err := utils.Encrypt(req.SecretID)
	if err != nil {
		return nil, errors.New("加密 SecretId 失败")
	}
	encryptedSecretKey, err := utils.Encrypt(req.SecretKey)
	if err != nil {
		return nil, errors.New("加密 SecretKey 失败")
	}

	syncInterval := req.SyncInterval
	if syncInterval <= 0 {
		syncInterval = 60
	}

	account := &model.CloudAccount{
		TenantID:     tenantID,
		Name:         req.Name,
		Provider:     provider,
		SecretID:     encryptedSecretID,
		SecretKey:    encryptedSecretKey,
		Status:       "active",
		SyncInterval: syncInterval,
		Description:  req.Description,
	}

	if err := s.repo.CreateAccount(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *CloudAccountService) UpdateInTenant(tenantID uint, req *CloudAccountUpdateRequest) (*model.CloudAccount, error) {
	account, err := s.repo.GetAccountByIDInTenant(tenantID, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("云账号不存在")
		}
		return nil, err
	}

	if req.Name != "" {
		account.Name = req.Name
	}
	if req.SecretID != "" {
		encrypted, err := utils.Encrypt(req.SecretID)
		if err != nil {
			return nil, errors.New("加密 SecretId 失败")
		}
		account.SecretID = encrypted
	}
	if req.SecretKey != "" {
		encrypted, err := utils.Encrypt(req.SecretKey)
		if err != nil {
			return nil, errors.New("加密 SecretKey 失败")
		}
		account.SecretKey = encrypted
	}
	if req.SyncInterval > 0 {
		account.SyncInterval = req.SyncInterval
	}
	if req.Description != "" {
		account.Description = req.Description
	}

	if err := s.repo.UpdateAccount(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *CloudAccountService) DeleteInTenant(tenantID, id uint) error {
	account, err := s.repo.GetAccountByIDInTenant(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("云账号不存在")
		}
		return err
	}
	// Delete associated resources
	_ = s.repo.DeleteResourcesByAccount(account.ID)
	return s.repo.DeleteAccountInTenant(tenantID, id)
}

func (s *CloudAccountService) ListResourcesInTenant(tenantID, accountID uint, resourceType string, page, pageSize int) ([]model.CloudResource, int64, error) {
	page, pageSize = s.normalizePage(page, pageSize)
	return s.repo.ListResourcesByAccountInTenant(tenantID, accountID, resourceType, page, pageSize)
}

// Sync methods

func (s *CloudAccountService) SyncAccount(tenantID, accountID uint) error {
	account, err := s.repo.GetAccountByIDInTenant(tenantID, accountID)
	if err != nil {
		return errors.New("云账号不存在")
	}

	secretID, err := utils.Decrypt(account.SecretID)
	if err != nil {
		return errors.New("解密 SecretId 失败")
	}
	secretKey, err := utils.Decrypt(account.SecretKey)
	if err != nil {
		return errors.New("解密 SecretKey 失败")
	}

	regions := getDefaultRegions()
	var syncErrors []string

	for _, region := range regions {
		if err := s.syncRegion(account, secretID, secretKey, region); err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("%s: %s", region, err.Error()))
		}
	}

	now := time.Now()
	account.LastSyncAt = &now

	if len(syncErrors) > 0 {
		account.Status = "error"
		account.LastSyncError = strings.Join(syncErrors, "; ")
	} else {
		account.Status = "active"
		account.LastSyncError = ""
	}

	return s.repo.UpdateAccount(account)
}

func getDefaultRegions() []string {
	regionsStr := config.Cfg.GetString("cloud.default_regions")
	if regionsStr == "" {
		return []string{"ap-guangzhou", "ap-shanghai", "ap-beijing"}
	}
	return strings.Split(regionsStr, ",")
}

func (s *CloudAccountService) syncRegion(account *model.CloudAccount, secretID, secretKey, region string) error {
	credential := common.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "GET"

	// Sync CVM
	if err := s.syncCVM(credential, cpf, account, region); err != nil {
		return fmt.Errorf("同步 CVM 失败: %w", err)
	}

	// Sync VPC
	if err := s.syncVPC(credential, cpf, account, region); err != nil {
		return fmt.Errorf("同步 VPC 失败: %w", err)
	}

	// Sync Subnets
	if err := s.syncSubnets(credential, cpf, account, region); err != nil {
		return fmt.Errorf("同步子网失败: %w", err)
	}

	// Sync Security Groups
	if err := s.syncSecurityGroups(credential, cpf, account, region); err != nil {
		return fmt.Errorf("同步安全组失败: %w", err)
	}

	// Sync CBS
	if err := s.syncCBS(credential, cpf, account, region); err != nil {
		return fmt.Errorf("同步云硬盘失败: %w", err)
	}

	return nil
}

func (s *CloudAccountService) syncCVM(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := cvm.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	request := cvm.NewDescribeInstancesRequest()
	request.Limit = common.Int64Ptr(100)

	response, err := client.DescribeInstances(request)
	if err != nil {
		return err
	}

	if response == nil || response.Response == nil {
		return nil
	}

	for _, instance := range response.Response.InstanceSet {
		instanceID := *instance.InstanceId
		instanceName := ""
		if instance.InstanceName != nil {
			instanceName = *instance.InstanceName
		}

		privateIP := ""
		if len(instance.PrivateIpAddresses) > 0 && instance.PrivateIpAddresses[0] != nil {
			privateIP = *instance.PrivateIpAddresses[0]
		}

		state := ""
		if instance.InstanceState != nil {
			state = *instance.InstanceState
		}

		zone := ""
		if instance.Placement != nil && instance.Placement.Zone != nil {
			zone = *instance.Placement.Zone
		}

		osName := ""
		if instance.OsName != nil {
			osName = *instance.OsName
		}

		cpu := 0
		if instance.CPU != nil {
			cpu = int(*instance.CPU)
		}

		memory := 0
		if instance.Memory != nil {
			memory = int(*instance.Memory)
		}

		// Save to CloudResource
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

		// Upsert Host record
		existing, err := s.repo.GetHostByCloudInstanceID(account.TenantID, instanceID)
		if err == nil && existing != nil {
			// Update existing host
			existing.Hostname = instanceName
			if privateIP != "" {
				existing.Ip = privateIP
			}
			existing.OsName = osName
			existing.CloudAccountID = &account.ID
			_ = s.repo.UpdateHost(existing)
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new host
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

	return nil
}

func (s *CloudAccountService) syncVPC(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := vpc.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	request := vpc.NewDescribeVpcsRequest()
	request.Limit = common.Int64Ptr(100)

	response, err := client.DescribeVpcs(request)
	if err != nil {
		return err
	}

	if response == nil || response.Response == nil {
		return nil
	}

	for _, v := range response.Response.VpcSet {
		vpcID := *v.VpcId
		name := ""
		if v.VpcName != nil {
			name = *v.VpcName
		}
		cidr := ""
		if v.CidrBlock != nil {
			cidr = *v.CidrBlock
		}
		state := ""
		if v.State != nil {
			state = *v.State
		}

		specJSON, _ := json.Marshal(map[string]string{"cidr": cidr})
		resource := &model.CloudResource{
			TenantID:       account.TenantID,
			CloudAccountID: account.ID,
			ResourceType:   "vpc",
			ResourceID:     vpcID,
			Region:         region,
			Name:           name,
			State:          state,
			Spec:           string(specJSON),
			SyncedAt:       time.Now(),
		}
		_ = s.repo.UpsertResource(resource)
	}

	return nil
}

func (s *CloudAccountService) syncSubnets(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := vpc.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	request := vpc.NewDescribeSubnetsRequest()
	request.Limit = common.Int64Ptr(100)

	response, err := client.DescribeSubnets(request)
	if err != nil {
		return err
	}

	if response == nil || response.Response == nil {
		return nil
	}

	for _, sub := range response.Response.SubnetSet {
		subnetID := *sub.SubnetId
		name := ""
		if sub.SubnetName != nil {
			name = *sub.SubnetName
		}
		cidr := ""
		if sub.CidrBlock != nil {
			cidr = *sub.CidrBlock
		}
		vpcID := ""
		if sub.VpcId != nil {
			vpcID = *sub.VpcId
		}
		zone := ""
		if sub.Zone != nil {
			zone = *sub.Zone
		}
		state := ""
		if sub.IsDefault != nil {
			if *sub.IsDefault {
				state = "available"
			}
		}

		specJSON, _ := json.Marshal(map[string]string{"cidr": cidr, "vpc_id": vpcID})
		resource := &model.CloudResource{
			TenantID:       account.TenantID,
			CloudAccountID: account.ID,
			ResourceType:   "subnet",
			ResourceID:     subnetID,
			Region:         region,
			Zone:           zone,
			Name:           name,
			State:          state,
			Spec:           string(specJSON),
			SyncedAt:       time.Now(),
		}
		_ = s.repo.UpsertResource(resource)
	}

	return nil
}

func (s *CloudAccountService) syncSecurityGroups(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := vpc.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	request := vpc.NewDescribeSecurityGroupsRequest()
	request.Limit = common.Int64Ptr(100)

	response, err := client.DescribeSecurityGroups(request)
	if err != nil {
		return err
	}

	if response == nil || response.Response == nil {
		return nil
	}

	for _, sg := range response.Response.SecurityGroupSet {
		sgID := *sg.SecurityGroupId
		name := ""
		if sg.SecurityGroupName != nil {
			name = *sg.SecurityGroupName
		}
		desc := ""
		if sg.SecurityGroupDesc != nil {
			desc = *sg.SecurityGroupDesc
		}

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

	return nil
}

func (s *CloudAccountService) syncCBS(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := cbs.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	request := cbs.NewDescribeDisksRequest()
	request.Limit = common.Int64Ptr(100)

	response, err := client.DescribeDisks(request)
	if err != nil {
		return err
	}

	if response == nil || response.Response == nil {
		return nil
	}

	for _, disk := range response.Response.DiskSet {
		diskID := *disk.DiskId
		name := ""
		if disk.DiskName != nil {
			name = *disk.DiskName
		}
		state := ""
		if disk.DiskState != nil {
			state = *disk.DiskState
		}
		diskType := ""
		if disk.DiskType != nil {
			diskType = *disk.DiskType
		}
		diskSize := int64(0)
		if disk.DiskSize != nil {
			diskSize = *disk.DiskSize
		}
		zone := ""
		if disk.Placement != nil && disk.Placement.Zone != nil {
			zone = *disk.Placement.Zone
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

	return nil
}

// Scheduled sync

var (
	syncMu       sync.Mutex
	syncCancelFn context.CancelFunc
)

func (s *CloudAccountService) StartScheduledSync() {
	syncMu.Lock()
	defer syncMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	syncCancelFn = cancel

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runScheduledSync()
			}
		}
	}()
}

func (s *CloudAccountService) StopScheduledSync() {
	syncMu.Lock()
	defer syncMu.Unlock()

	if syncCancelFn != nil {
		syncCancelFn()
		syncCancelFn = nil
	}
}

func (s *CloudAccountService) runScheduledSync() {
	accounts, err := s.repo.ListAllActiveAccounts()
	if err != nil {
		return
	}

	now := time.Now()
	for _, account := range accounts {
		if account.LastSyncAt == nil || now.Sub(*account.LastSyncAt) >= time.Duration(account.SyncInterval)*time.Minute {
			_ = s.SyncAccount(account.TenantID, account.ID)
		}
	}
}
```

- [ ] **Step 2: Build**

Run: `cd backend && go build ./...`
Expected: compiles with no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/modules/cmdb/service/cloud_sync.go
git commit -m "feat(cmdb): add cloud sync service with Tencent Cloud integration

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 4: API Handlers + Routes

**Files:**
- Create: `backend/internal/modules/cmdb/api/cloud.go`
- Modify: `backend/internal/modules/cmdb/api/common.go` — add cloudSvcInstance + getter
- Modify: `backend/routers/v1/cmdb.go` — register cloud routes

- [ ] **Step 1: Update common.go**

Add `cloudSvcInstance *service.CloudAccountService` to the var block in `backend/internal/modules/cmdb/api/common.go` (after `permSvcInstance`).

Add `cloudSvcInstance = nil` to `SetDB` (after `permSvcInstance = nil`).

Add getter after `getPermissionService()`:

```go
func getCloudAccountService() *service.CloudAccountService {
	cmdbMu.Lock()
	defer cmdbMu.Unlock()
	if cloudSvcInstance != nil {
		return cloudSvcInstance
	}
	if cmdbDB == nil {
		return nil
	}
	cloudSvcInstance = service.NewCloudAccountService(cmdbDB)
	return cloudSvcInstance
}
```

- [ ] **Step 2: Create cloud API handler**

```go
// backend/internal/modules/cmdb/api/cloud.go
package api

import (
	"errors"
	"net/http"
	"strconv"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CloudAccountList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	svc := getCloudAccountService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	accounts, total, err := svc.ListInTenant(tenantID, page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取云账号列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "获取成功",
		"data":     accounts,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func CloudAccountDetail(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的 ID"})
		return
	}

	svc := getCloudAccountService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	account, err := svc.GetByIDInTenant(tenantID, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "云账号不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取云账号详情失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    account,
	})
}

func CloudAccountCreate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.CloudAccountCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getCloudAccountService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	account, err := svc.CreateInTenant(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "创建云账号失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    account,
	})
}

func CloudAccountUpdate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.CloudAccountUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getCloudAccountService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	account, err := svc.UpdateInTenant(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "更新云账号失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    account,
	})
}

func CloudAccountDelete(c *gin.Context) {
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

	svc := getCloudAccountService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	if err := svc.DeleteInTenant(tenantID, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "删除云账号失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

func CloudAccountSync(c *gin.Context) {
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

	svc := getCloudAccountService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	if err := svc.SyncAccount(tenantID, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "同步失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "同步成功",
	})
}

func CloudResourceList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	accountID, _ := strconv.ParseUint(c.DefaultQuery("cloudAccountId", "0"), 10, 32)
	if accountID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少 cloudAccountId 参数"})
		return
	}

	resourceType := c.Query("resourceType")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	svc := getCloudAccountService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	resources, total, err := svc.ListResourcesInTenant(tenantID, uint(accountID), resourceType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取云资源失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "获取成功",
		"data":     resources,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
```

- [ ] **Step 3: Register routes**

Edit `backend/routers/v1/cmdb.go`. Add after the existing permission middleware definitions:

```go
	cloudListPerm := middleware.RequirePermission("cmdb:cloud", "list")
	cloudGetPerm := middleware.RequirePermission("cmdb:cloud", "get")
	cloudCreatePerm := middleware.RequirePermission("cmdb:cloud", "create")
	cloudUpdatePerm := middleware.RequirePermission("cmdb:cloud", "update")
	cloudDeletePerm := middleware.RequirePermission("cmdb:cloud", "delete")
	cloudSyncPerm := middleware.RequirePermission("cmdb:cloud", "sync")
```

Add routes inside the route block (after permission routes):

```go
			// 云账号管理
			g.GET("/cloud-account/list", cloudListPerm, api.CloudAccountList)
			g.GET("/cloud-account/detail", cloudGetPerm, api.CloudAccountDetail)
			g.POST("/cloud-account/create", cloudCreatePerm, middleware.SetAuditOperation("创建云账号"), api.CloudAccountCreate)
			g.POST("/cloud-account/update", cloudUpdatePerm, middleware.SetAuditOperation("更新云账号"), api.CloudAccountUpdate)
			g.POST("/cloud-account/delete", cloudDeletePerm, middleware.SetAuditOperation("删除云账号"), api.CloudAccountDelete)
			g.POST("/cloud-account/sync", cloudSyncPerm, middleware.SetAuditOperation("同步云资源"), api.CloudAccountSync)
			g.GET("/cloud-account/resources", cloudListPerm, api.CloudResourceList)
```

- [ ] **Step 4: Build**

Run: `cd backend && go build ./...`
Expected: compiles with no errors

- [ ] **Step 5: Commit**

```bash
git add backend/internal/modules/cmdb/api/cloud.go backend/internal/modules/cmdb/api/common.go backend/routers/v1/cmdb.go
git commit -m "feat(cmdb): add cloud account API handlers and routes

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 5: Frontend API + Route + Sidebar

**Files:**
- Create: `frontend/src/api/cmdb/cloud.js`
- Modify: `frontend/src/router/index.js`
- Modify: `frontend/src/components/Layout/MainLayout.vue`

- [ ] **Step 1: Create API client**

```javascript
// frontend/src/api/cmdb/cloud.js
import request from '../request'

export const getCloudAccountList = (params) => request.get('/cmdb/cloud-account/list', { params })
export const getCloudAccountDetail = (params) => request.get('/cmdb/cloud-account/detail', { params })
export const createCloudAccount = (data) => request.post('/cmdb/cloud-account/create', data)
export const updateCloudAccount = (data) => request.post('/cmdb/cloud-account/update', data)
export const deleteCloudAccount = (data) => request.post('/cmdb/cloud-account/delete', data)
export const syncCloudAccount = (data) => request.post('/cmdb/cloud-account/sync', data)
export const getCloudResources = (params) => request.get('/cmdb/cloud-account/resources', { params })
```

- [ ] **Step 2: Add route**

In `frontend/src/router/index.js`, add after the permissions route:

```javascript
      {
        path: 'cmdb/cloud-accounts',
        component: () => import('../views/Cmdb/CloudAccountList.vue')
      }
```

- [ ] **Step 3: Add sidebar item**

In `frontend/src/components/Layout/MainLayout.vue`, add after "权限配置" menu item:

```html
          <el-menu-item index="/cmdb/cloud-accounts">云账号</el-menu-item>
```

- [ ] **Step 4: Build**

Run: `cd frontend && npm run build`
Expected: build succeeds

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/cmdb/cloud.js frontend/src/router/index.js frontend/src/components/Layout/MainLayout.vue
git commit -m "feat(cmdb): add cloud account frontend API, route, and sidebar

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 6: Frontend CloudAccountList.vue

**Files:**
- Create: `frontend/src/views/Cmdb/CloudAccountList.vue`

- [ ] **Step 1: Create CloudAccountList.vue**

```vue
<template>
  <div class="page-container">
    <div class="page-header">
      <h3>云账号管理</h3>
      <el-button type="primary" @click="showCreateDialog">添加云账号</el-button>
    </div>
    <div class="toolbar">
      <el-select v-model="filterStatus" placeholder="全部状态" clearable style="width: 150px" @change="fetchData">
        <el-option label="正常" value="active" />
        <el-option label="错误" value="error" />
      </el-select>
    </div>
    <el-table :data="tableData" stripe v-loading="loading">
      <el-table-column prop="name" label="账号名称" min-width="150" />
      <el-table-column prop="provider" label="云厂商" width="100">
        <template #default="{ row }">{{ providerLabel(row.provider) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
            {{ row.status === 'active' ? '正常' : '错误' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="最后同步" width="180">
        <template #default="{ row }">{{ row.lastSyncAt || '-' }}</template>
      </el-table-column>
      <el-table-column label="同步间隔" width="100">
        <template #default="{ row }">{{ row.syncInterval }}分钟</template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleSync(row)" :loading="syncingId === row.id">同步</el-button>
          <el-button link type="primary" size="small" @click="showResources(row)">资源</el-button>
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

    <!-- Create/Edit dialog -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑云账号' : '添加云账号'" width="500px" destroy-on-close>
      <el-form :model="form" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="账号名称" prop="name">
          <el-input v-model="form.name" placeholder="输入账号名称" />
        </el-form-item>
        <el-form-item label="SecretId" prop="secretId">
          <el-input v-model="form.secretId" :placeholder="isEdit ? '不修改则留空' : '输入 SecretId'" />
        </el-form-item>
        <el-form-item label="SecretKey" prop="secretKey">
          <el-input v-model="form.secretKey" type="password" show-password :placeholder="isEdit ? '不修改则留空' : '输入 SecretKey'" />
        </el-form-item>
        <el-form-item label="同步间隔" prop="syncInterval">
          <el-input-number v-model="form.syncInterval" :min="5" :max="1440" />
          <span style="margin-left: 8px; color: #909399">分钟</span>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- Resources dialog -->
    <el-dialog v-model="resourceDialogVisible" :title="`云资源 - ${resourceAccountName}`" width="80%" top="5vh" destroy-on-close>
      <el-tabs v-model="resourceType" @tab-change="fetchResources">
        <el-tab-pane label="CVM" name="cvm" />
        <el-tab-pane label="VPC" name="vpc" />
        <el-tab-pane label="子网" name="subnet" />
        <el-tab-pane label="安全组" name="security_group" />
        <el-tab-pane label="云硬盘" name="cbs" />
      </el-tabs>
      <el-table :data="resourceData" stripe v-loading="resourceLoading" max-height="400">
        <el-table-column prop="resourceId" label="资源 ID" min-width="180" show-overflow-tooltip />
        <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="region" label="地域" width="120" />
        <el-table-column prop="zone" label="可用区" width="120" />
        <el-table-column prop="state" label="状态" width="100" />
        <el-table-column label="规格" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ formatSpec(row.spec) }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCloudAccountList, createCloudAccount, updateCloudAccount, deleteCloudAccount, syncCloudAccount, getCloudResources } from '@/api/cmdb/cloud'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const filterStatus = ref('')
const syncingId = ref(0)

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()

const form = reactive({
  id: 0,
  name: '',
  secretId: '',
  secretKey: '',
  syncInterval: 60,
  description: ''
})

const formRules = {
  name: [{ required: true, message: '请输入账号名称', trigger: 'blur' }]
}

const resourceDialogVisible = ref(false)
const resourceAccountName = ref('')
const resourceAccountId = ref(0)
const resourceType = ref('cvm')
const resourceData = ref([])
const resourceLoading = ref(false)

const providerLabel = (p) => {
  const map = { tencent: '腾讯云', aliyun: '阿里云', aws: 'AWS' }
  return map[p] || p
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = { page: page.value, pageSize: pageSize.value }
    if (filterStatus.value) params.status = filterStatus.value
    const res = await getCloudAccountList(params)
    tableData.value = res.data || []
    total.value = res.total || 0
  } catch (e) {
    ElMessage.error('获取云账号列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  form.id = 0
  form.name = ''
  form.secretId = ''
  form.secretKey = ''
  form.syncInterval = 60
  form.description = ''
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.id = row.id
  form.name = row.name
  form.secretId = ''
  form.secretKey = ''
  form.syncInterval = row.syncInterval
  form.description = row.description
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try { await formRef.value.validate() } catch { return }
  submitting.value = true
  try {
    if (isEdit.value) {
      const payload = { id: form.id, name: form.name, syncInterval: form.syncInterval, description: form.description }
      if (form.secretId) payload.secretId = form.secretId
      if (form.secretKey) payload.secretKey = form.secretKey
      await updateCloudAccount(payload)
      ElMessage.success('更新成功')
    } else {
      await createCloudAccount({
        name: form.name,
        secretId: form.secretId,
        secretKey: form.secretKey,
        syncInterval: form.syncInterval,
        description: form.description
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  try { await ElMessageBox.confirm('确定删除该云账号？关联的云资源也会被删除。', '确认', { type: 'warning' }) } catch { return }
  try {
    await deleteCloudAccount({ id: row.id })
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '删除失败')
  }
}

const handleSync = async (row) => {
  syncingId.value = row.id
  try {
    await syncCloudAccount({ id: row.id })
    ElMessage.success('同步成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '同步失败')
  } finally {
    syncingId.value = 0
  }
}

const showResources = (row) => {
  resourceAccountId.value = row.id
  resourceAccountName.value = row.name
  resourceType.value = 'cvm'
  resourceDialogVisible.value = true
  fetchResources()
}

const fetchResources = async () => {
  resourceLoading.value = true
  try {
    const res = await getCloudResources({
      cloudAccountId: resourceAccountId.value,
      resourceType: resourceType.value,
      page: 1,
      pageSize: 100
    })
    resourceData.value = res.data || []
  } catch {
    resourceData.value = []
  } finally {
    resourceLoading.value = false
  }
}

const formatSpec = (spec) => {
  if (!spec) return '-'
  try {
    const obj = typeof spec === 'string' ? JSON.parse(spec) : spec
    return Object.entries(obj).map(([k, v]) => `${k}: ${v}`).join(', ')
  } catch {
    return spec
  }
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
```

- [ ] **Step 2: Build**

Run: `cd frontend && npm run build`
Expected: build succeeds

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/Cmdb/CloudAccountList.vue
git commit -m "feat(cmdb): add CloudAccountList page with sync and resource viewer

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 7: API Verification

**Files:** None (verification only)

- [ ] **Step 1: Start backend**

Run: `cd backend && go build ./... && DEVOPS_SERVER_PORT=8001 go run cmd/server/main.go`
Wait for "服务启动" log. Verify `cloud_accounts` and `cloud_resources` tables created.

- [ ] **Step 2: Login**

```bash
TOKEN=$(curl -s http://localhost:8001/api/v1/user/login -H 'Content-Type: application/json' -d '{"tenantCode":"default","username":"admin","password":"admin@2026"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "Token: $TOKEN"
```

- [ ] **Step 3: Test empty list**

```bash
curl -s http://localhost:8001/api/v1/cmdb/cloud-account/list -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```
Expected: `{"code":200,"data":[],"total":0}`

- [ ] **Step 4: Test create account**

```bash
curl -s http://localhost:8001/api/v1/cmdb/cloud-account/create -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"name":"test-tencent","secretId":"test-id","secretKey":"test-key"}' | python3 -m json.tool
```
Expected: `{"code":200,"data":{"id":1,"name":"test-tencent","status":"active",...}}`

- [ ] **Step 5: Test update**

```bash
curl -s http://localhost:8001/api/v1/cmdb/cloud-account/update -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"id":1,"syncInterval":30}' | python3 -m json.tool
```
Expected: `{"code":200,"data":{"syncInterval":30,...}}`

- [ ] **Step 6: Test list with data**

```bash
curl -s "http://localhost:8001/api/v1/cmdb/cloud-account/list?page=1&pageSize=10" -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```
Expected: `total: 1`

- [ ] **Step 7: Test delete**

```bash
curl -s http://localhost:8001/api/v1/cmdb/cloud-account/delete -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"id":1}' | python3 -m json.tool
```
Expected: `{"code":200,"message":"删除成功"}`

- [ ] **Step 8: Start frontend dev and verify page**

Run: `cd frontend && npm run dev`
Open browser at the CMDB > 云账号 page. Verify the list, create dialog, and resource dialog render correctly.

- [ ] **Step 9: Final commit if fixes needed**

```bash
git add -A
git commit -m "fix(cmdb): address cloud account verification findings

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```
