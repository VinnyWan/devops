package k8s

import "testing"

// TestIsVersionSupported 测试版本支持检查
func TestIsVersionSupported(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		// 支持的版本（>= 1.17）
		{"K8s 1.17.0", "v1.17.0", true},
		{"K8s 1.17.15", "v1.17.15", true},
		{"K8s 1.18.0", "v1.18.0", true},
		{"K8s 1.19.0", "v1.19.0", true},
		{"K8s 1.20.0", "v1.20.0", true},
		{"K8s 1.21.0", "v1.21.0", true},
		{"K8s 1.22.0", "v1.22.0", true},
		{"K8s 1.23.0", "v1.23.0", true},
		{"K8s 1.24.0", "v1.24.0", true},
		{"K8s 1.25.0", "v1.25.0", true},
		{"K8s 1.26.0", "v1.26.0", true},
		{"K8s 1.27.0", "v1.27.0", true},
		{"K8s 1.28.0", "v1.28.0", true},
		{"K8s 1.29.0", "v1.29.0", true},
		{"K8s 1.30.0", "v1.30.0", true},

		// 带额外标记的版本
		{"K3s 1.28.5+k3s1", "v1.28.5+k3s1", true},
		{"K3s 1.27.8+k3s2", "v1.27.8+k3s2", true},
		{"K8s 1.23.0-alpha.1", "v1.23.0-alpha.1", true},
		{"K8s 1.24.0-rc.0", "v1.24.0-rc.0", true},
		{"K8s 1.17.0-eks", "v1.17.0-eks", true},

		// 未来版本（应该支持）
		{"K8s 1.31.0", "v1.31.0", true},
		{"K8s 1.35.0", "v1.35.0", true},
		{"K8s 2.0.0", "v2.0.0", true},

		// 不支持的版本（< 1.17）
		{"K8s 1.16.0", "v1.16.0", false},
		{"K8s 1.16.15", "v1.16.15", false},
		{"K8s 1.15.0", "v1.15.0", false},
		{"K8s 1.14.0", "v1.14.0", false},
		{"K8s 1.13.0", "v1.13.0", false},
		{"K8s 1.10.0", "v1.10.0", false},
		{"K8s 1.8.0", "v1.8.0", false},

		// 无效格式
		{"Invalid version 1", "invalid", false},
		{"Invalid version 2", "v1", false},
		{"Invalid version 3", "1.x.0", false},
		{"Empty version", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isVersionSupported(tt.version); got != tt.want {
				t.Errorf("isVersionSupported(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

// TestVersionParsing 测试版本解析
func TestVersionParsing(t *testing.T) {
	// 测试边界情况
	tests := []struct {
		name    string
		version string
		want    bool
		desc    string
	}{
		{
			"Exactly 1.17",
			"v1.17.0",
			true,
			"最低支持版本 1.17",
		},
		{
			"Just below 1.17",
			"v1.16.99",
			false,
			"1.16 不支持",
		},
		{
			"K3s with plus sign",
			"v1.28.5+k3s1",
			true,
			"K3s 版本格式支持",
		},
		{
			"EKS version",
			"v1.27.9-eks-ba74326",
			true,
			"EKS 版本格式支持",
		},
		{
			"GKE version",
			"v1.26.5-gke.1200",
			true,
			"GKE 版本格式支持",
		},
		{
			"OpenShift version",
			"v1.25.0+c4e6d60",
			true,
			"OpenShift 版本格式支持",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isVersionSupported(tt.version)
			if got != tt.want {
				t.Errorf("%s: isVersionSupported(%q) = %v, want %v", tt.desc, tt.version, got, tt.want)
			} else {
				t.Logf("✓ %s: %q -> %v", tt.desc, tt.version, got)
			}
		})
	}
}
