package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// NodeService 节点服务
// 方法挂载在 K8sService 上以复用基础设施

// NodeListResponse 节点列表响应
type NodeListResponse struct {
	Total int64          `json:"total"`
	Items []NodeListItem `json:"items"`
}

// NodeListItem 节点列表项
type NodeListItem struct {
	Name              string            `json:"name"`
	Status            string            `json:"status"` // Ready, NotReady, Unknown
	Role              string            `json:"role"`   // master, worker
	IP                string            `json:"ip"`
	ExternalIP        string            `json:"externalIP"` // 外部IP
	KubeletVersion    string            `json:"kubeletVersion"`
	K8sVersion        string            `json:"k8sVersion"` // Kubelet 版本 (别名)
	OsImage           string            `json:"osImage"`    // 操作系统镜像
	KernelVersion     string            `json:"kernelVersion"`
	Labels            map[string]string `json:"labels"`
	Taints            []interface{}     `json:"taints"`
	Unschedulable     bool              `json:"unschedulable"`
	Age               string            `json:"age"`
	CreatedAt         time.Time         `json:"createdAt"`
	CreationTimestamp time.Time         `json:"creationTimestamp"` // 创建时间 (别名)

	// 资源统计
	CpuCapacity    string `json:"cpuCapacity"`
	CpuUsage       string `json:"cpuUsage"` // 使用量 (Core)
	MemoryCapacity string `json:"memoryCapacity"`
	MemoryUsage    string `json:"memoryUsage"` // 使用量 (Gi)
	PodCount       int    `json:"podCount"`
	PodCapacity    int64  `json:"podCapacity"`
}

// NodeLeaseInfo 节点 Lease 信息
type NodeLeaseInfo struct {
	HolderIdentity string `json:"holderIdentity"`
	AcquireTime    string `json:"acquireTime"`
	RenewTime      string `json:"renewTime"`
}

// NodeAllocatedResource 节点已分配资源汇总
type NodeAllocatedResource struct {
	CPURequests              string `json:"cpuRequests"`
	CPURequestsPercentage    string `json:"cpuRequestsPercentage"`
	CPULimits                string `json:"cpuLimits"`
	CPULimitsPercentage      string `json:"cpuLimitsPercentage"`
	MemoryRequests           string `json:"memoryRequests"`
	MemoryRequestsPercentage string `json:"memoryRequestsPercentage"`
	MemoryLimits             string `json:"memoryLimits"`
	MemoryLimitsPercentage   string `json:"memoryLimitsPercentage"`
	EphemeralStorageRequests string `json:"ephemeralStorageRequests"`
	EphemeralStorageLimits   string `json:"ephemeralStorageLimits"`
	Hugepages1GiRequests     string `json:"hugepages1GiRequests"`
	Hugepages1GiLimits       string `json:"hugepages1GiLimits"`
	Hugepages2MiRequests     string `json:"hugepages2MiRequests"`
	Hugepages2MiLimits       string `json:"hugepages2MiLimits"`
}

// NodePodItem 节点上的 Pod 条目
type NodePodItem struct {
	Namespace     string `json:"namespace"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	CPURequest    string `json:"cpuRequest"`
	CPULimit      string `json:"cpuLimit"`
	MemoryRequest string `json:"memoryRequest"`
	MemoryLimit   string `json:"memoryLimit"`
	RestartCount  int32  `json:"restartCount"`
	CreatedAt     string `json:"createdAt"`
	Age           string `json:"age"`
}

// NodePodSummary 节点 Pod 汇总
type NodePodSummary struct {
	Total int           `json:"total"`
	Items []NodePodItem `json:"items"`
}

// NodeResourceQuantity 节点资源数量
type NodeResourceQuantity struct {
	CPU              string `json:"cpu"`
	Memory           string `json:"memory"`
	Pods             string `json:"pods"`
	EphemeralStorage string `json:"ephemeralStorage"`
	Hugepages1Gi     string `json:"hugepages1Gi"`
	Hugepages2Mi     string `json:"hugepages2Mi"`
}

// NodeDetail 节点详情
type NodeDetail struct {
	NodeListItem
	Annotations        map[string]string     `json:"annotations"`
	Lease              *NodeLeaseInfo        `json:"lease"`
	Conditions         []interface{}         `json:"conditions"`
	Addresses          []interface{}         `json:"addresses"`
	Capacity           NodeResourceQuantity  `json:"capacity"`
	Allocatable        NodeResourceQuantity  `json:"allocatable"`
	SystemInfo         interface{}           `json:"systemInfo"`
	PodCIDR            string                `json:"podCIDR"`
	PodCIDRs           []string              `json:"podCIDRs"`
	ProviderID         string                `json:"providerID"`
	Pods               NodePodSummary        `json:"pods"`
	AllocatedResources NodeAllocatedResource `json:"allocatedResources"`
	Images             []interface{}         `json:"images"`
}

// ListNodes 获取节点列表
func (s *K8sService) ListNodes(clusterName string, page, pageSize int, name string, status string, role string) (*NodeListResponse, error) {
	client, err := s.getClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 获取 Metrics Client (允许失败，用于降级)
	var metricsClient *metrics.Clientset
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err == nil {
		metricsClient, _ = s.clientFactory.GetMetricsClient(cluster)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 获取所有节点
	nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %w", err)
	}

	// 2. 内存过滤
	var filtered []corev1.Node
	for _, node := range nodeList.Items {
		// 按名称模糊搜索
		if name != "" && !strings.Contains(node.Name, name) {
			continue
		}

		// 计算角色
		nodeRole := "worker"
		for k := range node.Labels {
			if strings.Contains(k, "node-role.kubernetes.io/control-plane") || strings.Contains(k, "node-role.kubernetes.io/master") {
				nodeRole = "master"
				break
			}
		}
		if role != "" && role != nodeRole {
			continue
		}

		// 计算状态
		nodeStatus := "Unknown"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				if cond.Status == corev1.ConditionTrue {
					nodeStatus = "Ready"
				} else {
					nodeStatus = "NotReady"
				}
				break
			}
		}
		if status != "" && status != nodeStatus {
			continue
		}

		filtered = append(filtered, node)
	}

	// 3. 分页
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

	targetNodes := filtered[start:end]

	// 4. 并发获取 Metrics 和 Pod 数量
	type nodeExtraData struct {
		nodeName string
		cpuUsage string
		memUsage string
		podCount int
	}

	var wg sync.WaitGroup
	dataCh := make(chan nodeExtraData, len(targetNodes))

	for _, node := range targetNodes {
		wg.Add(1)
		go func(n corev1.Node) {
			defer wg.Done()
			data := nodeExtraData{nodeName: n.Name}

			// 获取 Pod 数量
			pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
				FieldSelector: "spec.nodeName=" + n.Name,
			})
			if err == nil {
				data.podCount = len(pods.Items)
			}

			// 获取 Metrics
			if metricsClient != nil {
				metrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, n.Name, metav1.GetOptions{})
				if err == nil {
					// CPU 转 Core
					cpu := metrics.Usage.Cpu().MilliValue()
					data.cpuUsage = fmt.Sprintf("%.2f Core", float64(cpu)/1000)

					// 内存 转 Gi
					mem := metrics.Usage.Memory().Value()
					data.memUsage = fmt.Sprintf("%.2f Gi", float64(mem)/(1024*1024*1024))
				}
			}

			dataCh <- data
		}(node)
	}

	wg.Wait()
	close(dataCh)

	extraMap := make(map[string]nodeExtraData)
	for d := range dataCh {
		extraMap[d.nodeName] = d
	}

	// 5. 组装结果
	var items []NodeListItem
	for _, node := range targetNodes {
		extra := extraMap[node.Name]
		items = append(items, buildNodeListItem(&node, extra.podCount, extra.cpuUsage, extra.memUsage))
	}

	return &NodeListResponse{Total: total, Items: items}, nil
}

type nodeDetailFetchResult struct {
	node  *corev1.Node
	pods  []corev1.Pod
	lease *coordinationv1.Lease
	item  NodeListItem
}

func (s *K8sService) fetchNodeDetailData(clusterName string, name string) (*nodeDetailFetchResult, error) {
	client, err := s.getClient(clusterName)
	if err != nil {
		return nil, err
	}

	var metricsClient *metrics.Clientset
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err == nil {
		metricsClient, _ = s.clientFactory.GetMetricsClient(cluster)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	node, err := client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("节点不存在")
		}
		return nil, fmt.Errorf("获取节点详情失败: %w", err)
	}

	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{FieldSelector: "spec.nodeName=" + name})
	if err != nil {
		return nil, fmt.Errorf("获取节点 Pods 失败: %w", err)
	}

	lease, err := client.CoordinationV1().Leases("kube-node-lease").Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("获取节点 Lease 失败: %w", err)
		}
		lease = nil
	}

	cpuUsage := ""
	memoryUsage := ""
	if metricsClient != nil {
		nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			cpu := nodeMetrics.Usage.Cpu().MilliValue()
			mem := nodeMetrics.Usage.Memory().Value()
			cpuUsage = fmt.Sprintf("%.2f Core", float64(cpu)/1000)
			memoryUsage = fmt.Sprintf("%.2f Gi", float64(mem)/(1024*1024*1024))
		}
	}

	return &nodeDetailFetchResult{
		node:  node,
		pods:  pods.Items,
		lease: lease,
		item:  buildNodeListItem(node, len(pods.Items), cpuUsage, memoryUsage),
	}, nil
}

func buildNodeListItem(node *corev1.Node, podCount int, cpuUsage string, memUsage string) NodeListItem {
	if cpuUsage == "" {
		cpuUsage = "-"
	}
	if memUsage == "" {
		memUsage = "-"
	}

	nodeRole := "worker"
	for k := range node.Labels {
		if strings.Contains(k, "node-role.kubernetes.io/control-plane") || strings.Contains(k, "node-role.kubernetes.io/master") {
			nodeRole = "master"
			break
		}
	}

	nodeStatus := "Unknown"
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				nodeStatus = "Ready"
			} else {
				nodeStatus = "NotReady"
			}
			break
		}
	}

	ip := ""
	externalIP := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			ip = addr.Address
		} else if addr.Type == corev1.NodeExternalIP {
			externalIP = addr.Address
		}
	}

	cpuCap := node.Status.Capacity.Cpu().MilliValue()
	memCap := node.Status.Capacity.Memory().Value()
	podCap := node.Status.Capacity.Pods().Value()

	return NodeListItem{
		Name:              node.Name,
		Status:            nodeStatus,
		Role:              nodeRole,
		IP:                ip,
		ExternalIP:        externalIP,
		KubeletVersion:    node.Status.NodeInfo.KubeletVersion,
		K8sVersion:        node.Status.NodeInfo.KubeletVersion,
		OsImage:           node.Status.NodeInfo.OSImage,
		KernelVersion:     node.Status.NodeInfo.KernelVersion,
		Labels:            node.Labels,
		Taints:            convertTaints(node.Spec.Taints),
		Unschedulable:     node.Spec.Unschedulable,
		CreatedAt:         node.CreationTimestamp.Time,
		CreationTimestamp: node.CreationTimestamp.Time,
		Age:               formatAge(time.Since(node.CreationTimestamp.Time)),
		CpuCapacity:       fmt.Sprintf("%.2f Core", float64(cpuCap)/1000),
		CpuUsage:          cpuUsage,
		MemoryCapacity:    fmt.Sprintf("%.2f Gi", float64(memCap)/(1024*1024*1024)),
		MemoryUsage:       memUsage,
		PodCount:          podCount,
		PodCapacity:       podCap,
	}
}

// GetNodeDetail 获取节点详情
func (s *K8sService) GetNodeDetail(clusterName string, name string) (*NodeDetail, error) {
	data, err := s.fetchNodeDetailData(clusterName, name)
	if err != nil {
		return nil, err
	}

	return buildNodeDetail(data.node, data.pods, data.lease, data.item), nil
}

// CordonNode 设置/取消调度
func (s *K8sService) CordonNode(clusterName string, name string, cordon bool) error {
	client, err := s.getClient(clusterName)
	if err != nil {
		return err
	}

	// 使用 Patch 更新
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  "/spec/unschedulable",
		Value: cordon,
	}}
	payloadBytes, _ := json.Marshal(payload)

	_, err = client.CoreV1().Nodes().Patch(context.Background(), name, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	return err
}

// DrainNode 驱逐节点
type DrainOptions struct {
	GracePeriodSeconds int  `json:"gracePeriodSeconds"`
	Force              bool `json:"force"`
	IgnoreDaemonSets   bool `json:"ignoreDaemonSets"`
	DeleteLocalData    bool `json:"deleteLocalData"`
}

func (s *K8sService) DrainNode(clusterName string, name string, opts DrainOptions) error {
	client, err := s.getClient(clusterName)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// 1. Cordon 节点
	if err := s.CordonNode(clusterName, name, true); err != nil {
		return fmt.Errorf("设置不可调度失败: %w", err)
	}

	// 2. 获取节点上的 Pod
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + name,
	})
	if err != nil {
		return fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	// 3. 驱逐 Pod
	var errs []string
	for _, pod := range pods.Items {
		// 忽略 DaemonSet (如果配置)
		if opts.IgnoreDaemonSets {
			isDaemonSet := false
			for _, ref := range pod.OwnerReferences {
				if ref.Kind == "DaemonSet" {
					isDaemonSet = true
					break
				}
			}
			if isDaemonSet {
				continue
			}
		}

		// 忽略 Mirror Pod
		if _, ok := pod.Annotations[corev1.MirrorPodAnnotationKey]; ok {
			continue
		}

		// 构建驱逐请求
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			DeleteOptions: &metav1.DeleteOptions{
				GracePeriodSeconds: int64Ptr(opts.GracePeriodSeconds),
			},
		}

		err := client.CoreV1().Pods(pod.Namespace).EvictV1(ctx, eviction)
		if err != nil {
			errs = append(errs, fmt.Sprintf("驱逐 Pod %s/%s 失败: %v", pod.Namespace, pod.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("驱逐部分 Pod 失败: %s", strings.Join(errs, "; "))
	}
	return nil
}

// UpdateNodeLabels 更新标签
func (s *K8sService) UpdateNodeLabels(clusterName string, name string, labels map[string]string) error {
	client, err := s.getClient(clusterName)
	if err != nil {
		return err
	}

	// 获取当前节点以进行 Merge Patch
	node, err := client.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	newNode := node.DeepCopy()
	newNode.Labels = labels

	patchBytes, err := createMergePatch(node, newNode)
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Nodes().Patch(context.Background(), name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	return err
}

// UpdateNodeTaints 更新污点
func (s *K8sService) UpdateNodeTaints(clusterName string, name string, taints []corev1.Taint) error {
	client, err := s.getClient(clusterName)
	if err != nil {
		return err
	}

	node, err := client.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	newNode := node.DeepCopy()
	newNode.Spec.Taints = taints

	patchBytes, err := createMergePatch(node, newNode)
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Nodes().Patch(context.Background(), name, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	return err
}

// GetNodeEvents 获取节点事件
func (s *K8sService) GetNodeEvents(clusterName string, nodeName string) ([]EventInfo, error) {
	client, err := s.getClient(clusterName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 筛选 involvedObject 为 Node 且名称匹配
	fieldSelector := fmt.Sprintf("involvedObject.kind=Node,involvedObject.name=%s", nodeName)
	events, err := client.CoreV1().Events("").List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	// 排序
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

// 辅助函数

func formatAge(d time.Duration) string {
	if d.Hours() > 24 {
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	}
	return d.Round(time.Minute).String()
}

func buildNodeDetail(node *corev1.Node, pods []corev1.Pod, lease *coordinationv1.Lease, base NodeListItem) *NodeDetail {
	items, allocated := buildNodePodsAndAllocatedResources(pods, node.Status.Allocatable)

	return &NodeDetail{
		NodeListItem:       base,
		Annotations:        node.Annotations,
		Lease:              convertNodeLease(lease),
		Conditions:         convertConditions(node.Status.Conditions),
		Addresses:          convertAddresses(node.Status.Addresses),
		Capacity:           convertNodeResourceList(node.Status.Capacity),
		Allocatable:        convertNodeResourceList(node.Status.Allocatable),
		SystemInfo:         node.Status.NodeInfo,
		PodCIDR:            fallbackDash(node.Spec.PodCIDR),
		PodCIDRs:           node.Spec.PodCIDRs,
		ProviderID:         fallbackDash(node.Spec.ProviderID),
		Pods:               NodePodSummary{Total: len(items), Items: items},
		AllocatedResources: allocated,
		Images:             convertImages(node.Status.Images),
	}
}

func buildNodePodsAndAllocatedResources(pods []corev1.Pod, allocatable corev1.ResourceList) ([]NodePodItem, NodeAllocatedResource) {
	items := make([]NodePodItem, 0, len(pods))
	var cpuReq, cpuLim, memReq, memLim, ephemeralReq, ephemeralLim resource.Quantity
	var huge1Req, huge1Lim, huge2Req, huge2Lim resource.Quantity

	for _, pod := range pods {
		if isTerminatedPod(pod) {
			continue
		}

		podCPUReq, podCPULim, podMemReq, podMemLim, podEphemeralReq, podEphemeralLim, podHuge1Req, podHuge1Lim, podHuge2Req, podHuge2Lim := sumPodResources(pod.Spec.Containers)
		item := NodePodItem{
			Namespace:     pod.Namespace,
			Name:          pod.Name,
			Status:        string(pod.Status.Phase),
			CPURequest:    quantityStringOrDash(podCPUReq),
			CPULimit:      quantityStringOrDash(podCPULim),
			MemoryRequest: quantityStringOrDash(podMemReq),
			MemoryLimit:   quantityStringOrDash(podMemLim),
			RestartCount:  sumRestartCount(pod.Status.ContainerStatuses),
			CreatedAt:     pod.CreationTimestamp.Time.Format(time.RFC3339),
			Age:           formatAge(time.Since(pod.CreationTimestamp.Time)),
		}

		cpuReq.Add(podCPUReq)
		cpuLim.Add(podCPULim)
		memReq.Add(podMemReq)
		memLim.Add(podMemLim)
		ephemeralReq.Add(podEphemeralReq)
		ephemeralLim.Add(podEphemeralLim)
		huge1Req.Add(podHuge1Req)
		huge1Lim.Add(podHuge1Lim)
		huge2Req.Add(podHuge2Req)
		huge2Lim.Add(podHuge2Lim)
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace == items[j].Namespace {
			return items[i].Name < items[j].Name
		}
		return items[i].Namespace < items[j].Namespace
	})

	return items, NodeAllocatedResource{
		CPURequests:              quantityStringOrDash(cpuReq),
		CPURequestsPercentage:    percentageString(cpuReq, allocatable.Cpu()),
		CPULimits:                quantityStringOrDash(cpuLim),
		CPULimitsPercentage:      percentageString(cpuLim, allocatable.Cpu()),
		MemoryRequests:           quantityStringOrDash(memReq),
		MemoryRequestsPercentage: percentageString(memReq, allocatable.Memory()),
		MemoryLimits:             quantityStringOrDash(memLim),
		MemoryLimitsPercentage:   percentageString(memLim, allocatable.Memory()),
		EphemeralStorageRequests: quantityStringOrDash(ephemeralReq),
		EphemeralStorageLimits:   quantityStringOrDash(ephemeralLim),
		Hugepages1GiRequests:     quantityStringOrDash(huge1Req),
		Hugepages1GiLimits:       quantityStringOrDash(huge1Lim),
		Hugepages2MiRequests:     quantityStringOrDash(huge2Req),
		Hugepages2MiLimits:       quantityStringOrDash(huge2Lim),
	}
}

func fallbackDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

func quantityStringOrDash(q resource.Quantity) string {
	if q.IsZero() {
		return "-"
	}
	return q.String()
}

func percentageString(used resource.Quantity, total *resource.Quantity) string {
	if total == nil || total.IsZero() || used.IsZero() {
		return "-"
	}
	return fmt.Sprintf("%.2f%%", float64(used.MilliValue())/float64(total.MilliValue())*100)
}

func convertNodeLease(lease *coordinationv1.Lease) *NodeLeaseInfo {
	if lease == nil {
		return nil
	}

	info := &NodeLeaseInfo{}
	if lease.Spec.HolderIdentity != nil {
		info.HolderIdentity = *lease.Spec.HolderIdentity
	}
	info.AcquireTime = microTimeStringOrDash(lease.Spec.AcquireTime)
	info.RenewTime = microTimeStringOrDash(lease.Spec.RenewTime)
	return info
}

func microTimeStringOrDash(v *metav1.MicroTime) string {
	if v == nil || v.Time.IsZero() {
		return "-"
	}
	return v.Time.Format(time.RFC3339)
}

func convertNodeResourceList(resources corev1.ResourceList) NodeResourceQuantity {
	return NodeResourceQuantity{
		CPU:              resourceValueOrDash(resources.Cpu()),
		Memory:           resourceValueOrDash(resources.Memory()),
		Pods:             resourceValueOrDash(resources.Pods()),
		EphemeralStorage: resourceValueOrDash(resources.StorageEphemeral()),
		Hugepages1Gi:     resourceValueOrDash(resources.Name(corev1.ResourceHugePagesPrefix+"1Gi", resource.BinarySI)),
		Hugepages2Mi:     resourceValueOrDash(resources.Name(corev1.ResourceHugePagesPrefix+"2Mi", resource.BinarySI)),
	}
}

func resourceValueOrDash(q *resource.Quantity) string {
	if q == nil || q.IsZero() {
		return "-"
	}
	return q.String()
}

func isTerminatedPod(pod corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed
}

func sumRestartCount(statuses []corev1.ContainerStatus) int32 {
	var total int32
	for _, status := range statuses {
		total += status.RestartCount
	}
	return total
}

func sumPodResources(containers []corev1.Container) (resource.Quantity, resource.Quantity, resource.Quantity, resource.Quantity, resource.Quantity, resource.Quantity, resource.Quantity, resource.Quantity, resource.Quantity, resource.Quantity) {
	var cpuReq, cpuLim, memReq, memLim resource.Quantity
	var ephemeralReq, ephemeralLim, huge1Req, huge1Lim, huge2Req, huge2Lim resource.Quantity
	for _, container := range containers {
		if q, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
			cpuReq.Add(q)
		}
		if q, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
			cpuLim.Add(q)
		}
		if q, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
			memReq.Add(q)
		}
		if q, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
			memLim.Add(q)
		}
		if q, ok := container.Resources.Requests[corev1.ResourceEphemeralStorage]; ok {
			ephemeralReq.Add(q)
		}
		if q, ok := container.Resources.Limits[corev1.ResourceEphemeralStorage]; ok {
			ephemeralLim.Add(q)
		}
		if q, ok := container.Resources.Requests[corev1.ResourceName(corev1.ResourceHugePagesPrefix+"1Gi")]; ok {
			huge1Req.Add(q)
		}
		if q, ok := container.Resources.Limits[corev1.ResourceName(corev1.ResourceHugePagesPrefix+"1Gi")]; ok {
			huge1Lim.Add(q)
		}
		if q, ok := container.Resources.Requests[corev1.ResourceName(corev1.ResourceHugePagesPrefix+"2Mi")]; ok {
			huge2Req.Add(q)
		}
		if q, ok := container.Resources.Limits[corev1.ResourceName(corev1.ResourceHugePagesPrefix+"2Mi")]; ok {
			huge2Lim.Add(q)
		}
	}
	return cpuReq, cpuLim, memReq, memLim, ephemeralReq, ephemeralLim, huge1Req, huge1Lim, huge2Req, huge2Lim
}

func sumPodComputeResources(containers []corev1.Container) (resource.Quantity, resource.Quantity, resource.Quantity, resource.Quantity) {
	cpuReq, cpuLim, memReq, memLim, _, _, _, _, _, _ := sumPodResources(containers)
	return cpuReq, cpuLim, memReq, memLim
}

type patchStringValue struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func int64Ptr(i int) *int64 {
	v := int64(i)
	return &v
}

func createMergePatch(original, modified interface{}) ([]byte, error) {
	_, err := json.Marshal(original)
	if err != nil {
		return nil, err
	}
	modBytes, err := json.Marshal(modified)
	if err != nil {
		return nil, err
	}

	// 这里简化处理，实际可以使用 strategicpatch 库，但 json merge patch 对 corev1 对象通常足够
	// 为了更严谨，对于 Node 对象建议使用 Strategic Merge Patch
	// 但此处为了不引入过多额外依赖，我们简单模拟 merge patch 差异
	// 注意：MergePatchType 只能做覆盖，无法做列表的精确删减（如 Taints），
	// 因此对于 Taints，上面的 Patch 调用是全量替换，这符合 JSON Merge Patch 语义
	return modBytes, nil
}

// 修正 UpdateNodeLabels 和 UpdateNodeTaints 为使用 Update 方法
// 覆盖原方法实现

func (s *K8sService) UpdateNodeLabels_Revised(clusterName string, name string, labels map[string]string) error {
	client, err := s.getClient(clusterName)
	if err != nil {
		return err
	}

	retryErr := retryOnConflict(func() error {
		node, err := client.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		node.Labels = labels
		_, err = client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
		return err
	})
	return retryErr
}

func (s *K8sService) UpdateNodeTaints_Revised(clusterName string, name string, taints []corev1.Taint) error {
	client, err := s.getClient(clusterName)
	if err != nil {
		return err
	}

	retryErr := retryOnConflict(func() error {
		node, err := client.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		node.Spec.Taints = taints
		_, err = client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
		return err
	})
	return retryErr
}

func retryOnConflict(fn func() error) error {
	for i := 0; i < 3; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "conflict") {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("重试次数超限")
}

func convertConditions(conds []corev1.NodeCondition) []interface{} {
	result := make([]interface{}, len(conds))
	for i, v := range conds {
		result[i] = v
	}
	return result
}

func convertTaints(taints []corev1.Taint) []interface{} {
	result := make([]interface{}, len(taints))
	for i, t := range taints {
		result[i] = t
	}
	return result
}

func convertAddresses(addrs []corev1.NodeAddress) []interface{} {
	result := make([]interface{}, len(addrs))
	for i, v := range addrs {
		result[i] = v
	}
	return result
}

func convertImages(imgs []corev1.ContainerImage) []interface{} {
	result := make([]interface{}, len(imgs))
	for i, v := range imgs {
		result[i] = v
	}
	return result
}
