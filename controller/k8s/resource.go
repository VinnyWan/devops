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
// @Param clusterId path int true "集群ID"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/namespaces [get]
func (ctrl *ResourceController) ListNamespaces(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)

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
// @Param clusterId path int true "集群ID"
// @Param name path string true "命名空间名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/namespaces/{name} [get]
func (ctrl *ResourceController) GetNamespace(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param data body object true "命名空间信息"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/namespaces [post]
func (ctrl *ResourceController) CreateNamespace(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)

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
// @Param clusterId path int true "集群ID"
// @Param name path string true "命名空间名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/namespaces/{name} [delete]
func (ctrl *ResourceController) DeleteNamespace(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/deployments [get]
func (ctrl *ResourceController) ListDeployments(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Deployment名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/deployments/{name} [get]
func (ctrl *ResourceController) GetDeployment(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param data body object true "Deployment信息"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/deployments [post]
func (ctrl *ResourceController) CreateDeployment(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Deployment名称"
// @Param data body object true "Deployment信息"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/deployments/{name} [put]
func (ctrl *ResourceController) UpdateDeployment(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Deployment名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/deployments/{name} [delete]
func (ctrl *ResourceController) DeleteDeployment(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Deployment名称"
// @Param replicas query int true "副本数"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/deployments/{name}/scale [post]
func (ctrl *ResourceController) ScaleDeployment(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")
	replicas, _ := strconv.ParseInt(c.Query("replicas"), 10, 32)

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Deployment名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/deployments/{name}/restart [post]
func (ctrl *ResourceController) RestartDeployment(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param labelSelector query string false "标签选择器"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/pods [get]
func (ctrl *ResourceController) ListPods(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Pod名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/pods/{name} [get]
func (ctrl *ResourceController) GetPod(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Pod名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/pods/{name} [delete]
func (ctrl *ResourceController) DeletePod(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Pod名称"
// @Param container query string false "容器名称"
// @Param tailLines query int false "尾部行数" default(100)
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/pods/{name}/logs [get]
func (ctrl *ResourceController) GetPodLogs(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")
	container := c.Query("container")
	tailLines, _ := strconv.ParseInt(c.DefaultQuery("tailLines", "100"), 10, 64)

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/statefulsets [get]
func (ctrl *ResourceController) ListStatefulSets(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/daemonsets [get]
func (ctrl *ResourceController) ListDaemonSets(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/services [get]
func (ctrl *ResourceController) ListServices(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Service名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/services/{name} [get]
func (ctrl *ResourceController) GetService(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param data body object true "Service信息"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/services [post]
func (ctrl *ResourceController) CreateService(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Service名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/services/{name} [delete]
func (ctrl *ResourceController) DeleteService(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/ingresses [get]
func (ctrl *ResourceController) ListIngresses(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Ingress名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/ingresses/{name} [get]
func (ctrl *ResourceController) GetIngress(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param data body object true "Ingress信息"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/ingresses [post]
func (ctrl *ResourceController) CreateIngress(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Ingress名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/ingresses/{name} [delete]
func (ctrl *ResourceController) DeleteIngress(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/configmaps [get]
func (ctrl *ResourceController) ListConfigMaps(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "ConfigMap名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/configmaps/{name} [get]
func (ctrl *ResourceController) GetConfigMap(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param data body object true "ConfigMap信息"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/configmaps [post]
func (ctrl *ResourceController) CreateConfigMap(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "ConfigMap名称"
// @Param data body object true "ConfigMap信息"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/configmaps/{name} [put]
func (ctrl *ResourceController) UpdateConfigMap(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "ConfigMap名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/configmaps/{name} [delete]
func (ctrl *ResourceController) DeleteConfigMap(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/secrets [get]
func (ctrl *ResourceController) ListSecrets(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Secret名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/secrets/{name} [get]
func (ctrl *ResourceController) GetSecret(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param data body object true "Secret信息"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/secrets [post]
func (ctrl *ResourceController) CreateSecret(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Secret名称"
// @Param data body object true "Secret信息"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/secrets/{name} [put]
func (ctrl *ResourceController) UpdateSecret(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "Secret名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/secrets/{name} [delete]
func (ctrl *ResourceController) DeleteSecret(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/pvs [get]
func (ctrl *ResourceController) ListPVs(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)

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
// @Param clusterId path int true "集群ID"
// @Param name path string true "PV名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/pvs/{name} [get]
func (ctrl *ResourceController) GetPV(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param name path string true "PV名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/pvs/{name} [delete]
func (ctrl *ResourceController) DeletePV(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/pvcs [get]
func (ctrl *ResourceController) ListPVCs(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "PVC名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/pvcs/{name} [get]
func (ctrl *ResourceController) GetPVC(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name path string true "PVC名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/pvcs/{name} [delete]
func (ctrl *ResourceController) DeletePVC(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	namespace := c.Query("namespace")
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/storageclasses [get]
func (ctrl *ResourceController) ListStorageClasses(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)

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
// @Param clusterId path int true "集群ID"
// @Param name path string true "StorageClass名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/storageclasses/{name} [get]
func (ctrl *ResourceController) GetStorageClass(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/nodes [get]
func (ctrl *ResourceController) ListNodes(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)

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
// @Param clusterId path int true "集群ID"
// @Param name path string true "节点名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/nodes/{name} [get]
func (ctrl *ResourceController) GetNode(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	name := c.Param("name")

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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/events [get]
func (ctrl *ResourceController) ListEvents(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
// @Param clusterId path int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param kind query string true "资源类型"
// @Param name query string true "资源名称"
// @Success 200 {object} common.Response
// @Router /api/k8s/clusters/{clusterId}/events/object [get]
func (ctrl *ResourceController) GetEventsByObject(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
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
