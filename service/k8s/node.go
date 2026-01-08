package k8s

import (
"context"

corev1 "k8s.io/api/core/v1"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeService 节点服务
type NodeService struct {
	clusterService *ClusterService
}

// ListNodes 获取节点列表
func (s *NodeService) ListNodes(clusterID uint) ([]corev1.Node, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodes.Items, nil
}

// GetNode 获取节点详情
func (s *NodeService) GetNode(clusterID uint, name string) (*corev1.Node, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	node, err := clientset.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return node, nil
}

// EventService 事件服务
type EventService struct {
	clusterService *ClusterService
}

// ListEvents 获取事件列表
func (s *EventService) ListEvents(clusterID uint, namespace string) ([]corev1.Event, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	events, err := clientset.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return events.Items, nil
}

// GetEventsByObject 获取指定资源的事件
func (s *EventService) GetEventsByObject(clusterID uint, namespace, kind, name string) ([]corev1.Event, error) {
	if s.clusterService == nil {
		s.clusterService = &ClusterService{}
	}

	clientset, err := s.clusterService.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	fieldSelector := "involvedObject.kind=" + kind + ",involvedObject.name=" + name
	events, err := clientset.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	return events.Items, nil
}
