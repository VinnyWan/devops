package service

import (
	"context"
	"fmt"
	"io"

	"devops-platform/internal/pkg/k8s"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

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

