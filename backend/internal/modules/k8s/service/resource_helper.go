package service

import (
	"fmt"
	"sort"
	"strings"

	"devops-platform/internal/modules/k8s/model"
	queryutil "devops-platform/internal/pkg/query"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// clusterClient 封装集群信息和客户端，减少重复的获取逻辑
type clusterClient struct {
	Cluster *model.Cluster
	Client  *kubernetes.Clientset
}

// getClusterClient 统一获取集群信息和 K8s 客户端
func (s *K8sService) getClusterClient(clusterId uint) (*clusterClient, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, fmt.Errorf("获取集群信息失败(id=%d): %w", clusterId, err)
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
func (s *K8sService) getClusterDynamicClient(clusterId uint) (*model.Cluster, dynamic.Interface, error) {
	if err := s.ensureReady(); err != nil {
		return nil, nil, err
	}

	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, nil, fmt.Errorf("获取集群信息失败(id=%d): %w", clusterId, err)
	}

	dynamicClient, err := s.clientFactory.GetDynamicClient(cluster)
	if err != nil {
		return nil, nil, fmt.Errorf("获取 dynamic 客户端失败(cluster=%s): %w", cluster.Name, err)
	}

	return cluster, dynamicClient, nil
}

// handleClientError 处理客户端错误，必要时移除缓存的客户端
func (s *K8sService) handleClientError(clusterId uint, err error) error {
	if err != nil {
		s.clientFactory.RemoveClient(clusterId)
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
