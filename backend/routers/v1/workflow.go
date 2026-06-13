package v1

import (
	"devops-platform/internal/modules/workflow/api"

	"github.com/gin-gonic/gin"
)

func registerWorkflowRoutes(r *gin.RouterGroup) {
	api.InstallWorkflowRoutes(r)
}
