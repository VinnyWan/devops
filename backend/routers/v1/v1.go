package v1

import (
	"devops-platform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.RequestContext(), middleware.Audit())

	// 鉴权接口
	auth := apiV1.Group("")
	auth.Use(authMiddlewares()...)

	// 旧业务路由（K8s / CMDB 等暂保留）
	registerCluster(auth)
	registerAlert(auth)
	registerLog(auth)
	registerMonitor(auth)
	registerHarbor(auth)
	registerCICD(auth)
	registerApp(auth)
	registerAudit(auth)
	registerLoginLog(auth)
	registerCMDB(auth)

	// 新路由分组（RESTful）
	registerAuthRoutes(apiV1)
	registerSystemRoutes(apiV1)
	registerPlatformRoutes(apiV1)
}
