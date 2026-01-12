#!/bin/bash

# API测试脚本
BASE_URL="http://localhost:8000/api"

echo "========================================"
echo "DevOps系统管理平台 API 测试"
echo "========================================"
echo ""

# 1. 获取验证码
echo "1. 获取验证码..."
CAPTCHA_RESPONSE=$(curl -s "$BASE_URL/captcha")
CAPTCHA_ID=$(echo $CAPTCHA_RESPONSE | grep -o '"captchaId":"[^"]*"' | cut -d'"' -f4)
echo "验证码ID: $CAPTCHA_ID"
echo "验证码URL: http://localhost:8000$BASE_URL/captcha/$CAPTCHA_ID"
echo ""

# 2. 登录（需要手动输入验证码）
echo "2. 用户登录..."
echo "请在浏览器中打开验证码URL并输入验证码："
read -p "验证码: " CAPTCHA_CODE

LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"admin\",
    \"password\": \"admin123\",
    \"captchaId\": \"$CAPTCHA_ID\",
    \"captchaCode\": \"$CAPTCHA_CODE\"
  }")

echo "登录响应: $LOGIN_RESPONSE"
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "登录失败，请检查验证码是否正确"
    exit 1
fi

echo "Token: $TOKEN"
echo ""

# 3. 获取当前用户信息
echo "3. 获取当前用户信息..."
curl -s -X GET "$BASE_URL/user/info" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 4. 获取用户列表
echo "4. 获取用户列表..."
curl -s -X GET "$BASE_URL/users?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# 5. 创建新用户
echo "5. 创建新用户..."
curl -s -X POST "$BASE_URL/users" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456",
    "nickname": "测试用户",
    "email": "test@example.com",
    "phone": "13800138001",
    "status": 1,
    "gender": 1
  }' | jq .
echo ""

echo "========================================"
echo "测试完成！"
echo "========================================"
echo ""
echo "更多接口请访问 Swagger 文档："
echo "http://localhost:8000/swagger/index.html"
