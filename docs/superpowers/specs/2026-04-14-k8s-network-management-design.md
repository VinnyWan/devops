# K8s 网络管理页面（Service & Ingress）改造设计

## 背景

`/k8s/network` 页面（NetworkList.vue）当前存在以下问题：
1. 进入页面后数据不会自动加载，必须手动点击"查询"按钮
2. Service 表格缺少"选择器(Selector)"列
3. 操作列只有"删除"，缺少创建、YAML 查看/编辑功能
4. Ingress 表格信息过于简单，缺少类型、路径、后端 Service 等关键列

后端已有 Service 和 Ingress 的完整 CRUD（list/detail/create/update/delete），且存在通用的 `/k8s/resource/yaml` GET 端点可获取任意资源的 YAML。

## 方案

采用方案 A：复用通用端点 + 最小后端改动。
- YAML 获取复用已有通用端点 `GET /k8s/resource/yaml`
- 新增 2 个 YAML 更新端点（与 deployment 模式一致）
- 前端参照 WorkloadList 的 YAML 弹窗模式

## 后端改动

### ServiceVO 扩展

文件：`backend/internal/modules/k8s/service/service.go`

```go
type ServiceVO struct {
    Name      string            `json:"name"`
    Namespace string            `json:"namespace"`
    Type      string            `json:"type"`
    ClusterIP string            `json:"clusterIP"`
    Ports     []string          `json:"ports"`     // 从 []int32 改为 []string，格式 "80:31234/TCP"
    Selector  map[string]string `json:"selector"`  // 新增
    Labels    map[string]string `json:"labels"`
    CreatedAt time.Time         `json:"createdAt"`
}
```

`ListServices` 方法需更新端口格式化逻辑和 Selector 字段填充。

### IngressVO 扩展

文件：`backend/internal/modules/k8s/service/ingress.go`

```go
type IngressVO struct {
    Name           string            `json:"name"`
    Namespace      string            `json:"namespace"`
    IngressClass   string            `json:"ingressClass"`    // 新增
    Hosts          []string          `json:"hosts"`
    Paths          []string          `json:"paths"`           // 新增
    BackendService string            `json:"backendService"`  // 新增
    BackendPort    string            `json:"backendPort"`     // 新增
    Labels         map[string]string `json:"labels"`
    CreatedAt      time.Time         `json:"createdAt"`
}
```

`ListIngresses` 方法需从 Ingress Spec 中提取 IngressClassName、Rules 中的 Paths、默认后端 Service 名称和端口。`ingressVOFromUnstructured` 同步更新。

### 新增 YAML 操作端点

新增文件：
- `backend/internal/modules/k8s/service/service_ops.go` — `GetServiceObject`、`UpdateServiceByYAML`
- `backend/internal/modules/k8s/service/ingress_ops.go` — `GetIngressObject`、`UpdateIngressByYAML`

`UpdateServiceByYAML` 模式（参照 `deployment_ops.go`）：
1. 获取当前 Service 对象
2. YAML Unmarshal 到 `corev1.Service`
3. 校验不可变字段（apiVersion、kind、name、namespace）
4. 继承 ResourceVersion，清空 Status 和 ManagedFields
5. 调用现有 `UpdateService` 方法

`UpdateIngressByYAML` 同理，操作 `networkingv1.Ingress`。

新增 API handler：
- `backend/internal/modules/k8s/api/service.go` — `UpdateServiceYAML` handler
- `backend/internal/modules/k8s/api/ingress.go` — `UpdateIngressYAML` handler

路由注册（`backend/routers/v1/k8s.go`）：
```go
// Service YAML 更新
g.POST("/service/yaml/update", updatePermission, middleware.SetAuditOperation("YAML更新Service"), api.UpdateServiceYAML)
// Ingress YAML 更新
g.POST("/ingress/yaml/update", updatePermission, middleware.SetAuditOperation("YAML更新Ingress"), api.UpdateIngressYAML)
```

## 前端改动

### API 文件更新

`frontend/src/api/service.js`：
```js
import request from './request'
export const getServiceList = (params) => request.get('/k8s/service/list', { params })
export const createService = (data) => request.post('/k8s/service/create', data)
export const deleteService = (data) => request.post('/k8s/service/delete', data)
export const getServiceYAML = (params) => request.get('/k8s/resource/yaml', { params })  // 新增
export const updateServiceByYAML = (data) => request.post('/k8s/service/yaml/update', data)  // 新增
```

`frontend/src/api/ingress.js`：
```js
import request from './request'
export const getIngressList = (params) => request.get('/k8s/ingress/list', { params })
export const createIngress = (data) => request.post('/k8s/ingress/create', data)
export const deleteIngress = (data) => request.post('/k8s/ingress/delete', data)
export const getIngressYAML = (params) => request.get('/k8s/resource/yaml', { params })  // 新增
export const updateIngressByYAML = (data) => request.post('/k8s/ingress/yaml/update', data)  // 新增
```

### NetworkList.vue 重写

整体结构：
```
.page-container
  .search-bar  (ClusterSelector + NamespaceSelector + 搜索框 + 查询按钮 + "创建"按钮)
  el-tabs      (Service | Ingress)
  el-table     (列根据 activeTab 条件渲染)
  el-pagination
  el-dialog    (YAML 编辑弹窗)
```

自动加载机制：
- `watch(clusterName)` → 重置 page 并 fetchData
- `watch(namespace)` → 重置 page 并 fetchData
- ClusterSelector 自动选中第一个集群时触发 watch → 自动加载数据

Service 表格列：
| 列 | 字段 | 宽度 | 说明 |
|---|---|---|---|
| 名称 | name | - | 蓝色链接样式 |
| 类型 | type | 120 | ClusterIP/NodePort/LoadBalancer |
| ClusterIP | clusterIP | 140 | |
| 端口 | ports | 180 | "80:31234/TCP" 格式，多个逗号分隔 |
| 选择器 | selector | - | el-tag 标签样式，多个 key=value |
| 创建时间 | createdAt | 180 | formatTime 格式化 |
| 操作 | - | 150 fixed=right | YAML / 删除 |

Ingress 表格列：
| 列 | 字段 | 宽度 | 说明 |
|---|---|---|---|
| 名称 | name | - | 蓝色链接样式 |
| 类型 | ingressClass | 120 | IngressClass 名称 |
| 主机 | hosts | 180 | 多个 host 逗号分隔 |
| 路径 | paths | 120 | /api, / 等 |
| 后端Service | backendService | 140 | |
| 端口 | backendPort | 100 | |
| 创建时间 | createdAt | 180 | formatTime 格式化 |
| 操作 | - | 150 fixed=right | YAML / 删除 |

YAML 弹窗（与 WorkloadList 模式一致）：
- `yamlMode`：`'create'` 或 `'edit'`
- `yamlVisible`：控制弹窗显示
- `yamlContent`：YAML 内容
- 创建模式：空编辑器，保存时调用 `createService`/`createIngress`（传 YAML + clusterName + namespace）
- 编辑模式：通过 `/k8s/resource/yaml` 获取当前 YAML，保存时调用 `updateServiceByYAML`/`updateIngressByYAML`
- 弹窗底部：复制按钮 + 取消 + 保存

操作逻辑：
- 顶部"创建"按钮 → yamlMode='create'，打开空 YAML 编辑器
- 行内"YAML"按钮 → yamlMode='edit'，获取当前 YAML 并打开编辑器
- 行内"删除"按钮 → ElMessageBox 确认 → 调用删除 API → 刷新列表

## 改动文件清单

### 后端（Go）
1. `backend/internal/modules/k8s/service/service.go` — ServiceVO 扩展 + ListServices 更新
2. `backend/internal/modules/k8s/service/ingress.go` — IngressVO 扩展 + ListIngresses 更新 + ingressVOFromUnstructured 更新
3. `backend/internal/modules/k8s/service/service_ops.go` — 新建，YAML 操作方法
4. `backend/internal/modules/k8s/service/ingress_ops.go` — 新建，YAML 操作方法
5. `backend/internal/modules/k8s/api/service.go` — 新增 UpdateServiceYAML handler
6. `backend/internal/modules/k8s/api/ingress.go` — 新增 UpdateIngressYAML handler
7. `backend/routers/v1/k8s.go` — 注册 2 条新路由

### 前端（Vue）
1. `frontend/src/api/service.js` — 新增 2 个 API 函数
2. `frontend/src/api/ingress.js` — 新增 2 个 API 函数
3. `frontend/src/views/k8s/NetworkList.vue` — 重写
