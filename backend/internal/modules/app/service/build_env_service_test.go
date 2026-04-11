package service

import (
	"testing"

	"devops-platform/internal/modules/app/model"

	"github.com/stretchr/testify/assert"
)

func TestBuildEnvService_CreateAndList(t *testing.T) {
	service := NewBuildEnvService()

	// 创建JDK版本
	jdk, err := service.CreateBuildEnv(model.BuildEnv{
		Name:    "JDK 17",
		Type:    "jdk",
		Version: "17",
	})
	assert.NoError(t, err)
	assert.Equal(t, "JDK 17", jdk.Name)

	// 创建Maven版本
	_, err = service.CreateBuildEnv(model.BuildEnv{
		Name:    "Maven 3.9",
		Type:    "maven",
		Version: "3.9",
	})
	assert.NoError(t, err)

	// 列出所有版本
	list := service.ListBuildEnvs()
	assert.Len(t, list, 2)
}

func TestBuildEnvService_Update(t *testing.T) {
	service := NewBuildEnvService()

	env, _ := service.CreateBuildEnv(model.BuildEnv{
		Name:    "Go 1.21",
		Type:    "golang",
		Version: "1.21",
	})

	// 更新版本
	env.Version = "1.22"
	updated, err := service.UpdateBuildEnv(env)
	assert.NoError(t, err)
	assert.Equal(t, "1.22", updated.Version)
}

func TestBuildEnvService_Delete(t *testing.T) {
	service := NewBuildEnvService()

	env, _ := service.CreateBuildEnv(model.BuildEnv{
		Name:    "Python 3.11",
		Type:    "python",
		Version: "3.11",
	})

	// 删除
	err := service.DeleteBuildEnv(env.ID)
	assert.NoError(t, err)

	// 验证已删除
	list := service.ListBuildEnvs()
	assert.Len(t, list, 0)
}
