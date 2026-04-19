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

	"github.com/go-ldap/ldap/v3"
	"gorm.io/gorm"
)

// LDAPAuthProvider LDAP认证提供者
type LDAPAuthProvider struct {
	userRepo   *repository.UserRepo
	tenantRepo *repository.TenantRepo
}

// NewLDAPAuthProvider 创建LDAP认证提供者
func NewLDAPAuthProvider(userRepo *repository.UserRepo, tenantRepo *repository.TenantRepo) *LDAPAuthProvider {
	return &LDAPAuthProvider{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// Name 返回提供者名称
func (p *LDAPAuthProvider) Name() string {
	return string(model.AuthTypeLDAP)
}

// Authenticate 执行LDAP认证
func (p *LDAPAuthProvider) Authenticate(ctx context.Context, cred Credentials) (*AuthResult, error) {
	// 1. 校验输入
	username := strings.TrimSpace(cred.Username)
	password := strings.TrimSpace(cred.Password)
	tenantCode := strings.ToLower(strings.TrimSpace(cred.TenantCode))

	if username == "" || password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	if tenantCode == "" {
		return nil, errors.New("租户编码不能为空")
	}

	// 2. 检查LDAP配置是否启用
	if config.Cfg == nil || !config.Cfg.GetBool("ldap.enable") {
		return nil, errors.New("LDAP未启用")
	}

	// 3. 查找租户
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

	// 4. 连接LDAP服务器
	ldapHost := config.Cfg.GetString("ldap.host")
	ldapPort := config.Cfg.GetInt("ldap.port")
	l, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", ldapHost, ldapPort))
	if err != nil {
		return nil, fmt.Errorf("连接LDAP服务器失败: %w", err)
	}
	defer l.Close()

	// 5. 管理员绑定搜索用户
	bindDN := config.Cfg.GetString("ldap.bind_dn")
	bindPassword := config.Cfg.GetString("ldap.bind_password")
	if err := l.Bind(bindDN, bindPassword); err != nil {
		return nil, fmt.Errorf("LDAP管理员绑定失败: %w", err)
	}

	// 6. 搜索用户
	baseDN := config.Cfg.GetString("ldap.base_dn")
	userFilter := config.Cfg.GetString("ldap.user_filter")
	attrEmail := config.Cfg.GetString("ldap.attributes.email")
	attrNickname := config.Cfg.GetString("ldap.attributes.nickname")
	attrUsername := config.Cfg.GetString("ldap.attributes.username")

	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(userFilter, username),
		[]string{"dn", attrEmail, attrNickname, attrUsername},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP搜索失败: %w", err)
	}
	if len(sr.Entries) != 1 {
		return nil, errors.New("LDAP中未找到用户或匹配到多个条目")
	}

	userEntry := sr.Entries[0]
	userDN := userEntry.DN

	// 7. 用户凭据绑定验证
	if err := l.Bind(userDN, password); err != nil {
		return nil, errors.New("LDAP密码错误")
	}

	// 8. 提取用户属性
	email := userEntry.GetAttributeValue(attrEmail)
	nickname := userEntry.GetAttributeValue(attrNickname)
	if nickname == "" {
		nickname = username
	}

	// 9. 在租户内查找或自动创建本地用户
	user, err := p.userRepo.GetByUsernameInTenant(tenant.ID, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	if user == nil {
		// 自动创建用户并绑定租户
		user = &model.User{
			TenantID:   &tenant.ID,
			Username:   username,
			Email:      email,
			Name:       nickname,
			AuthType:   model.AuthTypeLDAP,
			ExternalID: userDN,
			Status:     "active",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := p.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("创建LDAP用户失败: %w", err)
		}
	} else {
		// 更新已有用户信息
		updates := map[string]interface{}{
			"email":       email,
			"name":        nickname,
			"external_id": userDN,
			"auth_type":   model.AuthTypeLDAP,
			"updated_at":  time.Now(),
		}
		if err := p.userRepo.UpdateByIDInTenant(tenant.ID, user.ID, updates); err != nil {
			return nil, fmt.Errorf("更新LDAP用户失败: %w", err)
		}
		user, _ = p.userRepo.GetByIDInTenant(tenant.ID, user.ID)
	}

	// 10. 检查用户状态
	if user.IsLocked {
		return nil, errors.New("账号已被锁定")
	}
	if user.Status != "active" {
		return nil, errors.New("账号未激活")
	}

	return &AuthResult{
		User:       user,
		Tenant:     tenant,
		AuthType:   model.AuthTypeLDAP,
		ExternalID: userDN,
	}, nil
}
