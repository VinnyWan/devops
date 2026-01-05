package routers

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {
	r := gin.New() // 不用 gin.Default()

	// 在这里注册路由
	// RegisterStatics(r)
	// apiRoutes(r)

	return r
}
