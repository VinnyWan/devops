# CMDB 资产管理模块 - 第二阶段：SSH Web 终端与审计

> **依赖前提**：本阶段依赖第一阶段（主机管理、分组管理、凭据管理）已完成。

## 概述

为 CMDB 模块增加 SSH Web 终端能力，使用户可以通过浏览器直接连接到已录入的服务器，并记录所有终端操作用于审计回放。

## 功能范围

1. **实时 SSH 终端**：通过 WebSocket 建立浏览器到服务器的 SSH 连接
2. **终端会话记录**：记录每次连接的元信息（用户、主机、时间、IP）
3. **操作录像录制**：以 asciinema v2 格式录制终端操作
4. **审计回放**：按用户、主机、时间筛选历史会话并回放录像

## 数据模型

### TerminalSession（终端会话）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| user_id | uint | 操作用户 ID |
| username | string(100) | 用户名（冗余，方便查询） |
| host_id | uint | 连接的主机 ID（外键 → cmdb_hosts） |
| host_ip | string(45) | 主机 IP（冗余） |
| credential_id | uint | 使用的凭据 ID |
| client_ip | string(45) | 客户端 IP |
| started_at | timestamp | 开始时间 |
| finished_at | timestamp | 结束时间（可为 NULL） |
| duration | int | 持续时间（秒） |
| recording_path | string(500) | 录像文件路径 |
| file_size | int64 | 录像文件大小（字节） |
| status | string(20) | 状态：active/closed/interrupted |

索引：
- `idx_terminal_user` (user_id)
- `idx_terminal_host` (host_id)
- `idx_terminal_time` (started_at)
- `idx_terminal_status` (status)

## API 设计

路由前缀：`/api/v1/cmdb/terminal/`

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| WS GET | `/connect` | SSH 终端 WebSocket 连接，参数：host_id | `cmdb:terminal:connect` |
| GET | `/sessions` | 终端会话列表（分页、筛选） | `cmdb:terminal:list` |
| GET | `/sessions/:id` | 会话详情 | `cmdb:terminal:get` |
| GET | `/sessions/:id/recording` | 获取会话录像文件 | `cmdb:terminal:replay` |

### WebSocket 连接协议

**连接 URL**：`ws://host/api/v1/cmdb/terminal/connect?host_id=X`

**客户端 → 服务端消息**（JSON）：
```json
{"operation": "stdin", "data": "ls\n"}
{"operation": "resize", "cols": 120, "rows": 30}
```

**服务端 → 客户端消息**（JSON）：
```json
{"operation": "stdout", "data": "total 48\r\n"}
{"operation": "stderr", "data": "error message\r\n"}
{"operation": "closed", "reason": "connection closed"}
```

## 核心流程

### 1. 建立终端连接

```
用户点击"连接终端"
    ↓
前端建立 WebSocket 连接（携带 Session Cookie）
    ↓
后端中间件校验 Session 有效性
    ↓
后端校验用户对目标主机的 cmdb:terminal:connect 权限
    ↓
查询主机信息（host_id）和关联凭据
    ↓
使用 utils.Decrypt() 解密凭据密码/私钥
    ↓
建立 SSH 连接（golang.org/x/crypto/ssh）
    ↓
创建 TerminalSession 记录（status=active）
    ↓
启动 asciinema v2 格式录像录制
    ↓
双向数据转发：WebSocket ↔ SSH
```

### 2. 录像格式（asciinema v2）

```json
{"version": 2, "width": 120, "height": 30, "timestamp": 1713136800.000}
[1.234, "o", "Welcome to Ubuntu\r\n"]
[2.456, "o", "user@host:~$ "]
[3.789, "i", "ls\r\n"]
[4.012, "o", "\r\ntotal 48\r\n"]
...
```

每行格式：`[elapsed_seconds, output_type, data]`
- `output_type`: "o" (output) 或 "i" (input)

### 3. 连接关闭

```
WebSocket 断开或用户主动关闭
    ↓
关闭 SSH 连接
    ↓
停止录像录制
    ↓
更新 TerminalSession：
    - status = closed
    - finished_at = now()
    - duration = finished_at - started_at
    - file_size = 录像文件大小
    ↓
返回给前端连接关闭消息
```

## 前端设计

### 页面路由

```javascript
{ path: '/cmdb/terminal', component: TerminalList }
{ path: '/cmdb/terminal/:id/replay', component: TerminalReplay }
```

### 组件复用

- 复用 `components/K8s/Terminal.vue` 的 xterm.js 封装
- 复用 Element Plus 的 ElTable、ElPagination、ElDatePicker 等组件

### TerminalList.vue（会话列表）

- 表格列：用户、主机 IP、客户端 IP、开始时间、持续时间、状态、操作
- 筛选条件：用户（下拉）、主机（下拉）、时间范围（日期选择器）、状态（全部/进行中/已关闭/中断）
- 操作列：[回放] 按钮

### TerminalReplay.vue（录像回放）

- 头部：会话信息（用户、主机、时间等）
- 主体：xterm.js 组件（只读模式）
- 控制栏：播放/暂停、进度条、播放速度（1x/2x/4x）

## 后端目录结构

```
backend/internal/modules/cmdb/
├── model/
│   └── terminal.go           # TerminalSession 模型
├── repository/
│   └── terminal.go           # 会话 CRUD
├── service/
│   └── terminal.go           # 终端会话管理逻辑
├── api/
│   └── terminal.go           # WebSocket 升级 + SSH
├── terminal/
│   ├── ssh.go                # SSH 连接管理
│   ├── recorder.go           # asciinema v2 录制
│   └── replay.go             # 录像文件读取
```

## 配置项

```yaml
# config.yaml 新增
terminal:
  recording_dir: "./data/recordings"  # 终端录像存储目录
  max_session_duration: 86400          # 最大会话时长（秒），默认 24 小时
  idle_timeout: 300                    # 空闲超时（秒），默认 5 分钟
```

## 依赖新增

### 后端

- `golang.org/x/crypto/ssh` — SSH 客户端（可能已存在）

### 前端

- 无新增（xterm.js 已在 K8s 模块中使用）

## 权限种子数据

在 `bootstrap/db.go` 的 `seedPermissions()` 中新增：

```go
{Name: "连接终端", Resource: "cmdb:terminal", Action: "connect", Description: "SSH Web 终端连接"},
{Name: "查看终端会话", Resource: "cmdb:terminal", Action: "list", Description: "查看终端会话列表"},
{Name: "查看会话详情", Resource: "cmdb:terminal", Action: "get", Description: "查看终端会话详情"},
{Name: "回放会话录像", Resource: "cmdb:terminal", Action: "replay", Description: "回放终端操作录像"},
```

## 侧边栏导航

在 MainLayout.vue 的 CMDB 菜单分组中新增：

```html
<el-menu-item index="/cmdb/terminal">终端审计</el-menu-item>
```

## 与现有模块的集成

- **复用凭据管理**：通过 credential_id 关联，使用 Phase 1 的凭据加解密逻辑
- **复用主机管理**：通过 host_id 关联，查询主机 IP、端口等信息
- **复用权限中间件**：使用 RequirePermission("cmdb:terminal", "connect")
- **复用审计中间件**：终端连接操作记录到审计日志

## 文件命名规范

录像文件命名：`terminal_{session_id}_{user_id}_{host_id}_{timestamp}.cast`

## 安全考虑

1. **凭据不传输到前端**：凭据解密仅在服务端进行
2. **会话超时**：idle_timeout 后自动断开空闲连接
3. **权限校验**：每次连接都校验用户是否有权限访问目标主机
4. **录像保护**：录像文件存储在服务端，前端只能读取无法篡改
