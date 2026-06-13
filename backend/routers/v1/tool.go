package v1

import (
	"devops-platform/internal/modules/tool/api"

	"github.com/gin-gonic/gin"
)

func registerToolRoutes(r *gin.RouterGroup) {
	api.InstallToolRoutes(r)
}
