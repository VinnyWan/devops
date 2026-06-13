package middleware

import (
	"net/http"
	"strings"

	"devops-platform/internal/modules/user/service"
	jwtpkg "devops-platform/internal/pkg/jwt"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var authDB *gorm.DB

func SetAuthDB(db *gorm.DB) {
	authDB = db
}

// UnifiedAuth is a unified authentication middleware that supports:
// 1. JWT Bearer token (for external API calls)
// 2. API Key (for machine-to-machine communication)
// 3. Session Cookie fallback (for browser-based web UI)
//
// It tries each mode in order and injects user context on success.
func UnifiedAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 {
				scheme := strings.ToLower(parts[0])
				token := strings.TrimSpace(parts[1])

				switch scheme {
				case "bearer":
					if tryJWT(c, token) {
						c.Next()
						return
					}
					if trySessionToken(c, token) {
						c.Next()
						return
					}
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"code":    401,
						"message": "未认证：token 无效或已过期",
					})
					return

				case "apikey":
					if tryAPIKey(c, token) {
						c.Next()
						return
					}
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"code":    401,
						"message": "未认证：API Key 无效",
					})
					return
				}
			}
		}

		// Fallback: Session Cookie for browser auth
		if trySessionCookie(c) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证：缺少认证凭据",
		})
	}
}

func tryJWT(c *gin.Context, token string) bool {
	claims, err := jwtpkg.Default().ValidateToken(token)
	if err != nil {
		return false
	}
	if claims.TokenType != "access" {
		return false
	}
	injectUserContext(c, claims.UserID, claims.Username, claims.TenantID, claims.TenantCode, "jwt")
	return true
}

func tryAPIKey(c *gin.Context, key string) bool {
	if authDB == nil {
		return false
	}
	apiKeySvc := service.NewApiKeyService(authDB)
	apiKey, err := apiKeySvc.Validate(key)
	if err != nil {
		logger.Log.Debug("API Key validation failed",
			zap.Error(err))
		return false
	}
	injectUserContext(c, apiKey.UserID, "", apiKey.TenantID, "", "apikey")
	return true
}

// trySessionToken handles session_id passed as Bearer token (for Swagger compatibility).
func trySessionToken(c *gin.Context, token string) bool {
	return trySession(c, token)
}

func trySessionCookie(c *gin.Context) bool {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		return false
	}
	return trySession(c, sessionID)
}

func trySession(c *gin.Context, sessionID string) bool {
	sessionSvc := service.NewSessionService()
	sessionData, err := sessionSvc.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		logger.Log.Debug("Session validation failed in unified auth",
			zap.String("sessionID", sessionID),
			zap.Error(err))
		return false
	}
	injectUserContext(c, sessionData.UserID, sessionData.Username, sessionData.TenantID, sessionData.TenantCode, "session")
	return true
}

func injectUserContext(c *gin.Context, userID uint, username string, tenantID uint, tenantCode string, authSource string) {
	c.Set("userID", userID)
	c.Set("username", username)
	c.Set("tenantID", tenantID)
	c.Set("tenantCode", tenantCode)
	c.Set("authSource", authSource)
}
