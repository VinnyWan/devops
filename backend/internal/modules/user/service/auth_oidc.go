package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"devops-platform/config"
	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/pkg/redis"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// OIDCAuthProvider OIDC认证提供者
type OIDCAuthProvider struct {
	userRepo   *repository.UserRepo
	tenantRepo *repository.TenantRepo
}

// NewOIDCAuthProvider 创建OIDC认证提供者
func NewOIDCAuthProvider(userRepo *repository.UserRepo, tenantRepo *repository.TenantRepo) *OIDCAuthProvider {
	return &OIDCAuthProvider{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// Name 返回提供者名称
func (p *OIDCAuthProvider) Name() string {
	return string(model.AuthTypeOIDC)
}

// Authenticate 执行OIDC认证
func (p *OIDCAuthProvider) Authenticate(ctx context.Context, cred Credentials) (*AuthResult, error) {
	// 1. 校验输入
	code := strings.TrimSpace(cred.Code)
	state := strings.TrimSpace(cred.State)
	if code == "" || state == "" {
		return nil, errors.New("授权码和状态参数不能为空")
	}

	// 2. 检查OIDC配置是否启用
	if config.Cfg == nil || !config.Cfg.GetBool("oidc.enable") {
		return nil, errors.New("OIDC未启用")
	}

	// 3. 验证state，获取nonce
	expectedNonce, err := redis.Get(ctx, oidcStateKey(state))
	if err != nil || expectedNonce == "" {
		return nil, errors.New("无效或已过期的OIDC状态")
	}
	_ = redis.Del(ctx, oidcStateKey(state))

	// 4. 构建OIDC配置
	conf, verifier, err := p.buildOIDCConfig(ctx)
	if err != nil {
		return nil, err
	}

	// 5. OAuth2 code exchange
	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("令牌交换失败: %w", err)
	}

	// 6. 验证ID Token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("响应中缺少id_token")
	}
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("ID Token验证失败: %w", err)
	}

	// 7. 解析claims
	var claims struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		Nonce             string `json:"nonce"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("解析ID Token声明失败: %w", err)
	}
	if claims.Nonce == "" || claims.Nonce != expectedNonce {
		return nil, errors.New("OIDC nonce不匹配")
	}
	if claims.Sub == "" {
		return nil, errors.New("ID Token中缺少sub声明")
	}

	// 8. 确定用户名
	username := claims.PreferredUsername
	if username == "" {
		username = claims.Email
	}
	if username == "" {
		username = claims.Sub
	}

	// 9. 查找已绑定租户的用户
	user, err := p.userRepo.GetByUsername(username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, errors.New("OIDC用户未绑定任何租户")
	}
	if user.TenantID == nil || *user.TenantID == 0 {
		return nil, errors.New("OIDC用户未绑定任何租户")
	}

	// 10. 查找租户
	tenant, err := p.tenantRepo.GetByID(*user.TenantID)
	if err != nil {
		return nil, fmt.Errorf("查询租户失败: %w", err)
	}
	if tenant.Status != "active" {
		return nil, errors.New("租户未激活")
	}
	if tenant.ExpiresAt != nil && tenant.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("租户已过期")
	}

	// 11. 更新用户信息
	updates := map[string]interface{}{
		"email":       claims.Email,
		"name":        claims.Name,
		"external_id": claims.Sub,
		"auth_type":   model.AuthTypeOIDC,
		"updated_at":  time.Now(),
	}
	if err := p.userRepo.UpdateByIDInTenant(tenant.ID, user.ID, updates); err != nil {
		return nil, fmt.Errorf("更新OIDC用户失败: %w", err)
	}

	// 12. 重新加载最新用户数据
	user, err = p.userRepo.GetByIDInTenant(tenant.ID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("重新加载OIDC用户失败: %w", err)
	}

	return &AuthResult{
		User:       user,
		Tenant:     tenant,
		AuthType:   model.AuthTypeOIDC,
		ExternalID: claims.Sub,
	}, nil
}

// buildOIDCConfig 构建OAuth2配置和ID Token验证器
func (p *OIDCAuthProvider) buildOIDCConfig(ctx context.Context) (*oauth2.Config, *oidc.IDTokenVerifier, error) {
	providerURL := strings.TrimSpace(config.Cfg.GetString("oidc.provider"))
	if providerURL == "" {
		return nil, nil, errors.New("OIDC Provider地址为空")
	}

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		return nil, nil, fmt.Errorf("加载OIDC Provider失败: %w", err)
	}

	conf := &oauth2.Config{
		ClientID:     config.Cfg.GetString("oidc.client_id"),
		ClientSecret: config.Cfg.GetString("oidc.client_secret"),
		RedirectURL:  config.Cfg.GetString("oidc.redirect_url"),
		Scopes:       config.Cfg.GetStringSlice("oidc.scopes"),
		Endpoint:     provider.Endpoint(),
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: conf.ClientID,
	})

	return conf, verifier, nil
}
