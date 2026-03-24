# DevOps Platform

基于 Go 和 Vue 3 构建的现代化 DevOps 管理平台，专注于 Kubernetes 多集群管理、细粒度权限控制及运维自动化。

## 📖 项目简介

DevOps Platform 是一个企业级的运维管理系统，旨在简化 Kubernetes 集群的管理复杂度。它提供了一套统一的控制平面，允许用户通过 Web 界面纳管多个 K8s 集群，管理工作负载、配置资源，并具备完善的 RBAC 权限体系和审计日志功能。

## ✨ 核心功能

- **多集群管理**: 统一纳管开发、测试、生产等多套 Kubernetes 环境。
- **Kubernetes 资源操作**: 可视化管理 Deployments, Pods, Services, ConfigMaps, Namespaces 等核心资源。
- **细粒度权限控制 (RBAC)**: 基于 Casbin 和自定义 RBAC 模型，支持用户、角色、权限的灵活配置。
- **用户认证与安全**: 集成 JWT 认证，支持审计日志记录，确保操作可追溯。
- **终端控制台**: (计划中) 提供 Web Terminal 直接连接 Pod。
- **监控集成**: (计划中) 集成 Prometheus 和 Grafana 指标展示。

## 🛠 技术栈

### 后端 (Backend)
- **语言**: Go (1.21+)
- **Web 框架**: [Gin](https://github.com/gin-gonic/gin)
- **数据库 ORM**: [GORM](https://gorm.io/) (MySQL)
- **K8s Client**: client-go
- **配置管理**: Viper
- **日志**: Zap
- **API 文档**: Swagger

### 前端 (Frontend)
- **框架**: Vue 3
- **构建工具**: Vite
- **UI 组件库**: (推测为 Element Plus 或 Ant Design Vue)
- **语言**: JavaScript / TypeScript

### 基础设施
- **数据库**: MySQL 5.7+
- **缓存**: Redis 6.0+

## 🚀 快速开始

### 前置要求
- Go 1.21+
- Node.js 16+
- MySQL
- Redis

### 1. 后端启动

```bash
# 克隆项目
git clone https://github.com/your-repo/devops-platform.git
cd devops-platform

# 配置数据库与环境
# 复制配置文件 (根据实际情况修改 config.yaml)
cp config/config.yaml.example config/config.yaml 

# 初始化数据库
# 执行 scripts/sql/01_init.sql 中的 SQL 语句

# 下载依赖
go mod tidy

# 运行服务
make run
# 或直接运行
go run cmd/server/main.go
```

### 2. 前端启动

```bash
cd web

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

### 3. 访问系统
- 后端 API 地址: `http://localhost:8080`
- 前端访问地址: `http://localhost:5173` (默认 Vite 端口)
- Swagger 文档: `http://localhost:8080/swagger/index.html`

## 📂 项目结构

```
devops-platform/
├── cmd/                # 程序入口
├── config/             # 配置文件结构
├── docs/               # 文档与 Swagger 定义
├── internal/           # 内部业务逻辑 (Clean Architecture)
│   ├── api/            # HTTP Handlers
│   ├── bootstrap/      # 启动初始化 (DB, Redis, K8s 等)
│   ├── middleware/     # Gin 中间件
│   ├── model/          # 数据库模型
│   ├── pkg/            # 内部工具包
│   ├── repository/     # 数据访问层
│   └── service/        # 业务逻辑层
├── pkg/                # 公共工具库
├── routers/            # 路由定义
├── scripts/            # SQL 和 Shell 脚本
├── web/                # Vue 前端代码
├── Makefile            # 构建命令
└── go.mod              # Go 依赖定义
```

## 🤝 贡献指南

欢迎提交 Pull Request 或 Issue！

## 📄 许可证

MIT License
