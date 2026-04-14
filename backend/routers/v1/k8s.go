package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/k8s/api"

	"github.com/gin-gonic/gin"
)

func registerCluster(r *gin.RouterGroup) {
	g := r.Group("/k8s")
	listPermission := middleware.RequirePermission("cluster", "list")
	createPermission := middleware.RequirePermission("cluster", "create")
	updatePermission := middleware.RequirePermission("cluster", "update")
	deletePermission := middleware.RequirePermission("cluster", "delete")
	{
		// k8s  cluster
		g.GET("/cluster/list", listPermission, api.ClusterList)
		g.GET("/cluster/detail", listPermission, api.ClusterDetail)
		g.GET("/cluster/default", listPermission, api.ClusterDefault)
		g.GET("/cluster/search", listPermission, api.ClusterSearch)
		g.POST("/cluster/create",
			createPermission,
			middleware.SetAuditOperation("创建集群"),
			middleware.SetAuditRetention(7), // 此接口保留 7 天
			api.ClusterCreate)
		g.POST("/cluster/update",
			updatePermission,
			middleware.SetAuditOperation("更新集群"),
			api.ClusterUpdate)
		g.POST("/cluster/delete",
			deletePermission,
			middleware.SetAuditOperation("删除集群"),
			api.ClusterDelete)
		g.POST("/cluster/set-default", updatePermission, api.ClusterSetDefault)
		g.GET("/cluster/health", listPermission, api.ClusterHealthCheck)

		// Cluster Stats
		g.GET("/cluster/stats/workload", listPermission, api.ClusterWorkloadStats)
		g.GET("/cluster/stats/network", listPermission, api.ClusterNetworkStats)
		g.GET("/cluster/stats/storage", listPermission, api.ClusterStorageStats)
		g.GET("/cluster/nodes", listPermission, api.ClusterNodes)
		g.GET("/cluster/events", listPermission, api.ClusterEvents)

		// Namespace管理
		g.GET("/namespace/list", listPermission, api.NamespaceList)
		g.POST("/namespace/create",
			createPermission,
			middleware.SetAuditOperation("创建Namespace"),
			api.CreateNamespace)
		g.POST("/namespace/delete",
			deletePermission,
			middleware.SetAuditOperation("删除Namespace"),
			api.DeleteNamespace)

		// Deployment
		g.GET("/deployment/list", listPermission, api.ListDeployments)
		g.GET("/deployment/detail", listPermission, api.GetDeploymentDetail)
		g.GET("/deployment/pods", listPermission, api.GetDeploymentPods)
		g.GET("/deployment/yaml", listPermission, api.GetDeploymentYAML)
		g.POST("/deployment/create",
			createPermission,
			middleware.SetAuditOperation("创建Deployment"),
			api.CreateDeployment)
		g.POST("/deployment/update",
			updatePermission,
			middleware.SetAuditOperation("更新Deployment"),
			api.UpdateDeployment)
		g.POST("/deployment/yaml/update",
			updatePermission,
			middleware.SetAuditOperation("YAML更新Deployment"),
			api.UpdateDeploymentYAML)
		g.POST("/deployment/restart",
			updatePermission,
			middleware.SetAuditOperation("重启Deployment"),
			api.RestartDeployment)
		g.POST("/deployment/scale",
			updatePermission,
			middleware.SetAuditOperation("扩缩容Deployment"),
			api.ScaleDeployment)
		g.POST("/deployment/delete",
			deletePermission,
			middleware.SetAuditOperation("删除Deployment"),
			api.DeleteDeployment)

		// StatefulSet
		g.GET("/statefulset/list", listPermission, api.ListStatefulSets)
		g.GET("/statefulset/detail", listPermission, api.GetStatefulSetDetail)
		g.GET("/statefulset/yaml", listPermission, api.GetStatefulSetYAML)
		g.POST("/statefulset/create",
			createPermission,
			middleware.SetAuditOperation("创建StatefulSet"),
			api.CreateStatefulSet)
		g.POST("/statefulset/update",
			updatePermission,
			middleware.SetAuditOperation("更新StatefulSet"),
			api.UpdateStatefulSet)
		g.POST("/statefulset/yaml/update",
			updatePermission,
			middleware.SetAuditOperation("YAML更新StatefulSet"),
			api.UpdateStatefulSetYAML)
		g.POST("/statefulset/restart",
			updatePermission,
			middleware.SetAuditOperation("重启StatefulSet"),
			api.RestartStatefulSet)
		g.POST("/statefulset/scale",
			updatePermission,
			middleware.SetAuditOperation("扩缩容StatefulSet"),
			api.ScaleStatefulSet)
		g.POST("/statefulset/delete",
			deletePermission,
			middleware.SetAuditOperation("删除StatefulSet"),
			api.DeleteStatefulSet)

		// DaemonSet
		g.GET("/daemonset/list", listPermission, api.ListDaemonSets)
		g.GET("/daemonset/detail", listPermission, api.GetDaemonSetDetail)
		g.GET("/daemonset/yaml", listPermission, api.GetDaemonSetYAML)
		g.POST("/daemonset/create",
			createPermission,
			middleware.SetAuditOperation("创建DaemonSet"),
			api.CreateDaemonSet)
		g.POST("/daemonset/update",
			updatePermission,
			middleware.SetAuditOperation("更新DaemonSet"),
			api.UpdateDaemonSet)
		g.POST("/daemonset/yaml/update",
			updatePermission,
			middleware.SetAuditOperation("YAML更新DaemonSet"),
			api.UpdateDaemonSetYAML)
		g.POST("/daemonset/restart",
			updatePermission,
			middleware.SetAuditOperation("重启DaemonSet"),
			api.RestartDaemonSet)
		g.POST("/daemonset/delete",
			deletePermission,
			middleware.SetAuditOperation("删除DaemonSet"),
			api.DeleteDaemonSet)

		// Job 路由
		g.GET("/job/list", listPermission, api.ListJobs)
		g.GET("/job/detail", listPermission, api.GetJobDetail)
		g.GET("/job/pods", listPermission, api.GetJobPods)
		g.GET("/job/yaml", listPermission, api.GetJobYAML)
		g.POST("/job/create", createPermission, middleware.SetAuditOperation("创建Job"), api.CreateJob)
		g.POST("/job/delete", deletePermission, middleware.SetAuditOperation("删除Job"), api.DeleteJob)

		// CronJob 路由
		g.GET("/cronjob/list", listPermission, api.ListCronJobs)
		g.GET("/cronjob/detail", listPermission, api.GetCronJobDetail)
		g.GET("/cronjob/pods", listPermission, api.GetCronJobPods)
		g.GET("/cronjob/yaml", listPermission, api.GetCronJobYAML)
		g.POST("/cronjob/create", createPermission, middleware.SetAuditOperation("创建CronJob"), api.CreateCronJob)
		g.POST("/cronjob/yaml/update", updatePermission, middleware.SetAuditOperation("YAML更新CronJob"), api.UpdateCronJobYAML)
		g.POST("/cronjob/suspend", updatePermission, middleware.SetAuditOperation("暂停恢复CronJob"), api.SuspendCronJob)
		g.POST("/cronjob/delete", deletePermission, middleware.SetAuditOperation("删除CronJob"), api.DeleteCronJob)

		// Pod
		g.GET("/pod/list", listPermission, api.ListPods)
		g.GET("/pod/list_by_owner", listPermission, api.ListPodsByOwner)
		g.GET("/pod/detail", listPermission, api.GetPodDetail)
		g.GET("/pod/describe", listPermission, api.DescribePod)
		g.GET("/pod/yaml", listPermission, api.GetPodYAML)
		g.POST("/pod/yaml/update",
			updatePermission,
			middleware.SetAuditOperation("YAML更新Pod"),
			api.UpdatePodYAML)
		g.GET("/pod/logs", listPermission, api.GetPodLogs)
		g.GET("/pod/events", listPermission, api.GetPodEvents)
		g.GET("/pod/detect-shell", listPermission, api.DetectPodShell)
		g.GET("/pod/terminal", listPermission, api.PodTerminal)
		g.POST("/pod/create",
			createPermission,
			middleware.SetAuditOperation("创建Pod"),
			api.CreatePod)
		g.POST("/pod/update",
			updatePermission,
			middleware.SetAuditOperation("更新Pod"),
			api.UpdatePod)
		g.POST("/pod/delete",
			deletePermission,
			middleware.SetAuditOperation("删除Pod"),
			api.DeletePod)

		// Service
		g.GET("/service/list", listPermission, api.ListServices)
		g.GET("/service/detail", listPermission, api.GetServiceDetail)
		g.POST("/service/create",
			createPermission,
			middleware.SetAuditOperation("创建Service"),
			api.CreateService)
		g.POST("/service/update",
			updatePermission,
			middleware.SetAuditOperation("更新Service"),
			api.UpdateService)
		g.POST("/service/delete",
			deletePermission,
			middleware.SetAuditOperation("删除Service"),
			api.DeleteService)
			g.POST("/service/yaml/update",
				updatePermission,
				middleware.SetAuditOperation("YAML更新Service"),
				api.UpdateServiceYAML)

			// ConfigMap
		g.GET("/configmap/list", listPermission, api.ListConfigMaps)
		g.GET("/configmap/detail", listPermission, api.GetConfigMapDetail)
		g.GET("/configmap/yaml", listPermission, api.GetConfigMapYAML)
		g.POST("/configmap/create",
			createPermission,
			middleware.SetAuditOperation("创建ConfigMap"),
			api.CreateConfigMap)
		g.POST("/configmap/update",
			updatePermission,
			middleware.SetAuditOperation("更新ConfigMap"),
			api.UpdateConfigMap)
		g.POST("/configmap/yaml/update",
			updatePermission,
			middleware.SetAuditOperation("YAML更新ConfigMap"),
			api.UpdateConfigMapYAML)
		g.POST("/configmap/delete",
			deletePermission,
			middleware.SetAuditOperation("删除ConfigMap"),
			api.DeleteConfigMap)

		// Ingress
		g.GET("/ingress/list", listPermission, api.ListIngresses)
		g.GET("/ingress/detail", listPermission, api.GetIngressDetail)
		g.POST("/ingress/create",
			createPermission,
			middleware.SetAuditOperation("创建Ingress"),
			api.CreateIngress)
		g.POST("/ingress/update",
			updatePermission,
			middleware.SetAuditOperation("更新Ingress"),
			api.UpdateIngress)
		g.POST("/ingress/delete",
			deletePermission,
			middleware.SetAuditOperation("删除Ingress"),
			api.DeleteIngress)
			g.POST("/ingress/yaml/update",
				updatePermission,
				middleware.SetAuditOperation("YAML更新Ingress"),
				api.UpdateIngressYAML)

			// Node Management
		g.GET("/nodes", listPermission, api.NodeList)
		g.GET("/node/detail", listPermission, api.GetNodeDetail)
		g.GET("/node/events", listPermission, api.GetNodeEvents)
		g.POST("/node/cordon",
			updatePermission,
			middleware.SetAuditOperation("设置节点调度"),
			api.CordonNode)
		g.POST("/node/drain",
			updatePermission,
			middleware.SetAuditOperation("驱逐节点"),
			api.DrainNode)
		g.POST("/node/labels",
			updatePermission,
			middleware.SetAuditOperation("更新节点标签"),
			api.UpdateNodeLabels)
		g.POST("/node/taints",
			updatePermission,
			middleware.SetAuditOperation("更新节点污点"),
			api.UpdateNodeTaints)

			// Storage
		g.GET("/storageclass/list", listPermission, api.ListStorageClasses)
		g.POST("/storageclass/yaml/update",
			updatePermission,
			middleware.SetAuditOperation("YAML更新StorageClass"),
			api.UpdateStorageClassYAML)
		g.GET("/pv/list", listPermission, api.ListPersistentVolumes)
		g.POST("/pv/yaml/update",
			updatePermission,
			middleware.SetAuditOperation("YAML更新PV"),
			api.UpdatePVYAML)
		g.GET("/pvc/list", listPermission, api.ListPersistentVolumeClaims)
		g.POST("/pvc/yaml/update",
			updatePermission,
			middleware.SetAuditOperation("YAML更新PVC"),
			api.UpdatePVCYAML)
		g.POST("/pvc/delete",
			deletePermission,
			middleware.SetAuditOperation("删除PVC"),
			api.DeletePVC)

			// 通用资源YAML接口
		g.GET("/resource/yaml", listPermission, api.GetResourceYAML)
		g.GET("/resource/types", listPermission, api.GetSupportedResourceTypes)
	}
}
