package v1

import (
	"devops-platform/internal/modules/sqlaudit/api"

	"github.com/gin-gonic/gin"
)

func registerSqlAuditRoutes(r *gin.RouterGroup) {
	api.InstallSqlAuditRoutes(r)
}
