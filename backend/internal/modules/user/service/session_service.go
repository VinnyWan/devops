package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"devops-platform/config"
	"devops-platform/internal/pkg/redis"

	"github.com/google/uuid"
)

type SessionService struct{}

type SessionData struct {
	UserID     uint      `json:"userId"`
	Username   string    `json:"username"`
	TenantID   uint      `json:"tenantId"`
	TenantCode string    `json:"tenantCode"`
	AuthSource string    `json:"authSource"` // LOCAL/LDAP/OIDC
	LoginTime  time.Time `json:"loginTime"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"userAgent"`
	// Permissions are stored in Session to reduce DB/Cache hits during requests
	// However, complex permission updates might require clearing this or re-login
	// For this spec, we will store them here but also check separate cache if needed
	// Actually, per spec "Session 中仅存储用户 ID、用户名等基础信息，不存储权限列表", we remove permissions from here.
}

func NewSessionService() *SessionService {
	return &SessionService{}
}

func sessionTTL() time.Duration {
	expireSeconds := 7200
	if config.Cfg != nil {
		v := config.Cfg.GetInt("session.expire")
		if v > 0 {
			expireSeconds = v
		}
	}
	return time.Duration(expireSeconds) * time.Second
}

func userSessionSetKey(tenantID, userID uint) string {
	return fmt.Sprintf("tenant:%d:user_sessions:%d", tenantID, userID)
}

// CreateSession 创建新会话
func (s *SessionService) CreateSession(
	ctx context.Context,
	userID uint,
	username string,
	tenantID uint,
	tenantCode string,
	authSource string,
	ip string,
	userAgent string,
) (string, error) {
	// 1. 生成 SessionID
	sessionID := uuid.New().String()

	// 2. 清理旧会话（单点登录 - 可选，这里实现为单点登录）
	userSessionsKey := userSessionSetKey(tenantID, userID)

	// 获取所有旧的 sessionID
	oldSessionIDs, err := redis.SMembers(ctx, userSessionsKey)
	if err == nil {
		for _, oldID := range oldSessionIDs {
			// 删除旧的 session 数据
			redis.Del(ctx, fmt.Sprintf("session:%s", oldID))
		}
		// 清空用户会话集合
		redis.Del(ctx, userSessionsKey)
	}

	// 3. 保存新会话
	sessionData := SessionData{
		UserID:     userID,
		Username:   username,
		TenantID:   tenantID,
		TenantCode: tenantCode,
		AuthSource: authSource,
		LoginTime:  time.Now(),
		IP:         ip,
		UserAgent:  userAgent,
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return "", err
	}

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	expiration := sessionTTL()

	// 存储 Session 数据
	if err := redis.Set(ctx, sessionKey, string(jsonData), expiration); err != nil {
		return "", err
	}

	// 记录 SessionID 到用户集合
	if err := redis.SAdd(ctx, userSessionsKey, sessionID); err != nil {
		return "", err
	}
	// 设置用户集合过期时间（略长于 session 过期时间）
	redis.Expire(ctx, userSessionsKey, expiration+10*time.Minute)

	return sessionID, nil
}

// ValidateSession 校验会话
func (s *SessionService) ValidateSession(ctx context.Context, sessionID string) (*SessionData, error) {
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	data, err := redis.Get(ctx, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("session invalid or expired")
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return nil, err
	}

	// 滑动过期：延长有效期
	expiration := sessionTTL()
	redis.Expire(ctx, sessionKey, expiration)

	// 同时延长用户会话集合有效期
	userSessionsKey := userSessionSetKey(sessionData.TenantID, sessionData.UserID)
	redis.Expire(ctx, userSessionsKey, expiration+10*time.Minute)

	return &sessionData, nil
}

// RevokeSession 注销会话
func (s *SessionService) RevokeSession(ctx context.Context, sessionID string) error {
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	// 获取 session 数据以拿到 userID
	data, err := redis.Get(ctx, sessionKey)
	if err != nil {
		return nil // 已经不存在了
	}

	var sessionData SessionData
	json.Unmarshal([]byte(data), &sessionData)

	// 删除 session 数据
	if err := redis.Del(ctx, sessionKey); err != nil {
		return err
	}

	// 从用户会话集合中移除
	if sessionData.UserID > 0 {
		userSessionsKey := userSessionSetKey(sessionData.TenantID, sessionData.UserID)
		redis.SRem(ctx, userSessionsKey, sessionID)
	}

	return nil
}

// RevokeAllUserSessions 强制下线用户所有端
func (s *SessionService) RevokeAllUserSessions(ctx context.Context, tenantID uint, userID uint) error {
	userSessionsKey := userSessionSetKey(tenantID, userID)

	sessionIDs, err := redis.SMembers(ctx, userSessionsKey)
	if err != nil {
		return err
	}

	for _, sessionID := range sessionIDs {
		redis.Del(ctx, fmt.Sprintf("session:%s", sessionID))
	}

	return redis.Del(ctx, userSessionsKey)
}
