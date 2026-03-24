package service

import (
	"errors"

	"devops-platform/internal/pkg/k8s"
)

type K8sService struct {
	clusterService *ClusterService
	clientFactory  *k8s.ClientFactory
}

func NewK8sService(
	clusterService *ClusterService,
	clientFactory *k8s.ClientFactory,
) *K8sService {
	return &K8sService{
		clusterService: clusterService,
		clientFactory:  clientFactory,
	}
}

func (s *K8sService) ensureReady() error {
	if s == nil {
		return errors.New("k8s service not initialized")
	}
	if s.clusterService == nil {
		return errors.New("cluster service not initialized")
	}
	if s.clientFactory == nil {
		return errors.New("k8s client factory not initialized")
	}
	return nil
}

// normalizePage 规范化分页参数
func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return page, pageSize
}

// paginateSlice 对切片进行分页，返回 (起始索引, 结束索引)
func paginateRange(total, page, pageSize int) (int, int) {
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end
}
