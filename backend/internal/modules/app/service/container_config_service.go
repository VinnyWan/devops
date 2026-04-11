package service

import (
	"fmt"
	"regexp"
)

type ContainerConfigService struct{}

func NewContainerConfigService() *ContainerConfigService {
	return &ContainerConfigService{}
}

// ValidateResourceFormat 校验CPU和内存格式
func ValidateResourceFormat(cpu, memory string) error {
	if cpu == "" {
		return fmt.Errorf("CPU不能为空")
	}
	if memory == "" {
		return fmt.Errorf("内存不能为空")
	}

	cpuRegex := regexp.MustCompile(`^(\d+m|\d+)$`)
	if !cpuRegex.MatchString(cpu) {
		return fmt.Errorf("CPU格式无效,应为: 100m, 1, 2")
	}

	// 纯数字必须是整数(不能是500这样的多位数,应该是1,2,4,8)
	if matched, _ := regexp.MatchString(`^\d{2,}$`, cpu); matched {
		return fmt.Errorf("CPU格式无效,整数应为: 1, 2, 4, 8")
	}

	memRegex := regexp.MustCompile(`^\d+(Mi|Gi)$`)
	if !memRegex.MatchString(memory) {
		return fmt.Errorf("内存格式无效,应为: 128Mi, 1Gi")
	}

	return nil
}

// GetEnvPresets 获取预设环境变量
func GetEnvPresets() map[string]string {
	return map[string]string{
		"JAVA_OPTS": "-Xms512m -Xmx1024m",
		"NODE_ENV":  "production",
	}
}
