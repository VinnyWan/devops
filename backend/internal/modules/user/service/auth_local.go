package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/pkg/utils"

	"gorm.io/gorm"
)

// LocalAuthProvider 本地账号密码认证
type LocalAuthProvider struct {
	userRepo   *repository.UserRepo
	tenantRepo *repository.TenantRepo
}

// NewLocalAuthProvider 创建本地认证提供者
func NewLocalAuthProvider(userRepo *repository.UserRepo, tenantRepo *repository.TenantRepo) *LocalAuthProvider {
	return &LocalAuthProvider{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// Name 返回提供者名称
func (p *LocalAuthProvider) Name() string {
	return string(model.AuthTypeLocal)
}

// Authenticate 执行本地账号密码认证
func (p *LocalAuthProvider) Authenticate(ctx context.Context, cred Credentials) (*AuthResult, error) {
	// 1. 校验用户名密码非空
	username := strings.TrimSpace(cred.Username)
	password := strings.TrimSpace(cred.Password)
	tenantCode := strings.ToLower(strings.TrimSpace(cred.TenantCode))

	if username == "" || password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	if tenantCode == "" {
		return nil, errors.New("租户编码不能为空")
	}

	// 2. 根据 tenantCode 查找租户
	tenant, err := p.tenantRepo.GetByCode(tenantCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("租户不存在")
		}
		return nil, fmt.Errorf("查询租户失败: %w", err)
	}
	if tenant.Status != "active" {
		return nil, errors.New("租户未激活")
	}

	// 3. 在租户内查找用户
	user, err := p.userRepo.GetByUsernameInTenant(tenant.ID, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 4. 检查用户状态
	if user.IsLocked {
		return nil, errors.New("账号已被锁定")
	}
	if user.Status != "active" {
		return nil, errors.New("账号未激活")
	}
	if user.AuthType != model.AuthTypeLocal {
		return nil, fmt.Errorf("用户认证类型为 %s，请使用对应登录方式", user.AuthType)
	}

	// 5. 校验密码
	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	return &AuthResult{
		User:     user,
		Tenant:   tenant,
		AuthType: model.AuthTypeLocal,
	}, nil
}
