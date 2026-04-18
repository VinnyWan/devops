# CMDB 云账号管理 - 实现规格

> **依赖前提**：Phase 1（主机管理、分组管理、凭据管理）已完成，Phase 2（终端审计）已完成，Phase 3 主机权限配置已完成。

## 概述

为 CMDB 模块增加腾讯云账号接入能力，支持手动/定时同步云资源（CVM、VPC、子网、安全组、CBS），并将 CVM 实例自动关联到 Host 记录。

## 功能范围

1. **云账号 CRUD**：添加、编辑、删除腾讯云账号（SecretId/SecretKey 加密存储）
2. **手动同步**：点击按钮触发一次完整同步
3. **定时同步**：按账号配置的间隔自动同步
4. **云资源展示**：查看同步的 CVM/VPC/子网/安全组/CBS 资源
5. **CVM 自动关联**：同步的 CVM 按 cloud_instance_id 匹配 Host 表，存在则更新，不存在则创建

## 数据模型

### CloudAccount（云账号）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| name | string(100) | 账号名称 |
| provider | string(20) | 云厂商：tencent |
| secret_id | string(500) | API 密钥 ID（AES-256 加密存储） |
| secret_key | string(500) | API 密钥 Key（AES-256 加密存储） |
| status | string(20) | 状态：active/error |
| last_sync_at | timestamp | 最后同步时间 |
| last_sync_error | text | 最后同步错误信息 |
| sync_interval | int | 同步间隔（分钟），默认 60 |
| description | string(500) | 描述 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除 |

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
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | timestamp | 软删除 |

索引：
- `uk_cloud_res` (cloud_account_id, resource_type, resource_id)
- `idx_cloud_res_type` (resource_type)
- `idx_cloud_res_region` (region)

## API 设计

路由前缀：`/api/v1/cmdb/cloud-accounts/`

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET | `/list` | 云账号列表（分页） | `cmdb:cloud:list` |
| GET | `/detail` | 账号详情，参数：id | `cmdb:cloud:get` |
| POST | `/create` | 添加云账号 | `cmdb:cloud:create` |
| POST | `/update` | 更新云账号 | `cmdb:cloud:update` |
| POST | `/delete` | 删除云账号 | `cmdb:cloud:delete` |
| POST | `/:id/sync` | 手动触发同步 | `cmdb:cloud:sync` |
| GET | `/:id/resources` | 同步的云资源列表（按类型筛选） | `cmdb:cloud:list` |

### 安全约束

- secret_id/secret_key 在 API 响应中不返回（`json:"-"`）
- 存储使用 `utils.Encrypt()` / `utils.Decrypt()`（AES-256，项目已有）
- 更新时如果未传 secret 字段，保留原值

## 同步流程

### 腾讯云同步

```
触发同步（手动/定时）
    ↓
解密 SecretId/SecretKey
    ↓
调用腾讯云 API（按地域遍历）：
├── DescribeInstances → 同步 CVM
│   └── 按 cloud_instance_id 匹配 Host 表：
│       存在 → 更新 hostname/ip/os_type/os_name
│       不存在 → 创建 Host 记录，设置 cloud_instance_id
├── DescribeVpcs → upsert 到 CloudResource (type=vpc)
├── DescribeSubnets → upsert 到 CloudResource (type=subnet)
├── DescribeSecurityGroups → upsert 到 CloudResource (type=security_group)
└── DescribeVolumes → upsert 到 CloudResource (type=cbs)
    ↓
更新 CloudAccount.last_sync_at = now()
    ↓
同步失败：记录 last_sync_error，设置 status=error
同步成功：清除 last_sync_error，设置 status=active
```

### 同步映射

| 腾讯云资源 | CMDB 目标 | 字段映射 |
|---|---|---|
| CVM 实例 | Host 表 | hostname=InstanceName, ip=PrivateIp, cloud_instance_id=InstanceId, os_type=OsName, os_name=OsName |
| CVM 实例 | CloudResource (type=cvm) | resource_id=InstanceId, name=InstanceName, state=InstanceState, spec={cpu,memory,zone} |
| VPC | CloudResource (type=vpc) | resource_id=VpcId, name=VpcName, spec={cidr,is_default} |
| 子网 | CloudResource (type=subnet) | resource_id=SubnetId, name=SubnetName, spec={cidr,vpc_id} |
| 安全组 | CloudResource (type=security_group) | resource_id=SecurityGroupId, spec={rules} |
| 云硬盘 | CloudResource (type=cbs) | resource_id=VolumeId, spec={size,type,disk_type,state} |

### 定时同步

- 使用 Go `time.Ticker` 实现
- 每个账号按 `sync_interval` 独立间隔
- 服务启动时从数据库加载所有 active 账号，启动对应 ticker
- 并发控制：最多 `sync_concurrency` 个同时同步

## 后端目录结构

```
backend/internal/modules/cmdb/
├── model/
│   ├── cloud_account.go      # CloudAccount 模型
│   └── cloud_resource.go     # CloudResource 模型
├── repository/
│   └── cloud.go              # 云账号/资源 CRUD + upsert
├── service/
│   └── cloud_sync.go         # 同步逻辑 + 定时任务 + 权限检查
└── api/
    └── cloud.go              # HTTP Handler
```

## 前端设计

### 页面路由

```javascript
{ path: '/cmdb/cloud-accounts', component: CloudAccountList }
```

### CloudAccountList.vue

- 表格列：账号名称、云厂商、状态（标签）、最后同步时间、同步间隔、操作
- 筛选：状态（全部/正常/错误）
- 操作列：[同步] [编辑] [查看资源] [删除]
- 添加/编辑弹窗：名称、SecretId、SecretKey（密码输入）、同步间隔、描述
  - 编辑时 SecretId/SecretKey 显示为 "******"，未修改则不传

### CloudResources.vue（资源弹窗）

- 从 CloudAccountList 点击"查看资源"时打开
- 标签页切换资源类型：CVM / VPC / 子网 / 安全组 / CBS
- 表格列：资源 ID、名称、地域、状态、同步时间

### 侧边栏

在 MainLayout.vue 的 CMDB 菜单中新增：
```html
<el-menu-item index="/cmdb/cloud-accounts">云账号</el-menu-item>
```

## 权限种子数据

```go
// 云账号管理
{Name: "查看云账号", Resource: "cmdb:cloud", Action: "list", Description: "查看云账号列表"},
{Name: "查看云账号详情", Resource: "cmdb:cloud", Action: "get", Description: "查看云账号详情"},
{Name: "添加云账号", Resource: "cmdb:cloud", Action: "create", Description: "添加云账号"},
{Name: "更新云账号", Resource: "cmdb:cloud", Action: "update", Description: "更新云账号"},
{Name: "删除云账号", Resource: "cmdb:cloud", Action: "delete", Description: "删除云账号"},
{Name: "同步云资源", Resource: "cmdb:cloud", Action: "sync", Description: "手动触发云资源同步"},
```

## Go 依赖

```bash
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs
```

## 配置项

```yaml
# config.yaml 新增
cloud:
  sync_concurrency: 5    # 并发同步数
  sync_timeout: 300      # 单次同步超时（秒）
```

## Host 模型扩展

在 Host 模型中新增 `CloudInstanceID` 字段：

```go
CloudInstanceID string `gorm:"size:100;index" json:"cloudInstanceId"`
```

用于关联腾讯云 CVM 实例 ID，同步时按此字段匹配。

## 不在本期范围

- 其他云厂商（AWS、阿里云等）— 后续通过 provider 扩展
- 云资源变更事件推送（依赖云厂商 webhook）
- 云资源拓扑图展示
