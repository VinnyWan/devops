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
	"devops-platform/internal/pkg/utils"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-ldap/ldap/v3"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo   *repository.UserRepo
	roleRepo   *repository.RoleRepo
	sessionSvc *SessionService
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		userRepo:   repository.NewUserRepo(db),
		roleRepo:   repository.NewRoleRepo(db),
		sessionSvc: NewSessionService(),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	AuthType string `json:"authType"` // local, ldap
}

// LoginResponse 登录响应
type LoginResponse struct {
	SessionID string      `json:"sessionId"`
	User      *model.User `json:"user"`
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *LoginRequest, ip, userAgent string) (*LoginResponse, error) {
	authType := model.AuthType(req.AuthType)
	if authType == "" {
		authType = model.AuthTypeLocal
	}

	var user *model.User
	var err error

	switch authType {
	case model.AuthTypeLocal:
		user, err = s.loginLocal(req.Username, req.Password)
	case model.AuthTypeLDAP:
		user, err = s.loginLDAP(req.Username, req.Password)
	case model.AuthTypeOIDC:
		return nil, errors.New("use OIDC callback endpoint for login")
	default:
		return nil, errors.New("invalid auth type")
	}

	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	_ = s.userRepo.UpdateLastLoginTime(user.ID)

	// 创建 Session
	sessionID, err := s.sessionSvc.CreateSession(ctx, user.ID, user.Username, string(authType), ip, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		SessionID: sessionID,
		User:      user,
	}, nil
}

// loginLocal 本地认证登录
func (s *AuthService) loginLocal(username, password string) (*model.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	// 检查账号状态
	if user.IsLocked {
		return nil, errors.New("account is locked")
	}
	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	// 只能本地用户登录
	if user.AuthType != model.AuthTypeLocal {
		return nil, fmt.Errorf("user auth type is %s, please use corresponding login method", user.AuthType)
	}

	// 验证密码
	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

// loginLDAP LDAP认证登录
func (s *AuthService) loginLDAP(username, password string) (*model.User, error) {
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

func (s *AuthService) GetOIDCAuthURL(ctx context.Context) (string, string, error) {
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

func (s *AuthService) LoginOIDC(ctx context.Context, code, state, ip, userAgent string) (*LoginResponse, error) {
	if !config.Cfg.GetBool("oidc.enable") {
		return nil, errors.New("OIDC is not enabled")
	}
	if state == "" {
		return nil, errors.New("missing oidc state")
	}

	expectedNonce, err := redis.Get(ctx, oidcStateKey(state))
	if err != nil || expectedNonce == "" {
		return nil, errors.New("invalid or expired oidc state")
	}
	_ = redis.Del(ctx, oidcStateKey(state))

	conf, verifier, err := s.buildOIDCConfig(ctx)
	if err != nil {
		return nil, err
	}

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token in response")
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verify id_token failed: %w", err)
	}

	var claims struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		Nonce             string `json:"nonce"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("parse id_token claims failed: %w", err)
	}
	if claims.Nonce == "" || claims.Nonce != expectedNonce {
		return nil, errors.New("oidc nonce mismatch")
	}
	if claims.Sub == "" {
		return nil, errors.New("sub claim not found in id_token")
	}

	username := claims.PreferredUsername
	if username == "" {
		username = claims.Email
	}
	if username == "" {
		username = claims.Sub
	}

	// Sync user
	// 查找或创建本地用户
	user, err := s.userRepo.GetByUsername(username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if user == nil {
		// 创建新用户
		user = &model.User{
			Username:   username,
			Email:      claims.Email,
			Name:       claims.Name,
			AuthType:   model.AuthTypeOIDC,
			ExternalID: claims.Sub,
			Status:     "active",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create OIDC user: %w", err)
		}
	} else {
		// 更新用户信息
		updates := map[string]interface{}{
			"email":       claims.Email,
			"name":        claims.Name,
			"external_id": claims.Sub,
			"auth_type":   model.AuthTypeOIDC,
			"updated_at":  time.Now(),
		}
		if err := s.userRepo.UpdateByID(user.ID, updates); err != nil {
			return nil, fmt.Errorf("failed to update OIDC user: %w", err)
		}
		// 重新获取最新数据
		user, _ = s.userRepo.GetByID(user.ID)
	}

	// Create Session
	sessionID, err := s.sessionSvc.CreateSession(ctx, user.ID, user.Username, string(model.AuthTypeOIDC), ip, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		SessionID: sessionID,
		User:      user,
	}, nil
}
