package k8s

import (
	k8sctrl "devops/controller/k8s"
	"devops/middleware"

	"github.com/gin-gonic/gin"
)

// SetupClusterRoutes 设置K8s集群管理路由
func SetupClusterRoutes(r *gin.RouterGroup) {
	clusterCtrl := k8sctrl.NewClusterController()

	// K8s集群管理路由
	k8s := r.Group("/k8s")
	k8s.Use(middleware.JWTAuth())
	{
		// 集群基础管理
		k8s.POST("/clusters", clusterCtrl.Create)
		k8s.GET("/clusters", clusterCtrl.GetList)
		k8s.GET("/clusters/:clusterId", clusterCtrl.GetByID)
		k8s.PUT("/clusters/:clusterId", middleware.K8sPermission("update"), clusterCtrl.Update)
		k8s.DELETE("/clusters/:clusterId", middleware.K8sPermission("delete"), clusterCtrl.Delete)
		k8s.GET("/clusters/:clusterId/health", middleware.K8sPermission("get"), clusterCtrl.HealthCheck)

		// 集群权限管理
		k8s.POST("/clusters/:clusterId/access", clusterCtrl.CreateAccess)
		k8s.GET("/clusters/:clusterId/access", clusterCtrl.GetAccessList)
		k8s.DELETE("/clusters/:clusterId/access/:accessId", clusterCtrl.DeleteAccess)
	}
}
