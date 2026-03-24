package bootstrap

import (
	"devops-platform/config"
	"devops-platform/routers"
)

func InitRouter() {
	r := routers.InitRouter()
	port := config.Cfg.GetString("server.port")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
