package service

import (
	"context"

	"devops-platform/internal/modules/user/model"
)

// Credentials 认证凭据
type Credentials struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	TenantCode string `json:"tenantCode"`
	AuthType   string `json:"authType"`
	// OIDC
	Code  string `json:"code,omitempty"`
	State string `json:"state,omitempty"`
	// SSO
	Provider string `json:"provider,omitempty"`
	AuthCode string `json:"authCode,omitempty"`
}

// AuthResult 认证结果
type AuthResult struct {
	User       *model.User
	Tenant     *model.Tenant
	AuthType   model.AuthType
	ExternalID string
}

// AuthProvider 认证提供者接口
type AuthProvider interface {
	Authenticate(ctx context.Context, cred Credentials) (*AuthResult, error)
	Name() string
}
