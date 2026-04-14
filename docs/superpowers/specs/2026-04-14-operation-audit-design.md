# 操作审计模块设计

## 概述

在侧边栏新增独立的「操作审计」一级菜单，包含操作日志和登录日志两个子页面。操作日志复用已有的后端审计日志 API；登录日志为全新功能，需构建完整的后端模块（model/repo/service/api）并在登录流程中集成日志记录。

## 需求

### 操作日志

表格列：用户账号、请求方式、登录IP、请求URL、操作描述、操作时间、操作（详情按钮）

筛选条件：用户账号、请求方式、操作描述、时间范围

点击详情弹窗展示：请求参数、返回结果、错误信息、HTTP 状态码、响应耗时

### 登录日志

表格列：用户账号、登录IP、登录地点、浏览器、操作系统、登录状态、提示消息、访问时间

筛选条件：用户账号、登录状态（成功/失败）、时间范围

记录范围：同时记录登录成功和失败事件

## 技术决策

| 决策项 | 选择 | 原因 |
|--------|------|------|
| 侧边栏位置 | 独立一级菜单「操作审计」 | 审计功能独立于系统管理，更清晰的职责划分 |
| 登录日志后端位置 | 在 user 模块中扩展，与 audit_log 同级 | 复用已有模式和认证代码，改动最小 |
| IP 地理解析 | ip2region 离线库 | 无外部依赖，查询速度快，精度到城市级别 |
| UA 解析 | github.com/mssola/useragent | 轻量 Go 库，提取浏览器和操作系统信息 |

## 后端设计

### 登录日志 Model

文件：`backend/internal/modules/user/model/login_log.go`

```
LoginLog (表名: login_logs)
├── ID         uint            主键
├── Username   string(50)      用户账号, 索引
├── IP         string(50)      登录IP
├── Location   string(100)     登录地点 (ip2region 解析)
├── Browser    string(100)     浏览器 (UA 解析, 如 "Chrome 124")
├── OS         string(100)     操作系统 (UA 解析, 如 "Windows 11")
├── Status     string(20)      登录状态 (success/failed), 索引
├── Message    string(200)     提示消息 ("登录成功"/"密码错误"/"用户不存在" 等)
├── UserAgent  string(500)     原始 UA 字符串
├── LoginAt    time            登录时间, 索引
├── CreatedAt  time            创建时间
└── DeletedAt  gorm.DeletedAt  软删除
```

### 登录日志 Repository

文件：`backend/internal/modules/user/repository/login_log_repo.go`

方法：
- `Create(log *model.LoginLog) error` - 创建登录日志
- `List(query LoginLogQuery) ([]model.LoginLog, int64, error)` - 分页查询，支持筛选：Username, Status, StartAt/EndAt

### 登录日志 Service

文件：`backend/internal/modules/user/service/login_log_service.go`

方法：
- `CreateLoginLog(username, ip, userAgent, status, message string) error` - 创建日志（内部调用 ip2region 解析地理位置、UA 解析浏览器和 OS）
- `List(req LoginLogListRequest) ([]map[string]interface{}, int64, error)` - 分页查询

### 登录日志 API

文件：`backend/internal/modules/user/api/login_log.go`

端点：
- GET `/api/v1/login-log/list` - 分页查询登录日志

查询参数：username, status, startAt, endAt, page, pageSize

### 登录日志路由

文件：`backend/internal/routers/v1/login_log.go`

权限要求：`login-log:list`

### 登录日志捕获集成

修改 `backend/internal/modules/user/api/auth.go`：

在 `Login()` 函数中：
1. 登录失败时（err != nil）：异步记录 status="failed", message=err.Error()
2. 登录成功时（返回响应前）：异步记录 status="success", message="登录成功"

在 `OIDCLogin()` 函数中：同样模式记录 OIDC 登录事件。

异步写入：参考 audit 中间件的 channel+worker 模式，使用带缓冲的 channel 和 goroutine worker 写入，避免阻塞登录请求。

### IP 地理解析

使用 `github.com/lionsoul2014/ip2region` Go 版本：
- 将 ip2region.xdb 数据库文件打包到项目中（或通过配置指定路径）
- Service 层初始化时加载 xdb 到内存，提供 IP → 地区的查询方法

### UA 解析

使用 `github.com/mssola/useragent`：
- 解析 User-Agent 字符串提取浏览器名称+版本、操作系统名称+版本
- 示例：`Chrome 124`, `Windows 11`, `macOS 14`, `Firefox 126`

### 权限种子

修改 `backend/internal/bootstrap/db.go`，新增：
```go
{Name: "查看登录日志", Resource: "login-log", Action: "list", Description: "查看登录日志"}
```

### Auto-migration

修改 `backend/internal/bootstrap/db.go`，在 `AutoMigrate` 中添加 `&model.LoginLog{}`。

### 操作日志（已有 API - 无需后端改动）

现有 `/api/v1/audit/list` 端点已返回所有前端需要的字段：
- username (用户账号)
- method (请求方式)
- ip (登录IP)
- path (请求URL)
- operation (操作描述)
- requestAt (操作时间)
- params, result, errorMessage, status, latency (详情弹窗使用)

## 前端设计

### 侧边栏

修改 `frontend/src/components/Layout/MainLayout.vue`：

在「系统管理」el-sub-menu 之后新增：
```html
<el-sub-menu index="audit">
  <template #title>
    <el-icon><Notebook /></el-icon>
    <span>操作审计</span>
  </template>
  <el-menu-item index="/audit/operation">操作日志</el-menu-item>
  <el-menu-item index="/audit/login">登录日志</el-menu-item>
</el-sub-menu>
```

新增图标导入：`Notebook` from `@element-plus/icons-vue`

### 路由

修改 `frontend/src/router/index.js`，在 children 中新增：
```javascript
{
  path: 'audit/operation',
  component: () => import('../views/Audit/OperationLog.vue')
},
{
  path: 'audit/login',
  component: () => import('../views/Audit/LoginLog.vue')
}
```

### API 文件

新增 `frontend/src/api/audit.js`：
- `getAuditList(params)` → GET `/api/v1/audit/list`
- 查询参数：username, operation, resource, keyword, startAt, endAt, page, pageSize

新增 `frontend/src/api/loginLog.js`：
- `getLoginLogList(params)` → GET `/api/v1/login-log/list`
- 查询参数：username, status, startAt, endAt, page, pageSize

### 操作日志页面

新增 `frontend/src/views/Audit/OperationLog.vue`：

遵循项目已有的列表页模式，使用 `useTableList` composable。

筛选栏：
- 用户账号：el-input
- 请求方式：el-select (GET/POST/PUT/DELETE)
- 操作描述：el-input
- 时间范围：el-date-picker (type="datetimerange")

表格列：
- 用户账号 (username)
- 请求方式 (method) - 彩色 el-tag：GET=success, POST=warning, DELETE=danger, PUT=primary
- 登录IP (ip)
- 请求URL (path) - show-overflow-tooltip
- 操作描述 (operation)
- 操作时间 (requestAt) - dayjs 格式化
- 操作 - "详情" 文字按钮

详情弹窗 (el-dialog)：
- 请求参数 (params) - JSON 格式化展示
- 返回结果 (result)
- 错误信息 (errorMessage)
- HTTP 状态码 (status)
- 响应耗时 (latency) ms

### 登录日志页面

新增 `frontend/src/views/Audit/LoginLog.vue`：

筛选栏：
- 用户账号：el-input
- 登录状态：el-select (成功/失败)
- 时间范围：el-date-picker (type="datetimerange")

表格列：
- 用户账号 (username)
- 登录IP (ip)
- 登录地点 (location)
- 浏览器 (browser)
- 操作系统 (os)
- 登录状态 (status) - 成功=success tag, 失败=danger tag
- 提示消息 (message)
- 访问时间 (loginAt) - dayjs 格式化

## 文件变更清单

### 后端新增文件

| 文件 | 说明 |
|------|------|
| `backend/internal/modules/user/model/login_log.go` | LoginLog 模型 |
| `backend/internal/modules/user/repository/login_log_repo.go` | 登录日志仓储层 |
| `backend/internal/modules/user/service/login_log_service.go` | 登录日志服务层 |
| `backend/internal/modules/user/api/login_log.go` | 登录日志 API handler |
| `backend/internal/routers/v1/login_log.go` | 登录日志路由注册 |

### 后端修改文件

| 文件 | 变更 |
|------|------|
| `backend/internal/modules/user/api/auth.go` | Login/OIDCLogin 中集成登录日志记录 |
| `backend/internal/bootstrap/db.go` | 添加 LoginLog AutoMigrate + 权限种子 |
| `backend/routers/v1/v1.go` | 注册 login-log 路由 |
| `backend/go.mod` / `backend/go.sum` | 添加 ip2region、useragent 依赖 |

### 前端新增文件

| 文件 | 说明 |
|------|------|
| `frontend/src/api/audit.js` | 操作日志 API |
| `frontend/src/api/loginLog.js` | 登录日志 API |
| `frontend/src/views/Audit/OperationLog.vue` | 操作日志页面 |
| `frontend/src/views/Audit/LoginLog.vue` | 登录日志页面 |

### 前端修改文件

| 文件 | 变更 |
|------|------|
| `frontend/src/components/Layout/MainLayout.vue` | 侧边栏新增「操作审计」菜单 |
| `frontend/src/router/index.js` | 新增 audit/operation、audit/login 路由 |
