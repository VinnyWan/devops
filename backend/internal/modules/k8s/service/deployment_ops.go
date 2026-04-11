package service

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

func (s *K8sService) GetDeploymentObject(clusterName string, namespace, name string) (*appsv1.Deployment, error) {
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
	return client.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) GetDeploymentYAML(clusterName string, namespace, name string) (string, error) {
	obj, err := s.GetDeploymentObject(clusterName, namespace, name)
	if err != nil {
		return "", err
	}
	b, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *K8sService) UpdateDeploymentByYAML(clusterName string, namespace, name, rawYAML string) (*appsv1.Deployment, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	current, err := s.GetDeploymentObject(clusterName, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current deployment: %w", err)
	}

	var desired appsv1.Deployment
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	// 校验不可变字段
	if desired.APIVersion != "" && desired.APIVersion != "apps/v1" {
		return nil, fmt.Errorf("不允许修改 apiVersion")
	}
	if desired.Kind != "" && desired.Kind != "Deployment" {
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
	desired.Status = appsv1.DeploymentStatus{}
	desired.ManagedFields = nil

	return s.UpdateDeployment(clusterName, namespace, &desired)
}

func (s *K8sService) RestartDeployment(clusterName string, namespace, name string) (*appsv1.Deployment, error) {
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

	now := time.Now().UTC().Format(time.RFC3339Nano)
	patch := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":%q}}}}}`, now)
	return client.AppsV1().Deployments(namespace).Patch(context.Background(), name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
}

func (s *K8sService) ScaleDeployment(clusterName string, namespace, name string, replicas int32) (*appsv1.Deployment, error) {
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

	patch := fmt.Sprintf(`{"spec":{"replicas":%d}}`, replicas)
	return client.AppsV1().Deployments(namespace).Patch(context.Background(), name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
}
