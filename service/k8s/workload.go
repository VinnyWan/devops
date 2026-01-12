package k8s

import (
	"context"
	"strings"

	k8smodels "devops/models/k8s"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkloadService 工作负载服务
type WorkloadService struct {
	clusterService *ClusterService
}

// extractImageVersion 从镜像地址中提取版本号（冒号后的部分）
// 例如：nginx:1.21 -> 1.21
//
//	registry.cn-hangzhou.aliyuncs.com/my-app:v2.0.1 -> v2.0.1
//	redis -> latest（如果没有冒号，返回 latest）
func extractImageVersion(image string) string {
	// 查找最后一个冒号的位置
	lastColonIndex := strings.LastIndex(image, ":")

	// 如果没有冒号，说明没有指定版本，返回 latest
	if lastColonIndex == -1 {
		return "latest"
	}

	// 提取冒号后的部分作为版本号
	version := image[lastColonIndex+1:]

	// 如果版本号为空，返回 latest
	if version == "" {
		return "latest"
	}

	return version
}

// Deployment相关方法

// ListDeployments 获取Deployment列表（简化版）
// 返回字段：名称、命名空间、副本数、镜像、标签、创建时间、更新时间
// 如需完整信息，使用 GetDeployment() 方法获取单个详情
func (s *WorkloadService) ListDeployments(clusterID uint, namespace string) ([]k8smodels.DeploymentDTO, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 转换为 DTO
	result := make([]k8smodels.DeploymentDTO, 0, len(deployments.Items))
	for _, deploy := range deployments.Items {
		// 提取所有容器的镜像版本号（冒号后的部分）
		versions := make([]string, 0)
		for _, container := range deploy.Spec.Template.Spec.Containers {
			version := extractImageVersion(container.Image)
			versions = append(versions, version)
		}

		// 获取副本数
		replicas := int32(0)
		if deploy.Spec.Replicas != nil {
			replicas = *deploy.Spec.Replicas
		}

		// 获取更新时间（如果没有 conditions，使用创建时间）
		updateTime := deploy.CreationTimestamp.Time
		if len(deploy.Status.Conditions) > 0 {
			updateTime = deploy.Status.Conditions[len(deploy.Status.Conditions)-1].LastUpdateTime.Time
		}

		result = append(result, k8smodels.DeploymentDTO{
			Name:       deploy.Name,
			Namespace:  deploy.Namespace,
			Replicas:   replicas,
			Versions:   versions,
			Labels:     deploy.Labels,
			CreateTime: deploy.CreationTimestamp.Time,
			UpdateTime: updateTime,
		})
	}

	return result, nil
}

// GetDeployment 获取Deployment详情
func (s *WorkloadService) GetDeployment(clusterID uint, namespace, name string) (*appsv1.Deployment, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

// CreateDeployment 创建Deployment
func (s *WorkloadService) CreateDeployment(clusterID uint, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	deploy, err := clientset.AppsV1().Deployments(namespace).Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return deploy, nil
}

// UpdateDeployment 更新Deployment
func (s *WorkloadService) UpdateDeployment(clusterID uint, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	deploy, err := clientset.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return deploy, nil
}

// DeleteDeployment 删除Deployment
func (s *WorkloadService) DeleteDeployment(clusterID uint, namespace, name string) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	return clientset.AppsV1().Deployments(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

// ScaleDeployment 扩缩容Deployment
func (s *WorkloadService) ScaleDeployment(clusterID uint, namespace, name string, replicas int32) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	deployment.Spec.Replicas = &replicas
	_, err = clientset.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	return err
}

// RestartDeployment 重启Deployment
func (s *WorkloadService) RestartDeployment(clusterID uint, namespace, name string) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// 通过更新注解触发重启
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().String()

	_, err = clientset.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	return err
}

// Pod相关方法

// ListPods 获取Pod列表
func (s *WorkloadService) ListPods(clusterID uint, namespace string, labelSelector string) ([]corev1.Pod, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

// GetPod 获取Pod详情
func (s *WorkloadService) GetPod(clusterID uint, namespace, name string) (*corev1.Pod, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// DeletePod 删除Pod
func (s *WorkloadService) DeletePod(clusterID uint, namespace, name string) error {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return err
	}

	return clientset.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

// GetPodLogs 获取Pod日志
func (s *WorkloadService) GetPodLogs(clusterID uint, namespace, name, container string, tailLines int64) (string, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return "", err
	}

	opts := &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLines,
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
	logs, err := req.DoRaw(context.Background())
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

// StatefulSet相关方法

// ListStatefulSets 获取StatefulSet列表
func (s *WorkloadService) ListStatefulSets(clusterID uint, namespace string) ([]appsv1.StatefulSet, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return statefulsets.Items, nil
}

// DaemonSet相关方法

// ListDaemonSets 获取DaemonSet列表
func (s *WorkloadService) ListDaemonSets(clusterID uint, namespace string) ([]appsv1.DaemonSet, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	daemonsets, err := clientset.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return daemonsets.Items, nil
}
