package v1

import (
	"devops-platform/internal/modules/task/api"

	"github.com/gin-gonic/gin"
)

func registerTaskRoutes(r *gin.RouterGroup) {
	api.InstallTaskRoutes(r)
}
