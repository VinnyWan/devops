# K8s集群管理 - 快速开始

## 🎯 功能概述

本文档帮助你快速上手K8s集群管理的新功能，包括：
- ✅ 自动版本检测（要求 >= 1.23）
- ✅ 实时健康监控
- ✅ 导入状态跟踪
- ✅ 支持重新导入配置

## 📦 准备工作

### 1. 确保服务运行

```bash
# 启动服务
./devops

# 或使用指定配置
./devops -c config.yaml
```

服务启动后访问：http://localhost:8000

### 2. 准备KubeConfig

确保你有一个K8s集群的KubeConfig文件，要求：
- K8s版本 >= 1.23
- 包含有效的认证信息
- 网络可达

### 3. 获取认证Token

```bash
# 方法1: 使用脚本
./scripts/get_token.sh

# 方法2: 手动登录
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

## 🚀 快速上手

### 步骤1: 添加K8s集群

#### 使用Swagger UI（推荐）

1. 访问：http://localhost:8000/swagger/index.html
2. 找到 `K8s-Cluster` 标签
3. 点击 `POST /api/k8s/clusters`
4. 点击 "Try it out"
5. 填写信息：

```json
{
  "name": "我的测试集群",
  "description": "开发环境K8s集群",
  "apiServer": "https://192.168.1.100:6443",
  "kubeConfig": "apiVersion: v1\nkind: Config\nclusters:\n...",
  "deptId": 1,
  "remark": "测试用"
}
```

6. 点击 "Execute"

#### 使用curl命令

```bash
export TOKEN="your_token_here"

curl -X POST http://localhost:8000/api/k8s/clusters \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "我的测试集群",
    "description": "开发环境K8s集群",
    "apiServer": "https://192.168.1.100:6443",
    "kubeConfig": "apiVersion: v1\nkind: Config...",
    "deptId": 1,
    "remark": "测试用"
  }'
```

### 步骤2: 查看结果

#### 成功响应

```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {
    "id": 1,
    "name": "我的测试集群",
    "description": "开发环境K8s集群",
    "version": "v1.28.4",          // ← 自动获取
    "importMethod": "kubeconfig",   // ← 自动设置
    "importStatus": "success",      // ← 导入成功
    "clusterStatus": "healthy",     // ← 集群健康
    "status": 1,
    "apiServer": "https://192.168.1.100:6443",
    "deptId": 1,
    "remark": "测试用",
    "createdAt": "2026-01-10T09:00:00Z",
    "updatedAt": "2026-01-10T09:00:00Z"
  }
}
```

#### 版本不符合要求

```json
{
  "code": 400,
  "msg": "K8s版本不支持，要求 >= 1.23，当前版本: v1.20.0"
}
```

### 步骤3: 查看集群列表

```bash
curl -X GET "http://localhost:8000/api/k8s/clusters?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN"
```

响应包含完整字段：
- ✅ ID
- ✅ 集群名称
- ✅ 描述
- ✅ 版本号
- ✅ 导入方式
- ✅ 导入状态
- ✅ 集群状态
- ✅ 创建时间

### 步骤4: 健康检查

```bash
curl -X GET http://localhost:8000/api/k8s/clusters/1/health \
  -H "Authorization: Bearer $TOKEN"
```

### 步骤5: 重新导入配置（如需要）

当证书过期或配置变更时：

```bash
curl -X POST http://localhost:8000/api/k8s/clusters/1/reimport \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeConfig": "新的KubeConfig内容..."
  }'
```

## 💡 使用技巧

### 技巧1: 快速测试本地集群

如果你使用Minikube或Kind：

```bash
# 获取KubeConfig
minikube kubectl -- config view --flatten --minify

# 或
kind get kubeconfig --name=your-cluster
```

### 技巧2: 批量操作

使用脚本批量添加集群：

```bash
#!/bin/bash
TOKEN="your_token"

# 集群列表
clusters=(
  "prod:https://prod.k8s.com:6443:生产集群"
  "dev:https://dev.k8s.com:6443:开发集群"
  "test:https://test.k8s.com:6443:测试集群"
)

for cluster in "${clusters[@]}"; do
  IFS=':' read -r name server desc <<< "$cluster"
  
  curl -X POST http://localhost:8000/api/k8s/clusters \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"name\": \"$name\",
      \"description\": \"$desc\",
      \"apiServer\": \"$server\",
      \"kubeConfig\": \"$(cat configs/${name}.yaml)\",
      \"deptId\": 1
    }"
done
```

### 技巧3: 监控集群状态

定期检查所有集群健康状态：

```bash
#!/bin/bash
TOKEN="your_token"

# 获取所有集群
clusters=$(curl -s -X GET http://localhost:8000/api/k8s/clusters \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data.list[].id')

# 检查每个集群
for id in $clusters; do
  echo "检查集群 $id..."
  curl -s -X GET http://localhost:8000/api/k8s/clusters/$id/health \
    -H "Authorization: Bearer $TOKEN" | jq '.data'
done
```

## 🔍 故障排查

### 问题1: 版本检测失败

**症状**: 提示"获取集群版本失败"

**可能原因**:
- 网络不通
- API Server地址错误
- 证书过期
- 认证信息错误

**解决方法**:
```bash
# 1. 检查网络连通性
curl -k https://your-k8s-apiserver:6443

# 2. 验证KubeConfig
kubectl --kubeconfig=your-config.yaml cluster-info

# 3. 查看详细错误
# 在Swagger UI中查看响应的message字段
```

### 问题2: 版本太低

**症状**: "K8s版本不支持，要求 >= 1.23"

**解决方法**:
- 升级K8s集群到1.23+
- 或修改源码降低版本要求（不推荐）

### 问题3: 导入状态为failed

**症状**: importStatus显示为"failed"

**排查步骤**:
1. 查看集群详情获取错误信息
2. 检查KubeConfig格式
3. 尝试健康检查接口
4. 重新导入配置

```bash
# 获取详细信息
curl -X GET http://localhost:8000/api/k8s/clusters/1 \
  -H "Authorization: Bearer $TOKEN" | jq .

# 重新导入
curl -X POST http://localhost:8000/api/k8s/clusters/1/reimport \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"kubeConfig": "..."}'
```

## 📖 完整文档

更详细的信息请参考：

- [K8s集群管理完整指南](./K8S_CLUSTER_GUIDE.md) - 完整的功能说明和API文档
- [实现总结文档](./K8S_IMPLEMENTATION_SUMMARY.md) - 技术实现细节
- [Swagger API文档](http://localhost:8000/swagger/index.html) - 在线API测试

## ❓ 常见问题

**Q: 支持哪些K8s版本？**  
A: 要求 >= 1.23.0，推荐使用1.23+的版本

**Q: KubeConfig会加密存储吗？**  
A: 当前版本明文存储，加密功能计划在后续版本实现

**Q: 可以修改已导入的集群版本吗？**  
A: 版本号是自动获取的，无法手动修改。如需更新，使用重新导入功能

**Q: 导入状态有哪些？**  
A: pending（待导入）、importing（导入中）、success（成功）、failed（失败）

**Q: 集群状态有哪些？**  
A: healthy（健康）、unhealthy（不健康）、unknown（未知）

**Q: 如何更新过期的证书？**  
A: 使用重新导入接口 `POST /api/k8s/clusters/{id}/reimport`

## 🎉 完成！

现在你已经可以：
- ✅ 添加K8s集群（自动检测版本）
- ✅ 查看完整的集群信息
- ✅ 监控集群健康状态
- ✅ 重新导入更新配置
- ✅ 管理多个K8s集群

开始体验吧！🚀
