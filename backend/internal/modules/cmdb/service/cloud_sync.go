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
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
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

func safeDereferenceString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeDereferenceInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func safeDereferenceUint64(i *uint64) uint64 {
	if i == nil {
		return 0
	}
	return *i
}

func (s *CloudAccountService) syncCVM(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := cvm.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	return s.paginateSync(func(offset int64) (int, error) {
		request := cvm.NewDescribeInstancesRequest()
		request.Limit = common.Int64Ptr(cloudSyncPageSize)
		request.Offset = common.Int64Ptr(offset)

		response, err := client.DescribeInstances(request)
		if err != nil {
			return 0, err
		}
		if response == nil || response.Response == nil || response.Response.InstanceSet == nil {
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

func (s *CloudAccountService) syncVPC(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := vpc.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	return s.paginateSync(func(offset int64) (int, error) {
		request := vpc.NewDescribeVpcsRequest()
		request.Limit = common.StringPtr(fmt.Sprintf("%d", cloudSyncPageSize))
		request.Offset = common.StringPtr(fmt.Sprintf("%d", offset))

		response, err := client.DescribeVpcs(request)
		if err != nil {
			return 0, err
		}
		if response == nil || response.Response == nil || response.Response.VpcSet == nil {
			return 0, nil
		}

		for _, v := range response.Response.VpcSet {
			vpcID := safeDereferenceString(v.VpcId)
			name := safeDereferenceString(v.VpcName)
			cidr := safeDereferenceString(v.CidrBlock)
			state := "available"

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

		return len(response.Response.VpcSet), nil
	})
}

func (s *CloudAccountService) syncSubnets(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := vpc.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	return s.paginateSync(func(offset int64) (int, error) {
		request := vpc.NewDescribeSubnetsRequest()
		request.Limit = common.StringPtr(fmt.Sprintf("%d", cloudSyncPageSize))
		request.Offset = common.StringPtr(fmt.Sprintf("%d", offset))

		response, err := client.DescribeSubnets(request)
		if err != nil {
			return 0, err
		}
		if response == nil || response.Response == nil || response.Response.SubnetSet == nil {
			return 0, nil
		}

		for _, sub := range response.Response.SubnetSet {
			subnetID := safeDereferenceString(sub.SubnetId)
			name := safeDereferenceString(sub.SubnetName)
			cidr := safeDereferenceString(sub.CidrBlock)
			vpcID := safeDereferenceString(sub.VpcId)
			zone := safeDereferenceString(sub.Zone)
			state := "available"

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

		return len(response.Response.SubnetSet), nil
	})
}

func (s *CloudAccountService) syncSecurityGroups(credential *common.Credential, cpf *profile.ClientProfile, account *model.CloudAccount, region string) error {
	client, err := vpc.NewClient(credential, region, cpf)
	if err != nil {
		return err
	}

	return s.paginateSync(func(offset int64) (int, error) {
		request := vpc.NewDescribeSecurityGroupsRequest()
		request.Limit = common.StringPtr(fmt.Sprintf("%d", cloudSyncPageSize))
		request.Offset = common.StringPtr(fmt.Sprintf("%d", offset))

		response, err := client.DescribeSecurityGroups(request)
		if err != nil {
			return 0, err
		}
		if response == nil || response.Response == nil || response.Response.SecurityGroupSet == nil {
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
		if response == nil || response.Response == nil || response.Response.DiskSet == nil {
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
