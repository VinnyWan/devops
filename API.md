# DevOps 系统管理平台

基于 Go + Gin + GORM + JWT + Swagger 构建的RESTful API系统管理平台

## 功能特性

### 已实现功能

#### 1. 系统管理
- ✅ **用户管理**：用户增删改查、状态管理、角色分配
- ✅ **角色权限**：角色增删改查、菜单权限分配
- ✅ **菜单管理**：菜单增删改查、树形结构展示
- ✅ **部门管理**：部门增删改查、树形结构展示
- ✅ **岗位管理**：岗位增删改查
- ✅ **操作日志**：操作日志记录与查询
- ✅ **登录日志**：登录日志记录与查询

#### 2. 身份认证
- ✅ **JWT认证**：基于JWT的Token认证机制
- ✅ **验证码**：图形验证码生成与验证
- ✅ **密码加密**：bcrypt加密算法

#### 3. API文档
- ✅ **Swagger文档**：自动生成API文档，支持在线测试

## 技术栈

- **Web框架**：Gin 1.11.0
- **ORM**：GORM 1.31.1
- **数据库**：MySQL 8.0+
- **缓存**：Redis
- **日志**：Zap
- **文档**：Swagger
- **认证**：JWT

## 快速开始

### 1. 环境要求

- Go 1.24+
- MySQL 8.0+
- Redis 6.0+

### 2. 配置文件

编辑 `config.yaml`：

```yaml
server:
  port: 8080
  model: release
  enableSwagger: true

db:
  dialects: mysql
  host: 10.177.42.165
  port: 3306
  db: devops
  username: root
  password: rootpassword
  charset: utf8
  maxIdle: 10
  maxOpen: 150

redis:
  address: 10.177.42.165:6379
  password: "admin"

jwt:
  secret: "devops-secret-key-2026"
  expire: 7200  # token过期时间(秒), 2小时
```

### 3. 运行项目

```bash
# 安装依赖
go mod tidy

# 运行项目
go run main.go

# 或者编译后运行
go build -o devops .
./devops
```

### 4. 访问接口

- **Swagger文档**：http://localhost:8080/swagger/index.html
- **API基础路径**：http://localhost:8080/api

### 5. 默认账号

系统会自动初始化以下账号：

- **用户名**：admin
- **密码**：admin123

## API接口说明

### 公开接口（无需认证）

#### 1. 验证码
```http
GET /api/captcha
```
生成验证码ID和图片URL

```http
GET /api/captcha/:id
```
获取验证码图片

#### 2. 登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123",
  "captchaId": "验证码ID",
  "captchaCode": "验证码"
}
```

### 认证接口（需要JWT Token）

所有认证接口需要在请求头中携带Token：
```
Authorization: Bearer <your_token>
```

#### 用户管理

```http
# 获取当前用户信息
GET /api/user/info

# 获取用户列表
GET /api/users?page=1&pageSize=10&username=&phone=&status=

# 创建用户
POST /api/users
Content-Type: application/json
{
  "username": "testuser",
  "password": "123456",
  "nickname": "测试用户",
  "email": "test@example.com",
  "phone": "13800138001",
  "status": 1,
  "gender": 1,
  "deptId": 1,
  "postId": 1
}

# 获取用户详情
GET /api/users/:id

# 更新用户
PUT /api/users/:id
Content-Type: application/json
{
  "nickname": "新昵称",
  "email": "newemail@example.com",
  "status": 1
}

# 删除用户
DELETE /api/users/:id

# 分配角色
POST /api/users/:id/roles
Content-Type: application/json
{
  "roleIds": [1, 2]
}
```

## 数据库表结构

系统会自动创建以下数据表：

- `sys_user` - 用户表
- `sys_role` - 角色表
- `sys_menu` - 菜单表
- `sys_department` - 部门表
- `sys_post` - 岗位表
- `sys_operation_log` - 操作日志表
- `sys_login_log` - 登录日志表
- `user_roles` - 用户角色关联表
- `role_menus` - 角色菜单关联表

## 项目结构

```
devops/
├── common/              # 公共包
│   ├── config/         # 配置管理
│   └── response.go     # 统一响应结构
├── controller/         # 控制器层
│   ├── user.go
│   └── captcha.go
├── docs/               # Swagger文档
├── internal/           # 内部包
│   ├── database/       # 数据库
│   │   ├── db.go
│   │   ├── redis.go
│   │   ├── migrate.go
│   │   └── init.go
│   └── logger/         # 日志
├── middleware/         # 中间件
│   ├── jwt.go         # JWT认证
│   └── zap.go         # 日志中间件
├── models/            # 数据模型
│   ├── user.go
│   ├── role.go
│   ├── menu.go
│   ├── department.go
│   ├── post.go
│   ├── operation_log.go
│   └── login_log.go
├── routers/           # 路由
│   └── router.go
├── service/           # 业务逻辑层
│   ├── user.go
│   ├── role.go
│   ├── menu.go
│   ├── department.go
│   ├── post.go
│   ├── captcha.go
│   ├── operation_log.go
│   └── login_log.go
├── utils/             # 工具包
│   ├── jwt.go        # JWT工具
│   └── password.go   # 密码加密
├── config.yaml        # 配置文件
├── go.mod
└── main.go           # 程序入口
```

## RESTful API 设计规范

本项目严格遵循RESTful API设计规范：

- 使用标准HTTP方法：GET（查询）、POST（创建）、PUT（更新）、DELETE（删除）
- 统一的JSON数据格式
- 标准的HTTP状态码
- 资源命名使用复数形式
- 清晰的URL层级结构

## 响应格式

### 成功响应
```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {}
}
```

### 分页响应
```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {
    "list": [],
    "total": 100,
    "page": 1,
    "pageSize": 10
  }
}
```

### 错误响应
```json
{
  "code": 500,
  "msg": "错误信息"
}
```

## 开发说明

### 添加新的模块

1. 在 `models/` 中定义数据模型
2. 在 `service/` 中实现业务逻辑
3. 在 `controller/` 中实现控制器
4. 在 `routers/router.go` 中注册路由
5. 添加Swagger注释
6. 运行 `swag init` 重新生成文档

### 数据库迁移

系统启动时会自动执行数据库迁移和初始化数据，无需手动创建表结构。

## 许可证

Apache 2.0
