package bootstrap

import (
	"devops-platform/config"
)

func InitConfig() error {
	return config.Load()
}
