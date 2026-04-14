package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func (s *K8sService) GetServiceObject(clusterName string, namespace, name string) (*corev1.Service, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) UpdateServiceByYAML(clusterName string, namespace, name, rawYAML string) (*corev1.Service, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	current, err := s.GetServiceObject(clusterName, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current service: %w", err)
	}

	var desired corev1.Service
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	if desired.APIVersion != "" && desired.APIVersion != "v1" {
		return nil, fmt.Errorf("不允许修改 apiVersion")
	}
	if desired.Kind != "" && desired.Kind != "Service" {
		return nil, fmt.Errorf("不允许修改 kind")
	}
	if desired.Name != "" && desired.Name != name {
		return nil, fmt.Errorf("不允许修改 metadata.name")
	}
	if desired.Namespace != "" && desired.Namespace != namespace {
		return nil, fmt.Errorf("不允许修改 metadata.namespace")
	}

	desired.Namespace = namespace
	desired.Name = name
	desired.ResourceVersion = current.ResourceVersion
	desired.Status = corev1.ServiceStatus{}
	desired.ManagedFields = nil

	return s.UpdateService(clusterName, namespace, &desired)
}
