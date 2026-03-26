package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodDescribeVO 聚合 Pod 的完整诊断信息
type PodDescribeVO struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"createdAt"`

	Phase    string `json:"phase"`
	PodIP    string `json:"podIP"`
	HostIP   string `json:"hostIP"`
	QOSClass string `json:"qosClass"`
	NodeName string `json:"nodeName"`

	Containers     []ContainerDescribeVO `json:"containers"`
	InitContainers []ContainerDescribeVO `json:"initContainers"`
	Conditions     []PodConditionVO      `json:"conditions"`
	Volumes        []VolumeInfoVO        `json:"volumes"`
	Events         []EventInfoVO         `json:"events"`
}

type ContainerDescribeVO struct {
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	State        string            `json:"state"`
	StateDetail  string            `json:"stateDetail"`
	LastState    string            `json:"lastState"`
	Ready        bool              `json:"ready"`
	RestartCount int32             `json:"restartCount"`
	Ports        []string          `json:"ports"`
	Env          map[string]string `json:"env"`
	Mounts       []string          `json:"mounts"`
	Resources    string            `json:"resources"`
}

type PodConditionVO struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
}

type VolumeInfoVO struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

type EventInfoVO struct {
	Type           string    `json:"type"`
	Reason         string    `json:"reason"`
	Message        string    `json:"message"`
	Count          int32     `json:"count"`
	FirstTimestamp time.Time `json:"firstTimestamp"`
	LastTimestamp  time.Time `json:"lastTimestamp"`
	Source         string    `json:"source"`
}

// GetPodObject 获取原始 Pod 对象
func (s *K8sService) GetPodObject(clusterId uint, namespace, name string) (*corev1.Pod, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

// DescribePod 聚合 Pod 的完整诊断信息（Pod + Events 并发获取）
func (s *K8sService) DescribePod(clusterId uint, namespace, name string) (*PodDescribeVO, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var pod *corev1.Pod
	var events *corev1.EventList

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		pod, err = client.CoreV1().Pods(namespace).Get(gctx, name, metav1.GetOptions{})
		return err
	})

	g.Go(func() error {
		var err error
		fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Pod", name, namespace)
		events, err = client.CoreV1().Events(namespace).List(gctx, metav1.ListOptions{
			FieldSelector: fieldSelector,
		})
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to describe pod: %w", err)
	}

	vo := &PodDescribeVO{
		Name:        pod.Name,
		Namespace:   pod.Namespace,
		Labels:      pod.Labels,
		Annotations: pod.Annotations,
		CreatedAt:   pod.CreationTimestamp.Time,
		Phase:       string(pod.Status.Phase),
		PodIP:       pod.Status.PodIP,
		HostIP:      pod.Status.HostIP,
		QOSClass:    string(pod.Status.QOSClass),
		NodeName:    pod.Spec.NodeName,
	}

	// 容器状态
	vo.InitContainers = buildContainerDescribes(pod.Spec.InitContainers, pod.Status.InitContainerStatuses)
	vo.Containers = buildContainerDescribes(pod.Spec.Containers, pod.Status.ContainerStatuses)

	// Conditions
	vo.Conditions = make([]PodConditionVO, 0, len(pod.Status.Conditions))
	for _, c := range pod.Status.Conditions {
		vo.Conditions = append(vo.Conditions, PodConditionVO{
			Type:               string(c.Type),
			Status:             string(c.Status),
			Reason:             c.Reason,
			Message:            c.Message,
			LastTransitionTime: c.LastTransitionTime.Time,
		})
	}

	// Volumes
	vo.Volumes = make([]VolumeInfoVO, 0, len(pod.Spec.Volumes))
	for _, v := range pod.Spec.Volumes {
		vo.Volumes = append(vo.Volumes, buildVolumeInfo(v))
	}

	// Events（按时间倒序）
	vo.Events = make([]EventInfoVO, 0, len(events.Items))
	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].LastTimestamp.After(events.Items[j].LastTimestamp.Time)
	})
	for _, e := range events.Items {
		vo.Events = append(vo.Events, EventInfoVO{
			Type:           e.Type,
			Reason:         e.Reason,
			Message:        e.Message,
			Count:          e.Count,
			FirstTimestamp: e.FirstTimestamp.Time,
			LastTimestamp:  e.LastTimestamp.Time,
			Source:         e.Source.Component,
		})
	}

	return vo, nil
}

func buildContainerDescribes(specs []corev1.Container, statuses []corev1.ContainerStatus) []ContainerDescribeVO {
	statusMap := make(map[string]corev1.ContainerStatus, len(statuses))
	for _, s := range statuses {
		statusMap[s.Name] = s
	}

	result := make([]ContainerDescribeVO, 0, len(specs))
	for _, spec := range specs {
		cd := ContainerDescribeVO{
			Name:      spec.Name,
			Image:     spec.Image,
			Resources: formatResources(spec.Resources),
		}

		// Ports
		cd.Ports = make([]string, 0, len(spec.Ports))
		for _, p := range spec.Ports {
			cd.Ports = append(cd.Ports, fmt.Sprintf("%d/%s", p.ContainerPort, p.Protocol))
		}

		// Env（过滤敏感字段）
		cd.Env = make(map[string]string, len(spec.Env))
		for _, e := range spec.Env {
			if isSensitiveEnvKey(e.Name) {
				cd.Env[e.Name] = "***"
			} else if e.ValueFrom != nil {
				cd.Env[e.Name] = "(from ref)"
			} else {
				cd.Env[e.Name] = e.Value
			}
		}

		// Mounts
		cd.Mounts = make([]string, 0, len(spec.VolumeMounts))
		for _, m := range spec.VolumeMounts {
			ro := ""
			if m.ReadOnly {
				ro = " (ro)"
			}
			cd.Mounts = append(cd.Mounts, fmt.Sprintf("%s -> %s%s", m.Name, m.MountPath, ro))
		}

		// Status
		if st, ok := statusMap[spec.Name]; ok {
			cd.Ready = st.Ready
			cd.RestartCount = st.RestartCount
			cd.State, cd.StateDetail = describeContainerState(st.State)
			cd.LastState, _ = describeContainerState(st.LastTerminationState)
		}

		result = append(result, cd)
	}
	return result
}

func describeContainerState(state corev1.ContainerState) (string, string) {
	if state.Running != nil {
		return "Running", fmt.Sprintf("Started at %s", state.Running.StartedAt.Format(time.RFC3339))
	}
	if state.Waiting != nil {
		return "Waiting", fmt.Sprintf("%s: %s", state.Waiting.Reason, state.Waiting.Message)
	}
	if state.Terminated != nil {
		return "Terminated", fmt.Sprintf("Exit %d, reason: %s", state.Terminated.ExitCode, state.Terminated.Reason)
	}
	return "", ""
}

var sensitiveKeywords = []string{"token", "password", "secret", "key", "credential", "auth"}

func isSensitiveEnvKey(name string) bool {
	lower := strings.ToLower(name)
	for _, kw := range sensitiveKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func buildVolumeInfo(v corev1.Volume) VolumeInfoVO {
	info := VolumeInfoVO{Name: v.Name}
	switch {
	case v.ConfigMap != nil:
		info.Type = "ConfigMap"
		info.Source = v.ConfigMap.Name
	case v.Secret != nil:
		info.Type = "Secret"
		info.Source = v.Secret.SecretName
	case v.PersistentVolumeClaim != nil:
		info.Type = "PVC"
		info.Source = v.PersistentVolumeClaim.ClaimName
	case v.EmptyDir != nil:
		info.Type = "EmptyDir"
	case v.HostPath != nil:
		info.Type = "HostPath"
		info.Source = v.HostPath.Path
	case v.Projected != nil:
		info.Type = "Projected"
	case v.DownwardAPI != nil:
		info.Type = "DownwardAPI"
	default:
		info.Type = "Other"
	}
	return info
}
