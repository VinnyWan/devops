package k8s

import (
	k8sctrl "devops/controller/k8s"
	"devops/middleware"

	"github.com/gin-gonic/gin"
)

// SetupResourceRoutes 设置K8s资源管理路由
func SetupResourceRoutes(r *gin.RouterGroup) {
	resourceCtrl := k8sctrl.NewResourceController()

	k8s := r.Group("/k8s")
	k8s.Use(middleware.JWTAuth())
	{
		// Namespace管理
		k8s.GET("/clusters/:clusterId/namespaces", middleware.K8sPermission("get"), resourceCtrl.ListNamespaces)
		k8s.GET("/clusters/:clusterId/namespaces/:name", middleware.K8sPermission("get"), resourceCtrl.GetNamespace)
		k8s.POST("/clusters/:clusterId/namespaces", middleware.K8sPermission("create"), resourceCtrl.CreateNamespace)
		k8s.DELETE("/clusters/:clusterId/namespaces/:name", middleware.K8sPermission("delete"), resourceCtrl.DeleteNamespace)

		// Deployment管理
		k8s.GET("/clusters/:clusterId/deployments", middleware.K8sPermission("get"), resourceCtrl.ListDeployments)
		k8s.GET("/clusters/:clusterId/deployments/:name", middleware.K8sPermission("get"), resourceCtrl.GetDeployment)
		k8s.POST("/clusters/:clusterId/deployments", middleware.K8sPermission("create"), resourceCtrl.CreateDeployment)
		k8s.PUT("/clusters/:clusterId/deployments/:name", middleware.K8sPermission("update"), resourceCtrl.UpdateDeployment)
		k8s.DELETE("/clusters/:clusterId/deployments/:name", middleware.K8sPermission("delete"), resourceCtrl.DeleteDeployment)
		k8s.POST("/clusters/:clusterId/deployments/:name/scale", middleware.K8sPermission("update"), resourceCtrl.ScaleDeployment)
		k8s.POST("/clusters/:clusterId/deployments/:name/restart", middleware.K8sPermission("update"), resourceCtrl.RestartDeployment)

		// Pod管理
		k8s.GET("/clusters/:clusterId/pods", middleware.K8sPermission("get"), resourceCtrl.ListPods)
		k8s.GET("/clusters/:clusterId/pods/:name", middleware.K8sPermission("get"), resourceCtrl.GetPod)
		k8s.DELETE("/clusters/:clusterId/pods/:name", middleware.K8sPermission("delete"), resourceCtrl.DeletePod)
		k8s.GET("/clusters/:clusterId/pods/:name/logs", middleware.K8sPermission("get"), resourceCtrl.GetPodLogs)

		// StatefulSet管理
		k8s.GET("/clusters/:clusterId/statefulsets", middleware.K8sPermission("get"), resourceCtrl.ListStatefulSets)

		// DaemonSet管理
		k8s.GET("/clusters/:clusterId/daemonsets", middleware.K8sPermission("get"), resourceCtrl.ListDaemonSets)

		// Service管理
		k8s.GET("/clusters/:clusterId/services", middleware.K8sPermission("get"), resourceCtrl.ListServices)
		k8s.GET("/clusters/:clusterId/services/:name", middleware.K8sPermission("get"), resourceCtrl.GetService)
		k8s.POST("/clusters/:clusterId/services", middleware.K8sPermission("create"), resourceCtrl.CreateService)
		k8s.DELETE("/clusters/:clusterId/services/:name", middleware.K8sPermission("delete"), resourceCtrl.DeleteService)

		// Ingress管理
		k8s.GET("/clusters/:clusterId/ingresses", middleware.K8sPermission("get"), resourceCtrl.ListIngresses)
		k8s.GET("/clusters/:clusterId/ingresses/:name", middleware.K8sPermission("get"), resourceCtrl.GetIngress)
		k8s.POST("/clusters/:clusterId/ingresses", middleware.K8sPermission("create"), resourceCtrl.CreateIngress)
		k8s.DELETE("/clusters/:clusterId/ingresses/:name", middleware.K8sPermission("delete"), resourceCtrl.DeleteIngress)

		// ConfigMap管理
		k8s.GET("/clusters/:clusterId/configmaps", middleware.K8sPermission("get"), resourceCtrl.ListConfigMaps)
		k8s.GET("/clusters/:clusterId/configmaps/:name", middleware.K8sPermission("get"), resourceCtrl.GetConfigMap)
		k8s.POST("/clusters/:clusterId/configmaps", middleware.K8sPermission("create"), resourceCtrl.CreateConfigMap)
		k8s.PUT("/clusters/:clusterId/configmaps/:name", middleware.K8sPermission("update"), resourceCtrl.UpdateConfigMap)
		k8s.DELETE("/clusters/:clusterId/configmaps/:name", middleware.K8sPermission("delete"), resourceCtrl.DeleteConfigMap)

		// Secret管理
		k8s.GET("/clusters/:clusterId/secrets", middleware.K8sPermission("get"), resourceCtrl.ListSecrets)
		k8s.GET("/clusters/:clusterId/secrets/:name", middleware.K8sPermission("get"), resourceCtrl.GetSecret)
		k8s.POST("/clusters/:clusterId/secrets", middleware.K8sPermission("create"), resourceCtrl.CreateSecret)
		k8s.PUT("/clusters/:clusterId/secrets/:name", middleware.K8sPermission("update"), resourceCtrl.UpdateSecret)
		k8s.DELETE("/clusters/:clusterId/secrets/:name", middleware.K8sPermission("delete"), resourceCtrl.DeleteSecret)

		// PV管理
		k8s.GET("/clusters/:clusterId/pvs", middleware.K8sPermission("get"), resourceCtrl.ListPVs)
		k8s.GET("/clusters/:clusterId/pvs/:name", middleware.K8sPermission("get"), resourceCtrl.GetPV)
		k8s.DELETE("/clusters/:clusterId/pvs/:name", middleware.K8sPermission("delete"), resourceCtrl.DeletePV)

		// PVC管理
		k8s.GET("/clusters/:clusterId/pvcs", middleware.K8sPermission("get"), resourceCtrl.ListPVCs)
		k8s.GET("/clusters/:clusterId/pvcs/:name", middleware.K8sPermission("get"), resourceCtrl.GetPVC)
		k8s.DELETE("/clusters/:clusterId/pvcs/:name", middleware.K8sPermission("delete"), resourceCtrl.DeletePVC)

		// StorageClass管理
		k8s.GET("/clusters/:clusterId/storageclasses", middleware.K8sPermission("get"), resourceCtrl.ListStorageClasses)
		k8s.GET("/clusters/:clusterId/storageclasses/:name", middleware.K8sPermission("get"), resourceCtrl.GetStorageClass)

		// 节点管理
		k8s.GET("/clusters/:clusterId/nodes", middleware.K8sPermission("get"), resourceCtrl.ListNodes)
		k8s.GET("/clusters/:clusterId/nodes/:name", middleware.K8sPermission("get"), resourceCtrl.GetNode)

		// 事件查看
		k8s.GET("/clusters/:clusterId/events", middleware.K8sPermission("get"), resourceCtrl.ListEvents)
		k8s.GET("/clusters/:clusterId/events/object", middleware.K8sPermission("get"), resourceCtrl.GetEventsByObject)
	}
}
