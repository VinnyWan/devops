package k8s

import (
	"devops/common"
	k8sservice "devops/service/k8s"
	"strconv"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type ResourceController struct{}

func NewResourceController() *ResourceController {
	return &ResourceController{}
}

// Namespace相关接口

// ListNamespaces 获取命名空间列表
// @Summary 获取命名空间列表
// @Tags K8s-Namespace
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/namespaces [get]
func (ctrl *ResourceController) ListNamespaces(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}

	service := &k8sservice.NamespaceService{}
	namespaces, err := service.List(uint(clusterID))
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", namespaces)
}

// GetNamespace 获取命名空间详情
// @Summary 获取命名空间详情
// @Tags K8s-Namespace
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param name query string true "命名空间名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/namespace/detail [get]
func (ctrl *ResourceController) GetNamespace(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	name := c.Query("name")

	service := &k8sservice.NamespaceService{}
	namespace, err := service.Get(uint(clusterID), name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", namespace)
}

// CreateNamespace 创建命名空间
// @Summary 创建命名空间
// @Tags K8s-Namespace
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param data body object true "命名空间信息"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/namespaces [post]
func (ctrl *ResourceController) CreateNamespace(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}

	var namespace corev1.Namespace
	if err := c.ShouldBindJSON(&namespace); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	service := &k8sservice.NamespaceService{}
	ns, err := service.Create(uint(clusterID), &namespace)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "创建成功", ns)
}

// DeleteNamespace 删除命名空间
// @Summary 删除命名空间
// @Tags K8s-Namespace
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param name query string true "命名空间名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/namespace/detail [post]
func (ctrl *ResourceController) DeleteNamespace(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	name := c.Query("name")

	service := &k8sservice.NamespaceService{}
	if err := service.Delete(uint(clusterID), name); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "删除成功", nil)
}

// Deployment相关接口

// ListDeployments 获取Deployment列表
// @Summary 获取Deployment列表
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/deployments [get]
func (ctrl *ResourceController) ListDeployments(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	service := &k8sservice.WorkloadService{}
	deployments, err := service.ListDeployments(uint(clusterID), namespace)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", deployments)
}

// GetDeployment 获取Deployment详情
// @Summary 获取Deployment详情
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/deployment/detail [get]
func (ctrl *ResourceController) GetDeployment(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.WorkloadService{}
	deployment, err := service.GetDeployment(uint(clusterID), namespace, name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", deployment)
}

// CreateDeployment 创建Deployment
// @Summary 创建Deployment
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param data body object true "Deployment信息"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/deployment/create [post]
func (ctrl *ResourceController) CreateDeployment(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	var deployment appsv1.Deployment
	if err := c.ShouldBindJSON(&deployment); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	service := &k8sservice.WorkloadService{}
	deploy, err := service.CreateDeployment(uint(clusterID), namespace, &deployment)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "创建成功", deploy)
}

// UpdateDeployment 更新Deployment
// @Summary 更新Deployment
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment名称"
// @Param data body object true "Deployment信息"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/deployment/update [post]
func (ctrl *ResourceController) UpdateDeployment(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	var deployment appsv1.Deployment
	if err := c.ShouldBindJSON(&deployment); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	service := &k8sservice.WorkloadService{}
	deploy, err := service.UpdateDeployment(uint(clusterID), namespace, &deployment)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "更新成功", deploy)
}

// DeleteDeployment 删除Deployment
// @Summary 删除Deployment
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/deployment/delete [post]
func (ctrl *ResourceController) DeleteDeployment(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.WorkloadService{}
	if err := service.DeleteDeployment(uint(clusterID), namespace, name); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "删除成功", nil)
}

// ScaleDeployment 扩缩容Deployment
// @Summary 扩缩容Deployment
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment名称"
// @Param replicas query int true "副本数"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/deployment/scale [post]
func (ctrl *ResourceController) ScaleDeployment(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")
	replicasStr := c.Query("replicas")
	replicas, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "replicas格式错误")
		return
	}

	service := &k8sservice.WorkloadService{}
	if err := service.ScaleDeployment(uint(clusterID), namespace, name, int32(replicas)); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "扩缩容成功", nil)
}

// RestartDeployment 重启Deployment
// @Summary 重启Deployment
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/deployment/restart [post]
func (ctrl *ResourceController) RestartDeployment(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.WorkloadService{}
	if err := service.RestartDeployment(uint(clusterID), namespace, name); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "重启成功", nil)
}

// Pod相关接口

// ListPods 获取Pod列表
// @Summary 获取Pod列表
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param labelSelector query string false "标签选择器"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/pods [get]
func (ctrl *ResourceController) ListPods(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	labelSelector := c.Query("labelSelector")

	service := &k8sservice.WorkloadService{}
	pods, err := service.ListPods(uint(clusterID), namespace, labelSelector)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", pods)
}

// GetPod 获取Pod详情
// @Summary 获取Pod详情
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Pod名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/pod/detail [get]
func (ctrl *ResourceController) GetPod(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.WorkloadService{}
	pod, err := service.GetPod(uint(clusterID), namespace, name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", pod)
}

// DeletePod 删除Pod
// @Summary 删除Pod
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Pod名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/pod/delete [post]
func (ctrl *ResourceController) DeletePod(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.WorkloadService{}
	if err := service.DeletePod(uint(clusterID), namespace, name); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "删除成功", nil)
}

// GetPodLogs 获取Pod日志
// @Summary 获取Pod日志
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Pod名称"
// @Param container query string false "容器名称"
// @Param tailLines query int false "尾部行数" default(100)
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/pod/logs [get]
func (ctrl *ResourceController) GetPodLogs(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")
	container := c.Query("container")
	tailLinesStr := c.DefaultQuery("tailLines", "100")
	tailLines, err := strconv.ParseInt(tailLinesStr, 10, 64)
	if err != nil {
		common.BadRequest(c, "tailLines格式错误")
		return
	}

	service := &k8sservice.WorkloadService{}
	logs, err := service.GetPodLogs(uint(clusterID), namespace, name, container, tailLines)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", map[string]string{"logs": logs})
}

// StatefulSet相关

// ListStatefulSets 获取StatefulSet列表
// @Summary 获取StatefulSet列表
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/statefulsets [get]
func (ctrl *ResourceController) ListStatefulSets(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	service := &k8sservice.WorkloadService{}
	statefulsets, err := service.ListStatefulSets(uint(clusterID), namespace)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", statefulsets)
}

// DaemonSet相关

// ListDaemonSets 获取DaemonSet列表
// @Summary 获取DaemonSet列表
// @Tags K8s-Workload
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/daemonsets [get]
func (ctrl *ResourceController) ListDaemonSets(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	service := &k8sservice.WorkloadService{}
	daemonsets, err := service.ListDaemonSets(uint(clusterID), namespace)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", daemonsets)
}

// Service相关接口

// ListServices 获取Service列表
// @Summary 获取Service列表
// @Tags K8s-Service
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/services [get]
func (ctrl *ResourceController) ListServices(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	service := &k8sservice.K8sServiceService{}
	services, err := service.ListServices(uint(clusterID), namespace)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", services)
}

// GetService 获取Service详情
// @Summary 获取Service详情
// @Tags K8s-Service
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Service名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/service/detail [get]
func (ctrl *ResourceController) GetService(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.K8sServiceService{}
	svc, err := service.GetService(uint(clusterID), namespace, name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", svc)
}

// CreateService 创建Service
// @Summary 创建Service
// @Tags K8s-Service
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param data body object true "Service信息"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/services [post]
func (ctrl *ResourceController) CreateService(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	var svc corev1.Service
	if err := c.ShouldBindJSON(&svc); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	service := &k8sservice.K8sServiceService{}
	result, err := service.CreateService(uint(clusterID), namespace, &svc)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "创建成功", result)
}

// DeleteService 删除Service
// @Summary 删除Service
// @Tags K8s-Service
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Service名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/service/detail [post]
func (ctrl *ResourceController) DeleteService(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.K8sServiceService{}
	if err := service.DeleteService(uint(clusterID), namespace, name); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "删除成功", nil)
}

// Ingress相关接口

// ListIngresses 获取Ingress列表
// @Summary 获取Ingress列表
// @Tags K8s-Service
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/ingresses [get]
func (ctrl *ResourceController) ListIngresses(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	service := &k8sservice.K8sServiceService{}
	ingresses, err := service.ListIngresses(uint(clusterID), namespace)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", ingresses)
}

// GetIngress 获取Ingress详情
// @Summary 获取Ingress详情
// @Tags K8s-Service
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/ingress/detail [get]
func (ctrl *ResourceController) GetIngress(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.K8sServiceService{}
	ingress, err := service.GetIngress(uint(clusterID), namespace, name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", ingress)
}

// CreateIngress 创建Ingress
// @Summary 创建Ingress
// @Tags K8s-Service
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param data body object true "Ingress信息"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/ingresses [post]
func (ctrl *ResourceController) CreateIngress(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	var ingress networkingv1.Ingress
	if err := c.ShouldBindJSON(&ingress); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	service := &k8sservice.K8sServiceService{}
	ing, err := service.CreateIngress(uint(clusterID), namespace, &ingress)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "创建成功", ing)
}

// DeleteIngress 删除Ingress
// @Summary 删除Ingress
// @Tags K8s-Service
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/ingress/detail [post]
func (ctrl *ResourceController) DeleteIngress(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.K8sServiceService{}
	if err := service.DeleteIngress(uint(clusterID), namespace, name); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "删除成功", nil)
}

// ConfigMap相关接口

// ListConfigMaps 获取ConfigMap列表
// @Summary 获取ConfigMap列表
// @Tags K8s-Config
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/configmaps [get]
func (ctrl *ResourceController) ListConfigMaps(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	service := &k8sservice.ConfigService{}
	configMaps, err := service.ListConfigMaps(uint(clusterID), namespace)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", configMaps)
}

// GetConfigMap 获取ConfigMap详情
// @Summary 获取ConfigMap详情
// @Tags K8s-Config
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "ConfigMap名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/configmap/detail [get]
func (ctrl *ResourceController) GetConfigMap(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.ConfigService{}
	configMap, err := service.GetConfigMap(uint(clusterID), namespace, name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", configMap)
}

// CreateConfigMap 创建ConfigMap
// @Summary 创建ConfigMap
// @Tags K8s-Config
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param data body object true "ConfigMap信息"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/configmaps [post]
func (ctrl *ResourceController) CreateConfigMap(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	var configMap corev1.ConfigMap
	if err := c.ShouldBindJSON(&configMap); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	service := &k8sservice.ConfigService{}
	cm, err := service.CreateConfigMap(uint(clusterID), namespace, &configMap)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "创建成功", cm)
}

// UpdateConfigMap 更新ConfigMap
// @Summary 更新ConfigMap
// @Tags K8s-Config
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "ConfigMap名称"
// @Param data body object true "ConfigMap信息"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/configmap/update [post]
func (ctrl *ResourceController) UpdateConfigMap(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	var configMap corev1.ConfigMap
	if err := c.ShouldBindJSON(&configMap); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	service := &k8sservice.ConfigService{}
	cm, err := service.UpdateConfigMap(uint(clusterID), namespace, &configMap)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "更新成功", cm)
}

// DeleteConfigMap 删除ConfigMap
// @Summary 删除ConfigMap
// @Tags K8s-Config
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "ConfigMap名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/configmap/delete [post]
func (ctrl *ResourceController) DeleteConfigMap(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.ConfigService{}
	if err := service.DeleteConfigMap(uint(clusterID), namespace, name); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "删除成功", nil)
}

// Secret相关接口

// ListSecrets 获取Secret列表
// @Summary 获取Secret列表
// @Tags K8s-Config
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/secrets [get]
func (ctrl *ResourceController) ListSecrets(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	service := &k8sservice.ConfigService{}
	secrets, err := service.ListSecrets(uint(clusterID), namespace)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", secrets)
}

// GetSecret 获取Secret详情
// @Summary 获取Secret详情
// @Tags K8s-Config
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Secret名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/secret/detail [get]
func (ctrl *ResourceController) GetSecret(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.ConfigService{}
	secret, err := service.GetSecret(uint(clusterID), namespace, name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", secret)
}

// CreateSecret 创建Secret
// @Summary 创建Secret
// @Tags K8s-Config
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param data body object true "Secret信息"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/secrets [post]
func (ctrl *ResourceController) CreateSecret(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	var secret corev1.Secret
	if err := c.ShouldBindJSON(&secret); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	service := &k8sservice.ConfigService{}
	sec, err := service.CreateSecret(uint(clusterID), namespace, &secret)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "创建成功", sec)
}

// UpdateSecret 更新Secret
// @Summary 更新Secret
// @Tags K8s-Config
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Secret名称"
// @Param data body object true "Secret信息"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/secret/update [post]
func (ctrl *ResourceController) UpdateSecret(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	var secret corev1.Secret
	if err := c.ShouldBindJSON(&secret); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	service := &k8sservice.ConfigService{}
	sec, err := service.UpdateSecret(uint(clusterID), namespace, &secret)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "更新成功", sec)
}

// DeleteSecret 删除Secret
// @Summary 删除Secret
// @Tags K8s-Config
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "Secret名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/secret/delete [post]
func (ctrl *ResourceController) DeleteSecret(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.ConfigService{}
	if err := service.DeleteSecret(uint(clusterID), namespace, name); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "删除成功", nil)
}

// 存储相关接口

// ListPVs 获取PV列表
// @Summary 获取PV列表
// @Tags K8s-Storage
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/pvs [get]
func (ctrl *ResourceController) ListPVs(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}

	service := &k8sservice.StorageService{}
	pvs, err := service.ListPVs(uint(clusterID))
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", pvs)
}

// GetPV 获取PV详情
// @Summary 获取PV详情
// @Tags K8s-Storage
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param name query string true "PV名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/pv/detail [get]
func (ctrl *ResourceController) GetPV(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	name := c.Query("name")

	service := &k8sservice.StorageService{}
	pv, err := service.GetPV(uint(clusterID), name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", pv)
}

// DeletePV 删除PV
// @Summary 删除PV
// @Tags K8s-Storage
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param name query string true "PV名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/pv/detail [post]
func (ctrl *ResourceController) DeletePV(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	name := c.Query("name")

	service := &k8sservice.StorageService{}
	if err := service.DeletePV(uint(clusterID), name); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "删除成功", nil)
}

// ListPVCs 获取PVC列表
// @Summary 获取PVC列表
// @Tags K8s-Storage
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/pvcs [get]
func (ctrl *ResourceController) ListPVCs(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	service := &k8sservice.StorageService{}
	pvcs, err := service.ListPVCs(uint(clusterID), namespace)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", pvcs)
}

// GetPVC 获取PVC详情
// @Summary 获取PVC详情
// @Tags K8s-Storage
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "PVC名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/pvc/detail [get]
func (ctrl *ResourceController) GetPVC(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.StorageService{}
	pvc, err := service.GetPVC(uint(clusterID), namespace, name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", pvc)
}

// DeletePVC 删除PVC
// @Summary 删除PVC
// @Tags K8s-Storage
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "PVC名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/pvc/detail [post]
func (ctrl *ResourceController) DeletePVC(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	name := c.Query("name")

	service := &k8sservice.StorageService{}
	if err := service.DeletePVC(uint(clusterID), namespace, name); err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "删除成功", nil)
}

// ListStorageClasses 获取StorageClass列表
// @Summary 获取StorageClass列表
// @Tags K8s-Storage
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/storageclasses [get]
func (ctrl *ResourceController) ListStorageClasses(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}

	service := &k8sservice.StorageService{}
	scs, err := service.ListStorageClasses(uint(clusterID))
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", scs)
}

// GetStorageClass 获取StorageClass详情
// @Summary 获取StorageClass详情
// @Tags K8s-Storage
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param name query string true "StorageClass名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/storageclass/detail [get]
func (ctrl *ResourceController) GetStorageClass(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	name := c.Query("name")

	service := &k8sservice.StorageService{}
	sc, err := service.GetStorageClass(uint(clusterID), name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", sc)
}

// 节点相关接口

// ListNodes 获取节点列表
// @Summary 获取节点列表
// @Tags K8s-Node
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/nodes [get]
func (ctrl *ResourceController) ListNodes(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}

	service := &k8sservice.NodeService{}
	nodes, err := service.ListNodes(uint(clusterID))
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", nodes)
}

// GetNode 获取节点详情
// @Summary 获取节点详情
// @Tags K8s-Node
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param name query string true "节点名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/node/detail [get]
func (ctrl *ResourceController) GetNode(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	name := c.Query("name")

	service := &k8sservice.NodeService{}
	node, err := service.GetNode(uint(clusterID), name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", node)
}

// 事件相关接口

// ListEvents 获取事件列表
// @Summary 获取事件列表
// @Tags K8s-Event
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/events [get]
func (ctrl *ResourceController) ListEvents(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")

	service := &k8sservice.EventService{}
	events, err := service.ListEvents(uint(clusterID), namespace)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", events)
}

// GetEventsByObject 获取指定资源的事件
// @Summary 获取指定资源的事件
// @Tags K8s-Event
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param kind query string true "资源类型"
// @Param name query string true "资源名称"
// @Success 200 {object} common.Response
// @Router /api/v1/k8s/events/object [get]
func (ctrl *ResourceController) GetEventsByObject(c *gin.Context) {
	clusterID, ok := common.RequireUintQuery(c, "clusterId")
	if !ok {
		return
	}
	namespace := c.Query("namespace")
	kind := c.Query("kind")
	name := c.Query("name")

	service := &k8sservice.EventService{}
	events, err := service.GetEventsByObject(uint(clusterID), namespace, kind, name)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.SuccessWithMsg(c, "获取成功", events)
}
