package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateResourceFormat(t *testing.T) {
	tests := []struct {
		name    string
		cpu     string
		memory  string
		wantErr bool
	}{
		{"有效CPU和内存", "500m", "512Mi", false},
		{"有效CPU整数", "1", "1Gi", false},
		{"无效CPU格式", "500", "512Mi", true},
		{"无效内存格式", "500m", "512M", true},
		{"空CPU", "", "512Mi", true},
		{"空内存", "500m", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceFormat(tt.cpu, tt.memory)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEnvPresets(t *testing.T) {
	presets := GetEnvPresets()

	assert.NotEmpty(t, presets, "预设环境变量不应为空")
	assert.Contains(t, presets, "JAVA_OPTS", "应包含JAVA_OPTS")
	assert.Contains(t, presets, "NODE_ENV", "应包含NODE_ENV")
}
