// 文件配置,解析yaml配种文件

package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// 总配文件
type config struct {
	Server server `yaml:"server"`
	Db     db     `yaml:"db"`
	Redis  redis  `yaml:"redis"`
	Jwt    jwt    `yaml:"jwt"`
	Log    Log    `yaml:"log"`
}

// 项目端口配置
type server struct {
	Port          string `yaml:"port"`
	Model         string `yaml:"model"`
	EnableSwagger bool   `yaml:"enableSwagger"` // 是否启用Swagger文档
}

// 数据库配置
type db struct {
	Dialects string `yaml:"dialects"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Db       string `yaml:"db"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Charset  string `yaml:"charset"`
	MaxIdle  int    `yaml:"maxIdle"`
	MaxOpen  int    `yaml:"maxOpen"`
}

// redis配置
type redis struct {
	Address      string `yaml:"address"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"poolSize"`
	MinIdleConns int    `yaml:"minIdleConns"`
}

// jwt配置
type jwt struct {
	Secret string `yaml:"secret"`
	Expire int64  `yaml:"expire"`
}

// Log配置
type Log struct {
	Output           string `yaml:"output"` // console, file, both
	FilePath         string `yaml:"filePath"`
	Level            string `yaml:"level"` // debug, info, warn, error
	EnableCaller     bool   `yaml:"enableCaller"`
	EnableStacktrace bool   `yaml:"enableStacktrace"`
}

var Config *config

// 配置初始化
func init() {
	// 初始化时先不加载配置文件，等待LoadConfig()被调用
}

// LoadConfig 从指定路径加载配置文件
func LoadConfig(configPath string) error {
	if configPath == "" {
		configPath = "./config.yaml" // 默认配置文件路径
	}

	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg config
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return err
	}

	Config = &cfg
	return nil
}

// GetConfig 获取数据库配置
func GetConfig() *db {
	if Config == nil {
		panic("Config is not initialized")
	}
	return &Config.Db
}

// GetRedisConfig 获取Redis配置
func GetRedisConfig() *redis {
	if Config == nil {
		panic("Config is not initialized")
	}
	return &Config.Redis
}

// GetJwtConfig 获取JWT配置
func GetJwtConfig() *jwt {
	if Config == nil {
		panic("Config is not initialized")
	}
	return &Config.Jwt
}

// GetLogConfig 获取日志配置
func GetLogConfig() *Log {
	if Config == nil {
		panic("Config is not initialized")
	}
	return &Config.Log
}
