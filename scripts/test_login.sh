#!/bin/bash

echo "正在启动服务..."
cd /Users/wangnan/code/devops
./devops > server.log 2>&1 &
SERVER_PID=$!
echo "服务PID: $SERVER_PID"

# 等待服务启动
sleep 3

# 检查服务是否启动
if ! ps -p $SERVER_PID > /dev/null; then
    echo "服务启动失败"
    cat server.log
    exit 1
fi

echo "服务已启动，开始测试..."
echo ""

# 测试登录（URL参数方式）
echo "测试1: 使用URL参数登录（关闭验证码）"
curl -s --location --request POST 'http://127.0.0.1:8080/api/auth/login?username=admin&password=admin123&captchaId=test&captchaCode=1234'
echo ""
echo ""

# 测试登录（JSON方式）
echo "测试2: 使用JSON方式登录（关闭验证码）"
curl -s -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "captchaId": "test",
    "captchaCode": "1234"
  }'
echo ""
echo ""

# 关闭服务
echo "测试完成，关闭服务..."
kill $SERVER_PID 2>/dev/null
sleep 1
kill -9 $SERVER_PID 2>/dev/null

echo "完成！"
