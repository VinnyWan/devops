# K8s集群管理功能快速测试指南

## ✅ 已完成功能

### 1. 核心功能
- ✅ K8s集群管理（CRUD）
- ✅ 集群健康检查
- ✅ 基于角色的访问权限控制
- ✅ 权限类型：readonly（只读）、admin（管理员）
- ✅ 操作日志记录
- ✅ K8s权限验证中间件

### 2. 数据库表
- ✅ k8s_clusters - 集群表
- ✅ k8s_cluster_accesses - 访问权限表
- ✅ k8s_namespaces - 命名空间表
- ✅ k8s_operation_logs - 操作日志表

### 3. API接口（已实现）

#### 集群管理
```
POST   /api/k8s/clusters              # 创建集群
GET    /api/k8s/clusters              # 获取集群列表
GET    /api/k8s/clusters/:id          # 获取集群详情
PUT    /api/k8s/clusters/:id          # 更新集群（需权限）
DELETE /api/k8s/clusters/:id          # 删除集群（需权限）
GET    /api/k8s/clusters/:id/health   # 健康检查（需权限）
```

#### 权限管理
```
POST   /api/k8s/clusters/:id/access   # 配置访问权限
GET    /api/k8s/clusters/:id/access   # 获取权限列表
DELETE /api/k8s/clusters/:id/access/:accessId # 删除权限
```

## 🚀 快速测试

### 步骤1：登录获取Token

```bash
# 1. 获取验证码
curl http://localhost:8080/api/captcha

# 2. 登录（如果验证码已关闭，随便填）
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "captchaId": "test",
    "captchaCode": "test"
  }'

# 3. 保存返回的token
export TOKEN="你的token"
```

### 步骤2：创建K8s集群

```bash
curl -X POST http://localhost:8080/api/k8s/clusters \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试集群",
    "description": "用于测试的K8s集群",
    "apiServer": "https://kubernetes.default.svc",
    "kubeConfig": "apiVersion: v1\nkind: Config\n...",
    "version": "v1.28.0",
    "status": 1,
    "deptId": 1,
    "remark": "测试集群"
  }'
```

**注意**：kubeConfig需要是有效的Kubernetes配置文件内容

### 步骤3：获取集群列表

```bash
curl -X GET "http://localhost:8080/api/k8s/clusters?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN"
```

### 步骤4：配置访问权限

```bash
# 为角色ID=2配置只读权限
curl -X POST http://localhost:8080/api/k8s/clusters/1/access \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "roleId": 2,
    "accessType": "readonly",
    "namespaces": "[\"default\", \"dev\"]"
  }'

# 为角色ID=1配置管理员权限
curl -X POST http://localhost:8080/api/k8s/clusters/1/access \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "roleId": 1,
    "accessType": "admin",
    "namespaces": ""
  }'
```

### 步骤5：健康检查

```bash
curl -X GET http://localhost:8080/api/k8s/clusters/1/health \
  -H "Authorization: Bearer $TOKEN"
```

### 步骤6：验证权限控制

```bash
# 1. 使用只读角色的用户Token尝试删除集群（应该被拒绝）
curl -X DELETE http://localhost:8080/api/k8s/clusters/1 \
  -H "Authorization: Bearer $READONLY_TOKEN"

# 预期响应：
# {
#   "code": 403,
#   "msg": "只读权限，无法执行写操作"
# }

# 2. 使用admin角色的用户Token删除集群（应该成功）
curl -X DELETE http://localhost:8080/api/k8s/clusters/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## 📊 权限验证流程

```
1. 用户请求访问K8s资源
   ↓
2. JWT中间件验证用户身份
   ↓
3. K8sPermission中间件验证权限
   - 获取用户的所有角色
   - 查询角色对集群的访问权限
   - 确定最高权限（readonly/admin）
   - 检查操作类型（读/写）
   ↓
4. 权限验证通过
   ↓
5. 执行K8s操作
   ↓
6. 记录操作日志
```

## 🔐 权限类型说明

### readonly（只读）
**允许的操作**：
- get（查询）
- list（列表）
- watch（监视）

**禁止的操作**：
- create（创建）
- update（更新）
- delete（删除）
- patch（补丁）
- scale（扩缩容）
- restart（重启）

### admin（管理员）
**允许所有操作**

## 📝 在Swagger中测试

1. 访问：http://localhost:8080/swagger/index.html
2. 找到"K8s集群管理"标签
3. 先登录获取Token
4. 点击右上角🔒 Authorize，输入：`Bearer YOUR_TOKEN`
5. 测试各个接口

## 🎯 后续开发计划

### 第二阶段：基础资源管理
- [ ] Namespace管理
- [ ] Deployment管理
- [ ] Pod查看和日志
- [ ] Service管理

### 第三阶段：高级资源管理
- [ ] StatefulSet管理
- [ ] DaemonSet管理
- [ ] ConfigMap/Secret管理
- [ ] PV/PVC/StorageClass管理
- [ ] Node管理
- [ ] Event查看

### 第四阶段：WebShell终端
- [ ] WebSocket连接
- [ ] 容器终端交互
- [ ] 命令执行
- [ ] 多容器支持

## ⚠️ 注意事项

1. **KubeConfig安全**
   - 当前KubeConfig以明文存储
   - 建议生产环境使用AES加密
   - 查询时不返回KubeConfig内容

2. **权限控制**
   - 必须先给用户分配角色
   - 必须配置角色对集群的访问权限
   - 权限检查基于用户的所有角色

3. **操作审计**
   - 所有K8s操作都会记录日志
   - 包含用户、集群、操作类型、结果等信息

4. **集群连接**
   - 使用client-go连接K8s集群
   - 支持健康检查验证连接
   - 连接失败会返回详细错误信息

## 🐛 故障排查

### 问题1：创建集群时提示"KubeConfig验证失败"
**原因**：KubeConfig格式不正确
**解决**：确保kubeConfig是有效的YAML格式

### 问题2：权限检查时提示"用户没有分配角色"
**原因**：用户未关联任何角色
**解决**：使用 `/api/users/{id}/roles` 接口为用户分配角色

### 问题3：删除操作被拒绝"只读权限，无法执行写操作"
**原因**：用户只有readonly权限
**解决**：
1. 为用户的角色配置admin权限
2. 或者使用具有admin权限的用户

### 问题4：健康检查失败
**原因**：无法连接到K8s集群
**解决**：
1. 检查apiServer地址是否正确
2. 检查kubeConfig配置
3. 确认网络连通性

## 📚 相关文档

- [K8S_IMPLEMENTATION_PLAN.md](./K8S_IMPLEMENTATION_PLAN.md) - 完整实现方案
- [Kubernetes Client-Go文档](https://kubernetes.io/docs/reference/using-api/client-libraries/)
- [Swagger API文档](http://localhost:8080/swagger/index.html)
