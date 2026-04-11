package k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	"devops-platform/internal/modules/k8s/model"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type cachedClient struct {
	client        *kubernetes.Clientset
	dynamicClient dynamic.Interface
	metricsClient *metrics.Clientset
	createdAt     time.Time
}

type ClientFactory struct {
	mu      sync.RWMutex
	clients map[string]*cachedClient

	// 可配置项
	maxAge time.Duration
}

func NewClientFactory() *ClientFactory {
	return &ClientFactory{
		clients: make(map[string]*cachedClient),
		maxAge:  30 * time.Minute, // Client 最长存活时间
	}
}

// GetClient 获取可用 client（自动重建）
func (f *ClientFactory) GetClient(cluster *model.Cluster) (*kubernetes.Clientset, error) {
	// 1️⃣ 读缓存
	f.mu.RLock()
	cc, ok := f.clients[cluster.Name]
	f.mu.RUnlock()

	if ok && !f.isExpired(cc) && f.isHealthy(cc.client) {
		return cc.client, nil
	}

	// 2️⃣ 创建新 client
	clientset, dynamicClient, metricsClient, err := f.buildClients(cluster)
	if err != nil {
		return nil, err
	}

	// 3️⃣ 写缓存（双检）
	f.mu.Lock()
	defer f.mu.Unlock()

	f.clients[cluster.Name] = &cachedClient{
		client:        clientset,
		dynamicClient: dynamicClient,
		metricsClient: metricsClient,
		createdAt:     time.Now(),
	}

	return clientset, nil
}

// GetDynamicClient 获取可用 dynamic client（自动重建）
func (f *ClientFactory) GetDynamicClient(cluster *model.Cluster) (dynamic.Interface, error) {
	f.mu.RLock()
	cc, ok := f.clients[cluster.Name]
	f.mu.RUnlock()

	if ok && !f.isExpired(cc) && cc.dynamicClient != nil && f.isHealthy(cc.client) {
		return cc.dynamicClient, nil
	}

	clientset, dynamicClient, metricsClient, err := f.buildClients(cluster)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.clients[cluster.Name] = &cachedClient{
		client:        clientset,
		dynamicClient: dynamicClient,
		metricsClient: metricsClient,
		createdAt:     time.Now(),
	}

	return dynamicClient, nil
}

// GetMetricsClient 获取可用 metrics client（自动重建）
func (f *ClientFactory) GetMetricsClient(cluster *model.Cluster) (*metrics.Clientset, error) {
	if cluster == nil {
		return nil, fmt.Errorf("cluster cannot be nil")
	}

	f.mu.RLock()
	cc, ok := f.clients[cluster.Name]
	f.mu.RUnlock()

	if ok && !f.isExpired(cc) && cc.metricsClient != nil && f.isHealthy(cc.client) {
		return cc.metricsClient, nil
	}

	clientset, dynamicClient, metricsClient, err := f.buildClients(cluster)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.clients[cluster.Name] = &cachedClient{
		client:        clientset,
		dynamicClient: dynamicClient,
		metricsClient: metricsClient,
		createdAt:     time.Now(),
	}

	return metricsClient, nil
}

func (f *ClientFactory) buildClients(cluster *model.Cluster) (*kubernetes.Clientset, dynamic.Interface, *metrics.Clientset, error) {
	restConfig, err := BuildRestConfigFromCluster(cluster)
	if err != nil {
		return nil, nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建 kubernetes 客户端失败: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建 dynamic 客户端失败: %w", err)
	}
	metricsClient, err := metrics.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建 metrics 客户端失败: %w", err)
	}

	return clientset, dynamicClient, metricsClient, nil
}

// isExpired 判断是否过期
func (f *ClientFactory) isExpired(cc *cachedClient) bool {
	return time.Since(cc.createdAt) > f.maxAge
}

// isHealthy 健康检查（轻量）
func (f *ClientFactory) isHealthy(client *kubernetes.Clientset) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 使用带超时的上下文，但 ServerVersion 本身不支持 context
	// 通过超时控制整体健康检查时间
	done := make(chan error, 1)
	go func() {
		_, err := client.Discovery().ServerVersion()
		done <- err
	}()

	select {
	case err := <-done:
		return err == nil
	case <-ctx.Done():
		return false
	}
}

// RemoveClient 主动移除（集群配置变更时）
func (f *ClientFactory) RemoveClient(clusterName string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.clients, clusterName)
}
