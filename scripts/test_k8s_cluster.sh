#!/bin/bash

# K8s集群管理功能测试脚本
# 用于测试版本检测、导入状态、重新导入等功能

API_BASE="http://localhost:8000/api"
TOKEN=""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印函数
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

print_section() {
    echo ""
    echo "=========================================="
    echo "$1"
    echo "=========================================="
}

# 获取Token
get_token() {
    print_section "1. 获取登录Token"
    
    response=$(curl -s -X POST "$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "admin",
            "password": "admin123"
        }')
    
    TOKEN=$(echo $response | jq -r '.data.token')
    
    if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
        print_success "登录成功，Token: ${TOKEN:0:20}..."
    else
        print_error "登录失败"
        exit 1
    fi
}

# 测试创建集群（成功案例）
test_create_cluster_success() {
    print_section "2. 创建K8s集群（版本检测）"
    
    print_info "准备KubeConfig..."
    
    # 注意：这里需要替换为真实的KubeConfig
    KUBECONFIG_CONTENT='apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: LS0tLS1...
    server: https://k8s.example.com:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: kubernetes-admin@kubernetes
users:
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1...
    client-key-data: LS0tLS1...'
    
    print_info "创建集群中..."
    
    response=$(curl -s -X POST "$API_BASE/k8s/clusters" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"测试集群-$(date +%s)\",
            \"description\": \"自动测试集群\",
            \"apiServer\": \"https://k8s.example.com:6443\",
            \"kubeConfig\": $(echo "$KUBECONFIG_CONTENT" | jq -Rs .),
            \"deptId\": 1,
            \"remark\": \"测试创建\"
        }")
    
    code=$(echo $response | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        CLUSTER_ID=$(echo $response | jq -r '.data.id')
        VERSION=$(echo $response | jq -r '.data.version')
        IMPORT_STATUS=$(echo $response | jq -r '.data.importStatus')
        CLUSTER_STATUS=$(echo $response | jq -r '.data.clusterStatus')
        
        print_success "集群创建成功"
        echo "  集群ID: $CLUSTER_ID"
        echo "  K8s版本: $VERSION"
        echo "  导入状态: $IMPORT_STATUS"
        echo "  集群状态: $CLUSTER_STATUS"
    else
        msg=$(echo $response | jq -r '.msg')
        print_error "集群创建失败: $msg"
        echo "响应: $response"
    fi
}

# 测试获取集群列表
test_get_cluster_list() {
    print_section "3. 获取集群列表"
    
    response=$(curl -s -X GET "$API_BASE/k8s/clusters?page=1&pageSize=10" \
        -H "Authorization: Bearer $TOKEN")
    
    code=$(echo $response | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        total=$(echo $response | jq -r '.data.total')
        print_success "获取集群列表成功，共 $total 个集群"
        
        # 显示集群信息
        echo "$response" | jq -r '.data.list[] | "  ID: \(.id) | 名称: \(.name) | 版本: \(.version) | 导入状态: \(.importStatus) | 集群状态: \(.clusterStatus)"'
    else
        print_error "获取列表失败"
    fi
}

# 测试健康检查
test_health_check() {
    print_section "4. 集群健康检查"
    
    if [ -z "$CLUSTER_ID" ]; then
        print_info "跳过健康检查（无可用集群ID）"
        return
    fi
    
    print_info "检查集群 $CLUSTER_ID 的健康状态..."
    
    response=$(curl -s -X GET "$API_BASE/k8s/clusters/$CLUSTER_ID/health" \
        -H "Authorization: Bearer $TOKEN")
    
    code=$(echo $response | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        healthy=$(echo $response | jq -r '.data.healthy')
        version=$(echo $response | jq -r '.data.version')
        message=$(echo $response | jq -r '.data.message')
        
        if [ "$healthy" = "true" ]; then
            print_success "集群健康"
        else
            print_error "集群不健康"
        fi
        
        echo "  版本: $version"
        echo "  消息: $message"
    else
        print_error "健康检查失败"
    fi
}

# 测试重新导入KubeConfig
test_reimport_kubeconfig() {
    print_section "5. 重新导入KubeConfig"
    
    if [ -z "$CLUSTER_ID" ]; then
        print_info "跳过重新导入（无可用集群ID）"
        return
    fi
    
    print_info "重新导入集群 $CLUSTER_ID 的配置..."
    
    # 使用相同的KubeConfig（实际使用中应该是新的配置）
    response=$(curl -s -X POST "$API_BASE/k8s/clusters/$CLUSTER_ID/reimport" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"kubeConfig\": $(echo "$KUBECONFIG_CONTENT" | jq -Rs .)
        }")
    
    code=$(echo $response | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        version=$(echo $response | jq -r '.data.version')
        import_status=$(echo $response | jq -r '.data.importStatus')
        cluster_status=$(echo $response | jq -r '.data.clusterStatus')
        
        print_success "重新导入成功"
        echo "  新版本: $version"
        echo "  导入状态: $import_status"
        echo "  集群状态: $cluster_status"
    else
        msg=$(echo $response | jq -r '.msg')
        print_error "重新导入失败: $msg"
    fi
}

# 测试获取集群详情
test_get_cluster_detail() {
    print_section "6. 获取集群详情"
    
    if [ -z "$CLUSTER_ID" ]; then
        print_info "跳过获取详情（无可用集群ID）"
        return
    fi
    
    response=$(curl -s -X GET "$API_BASE/k8s/clusters/$CLUSTER_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    code=$(echo $response | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        print_success "获取集群详情成功"
        echo "$response" | jq '.data | {
            id: .id,
            name: .name,
            description: .description,
            version: .version,
            importMethod: .importMethod,
            importStatus: .importStatus,
            clusterStatus: .clusterStatus,
            status: .status,
            createdAt: .createdAt,
            updatedAt: .updatedAt
        }'
    else
        print_error "获取详情失败"
    fi
}

# 测试更新集群
test_update_cluster() {
    print_section "7. 更新集群信息"
    
    if [ -z "$CLUSTER_ID" ]; then
        print_info "跳过更新（无可用集群ID）"
        return
    fi
    
    print_info "更新集群 $CLUSTER_ID 的描述..."
    
    response=$(curl -s -X PUT "$API_BASE/k8s/clusters/$CLUSTER_ID" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "description": "更新后的描述信息",
            "remark": "测试更新功能"
        }')
    
    code=$(echo $response | jq -r '.code')
    
    if [ "$code" = "200" ]; then
        print_success "集群更新成功"
    else
        msg=$(echo $response | jq -r '.msg')
        print_error "集群更新失败: $msg"
    fi
}

# 测试版本验证（低版本）
test_version_validation() {
    print_section "8. 测试版本验证（低版本拒绝）"
    
    print_info "尝试创建低版本集群（预期失败）..."
    
    # 注意：需要一个真实的低版本K8s集群KubeConfig
    print_info "此测试需要真实的K8s 1.22或更低版本集群"
    print_info "跳过此测试"
}

# 显示功能清单
show_features() {
    print_section "功能清单"
    
    echo "✓ 自动版本检测"
    echo "✓ 版本验证（>= 1.23）"
    echo "✓ 导入状态跟踪（pending/importing/success/failed）"
    echo "✓ 集群状态监控（healthy/unhealthy/unknown）"
    echo "✓ 重新导入KubeConfig"
    echo "✓ 健康检查"
    echo "✓ 完整字段展示（ID、名称、描述、版本、导入方式、状态、时间）"
    echo ""
}

# 主函数
main() {
    echo "=========================================="
    echo "K8s集群管理功能测试"
    echo "=========================================="
    echo ""
    
    # 检查依赖
    if ! command -v jq &> /dev/null; then
        print_error "需要安装 jq 命令行工具"
        echo "安装方法："
        echo "  macOS: brew install jq"
        echo "  Ubuntu: apt-get install jq"
        exit 1
    fi
    
    # 显示功能清单
    show_features
    
    # 执行测试
    get_token
    
    # 注意：以下测试需要真实的K8s集群KubeConfig
    print_info "注意：创建集群测试需要真实的KubeConfig配置"
    print_info "请手动替换脚本中的KUBECONFIG_CONTENT变量"
    echo ""
    
    # test_create_cluster_success
    test_get_cluster_list
    # test_health_check
    # test_reimport_kubeconfig
    # test_get_cluster_detail
    # test_update_cluster
    # test_version_validation
    
    print_section "测试完成"
    print_info "请访问 Swagger 文档查看完整API: http://localhost:8000/swagger/index.html"
}

# 运行主函数
main
