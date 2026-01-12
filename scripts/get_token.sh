#!/bin/bash

# 简单的Token获取示例脚本
BASE_URL="http://localhost:8000/api"

echo "=========================================="
echo "获取 Bearer Token 示例"
echo "=========================================="
echo ""

# 检查是否关闭了验证码
echo "提示：如果测试环境想跳过验证码，可以在 config.yaml 中设置："
echo "captcha:"
echo "  enabled: false"
echo ""

# 1. 获取验证码
echo "步骤1: 获取验证码..."
CAPTCHA_RESPONSE=$(curl -s "$BASE_URL/captcha")
echo "响应: $CAPTCHA_RESPONSE"
echo ""

CAPTCHA_ID=$(echo $CAPTCHA_RESPONSE | grep -o '"captchaId":"[^"]*"' | cut -d'"' -f4)
IMAGE_URL=$(echo $CAPTCHA_RESPONSE | grep -o '"imageUrl":"[^"]*"' | cut -d'"' -f4)

if [ -z "$CAPTCHA_ID" ]; then
    echo "错误：无法获取验证码ID"
    exit 1
fi

echo "验证码ID: $CAPTCHA_ID"
echo "验证码图片: http://localhost:8000$IMAGE_URL"
echo ""

# 2. 提示用户查看验证码
echo "步骤2: 请在浏览器中打开以下链接查看验证码："
echo "http://localhost:8000$IMAGE_URL"
echo ""
read -p "请输入您看到的验证码: " CAPTCHA_CODE

# 3. 登录获取Token
echo ""
echo "步骤3: 正在登录..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"admin\",
    \"password\": \"admin123\",
    \"captchaId\": \"$CAPTCHA_ID\",
    \"captchaCode\": \"$CAPTCHA_CODE\"
  }")

echo "登录响应:"
echo "$LOGIN_RESPONSE" | jq . 2>/dev/null || echo "$LOGIN_RESPONSE"
echo ""

# 4. 提取Token
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ 登录失败，无法获取Token"
    echo "可能的原因："
    echo "  1. 验证码输入错误"
    echo "  2. 用户名或密码错误"
    echo "  3. 服务未启动"
    exit 1
fi

echo "=========================================="
echo "✅ 成功获取 Bearer Token！"
echo "=========================================="
echo ""
echo "Token: $TOKEN"
echo ""
echo "使用方式："
echo "----------------------------------------"
echo "1. curl 命令："
echo "   curl -H \"Authorization: Bearer $TOKEN\" http://localhost:8000/api/user/info"
echo ""
echo "2. Swagger授权："
echo "   在 http://localhost:8000/swagger/index.html 点击 Authorize"
echo "   输入: Bearer $TOKEN"
echo ""
echo "3. Postman："
echo "   Authorization -> Type: Bearer Token -> Token: $TOKEN"
echo ""

# 5. 测试Token
echo "=========================================="
echo "测试 Token 是否有效..."
echo "=========================================="
echo ""

USER_INFO=$(curl -s -X GET "$BASE_URL/user/info" \
  -H "Authorization: Bearer $TOKEN")

echo "当前用户信息:"
echo "$USER_INFO" | jq . 2>/dev/null || echo "$USER_INFO"
echo ""

# 检查是否成功
if echo "$USER_INFO" | grep -q '"code":200'; then
    echo "✅ Token 验证成功！可以正常使用了。"
else
    echo "❌ Token 验证失败，请检查。"
fi

echo ""
echo "Token 有效期: 2小时"
echo "过期后需要重新登录获取新Token"
