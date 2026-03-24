package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getClient 获取 K8s 客户端
func (s *K8sService) getClient(clusterID uint) (*kubernetes.Clientset, error) {
	cluster, err := s.clusterService.GetByID(clusterID)
	if err != nil {
		return nil, err
	}
	return s.clientFactory.GetClient(cluster)
}

// WorkloadCounts 工作负载统计
type WorkloadCounts struct {
	Deployment  int `json:"deployment"`
	StatefulSet int `json:"statefulset"`
	DaemonSet   int `json:"daemonset"`
	Job         int `json:"job"`
	CronJob     int `json:"cronjob"`
}

// GetWorkloadCounts 获取工作负载统计
func (s *K8sService) GetWorkloadCounts(clusterID uint) (*WorkloadCounts, error) {
	client, err := s.getClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []string
	counts := &WorkloadCounts{}

	// 并发获取各资源数量
	wg.Add(5)

	go func() {
		defer wg.Done()
		list, err := client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("deployment: %v", err))
			mu.Unlock()
			return
		}
		if list != nil {
			mu.Lock()
			counts.Deployment = len(list.Items)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		list, err := client.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("statefulset: %v", err))
			mu.Unlock()
			return
		}
		if list != nil {
			mu.Lock()
			counts.StatefulSet = len(list.Items)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		list, err := client.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("daemonset: %v", err))
			mu.Unlock()
			return
		}
		if list != nil {
			mu.Lock()
			counts.DaemonSet = len(list.Items)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		list, err := client.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("job: %v", err))
			mu.Unlock()
			return
		}
		if list != nil {
			mu.Lock()
			counts.Job = len(list.Items)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		list, err := client.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("cronjob: %v", err))
			mu.Unlock()
			return
		}
		if list != nil {
			mu.Lock()
			counts.CronJob = len(list.Items)
			mu.Unlock()
		}
	}()

	wg.Wait()

	if len(errs) > 0 {
		return counts, fmt.Errorf("部分统计失败: %s", strings.Join(errs, "; "))
	}

	return counts, nil
}

// NetworkCounts 网络资源统计
type NetworkCounts struct {
	Service int `json:"service"`
	Ingress int `json:"ingress"`
}

// GetNetworkCounts 获取网络资源统计
func (s *K8sService) GetNetworkCounts(clusterID uint) (*NetworkCounts, error) {
	client, err := s.getClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	counts := &NetworkCounts{}

	svcList, err := client.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Service 失败: %w", err)
	}
	counts.Service = len(svcList.Items)

	ingList, err := client.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Ingress 失败: %w", err)
	}
	counts.Ingress = len(ingList.Items)

	return counts, nil
}

// StorageCounts 存储资源统计
type StorageCounts struct {
	PV  int `json:"pv"`
	PVC int `json:"pvc"`
}

// GetStorageCounts 获取存储资源统计
func (s *K8sService) GetStorageCounts(clusterID uint) (*StorageCounts, error) {
	client, err := s.getClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	counts := &StorageCounts{}

	pvList, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PV 失败: %w", err)
	}
	counts.PV = len(pvList.Items)

	pvcList, err := client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PVC 失败: %w", err)
	}
	counts.PVC = len(pvcList.Items)

	return counts, nil
}

// NodeInfo 节点信息
type NodeInfo struct {
	Name            string `json:"name"`
	IP              string `json:"ip"`
	Role            string `json:"role"`
	Status          string `json:"status"`
	CpuUsed         string `json:"cpuUsed"`
	CpuCapacity     string `json:"cpuCapacity"`
	MemoryUsed      string `json:"memoryUsed"`
	MemoryCapacity  string `json:"memoryCapacity"`
	StorageUsed     string `json:"storageUsed"`
	StorageCapacity string `json:"storageCapacity"`
}

type NodeListResponse_Simple struct {
	Total int64      `json:"total"`
	Items []NodeInfo `json:"items"`
}

// GetNodeList 获取节点列表
func (s *K8sService) GetNodeList(clusterID uint, page, pageSize int, name string) (*NodeListResponse_Simple, error) {
	client, err := s.getClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}

	var filtered []corev1.Node
	for _, node := range nodeList.Items {
		if name != "" && !strings.Contains(node.Name, name) {
			continue
		}
		filtered = append(filtered, node)
	}

	total := int64(len(filtered))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start < 0 {
		start = 0
	}
	if start > int(total) {
		start = int(total)
	}
	if end > int(total) {
		end = int(total)
	}

	var items []NodeInfo
	for _, node := range filtered[start:end] {
		// IP
		ip := ""
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				ip = addr.Address
				break
			}
		}
		if ip == "" && len(node.Status.Addresses) > 0 {
			ip = node.Status.Addresses[0].Address
		}

		// Role
		role := "worker"
		for k := range node.Labels {
			if strings.Contains(k, "node-role.kubernetes.io/control-plane") || strings.Contains(k, "node-role.kubernetes.io/master") {
				role = "master"
				break
			}
		}

		// Status
		status := "Unknown"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		// Capacity
		cpuCap := node.Status.Capacity.Cpu().String()
		memCap := node.Status.Capacity.Memory().String()
		storageCap := node.Status.Capacity.StorageEphemeral().String()

		items = append(items, NodeInfo{
			Name:            node.Name,
			IP:              ip,
			Role:            role,
			Status:          status,
			CpuUsed:         "0", // 暂无 metrics
			CpuCapacity:     cpuCap,
			MemoryUsed:      "0", // 暂无 metrics
			MemoryCapacity:  memCap,
			StorageUsed:     "0", // 暂无 metrics
			StorageCapacity: storageCap,
		})
	}

	return &NodeListResponse_Simple{
		Total: total,
		Items: items,
	}, nil
}

// EventInfo 事件信息
type EventInfo struct {
	Time    string `json:"time"`
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Object  string `json:"object"`
	Message string `json:"message"`
}

type EventListResponse struct {
	Total int64       `json:"total"`
	Items []EventInfo `json:"items"`
}

// GetEventList 获取事件列表
func (s *K8sService) GetEventList(clusterID uint, page, pageSize int) (*EventListResponse, error) {
	client, err := s.getClient(clusterID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	eventList, err := client.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取事件列表失败: %w", err)
	}

	// 按时间倒序排序
	sort.Slice(eventList.Items, func(i, j int) bool {
		return eventList.Items[i].LastTimestamp.Time.After(eventList.Items[j].LastTimestamp.Time)
	})

	total := int64(len(eventList.Items))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start < 0 {
		start = 0
	}
	if start > int(total) {
		start = int(total)
	}
	if end > int(total) {
		end = int(total)
	}

	var items []EventInfo
	for _, event := range eventList.Items[start:end] {
		items = append(items, EventInfo{
			Time:    event.LastTimestamp.Time.Format(time.RFC3339),
			Type:    event.Type,
			Reason:  event.Reason,
			Object:  fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
			Message: event.Message,
		})
	}

	return &EventListResponse{
		Total: total,
		Items: items,
	}, nil
}
