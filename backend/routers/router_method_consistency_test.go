package routers

import (
	"net/http"
	"testing"
)

func TestReadRoutesUseGET(t *testing.T) {
	router := setupRouterForSwaggerTest()

	routeMethods := make(map[string]map[string]bool)
	for _, route := range router.Routes() {
		methodBucket := routeMethods[route.Path]
		if methodBucket == nil {
			methodBucket = make(map[string]bool)
			routeMethods[route.Path] = methodBucket
		}
		methodBucket[route.Method] = true
	}

	readPaths := []string{
		"/api/v1/alert/rules",
		"/api/v1/alert/history",
		"/api/v1/alert/silences",
		"/api/v1/alert/channels",
		"/api/v1/alert/config",
		"/api/v1/log/search",
		"/api/v1/monitor/query",
		"/api/v1/monitor/config",
		"/api/v1/harbor/list",
		"/api/v1/harbor/images",
		"/api/v1/harbor/config",
		"/api/v1/cicd/list",
		"/api/v1/cicd/logs",
		"/api/v1/cicd/templates",
		"/api/v1/cicd/runs",
		"/api/v1/cicd/config",
		"/api/v1/app/list",
		"/api/v1/app/template/list",
		"/api/v1/app/deployment/list",
		"/api/v1/app/version/list",
		"/api/v1/app/topology",
	}

	for _, path := range readPaths {
		methodBucket := routeMethods[path]
		if methodBucket == nil {
			t.Fatalf("missing route path %s", path)
		}
		if !methodBucket[http.MethodGet] {
			t.Fatalf("read route %s should register GET", path)
		}
		if methodBucket[http.MethodPost] {
			t.Fatalf("read route %s should not register POST", path)
		}
	}
}

func TestResourceListLegacyPOSTRemoved(t *testing.T) {
	router := setupRouterForSwaggerTest()

	for _, route := range router.Routes() {
		if route.Path == "/api/v1/resource/list" && route.Method == http.MethodPost {
			t.Fatalf("legacy POST route should be removed: %s %s", route.Method, route.Path)
		}
	}
}
