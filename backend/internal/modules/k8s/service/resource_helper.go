package service

import (
	"fmt"
	"sort"
	"strings"

	"devops-platform/internal/modules/k8s/model"
	queryutil "devops-platform/internal/pkg/query"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// clusterClient 封装集群信息和客户端，减少重复的获取逻辑
type clusterClient struct {
	Cluster *model.Cluster
	Client  *kubernetes.Clientset
}

// getClusterClient 统一获取集群信息和 K8s 客户端
func (s *K8sService) getClusterClient(clusterName string) (*clusterClient, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群信息失败(name=%s): %w", clusterName, err)
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, fmt.Errorf("获取集群客户端失败(cluster=%s): %w", cluster.Name, err)
	}

	return &clusterClient{
		Cluster: cluster,
		Client:  client,
	}, nil
}

// getClusterDynamicClient 获取集群的 dynamic client
func (s *K8sService) getClusterDynamicClient(clusterName string) (*model.Cluster, dynamic.Interface, error) {
	if err := s.ensureReady(); err != nil {
		return nil, nil, err
	}

	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, nil, fmt.Errorf("获取集群信息失败(name=%s): %w", clusterName, err)
	}

	dynamicClient, err := s.clientFactory.GetDynamicClient(cluster)
	if err != nil {
		return nil, nil, fmt.Errorf("获取 dynamic 客户端失败(cluster=%s): %w", cluster.Name, err)
	}

	return cluster, dynamicClient, nil
}

// handleClientError 处理客户端错误，必要时移除缓存的客户端
func (s *K8sService) handleClientError(clusterName string, err error) error {
	if err != nil {
		s.clientFactory.RemoveClient(clusterName)
	}
	return err
}

func filterByKeywordFields[T any](items []T, keyword string, fieldsFunc func(T) []string) []T {
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		if queryutil.MatchKeywordAny(keyword, fieldsFunc(item)...) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func flattenLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(labels))
	for _, key := range keys {
		parts = append(parts, key+"="+labels[key])
	}
	return strings.Join(parts, ",")
}

// paginateItems 通用分页处理
func paginateItems[T any](items []T, page, pageSize int) ([]T, int64) {
	total := int64(len(items))
	page, pageSize = normalizePage(page, pageSize)
	start, end := paginateRange(len(items), page, pageSize)
	return items[start:end], total
}

// buildContainerInfos 构建容器信息列表
func buildContainerInfos(containers []corev1.Container) []ContainerInfo {
	result := make([]ContainerInfo, 0, len(containers))
	for _, c := range containers {
		result = append(result, ContainerInfo{
			Name:      c.Name,
			Image:     c.Image,
			Resources: formatResources(c.Resources),
		})
	}
	return result
}

// formatResources 格式化资源请求和限制
func formatResources(req corev1.ResourceRequirements) string {
	var parts []string
	if !req.Requests.Cpu().IsZero() {
		parts = append(parts, fmt.Sprintf("Req CPU: %s", req.Requests.Cpu().String()))
	}
	if !req.Requests.Memory().IsZero() {
		parts = append(parts, fmt.Sprintf("Req Mem: %s", req.Requests.Memory().String()))
	}
	if !req.Limits.Cpu().IsZero() {
		parts = append(parts, fmt.Sprintf("Lim CPU: %s", req.Limits.Cpu().String()))
	}
	if !req.Limits.Memory().IsZero() {
		parts = append(parts, fmt.Sprintf("Lim Mem: %s", req.Limits.Memory().String()))
	}
	if len(parts) == 0 {
		return "None"
	}
	return strings.Join(parts, ", ")
}

// aggregateResources 聚合所有容器的 CPU/Memory requests 和 limits
func aggregateResources(containers []corev1.Container) ResourceSummary {
	cpuReq, cpuLim, memReq, memLim := resource.Quantity{}, resource.Quantity{}, resource.Quantity{}, resource.Quantity{}
	hasCPUReq, hasCPULim, hasMemReq, hasMemLim := false, false, false, false

	for _, c := range containers {
		if v, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
			cpuReq.Add(v)
			hasCPUReq = true
		}
		if v, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
			cpuLim.Add(v)
			hasCPULim = true
		}
		if v, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
			memReq.Add(v)
			hasMemReq = true
		}
		if v, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
			memLim.Add(v)
			hasMemLim = true
		}
	}

	summary := ResourceSummary{}
	if hasCPUReq {
		summary.CPURequest = cpuReq.String()
	}
	if hasCPULim {
		summary.CPULimit = cpuLim.String()
	}
	if hasMemReq {
		summary.MemoryRequest = memReq.String()
	}
	if hasMemLim {
		summary.MemoryLimit = memLim.String()
	}
	return summary
}
