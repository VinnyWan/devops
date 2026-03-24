# DevOps 运维平台

一个基于 Kubernetes 的企业级 DevOps 运维管理平台，提供集群管理、工作负载管理、监控告警、CI/CD 流水线等功能。

## 技术栈

### 后端 (Backend)

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.24+ | 主要开发语言 |
| Gin | 1.10 | Web 框架 |
| GORM | 1.31 | ORM 框架 |
| MySQL | 8.0+ | 主数据库 |
| Redis | 7.0+ | 缓存/Session |
| Kubernetes | 1.31+ | 容器编排 |
| Casbin | - | RBAC 权限控制 |
| Nacos | 2.3+ | 服务注册/配置中心 |
| Swagger | - | API 文档 |

### 前端 (Frontend)

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | 3.5+ | 前端框架 |
| TypeScript | 5.9+ | 类型安全 |
| Vite | 7.3+ | 构建工具 |
| Naive UI | 2.43+ | UI 组件库 |
| Pinia | 3.0+ | 状态管理 |
| Vue Router | 4.6+ | 路由管理 |
| xterm.js | 6.0+ | Web 终端 |

## 功能模块

### 核心功能

- **仪表盘** - 集群资源概览、关键指标展示
- **集群管理** - 多集群接入与管理
- **资产管理** - 资产清单管理

### Kubernetes 资源管理

- **工作负载** - Deployment、StatefulSet、DaemonSet、Job、CronJob
- **节点管理** - 节点状态、资源使用、调度控制
- **网络管理** - Service、Ingress、NetworkPolicy
- **存储管理** - PV、PVC、StorageClass
- **配置管理** - ConfigMap、Secret
- **命名空间** - 命名空间隔离与配额

### 运维中心

- **告警中心** - 告警规则配置、告警通知
- **日志检索** - 集中式日志查询与分析
- **监控配置** - Prometheus 规则管理
- **CI/CD 流水线** - 构建与部署流水线管理
- **Harbor 管理** - 镜像仓库集成
- **应用管理** - 应用生命周期管理
- **审计日志** - 操作审计与追溯

### 系统管理

- **用户管理** - 用户账号与认证
- **角色管理** - 角色定义与权限绑定
- **权限管理** - 细粒度权限控制
- **部门管理** - 组织架构管理

### 认证方式

- 本地账号登录
- LDAP 集成认证
- OIDC (OpenID Connect) 单点登录

## 项目结构

```
.
├── backend/                    # 后端服务
│   ├── cmd/server/            # 应用入口
│   │   └── main.go
│   ├── config/                # 配置管理
│   │   ├── config.go
│   │   ├── config.yaml.example
│   │   ├── nacos.go           # Nacos 配置中心
│   │   └── nacos_registry.go  # 服务注册
│   ├── internal/
│   │   ├── bootstrap/         # 初始化逻辑
│   │   │   ├── config.go
│   │   │   ├── db.go
│   │   │   ├── redis.go
│   │   │   ├── casbin.go
│   │   │   └── k8s.go
│   │   ├── middleware/        # HTTP 中间件
│   │   │   ├── cors.go
│   │   │   ├── permission.go
│   │   │   ├── session.go
│   │   │   └── audit.go
│   │   ├── modules/           # 业务模块
│   │   │   ├── user/          # 用户模块
│   │   │   │   ├── api/
│   │   │   │   ├── service/
│   │   │   │   ├── repository/
│   │   │   │   └── model/
│   │   │   └── k8s/           # Kubernetes 模块
│   │   │       ├── api/
│   │   │       ├── service/
│   │   │       └── repository/
│   │   └── pkg/               # 公共组件
│   │       ├── k8s/           # K8s 客户端
│   │       ├── logger/        # 日志
│   │       ├── redis/         # Redis 封装
│   │       └── utils/         # 工具函数
│   ├── routers/               # 路由定义
│   │   ├── router.go
│   │   └── v1/                # API v1
│   └── scripts/               # 脚本
│       └── bash/
│
├── frontend/                   # 前端应用
│   ├── src/
│   │   ├── api/               # API 请求
│   │   │   └── generated/     # Swagger 生成的 API
│   │   ├── components/        # 公共组件
│   │   │   ├── ClusterSelector.vue
│   │   │   ├── CrudTable.vue
│   │   │   ├── K8sTerminal.vue
│   │   │   └── ...
│   │   ├── composables/       # 组合式函数
│   │   ├── layouts/           # 布局组件
│   │   ├── router/            # 路由配置
│   │   ├── stores/            # Pinia 状态
│   │   ├── types/             # TypeScript 类型
│   │   │   └── generated/     # Swagger 生成的类型
│   │   ├── utils/             # 工具函数
│   │   └── views/             # 页面组件
│   │       ├── cluster/
│   │       ├── dashboard/
│   │       ├── login/
│   │       ├── node/
│   │       ├── ops/
│   │       ├── system/
│   │       └── ...
│   └── public/
│
├── docs/                       # 文档
├── tasks/                      # 任务跟踪
└── .claude/                    # Claude Code 配置
```

## 快速开始

### 环境要求

- Go 1.24+
- Node.js 18+
- MySQL 8.0+
- Redis 7.0+
- Kubernetes 集群 (可选，用于 K8s 管理功能)

### 后端启动

```bash
# 进入后端目录
cd backend

# 安装依赖
go mod download

# 复制配置文件
cp config/config.yaml.example config/config.yaml

# 编辑配置文件
vim config/config.yaml

# 运行服务
go run cmd/server/main.go
```

### 前端启动

```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 开发模式运行
npm run dev

# 生产构建
npm run build
```

### 访问服务

- 前端: http://localhost:5173
- 后端 API: http://localhost:8000
- Swagger 文档: http://localhost:8000/swagger/index.html

## 配置说明

### 后端配置 (config.yaml)

```yaml
# 服务配置
server:
  port: 8080
  model: debug          # debug / release
  enableSwagger: true

# 数据库配置
db:
  dialects: mysql
  host: 127.0.0.1
  port: 3306
  db: devops_platform
  username: root
  password: your_password
  charset: utf8mb4
  maxIdle: 10
  maxOpen: 100

# Redis 配置
redis:
  address: 127.0.0.1:6379
  password: ""

# Session 配置
session:
  expire: 7200          # 2小时

# LDAP 配置 (可选)
ldap:
  enable: false
  host: "ldap.example.com"
  port: 389
  base_dn: "dc=example,dc=com"

# OIDC 配置 (可选)
oidc:
  enable: false
  provider: "https://accounts.google.com"
  client_id: "your-client-id"
  client_secret: "your-client-secret"

# 加密配置
crypto:
  secret: "your-crypto-secret-key-32-bytes"

# 日志配置
log:
  output: console       # console / file / both
  filePath: ./logs/app.log
  level: debug
```

### 前端配置

前端支持通过环境变量配置 API 地址：

```bash
# .env.local
VITE_API_BASE_URL=http://localhost:8000
```

## API 文档

启动后端服务后，访问 Swagger 文档：

```
http://localhost:8000/swagger/index.html
```

API 认证方式：
1. 调用登录接口获取 `session_id`
2. 在 Swagger 中点击 "Authorize"
3. 输入 `Bearer {session_id}`

## 开发指南

### 代码规范

- 后端遵循 [Effective Go](https://golang.org/doc/effective_go) 规范
- 前端遵循 [Vue 风格指南](https://vuejs.org/style-guide/)
- 使用 `gofmt`、`goimports` 格式化 Go 代码
- 使用 ESLint + Prettier 格式化前端代码

### 分支管理

- `main` - 主分支，稳定版本
- `develop` - 开发分支
- `feature/*` - 功能分支
- `hotfix/*` - 热修复分支

### 提交规范

```
<type>: <description>

# type:
# feat - 新功能
# fix - 修复 bug
# refactor - 重构
# docs - 文档更新
# test - 测试
# chore - 构建/工具
```

## 许可证

Apache 2.0 License
