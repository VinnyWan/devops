# Backend

DevOps 平台后端服务，基于 Go + Gin 框架构建，提供 Kubernetes 集群管理、用户认证、权限控制等核心功能。

## 技术栈

- **语言**: Go 1.24.0
- **Web 框架**: Gin 1.10.0
- **数据库**: PostgreSQL + Redis
- **ORM**: GORM
- **认证**: JWT (golang-jwt/jwt/v5)
- **权限**: Casbin
- **配置管理**: Viper
- **服务发现**: Nacos
- **日志**: Zap
- **API 文档**: Swagger (Swaggo)
- **WebSocket**: Gorilla WebSocket

## 项目结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
│
├── config/                       # 配置管理
│   ├── config.go                 # 配置结构定义
│   ├── config.yaml               # 配置文件
│   ├── config.yaml.example       # 配置示例
│   ├── defaults.go               # 默认配置
│   ├── env.go                    # 环境变量加载
│   ├── env_override_test.go      # 环境变量测试
│   ├── nacos.go                  # Nacos 配置中心
│   └── nacos_registry.go         # Nacos 服务注册
│
├── internal/                     # 内部应用代码
│   ├── bootstrap/                # 初始化引导
│   │   ├── config.go             # 配置初始化
│   │   ├── db.go                 # 数据库初始化
│   │   ├── redis.go              # Redis 初始化
│   │   ├── casbin.go             # 权限引擎初始化
│   │   ├── k8s.go                # Kubernetes 客户端初始化
│   │   └── router.go             # 路由初始化
│   │
│   ├── middleware/               # 中间件
│   │   ├── permission.go         # 权限检查中间件
│   │   ├── cors.go               # 跨域中间件
│   │   ├── audit.go              # 审计日志中间件
│   │   ├── casbin.go             # Casbin 权限中间件
│   │   ├── session.go            # 会话中间件
│   │   ├── request_context.go    # 请求上下文中间件
│   │   ├── charset.go             # 字符集中间件
│   │   ├── recover.go            # 异常恢复中间件
│   │   ├── request_log.go        # 请求日志中间件
│   │   ├── mask_sensitive_test.go
│   │   └── audit_test.go
│   │
│   ├── modules/                  # 业务模块（DDD 架构）
│   │   │
│   │   ├── user/                 # 用户管理模块
│   │   │   ├── api/              # HTTP 处理器
│   │   │   ├── service/          # 业务逻辑层
│   │   │   ├── repository/       # 数据访问层
│   │   │   └── model/            # 数据模型
│   │   │
│   │   ├── alert/                # 告警管理模块
│   │   │   ├── api/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   └── model/
│   │   │
│   │   ├── app/                  # 应用管理模块
│   │   │   ├── api/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   └── model/
│   │   │
│   │   ├── cicd/                 # CI/CD 模块
│   │   │   ├── api/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   └── model/
│   │   │
│   │   ├── harbor/               # Harbor 镜像仓库模块
│   │   │   ├── api/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   └── model/
│   │   │
│   │   ├── k8s/                  # Kubernetes 管理模块
│   │   │   ├── api/              # K8s API 处理器
│   │   │   ├── service/          # K8s 业务逻辑
│   │   │   ├── repository/       # K8s 数据访问
│   │   │   └── model/            # K8s 资源模型
│   │   │
│   │   ├── log/                  # 日志管理模块
│   │   │   ├── api/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   └── model/
│   │   │
│   │   └── monitor/              # 监控管理模块
│   │       ├── api/
│   │       ├── service/
│   │       ├── repository/
│   │       └── model/
│   │
│   └── pkg/                      # 内部公共包
│       └── ...
│
├── routers/                      # 路由定义
│   ├── router.go                 # 主路由
│   ├── v1/                       # v1 API 路由
│   │   ├── v1.go                 # v1 路由聚合
│   │   ├── user.go               # 用户路由
│   │   ├── k8s.go                # K8s 路由
│   │   ├── alert.go              # 告警路由
│   │   ├── log.go                # 日志路由
│   │   ├── monitor.go            # 监控路由
│   │   ├── cicd.go               # CI/CD 路由
│   │   ├── audit.go              # 审计路由
│   │   ├── harbor.go             # Harbor 路由
│   │   ├── app.go                # 应用路由
│   │   ├── middleware.go         # 中间件路由
│   │   ├── role.go               # 角色路由
│   │   └── department.go         # 部门路由
│   ├── router_cors_test.go
│   ├── router_method_consistency_test.go
│   ├── router_probe_auth_test.go
│   └── router_swagger_contract_test.go
│
├── docs/                         # API 文档
│   ├── docs.go                   # Swagger 总文档
│   ├── openapi/                  # OpenAPI 规范
│   ├── swagger/                  # Swagger UI
│   ├── swagger.json              # Swagger JSON
│   └── swagger.yaml              # Swagger YAML
│
├── scripts/                      # 脚本工具
│   ├── bash/                     # Bash 脚本
│   ├── openapi/                  # OpenAPI 生成
│   ├── perf/                     # 性能测试
│   └── sql/                      # SQL 脚本
│
├── Dockerfile                    # Docker 镜像构建
├── Makefile                      # 构建任务
├── go.mod                        # Go 模块依赖
├── go.sum                        # 依赖校验和
└── PLAN.md                       # 开发计划
```

## 模块架构

每个业务模块遵循 DDD（领域驱动设计）分层架构：

```
modules/{name}/
├── api/           # API 层 - 处理 HTTP 请求/响应
├── service/       # 服务层 - 核心业务逻辑
├── repository/    # 仓储层 - 数据访问抽象
└── model/         # 模型层 - 领域模型定义
```

### 分层职责

| 层级 | 职责 | 示例 |
|------|------|------|
| **api** | HTTP 请求处理、参数验证、响应封装 | `UserController.Create()` |
| **service** | 业务逻辑编排、事务管理 | `UserService.CreateUser()` |
| **repository** | 数据库 CRUD 操作 | `UserRepository.Save()` |
| **model** | 数据结构定义、领域规则 | `User` 实体 |

## 核心模块

### 用户管理 (user)
- 用户 CRUD 操作
- 用户角色分配
- 用户部门关联
- LDAP 集成认证

### 告警管理 (alert)
- 告警规则配置
- 告警历史记录
- 告警通知发送

### Kubernetes 管理 (k8s)
- 集群连接管理
- 节点管理
- 工作负载管理
- 存储管理
- 网络管理

### CI/CD (cicd)
- 流水线配置
- 构建任务管理
- 部署记录

### Harbor 集成 (harbor)
- 镜像仓库同步
- 项目管理
- Webhook 处理

### 日志管理 (log)
- 日志采集配置
- 日志检索接口
- 日志保留策略

### 监控管理 (monitor)
- 监控指标配置
- 告警规则
- 数据源管理

## 中间件

| 中间件 | 功能 | 文件 |
|--------|------|------|
| **CORS** | 跨域资源共享 | `middleware/cors.go` |
| **Permission** | 权限验证 | `middleware/permission.go` |
| **Audit** | 操作审计 | `middleware/audit.go` |
| **Casbin** | RBAC 权限控制 | `middleware/casbin.go` |
| **Session** | 会话管理 | `middleware/session.go` |
| **Recover** | 异常恢复 | `middleware/recover.go` |
| **RequestLog** | 请求日志 | `middleware/request_log.go` |

## API 路由

### v1 API 端点

```
POST   /api/v1/auth/login       # 用户登录
POST   /api/v1/auth/logout      # 用户登出
GET    /api/v1/users            # 获取用户列表
POST   /api/v1/users            # 创建用户
PUT    /api/v1/users/:id        # 更新用户
DELETE /api/v1/users/:id        # 删除用户

GET    /api/v1/clusters         # 获取集群列表
POST   /api/v1/clusters         # 创建集群
GET    /api/v1/clusters/:id      # 获取集群详情

GET    /api/v1/k8s/nodes        # 获取节点列表
POST   /api/v1/k8s/nodes/:name/cordon   # 禁止调度
POST   /api/v1/k8s/nodes/:name/drain    # 驱逐节点

GET    /api/v1/alerts           # 获取告警列表
POST   /api/v1/alerts           # 创建告警规则

GET    /api/v1/logs             # 日志检索
GET    /api/v1/monitor/metrics  # 获取监控指标
```

## 配置管理

### 配置文件 (config.yaml)

```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  name: devops
  user: postgres
  password: ""

redis:
  host: localhost
  port: 6379
  db: 0

jwt:
  secret: your-secret-key
  expire: 24h

casbin:
  model: config/casbin.conf
```

### 环境变量

支持通过环境变量覆盖配置（优先级高于配置文件）：

```bash
export SERVER_PORT=9000
export DATABASE_HOST=postgres.example.com
export JWT_SECRET=production-secret
```

## 开发命令

```bash
# 运行开发服务器
go run cmd/server/main.go

# 构建
go build -o bin/server cmd/server/main.go

# 运行测试
go test ./...

# 运行测试并生成覆盖率
go test -cover ./...

# 生成 Swagger 文档
swag init -g cmd/server/main.go

# 代码格式化
go fmt ./...

# 代码检查
go vet ./...

# 使用 Makefile
make build
make test
make swagger
```

## 初始化流程

应用启动时的初始化顺序（`internal/bootstrap/`）：

1. **config.go** - 加载配置
2. **db.go** - 连接数据库
3. **redis.go** - 连接 Redis
4. **casbin.go** - 初始化权限引擎
5. **k8s.go** - 初始化 K8s 客户端
6. **router.go** - 注册路由

## 权限控制

基于 Casbin 的 RBAC 权限模型：

```
用户 (user) → 角色 (role) → 权限 (permission)
```

权限检查示例：

```go
// 中间件自动检查
r.GET("/clusters", middleware.Permission("cluster:list"), handler.GetClusters)

// 代码中手动检查
if hasPermission(c, "cluster:create") {
    // 允许操作
}
```

## WebSocket 支持

用于实时推送：

- 终端输出流 (K8s Pod 日志)
- 实时监控数据
- 告警通知推送

## Docker 部署

```bash
# 构建镜像
docker build -t devops-backend:latest .

# 运行容器
docker run -p 8080:8080 \
  -e DATABASE_HOST=postgres \
  -e REDIS_HOST=redis \
  devops-backend:latest
```

## Swagger 文档

启动服务后访问：

- Swagger UI: `http://localhost:8080/swagger/index.html`
- Swagger JSON: `http://localhost:8080/swagger/doc.json`

## 相关文档

- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [Casbin 文档](https://casbin.org/docs/)
- [Swaggo 文档](https://github.com/swaggo/swag)
