package k8s

import "time"

// NamespaceDTO 命名空间列表返回对象
type NamespaceDTO struct {
	Name string `json:"name"` // 命名空间名称
}

// DeploymentDTO Deployment列表返回对象
type DeploymentDTO struct {
	Name       string            `json:"name"`       // 名称
	Namespace  string            `json:"namespace"`  // 命名空间
	Replicas   int32             `json:"replicas"`   // 副本数
	Versions   []string          `json:"versions"`   // 镜像版本号列表
	Labels     map[string]string `json:"labels"`     // 标签
	CreateTime time.Time         `json:"createTime"` // 创建时间
	UpdateTime time.Time         `json:"updateTime"` // 更新时间
}
