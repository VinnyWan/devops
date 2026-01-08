# 快速使用指南

## 系统启动

```bash
# 1. 确保MySQL和Redis已经启动

# 2. 修改配置文件 config.yaml，设置正确的数据库和Redis连接信息

# 3. 启动服务
go run main.go

# 或者编译后运行
go build -o devops .
./devops
```

## 访问系统

1. **Swagger API文档**：http://localhost:8080/swagger/index.html
2. **默认账号**：admin / admin123
3. **Token获取指南**：查看 `TOKEN_GUIDE.md` 文件

## API使用流程

### 1. 获取验证码
```bash
curl http://localhost:8080/api/captcha
```

返回示例：
```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {
    "captchaId": "xxxx",
    "imageUrl": "/api/captcha/xxxx.png"
  }
}
```

### 2. 用户登录
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "captchaId": "验证码ID",
    "captchaCode": "验证码内容"
  }'
```

返回示例：
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

### 3. 调用需要认证的接口

在请求头中添加Token：
```bash
curl http://localhost:8080/api/user/info \
  -H "Authorization: Bearer <your_token>"
```

## 主要功能模块

### 用户管理
- GET /api/users - 获取用户列表
- POST /api/users - 创建用户
- GET /api/users/:id - 获取用户详情
- PUT /api/users/:id - 更新用户
- DELETE /api/users/:id - 删除用户
- POST /api/users/:id/roles - 分配角色

### 角色管理（类似的RESTful风格）
- /api/roles

### 菜单管理
- /api/menus

### 部门管理
- /api/departments

### 岗位管理
- /api/posts

### 操作日志
- /api/operation-logs

### 登录日志
- /api/login-logs

## 数据格式说明

所有接口统一使用JSON格式：
- 请求：`Content-Type: application/json`
- 响应：JSON格式

## 测试脚本

运行测试脚本：
```bash
./test_api.sh
```

## 常见问题

### 1. 数据库连接失败
检查 config.yaml 中的数据库配置是否正确

### 2. Redis连接失败
检查 config.yaml 中的Redis配置是否正确

### 3. Token过期
重新登录获取新的Token（默认有效期2小时）

### 4. Swagger文档无法访问
确保 config.yaml 中 enableSwagger 设置为 true

### 5. 关闭验证码验证
在 config.yaml 中设置：
```yaml
captcha:
  enabled: false  # 关闭验证码验证
```

## 更多信息

详细API文档请访问：http://localhost:8080/swagger/index.html
