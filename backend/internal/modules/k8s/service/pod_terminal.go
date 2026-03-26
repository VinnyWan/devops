package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"devops-platform/internal/pkg/k8s"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// DetectShellRequest shell检测请求
type DetectShellRequest struct {
	Namespace string `json:"namespace"`
	PodName   string `json:"podName"`
	Container string `json:"container,omitempty"`
}

// DetectShellResponse shell检测响应
type DetectShellResponse struct {
	AvailableShells  []string `json:"availableShells"`  // 可用的shell列表
	RecommendedShell string   `json:"recommendedShell"` // 推荐的shell
}

// DetectContainerShell 检测容器可用的shell
func (s *K8sService) DetectContainerShell(clusterId uint, namespace, podName, container string) (*DetectShellResponse, error) {
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

	// 获取Pod信息以确定容器
	pod, err := client.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod失败: %w", err)
	}

	// 如果未指定容器，使用第一个容器
	if container == "" {
		if len(pod.Spec.Containers) == 0 {
			return nil, fmt.Errorf("Pod中没有容器")
		}
		container = pod.Spec.Containers[0].Name
	}

	// 常见shell列表，按优先级排序
	type shellInfo struct {
		path     string
		name     string
		priority int
	}

	shellsToTry := []shellInfo{
		{"/bin/bash", "bash", 1},
		{"/usr/bin/bash", "bash", 2},
		{"/bin/sh", "sh", 3},
		{"/usr/bin/sh", "sh", 4},
	}

	availableShellsMap := make(map[string]bool)
	var foundShells []shellInfo

	restCfg, err := k8s.BuildRestConfigFromCluster(cluster)
	if err != nil {
		return nil, err
	}

	// 测试每个shell是否可用
	for _, shellInfo := range shellsToTry {
		if s.testShellAvailability(client, restCfg, namespace, podName, container, shellInfo.path) {
			if !availableShellsMap[shellInfo.name] {
				availableShellsMap[shellInfo.name] = true
				foundShells = append(foundShells, shellInfo)
			}
		}
	}

	// 如果没有找到任何shell，默认返回sh
	if len(foundShells) == 0 {
		return &DetectShellResponse{
			AvailableShells:  []string{"sh"},
			RecommendedShell: "sh",
		}, nil
	}

	// 按优先级排序找到的shells
	for i := 0; i < len(foundShells); i++ {
		for j := i + 1; j < len(foundShells); j++ {
			if foundShells[i].priority > foundShells[j].priority {
				foundShells[i], foundShells[j] = foundShells[j], foundShells[i]
			}
		}
	}

	// 构建返回的shell列表
	availableShells := make([]string, 0, len(foundShells))
	for _, shell := range foundShells {
		availableShells = append(availableShells, shell.name)
	}

	// 推荐使用优先级最高的（应该是bash如果可用）
	recommended := foundShells[0].name

	return &DetectShellResponse{
		AvailableShells:  availableShells,
		RecommendedShell: recommended,
	}, nil
}

// testShellAvailability 测试指定shell是否在容器中可用
func (s *K8sService) testShellAvailability(clientSet interface{}, restCfg *rest.Config, namespace, podName, container, shellPath string) bool {
	// 类型断言获取Clientset
	k8sClient, ok := clientSet.(*kubernetes.Clientset)
	if !ok {
		return false
	}

	// 使用test命令检查shell是否存在
	// 同时检查文件是否存在和可执行
	testCmd := []string{"sh", "-c", fmt.Sprintf("test -x %s && %s --version", shellPath, shellPath)}

	req := k8sClient.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   testCmd,
		Stdin:     false,
		Stdout:    false,
		Stderr:    false,
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return false
	}

	// 创建一个超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 尝试执行命令
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: nil,
		Stderr: nil,
		Tty:    false,
	})

	// 如果执行成功，shell可用
	if err == nil {
		return true
	}

	// 检查是否是"not found"或"no such file"错误
	errStr := err.Error()
	if strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "No such file") ||
		strings.Contains(errStr, "command not found") {
		return false
	}

	// 容器正在启动或其他错误，暂时认为shell可能存在
	// 这样可以避免在容器启动阶段误判
	return true
}

// CreatePodExecutor 创建Pod的executor（用于WebSocket终端）
// 支持自动降级：优先使用bash，失败则使用sh
func (s *K8sService) CreatePodExecutor(clusterId uint, namespace, podName, container, shell string) (remotecommand.Executor, string, error) {
	if err := s.ensureReady(); err != nil {
		return nil, "", err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return nil, "", err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, "", err
	}

	restCfg, err := k8s.BuildRestConfigFromCluster(cluster)
	if err != nil {
		return nil, "", err
	}

	// 确定要尝试的shell列表
	var shellsToTry []string
	if shell == "" || shell == "bash" {
		// 优先尝试bash，失败则降级到sh
		shellsToTry = []string{"/bin/bash", "/bin/sh"}
	} else if strings.HasPrefix(shell, "/") {
		// 绝对路径，直接使用
		shellsToTry = []string{shell}
	} else {
		// 简单名称，转换为标准路径
		shellsToTry = []string{"/bin/" + shell}
	}

	var lastErr error
	var usedShell string

	// 依次尝试每个shell
	for _, shellPath := range shellsToTry {
		req := client.CoreV1().RESTClient().
			Post().
			Resource("pods").
			Name(podName).
			Namespace(namespace).
			SubResource("exec")

		req.VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   []string{shellPath},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

		executor, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
		if err != nil {
			lastErr = fmt.Errorf("create executor for %s failed: %w", shellPath, err)
			continue
		}

		// 记录实际使用的shell（去除路径前缀）
		if strings.Contains(shellPath, "bash") {
			usedShell = "bash"
		} else {
			usedShell = "sh"
		}

		return executor, usedShell, nil
	}

	// 所有shell都失败了
	return nil, "", fmt.Errorf("所有shell尝试失败，最后错误: %w", lastErr)
}

func (s *K8sService) ExecPodTerminal(
	ctx context.Context,
	clusterId uint,
	namespace string,
	pod string,
	container string,
	command []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	tty bool,
	sizeQueue remotecommand.TerminalSizeQueue,
) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	cluster, err := s.clusterService.GetByID(clusterId)
	if err != nil {
		return err
	}

	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return err
	}

	restCfg, err := k8s.BuildRestConfigFromCluster(cluster)
	if err != nil {
		return err
	}

	req := client.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     stdin != nil,
		Stdout:    stdout != nil,
		Stderr:    stderr != nil && !tty,
		TTY:       tty,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("create executor failed: %w", err)
	}

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               tty,
		TerminalSizeQueue: sizeQueue,
	})
}
