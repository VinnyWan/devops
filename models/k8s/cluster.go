package k8s

import (
	"time"

	"gorm.io/gorm"
)

// Cluster K8s集群模型
type Cluster struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"name" binding:"required"` // 集群名称
	Description string         `gorm:"type:varchar(500)" json:"description"`                                  // 集群描述
	ApiServer   string         `gorm:"type:varchar(500);not null" json:"apiServer" binding:"required"`        // API Server地址
	KubeConfig  string         `gorm:"type:text;not null" json:"kubeConfig,omitempty" binding:"required"`     // KubeConfig配置
	Version     string         `gorm:"type:varchar(50)" json:"version"`                                       // K8s版本
	Status      int            `gorm:"type:tinyint;default:1" json:"status"`                                  // 状态：1-正常 0-禁用
	DeptID      uint           `gorm:"index" json:"deptId"`                                                   // 所属部门ID
	Remark      string         `gorm:"type:varchar(500)" json:"remark"`                                       // 备注
}

// TableName 指定表名
func (Cluster) TableName() string {
	return "k8s_clusters"
}

// ClusterAccess 集群访问权限模型
type ClusterAccess struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	ClusterID  uint           `gorm:"index;not null" json:"clusterId"`             // 集群ID
	RoleID     uint           `gorm:"index;not null" json:"roleId"`                // 角色ID
	AccessType string         `gorm:"type:varchar(20);not null" json:"accessType"` // 访问类型：readonly-只读，admin-管理员
	Namespaces string         `gorm:"type:text" json:"namespaces"`                 // 可访问的命名空间（JSON数组），为空表示所有
}

// TableName 指定表名
func (ClusterAccess) TableName() string {
	return "k8s_cluster_accesses"
}

// Namespace 命名空间记录（用于审计和管理）
type Namespace struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	ClusterID   uint           `gorm:"index;not null" json:"clusterId"`        // 所属集群ID
	Name        string         `gorm:"type:varchar(253);not null" json:"name"` // 命名空间名称
	Labels      string         `gorm:"type:text" json:"labels"`                // 标签（JSON）
	Annotations string         `gorm:"type:text" json:"annotations"`           // 注解（JSON）
	Status      string         `gorm:"type:varchar(20)" json:"status"`         // 状态
}

// TableName 指定表名
func (Namespace) TableName() string {
	return "k8s_namespaces"
}

// OperationLog K8s操作日志
type OperationLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	ClusterID uint      `gorm:"index;not null" json:"clusterId"`            // 集群ID
	UserID    uint      `gorm:"index;not null" json:"userId"`               // 操作用户ID
	Username  string    `gorm:"type:varchar(100)" json:"username"`          // 操作用户名
	Operation string    `gorm:"type:varchar(50);not null" json:"operation"` // 操作类型：create/update/delete/get/list/exec
	Resource  string    `gorm:"type:varchar(50);not null" json:"resource"`  // 资源类型：deployment/pod/service等
	Namespace string    `gorm:"type:varchar(253)" json:"namespace"`         // 命名空间
	Name      string    `gorm:"type:varchar(253)" json:"name"`              // 资源名称
	Result    string    `gorm:"type:varchar(20)" json:"result"`             // 操作结果：success/failed
	Message   string    `gorm:"type:text" json:"message"`                   // 详细信息
	IP        string    `gorm:"type:varchar(50)" json:"ip"`                 // 操作IP
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "k8s_operation_logs"
}
