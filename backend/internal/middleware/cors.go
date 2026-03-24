package middleware

import (
	"net/http"
	"strings"

	"devops-platform/config"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	allowOrigins := getStringSlice("cors.allow_origins")
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"*"}
	}

	allowMethods := getStringSlice("cors.allow_methods")
	if len(allowMethods) == 0 {
		allowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}

	allowHeaders := getStringSlice("cors.allow_headers")
	if len(allowHeaders) == 0 {
		allowHeaders = []string{"Content-Type", "Authorization", "X-Token"}
	}

	exposeHeaders := getStringSlice("cors.expose_headers")
	if len(exposeHeaders) == 0 {
		exposeHeaders = []string{"Set-Cookie"}
	}

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin != "" {
			if allowedOrigin(origin, allowOrigins) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Methods", strings.Join(allowMethods, ", "))
				c.Header("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
				c.Header("Access-Control-Expose-Headers", strings.Join(exposeHeaders, ", "))
				c.Header("Access-Control-Max-Age", "86400")
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func getStringSlice(key string) []string {
	if config.Cfg == nil {
		return nil
	}
	return config.Cfg.GetStringSlice(key)
}

func allowedOrigin(origin string, allowOrigins []string) bool {
	for _, ao := range allowOrigins {
		ao = strings.TrimSpace(ao)
		if ao == "" {
			continue
		}
		if ao == "*" {
			return true
		}
		if strings.EqualFold(ao, origin) {
			return true
		}
	}
	return false
}
