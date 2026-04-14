package service

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func (s *K8sService) GetIngressObject(clusterName string, namespace, name string) (*networkingv1.Ingress, error) {
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
	return client.NetworkingV1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) UpdateIngressByYAML(clusterName string, namespace, name, rawYAML string) (*networkingv1.Ingress, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	current, err := s.GetIngressObject(clusterName, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current ingress: %w", err)
	}

	var desired networkingv1.Ingress
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	if desired.APIVersion != "" && desired.APIVersion != "networking.k8s.io/v1" {
		return nil, fmt.Errorf("不允许修改 apiVersion")
	}
	if desired.Kind != "" && desired.Kind != "Ingress" {
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
	desired.Status = networkingv1.IngressStatus{}
	desired.ManagedFields = nil

	return s.UpdateIngress(clusterName, namespace, &desired)
}
