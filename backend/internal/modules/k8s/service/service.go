package service

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceVO struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	ClusterIP   string            `json:"clusterIP"`
	Ports       []string          `json:"ports"`
	TargetPort  []string          `json:"targetPort"`
	Selector    map[string]string `json:"selector"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Endpoints   []string          `json:"endpoints"`
	CreatedAt   time.Time         `json:"createdAt"`
}

type ServiceListResponse struct {
	Total int64       `json:"total"`
	Items []ServiceVO `json:"items"`
}

func (s *K8sService) ListServices(clusterName string, namespace string, page, pageSize int, keyword string) (*ServiceListResponse, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}

	list, err := cc.Client.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, s.handleClientError(clusterName, err)
	}

	// Fetch Endpoints resources to resolve backend IPs
	endpointsList, _ := cc.Client.CoreV1().Endpoints(namespace).List(context.Background(), metav1.ListOptions{})
	endpointsMap := make(map[string]*corev1.Endpoints)
	if endpointsList != nil {
		for i := range endpointsList.Items {
			ep := &endpointsList.Items[i]
			endpointsMap[ep.Namespace+"/"+ep.Name] = ep
		}
	}

	filtered := filterByKeywordFields(list.Items, keyword, func(item corev1.Service) []string {
		return []string{
			item.Name,
			item.Namespace,
			string(item.Spec.Type),
			item.Spec.ClusterIP,
			flattenLabels(item.Labels),
		}
	})

	paged, total := paginateItems(filtered, page, pageSize)

	result := make([]ServiceVO, 0, len(paged))
	for _, item := range paged {
		ports := make([]string, 0, len(item.Spec.Ports))
		targetPorts := make([]string, 0, len(item.Spec.Ports))
		for _, p := range item.Spec.Ports {
			if p.NodePort != 0 {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, p.Protocol))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Protocol))
			}
			tp := ""
			if p.TargetPort.IntVal != 0 {
				tp = fmt.Sprintf("%d/%s", p.TargetPort.IntVal, p.Protocol)
			} else if p.TargetPort.StrVal != "" {
				tp = p.TargetPort.StrVal
			}
			targetPorts = append(targetPorts, tp)
		}

		// Resolve endpoints
		epList := make([]string, 0)
		if ep, ok := endpointsMap[item.Namespace+"/"+item.Name]; ok {
			for _, subset := range ep.Subsets {
				for _, addr := range subset.Addresses {
					for _, port := range subset.Ports {
						epList = append(epList, fmt.Sprintf("%s:%d", addr.IP, port.Port))
					}
				}
			}
		}

		result = append(result, ServiceVO{
			Name:        item.Name,
			Namespace:   item.Namespace,
			Type:        string(item.Spec.Type),
			ClusterIP:   item.Spec.ClusterIP,
			Ports:       ports,
			TargetPort:  targetPorts,
			Selector:    item.Spec.Selector,
			Labels:      item.Labels,
			Annotations: item.Annotations,
			Endpoints:   epList,
			CreatedAt:   item.CreationTimestamp.Time,
		})
	}
	return &ServiceListResponse{Total: total, Items: result}, nil
}

func (s *K8sService) GetServiceDetail(clusterName string, namespace, name string) (*ServiceVO, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}

	item, err := cc.Client.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	ports := make([]string, 0, len(item.Spec.Ports))
	targetPorts := make([]string, 0, len(item.Spec.Ports))
	for _, p := range item.Spec.Ports {
		if p.NodePort != 0 {
			ports = append(ports, fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, p.Protocol))
		} else {
			ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Protocol))
		}
		tp := ""
		if p.TargetPort.IntVal != 0 {
			tp = fmt.Sprintf("%d/%s", p.TargetPort.IntVal, p.Protocol)
		} else if p.TargetPort.StrVal != "" {
			tp = p.TargetPort.StrVal
		}
		targetPorts = append(targetPorts, tp)
	}

	// Resolve endpoints
	epList := make([]string, 0)
	if ep, err := cc.Client.CoreV1().Endpoints(namespace).Get(context.Background(), name, metav1.GetOptions{}); err == nil {
		for _, subset := range ep.Subsets {
			for _, addr := range subset.Addresses {
				for _, port := range subset.Ports {
					epList = append(epList, fmt.Sprintf("%s:%d", addr.IP, port.Port))
				}
			}
		}
	}

	return &ServiceVO{
		Name:        item.Name,
		Namespace:   item.Namespace,
		Type:        string(item.Spec.Type),
		ClusterIP:   item.Spec.ClusterIP,
		Ports:       ports,
		TargetPort:  targetPorts,
		Selector:    item.Spec.Selector,
		Labels:      item.Labels,
		Annotations: item.Annotations,
		Endpoints:   epList,
		CreatedAt:   item.CreationTimestamp.Time,
	}, nil
}

func (s *K8sService) DeleteService(clusterName string, namespace, name string) error {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return err
	}
	return cc.Client.CoreV1().Services(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (s *K8sService) CreateService(clusterName string, namespace string, svc *corev1.Service) (*corev1.Service, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}
	return cc.Client.CoreV1().Services(namespace).Create(context.Background(), svc, metav1.CreateOptions{})
}

func (s *K8sService) UpdateService(clusterName string, namespace string, svc *corev1.Service) (*corev1.Service, error) {
	cc, err := s.getClusterClient(clusterName)
	if err != nil {
		return nil, err
	}
	return cc.Client.CoreV1().Services(namespace).Update(context.Background(), svc, metav1.UpdateOptions{})
}
