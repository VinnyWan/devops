package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var Cfg *viper.Viper

// 命令行参数
var (
	configFile  string
	nacosAddr   string
	nacosEnable bool
	nacosDataId string
	nacosGroup  string
)

func init() {
	// 定义命令行参数 (在 Load 之前解析)
	flag.StringVar(&configFile, "config", "", "path to config file (e.g. ./config/config.yaml)")
	flag.StringVar(&nacosAddr, "nacos-addr", "", "nacos address (e.g. 127.0.0.1:8848)")
	flag.BoolVar(&nacosEnable, "nacos-enable", false, "enable nacos config")
	flag.StringVar(&nacosDataId, "nacos-data-id", "", "nacos data id")
	flag.StringVar(&nacosGroup, "nacos-group", "", "nacos group")
}

func Load() error {
	// 解析命令行参数
	if !flag.Parsed() {
		flag.Parse()
	}

	v := viper.New()

	// 1. 设置默认配置 (Layer 1: Default)
	SetDefaults(v)

	// 1.5. 环境变量覆盖 (Layer 1.5: Environment Variables)
	// 格式: DEVOPS_DB_HOST -> db.host
	LoadEnvOverrides(v)

	// 2. 加载本地配置 (Layer 2: Local File)
	// 优先级：命令行参数指定的文件 > 默认 ./config/config.yaml
	if configFile != "" {
		fmt.Printf("Loading config from file: %s\n", configFile)
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
	}

	if err := v.ReadInConfig(); err != nil {
		// 如果指定了配置文件但读取失败，直接报错
		if configFile != "" {
			return fmt.Errorf("failed to read specified config file: %w", err)
		}
		// 默认路径未找到则降级
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Warning: Local config file not found, using defaults.")
		} else {
			return fmt.Errorf("failed to read local config: %w", err)
		}
	} else {
		if configFile == "" {
			fmt.Println("Successfully loaded default local config file.")
		}
	}

	// 3. 命令行参数覆盖 (Layer 2.5: Command Line Args)
	// 允许通过命令行临时覆盖 Nacos 连接参数
	if nacosEnable {
		v.Set("nacos.enable", true)
	}
	if nacosAddr != "" {
		parts := strings.Split(nacosAddr, ":")
		if len(parts) == 2 {
			v.Set("nacos.host", parts[0])
			v.Set("nacos.port", parts[1])
		} else {
			v.Set("nacos.host", parts[0])
		}
	}
	if nacosDataId != "" {
		v.Set("nacos.data_id", nacosDataId)
	}
	if nacosGroup != "" {
		v.Set("nacos.group", nacosGroup)
	}

	// 4. 加载 Nacos 配置 (Layer 3: Remote Config Center)
	// 如果本地配置或命令行启用了 Nacos，则尝试从 Nacos 加载并覆盖当前配置
	if err := LoadFromNacos(v); err != nil {
		// 如果配置了 Nacos 但连接失败，建议报错阻断启动，避免配置不一致
		return fmt.Errorf("failed to load config from nacos: %w", err)
	}

	Cfg = v

	// 5. 配置验证
	if err := validateConfig(v); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	return nil
}

// validateConfig 校验关键配置项的合法性
func validateConfig(v *viper.Viper) error {
	if v.GetString("db.host") == "" {
		return fmt.Errorf("db.host 不能为空")
	}
	if v.GetInt("db.port") <= 0 || v.GetInt("db.port") > 65535 {
		return fmt.Errorf("db.port 必须在 1-65535 之间，当前值: %d", v.GetInt("db.port"))
	}
	if v.GetString("db.db") == "" {
		return fmt.Errorf("db.db（数据库名）不能为空")
	}

	if v.GetString("redis.addr") == "" {
		return fmt.Errorf("redis.addr 不能为空")
	}

	serverPort := v.GetInt("server.port")
	if serverPort <= 0 || serverPort > 65535 {
		return fmt.Errorf("server.port 必须在 1-65535 之间，当前值: %d", serverPort)
	}

	serverMode := v.GetString("server.mode")
	if serverMode != "debug" && serverMode != "release" && serverMode != "test" {
		return fmt.Errorf("server.mode 必须是 debug/release/test 之一，当前值: %s", serverMode)
	}

	logLevel := v.GetString("log.level")
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[logLevel] {
		return fmt.Errorf("log.level 必须是 debug/info/warn/error 之一，当前值: %s", logLevel)
	}

	logOutput := v.GetString("log.output")
	validOutputs := map[string]bool{"console": true, "file": true, "both": true}
	if !validOutputs[logOutput] {
		return fmt.Errorf("log.output 必须是 console/file/both 之一，当前值: %s", logOutput)
	}

	if (logOutput == "file" || logOutput == "both") && v.GetString("log.filePath") == "" {
		return fmt.Errorf("log.output 为 %s 时，log.filePath 不能为空", logOutput)
	}

	if v.GetInt("session.expire") <= 0 {
		return fmt.Errorf("session.expire 必须大于 0")
	}

	return nil
}
