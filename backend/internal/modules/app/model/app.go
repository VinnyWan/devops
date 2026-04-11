package model

import "time"

type Application struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AppTemplate struct {
	ID          uint              `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Environment []string          `json:"environment"`
	Variables   map[string]string `json:"variables"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type ApplicationDeployment struct {
	ID           uint              `json:"id"`
	AppID        uint              `json:"appId"`
	AppName      string            `json:"appName"`
	TemplateID   uint              `json:"templateId"`
	TemplateName string            `json:"templateName"`
	Cluster      string            `json:"cluster"`
	Environment  string            `json:"environment"`
	Namespace    string            `json:"namespace"`
	Version      string            `json:"version"`
	Status       string            `json:"status"`
	Operator     string            `json:"operator"`
	Variables    map[string]string `json:"variables"`
	CreatedAt    time.Time         `json:"createdAt"`
}

type ApplicationVersion struct {
	ID          uint      `json:"id"`
	AppID       uint      `json:"appId"`
	Version     string    `json:"version"`
	Cluster     string    `json:"cluster"`
	Environment string    `json:"environment"`
	Image       string    `json:"image"`
	Status      string    `json:"status"`
	Operator    string    `json:"operator"`
	CreatedAt   time.Time `json:"createdAt"`
}

type TopologyNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
	Cluster  string `json:"cluster"`
	Metadata string `json:"metadata"`
}

type TopologyEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

type ApplicationTopology struct {
	AppID        uint           `json:"appId"`
	AppName      string         `json:"appName"`
	Environment  string         `json:"environment"`
	Nodes        []TopologyNode `json:"nodes"`
	Edges        []TopologyEdge `json:"edges"`
	LastSyncTime time.Time      `json:"lastSyncTime"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AppID        uint      `json:"appId" gorm:"uniqueIndex"`
	Name         string    `json:"name"`
	Owner        string    `json:"owner"`         // 运维负责人
	Developers   string    `json:"developers"`    // 开发负责人
	Testers      string    `json:"testers"`       // 测试负责人
	GitAddress   string    `json:"gitAddress"`    // Git地址
	AppState     string    `json:"appState"`      // 应用状态：pending/online
	Status       string    `json:"status"`        // 运行状态：running/offline
	InstanceType string    `json:"instanceType"`  // 实例类型：container/native
	Language     string    `json:"language"`      // 开发语言
	Port         int       `json:"port"`          // 服务端口
	Domain       string    `json:"domain"`        // 域名
	HealthCheck  string    `json:"healthCheck"`   // 健康检查接口
	Description  string    `json:"description"`   // 描述
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// BuildConfig 构建配置
type BuildConfig struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AppID        uint      `json:"appId" gorm:"uniqueIndex"`
	BuildEnv     string    `json:"buildEnv"`
	BuildTool    string    `json:"buildTool"`
	BuildConfig  string    `json:"buildConfig"`
	CustomConfig string    `json:"customConfig"`
	Dockerfile   string    `json:"dockerfile"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// DeployConfig 部署配置（按环境区分）
type DeployConfig struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	AppID         uint      `json:"appId" gorm:"uniqueIndex:idx_app_env"`
	Environment   string    `json:"environment" gorm:"uniqueIndex:idx_app_env"` // 环境：dev/test/staging/prod
	ClusterName   string    `json:"clusterName"`                                // K8s集群名称
	Replicas      int       `json:"replicas"`                                    // 副本数
	ServicePort   int       `json:"servicePort"`                                 // 服务端口
	CPURequest    string    `json:"cpuRequest"`                                  // CPU请求
	CPULimit      string    `json:"cpuLimit"`                                    // CPU限制
	MemoryRequest string    `json:"memoryRequest"`                               // 内存请求
	MemoryLimit   string    `json:"memoryLimit"`                                 // 内存限制
	EnvVars       string    `json:"envVars"`                                     // 环境变量
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// DeployConfigRequest 部署配置请求
type DeployConfigRequest struct {
	ID            uint   `json:"id"`
	AppID         uint   `json:"appId" binding:"required"`
	Environment   string `json:"environment" binding:"required"`
	ClusterName   string `json:"clusterName"`
	Replicas      int    `json:"replicas"`
	ServicePort   int    `json:"servicePort"`
	CPURequest    string `json:"cpuRequest"`
	CPULimit      string `json:"cpuLimit"`
	MemoryRequest string `json:"memoryRequest"`
	MemoryLimit   string `json:"memoryLimit"`
	EnvVars       string `json:"envVars"`
}

// DeleteDeployConfigRequest 删除部署配置请求
type DeleteDeployConfigRequest struct {
	AppID       uint   `json:"appId" binding:"required"`
	Environment string `json:"environment" binding:"required"`
}

// TechStackConfig 技术栈配置
type TechStackConfig struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AppID        uint      `json:"appId" gorm:"uniqueIndex"`
	Name         string    `json:"name"`
	Language     string    `json:"language"`
	Version      string    `json:"version"`
	BaseImage    string    `json:"baseImage"`
	BuildImage   string    `json:"buildImage"`
	RuntimeImage string    `json:"runtimeImage"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// BuildEnv 构建环境版本
type BuildEnv struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ContainerConfig 容器配置（按环境区分）
type ContainerConfig struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	AppID         uint      `json:"appId" gorm:"uniqueIndex:idx_container_app_env"`
	Environment   string    `json:"environment" gorm:"uniqueIndex:idx_container_app_env"` // 环境：dev/test/staging/prod
	Namespace     string    `json:"namespace"`                                             // K8s命名空间
	Image         string    `json:"image"`                                                 // Harbor镜像地址
	CPURequest    string    `json:"cpuRequest"`                                            // CPU请求
	CPULimit      string    `json:"cpuLimit"`                                              // CPU限制
	MemoryRequest string    `json:"memoryRequest"`                                         // 内存请求
	MemoryLimit   string    `json:"memoryLimit"`                                           // 内存限制
	MountPaths    string    `json:"mountPaths"`                                            // 挂载目录（JSON数组或逗号分隔）
	EnvVars       string    `json:"envVars"`                                               // 环境变量JSON
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ContainerConfigRequest 容器配置请求
type ContainerConfigRequest struct {
	ID            uint   `json:"id"`
	AppID         uint   `json:"appId" binding:"required"`
	Environment   string `json:"environment" binding:"required"`
	Namespace     string `json:"namespace"`
	Image         string `json:"image"`
	CPURequest    string `json:"cpuRequest"`
	CPULimit      string `json:"cpuLimit"`
	MemoryRequest string `json:"memoryRequest"`
	MemoryLimit   string `json:"memoryLimit"`
	MountPaths    string `json:"mountPaths"`
	EnvVars       string `json:"envVars"`
}

// DeleteContainerConfigRequest 删除容器配置请求
type DeleteContainerConfigRequest struct {
	AppID       uint   `json:"appId" binding:"required"`
	Environment string `json:"environment" binding:"required"`
}

// EnvVar 环境变量键值对
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PageRequest 分页请求
type PageRequest struct {
	Page     int    `form:"page" json:"page"`           // 页码，默认1
	PageSize int    `form:"size" json:"size"`           // 每页条数，默认10
	Keyword  string `form:"keyword" json:"keyword"`     // 搜索关键字
}

// AppListRequest 应用列表请求
type AppListRequest struct {
	PageRequest
	InstanceType string `form:"instance_type" json:"instanceType"` // 实例类型筛选
	Status       string `form:"status" json:"status"`              // 状态筛选
}

// PageResponse 分页响应
type PageResponse struct {
	Total int64       `json:"total"` // 总条数
	List  interface{} `json:"list"`  // 列表数据
}

// StatusToggleRequest 状态切换请求
type StatusToggleRequest struct {
	ID     uint   `json:"id" binding:"required"`     // 应用ID
	Status string `json:"status" binding:"required"` // 目标状态：running/offline
}

// DeleteRequest 删除请求
type DeleteRequest struct {
	ID uint `json:"id" binding:"required"` // ID
}

// 预定义的枚举值
const (
	// 应用状态
	StatusRunning = "running" // 运行中
	StatusOffline = "offline" // 已下线

	// 实例类型
	InstanceTypeContainer = "container" // 容器部署
	InstanceTypeNative    = "native"    // 原方式部署

	// 应用状态（兼容旧字段）
	AppStateRunning    = "running"
	AppStateStopped    = "stopped"
	AppStateDeveloping = "developing"

	// 构建环境
	BuildEnvDevelopment = "development"
	BuildEnvStaging     = "staging"
	BuildEnvProduction  = "production"

	// 部署环境
	EnvironmentDev     = "dev"
	EnvironmentTest    = "test"
	EnvironmentStaging = "staging"
	EnvironmentProd    = "prod"

	// 开发语言
	LanguageJava   = "java"
	LanguageGo     = "go"
	LanguagePython = "python"
	LanguageNodeJS = "nodejs"
)

// CPU/Memory 选项
var CPUOptions = []string{
	"100m", "200m", "500m", "1", "2", "4", "8",
}

var MemoryOptions = []string{
	"128Mi", "256Mi", "512Mi", "1Gi", "2Gi", "4Gi", "8Gi",
}

// GetCPUOptions 获取CPU选项列表
func GetCPUOptions() []string {
	return CPUOptions
}

// GetMemoryOptions 获取内存选项列表
func GetMemoryOptions() []string {
	return MemoryOptions
}

// ========== 枚举管理 ==========

// Enum 枚举值定义
type Enum struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	EnumType  string    `json:"enumType" gorm:"index"`           // 枚举类型：app_status, run_status, dev_language, build_tool, environment
	EnumKey   string    `json:"enumKey"`                         // 枚举键（代码中使用）
	EnumValue string    `json:"enumValue"`                       // 枚举显示值（UI展示）
	SortOrder int       `json:"sortOrder"`                       // 排序序号
	IsActive  bool      `json:"isActive"`                        // 是否启用
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// EnumRequest 枚举保存请求
type EnumRequest struct {
	ID        uint   `json:"id"`
	EnumType  string `json:"enumType" binding:"required"`
	EnumKey   string `json:"enumKey" binding:"required"`
	EnumValue string `json:"enumValue" binding:"required"`
	SortOrder int    `json:"sortOrder"`
	IsActive  bool   `json:"isActive"`
}

// EnumListRequest 枚举列表请求
type EnumListRequest struct {
	EnumType string `form:"enum_type" json:"enumType"` // 枚举类型筛选
}

// DeleteEnumRequest 删除枚举请求
type DeleteEnumRequest struct {
	ID uint `json:"id" binding:"required"`
}
