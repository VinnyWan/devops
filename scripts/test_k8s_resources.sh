#!/bin/bash

# K8s 资源接口测试脚本
# 测试命名空间和 Deployment 的序列化返回

echo "=========================================="
echo "K8s 资源接口测试"
echo "=========================================="
echo ""

# 配置
BASE_URL="http://127.0.0.1:8000"
CLUSTER_ID=14  # 请根据实际集群ID修改

# 检查是否传入了 Token
if [ -z "$1" ]; then
    echo "❌ 请提供 JWT Token"
    echo "用法: $0 <JWT_TOKEN>"
    echo ""
    echo "获取 Token 的方法："
    echo "  curl -X POST \"$BASE_URL/api/auth/login\" \\"
    echo "    -H \"Content-Type: application/json\" \\"
    echo "    -d '{\"username\":\"admin\",\"password\":\"your_password\"}'"
    exit 1
fi

TOKEN=$1

echo "📋 配置信息:"
echo "  - 接口地址: $BASE_URL"
echo "  - 集群ID: $CLUSTER_ID"
echo "  - Token: ${TOKEN:0:20}..."
echo ""

# 测试1: 获取命名空间列表（简化版 - 仅返回名称）
echo "=========================================="
echo "测试1: 获取命名空间列表"
echo "=========================================="
echo "📡 请求: GET /api/k8s/clusters/$CLUSTER_ID/namespaces"
echo ""

response=$(curl -s -w "\n%{http_code}" -X GET \
  "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces" \
  -H "Authorization: Bearer $TOKEN")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "📊 HTTP 状态码: $http_code"
echo ""
echo "📦 响应数据:"
echo "$body" | jq '.'
echo ""

if [ "$http_code" = "200" ]; then
    echo "✅ 测试通过 - 命名空间列表格式："
    echo "$body" | jq '.data[0]' 2>/dev/null
    echo ""
    echo "   返回字段："
    echo "   - name: 命名空间名称"
else
    echo "❌ 测试失败"
fi

echo ""
echo "=========================================="
echo "测试2: 获取 Deployment 列表"
echo "=========================================="

# 从命名空间列表中提取第一个命名空间
NAMESPACE=$(echo "$body" | jq -r '.data[0].name' 2>/dev/null)

if [ -z "$NAMESPACE" ] || [ "$NAMESPACE" = "null" ]; then
    NAMESPACE="default"
fi

echo "📡 请求: GET /api/k8s/clusters/$CLUSTER_ID/deployments?namespace=$NAMESPACE"
echo ""

response=$(curl -s -w "\n%{http_code}" -X GET \
  "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/deployments?namespace=$NAMESPACE" \
  -H "Authorization: Bearer $TOKEN")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "📊 HTTP 状态码: $http_code"
echo ""
echo "📦 响应数据:"
echo "$body" | jq '.'
echo ""

if [ "$http_code" = "200" ]; then
    echo "✅ 测试通过 - Deployment 列表格式："
    echo "$body" | jq '.data[0]' 2>/dev/null
    echo ""
    echo "   返回字段："
    echo "   - name: 名称"
    echo "   - namespace: 命名空间"
    echo "   - replicas: 副本数"
    echo "   - images: 镜像列表 (数组)"
    echo "   - labels: 标签 (对象)"
    echo "   - createTime: 创建时间"
    echo "   - updateTime: 更新时间"
else
    echo "❌ 测试失败"
fi

echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="
