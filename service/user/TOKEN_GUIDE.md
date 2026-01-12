# 如何获取 Bearer Token

## 方式一：通过API接口获取（推荐）

### 步骤1：获取验证码

**请求：**
```bash
curl -X GET http://localhost:8000/api/captcha
```

**响应示例：**
```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {
    "captchaId": "xyz123",
    "imageUrl": "/api/captcha/xyz123.png"
  }
}
```

### 步骤2：查看验证码图片

在浏览器中访问：
```
http://localhost:8000/api/captcha/xyz123.png
```

### 步骤3：使用验证码登录获取Token

**请求：**
```bash
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "captchaId": "xyz123",
    "captchaCode": "您看到的验证码"
  }'
```

**响应示例：**
```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOjEsInVzZXJuYW1lIjoiYWRtaW4iLCJleHAiOjE3MDQ3MTI4MDAsImlhdCI6MTcwNDcwNTYwMCwibmJmIjoxNzA0NzA1NjAwLCJpc3MiOiJkZXZvcHMifQ.xxxxxxxxxxxxxxxxxxxx",
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "超级管理员",
      "email": "admin@example.com",
      "phone": "13800138000",
      "status": 1,
      "gender": 1
    }
  }
}
```

**提取Token：**
从响应中提取 `data.token` 字段的值。

---

## 方式二：通过Swagger获取

### 步骤1：访问Swagger文档
```
http://localhost:8000/swagger/index.html
```

### 步骤2：测试登录接口

1. 找到 `用户管理` -> `POST /api/auth/login` 接口
2. 点击 "Try it out"
3. 先调用 `/api/captcha` 获取验证码ID
4. 在浏览器中查看验证码图片
5. 填写登录信息：
   ```json
   {
     "username": "admin",
     "password": "admin123",
     "captchaId": "验证码ID",
     "captchaCode": "验证码内容"
   }
   ```
6. 点击 "Execute" 执行
7. 从响应中复制 `data.token` 的值

### 步骤3：在Swagger中使用Token

1. 点击页面右上角的 **Authorize** 按钮（锁形图标）
2. 在弹出框中输入：`Bearer <your_token>`
   - 注意：`Bearer` 和 token 之间有一个空格
   - 例如：`Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
3. 点击 "Authorize"
4. 关闭对话框

现在所有需要认证的接口都会自动携带这个Token。

---

## 方式三：使用Postman

### 步骤1：获取Token
1. 创建 POST 请求：`http://localhost:8000/api/auth/login`
2. 设置 Headers：
   - `Content-Type: application/json`
3. 在 Body 中选择 `raw` 和 `JSON`，填入：
   ```json
   {
     "username": "admin",
     "password": "admin123",
     "captchaId": "验证码ID",
     "captchaCode": "验证码"
   }
   ```
4. 发送请求，从响应中复制 token

### 步骤2：使用Token
在后续需要认证的请求中：
1. 进入 `Authorization` 标签
2. Type 选择 `Bearer Token`
3. 在 Token 输入框中粘贴你的 token
4. 发送请求

**或者手动设置 Header：**
- Key: `Authorization`
- Value: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`

---

## 验证码开关配置

如果您在测试环境想要关闭验证码验证，可以修改 `config.yaml`：

```yaml
# 验证码配置
captcha:
  enabled: false  # 设置为 false 关闭验证码验证
```

**注意：** 关闭验证码后，登录时仍需传递 `captchaId` 和 `captchaCode` 字段，但不会进行实际验证。

---

## Token 使用示例

### curl 命令示例
```bash
# 使用 Token 访问需要认证的接口
curl -X GET http://localhost:8000/api/user/info \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### JavaScript/Axios 示例
```javascript
axios.get('http://localhost:8000/api/user/info', {
  headers: {
    'Authorization': 'Bearer ' + token
  }
})
```

### Python/Requests 示例
```python
import requests

headers = {
    'Authorization': f'Bearer {token}'
}
response = requests.get('http://localhost:8000/api/user/info', headers=headers)
```

---

## Token 相关信息

- **有效期：** 2小时（7200秒）
- **格式：** JWT (JSON Web Token)
- **使用方式：** 在请求头中添加 `Authorization: Bearer <token>`
- **过期处理：** Token过期后需要重新登录获取新Token

---

## 常见问题

### Q1: Token格式错误
**错误信息：** "Token格式错误"

**解决方案：** 确保Authorization header格式为：`Bearer <token>`，注意 Bearer 和 token 之间有一个空格。

### Q2: Token无效或已过期
**错误信息：** "Token无效或已过期"

**解决方案：** 重新登录获取新的Token。

### Q3: 验证码错误
**错误信息：** "验证码错误"

**解决方案：**
1. 确保验证码ID和验证码内容正确
2. 验证码有效期为5分钟，过期需重新获取
3. 验证码只能使用一次，使用后会自动删除

### Q4: 想关闭验证码验证
**解决方案：** 修改 `config.yaml` 中的 `captcha.enabled` 为 `false`

---

## 默认账号信息

- **用户名：** admin
- **密码：** admin123

系统启动时会自动创建此管理员账号。
