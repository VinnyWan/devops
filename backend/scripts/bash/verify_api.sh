#!/bin/bash

# 配置
API_BASE="http://localhost:8000/api/v1"
USERNAME="admin"
PASSWORD="123456"
CLUSTER_ID=1
NAMESPACE="default"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[TEST]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# 1. 登录获取 Token
log "正在登录..."
LOGIN_RES=$(curl -s -X POST "$API_BASE/user/login" \
    -H "Content-Type: application/json" \
    -d "{\"authType\": \"local\", \"username\": \"$USERNAME\", \"password\": \"$PASSWORD\"}")

TOKEN=$(echo $LOGIN_RES | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

if [ -z "$TOKEN" ]; then
    error "登录失败: $LOGIN_RES"
fi

log "登录成功，Token: ${TOKEN:0:20}..."

# 2. 验证 K8s 资源接口

# Deployment List
log "测试 Deployment 列表..."
RES=$(curl -s -X GET "$API_BASE/k8s/deployment/list?clusterId=$CLUSTER_ID&namespace=$NAMESPACE" \
    -H "Authorization: Bearer $TOKEN")
if [[ $RES != *"code\":200"* ]]; then
    error "Deployment 列表获取失败: $RES"
fi

# Pod List
log "测试 Pod 列表..."
RES=$(curl -s -X GET "$API_BASE/k8s/pod/list?clusterId=$CLUSTER_ID&namespace=$NAMESPACE" \
    -H "Authorization: Bearer $TOKEN")
if [[ $RES != *"code\":200"* ]]; then
    error "Pod 列表获取失败: $RES"
fi

# Service List
log "测试 Service 列表..."
RES=$(curl -s -X GET "$API_BASE/k8s/service/list?clusterId=$CLUSTER_ID&namespace=$NAMESPACE" \
    -H "Authorization: Bearer $TOKEN")
if [[ $RES != *"code\":200"* ]]; then
    error "Service 列表获取失败: $RES"
fi

# ConfigMap List
log "测试 ConfigMap 列表..."
RES=$(curl -s -X GET "$API_BASE/k8s/configmap/list?clusterId=$CLUSTER_ID&namespace=$NAMESPACE" \
    -H "Authorization: Bearer $TOKEN")
if [[ $RES != *"code\":200"* ]]; then
    error "ConfigMap 列表获取失败: $RES"
fi

# Ingress List
log "测试 Ingress 列表..."
RES=$(curl -s -X GET "$API_BASE/k8s/ingress/list?clusterId=$CLUSTER_ID&namespace=$NAMESPACE" \
    -H "Authorization: Bearer $TOKEN")
if [[ $RES != *"code\":200"* ]]; then
    error "Ingress 列表获取失败: $RES"
fi

log "所有接口验证通过！"
