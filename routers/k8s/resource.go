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
		k8s.GET("/namespaces", middleware.K8sPermission("get"), resourceCtrl.ListNamespaces)
		k8s.GET("/namespace/detail", middleware.K8sPermission("get"), resourceCtrl.GetNamespace)
		k8s.POST("/namespace/create", middleware.K8sPermission("create"), resourceCtrl.CreateNamespace)
		k8s.POST("/namespace/delete", middleware.K8sPermission("delete"), resourceCtrl.DeleteNamespace)  // 改为POST

		// Deployment管理
		k8s.GET("/deployments", middleware.K8sPermission("get"), resourceCtrl.ListDeployments)
		k8s.GET("/deployment/detail", middleware.K8sPermission("get"), resourceCtrl.GetDeployment)
		k8s.POST("/deployment/create", middleware.K8sPermission("create"), resourceCtrl.CreateDeployment)
		k8s.POST("/deployment/update", middleware.K8sPermission("update"), resourceCtrl.UpdateDeployment)  // 改为POST
		k8s.POST("/deployment/delete", middleware.K8sPermission("delete"), resourceCtrl.DeleteDeployment)  // 改为POST
		k8s.POST("/deployment/scale", middleware.K8sPermission("update"), resourceCtrl.ScaleDeployment)
		k8s.POST("/deployment/restart", middleware.K8sPermission("update"), resourceCtrl.RestartDeployment)

		// Pod管理
		k8s.GET("/pods", middleware.K8sPermission("get"), resourceCtrl.ListPods)
		k8s.GET("/pod/detail", middleware.K8sPermission("get"), resourceCtrl.GetPod)
		k8s.POST("/pod/delete", middleware.K8sPermission("delete"), resourceCtrl.DeletePod)  // 改为POST
		k8s.GET("/pod/logs", middleware.K8sPermission("get"), resourceCtrl.GetPodLogs)

		// StatefulSet管理
		k8s.GET("/statefulsets", middleware.K8sPermission("get"), resourceCtrl.ListStatefulSets)

		// DaemonSet管理
		k8s.GET("/daemonsets", middleware.K8sPermission("get"), resourceCtrl.ListDaemonSets)

		// Service管理
		k8s.GET("/services", middleware.K8sPermission("get"), resourceCtrl.ListServices)
		k8s.GET("/service/detail", middleware.K8sPermission("get"), resourceCtrl.GetService)
		k8s.POST("/service/create", middleware.K8sPermission("create"), resourceCtrl.CreateService)
		k8s.POST("/service/delete", middleware.K8sPermission("delete"), resourceCtrl.DeleteService)  // 改为POST

		// Ingress管理
		k8s.GET("/ingresses", middleware.K8sPermission("get"), resourceCtrl.ListIngresses)
		k8s.GET("/ingress/detail", middleware.K8sPermission("get"), resourceCtrl.GetIngress)
		k8s.POST("/ingress/create", middleware.K8sPermission("create"), resourceCtrl.CreateIngress)
		k8s.POST("/ingress/delete", middleware.K8sPermission("delete"), resourceCtrl.DeleteIngress)  // 改为POST

		// ConfigMap管理
		k8s.GET("/configmaps", middleware.K8sPermission("get"), resourceCtrl.ListConfigMaps)
		k8s.GET("/configmap/detail", middleware.K8sPermission("get"), resourceCtrl.GetConfigMap)
		k8s.POST("/configmap/create", middleware.K8sPermission("create"), resourceCtrl.CreateConfigMap)
		k8s.POST("/configmap/update", middleware.K8sPermission("update"), resourceCtrl.UpdateConfigMap)  // 改为POST
		k8s.POST("/configmap/delete", middleware.K8sPermission("delete"), resourceCtrl.DeleteConfigMap)  // 改为POST

		// Secret管理
		k8s.GET("/secrets", middleware.K8sPermission("get"), resourceCtrl.ListSecrets)
		k8s.GET("/secret/detail", middleware.K8sPermission("get"), resourceCtrl.GetSecret)
		k8s.POST("/secret/create", middleware.K8sPermission("create"), resourceCtrl.CreateSecret)
		k8s.POST("/secret/update", middleware.K8sPermission("update"), resourceCtrl.UpdateSecret)  // 改为POST
		k8s.POST("/secret/delete", middleware.K8sPermission("delete"), resourceCtrl.DeleteSecret)  // 改为POST

		// PV管理
		k8s.GET("/pvs", middleware.K8sPermission("get"), resourceCtrl.ListPVs)
		k8s.GET("/pv/detail", middleware.K8sPermission("get"), resourceCtrl.GetPV)
		k8s.POST("/pv/delete", middleware.K8sPermission("delete"), resourceCtrl.DeletePV)  // 改为POST

		// PVC管理
		k8s.GET("/pvcs", middleware.K8sPermission("get"), resourceCtrl.ListPVCs)
		k8s.GET("/pvc/detail", middleware.K8sPermission("get"), resourceCtrl.GetPVC)
		k8s.POST("/pvc/delete", middleware.K8sPermission("delete"), resourceCtrl.DeletePVC)  // 改为POST

		// StorageClass管理
		k8s.GET("/storageclasses", middleware.K8sPermission("get"), resourceCtrl.ListStorageClasses)
		k8s.GET("/storageclass/detail", middleware.K8sPermission("get"), resourceCtrl.GetStorageClass)

		// 节点管理
		k8s.GET("/nodes", middleware.K8sPermission("get"), resourceCtrl.ListNodes)
		k8s.GET("/node/detail", middleware.K8sPermission("get"), resourceCtrl.GetNode)

		// 事件查看
		k8s.GET("/events", middleware.K8sPermission("get"), resourceCtrl.ListEvents)
		k8s.GET("/events/object", middleware.K8sPermission("get"), resourceCtrl.GetEventsByObject)
	}
}
