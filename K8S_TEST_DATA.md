# K8s集群测试数据说明

## 概述

系统启动时会自动初始化4个K8s集群测试数据，可直接用于API功能测试，无需手动创建集群。

## 自动初始化的测试集群

### 1. 本地开发集群

**基本信息**:
- **ID**: 1（自动生成）
- **名称**: 本地开发集群
- **描述**: 本地K3s开发测试集群，用于功能测试和开发
- **API Server**: https://127.0.0.1:6443
- **版本**: v1.28.5+k3s1
- **导入方式**: kubeconfig
- **导入状态**: success ✅
- **集群状态**: unknown（需健康检查）
- **启用状态**: 启用
- **备注**: 系统自动初始化的测试集群，可直接用于API测试

**使用场景**: 
- 本地K3s/Minikube集群测试
- API功能验证
- 开发环境测试

### 2. 测试集群-1.27

**基本信息**:
- **ID**: 2（自动生成）
- **名称**: 测试集群-1.27
- **描述**: Kubernetes 1.27版本测试集群
- **API Server**: https://test-k8s-1-27.example.com:6443
- **版本**: v1.27.8
- **导入方式**: kubeconfig
- **导入状态**: failed ❌
- **集群状态**: unhealthy
- **启用状态**: 禁用
- **备注**: 模拟不可访问的集群，用于测试失败场景

**使用场景**:
- 测试失败状态处理
- 测试禁用集群逻辑
- 测试不健康集群展示

### 3. 生产集群-示例

**基本信息**:
- **ID**: 3（自动生成）
- **名称**: 生产集群-示例
- **描述**: 生产环境K8s集群示例（仅供展示）
- **API Server**: https://prod-k8s.example.com:6443
- **版本**: v1.29.0
- **导入方式**: kubeconfig
- **导入状态**: success ✅
- **集群状态**: healthy
- **启用状态**: 启用
- **备注**: 示例集群，展示完整的字段信息

**使用场景**:
- 展示完整的集群信息字段
- 测试正常集群的所有功能
- 演示健康的集群状态

### 4. 待导入集群

**基本信息**:
- **ID**: 4（自动生成）
- **名称**: 待导入集群
- **描述**: 正在导入中的集群
- **API Server**: https://pending-k8s.example.com:6443
- **版本**: （空）
- **导入方式**: kubeconfig
- **导入状态**: importing 🔄
- **集群状态**: unknown
- **启用状态**: 启用
- **备注**: 模拟导入过程中的集群状态

**使用场景**:
- 测试导入中的状态展示
- 测试版本号为空的情况
- 测试过渡状态处理

## 初始化时机

测试数据会在以下情况下自动初始化：

1. **首次启动服务** - 数据库为空时
2. **数据库重置后** - 删除数据后重新启动
3. **执行迁移时** - 运行 `AutoMigrate()` 且无用户数据

## 查看测试数据

### 方法1: 通过API接口

```bash
# 1. 获取Token
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token'

# 2. 查看集群列表
curl -X GET "http://localhost:8000/api/k8s/clusters?page=1&pageSize=10" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  | jq .
```

### 方法2: 通过Swagger UI

1. 访问: http://localhost:8000/swagger/index.html
2. 点击 `Authorize` 输入Token
3. 找到 `K8s-Cluster` 标签
4. 执行 `GET /api/k8s/clusters` 接口

### 方法3: 直接查询数据库

```sql
SELECT 
    id,
    name,
    version,
    import_method,
    import_status,
    cluster_status,
    status,
    created_at
FROM k8s_clusters
ORDER BY id;
```

## 预期返回数据示例

```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "本地开发集群",
        "description": "本地K3s开发测试集群，用于功能测试和开发",
        "version": "v1.28.5+k3s1",
        "importMethod": "kubeconfig",
        "importStatus": "success",
        "clusterStatus": "unknown",
        "status": 1,
        "apiServer": "https://127.0.0.1:6443",
        "deptId": 1,
        "remark": "系统自动初始化的测试集群，可直接用于API测试",
        "createdAt": "2026-01-10T09:30:00Z",
        "updatedAt": "2026-01-10T09:30:00Z"
      },
      {
        "id": 2,
        "name": "测试集群-1.27",
        "description": "Kubernetes 1.27版本测试集群",
        "version": "v1.27.8",
        "importMethod": "kubeconfig",
        "importStatus": "failed",
        "clusterStatus": "unhealthy",
        "status": 0,
        "apiServer": "https://test-k8s-1-27.example.com:6443",
        "deptId": 1,
        "remark": "模拟不可访问的集群，用于测试失败场景",
        "createdAt": "2026-01-10T09:30:00Z",
        "updatedAt": "2026-01-10T09:30:00Z"
      },
      {
        "id": 3,
        "name": "生产集群-示例",
        "description": "生产环境K8s集群示例（仅供展示）",
        "version": "v1.29.0",
        "importMethod": "kubeconfig",
        "importStatus": "success",
        "clusterStatus": "healthy",
        "status": 1,
        "apiServer": "https://prod-k8s.example.com:6443",
        "deptId": 1,
        "remark": "示例集群，展示完整的字段信息",
        "createdAt": "2026-01-10T09:30:00Z",
        "updatedAt": "2026-01-10T09:30:00Z"
      },
      {
        "id": 4,
        "name": "待导入集群",
        "description": "正在导入中的集群",
        "version": "",
        "importMethod": "kubeconfig",
        "importStatus": "importing",
        "clusterStatus": "unknown",
        "status": 1,
        "apiServer": "https://pending-k8s.example.com:6443",
        "deptId": 1,
        "remark": "模拟导入过程中的集群状态",
        "createdAt": "2026-01-10T09:30:00Z",
        "updatedAt": "2026-01-10T09:30:00Z"
      }
    ],
    "total": 4,
    "page": 1,
    "pageSize": 10
  }
}
```

## 测试用例覆盖

这4个测试集群覆盖了以下测试场景：

### 状态覆盖

| 状态类型 | 覆盖的集群 |
|---------|-----------|
| **导入状态-成功** | 本地开发集群、生产集群-示例 |
| **导入状态-失败** | 测试集群-1.27 |
| **导入状态-进行中** | 待导入集群 |
| **集群状态-健康** | 生产集群-示例 |
| **集群状态-不健康** | 测试集群-1.27 |
| **集群状态-未知** | 本地开发集群、待导入集群 |
| **启用状态-启用** | 本地开发集群、生产集群-示例、待导入集群 |
| **启用状态-禁用** | 测试集群-1.27 |

### 版本覆盖

| 版本范围 | 集群 |
|---------|------|
| **1.27.x** | 测试集群-1.27 |
| **1.28.x** | 本地开发集群 |
| **1.29.x** | 生产集群-示例 |
| **空版本** | 待导入集群 |

## 使用测试数据进行API测试

### 测试1: 获取健康集群

```bash
# 获取ID为3的集群（健康状态）
curl -X GET http://localhost:8000/api/k8s/clusters/3 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 测试2: 测试健康检查

```bash
# 对不同状态的集群进行健康检查
curl -X GET http://localhost:8000/api/k8s/clusters/1/health \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 测试3: 测试重新导入

```bash
# 对失败的集群重新导入
curl -X POST http://localhost:8000/api/k8s/clusters/2/reimport \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"kubeConfig": "新的配置..."}'
```

### 测试4: 测试更新集群

```bash
# 更新集群描述
curl -X PUT http://localhost:8000/api/k8s/clusters/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "更新后的描述",
    "remark": "测试更新功能"
  }'
```

### 测试5: 测试筛选功能

```bash
# 按状态筛选（如果实现了筛选功能）
curl -X GET "http://localhost:8000/api/k8s/clusters?importStatus=success" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 测试数据的特点

### ✅ 优点

1. **开箱即用** - 无需手动创建，启动即可测试
2. **状态完整** - 覆盖所有可能的状态组合
3. **真实场景** - 模拟实际使用中的各种情况
4. **易于识别** - 命名清晰，描述详细
5. **不影响生产** - 使用示例域名，不会误操作真实集群

### ⚠️ 注意事项

1. **仅供测试** - 这些是测试数据，不包含真实的集群连接
2. **KubeConfig无效** - 配置文件是示例格式，无法真实连接
3. **健康检查会失败** - 由于集群不存在，健康检查会返回失败
4. **可以删除** - 如不需要可手动删除，不影响系统功能

## 清理测试数据

如果需要清理测试数据：

### 方法1: 通过API删除

```bash
# 删除单个集群
curl -X DELETE http://localhost:8000/api/k8s/clusters/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 方法2: 直接操作数据库

```sql
-- 删除所有测试集群
DELETE FROM k8s_clusters WHERE id IN (1, 2, 3, 4);

-- 重置自增ID（可选）
ALTER TABLE k8s_clusters AUTO_INCREMENT = 1;
```

### 方法3: 重置整个数据库

```bash
# 停止服务
# 删除数据库
mysql -u root -p -e "DROP DATABASE devops;"
# 重新创建
mysql -u root -p -e "CREATE DATABASE devops CHARACTER SET utf8mb4;"
# 重启服务，会自动重新初始化
```

## 日志查看

启动服务时，可在日志中看到初始化过程：

```
2026-01-10T09:30:00.000+0800    INFO    database/init.go:203    初始化 K8s 测试集群数据...
2026-01-10T09:30:00.001+0800    INFO    database/init.go:283    创建测试集群成功    {"ID": 1, "名称": "本地开发集群", "版本": "v1.28.5+k3s1", "导入状态": "success", "集群状态": "unknown"}
2026-01-10T09:30:00.002+0800    INFO    database/init.go:283    创建测试集群成功    {"ID": 2, "名称": "测试集群-1.27", "版本": "v1.27.8", "导入状态": "failed", "集群状态": "unhealthy"}
2026-01-10T09:30:00.003+0800    INFO    database/init.go:283    创建测试集群成功    {"ID": 3, "名称": "生产集群-示例", "版本": "v1.29.0", "导入状态": "success", "集群状态": "healthy"}
2026-01-10T09:30:00.004+0800    INFO    database/init.go:283    创建测试集群成功    {"ID": 4, "名称": "待导入集群", "版本": "", "导入状态": "importing", "集群状态": "unknown"}
2026-01-10T09:30:00.005+0800    INFO    database/init.go:291    K8s 测试集群初始化完成    {"集群数量": 4}
```

## 常见问题

### Q1: 为什么健康检查都失败？
A: 因为这些是测试数据，使用的是示例KubeConfig，并没有真实的K8s集群。如需测试真实连接，请添加真实的集群。

### Q2: 可以修改这些测试数据吗？
A: 可以，通过API或直接修改数据库都可以。但重置数据库后会恢复到初始状态。

### Q3: 测试数据会影响生产使用吗？
A: 不会。这些只是示例数据，可以随时删除或替换为真实集群。

### Q4: 如何添加更多测试集群？
A: 修改 `internal/database/init.go` 中的 `testClusters` 数组，添加更多集群定义。

### Q5: 能否禁用自动初始化？
A: 可以。注释掉 `InitData()` 函数中调用 `initK8sTestClusters()` 的代码即可。

## 总结

- ✅ 4个测试集群自动初始化
- ✅ 覆盖所有状态组合
- ✅ 开箱即用，无需配置
- ✅ 可直接用于API测试
- ✅ 提供完整的测试用例数据

**立即开始测试！** 🚀

访问 http://localhost:8000/swagger/index.html 查看所有可用的API接口。
