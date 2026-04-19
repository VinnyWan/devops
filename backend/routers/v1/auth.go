package v1

import (
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(v1 *gin.RouterGroup) {
	auth := v1.Group("/auth")
	mw := authMiddlewares()
	{
		auth.POST("/login", api.Login)
		auth.POST("/register", api.Register)
		auth.POST("/logout", append(mw, api.Logout)...)
		auth.GET("/permissions", append(mw, api.GetAllPermissions)...)
		auth.GET("/user-info", append(mw, api.GetUserInfo)...)
		auth.POST("/user-permissions", append(mw, api.GetUserPermissions)...)
		auth.POST("/change-password", append(mw, api.ChangePassword)...)

		// OIDC
		auth.GET("/oidc/login", api.OIDCLogin)
		auth.GET("/oidc/callback", api.OIDCCallback)
	}
}
