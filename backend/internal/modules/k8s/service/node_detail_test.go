package service

import (
	"testing"
	"time"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildNodeDetailFromNode_IncludesLeasePodsAndAllocatedResources(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "vm-0-15-rockylinux",
			CreationTimestamp: metav1.NewTime(time.Date(2026, 3, 25, 12, 28, 35, 0, time.UTC)),
			Labels: map[string]string{
				"node-role.kubernetes.io/control-plane": "true",
				"kubernetes.io/hostname":                "vm-0-15-rockylinux",
			},
			Annotations: map[string]string{
				"k3s.io/internal-ip": "10.0.0.15",
			},
		},
		Spec: corev1.NodeSpec{
			PodCIDR:    "10.42.0.0/24",
			PodCIDRs:   []string{"10.42.0.0/24"},
			ProviderID: "k3s://vm-0-15-rockylinux",
		},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("2"),
				corev1.ResourceMemory:           resource.MustParse("3743404Ki"),
				corev1.ResourcePods:             resource.MustParse("110"),
				corev1.ResourceEphemeralStorage: resource.MustParse("72128976Ki"),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("2"),
				corev1.ResourceMemory:           resource.MustParse("3743404Ki"),
				corev1.ResourcePods:             resource.MustParse("110"),
				corev1.ResourceEphemeralStorage: resource.MustParse("70167067798"),
			},
			NodeInfo: corev1.NodeSystemInfo{
				OSImage:                 "Rocky Linux 9.4 (Blue Onyx)",
				KernelVersion:           "5.14.0-570.58.1.el9_6.x86_64",
				KubeletVersion:          "v1.34.5+k3s1",
				ContainerRuntimeVersion: "containerd://2.1.5-k3s1",
			},
		},
	}

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mysql", Namespace: "default", CreationTimestamp: metav1.NewTime(time.Now().Add(-2 * time.Hour))},
			Spec: corev1.PodSpec{NodeName: node.Name, Containers: []corev1.Container{{
				Name: "mysql",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m"), corev1.ResourceMemory: resource.MustParse("512Mi")},
					Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1"), corev1.ResourceMemory: resource.MustParse("1Gi")},
				},
			}}},
			Status: corev1.PodStatus{Phase: corev1.PodRunning, ContainerStatuses: []corev1.ContainerStatus{{Name: "mysql", RestartCount: 2}}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "finished-job", Namespace: "default"},
			Spec:       corev1.PodSpec{NodeName: node.Name},
			Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
		},
	}

	lease := &coordinationv1.Lease{
		Spec: coordinationv1.LeaseSpec{
			HolderIdentity: strPtr("vm-0-15-rockylinux"),
			RenewTime:      &metav1.MicroTime{Time: time.Date(2026, 4, 11, 11, 55, 35, 0, time.UTC)},
		},
	}

	detail := buildNodeDetail(node, pods, lease, NodeListItem{Name: node.Name, CpuUsage: "0.90 Core", MemoryUsage: "0.89 Gi", PodCount: 7, PodCapacity: 110})

	if detail.Lease == nil || detail.Lease.HolderIdentity != "vm-0-15-rockylinux" {
		t.Fatalf("expected lease holder identity populated, got %#v", detail.Lease)
	}
	if detail.Pods.Total != 1 {
		t.Fatalf("expected succeeded pod filtered out, got total=%d", detail.Pods.Total)
	}
	if detail.AllocatedResources.CPURequests != "500m" || detail.AllocatedResources.MemoryLimits != "1Gi" {
		t.Fatalf("unexpected allocated resources: %#v", detail.AllocatedResources)
	}
	if detail.PodCIDR != "10.42.0.0/24" || detail.ProviderID != "k3s://vm-0-15-rockylinux" {
		t.Fatalf("expected pod network fields populated, got podCIDR=%s providerID=%s", detail.PodCIDR, detail.ProviderID)
	}
}

func TestBuildNodeDetail_UsesDashForMissingOptionalFields(t *testing.T) {
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-with-missing-fields"}}
	detail := buildNodeDetail(node, nil, nil, NodeListItem{Name: node.Name})

	if detail.Lease != nil {
		t.Fatalf("expected nil lease when node lease missing, got %#v", detail.Lease)
	}
	if detail.ProviderID != "-" {
		t.Fatalf("expected providerID fallback '-', got %q", detail.ProviderID)
	}
	if detail.Pods.Total != 0 || len(detail.Pods.Items) != 0 {
		t.Fatalf("expected empty pod summary, got %#v", detail.Pods)
	}
}

func TestBuildNodeListItem_MapsCoreFields(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-a",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-48 * time.Hour)),
			Labels: map[string]string{
				"node-role.kubernetes.io/control-plane": "true",
				"topology.kubernetes.io/zone":          "cn-shanghai-a",
			},
		},
		Spec: corev1.NodeSpec{
			Unschedulable: true,
			Taints: []corev1.Taint{{
				Key:    "dedicated",
				Value:  "control-plane",
				Effect: corev1.TaintEffectNoSchedule,
			}},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
			Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "10.0.0.10"}, {Type: corev1.NodeExternalIP, Address: "1.2.3.4"}},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("8Gi"),
				corev1.ResourcePods:   resource.MustParse("120"),
			},
			NodeInfo: corev1.NodeSystemInfo{
				KubeletVersion: "v1.30.0",
				OSImage:        "Rocky Linux",
				KernelVersion:  "6.1.0",
			},
		},
	}

	item := buildNodeListItem(node, 12, "", "")

	if item.Role != "master" || item.Status != "Ready" {
		t.Fatalf("expected role/status mapped, got %#v", item)
	}
	if item.IP != "10.0.0.10" || item.ExternalIP != "1.2.3.4" {
		t.Fatalf("expected IPs mapped, got %#v", item)
	}
	if item.CpuCapacity != "4.00 Core" || item.MemoryCapacity != "8.00 Gi" || item.PodCapacity != 120 {
		t.Fatalf("expected capacities mapped, got %#v", item)
	}
	if item.PodCount != 12 || item.CpuUsage != "-" || item.MemoryUsage != "-" {
		t.Fatalf("expected usage fallback applied, got %#v", item)
	}
	if item.Age == "" {
		t.Fatalf("expected age populated, got %#v", item)
	}
	if !item.Unschedulable || item.Labels["topology.kubernetes.io/zone"] != "cn-shanghai-a" {
		t.Fatalf("expected labels and unschedulable preserved, got %#v", item)
	}
	if len(item.Taints) != 1 {
		t.Fatalf("expected taints converted, got %#v", item)
	}
}

func TestBuildNodePodsAndAllocatedResources_SummarizesNonTerminatedPods(t *testing.T) {
	allocatable := corev1.ResourceList{
		corev1.ResourceCPU:              resource.MustParse("2"),
		corev1.ResourceMemory:           resource.MustParse("4Gi"),
		corev1.ResourceEphemeralStorage: resource.MustParse("10Gi"),
	}
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mysql", Namespace: "default", CreationTimestamp: metav1.NewTime(time.Now().Add(-1 * time.Hour))},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:              resource.MustParse("500m"),
						corev1.ResourceMemory:           resource.MustParse("512Mi"),
						corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:              resource.MustParse("1"),
						corev1.ResourceMemory:           resource.MustParse("1Gi"),
						corev1.ResourceEphemeralStorage: resource.MustParse("2Gi"),
					},
				},
			}}},
			Status: corev1.PodStatus{Phase: corev1.PodRunning, ContainerStatuses: []corev1.ContainerStatus{{RestartCount: 1}}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "redis", Namespace: "default", CreationTimestamp: metav1.NewTime(time.Now().Add(-30 * time.Minute))},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("256Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("500m"),
						corev1.ResourceMemory: resource.MustParse("512Mi"),
					},
				},
			}}},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
		{ObjectMeta: metav1.ObjectMeta{Name: "job-done", Namespace: "default"}, Status: corev1.PodStatus{Phase: corev1.PodSucceeded}},
	}

	items, allocated := buildNodePodsAndAllocatedResources(pods, allocatable)

	if len(items) != 2 {
		t.Fatalf("expected 2 non-terminated pods, got %d", len(items))
	}
	if items[0].Namespace == "" || items[0].Status == "" || items[0].Age == "" {
		t.Fatalf("expected pod row fields populated, got %#v", items[0])
	}
	if allocated.CPURequests != "700m" || allocated.CPULimits != "1500m" {
		t.Fatalf("unexpected cpu allocation summary: %#v", allocated)
	}
	if allocated.MemoryRequests == "" || allocated.MemoryLimits == "" {
		t.Fatalf("expected memory summary populated, got %#v", allocated)
	}
	if allocated.EphemeralStorageRequests != "1Gi" || allocated.EphemeralStorageLimits != "2Gi" {
		t.Fatalf("expected ephemeral storage summary populated, got %#v", allocated)
	}
	if allocated.CPURequestsPercentage == "-" || allocated.MemoryRequestsPercentage == "-" {
		t.Fatalf("expected allocation percentages populated, got %#v", allocated)
	}
}

func strPtr(v string) *string { return &v }
