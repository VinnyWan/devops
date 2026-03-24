package api

import (
	"net/http"
	"sync"

	"devops-platform/config"
	"devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	authService *service.AuthService
	authOnce    sync.Once
)

func getAuthService() *service.AuthService {
	authOnce.Do(func() {
		authService = service.NewAuthService(db)
	})
	return authService
}

func setSessionCookie(c *gin.Context, value string, maxAge int) {
	secure := config.Cfg.GetString("server.mode") == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("session_id", value, maxAge, "/", "", secure, true)
}

func setOIDCStateCookie(c *gin.Context, value string, maxAge int) {
	secure := config.Cfg.GetString("server.mode") == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oidc_state", value, maxAge, "/", "", secure, true)
}

// Login godoc
// @Summary 用户登录
// @Description 用户登录（支持本地、LDAP），登录成功返回 session_id (Cookie & Body)
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{} "登录成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "认证失败"
// @Router /user/login [post]
func Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	resp, err := getAuthService().Login(c.Request.Context(), &req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		logger.Log.Warn("Login failed", zap.String("username", req.Username), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": err.Error(),
		})
		return
	}

	setSessionCookie(c, resp.SessionID, 7200)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"token": resp.SessionID,
			"user":  resp.User,
		},
	})
}

// Logout godoc
// @Summary 用户登出
// @Description 清除 Session
// @Tags 认证管理
// @Produce json
// @Success 200 {object} map[string]interface{} "登出成功"
// @Router /user/logout [post]
func Logout(c *gin.Context) {
	// 获取 SessionID
	sessionID, err := c.Cookie("session_id")
	if err == nil && sessionID != "" {
		sessionSvc := service.NewSessionService()
		_ = sessionSvc.RevokeSession(c.Request.Context(), sessionID)
	}

	setSessionCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}

// OIDCLogin godoc
// @Summary 获取OIDC登录地址
// @Description 获取跳转到 OIDC 提供商的 URL
// @Tags 认证管理
// @Produce json
// @Success 200 {object} map[string]interface{} "获取成功"
// @Router /auth/oidc/login [get]
func OIDCLogin(c *gin.Context) {
	url, state, err := getAuthService().GetOIDCAuthURL(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}
	setOIDCStateCookie(c, state, 300)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"url": url,
		},
	})
}

// OIDCCallback godoc
// @Summary OIDC 回调
// @Description 处理 OIDC 登录回调
// @Tags 认证管理
// @Param code query string true "Authorization Code"
// @Success 200 {object} map[string]interface{} "登录成功"
// @Router /auth/oidc/callback [get]
func OIDCCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "missing code",
		})
		return
	}
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "missing state",
		})
		return
	}

	expectedState, _ := c.Cookie("oidc_state")
	if expectedState == "" || expectedState != state {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "invalid oidc state",
		})
		return
	}
	setOIDCStateCookie(c, "", -1)

	resp, err := getAuthService().LoginOIDC(c.Request.Context(), code, state, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		logger.Log.Error("OIDC Login failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": err.Error(),
		})
		return
	}

	setSessionCookie(c, resp.SessionID, 7200)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"token": resp.SessionID,
			"user":  resp.User,
		},
	})
}
