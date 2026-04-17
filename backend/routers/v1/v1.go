package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.RequestContext(), middleware.Audit())

	// 认证相关公开接口
	apiV1.POST("/user/login", api.Login)
	apiV1.POST("/user/register", api.Register)

	// OIDC
	apiV1.GET("/auth/oidc/login", api.OIDCLogin)
	apiV1.GET("/auth/oidc/callback", api.OIDCCallback)

	// 鉴权接口
	auth := apiV1.Group("")
	auth.Use(authMiddlewares()...)

	auth.POST("/user/logout", api.Logout)

	// 按模块注册
	registerCluster(auth)
	registerAlert(auth)
	registerLog(auth)
	registerMonitor(auth)
	registerHarbor(auth)
	registerCICD(auth)
	registerApp(auth)
	registerDepartment(auth)
	registerUser(auth)
	registerRole(auth)
	registerAudit(auth)
	registerLoginLog(auth)
	registerTenant(auth)
	registerCMDB(auth)
}
