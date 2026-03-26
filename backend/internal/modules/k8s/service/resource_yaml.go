package service

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

// ResourceGVR 资源类型到 GroupVersionResource 的映射
var resourceGVRMap = map[string]schema.GroupVersionResource{
	// 工作负载
	"pod":         {Group: "", Version: "v1", Resource: "pods"},
	"deployment":  {Group: "apps", Version: "v1", Resource: "deployments"},
	"statefulset": {Group: "apps", Version: "v1", Resource: "statefulsets"},
	"daemonset":   {Group: "apps", Version: "v1", Resource: "daemonsets"},
	"job":         {Group: "batch", Version: "v1", Resource: "jobs"},
	"cronjob":     {Group: "batch", Version: "v1", Resource: "cronjobs"},
	"replicaset":  {Group: "apps", Version: "v1", Resource: "replicasets"},
	// 服务与网络
	"service":       {Group: "", Version: "v1", Resource: "services"},
	"ingress":       {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	"endpoint":      {Group: "", Version: "v1", Resource: "endpoints"},
	"networkpolicy": {Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	// 配置与存储
	"configmap":    {Group: "", Version: "v1", Resource: "configmaps"},
	"secret":       {Group: "", Version: "v1", Resource: "secrets"},
	"pvc":          {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	"pv":           {Group: "", Version: "v1", Resource: "persistentvolumes"},
	"storageclass": {Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
	// 其他
	"namespace":      {Group: "", Version: "v1", Resource: "namespaces"},
	"node":           {Group: "", Version: "v1", Resource: "nodes"},
	"event":          {Group: "", Version: "v1", Resource: "events"},
	"limitrange":     {Group: "", Version: "v1", Resource: "limitranges"},
	"resourcequota":  {Group: "", Version: "v1", Resource: "resourcequotas"},
	"serviceaccount": {Group: "", Version: "v1", Resource: "serviceaccounts"},
	// RBAC
	"role":               {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
	"rolebinding":        {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
	"clusterrole":        {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
	"clusterrolebinding": {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
}

// GetResourceYAML 通用获取资源YAML的方法
// resourceType: 资源类型，如 pod, deployment, service 等
// namespace: 命名空间（集群级别资源传空字符串）
// name: 资源名称
func (s *K8sService) GetResourceYAML(clusterID uint, resourceType, namespace, name string) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}

	gvr, ok := resourceGVRMap[resourceType]
	if !ok {
		return "", fmt.Errorf("不支持的资源类型: %s", resourceType)
	}

	_, dynamicClient, err := s.getClusterDynamicClient(clusterID)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var obj *unstructured.Unstructured
	if gvr.Resource == "namespaces" || gvr.Resource == "nodes" || gvr.Resource == "persistentvolumes" ||
		gvr.Resource == "clusterroles" || gvr.Resource == "clusterrolebindings" || gvr.Resource == "storageclasses" {
		// 集群级别资源
		obj, err = dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	} else {
		// 命名空间级别资源
		if namespace == "" {
			return "", fmt.Errorf("命名空间不能为空")
		}
		obj, err = dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return "", fmt.Errorf("获取资源失败: %w", err)
	}

	// 移除 managedFields 字段，减少输出内容
	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")

	yamlBytes, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", fmt.Errorf("序列化YAML失败: %w", err)
	}

	return string(yamlBytes), nil
}

// GetResourceYAMLByGVR 通过 GVR 获取资源YAML（更灵活的方式）
func (s *K8sService) GetResourceYAMLByGVR(clusterID uint, gvr schema.GroupVersionResource, namespace, name string) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}

	_, dynamicClient, err := s.getClusterDynamicClient(clusterID)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var obj *unstructured.Unstructured
	if namespace == "" {
		obj, err = dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return "", fmt.Errorf("获取资源失败: %w", err)
	}

	// 移除 managedFields 字段
	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")

	yamlBytes, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", fmt.Errorf("序列化YAML失败: %w", err)
	}

	return string(yamlBytes), nil
}

// GetSupportedResourceTypes 获取支持的资源类型列表
func GetSupportedResourceTypes() []string {
	types := make([]string, 0, len(resourceGVRMap))
	for k := range resourceGVRMap {
		types = append(types, k)
	}
	return types
}
