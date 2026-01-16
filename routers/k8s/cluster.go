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
		k8s.POST("/cluster/create", clusterCtrl.Create)
		k8s.GET("/cluster/list", clusterCtrl.GetList)
		k8s.GET("/clusters", clusterCtrl.GetList)
		k8s.GET("/cluster/detail", clusterCtrl.GetByID)
		k8s.POST("/cluster/update", middleware.K8sPermission("update"), clusterCtrl.Update)  // 改为POST
		k8s.POST("/cluster/delete", middleware.K8sPermission("delete"), clusterCtrl.Delete)  // 改为POST
		k8s.GET("/cluster/health", middleware.K8sPermission("get"), clusterCtrl.HealthCheck)
		k8s.POST("/cluster/reimport", middleware.K8sPermission("update"), clusterCtrl.ReimportKubeConfig)

		// 集群权限管理
		k8s.POST("/cluster/access", clusterCtrl.CreateAccess)
		k8s.GET("/cluster/access", clusterCtrl.GetAccessList)
		k8s.POST("/cluster/access/delete", clusterCtrl.DeleteAccess)  // 改为POST
	}
}
