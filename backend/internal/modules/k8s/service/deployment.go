package service

import (
	"context"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentConditionVO 表示 Deployment 的状态条件
type DeploymentConditionVO struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	LastUpdateTime     time.Time `json:"lastUpdateTime"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
}

// ResourceSummary 聚合所有容器的资源请求与限制
type ResourceSummary struct {
	CPURequest    string `json:"cpuRequest"` // 如 "500m"，未设置为空
	CPULimit      string `json:"cpuLimit"`
	MemoryRequest string `json:"memoryRequest"` // 如 "256Mi"，未设置为空
	MemoryLimit   string `json:"memoryLimit"`
}

type ContainerInfo struct {
	Name      string `json:"name"`
	Image     string `json:"image"`
	Resources string `json:"resources"`
}

type DeploymentListVO struct {
	Name               string            `json:"name"`
	Namespace          string            `json:"namespace"`
	Replicas           int32             `json:"replicas"`
	ReadyReplicas      int32             `json:"readyReplicas"`
	UpdatedReplicas    int32             `json:"updatedReplicas"`
	AvailableReplicas  int32             `json:"availableReplicas"`
	Labels             map[string]string `json:"labels"`
	Containers         []ContainerInfo   `json:"containers"`
	ResourceSummary    ResourceSummary   `json:"resourceSummary"`
	Strategy           string            `json:"strategy"`
	Status             string            `json:"status"`       // Progressing / Available / Unavailable / Degraded
	StatusReason       string            `json:"statusReason"` // 状态原因
	Generation         int64             `json:"generation"`
	ObservedGeneration int64             `json:"observedGeneration"`
	CreatedAt          time.Time         `json:"createdAt"`
}

type DeploymentVO struct {
	Name               string                  `json:"name"`
	Namespace          string                  `json:"namespace"`
	Replicas           int32                   `json:"replicas"`
	ReadyReplicas      int32                   `json:"readyReplicas"`
	UpdatedReplicas    int32                   `json:"updatedReplicas"`
	AvailableReplicas  int32                   `json:"availableReplicas"`
	Labels             map[string]string       `json:"labels"`
	Annotations        map[string]string       `json:"annotations"`
	Selector           map[string]string       `json:"selector"`
	Containers         []ContainerInfo         `json:"containers"`
	ResourceSummary    ResourceSummary         `json:"resourceSummary"`
	Strategy           string                  `json:"strategy"`
	Status             string                  `json:"status"`
	StatusReason       string                  `json:"statusReason"`
	Conditions         []DeploymentConditionVO `json:"conditions"`
	Generation         int64                   `json:"generation"`
	ObservedGeneration int64                   `json:"observedGeneration"`
	CreatedAt          time.Time               `json:"createdAt"`
}

// DeploymentListResponse Deployment 列表分页响应
type DeploymentListResponse struct {
	Total int64              `json:"total"`
	Items []DeploymentListVO `json:"items"`
}

func (s *K8sService) ListDeployments(clusterName string, namespace string, page, pageSize int, keyword string) (*DeploymentListResponse, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	// namespace 为空时查询所有命名空间
	list, err := client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.clientFactory.RemoveClient(cluster.Name)
		return nil, err
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item appsv1.Deployment) []string {
		images := make([]string, 0, len(item.Spec.Template.Spec.Containers))
		for _, container := range item.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}
		return []string{
			item.Name,
			item.Namespace,
			strings.Join(images, ","),
			flattenLabels(item.Labels),
		}
	})

	total := int64(len(filtered))

	// 分页
	page, pageSize = normalizePage(page, pageSize)
	start, end := paginateRange(len(filtered), page, pageSize)
	paged := filtered[start:end]

	// 转换为 VO
	result := make([]DeploymentListVO, 0, len(paged))
	for _, item := range paged {
		containers := buildContainerInfos(item.Spec.Template.Spec.Containers)
		resSummary := aggregateResources(item.Spec.Template.Spec.Containers)
		status, reason := computeDeploymentStatus(&item)

		var replicas int32
		if item.Spec.Replicas != nil {
			replicas = *item.Spec.Replicas
		}

		result = append(result, DeploymentListVO{
			Name:               item.Name,
			Namespace:          item.Namespace,
			Replicas:           replicas,
			ReadyReplicas:      item.Status.ReadyReplicas,
			UpdatedReplicas:    item.Status.UpdatedReplicas,
			AvailableReplicas:  item.Status.AvailableReplicas,
			Labels:             item.Labels,
			Containers:         containers,
			ResourceSummary:    resSummary,
			Strategy:           string(item.Spec.Strategy.Type),
			Status:             status,
			StatusReason:       reason,
			Generation:         item.Generation,
			ObservedGeneration: item.Status.ObservedGeneration,
			CreatedAt:          item.CreationTimestamp.Time,
		})
	}

	return &DeploymentListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) GetDeploymentDetail(clusterName string, namespace, name string) (*DeploymentVO, error) {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	item, err := client.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := buildContainerInfos(item.Spec.Template.Spec.Containers)
	resSummary := aggregateResources(item.Spec.Template.Spec.Containers)
	status, reason := computeDeploymentStatus(item)
	conditions := buildConditionVOs(item.Status.Conditions)

	var selector map[string]string
	if item.Spec.Selector != nil {
		selector = item.Spec.Selector.MatchLabels
	}

	var replicas int32
	if item.Spec.Replicas != nil {
		replicas = *item.Spec.Replicas
	}

	return &DeploymentVO{
		Name:               item.Name,
		Namespace:          item.Namespace,
		Replicas:           replicas,
		ReadyReplicas:      item.Status.ReadyReplicas,
		UpdatedReplicas:    item.Status.UpdatedReplicas,
		AvailableReplicas:  item.Status.AvailableReplicas,
		Labels:             item.Labels,
		Annotations:        item.Annotations,
		Selector:           selector,
		Containers:         containers,
		ResourceSummary:    resSummary,
		Strategy:           string(item.Spec.Strategy.Type),
		Status:             status,
		StatusReason:       reason,
		Conditions:         conditions,
		Generation:         item.Generation,
		ObservedGeneration: item.Status.ObservedGeneration,
		CreatedAt:          item.CreationTimestamp.Time,
	}, nil
}

func (s *K8sService) DeleteDeployment(clusterName string, namespace, name string) error {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return err
	}

	return client.AppsV1().Deployments(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (s *K8sService) CreateDeployment(clusterName string, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.AppsV1().Deployments(namespace).Create(context.Background(), deployment, metav1.CreateOptions{})
}

func (s *K8sService) UpdateDeployment(clusterName string, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
}

// computeDeploymentStatus 从 Deployment 的 conditions 和副本状态计算聚合状态枚举
// 返回 (状态枚举, 原因字符串)
func computeDeploymentStatus(deploy *appsv1.Deployment) (string, string) {
	var replicas int32
	if deploy.Spec.Replicas != nil {
		replicas = *deploy.Spec.Replicas
	}

	var (
		availableCond   *appsv1.DeploymentCondition
		progressingCond *appsv1.DeploymentCondition
	)
	for i := range deploy.Status.Conditions {
		cond := &deploy.Status.Conditions[i]
		switch cond.Type {
		case appsv1.DeploymentAvailable:
			availableCond = cond
		case appsv1.DeploymentProgressing:
			progressingCond = cond
		}
	}

	// 若 Available condition 为 False → Unavailable
	if availableCond != nil && availableCond.Status == corev1.ConditionFalse {
		return "Unavailable", availableCond.Message
	}

	// 若 Progressing condition 为 False（如超时）→ Degraded
	if progressingCond != nil && progressingCond.Status == corev1.ConditionFalse {
		return "Degraded", progressingCond.Message
	}

	// 若副本未全部就绪且正在滚动 → Progressing
	if deploy.Status.AvailableReplicas < replicas || deploy.Status.UpdatedReplicas < replicas {
		reason := ""
		if progressingCond != nil {
			reason = progressingCond.Message
		}
		return "Progressing", reason
	}

	return "Running", ""
}

func buildConditionVOs(conditions []appsv1.DeploymentCondition) []DeploymentConditionVO {
	result := make([]DeploymentConditionVO, 0, len(conditions))
	for _, c := range conditions {
		result = append(result, DeploymentConditionVO{
			Type:               string(c.Type),
			Status:             string(c.Status),
			Reason:             c.Reason,
			Message:            c.Message,
			LastUpdateTime:     c.LastUpdateTime.Time,
			LastTransitionTime: c.LastTransitionTime.Time,
		})
	}
	return result
}
