package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListK8sClusters(t *testing.T) {
	service := NewDeployConfigService()
	clusters := service.ListK8sClusters()

	assert.NotNil(t, clusters, "集群列表不应为nil")
}

func TestListImageVersions(t *testing.T) {
	service := NewDeployConfigService()
	versions := service.ListImageVersions("test-app")

	assert.NotNil(t, versions, "镜像版本列表不应为nil")
}
