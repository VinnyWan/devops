# CMDB 资产管理模块 - 第三阶段：云账号管理与权限配置

> **依赖前提**：本阶段依赖第一阶段（主机管理、分组管理、凭据管理）已完成。

## 概述

为 CMDB 模块增加云账号接入能力和主机级别的细粒度权限控制。

## 功能范围

1. **云账号管理**：添加腾讯云账号，自动同步云资源
2. **云资源展示**：查看同步的 CVM、VPC、子网、安全组、CBS 等资源
3. **主机权限配置**：按分组为用户配置主机访问权限（view/terminal/admin）

## 数据模型

### CloudAccount（云账号）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| name | string(100) | 账号名称 |
| provider | string(20) | 云厂商：tencent |
| secret_id | string(500) | API 密钥 ID（AES-256 加密） |
| secret_key | string(500) | API 密钥 Key（AES-256 加密） |
| status | string(20) | 状态：active/error |
| last_sync_at | timestamp | 最后同步时间 |
| last_sync_error | text | 最后同步错误信息 |
| sync_interval | int | 同步间隔（分钟），默认 60 |
| description | string(500) | 描述 |

索引：
- `uk_cloud_tenant_provider_name` (tenant_id, provider, name)

### CloudResource（云资源）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| cloud_account_id | uint | 云账号 ID（外键） |
| resource_type | string(30) | 资源类型：cvm/vpc/subnet/security_group/cbs |
| resource_id | string(100) | 云资源 ID |
| region | string(50) | 地域 |
| zone | string(50) | 可用区 |
| name | string(200) | 资源名称 |
| state | string(30) | 资源状态 |
| spec | json | 资源规格详情 |
| synced_at | timestamp | 同步时间 |

索引：
- `uk_cloud_res` (cloud_account_id, resource_type, resource_id)
- `idx_cloud_res_type` (resource_type)
- `idx_cloud_res_region` (region)

### HostPermission（主机权限）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| user_id | uint | 用户 ID（外键） |
| host_group_id | uint | 授权的分组 ID（外键 → cmdb_host_groups） |
| permission | string(20) | 权限：view/terminal/admin |
| created_by | uint | 创建者用户 ID |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

索引：
- `uk_perm_user_group` (tenant_id, user_id, host_group_id, permission)
- `idx_perm_user` (user_id)
- `idx_perm_group` (host_group_id)

**权限继承规则**：
- 授权到父级分组时，子分组及其中主机自动继承该权限
- admin 权限包含 view + terminal
- terminal 权限包含 view

## API 设计

### 云账号管理

路由前缀：`/api/v1/cmdb/cloud-accounts/`

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET | `/list` | 云账号列表（分页） | `cmdb:cloud:list` |
| GET | `/detail` | 账号详情，参数：id | `cmdb:cloud:get` |
| POST | `/create` | 添加云账号 | `cmdb:cloud:create` |
| POST | `/update` | 更新云账号 | `cmdb:cloud:update` |
| POST | `/delete` | 删除云账号 | `cmdb:cloud:delete` |
| POST | `/:id/sync` | 手动触发同步 | `cmdb:cloud:sync` |
| GET | `/:id/resources` | 同步的资源列表（按类型筛选） | `cmdb:cloud:list` |
| GET | `/providers` | 支持的云厂商列表 | `cmdb:cloud:list` |

### 主机权限配置

路由前缀：`/api/v1/cmdb/permissions/`

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET | `/list` | 权限规则列表（按用户/分组筛选） | `cmdb:permission:list` |
| POST | `/create` | 授予权限 | `cmdb:permission:create` |
| POST | `/update` | 更新权限 | `cmdb:permission:update` |
| POST | `/delete` | 删除权限 | `cmdb:permission:delete` |
| GET | `/my-hosts` | 当前用户可访问的主机列表（含权限） | 登录即可 |
| GET | `/check` | 检查用户对主机的权限，参数：host_id, action | 登录即可 |

## 云账号同步流程

### 腾讯云同步（v1.0 支持）

```
添加云账号（SecretId/SecretKey 加密存储）
    ↓
手动/定时触发同步
    ↓
解密密钥（utils.Decrypt）
    ↓
调用腾讯云 API：
├── DescribeInstances → 同步 CVM 到 Host 表
│   └── 按 cloud_instance_id 匹配：存在则更新，不存在则创建
├── DescribeVpcs → 同步 VPC 到 CloudResource 表
├── DescribeSubnets → 同步子网到 CloudResource 表
├── DescribeSecurityGroups → 同步安全组到 CloudResource 表
└── DescribeVolumes → 同步 CBS 到 CloudResource 表
    ↓
更新 CloudAccount.last_sync_at
    ↓
同步失败时记录错误、更新 status=error
```

### 同步映射规则

| 腾讯云资源 | CMDB 模型 | 字段映射 |
|---|---|---|
| CVM 实例 | Host | hostname=InstanceName, ip=PrivateIP, cloud_instance_id=InstanceId |
| VPC | CloudResource | resource_type=vpc, spec={cidr, is_default} |
| 子网 | CloudResource | resource_type=subnet, spec={cidr, vpc_id} |
| 安全组 | CloudResource | resource_type=security_group, spec={rules} |
| 云硬盘 | CloudResource | resource_type=cbs, spec={size, type, status} |

### 定时同步

- 使用 Go time.Ticker 实现
- 间隔由 CloudAccount.sync_interval 配置（分钟）
- 支持手动触发同步

## 前端设计

### 页面路由

```javascript
{ path: '/cmdb/cloud-accounts', component: CloudAccountList }
{ path: '/cmdb/permissions', component: PermissionList }
```

### CloudAccountList.vue（云账号列表）

- 表格列：账号名称、云厂商、状态、最后同步时间、同步间隔、操作
- 筛选条件：云厂商（下拉）、状态（全部/正常/错误）
- 操作列：[同步] [编辑] [删除] [查看资源]
- 添加/编辑弹窗：名称、云厂商（下拉）、SecretId、SecretKey、同步间隔、描述

### CloudResources.vue（云资源列表，可选）

- 作为 CloudAccountList 的详情页或独立页
- 表格列：资源类型、资源 ID、名称、地域、状态、同步时间
- 筛选条件：资源类型（标签页）、地域（下拉）

### PermissionList.vue（权限配置）

- 左侧：分组树（复用 Phase 1 的 HostGroup 树）
- 右侧：权限规则列表
  - 列：用户、分组、权限（标签）、操作
- 筛选条件：用户（下拉）、分组（树选择）
- 授予权限弹窗：用户（下拉）、分组（树选择）、权限（多选：view/terminal/admin）
- 权限继承提示：显示授权影响的主机数量

## 后端目录结构

```
backend/internal/modules/cmdb/
├── model/
│   ├── cloud_account.go      # CloudAccount 模型
│   ├── cloud_resource.go     # CloudResource 模型
│   └── permission.go         # HostPermission 模型
├── repository/
│   ├── cloud_account.go      # 云账号 CRUD
│   ├── cloud_resource.go     # 云资源 CRUD
│   └── permission.go         # 权限 CRUD
├── service/
│   ├── cloud_sync.go         # 云同步逻辑
│   └── permission.go         # 权限检查逻辑
└── api/
    ├── cloud_account.go      # 云账号 API
    └── permission.go         # 权限 API
```

## 配置项

```yaml
# config.yaml 新增
cloud:
  sync_concurrency: 5    # 并发同步数
  sync_timeout: 300      # 单次同步超时（秒）
  providers:
    tencent:
      enabled: true
      endpoint: ""       # 可选：自定义 API 端点
```

## 依赖新增

### 后端 Go 依赖

```bash
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs
```

## 权限种子数据

在 `bootstrap/db.go` 的 `seedPermissions()` 中新增：

```go
// 云账号管理
{Name: "查看云账号", Resource: "cmdb:cloud", Action: "list", Description: "查看云账号列表"},
{Name: "查看云账号详情", Resource: "cmdb:cloud", Action: "get", Description: "查看云账号详情"},
{Name: "添加云账号", Resource: "cmdb:cloud", Action: "create", Description: "添加云账号"},
{Name: "更新云账号", Resource: "cmdb:cloud", Action: "update", Description: "更新云账号"},
{Name: "删除云账号", Resource: "cmdb:cloud", Action: "delete", Description: "删除云账号"},
{Name: "同步云资源", Resource: "cmdb:cloud", Action: "sync", Description: "手动触发云资源同步"},
// 权限配置
{Name: "查看权限配置", Resource: "cmdb:permission", Action: "list", Description: "查看主机权限配置"},
{Name: "授予权限", Resource: "cmdb:permission", Action: "create", Description: "授予主机权限"},
{Name: "更新权限", Resource: "cmdb:permission", Action: "update", Description: "更新主机权限"},
{Name: "删除权限", Resource: "cmdb:permission", Action: "delete", Description: "删除主机权限"},
```

## 侧边栏导航

在 MainLayout.vue 的 CMDB 菜单分组中新增：

```html
<el-menu-item index="/cmdb/cloud-accounts">云账号</el-menu-item>
<el-menu-item index="/cmdb/permissions">权限配置</el-menu-item>
```

## 权限检查逻辑

在主机相关 API 中增加权限校验：

```go
// 在 api/host.go 的各个 Handler 中
func checkHostPermission(c *gin.Context, hostID uint, action string) error {
    userID := c.GetUint("userID")
    tenantID := c.GetUint("tenantID")

    // 检查用户是否有该主机的权限
    hasPermission, err := permissionService.CheckHostPermission(
        context.Background(),
        tenantID,
        userID,
        hostID,
        action, // "view", "terminal", "admin"
    )
    if err != nil {
        return err
    }
    if !hasPermission {
        return errors.New("无权访问该主机")
    }
    return nil
}
```

权限检查逻辑（service/permission.go）：
1. 查询用户的所有 HostPermission 规则
2. 找到主机所属分组（含父级分组）
3. 检查是否有任一分组上的权限满足要求
4. 权限继承：admin ≥ terminal ≥ view

## 与现有模块的集成

- **复用加密工具**：云账号密钥使用 utils.Encrypt/Decrypt
- **复用主机管理**：同步的 CVM 自动创建 Host 记录
- **复用分组管理**：权限按分组授予，自动继承
- **复用用户模块**：权限配置关联用户表

## 安全考虑

1. **密钥加密存储**：SecretId/SecretKey 使用 AES-256 加密
2. **API 传输保护**：凭据相关 API 不返回加密字段
3. **权限隔离**：用户只能访问被授权的主机
4. **审计日志**：权限变更记录到审计日志
5. **同步限流**：控制并发同步数，避免过度消耗 API 配额

## 扩展性

未来支持其他云厂商时，只需：
1. 在 CloudAccount.provider 增加新值
2. 在 service/cloud_sync.go 中增加对应的同步逻辑
3. 更新前端云厂商下拉选项
