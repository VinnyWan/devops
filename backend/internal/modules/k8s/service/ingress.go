package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type IngressVO struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	IngressClass   string            `json:"ingressClass"`
	Hosts          []string          `json:"hosts"`
	Paths          []string          `json:"paths"`
	BackendService string            `json:"backendService"`
	BackendPort    string            `json:"backendPort"`
	Labels         map[string]string `json:"labels"`
	CreatedAt      time.Time         `json:"createdAt"`
}

// IngressListResponse Ingress 列表分页响应
type IngressListResponse struct {
	Total int64       `json:"total"`
	Items []IngressVO `json:"items"`
}

func (s *K8sService) ListIngresses(clusterName string, namespace string, page, pageSize int, keyword string) (*IngressListResponse, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}

	var allItems []IngressVO
	list, err := cc.Client.NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		if isIngressAPINotSupported(err) {
			dynamicClient, derr := s.clientFactory.GetDynamicClient(cc.Cluster)
			if derr != nil {
				return nil, s.handleClientError(clusterName, derr)
			}
			if data, ok, derr := listIngressesWithFallback(dynamicClient, namespace); derr == nil && ok {
				allItems = data
			} else if derr != nil {
				return nil, s.handleClientError(clusterName, derr)
			}
		}
		if allItems == nil {
			return nil, s.handleClientError(clusterName, err)
		}
	} else {
		allItems = make([]IngressVO, 0, len(list.Items))
		for _, item := range list.Items {
			ingressClass, hosts, paths, backendService, backendPort := extractIngressDetails(item)
			allItems = append(allItems, IngressVO{
				Name:           item.Name,
				Namespace:      item.Namespace,
				IngressClass:   ingressClass,
				Hosts:          hosts,
				Paths:          paths,
				BackendService: backendService,
				BackendPort:    backendPort,
				Labels:         item.Labels,
				CreatedAt:      item.CreationTimestamp.Time,
			})
		}
	}

	filtered := filterByKeywordFields(allItems, keyword, func(item IngressVO) []string {
		return []string{
			item.Name,
			item.Namespace,
			flattenLabels(item.Labels),
			strings.Join(item.Hosts, ","),
		}
	})

	paged, total := paginateItems(filtered, page, pageSize)
	return &IngressListResponse{Total: total, Items: paged}, nil
}

func extractIngressDetails(item networkingv1.Ingress) (ingressClass string, hosts []string, paths []string, backendService string, backendPort string) {
	if item.Spec.IngressClassName != nil {
		ingressClass = *item.Spec.IngressClassName
	}
	if item.Spec.DefaultBackend != nil {
		backendService = item.Spec.DefaultBackend.Service.Name
		if item.Spec.DefaultBackend.Service.Port.Name != "" {
			backendPort = item.Spec.DefaultBackend.Service.Port.Name
		} else {
			backendPort = fmt.Sprintf("%d", item.Spec.DefaultBackend.Service.Port.Number)
		}
	}
	for _, rule := range item.Spec.Rules {
		if rule.Host != "" {
			hosts = append(hosts, rule.Host)
		}
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				paths = append(paths, path.Path)
				if backendService == "" && path.Backend.Service != nil {
					backendService = path.Backend.Service.Name
					if path.Backend.Service.Port.Name != "" {
						backendPort = path.Backend.Service.Port.Name
					} else {
						backendPort = fmt.Sprintf("%d", path.Backend.Service.Port.Number)
					}
				}
			}
		}
	}
	return
}

func isIngressAPINotSupported(err error) bool {
	if apierrors.IsNotFound(err) {
		return true
	}
	return strings.Contains(err.Error(), "could not find the requested resource")
}

func listIngressesWithFallback(dynamicClient dynamic.Interface, namespace string) ([]IngressVO, bool, error) {
	gvrs := []schema.GroupVersionResource{
		{Group: "networking.k8s.io", Version: "v1beta1", Resource: "ingresses"},
		{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},
	}

	for _, gvr := range gvrs {
		list, err := dynamicClient.Resource(gvr).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) || strings.Contains(err.Error(), "could not find the requested resource") {
				continue
			}
			return nil, false, err
		}

		result := make([]IngressVO, 0, len(list.Items))
		for _, item := range list.Items {
			result = append(result, ingressVOFromUnstructured(item))
		}
		return result, true, nil
	}

	return nil, false, nil
}

func ingressVOFromUnstructured(item unstructured.Unstructured) IngressVO {
	hosts := make([]string, 0)
	paths := make([]string, 0)
	var ingressClass, backendService, backendPort string

	if ic, found, _ := unstructured.NestedString(item.Object, "spec", "ingressClassName"); found {
		ingressClass = ic
	}

	if svcName, found, _ := unstructured.NestedString(item.Object, "spec", "defaultBackend", "serviceName"); found {
		backendService = svcName
	}
	if portName, found, _ := unstructured.NestedString(item.Object, "spec", "defaultBackend", "servicePort"); found {
		backendPort = portName
	}

	if rules, found, _ := unstructured.NestedSlice(item.Object, "spec", "rules"); found {
		for _, rule := range rules {
			ruleMap, ok := rule.(map[string]interface{})
			if !ok {
				continue
			}
			host, ok := ruleMap["host"].(string)
			if ok && host != "" {
				hosts = append(hosts, host)
			}
			httpVal, ok := ruleMap["http"].(map[string]interface{})
			if !ok {
				continue
			}
			httpPaths, ok := httpVal["paths"].([]interface{})
			if !ok {
				continue
			}
			for _, p := range httpPaths {
				pathMap, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				if pathStr, ok := pathMap["path"].(string); ok {
					paths = append(paths, pathStr)
				}
				if backendService == "" {
					if svcName, ok := pathMap["backend"].(map[string]interface{}); ok {
						if sn, ok := svcName["serviceName"].(string); ok {
							backendService = sn
						}
						if sp, ok := svcName["servicePort"]; ok {
							switch v := sp.(type) {
							case string:
								backendPort = v
							case float64:
								backendPort = fmt.Sprintf("%.0f", v)
							case int:
								backendPort = fmt.Sprintf("%d", v)
							}
						}
					}
				}
			}
		}
	}

	return IngressVO{
		Name:           item.GetName(),
		Namespace:      item.GetNamespace(),
		IngressClass:   ingressClass,
		Hosts:          hosts,
		Paths:          paths,
		BackendService: backendService,
		BackendPort:    backendPort,
		Labels:         item.GetLabels(),
		CreatedAt:      item.GetCreationTimestamp().Time,
	}
}

func (s *K8sService) GetIngressDetail(clusterName string, namespace, name string) (*networkingv1.Ingress, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}
	return cc.Client.NetworkingV1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) CreateIngress(clusterName string, namespace string, ingress *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}
	return cc.Client.NetworkingV1().Ingresses(namespace).Create(context.Background(), ingress, metav1.CreateOptions{})
}

func (s *K8sService) UpdateIngress(clusterName string, namespace string, ingress *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}
	return cc.Client.NetworkingV1().Ingresses(namespace).Update(context.Background(), ingress, metav1.UpdateOptions{})
}

func (s *K8sService) DeleteIngress(clusterName string, namespace, name string) error {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return err
	}
	return cc.Client.NetworkingV1().Ingresses(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
