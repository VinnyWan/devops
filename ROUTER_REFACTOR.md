# 路由重构和登录修复说明

## 修改内容总结

### 1. 路由结构重构 ✅

已将路由按功能模块拆分到不同文件：

```
routers/
├── router.go         # 主路由文件（只包含汇总逻辑）
├── auth.go          # 认证路由（登录、验证码）
├── user.go          # 用户管理路由
├── role.go          # 角色管理路由
├── menu.go          # 菜单管理路由
├── department.go    # 部门管理路由
├── post.go          # 岗位管理路由
└── log.go           # 日志管理路由
```

**主路由文件 (router.go) 现在只包含汇总逻辑：**
```go
func SetupRouter() *gin.Engine {
    r := gin.New()
    
    // Swagger文档
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    
    // API路由组
    api := r.Group("/api")
    {
        SetupAuthRoutes(api)       // 认证相关
        SetupUserRoutes(api)       // 用户管理
        SetupRoleRoutes(api)       // 角色管理
        SetupMenuRoutes(api)       // 菜单管理
        SetupDepartmentRoutes(api) // 部门管理
        SetupPostRoutes(api)       // 岗位管理
        SetupLogRoutes(api)        // 日志管理
    }
    
    return r
}
```

### 2. 修复登录参数问题 ✅

**问题原因：**
- 原代码使用 `ShouldBindJSON`，只支持JSON格式
- 你的curl命令使用URL参数（query string）

**解决方案：**
```go
// 修改前
type LoginRequest struct {
    Username    string `json:"username" binding:"required"`
    Password    string `json:"password" binding:"required"`
    CaptchaID   string `json:"captchaId" binding:"required"`
    CaptchaCode string `json:"captchaCode" binding:"required"`
}

// 修改后 - 同时支持JSON和URL参数
type LoginRequest struct {
    Username    string `json:"username" form:"username" binding:"required"`
    Password    string `json:"password" form:"password" binding:"required"`
    CaptchaID   string `json:"captchaId" form:"captchaId" binding:"required"`
    CaptchaCode string `json:"captchaCode" form:"captchaCode" binding:"required"`
}

// 绑定方法改为 ShouldBind（自动识别JSON或URL参数）
func (ctrl *UserController) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBind(&req); err != nil {
        common.Fail(c, "参数错误: "+err.Error())
        return
    }
    // ...
}
```

## 测试方法

### 方式1: 使用URL参数（你的原始curl命令）

```bash
# 注意：需要先关闭验证码验证（见下方配置）
curl --location --request POST 'http://127.0.0.1:8080/api/auth/login?username=admin&password=admin123&captchaId=test&captchaCode=1234'
```

### 方式2: 使用JSON格式（推荐）

```bash
curl -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "captchaId": "test",
    "captchaCode": "1234"
  }'
```

### 方式3: 使用POST表单

```bash
curl -X POST http://127.0.0.1:8080/api/auth/login \
  -d "username=admin" \
  -d "password=admin123" \
  -d "captchaId=test" \
  -d "captchaCode=1234"
```

## 关闭验证码验证（测试环境）

如果想跳过验证码验证，修改 `config.yaml`：

```yaml
# 验证码配置
captcha:
  enabled: false  # 改为 false
```

## 完整测试流程

### 1. 启动服务
```bash
cd /Users/wangnan/code/devops
./devops
```

### 2. 测试登录（URL参数方式）
```bash
curl --location --request POST \
  'http://127.0.0.1:8080/api/auth/login?username=admin&password=admin123&captchaId=test&captchaCode=1234'
```

**预期响应（验证码开启时会失败）：**
```json
{
    "code": 500,
    "msg": "验证码错误"
}
```

### 3. 关闭验证码后再测试
```bash
# 修改 config.yaml 中 captcha.enabled 为 false
# 重启服务
curl --location --request POST \
  'http://127.0.0.1:8080/api/auth/login?username=admin&password=admin123&captchaId=test&captchaCode=1234'
```

**预期响应（成功）：**
```json
{
    "code": 200,
    "msg": "操作成功",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "user": {
            "id": 1,
            "username": "admin",
            "nickname": "超级管理员"
        }
    }
}
```

### 4. 使用正确的验证码流程

```bash
# 步骤1: 获取验证码
curl http://127.0.0.1:8080/api/captcha

# 步骤2: 在浏览器中查看验证码图片
# http://127.0.0.1:8080/api/captcha/{captchaId}.png

# 步骤3: 使用正确的验证码登录
curl -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "captchaId": "实际的验证码ID",
    "captchaCode": "看到的验证码"
  }'
```

## 路由模块说明

### 已实现模块
- ✅ **auth.go** - 认证路由（登录、验证码）
- ✅ **user.go** - 用户管理路由（已有控制器）

### 待实现模块（已预留路由）
- ⏳ **role.go** - 角色管理（需要创建控制器）
- ⏳ **menu.go** - 菜单管理（需要创建控制器）
- ⏳ **department.go** - 部门管理（需要创建控制器）
- ⏳ **post.go** - 岗位管理（需要创建控制器）
- ⏳ **log.go** - 日志管理（需要创建控制器）

## 新增其他模块的步骤

以角色管理为例：

1. 创建控制器 `controller/role.go`
2. 在 `routers/role.go` 中取消注释相关代码
3. 添加相应的Swagger注释
4. 运行 `swag init` 更新文档

## 优势

### 重构前
- 所有路由都在一个文件中
- 难以维护和扩展
- 代码耦合度高

### 重构后
- 按功能模块分离
- 每个模块独立管理
- 易于维护和扩展
- 主路由文件简洁清晰
- 符合单一职责原则

## 文件变更清单

**新增文件：**
- routers/auth.go
- routers/user.go
- routers/role.go
- routers/menu.go
- routers/department.go
- routers/post.go
- routers/log.go

**修改文件：**
- routers/router.go（简化为汇总逻辑）
- controller/user.go（支持多种参数格式）

**配置说明：**
- config.yaml 中可通过 `captcha.enabled` 控制验证码开关
