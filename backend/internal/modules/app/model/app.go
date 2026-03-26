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
	ID          uint      `json:"id" gorm:"primaryKey"`
	AppID       uint      `json:"appId" gorm:"uniqueIndex"`
	Name        string    `json:"name"`
	Owner       string    `json:"owner"`
	Developers  string    `json:"developers"`
	Testers     string    `json:"testers"`
	GitAddress  string    `json:"gitAddress"`
	AppState    string    `json:"appState"`
	Language    string    `json:"language"`
	Description string    `json:"description"`
	Domain      string    `json:"domain"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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

// DeployConfig 部署配置
type DeployConfig struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	AppID         uint      `json:"appId" gorm:"uniqueIndex"`
	ServicePort   int       `json:"servicePort"`
	CPURequest    string    `json:"cpuRequest"`
	CPULimit      string    `json:"cpuLimit"`
	MemoryRequest string    `json:"memoryRequest"`
	MemoryLimit   string    `json:"memoryLimit"`
	Environment   string    `json:"environment"`
	EnvVars       string    `json:"envVars"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
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

// 预定义的枚举值
const (
	AppStateRunning    = "running"
	AppStateStopped    = "stopped"
	AppStateDeveloping = "developing"

	BuildEnvDevelopment = "development"
	BuildEnvStaging     = "staging"
	BuildEnvProduction  = "production"

	EnvironmentDev     = "dev"
	EnvironmentTest    = "test"
	EnvironmentStaging = "staging"
	EnvironmentProd    = "prod"

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
