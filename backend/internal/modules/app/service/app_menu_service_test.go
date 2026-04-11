package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAppManagementMenuOptions(t *testing.T) {
	service := NewAppMenuService()

	options := service.GetAppManagementMenuOptions()

	assert.Len(t, options, 4, "应该返回4个菜单选项")
	assert.Equal(t, "应用配置", options[0].Label)
	assert.Equal(t, "app-config", options[0].Value)
	assert.Equal(t, "构建配置", options[1].Label)
	assert.Equal(t, "build-config", options[1].Value)
	assert.Equal(t, "部署配置", options[2].Label)
	assert.Equal(t, "deploy-config", options[2].Value)
	assert.Equal(t, "容器配置", options[3].Label)
	assert.Equal(t, "container-config", options[3].Value)
}
