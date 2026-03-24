package service

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceListVO struct {
	Name string `json:"name"`
}

func (s *K8sService) ListNamespaces(clusterId uint) ([]NamespaceListVO, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	// 1️⃣ 查询集群
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}

	// 2️⃣ 获取 client（自动重建）
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	// 3️⃣ 调用 K8s API
	nsList, err := client.CoreV1().
		Namespaces().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		// ❗调用失败 → 移除 client，下次自动重建
		s.clientFactory.RemoveClient(clusterId)
		return nil, err
	}

	// 4️⃣ 转 VO
	result := make([]NamespaceListVO, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		result = append(result, NamespaceListVO{Name: ns.Name})
	}

	return result, nil
}
