package middleware

import (
	"net/http"
	"strings"

	"devops-platform/config"
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

		// 3. 注入上下文
		c.Set("userID", sessionData.UserID)
		c.Set("username", sessionData.Username)
		c.Set("authSource", sessionData.AuthSource)
		c.Next()
	}
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
