# K8s 网络管理页面（Service & Ingress）改造实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 改造 `/k8s/network` 页面，实现 Service/Ingress 的自动加载、完整列展示、CRUD + YAML 操作

**Architecture:** 后端扩展 ServiceVO/IngressVO 并新增 YAML 更新端点；前端重写 NetworkList.vue，复用通用 YAML 获取端点和 WorkloadList 的 YAML 弹窗模式

**Tech Stack:** Go (Gin, k8s client-go), Vue 3 (Composition API, Element Plus)

---

## Task 1: 扩展 ServiceVO + 更新 ListServices

**Files:**
- Modify: `backend/internal/modules/k8s/service/service.go`

- [ ] **Step 1: 更新 ServiceVO 结构体**

在 `backend/internal/modules/k8s/service/service.go` 中，修改 ServiceVO：

```go
type ServiceVO struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"clusterIP"`
	Ports     []string          `json:"ports"`
	Selector  map[string]string `json:"selector"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"createdAt"`
}
```

改动点：
- `Ports` 类型从 `[]int32` 改为 `[]string`
- 新增 `Selector` 字段

- [ ] **Step 2: 更新 ListServices 中的端口格式化和 Selector 填充**

在 `ListServices` 方法的 for 循环中，更新 VO 构建逻辑：

```go
ports := make([]string, 0, len(item.Spec.Ports))
for _, p := range item.Spec.Ports {
	portStr := fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, string(p.Protocol))
	if p.NodePort == 0 {
		portStr = fmt.Sprintf("%d/%s", p.Port, string(p.Protocol))
	}
	ports = append(ports, portStr)
}
result = append(result, ServiceVO{
	Name:      item.Name,
	Namespace: item.Namespace,
	Type:      string(item.Spec.Type),
	ClusterIP: item.Spec.ClusterIP,
	Ports:     ports,
	Selector:  item.Spec.Selector,
	Labels:    item.Labels,
	CreatedAt: item.CreationTimestamp.Time,
})
```

需要在文件顶部 import 中添加 `"fmt"`。

- [ ] **Step 3: 更新 GetServiceDetail 中同样的逻辑**

在 `GetServiceDetail` 方法中应用同样的端口格式化和 Selector 填充：

```go
ports := make([]string, 0, len(item.Spec.Ports))
for _, p := range item.Spec.Ports {
	portStr := fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, string(p.Protocol))
	if p.NodePort == 0 {
		portStr = fmt.Sprintf("%d/%s", p.Port, string(p.Protocol))
	}
	ports = append(ports, portStr)
}

return &ServiceVO{
	Name:      item.Name,
	Namespace: item.Namespace,
	Type:      string(item.Spec.Type),
	ClusterIP: item.Spec.ClusterIP,
	Ports:     ports,
	Selector:  item.Spec.Selector,
	Labels:    item.Labels,
	CreatedAt: item.CreationTimestamp.Time,
}, nil
```

- [ ] **Step 4: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功，无错误

- [ ] **Step 5: 提交**

```bash
git add backend/internal/modules/k8s/service/service.go
git commit -m "feat: extend ServiceVO with Selector field and rich port format"
```

---

## Task 2: 扩展 IngressVO + 更新 ListIngresses

**Files:**
- Modify: `backend/internal/modules/k8s/service/ingress.go`

- [ ] **Step 1: 更新 IngressVO 结构体**

```go
type IngressVO struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	IngressClass   string            `json:"ingressClass"`
	Hosts          []string          `json:"hosts"`
	Paths          []string          `json:"paths"`
	BackendService string            `json:"backendService"`
	BackendPort    string            `json:"backendPort"`
	Labels         map[string]string `json:"labels"`
	CreatedAt      time.Time         `json:"createdAt"`
}
```

- [ ] **Step 2: 新增辅助函数从 Ingress 提取详细信息**

在 `ingress.go` 中添加一个辅助函数：

```go
func extractIngressDetails(item networkingv1.Ingress) (ingressClass string, hosts []string, paths []string, backendService string, backendPort string) {
	if item.Spec.IngressClassName != nil {
		ingressClass = *item.Spec.IngressClassName
	}

	// 默认后端
	if item.Spec.DefaultBackend != nil {
		backendService = item.Spec.DefaultBackend.Service.Name
		if item.Spec.DefaultBackend.Service.Port.Name != "" {
			backendPort = item.Spec.DefaultBackend.Service.Port.Name
		} else {
			backendPort = fmt.Sprintf("%d", item.Spec.DefaultBackend.Service.Port.Number)
		}
	}

	for _, rule := range item.Spec.Rules {
		if rule.Host != "" {
			hosts = append(hosts, rule.Host)
		}
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				paths = append(paths, path.Path)
				if backendService == "" && path.Backend.Service != nil {
					backendService = path.Backend.Service.Name
					if path.Backend.Service.Port.Name != "" {
						backendPort = path.Backend.Service.Port.Name
					} else {
						backendPort = fmt.Sprintf("%d", path.Backend.Service.Port.Number)
					}
				}
			}
		}
	}
	return
}
```

需要在文件顶部 import 中添加 `"fmt"`。

- [ ] **Step 3: 更新 ListIngresses 中的 VO 构建逻辑**

在 `ListIngresses` 方法的 for 循环中（`allItems` 构建部分）：

```go
for _, item := range list.Items {
	ingressClass, hosts, paths, backendService, backendPort := extractIngressDetails(item)
	allItems = append(allItems, IngressVO{
		Name:           item.Name,
		Namespace:      item.Namespace,
		IngressClass:   ingressClass,
		Hosts:          hosts,
		Paths:          paths,
		BackendService: backendService,
		BackendPort:    backendPort,
		Labels:         item.Labels,
		CreatedAt:      item.CreationTimestamp.Time,
	})
}
```

- [ ] **Step 4: 更新 ingressVOFromUnstructured 函数**

这个函数处理旧版 K8s API 的 fallback 路径，也需要更新：

```go
func ingressVOFromUnstructured(item unstructured.Unstructured) IngressVO {
	hosts := make([]string, 0)
	paths := make([]string, 0)
	var backendService, backendPort, ingressClass string

	// 提取 ingressClassName
	if icn, found, _ := unstructured.NestedString(item.Object, "spec", "ingressClassName"); found {
		ingressClass = icn
	}

	// 提取默认后端
	if defBackend, found, _ := unstructured.NestedMap(item.Object, "spec", "defaultBackend"); found {
		if svc, found, _ := unstructured.NestedMap(defBackend, "service"); found {
			if name, found, _ := unstructured.NestedString(svc, "name"); found {
				backendService = name
			}
			if port, found, _ := unstructured.NestedMap(svc, "port"); found {
				if name, found, _ := unstructured.NestedString(port, "name"); found && name != "" {
					backendPort = name
				} else if num, found, _ := unstructured.NestedInt64(port, "number"); found {
					backendPort = fmt.Sprintf("%d", num)
				}
			}
		}
	}

	// 提取 rules
	if rules, found, _ := unstructured.NestedSlice(item.Object, "spec", "rules"); found {
		for _, rule := range rules {
			ruleMap, ok := rule.(map[string]interface{})
			if !ok {
				continue
			}
			host, ok := ruleMap["host"].(string)
			if ok && host != "" {
				hosts = append(hosts, host)
			}
			if http, ok := ruleMap["http"].(map[string]interface{}); ok {
				if httpPaths, ok := http["paths"].([]interface{}); ok {
					for _, p := range httpPaths {
						if pathMap, ok := p.(map[string]interface{}); ok {
							if path, ok := pathMap["path"].(string); ok && path != "" {
								paths = append(paths, path)
							}
							if backendService == "" {
								if backend, ok := pathMap["backend"].(map[string]interface{}); ok {
									if svc, ok := backend["service"].(map[string]interface{}); ok {
										if name, ok := svc["name"].(string); ok {
											backendService = name
										}
										if port, ok := svc["port"].(map[string]interface{}); ok {
											if name, ok := port["name"].(string); ok && name != "" {
												backendPort = name
											} else if num, ok := port["number"].(int64); ok {
												backendPort = fmt.Sprintf("%d", num)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return IngressVO{
		Name:           item.GetName(),
		Namespace:      item.GetNamespace(),
		IngressClass:   ingressClass,
		Hosts:          hosts,
		Paths:          paths,
		BackendService: backendService,
		BackendPort:    backendPort,
		Labels:         item.GetLabels(),
		CreatedAt:      item.GetCreationTimestamp().Time,
	}
}
```

- [ ] **Step 5: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功

- [ ] **Step 6: 提交**

```bash
git add backend/internal/modules/k8s/service/ingress.go
git commit -m "feat: extend IngressVO with ingressClass, paths, backendService, backendPort"
```

---

## Task 3: 新建 Service YAML 操作方法

**Files:**
- Create: `backend/internal/modules/k8s/service/service_ops.go`

- [ ] **Step 1: 创建 service_ops.go**

```go
package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func (s *K8sService) GetServiceObject(clusterName string, namespace, name string) (*corev1.Service, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) UpdateServiceByYAML(clusterName string, namespace, name, rawYAML string) (*corev1.Service, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	current, err := s.GetServiceObject(clusterName, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current service: %w", err)
	}

	var desired corev1.Service
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	// 校验不可变字段
	if desired.APIVersion != "" && desired.APIVersion != "v1" {
		return nil, fmt.Errorf("不允许修改 apiVersion")
	}
	if desired.Kind != "" && desired.Kind != "Service" {
		return nil, fmt.Errorf("不允许修改 kind")
	}
	if desired.Name != "" && desired.Name != name {
		return nil, fmt.Errorf("不允许修改 metadata.name")
	}
	if desired.Namespace != "" && desired.Namespace != namespace {
		return nil, fmt.Errorf("不允许修改 metadata.namespace")
	}

	desired.Namespace = namespace
	desired.Name = name
	desired.ResourceVersion = current.ResourceVersion
	desired.Status = corev1.ServiceStatus{}
	desired.ManagedFields = nil

	return s.UpdateService(clusterName, namespace, &desired)
}
```

- [ ] **Step 2: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add backend/internal/modules/k8s/service/service_ops.go
git commit -m "feat: add Service YAML update operations"
```

---

## Task 4: 新建 Ingress YAML 操作方法

**Files:**
- Create: `backend/internal/modules/k8s/service/ingress_ops.go`

- [ ] **Step 1: 创建 ingress_ops.go**

```go
package service

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func (s *K8sService) GetIngressObject(clusterName string, namespace, name string) (*networkingv1.Ingress, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	cluster, err := s.clusterService.GetByExactName(clusterName)
	if err != nil {
		return nil, err
	}
	client, err := s.clientFactory.GetClient(cluster)
	if err != nil {
		return nil, err
	}
	return client.NetworkingV1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (s *K8sService) UpdateIngressByYAML(clusterName string, namespace, name, rawYAML string) (*networkingv1.Ingress, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	current, err := s.GetIngressObject(clusterName, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current ingress: %w", err)
	}

	var desired networkingv1.Ingress
	if err := yaml.Unmarshal([]byte(rawYAML), &desired); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	// 校验不可变字段
	if desired.APIVersion != "" && desired.APIVersion != "networking.k8s.io/v1" {
		return nil, fmt.Errorf("不允许修改 apiVersion")
	}
	if desired.Kind != "" && desired.Kind != "Ingress" {
		return nil, fmt.Errorf("不允许修改 kind")
	}
	if desired.Name != "" && desired.Name != name {
		return nil, fmt.Errorf("不允许修改 metadata.name")
	}
	if desired.Namespace != "" && desired.Namespace != namespace {
		return nil, fmt.Errorf("不允许修改 metadata.namespace")
	}

	desired.Namespace = namespace
	desired.Name = name
	desired.ResourceVersion = current.ResourceVersion
	desired.Status = networkingv1.IngressStatus{}
	desired.ManagedFields = nil

	return s.UpdateIngress(clusterName, namespace, &desired)
}
```

- [ ] **Step 2: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add backend/internal/modules/k8s/service/ingress_ops.go
git commit -m "feat: add Ingress YAML update operations"
```

---

## Task 5: 新增 API handlers + 路由注册

**Files:**
- Modify: `backend/internal/modules/k8s/api/service.go`
- Modify: `backend/internal/modules/k8s/api/ingress.go`
- Modify: `backend/routers/v1/k8s.go`

- [ ] **Step 1: 在 service.go 中添加 UpdateServiceYAML handler**

在 `backend/internal/modules/k8s/api/service.go` 文件末尾添加：

```go
// UpdateServiceYAML godoc
// @Summary 通过 YAML 更新 Service
// @Description 使用 YAML 内容更新指定 Service
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name, yaml}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/service/yaml/update [post]
func UpdateServiceYAML(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
		YAML        string `json:"yaml" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	clusterName := req.ClusterName
	if clusterName == "" {
		var err error
		clusterName, err = resolveClusterName(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := svc.UpdateServiceByYAML(clusterName, req.Namespace, req.Name, req.YAML)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}
```

- [ ] **Step 2: 在 ingress.go 中添加 UpdateIngressYAML handler**

在 `backend/internal/modules/k8s/api/ingress.go` 文件末尾添加：

```go
// UpdateIngressYAML godoc
// @Summary 通过 YAML 更新 Ingress
// @Description 使用 YAML 内容更新指定 Ingress
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name, yaml}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/ingress/yaml/update [post]
func UpdateIngressYAML(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
		YAML        string `json:"yaml" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	clusterName := req.ClusterName
	if clusterName == "" {
		var err error
		clusterName, err = resolveClusterName(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.UpdateIngressByYAML(clusterName, req.Namespace, req.Name, req.YAML)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}
```

- [ ] **Step 3: 在 k8s.go 中注册新路由**

在 `backend/routers/v1/k8s.go` 中，找到 Service 路由组（约 199 行 `api.DeleteService` 之后），添加：

```go
g.POST("/service/yaml/update",
	updatePermission,
	middleware.SetAuditOperation("YAML更新Service"),
	api.UpdateServiceYAML)
```

找到 Ingress 路由组（约 231 行 `api.DeleteIngress` 之后），添加：

```go
g.POST("/ingress/yaml/update",
	updatePermission,
	middleware.SetAuditOperation("YAML更新Ingress"),
	api.UpdateIngressYAML)
```

- [ ] **Step 4: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译成功

- [ ] **Step 5: 提交**

```bash
git add backend/internal/modules/k8s/api/service.go backend/internal/modules/k8s/api/ingress.go backend/routers/v1/k8s.go
git commit -m "feat: add Service/Ingress YAML update API handlers and routes"
```

---

## Task 6: 更新前端 API 文件

**Files:**
- Modify: `frontend/src/api/service.js`
- Modify: `frontend/src/api/ingress.js`

- [ ] **Step 1: 更新 service.js**

将 `frontend/src/api/service.js` 内容替换为：

```js
import request from './request'

export const getServiceList = (params) => request.get('/k8s/service/list', { params })

export const createService = (data) => request.post('/k8s/service/create', data)

export const deleteService = (data) => request.post('/k8s/service/delete', data)

export const getServiceYAML = (params) => request.get('/k8s/resource/yaml', { params })

export const updateServiceByYAML = (data) => request.post('/k8s/service/yaml/update', data)
```

- [ ] **Step 2: 更新 ingress.js**

将 `frontend/src/api/ingress.js` 内容替换为：

```js
import request from './request'

export const getIngressList = (params) => request.get('/k8s/ingress/list', { params })

export const createIngress = (data) => request.post('/k8s/ingress/create', data)

export const deleteIngress = (data) => request.post('/k8s/ingress/delete', data)

export const getIngressYAML = (params) => request.get('/k8s/resource/yaml', { params })

export const updateIngressByYAML = (data) => request.post('/k8s/ingress/yaml/update', data)
```

- [ ] **Step 3: 提交**

```bash
git add frontend/src/api/service.js frontend/src/api/ingress.js
git commit -m "feat: add Service/Ingress YAML API functions"
```

---

## Task 7: 重写 NetworkList.vue

**Files:**
- Modify: `frontend/src/views/k8s/NetworkList.vue`

- [ ] **Step 1: 重写 NetworkList.vue**

将 `frontend/src/views/k8s/NetworkList.vue` 内容替换为：

```vue
<template>
  <div class="page-container">
    <div class="search-bar">
      <ClusterSelector v-model="clusterName" />
      <NamespaceSelector v-model="namespace" :cluster-name="clusterName" style="margin-left: 12px" />
      <el-input v-model="keyword" placeholder="搜索" style="width: 200px; margin-left: 12px" clearable />
      <el-button type="primary" @click="fetchData" style="margin-left: 12px">查询</el-button>
      <el-button type="success" @click="handleCreate" style="margin-left: 12px">创建</el-button>
    </div>

    <el-tabs v-model="activeTab" @tab-change="handleTabChange" style="margin-top: 16px">
      <el-tab-pane label="Service" name="service" />
      <el-tab-pane label="Ingress" name="ingress" />
    </el-tabs>

    <!-- Service 表格 -->
    <el-table v-if="activeTab === 'service'" :data="tableData" stripe>
      <el-table-column label="名称" min-width="160">
        <template #default="{ row }">
          <el-link type="primary">{{ row.name }}</el-link>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="120" />
      <el-table-column prop="clusterIP" label="ClusterIP" width="140" />
      <el-table-column label="端口" min-width="180">
        <template #default="{ row }">{{ (row.ports || []).join(', ') }}</template>
      </el-table-column>
      <el-table-column label="选择器" min-width="160">
        <template #default="{ row }">
          <template v-if="row.selector && Object.keys(row.selector).length">
            <el-tag v-for="(value, key) in row.selector" :key="key" size="small" style="margin: 2px">{{ key }}={{ value }}</el-tag>
          </template>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleYaml(row)">YAML</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Ingress 表格 -->
    <el-table v-if="activeTab === 'ingress'" :data="tableData" stripe>
      <el-table-column label="名称" min-width="160">
        <template #default="{ row }">
          <el-link type="primary">{{ row.name }}</el-link>
        </template>
      </el-table-column>
      <el-table-column prop="ingressClass" label="类型" width="120">
        <template #default="{ row }">{{ row.ingressClass || '-' }}</template>
      </el-table-column>
      <el-table-column label="主机" min-width="180">
        <template #default="{ row }">{{ (row.hosts || []).join(', ') || '-' }}</template>
      </el-table-column>
      <el-table-column label="路径" width="120">
        <template #default="{ row }">{{ (row.paths || []).join(', ') || '-' }}</template>
      </el-table-column>
      <el-table-column prop="backendService" label="后端Service" width="140">
        <template #default="{ row }">{{ row.backendService || '-' }}</template>
      </el-table-column>
      <el-table-column prop="backendPort" label="端口" width="100">
        <template #default="{ row }">{{ row.backendPort || '-' }}</template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleYaml(row)">YAML</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      @current-change="fetchData"
      style="margin-top: 16px; justify-content: flex-end"
    />

    <!-- YAML 编辑弹窗 -->
    <el-dialog v-model="yamlVisible" :title="yamlTitle" width="900px" destroy-on-close top="3vh">
      <YamlEditor ref="yamlEditorRef" v-model="yamlContent" :readonly="yamlReadonly" :show-copy="false" min-height="600px" />
      <template #footer>
        <div style="display: flex; justify-content: space-between; width: 100%">
          <el-button @click="handleYamlCopy">{{ yamlCopyText }}</el-button>
          <div>
            <el-button @click="yamlVisible = false">取消</el-button>
            <el-button type="primary" @click="handleYamlSave" v-if="!yamlReadonly">保存</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ClusterSelector from '@/components/K8s/ClusterSelector.vue'
import NamespaceSelector from '@/components/K8s/NamespaceSelector.vue'
import YamlEditor from '@/components/K8s/YamlEditor.vue'
import { getServiceList, createService, deleteService, getServiceYAML, updateServiceByYAML } from '@/api/service'
import { getIngressList, createIngress, deleteIngress, getIngressYAML, updateIngressByYAML } from '@/api/ingress'
import { formatTime } from '@/utils/format'

const clusterName = ref('')
const namespace = ref('')
const keyword = ref('')
const activeTab = ref('service')
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

// YAML 弹窗
const yamlVisible = ref(false)
const yamlTitle = ref('YAML')
const yamlContent = ref('')
const yamlReadonly = ref(false)
const yamlMode = ref('')
const yamlEditorRef = ref(null)
const yamlCopyText = ref('复制')
const currentRow = ref(null)

// 自动加载：集群变化时触发
watch(clusterName, (val) => {
  if (val) {
    page.value = 1
    fetchData()
  }
})

// 自动加载：命名空间变化时触发
watch(namespace, () => {
  if (clusterName.value) {
    page.value = 1
    fetchData()
  }
})

// 页面加载时若集群已就绪则自动获取
onMounted(() => {
  if (clusterName.value) fetchData()
})

const fetchData = async () => {
  const params = {
    clusterName: clusterName.value,
    namespace: namespace.value,
    keyword: keyword.value,
    page: page.value,
    pageSize: pageSize.value
  }
  const res = activeTab.value === 'service'
    ? await getServiceList(params)
    : await getIngressList(params)
  tableData.value = res.data?.items || res.data || []
  total.value = res.data?.total || res.total || 0
}

const handleTabChange = () => {
  page.value = 1
  fetchData()
}

// 删除
const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除?', '提示', { type: 'warning' })
  const data = { clusterName: clusterName.value, namespace: row.namespace, name: row.name }
  activeTab.value === 'service' ? await deleteService(data) : await deleteIngress(data)
  ElMessage.success('删除成功')
  fetchData()
}

// YAML 复制
const handleYamlCopy = async () => {
  try {
    await navigator.clipboard.writeText(yamlContent.value)
    yamlCopyText.value = '已复制'
    setTimeout(() => { yamlCopyText.value = '复制' }, 2000)
  } catch {
    ElMessage.error('复制失败')
  }
}

// 打开 YAML 编辑
const handleYaml = async (row) => {
  currentRow.value = row
  yamlMode.value = 'edit'
  yamlReadonly.value = false
  yamlTitle.value = `YAML - ${row.name}`
  try {
    const res = await (activeTab.value === 'service'
      ? getServiceYAML({ resourceType: 'service', clusterName: clusterName.value, namespace: row.namespace, name: row.name })
      : getIngressYAML({ resourceType: 'ingress', clusterName: clusterName.value, namespace: row.namespace, name: row.name }))
    yamlContent.value = res.data?.yaml || ''
  } catch {
    yamlContent.value = ''
  }
  yamlVisible.value = true
}

// 创建
const handleCreate = () => {
  currentRow.value = { namespace: namespace.value }
  yamlMode.value = 'create'
  yamlReadonly.value = false
  yamlTitle.value = `新建 ${activeTab.value === 'service' ? 'Service' : 'Ingress'}`
  yamlContent.value = ''
  yamlVisible.value = true
}

// YAML 保存
const handleYamlSave = async () => {
  if (!yamlContent.value.trim()) {
    ElMessage.warning('YAML 内容不能为空')
    return
  }

  if (yamlMode.value === 'create') {
    const ns = currentRow.value.namespace || namespace.value || 'default'
    if (activeTab.value === 'service') {
      await createService({ yaml: yamlContent.value, clusterName: clusterName.value, namespace: ns })
    } else {
      await createIngress({ yaml: yamlContent.value, clusterName: clusterName.value, namespace: ns })
    }
    ElMessage.success('创建成功')
  } else {
    const data = {
      clusterName: clusterName.value,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      yaml: yamlContent.value
    }
    if (activeTab.value === 'service') {
      await updateServiceByYAML(data)
    } else {
      await updateIngressByYAML(data)
    }
    ElMessage.success('保存成功')
  }
  yamlVisible.value = false
  fetchData()
}
</script>

<style scoped>
.page-container {
  padding: 20px;
}
.search-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}
</style>
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/views/k8s/NetworkList.vue
git commit -m "feat: rewrite NetworkList.vue with auto-load, rich columns, and YAML CRUD"
```

---

## Task 8: 集成验证

- [ ] **Step 1: 后端编译**

Run: `cd backend && go build ./...`
Expected: 编译成功

- [ ] **Step 2: 启动后端服务**

Run: `cd backend && go run cmd/main.go`（或项目的启动命令）

- [ ] **Step 3: 启动前端开发服务器**

Run: `cd frontend && npm run dev`

- [ ] **Step 4: 浏览器验证**

打开 `http://localhost:3000/k8s/network`，验证：

1. 选择集群后数据自动加载
2. Service 表格展示：名称、类型、ClusterIP、端口（新格式）、选择器、创建时间、操作
3. 切换到 Ingress Tab，表格展示：名称、类型、主机、路径、后端Service、端口、创建时间、操作
4. 点击"创建"按钮 → 空 YAML 编辑器弹出 → 输入 YAML 保存
5. 点击行内"YAML"按钮 → 当前资源 YAML 加载到编辑器 → 修改保存
6. 点击"删除"按钮 → 确认弹窗 → 删除成功

- [ ] **Step 5: 提交最终状态（如有修复）**

```bash
git add -A
git commit -m "fix: integration fixes for network management page"
```
