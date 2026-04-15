# CMDB 资产管理模块 - 第一阶段：核心管理

> **分阶段说明**：CMDB 模块分三阶段实现。
> - **第一阶段（本文档）**：主机管理、分组管理、凭据管理
> - [第二阶段](./2026-04-15-cmdb-phase2-terminal-design.md)：SSH Web 终端、终端审计
> - [第三阶段](./2026-04-15-cmdb-phase3-cloud-permission-design.md)：云账号管理、权限配置

## 概述

为 DevOps 运维平台新增 CMDB（Configuration Management Database）模块的第一阶段，提供核心的 IT 资产管理能力。

**本阶段包含三个子系统**：
1. **主机管理**：服务器资产录入、批量导入、连接测试、状态监控
2. **分组管理**：三级固定层级（业务→环境→地域）的主机分组
3. **凭据管理**：SSH 密码和密钥的统一管理、加密存储

所有功能通过统一的 RESTful API 提供服务，复用现有 Session + RBAC 权限体系进行鉴权。

## 架构方案

采用统一 CMDB 模块方案，所有功能放置在 `internal/modules/cmdb/` 下，按子系统分文件组织。与现有 k8s、user 等模块风格一致。

## 数据模型

### Host（主机）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| hostname | string(255) | 主机名 |
| ip | string(45) | 主 IP 地址（IPv4/IPv6） |
| port | int | SSH 端口，默认 22 |
| os_type | string(20) | 操作系统类型（linux/windows） |
| os_name | string(255) | 操作系统全名 |
| cpu_cores | int | CPU 核数 |
| memory_total | int | 内存总量（MB） |
| disk_total | int | 磁盘总量（GB） |
| status | string(20) | 状态：online/offline/unknown |
| credential_id | uint | 关联凭据 ID（外键） |
| group_id | uint | 所属分组 ID（第三级，外键） |
| cloud_account_id | uint | 来源云账号 ID（可选，外键） |
| cloud_instance_id | string(100) | 云实例 ID（可选） |
| labels | json | 自定义标签 |
| description | string(500) | 描述 |
| agent_version | string(50) | Agent 版本（预留） |
| last_active_at | timestamp | 最后活跃时间 |

索引：`uk_host_tenant_ip_port` (tenant_id, ip, port)、`idx_host_group` (group_id)、`idx_host_status` (status)

### HostGroup（分组）— 三级固定层级

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| name | string(100) | 分组名称 |
| level | int | 层级：1=业务、2=环境、3=地域/机房 |
| parent_id | uint | 父分组 ID（一级为 0） |
| sort_order | int | 排序 |

索引：`uk_group_tenant_parent_name` (tenant_id, parent_id, name)

层级约束：level=1 的 parent_id=0；level=2 的 parent 指向 level=1；level=3 的 parent 指向 level=2。删除分组时检查是否有子分组或关联主机。

### Credential（凭据）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| name | string(100) | 凭据名称 |
| type | string(20) | 类型：password/key |
| username | string(100) | SSH 用户名 |
| password | text | AES-256-GCM 加密后的密码 |
| private_key | text | AES-256-GCM 加密后的私钥 |
| passphrase | string(500) | 私钥密码（加密） |
| description | string(500) | 描述 |

索引：`uk_credential_tenant_name` (tenant_id, name)

安全策略：
- 凭据列表和详情 API 不返回 password、private_key、passphrase 字段
- 创建/更新时接收明文，入库前加密
- 使用时服务端解密，不传输到前端

<!-- 云账号管理模型 → 见第三阶段 -->

<!-- 终端会话模型 → 见第二阶段 -->

<!-- 主机权限模型 → 见第三阶段 -->

## API 设计

路由前缀：`/api/v1/cmdb/`，复用 SessionAuth 中间件 + RequirePermission 权限控制。

### 主机管理

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET | `/cmdb/hosts` | 主机列表（分页、按分组/状态/IP/标签筛选） | `cmdb:host:list` |
| GET | `/cmdb/hosts/stats` | 主机统计（总数、在线/离线、按分组分布） | `cmdb:host:list` |
| GET | `/cmdb/hosts/:id` | 主机详情 | `cmdb:host:get` |
| POST | `/cmdb/hosts` | 创建主机 | `cmdb:host:create` |
| POST | `/cmdb/hosts/batch` | 批量导入主机（JSON 数组） | `cmdb:host:create` |
| PUT | `/cmdb/hosts/:id` | 更新主机 | `cmdb:host:update` |
| DELETE | `/cmdb/hosts/:id` | 删除主机 | `cmdb:host:delete` |
| POST | `/cmdb/hosts/:id/test` | 测试主机 SSH 连接 | `cmdb:host:test` |

### 分组管理

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET | `/cmdb/groups` | 分组树（三级嵌套结构） | `cmdb:group:list` |
| GET | `/cmdb/groups/:id` | 分组详情（含子分组和主机统计） | `cmdb:group:get` |
| POST | `/cmdb/groups` | 创建分组（校验层级约束） | `cmdb:group:create` |
| PUT | `/cmdb/groups/:id` | 更新分组 | `cmdb:group:update` |
| DELETE | `/cmdb/groups/:id` | 删除分组（检查是否有子分组/主机） | `cmdb:group:delete` |
| GET | `/cmdb/groups/:id/hosts` | 分组下主机列表 | `cmdb:host:list` |

### 凭据管理

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET | `/cmdb/credentials` | 凭据列表（不返回敏感字段） | `cmdb:credential:list` |
| GET | `/cmdb/credentials/:id` | 凭据详情（不返回敏感字段） | `cmdb:credential:get` |
| POST | `/cmdb/credentials` | 创建凭据（自动加密） | `cmdb:credential:create` |
| PUT | `/cmdb/credentials/:id` | 更新凭据 | `cmdb:credential:update` |
| DELETE | `/cmdb/credentials/:id` | 删除凭据（检查是否被主机引用） | `cmdb:credential:delete` |
| POST | `/cmdb/credentials/:id/test` | 测试凭据（连接指定主机验证） | `cmdb:credential:test` |

<!-- Web 终端 API → 见第二阶段 -->

<!-- 云账号管理 API → 见第三阶段 -->

<!-- 主机权限 API → 见第三阶段 -->

## 凭据加密方案

使用 AES-256-GCM 对称加密，密钥配置在 `config.yaml` 的 `crypto.aes_key` 字段（32 字节 hex 编码）。

- **加密流程**：明文 → AES-256-GCM 加密（随机 nonce） → base64 编码 → 存入数据库
- **解密流程**：base64 解码 → 解析 nonce + ciphertext → AES-256-GCM 解密 → 明文
- **密钥管理**：密钥配置在服务端，不入数据库，不通过网络传输
- **API 安全**：凭据的 CRUD 接口永远不返回加密字段（password、private_key、passphrase）

<!-- SSH Web 终端 → 见第二阶段 -->

<!-- 云账号同步 → 见第三阶段 -->

<!-- 权限模型（主机权限） → 见第三阶段 -->

## 前端设计

### 路由

```javascript
// 第一阶段 CMDB 路由
{ path: '/cmdb/hosts', component: HostList }
{ path: '/cmdb/groups', component: GroupList }
{ path: '/cmdb/credentials', component: CredentialList }

// 第二阶段新增
// { path: '/cmdb/terminal', component: TerminalList }
// { path: '/cmdb/terminal/:id/replay', component: TerminalReplay }

// 第三阶段新增
// { path: '/cmdb/cloud-accounts', component: CloudAccountList }
// { path: '/cmdb/permissions', component: PermissionList }
```

### API 模块

```
frontend/src/api/cmdb/
├── host.js           # 主机管理 API
├── group.js          # 分组管理 API
└── credential.js     # 凭据管理 API
# 第二阶段新增：terminal.js
# 第三阶段新增：cloudAccount.js, permission.js
```

### 侧边栏导航

在 MainLayout.vue 中新增 CMDB 菜单分组：

```
资产管理（CMDB）
├── 主机管理    /cmdb/hosts
├── 分组管理    /cmdb/groups
└── 凭据管理    /cmdb/credentials
# 第二阶段新增：├── 终端审计    /cmdb/terminal
# 第三阶段新增：├── 云账号      /cmdb/cloud-accounts
#              └── 权限配置    /cmdb/permissions
```

### 复用现有组件

- 复用 `frontend/src/components/K8s/Terminal.vue` 的 xterm.js 封装模式
- 复用 Element Plus 组件（ElTable、ElTree、ElForm 等）
- 复用 MainLayout.vue 的布局框架

## 后端目录结构

```
backend/internal/modules/cmdb/
├── model/
│   ├── host.go
│   ├── group.go
│   └── credential.go
# 第二阶段新增：├── terminal.go
# 第三阶段新增：├── cloud_account.go, cloud_resource.go, permission.go
├── repository/
│   ├── host.go
│   ├── group.go
│   └── credential.go
# 第二阶段新增：├── terminal.go
# 第三阶段新增：├── cloud_account.go, cloud_resource.go, permission.go
├── service/
│   ├── host.go           # 含批量导入、连接测试、统计
│   ├── group.go          # 含层级校验
│   └── credential.go     # 含加解密（复用 utils.Crypto）
# 第二阶段新增：├── terminal.go
# 第三阶段新增：├── cloud_sync.go, permission.go
├── api/
│   ├── host.go
│   ├── group.go
│   └── credential.go
# 第二阶段新增：├── terminal.go (WebSocket)
# 第三阶段新增：├── cloud_account.go, permission.go
# 第二阶段新增：terminal/ 子目录（ssh.go, recorder.go, replay.go）
```

## 依赖新增

### 后端 Go 依赖

**第一阶段无新增依赖**（复用现有的 `golang.org/x/crypto` 进行加解密）
**第二阶段新增**：`golang.org/x/crypto/ssh` — SSH 客户端
**第三阶段新增**：腾讯云 SDK

### 前端依赖

- 无新增（Element Plus 已有）

## 配置项

```yaml
# 第一阶段：复用现有 crypto.secret 配置，无需新增

# 第二阶段新增
# terminal:
#   recording_dir: "./data/recordings"
#   max_session_duration: 86400

# 第三阶段新增
# cloud:
#   sync_concurrency: 5
```
