package config

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
)

// LoadFromNacos 从 Nacos 加载配置并合并到 Viper
func LoadFromNacos(v *viper.Viper) error {
	// 1. 检查开关
	if !v.GetBool("nacos.enable") {
		return nil
	}

	// 2. 读取 Nacos 连接参数
	dataId := v.GetString("nacos.data_id")
	group := v.GetString("nacos.group")
	sc, err := parseNacosServerConfigs(v)
	if err != nil {
		return err
	}
	cc := buildNacosClientConfig(v)

	// 5. 创建 Config Client
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return fmt.Errorf("create nacos config client failed: %w", err)
	}

	// 6. 获取配置
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		return fmt.Errorf("get config from nacos failed: %w", err)
	}

	// DEBUG: 打印获取到的配置内容摘要
	if len(content) > 0 {
		preview := content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("Nacos Config Content Preview:\n%s\n", preview)
	} else {
		fmt.Println("Warning: Nacos config content is empty!")
	}

	// 7. 合并配置 (Viper MergeConfig 会覆盖已有的 key)
	if err := v.MergeConfig(strings.NewReader(content)); err != nil {
		return fmt.Errorf("merge nacos config failed: %w", err)
	}

	fmt.Printf("Successfully loaded config from Nacos (DataId: %s, Group: %s)\n", dataId, group)

	// DEBUG: 打印合并后的关键配置
	fmt.Printf("Merged Config - DB Host: %s, Port: %s\n", v.GetString("db.host"), v.GetString("db.port"))
	fmt.Printf("Merged Config - Redis Addr: %s\n", v.GetString("redis.addr"))

	// 8. 监听配置变更 (动态刷新)
	err = client.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Printf("Config changed in Nacos: group=%s, dataId=%s\n", group, dataId)
			// 再次合并配置
			if err := v.MergeConfig(bytes.NewBufferString(data)); err != nil {
				fmt.Printf("Failed to merge changed config: %v\n", err)
				return
			}
			// 可以在这里触发其他回调，或依赖 Viper 的 OnConfigChange
		},
	})

	if err != nil {
		fmt.Printf("Listen nacos config failed: %v\n", err)
		// 监听失败不应阻断启动
	}

	return nil
}
