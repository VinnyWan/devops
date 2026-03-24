package service

import (
	"context"
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
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Hosts     []string          `json:"hosts"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"createdAt"`
}

// IngressListResponse Ingress 列表分页响应
type IngressListResponse struct {
	Total int64       `json:"total"`
	Items []IngressVO `json:"items"`
}

func (s *K8sService) ListIngresses(clusterId uint, namespace string, page, pageSize int, keyword string) (*IngressListResponse, error) {
	cc, err := s.getClusterClient(clusterId)
	if err != nil {
		return nil, err
	}

	var allItems []IngressVO
	list, err := cc.Client.NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		if isIngressAPINotSupported(err) {
			dynamicClient, derr := s.clientFactory.GetDynamicClient(cc.Cluster)
			if derr != nil {
				return nil, s.handleClientError(clusterId, derr)
			}
			if data, ok, derr := listIngressesWithFallback(dynamicClient, namespace); derr == nil && ok {
				allItems = data
			} else if derr != nil {
				return nil, s.handleClientError(clusterId, derr)
			}
		}
		if allItems == nil {
			return nil, s.handleClientError(clusterId, err)
		}
	} else {
		allItems = make([]IngressVO, 0, len(list.Items))
		for _, item := range list.Items {
			hosts := make([]string, 0)
			for _, rule := range item.Spec.Rules {
				hosts = append(hosts, rule.Host)
			}
			allItems = append(allItems, IngressVO{
				Name:      item.Name,
				Namespace: item.Namespace,
				Hosts:     hosts,
				Labels:    item.Labels,
				CreatedAt: item.CreationTimestamp.Time,
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
		}
	}

	return IngressVO{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
		Hosts:     hosts,
		Labels:    item.GetLabels(),
		CreatedAt: item.GetCreationTimestamp().Time,
	}
}

func (s *K8sService) GetIngressDetail(clusterId uint, namespace, name string) (*networkingv1.Ingress, error) {
	cc, err := s.getClusterClient(clusterId)
	if err != nil {
		return nil, err
	}
	return cc.Client.NetworkingV1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) CreateIngress(clusterId uint, namespace string, ingress *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	cc, err := s.getClusterClient(clusterId)
	if err != nil {
		return nil, err
	}
	return cc.Client.NetworkingV1().Ingresses(namespace).Create(context.Background(), ingress, metav1.CreateOptions{})
}

func (s *K8sService) UpdateIngress(clusterId uint, namespace string, ingress *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	cc, err := s.getClusterClient(clusterId)
	if err != nil {
		return nil, err
	}
	return cc.Client.NetworkingV1().Ingresses(namespace).Update(context.Background(), ingress, metav1.UpdateOptions{})
}

func (s *K8sService) DeleteIngress(clusterId uint, namespace, name string) error {
	cc, err := s.getClusterClient(clusterId)
	if err != nil {
		return err
	}
	return cc.Client.NetworkingV1().Ingresses(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}
