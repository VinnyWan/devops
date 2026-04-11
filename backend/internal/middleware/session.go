package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"devops-platform/config"
	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func clearSessionCookie(c *gin.Context) {
	secure := config.Cfg.GetString("server.mode") == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("session_id", "", -1, "/", "", secure, true)
}

// SessionAuth 会话认证中间件
// 优先从 Cookie 获取 session_id
func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 SessionID
		sessionID := extractSessionID(c)
		if sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证：缺少会话凭证",
			})
			return
		}

		// 2. 校验会话
		sessionSvc := service.NewSessionService()
		sessionData, err := sessionSvc.ValidateSession(c.Request.Context(), sessionID)
		if err != nil {
			logger.Log.Warn("Invalid session", zap.String("sessionID", sessionID), zap.Error(err))
			clearSessionCookie(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证：会话已过期或无效",
			})
			return
		}
		if db == nil {
			logger.Log.Error("Session auth failed: database not initialized")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "系统错误：数据库连接未初始化",
			})
			return
		}
		if err := validateSessionTenant(sessionData); err != nil {
			logger.Log.Warn("Invalid tenant session",
				zap.String("sessionID", sessionID),
				zap.Uint("tenantID", sessionData.TenantID),
				zap.Error(err),
			)
			_ = sessionSvc.RevokeSession(c.Request.Context(), sessionID)
			clearSessionCookie(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证：租户已失效或不可用",
			})
			return
		}

		// 3. 注入上下文
		c.Set("userID", sessionData.UserID)
		c.Set("username", sessionData.Username)
		c.Set("tenantID", sessionData.TenantID)
		c.Set("tenantCode", sessionData.TenantCode)
		c.Set("authSource", sessionData.AuthSource)
		c.Next()
	}
}

func validateSessionTenant(sessionData *service.SessionData) error {
	if sessionData == nil || sessionData.TenantID == 0 {
		return errors.New("session tenant is missing")
	}

	var tenant model.Tenant
	if err := db.Select("id", "status", "expires_at").First(&tenant, sessionData.TenantID).Error; err != nil {
		return err
	}
	if tenant.Status != "active" {
		return errors.New("tenant is not active")
	}
	if tenant.ExpiresAt != nil && tenant.ExpiresAt.Before(time.Now()) {
		return errors.New("tenant has expired")
	}
	return nil
}

func extractSessionID(c *gin.Context) string {
	// 1. Cookie (Preferred)
	sessionID, err := c.Cookie("session_id")
	if err == nil && sessionID != "" {
		return sessionID
	}

	// 2. Authorization Header (Bearer) - For API debugging or non-browser clients
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	return ""
}
