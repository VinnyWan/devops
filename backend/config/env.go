package config

import (
	"strings"

	"github.com/spf13/viper"
)

// LoadEnvOverrides 加载环境变量覆盖配置
// 环境变量格式: DEVOPS_DB_HOST -> db.host
// 规则: 前缀 DEVOPS_，下划线分隔层级，全部大写
func LoadEnvOverrides(v *viper.Viper) {
	v.SetEnvPrefix("DEVOPS")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}
