package bootstrap

import (
	"devops-platform/internal/pkg/k8s"
)

var K8sFactory *k8s.ClientFactory

func InitK8sFactory() error {
	K8sFactory = k8s.NewClientFactory()
	return nil
}
