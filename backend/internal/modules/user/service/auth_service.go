package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"devops-platform/config"
	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/pkg/redis"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-ldap/ldap/v3"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// AuthService 认证服务，使用策略模式管理多种认证方式
type AuthService struct {
	providers map[string]AuthProvider
	session   *SessionService
	userRepo  *repository.UserRepo
	roleRepo  *repository.RoleRepo
	tenantRepo *repository.TenantRepo
}

// NewAuthService 创建认证服务并注册所有认证提供者
func NewAuthService(db *gorm.DB) *AuthService {
	userRepo := repository.NewUserRepo(db)
	roleRepo := repository.NewRoleRepo(db)
	tenantRepo := repository.NewTenantRepo(db)

	s := &AuthService{
		providers: make(map[string]AuthProvider),
		session:   NewSessionService(),
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		tenantRepo: tenantRepo,
	}

	// 注册内置认证提供者
	s.RegisterProvider(NewLocalAuthProvider(userRepo, tenantRepo))
	s.RegisterProvider(NewLDAPAuthProvider(userRepo, tenantRepo))
	s.RegisterProvider(NewOIDCAuthProvider(userRepo, tenantRepo))

	// 注册第三方 SSO 认证提供者（框架已就绪，具体 API 对接待实现）
	s.RegisterProvider(NewSSOAuthProvider("feishu", userRepo, tenantRepo))
	s.RegisterProvider(NewSSOAuthProvider("dingtalk", userRepo, tenantRepo))
	s.RegisterProvider(NewSSOAuthProvider("wecom", userRepo, tenantRepo))

	return s
}

// RegisterProvider 注册认证提供者
func (s *AuthService) RegisterProvider(p AuthProvider) {
	s.providers[p.Name()] = p
}

// LoginRequest 登录请求
type LoginRequest struct {
	TenantCode string `json:"tenantCode" binding:"required"`
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	AuthType   string `json:"authType"` // local, ldap, oidc, sso_feishu, sso_dingtalk, sso_wecom
}

// LoginResponse 登录响应
type LoginResponse struct {
	SessionID string      `json:"sessionId"`
	User      *model.User `json:"user"`
}

// Login 统一登录入口
func (s *AuthService) Login(ctx context.Context, req *LoginRequest, ip, userAgent string) (*LoginResponse, error) {
	tenantCode := strings.ToLower(strings.TrimSpace(req.TenantCode))
	if tenantCode == "" {
		return nil, errors.New("tenantCode is required")
	}
	tenant, err := s.tenantRepo.GetByCode(tenantCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}
	if tenant.Status != "active" {
		return nil, errors.New("tenant is not active")
	}

	authType := req.AuthType
	if authType == "" {
		authType = string(model.AuthTypeLocal)
	}

	// 构造凭据
	cred := Credentials{
		Username:   req.Username,
		Password:   req.Password,
		TenantCode: tenantCode,
		AuthType:   authType,
	}

	// 查找对应的认证提供者
	provider := s.providers[authType]
	if provider == nil {
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}

	// 执行认证
	result, err := provider.Authenticate(ctx, cred)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	_ = s.userRepo.UpdateLastLoginTimeInTenant(tenant.ID, result.User.ID)

	// 创建会话
	sessionID, err := s.session.CreateSession(ctx, result.User.ID, result.User.Username, tenant.ID, tenant.Code, string(result.AuthType), ip, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		SessionID: sessionID,
		User:      result.User,
	}, nil
}

// GetOIDCAuthURL 获取 OIDC 登录跳转地址（保留对 auth handler 的兼容）
func (s *AuthService) GetOIDCAuthURL(ctx context.Context) (string, string, error) {
	if config.Cfg == nil || !config.Cfg.GetBool("auth.enable_external") {
		return "", "", errors.New("oidc login is temporarily disabled in tenant mode")
	}

	if !config.Cfg.GetBool("oidc.enable") {
		return "", "", errors.New("OIDC is not enabled")
	}

	conf, _, err := s.buildOIDCConfig(ctx)
	if err != nil {
		return "", "", err
	}

	state, err := generateSecureRandomString(24)
	if err != nil {
		return "", "", fmt.Errorf("generate state failed: %w", err)
	}
	nonce, err := generateSecureRandomString(24)
	if err != nil {
		return "", "", fmt.Errorf("generate nonce failed: %w", err)
	}

	if err := redis.Set(ctx, oidcStateKey(state), nonce, 5*time.Minute); err != nil {
		return "", "", fmt.Errorf("save oidc state failed: %w", err)
	}

	url := conf.AuthCodeURL(state, oidc.Nonce(nonce))
	return url, state, nil
}

// LoginOIDC OIDC 回调登录（保留对 auth handler 的兼容）
func (s *AuthService) LoginOIDC(ctx context.Context, code, state, ip, userAgent string) (*LoginResponse, error) {
	// 直接委托给 OIDC AuthProvider
	cred := Credentials{
		Code:    code,
		State:   state,
		AuthType: string(model.AuthTypeOIDC),
	}

	provider := s.providers[string(model.AuthTypeOIDC)]
	if provider == nil {
		return nil, errors.New("OIDC provider not registered")
	}

	result, err := provider.Authenticate(ctx, cred)
	if err != nil {
		return nil, err
	}

	// 创建会话
	sessionID, err := s.session.CreateSession(ctx, result.User.ID, result.User.Username, result.Tenant.ID, result.Tenant.Code, string(result.AuthType), ip, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		SessionID: sessionID,
		User:      result.User,
	}, nil
}

// ---------- 以下为辅助方法，保留兼容性 ----------

// buildOIDCConfig 构建 OAuth2 配置和 ID Token 验证器
func (s *AuthService) buildOIDCConfig(ctx context.Context) (*oauth2.Config, *oidc.IDTokenVerifier, error) {
	providerURL := strings.TrimSpace(config.Cfg.GetString("oidc.provider"))
	if providerURL == "" {
		return nil, nil, errors.New("OIDC provider is empty")
	}

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		return nil, nil, fmt.Errorf("load oidc provider failed: %w", err)
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

func generateSecureRandomString(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func oidcStateKey(state string) string {
	return "oidc:state:" + state
}

// loginLDAP 保留旧版 LDAP 登录方法（兼容性），内部委托给 LDAPAuthProvider
func (s *AuthService) loginLDAP(username, password string) (*model.User, error) {
	if config.Cfg == nil || !config.Cfg.GetBool("auth.enable_external") {
		return nil, errors.New("ldap login is temporarily disabled in tenant mode")
	}

	if !config.Cfg.GetBool("ldap.enable") {
		return nil, errors.New("LDAP is not enabled")
	}

	ldapHost := config.Cfg.GetString("ldap.host")
	ldapPort := config.Cfg.GetInt("ldap.port")

	// 连接 LDAP
	l, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", ldapHost, ldapPort))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	defer l.Close()

	// 绑定管理员账号（用于搜索）
	bindDN := config.Cfg.GetString("ldap.bind_dn")
	bindPassword := config.Cfg.GetString("ldap.bind_password")
	if err := l.Bind(bindDN, bindPassword); err != nil {
		return nil, fmt.Errorf("LDAP admin bind failed: %w", err)
	}

	// 搜索用户
	baseDN := config.Cfg.GetString("ldap.base_dn")
	userFilter := config.Cfg.GetString("ldap.user_filter")
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(userFilter, username),
		[]string{"dn", config.Cfg.GetString("ldap.attributes.email"), config.Cfg.GetString("ldap.attributes.nickname"), config.Cfg.GetString("ldap.attributes.username")},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %w", err)
	}

	if len(sr.Entries) != 1 {
		return nil, errors.New("user not found or too many entries in LDAP")
	}

	userEntry := sr.Entries[0]
	userDN := userEntry.DN

	// 使用用户凭据绑定（验证密码）
	if err := l.Bind(userDN, password); err != nil {
		return nil, errors.New("invalid LDAP password")
	}

	// 认证成功，同步用户
	attrEmail := config.Cfg.GetString("ldap.attributes.email")
	attrNickname := config.Cfg.GetString("ldap.attributes.nickname")

	email := userEntry.GetAttributeValue(attrEmail)
	nickname := userEntry.GetAttributeValue(attrNickname)
	if nickname == "" {
		nickname = username
	}

	// 查找或创建本地用户
	user, err := s.userRepo.GetByUsername(username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if user == nil {
		// 创建新用户
		user = &model.User{
			Username:   username,
			Email:      email,
			Name:       nickname,
			AuthType:   model.AuthTypeLDAP,
			ExternalID: userDN,
			Status:     "active",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create LDAP user: %w", err)
		}
	} else {
		// 更新用户信息
		updates := map[string]interface{}{
			"email":       email,
			"name":        nickname,
			"external_id": userDN,
			"auth_type":   model.AuthTypeLDAP,
			"updated_at":  time.Now(),
		}
		if err := s.userRepo.UpdateByID(user.ID, updates); err != nil {
			return nil, fmt.Errorf("failed to update LDAP user: %w", err)
		}
		// 重新获取最新数据
		user, _ = s.userRepo.GetByID(user.ID)
	}

	return user, nil
}
