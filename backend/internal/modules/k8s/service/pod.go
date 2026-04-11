package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type PodVO struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Status       string            `json:"status"`
	IP           string            `json:"ip"`
	Node         string            `json:"node"`
	Labels       map[string]string `json:"labels"`
	CreatedAt    time.Time         `json:"createdAt"`
	RestartCount int32             `json:"restartCount"`
	Age          string            `json:"age"`
	Containers   []ContainerInfo   `json:"containers,omitempty"`
}

type PodListVO struct {
	Name         string          `json:"name"`
	Namespace    string          `json:"namespace"`
	Status       string          `json:"status"`
	IP           string          `json:"ip"`
	Node         string          `json:"node"`
	CreatedAt    time.Time       `json:"createdAt"`
	RestartCount int32           `json:"restartCount"`
	Age          string          `json:"age"`
	Containers   []ContainerInfo `json:"containers,omitempty"`
}

// PodListResponse Pod 列表分页响应
type PodListResponse struct {
	Total int64       `json:"total"`
	Items []PodListVO `json:"items"`
}

func (s *K8sService) ListPods(clusterName string, namespace string, page, pageSize int, keyword string) (*PodListResponse, error) {
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

	list, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.clientFactory.RemoveClient(cluster.Name)
		return nil, err
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item corev1.Pod) []string {
		return []string{
			item.Name,
			item.Namespace,
			string(item.Status.Phase),
			item.Spec.NodeName,
			item.Status.PodIP,
			flattenLabels(item.Labels),
		}
	})

	total := int64(len(filtered))

	// 分页
	page, pageSize = normalizePage(page, pageSize)
	start, end := paginateRange(len(filtered), page, pageSize)
	paged := filtered[start:end]

	result := make([]PodListVO, 0, len(paged))
	for _, item := range paged {
		// 计算重启次数
		restartCount := int32(0)
		for _, cs := range item.Status.ContainerStatuses {
			restartCount += cs.RestartCount
		}

		// 计算运行时间
		age := calculateAge(item.CreationTimestamp.Time)

		result = append(result, PodListVO{
			Name:         item.Name,
			Namespace:    item.Namespace,
			Status:       string(item.Status.Phase),
			IP:           item.Status.PodIP,
			Node:         item.Spec.NodeName,
			CreatedAt:    item.CreationTimestamp.Time,
			RestartCount: restartCount,
			Age:          age,
			Containers:   buildContainerInfos(item.Spec.Containers),
		})
	}
	return &PodListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) GetPodDetail(clusterName string, namespace, name string) (*PodVO, error) {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	item, err := client.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 计算重启次数
	restartCount := int32(0)
	for _, cs := range item.Status.ContainerStatuses {
		restartCount += cs.RestartCount
	}

	// 计算运行时间
	age := calculateAge(item.CreationTimestamp.Time)

	return &PodVO{
		Name:         item.Name,
		Namespace:    item.Namespace,
		Status:       string(item.Status.Phase),
		IP:           item.Status.PodIP,
		Node:         item.Spec.NodeName,
		Labels:       item.Labels,
		CreatedAt:    item.CreationTimestamp.Time,
		RestartCount: restartCount,
		Age:          age,
		Containers:   buildContainerInfos(item.Spec.Containers),
	}, nil
}

func (s *K8sService) DeletePod(clusterName string, namespace, name string) error {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return err
	}

	return client.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (s *K8sService) CreatePod(clusterName string, namespace string, pod *corev1.Pod) (*corev1.Pod, error) {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})
}

// GetPodLogs 获取 Pod 日志
func (s *K8sService) GetPodLogs(clusterName string, namespace, name, container string, tailLines int64) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return "", err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return "", err
	}

	// 获取 Pod 对象以确定容器
	pod, err := client.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// 如果未指定容器，使用第一个容器
	if container == "" {
		if len(pod.Spec.Containers) == 0 {
			return "", fmt.Errorf("pod 中没有容器")
		}
		container = pod.Spec.Containers[0].Name
	}

	// 获取日志
	req := client.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLines,
	})

	logs, err := req.Do(context.Background()).Raw()
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

func (s *K8sService) UpdatePod(clusterName string, namespace string, pod *corev1.Pod) (*corev1.Pod, error) {
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	return client.CoreV1().Pods(namespace).Update(context.Background(), pod, metav1.UpdateOptions{})
}

func (s *K8sService) GetPodYAML(clusterName string, namespace, name string) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}
	obj, err := s.GetPodObject(clusterName, namespace, name)
	if err != nil {
		return "", err
	}
	b, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *K8sService) UpdatePodByYAML(clusterName string, namespace, name, rawYAML string) (*corev1.Pod, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	current, err := s.GetPodObject(clusterName, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current pod: %w", err)
	}

	var desired corev1.Pod
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	// 校验不可变字段
	if desired.APIVersion != "" && desired.APIVersion != "v1" {
		return nil, fmt.Errorf("不允许修改 apiVersion")
	}
	if desired.Kind != "" && desired.Kind != "Pod" {
		return nil, fmt.Errorf("不允许修改 kind")
	}
	if desired.Name != "" && desired.Name != name {
		return nil, fmt.Errorf("不允许修改 metadata.name")
	}
	if desired.Namespace != "" && desired.Namespace != namespace {
		return nil, fmt.Errorf("不允许修改 metadata.namespace")
	}

	desired.Namespace = namespace
	desired.Name = name
	desired.ResourceVersion = current.ResourceVersion
	desired.Status = corev1.PodStatus{}
	desired.ManagedFields = nil

	return s.UpdatePod(clusterName, namespace, &desired)
}

// ListPodsByOwner 根据控制器类型和名称获取 Pod 列表
func (s *K8sService) ListPodsByOwner(clusterName string, namespace string, ownerType string, ownerName string) ([]PodListVO, error) {
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

	var selector string
	switch ownerType {
	case "Deployment":
		deploy, err := client.AppsV1().Deployments(namespace).Get(context.Background(), ownerName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		selector = metav1.FormatLabelSelector(deploy.Spec.Selector)
	case "StatefulSet":
		sts, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), ownerName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		selector = metav1.FormatLabelSelector(sts.Spec.Selector)
	case "DaemonSet":
		ds, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), ownerName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		selector = metav1.FormatLabelSelector(ds.Spec.Selector)
	default:
		return nil, fmt.Errorf("不支持的控制器类型: %s", ownerType)
	}

	list, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}

	result := make([]PodListVO, 0, len(list.Items))
	for _, item := range list.Items {
		// 计算重启次数
		restartCount := int32(0)
		for _, cs := range item.Status.ContainerStatuses {
			restartCount += cs.RestartCount
		}

		// 计算运行时间
		age := calculateAge(item.CreationTimestamp.Time)

		result = append(result, PodListVO{
			Name:         item.Name,
			Namespace:    item.Namespace,
			Status:       string(item.Status.Phase),
			IP:           item.Status.PodIP,
			Node:         item.Spec.NodeName,
			CreatedAt:    item.CreationTimestamp.Time,
			RestartCount: restartCount,
			Age:          age,
			Containers:   buildContainerInfos(item.Spec.Containers),
		})
	}
	return result, nil
}

// calculateAge 计算运行时间
func calculateAge(createdAt time.Time) string {
	duration := time.Since(createdAt)
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}

// GetPodEvents 获取 Pod 事件
func (s *K8sService) GetPodEvents(clusterName string, namespace, name string) ([]EventInfo, error) {
	client, err := s.getClient(clusterName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 筛选 involvedObject 为 Pod 且名称匹配
	fieldSelector := fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s,involvedObject.namespace=%s", name, namespace)
	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	// 按时间倒序排序
	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].LastTimestamp.Time.After(events.Items[j].LastTimestamp.Time)
	})

	var result []EventInfo
	for _, e := range events.Items {
		result = append(result, EventInfo{
			Time:    e.LastTimestamp.Time.Format(time.RFC3339),
			Type:    e.Type,
			Reason:  e.Reason,
			Object:  fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name),
			Message: e.Message,
		})
	}
	return result, nil
}
