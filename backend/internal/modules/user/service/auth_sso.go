package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"devops-platform/internal/modules/user/repository"

	"gorm.io/gorm"
)

// SSOConfig 单个 SSO 提供商的配置
type SSOConfig struct {
	Provider     string `json:"provider"`      // feishu / dingtalk / wecom
	ClientID     string `json:"clientId"`      // 应用 Client ID
	ClientSecret string `json:"clientSecret"`  // 应用 Client Secret
	AuthURL      string `json:"authUrl"`       // 授权地址
	TokenURL     string `json:"tokenUrl"`      // Token 交换地址
	UserInfoURL  string `json:"userInfoUrl"`   // 用户信息地址
	RedirectURL  string `json:"redirectUrl"`   // 回调地址
	Enabled      bool   `json:"enabled"`       // 是否启用
}

// SSOAuthProvider 第三方 SSO 认证提供者
type SSOAuthProvider struct {
	provider   string // feishu / dingtalk / wecom
	userRepo   *repository.UserRepo
	tenantRepo *repository.TenantRepo
}

// NewSSOAuthProvider 创建 SSO 认证提供者
func NewSSOAuthProvider(provider string, userRepo *repository.UserRepo, tenantRepo *repository.TenantRepo) *SSOAuthProvider {
	return &SSOAuthProvider{
		provider:   provider,
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// Name 返回提供者名称，格式为 "sso_feishu" / "sso_dingtalk" / "sso_wecom"
func (p *SSOAuthProvider) Name() string {
	return "sso_" + p.provider
}

// Authenticate 执行 SSO 认证
func (p *SSOAuthProvider) Authenticate(ctx context.Context, cred Credentials) (*AuthResult, error) {
	// 1. 校验输入
	tenantCode := strings.ToLower(strings.TrimSpace(cred.TenantCode))
	authCode := strings.TrimSpace(cred.AuthCode)
	if tenantCode == "" {
		return nil, errors.New("租户编码不能为空")
	}
	if authCode == "" {
		return nil, errors.New("授权码不能为空")
	}

	// 2. 查找租户
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
	if tenant.ExpiresAt != nil && tenant.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("租户已过期")
	}

	// 3. 从租户 SSO 配置中提取对应 provider 的配置
	ssoCfg, err := getSSOConfig(tenant.SSOConfig, p.provider)
	if err != nil {
		return nil, fmt.Errorf("获取SSO配置失败: %w", err)
	}
	if !ssoCfg.Enabled {
		return nil, fmt.Errorf("SSO %s 未启用", p.provider)
	}

	// 4. SSO 各 provider 的具体 API 对接（待实现）
	// 步骤：用 authCode 换取 accessToken -> 调用 UserInfoURL 获取用户信息 -> 查找或创建本地用户
	return nil, fmt.Errorf("SSO %s 认证待实现", p.provider)
}

// getSSOConfig 从租户 SSO JSON 配置中提取指定 provider 的配置
// JSON 格式示例: {"feishu": {...}, "dingtalk": {...}, "wecom": {...}}
func getSSOConfig(ssoConfigJSON string, provider string) (*SSOConfig, error) {
	if ssoConfigJSON == "" {
		return nil, errors.New("租户未配置SSO")
	}

	var configs map[string]SSOConfig
	if err := json.Unmarshal([]byte(ssoConfigJSON), &configs); err != nil {
		return nil, fmt.Errorf("解析SSO配置失败: %w", err)
	}

	cfg, ok := configs[provider]
	if !ok {
		return nil, fmt.Errorf("未找到 %s 的SSO配置", provider)
	}

	return &cfg, nil
}
