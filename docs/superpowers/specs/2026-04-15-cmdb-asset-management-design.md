# CMDB 资产管理模块设计

## 概述

为 DevOps 运维平台新增 CMDB（Configuration Management Database）模块，提供完整的 IT 资产管理能力。模块包含 6 个子系统：主机管理、分组管理、凭据管理、Web 终端（实时 SSH + 审计回放）、云账号管理、权限配置。

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
| sync_interval | int | 同步间隔（分钟），默认 60 |
| description | string(500) | 描述 |

索引：`uk_cloud_tenant_provider_name` (tenant_id, provider, name)

### CloudResource（云资源）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| cloud_account_id | uint | 云账号 ID |
| resource_type | string(30) | 资源类型：cvm/vpc/subnet/security_group/cbs |
| resource_id | string(100) | 云资源 ID |
| region | string(50) | 地域 |
| name | string(200) | 资源名称 |
| state | string(30) | 资源状态 |
| spec | json | 资源规格详情 |
| synced_at | timestamp | 同步时间 |

索引：`uk_cloud_res` (cloud_account_id, resource_type, resource_id)

### TerminalSession（终端会话）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| user_id | uint | 操作用户 ID |
| username | string(100) | 用户名（冗余） |
| host_id | uint | 连接的主机 ID |
| host_ip | string(45) | 主机 IP（冗余） |
| credential_id | uint | 使用的凭据 ID |
| client_ip | string(45) | 客户端 IP |
| started_at | timestamp | 开始时间 |
| finished_at | timestamp | 结束时间 |
| duration | int | 持续时间（秒） |
| recording_path | string(500) | 录像文件路径（asciinema v2 格式） |
| file_size | int64 | 录像文件大小（字节） |
| status | string(20) | 状态：active/closed/interrupted |

索引：`idx_terminal_user` (user_id)、`idx_terminal_host` (host_id)、`idx_terminal_time` (started_at)

### HostPermission（主机权限）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| user_id | uint | 用户 ID |
| host_group_id | uint | 授权的分组 ID |
| permission | string(20) | 权限：view/terminal/admin |
| created_by | uint | 创建者用户 ID |

索引：`uk_perm_user_group` (tenant_id, user_id, host_group_id, permission)

权限继承：授权到父级分组时，子分组及其中主机自动继承权限。admin 权限包含 view + terminal。

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

### Web 终端

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET (WS) | `/cmdb/terminal/connect` | SSH 终端 WebSocket 连接，参数：host_id | `cmdb:terminal:connect` |
| GET | `/cmdb/terminal/sessions` | 终端会话列表（按用户/主机/时间筛选） | `cmdb:terminal:list` |
| GET | `/cmdb/terminal/sessions/:id` | 会话详情 | `cmdb:terminal:get` |
| GET | `/cmdb/terminal/sessions/:id/recording` | 获取会话录像（asciinema v2 格式） | `cmdb:terminal:replay` |

### 云账号管理

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET | `/cmdb/cloud-accounts` | 云账号列表 | `cmdb:cloud:list` |
| GET | `/cmdb/cloud-accounts/:id` | 账号详情 | `cmdb:cloud:get` |
| POST | `/cmdb/cloud-accounts` | 添加云账号 | `cmdb:cloud:create` |
| PUT | `/cmdb/cloud-accounts/:id` | 更新云账号 | `cmdb:cloud:update` |
| DELETE | `/cmdb/cloud-accounts/:id` | 删除云账号 | `cmdb:cloud:delete` |
| POST | `/cmdb/cloud-accounts/:id/sync` | 手动触发同步 | `cmdb:cloud:sync` |
| GET | `/cmdb/cloud-accounts/:id/resources` | 同步的资源列表（按类型筛选） | `cmdb:cloud:list` |

### 主机权限

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET | `/cmdb/permissions` | 权限规则列表（按用户/分组筛选） | `cmdb:permission:list` |
| POST | `/cmdb/permissions` | 授予权限 | `cmdb:permission:create` |
| PUT | `/cmdb/permissions/:id` | 更新权限 | `cmdb:permission:update` |
| DELETE | `/cmdb/permissions/:id` | 删除权限 | `cmdb:permission:delete` |
| GET | `/cmdb/permissions/my-hosts` | 当前用户可访问的主机列表 | 登录即可访问 |

## 凭据加密方案

使用 AES-256-GCM 对称加密，密钥配置在 `config.yaml` 的 `crypto.aes_key` 字段（32 字节 hex 编码）。

- **加密流程**：明文 → AES-256-GCM 加密（随机 nonce） → base64 编码 → 存入数据库
- **解密流程**：base64 解码 → 解析 nonce + ciphertext → AES-256-GCM 解密 → 明文
- **密钥管理**：密钥配置在服务端，不入数据库，不通过网络传输
- **API 安全**：凭据的 CRUD 接口永远不返回加密字段（password、private_key、passphrase）

## SSH Web 终端

### 实时终端流程

1. 前端通过 WebSocket 连接 `/api/v1/cmdb/terminal/connect?host_id=X`，携带 Session Cookie
2. 后端中间件校验 Session 有效性
3. 后端校验用户对目标主机的终端权限（HostPermission）
4. 查询主机信息和关联凭据，服务端解密密码/密钥
5. 通过 `golang.org/x/crypto/ssh` 建立到目标主机的 SSH 连接
6. 创建 TerminalSession 记录（status=active）
7. 双向数据转发：WebSocket ↔ SSH 连接
8. 同时开启终端录像录制（asciinema v2 格式，写入文件）
9. WebSocket 断开时：关闭 SSH 连接、停止录制、更新 TerminalSession（status=closed, finished_at, duration, file_size）

### 终端审计回放

- 录像格式采用 **asciinema v2** 标准格式（JSON Lines，每行包含时间戳和输出内容）
- 前端使用 xterm.js + 自建播放器渲染录像（或使用 asciinema-player 组件）
- 会话列表支持按用户、主机 IP、时间范围筛选
- 录像文件存储在服务端本地磁盘，路径配置在 `config.yaml: terminal.recording_dir`

## 云账号同步

### 腾讯云同步流程

1. 解密 CloudAccount 的 SecretId/SecretKey
2. 使用腾讯云 SDK 调用以下 API：
   - `DescribeInstances` → 同步 CVM 实例到 Host 表（按 cloud_instance_id 匹配更新或新建）
   - `DescribeVpcs` → 同步 VPC 到 CloudResource 表
   - `DescribeSubnets` → 同步子网到 CloudResource 表
   - `DescribeSecurityGroups` → 同步安全组到 CloudResource 表
   - `DescribeVolumes` → 同步 CBS 云硬盘到 CloudResource 表
3. 更新 CloudAccount.last_sync_at
4. 同步失败时记录错误日志，更新 status=error

### 定时同步

- 使用 Go ticker 或 cron 实现定时同步，间隔由 CloudAccount.sync_interval 配置
- 支持手动触发同步（POST /cmdb/cloud-accounts/:id/sync）

## 权限模型

### CMDB 权限种子数据

在系统启动时，向 permissions 表注册以下权限：

| 资源 | 动作 | 说明 |
|---|---|---|
| cmdb:host | list/get/create/update/delete/test | 主机管理 |
| cmdb:group | list/get/create/update/delete | 分组管理 |
| cmdb:credential | list/get/create/update/delete/test | 凭据管理 |
| cmdb:terminal | connect/list/get/replay | Web 终端 |
| cmdb:cloud | list/get/create/update/delete/sync | 云账号 |
| cmdb:permission | list/create/update/delete | 权限配置 |

### 主机访问权限（HostPermission）

独立于系统 RBAC 的资产级权限控制，决定用户可以访问哪些主机：

- **view**：查看主机详情
- **terminal**：使用 Web 终端连接主机（隐含 view）
- **admin**：管理主机（修改、删除，隐含 view + terminal）

权限基于分组授权，支持继承：授权到一级分组时，其下所有二级、三级分组及主机都继承权限。

## 前端设计

### 路由

```javascript
// 新增 CMDB 路由
{ path: '/cmdb/hosts', component: HostList }
{ path: '/cmdb/groups', component: GroupList }
{ path: '/cmdb/credentials', component: CredentialList }
{ path: '/cmdb/terminal', component: TerminalList }
{ path: '/cmdb/terminal/:id/replay', component: TerminalReplay }
{ path: '/cmdb/cloud-accounts', component: CloudAccountList }
{ path: '/cmdb/permissions', component: PermissionList }
```

### API 模块

```
frontend/src/api/cmdb/
├── host.js           # 主机管理 API
├── group.js          # 分组管理 API
├── credential.js     # 凭据管理 API
├── terminal.js       # 终端 API
├── cloudAccount.js   # 云账号 API
└── permission.js     # 权限 API
```

### 侧边栏导航

在 MainLayout.vue 中新增 CMDB 菜单分组：

```
资产管理
├── 主机管理    /cmdb/hosts
├── 分组管理    /cmdb/groups
├── 凭据管理    /cmdb/credentials
├── 终端审计    /cmdb/terminal
├── 云账号      /cmdb/cloud-accounts
└── 权限配置    /cmdb/permissions
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
│   ├── credential.go
│   ├── cloud_account.go
│   ├── cloud_resource.go
│   ├── terminal.go
│   └── permission.go
├── repository/
│   ├── host.go
│   ├── group.go
│   ├── credential.go
│   ├── cloud_account.go
│   ├── cloud_resource.go
│   ├── terminal.go
│   └── permission.go
├── service/
│   ├── host.go           # 含批量导入、连接测试、统计
│   ├── group.go          # 含层级校验
│   ├── credential.go     # 含加解密
│   ├── cloud_sync.go     # 腾讯云同步
│   ├── terminal.go       # 终端会话管理
│   └── permission.go     # 权限检查
├── api/
│   ├── host.go
│   ├── group.go
│   ├── credential.go
│   ├── cloud_account.go
│   ├── terminal.go       # WebSocket 升级 + SSH
│   └── permission.go
├── terminal/
│   ├── ssh.go            # SSH 连接池管理
│   ├── recorder.go       # asciinema v2 录制
│   └── replay.go         # 录像读取
└── crypto/
    └── aes.go            # AES-256-GCM 加解密
```

## 依赖新增

### 后端 Go 依赖

- `golang.org/x/crypto/ssh` — SSH 客户端
- `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common` — 腾讯云 SDK 基础
- `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm` — CVM API
- `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc` — VPC API
- `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs` — CBS API

### 前端依赖

- 无新增主要依赖（xterm.js 已有，Element Plus 已有）

## 配置项

```yaml
# config.yaml 新增
crypto:
  aes_key: "32-byte-hex-encoded-key"  # AES-256 密钥

terminal:
  recording_dir: "./data/recordings"  # 终端录像存储目录
  max_session_duration: 86400          # 最大会话时长（秒），默认 24 小时

cloud:
  sync_concurrency: 5                  # 并发同步数
```
