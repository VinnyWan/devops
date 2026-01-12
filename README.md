# DevOps 管理平台

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.24.11-blue.svg)](https://golang.org)
[![Gin Framework](https://img.shields.io/badge/Gin-1.11.0-green.svg)](https://github.com/gin-gonic/gin)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

一个基于 Gin + GORM + JWT + Kubernetes 的企业级 DevOps 管理平台

</div>

## 📖 项目简介

本项目是一个功能完整的企业级DevOps管理平台，集成了**用户权限管理**和**Kubernetes多集群管理**两大核心功能模块。采用前后端分离架构，后端使用Go语言开发，提供RESTful API接口，支持Swagger在线文档。

### ✨ 核心特性

- 🔐 **完整的RBAC权限体系** - 用户、角色、部门、岗位、菜单多维度权限控制
- ☸️ **K8s多集群管理** - 支持多个Kubernetes集群统一管理
- 🎯 **资源全生命周期管理** - Workload、Service、ConfigMap、Storage等资源管理
- 📊 **细粒度权限控制** - 集群级、命名空间级权限隔离
- 📝 **完整的审计日志** - 操作日志、登录日志全记录
- 🚀 **高性能** - 基于Gin框架，支持高并发
- 📚 **API文档自动生成** - Swagger在线文档，开箱即用
- 🔄 **优雅的错误处理** - 统一响应格式，友好的错误提示

## 🏗️ 技术架构

### 技术栈

- **后端框架**: [Gin](https://github.com/gin-gonic/gin) v1.11.0
- **ORM框架**: [GORM](https://gorm.io) v1.31.1
- **数据库**: MySQL 8.0+
- **缓存**: Redis 6.0+
- **认证**: JWT ([golang-jwt/jwt](https://github.com/golang-jwt/jwt)) v5.3.0
- **日志**: [Zap](https://github.com/uber-go/zap) v1.27.1
- **API文档**: [Swag](https://github.com/swaggo/swag) v1.16.6
- **K8s客户端**: [client-go](https://github.com/kubernetes/client-go) v0.28.4
- **验证码**: [captcha](https://github.com/dchest/captcha) v1.1.0

### 项目结构

```
devops/
├── common/                  # 公共模块
│   ├── config/             # 配置管理
│   └── response.go         # 统一响应处理
├── controller/              # 控制器层
│   ├── k8s/                # K8s控制器
│   │   ├── cluster.go      # 集群管理 (6.4KB)
│   │   └── resource.go     # 资源管理 (36.2KB, 56个接口)
│   └── user/               # 用户控制器
│       ├── captcha.go      # 验证码
│       └── user.go         # 用户管理
├── docs/                    # Swagger文档
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/                # 内部模块
│   ├── database/           # 数据库
│   │   ├── db.go          # 数据库连接
│   │   ├── init.go        # 数据初始化
│   │   ├── migrate.go     # 表迁移
│   │   └── redis.go       # Redis连接
│   └── logger/             # 日志
│       └── logger.go
├── middleware/              # 中间件
│   ├── jwt.go              # JWT认证
│   ├── k8s_permission.go   # K8s权限验证
│   └── zap.go              # 日志中间件
├── models/                  # 数据模型
│   ├── k8s/                # K8s模型
│   │   └── cluster.go      # 集群、权限、命名空间模型
│   └── user/               # 用户模型
│       ├── user.go         # 用户
│       ├── role.go         # 角色
│       ├── menu.go         # 菜单
│       ├── department.go   # 部门
│       ├── post.go         # 岗位
│       ├── login_log.go    # 登录日志
│       └── operation_log.go # 操作日志
├── routers/                 # 路由层
│   ├── k8s/                # K8s路由
│   │   ├── cluster.go      # 集群路由
│   │   └── resource.go     # 资源路由
│   ├── user/               # 用户路由
│   │   ├── auth.go         # 认证路由
│   │   ├── user.go         # 用户路由
│   │   ├── role.go         # 角色路由
│   │   ├── menu.go         # 菜单路由
│   │   ├── department.go   # 部门路由
│   │   ├── post.go         # 岗位路由
│   │   └── log.go          # 日志路由
│   └── router.go           # 路由汇总
├── service/                 # 业务逻辑层
│   ├── k8s/                # K8s服务
│   │   ├── cluster.go      # 集群管理服务
│   │   ├── namespace.go    # 命名空间服务
│   │   ├── workload.go     # 工作负载服务
│   │   ├── service.go      # Service & Ingress服务
│   │   ├── config.go       # ConfigMap & Secret服务
│   │   ├── storage.go      # 存储服务
│   │   └── node.go         # 节点 & 事件服务
│   └── user/               # 用户服务
│       ├── user.go         # 用户服务
│       ├── role.go         # 角色服务
│       ├── menu.go         # 菜单服务
│       ├── department.go   # 部门服务
│       ├── post.go         # 岗位服务
│       ├── login_log.go    # 登录日志服务
│       ├── operation_log.go # 操作日志服务
│       └── captcha.go      # 验证码服务
├── utils/                   # 工具类
│   ├── jwt.go              # JWT工具
│   └── password.go         # 密码加密
├── scripts/                 # 脚本文件
│   ├── get_token.sh        # 获取Token脚本
│   ├── test_api.sh         # API测试脚本
│   └── test_login.sh       # 登录测试脚本
├── config.yaml              # 配置文件
├── go.mod                   # Go模块定义
├── go.sum                   # 依赖校验
└── main.go                  # 程序入口
```

### 代码统计

- **Go文件总数**: 51个
- **代码总行数**: 5,486行
- **项目大小**: 83.6 MB (包含二进制)

## 🚀 快速开始

### 环境要求

- Go 1.24+ 
- MySQL 8.0+
- Redis 6.0+
- Kubernetes 1.28+ (可选，用于K8s管理功能)

### 安装步骤

#### 1. 克隆项目

```bash
git clone <repository-url>
cd devops
```

#### 2. 安装依赖

```bash
go mod download
```

#### 3. 配置文件

复制并修改配置文件 `config.yaml`：

```yaml
# 服务配置
server:
  port: 8000              # 服务端口
  model: debug            # 模式: debug/release
  enableSwagger: true     # 启用Swagger文档

# 数据库配置
db:
  dialects: mysql
  host: 127.0.0.1
  port: 3306
  db: devops
  username: root
  password: your_password
  charset: utf8
  maxIdle: 10            # 最大空闲连接
  maxOpen: 150           # 最大连接数

# Redis配置
redis:
  address: 127.0.0.1:6379
  password: ""

# JWT配置
jwt:
  secret: "your-secret-key"
  expire: 7200           # Token过期时间(秒), 2小时

# 验证码配置
captcha:
  enabled: false         # 是否启用验证码

# 日志配置
log:
  output: console        # 输出目标: console/file/both
  filePath: ./logs/app.log
  level: debug           # 日志级别: debug/info/warn/error
  enableCaller: true
  enableStacktrace: true
```

#### 4. 初始化数据库

创建数据库：

```bash
mysql -u root -p -e "CREATE DATABASE devops CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

程序首次运行时会自动创建表结构和初始数据。

#### 5. 启动服务

```bash
# 开发模式
go run main.go

# 或者编译后运行
go build -o devops .
./devops

# 指定配置文件
./devops -c /path/to/config.yaml
```

#### 6. 验证运行

访问 Swagger 文档：http://localhost:8000/swagger/index.html

## 📚 功能模块

### 1️⃣ 用户权限管理模块

#### 功能列表

- **认证管理**
  - 用户登录/登出
  - JWT Token认证
  - 验证码验证（可选）
  - 登录日志记录

- **用户管理**
  - 用户CRUD
  - 用户角色分配
  - 用户部门/岗位管理
  - 密码加密存储

- **角色管理**
  - 角色CRUD
  - 角色权限配置
  - 数据权限控制

- **菜单管理**
  - 菜单树结构
  - 动态路由
  - 按钮权限

- **部门管理**
  - 部门树结构
  - 部门用户管理

- **岗位管理**
  - 岗位CRUD
  - 岗位用户关联

- **日志管理**
  - 操作日志查询
  - 登录日志查询

#### API接口 (用户模块)

```
# 认证
POST   /api/auth/login          # 用户登录
POST   /api/auth/logout         # 用户登出
GET    /api/captcha             # 获取验证码

# 用户管理
GET    /api/users               # 获取用户列表
POST   /api/users               # 创建用户
GET    /api/users/:id           # 获取用户详情
PUT    /api/users/:id           # 更新用户
DELETE /api/users/:id           # 删除用户

# 角色管理
GET    /api/roles               # 获取角色列表
POST   /api/roles               # 创建角色
PUT    /api/roles/:id           # 更新角色
DELETE /api/roles/:id           # 删除角色

# 菜单管理
GET    /api/menus               # 获取菜单树
POST   /api/menus               # 创建菜单
PUT    /api/menus/:id           # 更新菜单
DELETE /api/menus/:id           # 删除菜单

# 部门管理
GET    /api/departments         # 获取部门树
POST   /api/departments         # 创建部门
PUT    /api/departments/:id     # 更新部门
DELETE /api/departments/:id     # 删除部门

# 岗位管理
GET    /api/posts               # 获取岗位列表
POST   /api/posts               # 创建岗位
PUT    /api/posts/:id           # 更新岗位
DELETE /api/posts/:id           # 删除岗位

# 日志查询
GET    /api/operation-logs      # 操作日志列表
GET    /api/login-logs          # 登录日志列表
```

### 2️⃣ Kubernetes 多集群管理模块

#### 功能列表

- **集群管理**
  - 多集群添加/删除/编辑
  - 集群健康检查
  - KubeConfig管理
  - 集群部门关联

- **权限管理**
  - 集群级权限控制
  - 命名空间级权限隔离
  - readonly/admin权限类型
  - 基于角色的访问控制

- **命名空间管理**
  - Namespace CRUD
  - 资源配额管理
  - 标签管理

- **工作负载管理**
  - Deployment 管理（CRUD、扩缩容、重启）
  - StatefulSet 管理
  - DaemonSet 管理
  - Pod 查看和日志

- **服务管理**
  - Service CRUD
  - Ingress CRUD
  - Endpoint 查看

- **配置管理**
  - ConfigMap CRUD
  - Secret CRUD
  - 数据版本管理

- **存储管理**
  - PersistentVolume 管理
  - PersistentVolumeClaim 管理
  - StorageClass 查看

- **节点管理**
  - 节点列表和详情
  - 节点标签管理
  - 节点污点管理

- **事件查看**
  - 集群事件
  - 命名空间事件
  - 资源事件

#### API接口 (K8s模块)

```
# 集群管理
POST   /api/k8s/clusters                    # 创建集群
GET    /api/k8s/clusters                    # 获取集群列表
GET    /api/k8s/clusters/:clusterId         # 获取集群详情
PUT    /api/k8s/clusters/:clusterId         # 更新集群
DELETE /api/k8s/clusters/:clusterId         # 删除集群
GET    /api/k8s/clusters/:clusterId/health  # 健康检查

# 集群权限
POST   /api/k8s/clusters/:clusterId/access           # 配置权限
GET    /api/k8s/clusters/:clusterId/access           # 获取权限列表
DELETE /api/k8s/clusters/:clusterId/access/:id      # 删除权限

# 命名空间
GET    /api/k8s/clusters/:clusterId/namespaces        # 命名空间列表
POST   /api/k8s/clusters/:clusterId/namespaces        # 创建命名空间
GET    /api/k8s/clusters/:clusterId/namespaces/:name  # 获取命名空间
DELETE /api/k8s/clusters/:clusterId/namespaces/:name  # 删除命名空间

# Deployment
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/deployments        # 列表
POST   /api/k8s/clusters/:clusterId/namespaces/:ns/deployments        # 创建
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/deployments/:name  # 详情
PUT    /api/k8s/clusters/:clusterId/namespaces/:ns/deployments/:name  # 更新
DELETE /api/k8s/clusters/:clusterId/namespaces/:ns/deployments/:name  # 删除
PUT    /api/k8s/clusters/:clusterId/namespaces/:ns/deployments/:name/scale   # 扩缩容
POST   /api/k8s/clusters/:clusterId/namespaces/:ns/deployments/:name/restart # 重启

# Pod
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/pods              # Pod列表
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/pods/:name        # Pod详情
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/pods/:name/logs   # Pod日志
DELETE /api/k8s/clusters/:clusterId/namespaces/:ns/pods/:name        # 删除Pod

# StatefulSet
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/statefulsets       # 列表
POST   /api/k8s/clusters/:clusterId/namespaces/:ns/statefulsets       # 创建
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/statefulsets/:name # 详情
PUT    /api/k8s/clusters/:clusterId/namespaces/:ns/statefulsets/:name # 更新
DELETE /api/k8s/clusters/:clusterId/namespaces/:ns/statefulsets/:name # 删除

# DaemonSet
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/daemonsets         # 列表
POST   /api/k8s/clusters/:clusterId/namespaces/:ns/daemonsets         # 创建
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/daemonsets/:name   # 详情
PUT    /api/k8s/clusters/:clusterId/namespaces/:ns/daemonsets/:name   # 更新
DELETE /api/k8s/clusters/:clusterId/namespaces/:ns/daemonsets/:name   # 删除

# Service
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/services           # 列表
POST   /api/k8s/clusters/:clusterId/namespaces/:ns/services           # 创建
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/services/:name     # 详情
PUT    /api/k8s/clusters/:clusterId/namespaces/:ns/services/:name     # 更新
DELETE /api/k8s/clusters/:clusterId/namespaces/:ns/services/:name     # 删除

# Ingress
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/ingresses          # 列表
POST   /api/k8s/clusters/:clusterId/namespaces/:ns/ingresses          # 创建
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/ingresses/:name    # 详情
PUT    /api/k8s/clusters/:clusterId/namespaces/:ns/ingresses/:name    # 更新
DELETE /api/k8s/clusters/:clusterId/namespaces/:ns/ingresses/:name    # 删除

# ConfigMap
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/configmaps         # 列表
POST   /api/k8s/clusters/:clusterId/namespaces/:ns/configmaps         # 创建
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/configmaps/:name   # 详情
PUT    /api/k8s/clusters/:clusterId/namespaces/:ns/configmaps/:name   # 更新
DELETE /api/k8s/clusters/:clusterId/namespaces/:ns/configmaps/:name   # 删除

# Secret
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/secrets            # 列表
POST   /api/k8s/clusters/:clusterId/namespaces/:ns/secrets            # 创建
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/secrets/:name      # 详情
PUT    /api/k8s/clusters/:clusterId/namespaces/:ns/secrets/:name      # 更新
DELETE /api/k8s/clusters/:clusterId/namespaces/:ns/secrets/:name      # 删除

# 存储
GET    /api/k8s/clusters/:clusterId/persistentvolumes                 # PV列表
GET    /api/k8s/clusters/:clusterId/storageclasses                    # StorageClass列表
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/persistentvolumeclaims  # PVC列表
POST   /api/k8s/clusters/:clusterId/namespaces/:ns/persistentvolumeclaims  # 创建PVC
DELETE /api/k8s/clusters/:clusterId/namespaces/:ns/persistentvolumeclaims/:name # 删除PVC

# 节点
GET    /api/k8s/clusters/:clusterId/nodes                             # 节点列表
GET    /api/k8s/clusters/:clusterId/nodes/:name                       # 节点详情

# 事件
GET    /api/k8s/clusters/:clusterId/events                            # 集群事件
GET    /api/k8s/clusters/:clusterId/namespaces/:ns/events             # 命名空间事件
```

## 🔐 使用指南

### 获取 Token

#### 方法1: 使用脚本

```bash
# 使用提供的脚本获取Token
./scripts/get_token.sh
```

#### 方法2: 手动调用API

```bash
# 1. 获取验证码（如果启用）
curl http://localhost:8000/api/captcha

# 2. 登录获取Token
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'

# 返回结果包含token字段
{
  "code": 200,
  "msg": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {...}
  }
}
```

### 调用受保护的API

在请求头中添加 Authorization 字段：

```bash
curl http://localhost:8000/api/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 默认账号

- **用户名**: admin
- **密码**: admin123
- **权限**: 超级管理员

### K8s集群管理示例

#### 1. 添加K8s集群

```bash
curl -X POST http://localhost:8000/api/k8s/clusters \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "生产集群",
    "description": "生产环境K8s集群",
    "apiServer": "https://k8s.example.com:6443",
    "kubeConfig": "<KubeConfig内容>",
    "deptId": 1
  }'
```

#### 2. 配置集群访问权限

```bash
curl -X POST http://localhost:8000/api/k8s/clusters/1/access \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "roleId": 2,
    "accessType": "readonly",
    "namespaces": ["default", "dev"]
  }'
```

#### 3. 获取Deployment列表

```bash
curl -X GET "http://localhost:8000/api/k8s/clusters/1/namespaces/default/deployments" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 4. 扩缩容Deployment

```bash
curl -X PUT "http://localhost:8000/api/k8s/clusters/1/namespaces/default/deployments/nginx/scale" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"replicas": 3}'
```

## 🧪 测试

### 运行测试脚本

```bash
# 登录测试
./scripts/test_login.sh

# API测试
./scripts/test_api.sh

# 获取Token
./scripts/get_token.sh
```

### 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./service/k8s

# 查看测试覆盖率
go test -cover ./...
```

## 📊 API 文档

### Swagger 文档

启动服务后访问：**http://localhost:8000/swagger/index.html**

Swagger提供了完整的API文档，包括：
- 所有接口的请求/响应示例
- 在线测试功能
- 数据模型定义
- 认证配置

### 重新生成Swagger文档

```bash
# 安装swag工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init --parseDependency --parseInternal
```

## ⚙️ 配置说明

### 配置文件结构

```yaml
server:                    # 服务器配置
  port: 8000              # 端口号
  model: debug            # 运行模式: debug/release
  enableSwagger: true     # 是否启用Swagger

db:                        # 数据库配置
  dialects: mysql
  host: 127.0.0.1
  port: 3306
  db: devops
  username: root
  password: password
  charset: utf8
  maxIdle: 10
  maxOpen: 150

redis:                     # Redis配置
  address: 127.0.0.1:6379
  password: ""

jwt:                       # JWT配置
  secret: "your-secret-key"
  expire: 7200             # Token有效期(秒)

captcha:                   # 验证码配置
  enabled: false           # 是否启用

log:                       # 日志配置
  output: console          # console/file/both
  filePath: ./logs/app.log
  level: debug             # debug/info/warn/error
  enableCaller: true
  enableStacktrace: true
```

## 🛠️ 开发指南

### 添加新功能

1. **定义数据模型** (models/)
2. **实现业务逻辑** (service/)
3. **创建控制器** (controller/)
4. **配置路由** (routers/)
5. **更新Swagger文档** (`swag init`)

### 代码规范

- 遵循Go官方代码规范
- 使用`gofmt`格式化代码
- 添加必要的注释
- Controller层添加Swagger注释

### 提交规范

```bash
# 功能开发
git commit -m "feat: 添加XXX功能"

# Bug修复
git commit -m "fix: 修复XXX问题"

# 文档更新
git commit -m "docs: 更新XXX文档"

# 代码重构
git commit -m "refactor: 重构XXX模块"
```

## ❓ 常见问题

### 1. 数据库连接失败

**问题**: `dial tcp: connect: connection refused`

**解决方案**:
- 检查MySQL是否启动
- 确认`config.yaml`中的数据库配置正确
- 检查数据库是否已创建

### 2. Redis连接失败

**问题**: `dial tcp: connect: connection refused`

**解决方案**:
- 检查Redis是否启动
- 确认`config.yaml`中的Redis配置正确
- 检查Redis密码是否正确

### 3. Token过期

**问题**: `401 Unauthorized: token过期`

**解决方案**:
- 重新登录获取新Token
- 默认有效期为2小时，可在配置文件中修改

### 4. Swagger文档无法访问

**问题**: `404 Not Found`

**解决方案**:
- 确保`config.yaml`中`enableSwagger: true`
- 检查docs目录是否存在
- 重新生成文档: `swag init --parseDependency --parseInternal`

### 5. K8s集群连接失败

**问题**: 无法连接到K8s集群

**解决方案**:
- 检查KubeConfig内容是否正确
- 确认集群API Server地址可访问
- 检查集群证书是否有效

### 6. 权限验证失败

**问题**: `403 Forbidden: 权限不足`

**解决方案**:
- 检查用户角色配置
- 确认集群访问权限已配置
- 检查命名空间权限设置

### 7. 编译失败

**问题**: `package XXX is not in GOROOT`

**解决方案**:
```bash
go mod download
go mod tidy
```

## 🔒 安全建议

1. **生产环境配置**
   - 修改JWT密钥为强密码
   - 启用验证码验证
   - 使用HTTPS
   - 限制Swagger访问

2. **数据库安全**
   - 使用强密码
   - 限制数据库访问IP
   - 定期备份数据

3. **K8s集群安全**
   - KubeConfig加密存储（待实现）
   - 最小权限原则
   - 定期审计操作日志
   - Secret数据脱敏

4. **Token管理**
   - 合理设置过期时间
   - 定期轮换JWT密钥
   - 禁用不活跃用户

## 📝 更新日志

### v1.0.0 (2026-01-09)

- ✅ 完成用户权限管理模块
- ✅ 完成K8s多集群管理模块
- ✅ 实现56个K8s资源管理接口
- ✅ 集成Swagger API文档
- ✅ 实现RBAC权限控制
- ✅ 实现操作日志记录
- ✅ 优化项目目录结构

## 🚧 待实现功能

- [ ] WebShell终端 (WebSocket + K8s Exec)
- [ ] K8s资源监控和告警
- [ ] Helm应用管理
- [ ] YAML模板库
- [ ] 资源拓扑图
- [ ] 操作审计导出
- [ ] 多语言支持

## 📄 许可证

Apache License 2.0

## 👥 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

如有问题或建议，请提交 Issue。
