# CMDB 运维能力扩展设计

> 方案 A：堡垒机核心增强 + 运维工具链扩展
> 目标：以现有终端审计/权限体系为基石，补齐文件管理、任务执行、数据库管理、监控告警四大运维能力，实现"一个平台替代堡垒机 + CMDB + 任务调度 + DB管理 + 监控"。

## 目标用户

中大型企业运维团队（10+人），需要一个全能型平台替代多个工具，更看重安全审计、权限管控、多租户等企业级特性。

## 核心差异化策略

每一项运维操作（终端、文件、脚本、SQL）都自带审计录制 + 权限管控。竞品 AutoOps 没有这些能力。

## 整体架构

所有新模块共享 CMDB 的资产/凭证/权限体系，不独立建设。

### 复用层

| 现有能力 | 复用到 |
|---|---|
| SSH 连接 + 凭证解密 | 文件管理(SFTP)、任务执行、Agent 部署 |
| HostPermission + 组继承 | 所有模块的操作权限 |
| 终端录制 (asciicast v2) | 任务执行日志录制、SQL 审计录制 |
| CloudAccount 同步 | 自动发现监控目标、数据库实例 |
| 终端会话管理 | 任务执行会话、数据库查询会话 |

### 模块划分

后端新增模块均放在 `backend/internal/modules/cmdb/` 下：

```
cmdb/
├── terminal/      # 现有：SSH 终端 + 录制
├── file/          # 新增：SFTP 文件管理
├── task/          # 新增：脚本任务执行
├── database/      # 新增：数据库管理 + SQL 查询
└── monitor/       # 新增：主机监控 + 告警
```

前端新增页面均挂在 `/cmdb/` 路由下：

```
/cmdb/terminal/sessions    # 现有
/cmdb/files                # 新增：文件浏览器
/cmdb/tasks                # 新增：任务中心（模板+执行+定时）
/cmdb/databases            # 新增：数据库管理
/cmdb/monitoring           # 新增：主机监控
```

---

## 模块一：文件管理（SFTP）

### 场景

用户在 CMDB 中看到主机列表，选中一台直接浏览/上传/下载文件，不需要再开 WinSCP/FileZilla。

### 功能清单

- **文件浏览器**：左树右表布局，目录树 + 文件列表（名称、大小、权限、修改时间、所有者）
- **文件操作**：上传（HTTP multipart）、下载、删除、重命名、新建目录、修改权限（chmod）
- **批量分发**：选多台主机 → 上传一个文件 → 并行分发到指定路径，显示每台主机的分发结果
- **文本预览/编辑**：文本文件在线查看/编辑（≤1MB），图片/PDF 预览
- **操作审计**：所有文件操作记录审计日志（操作人、时间、主机、文件路径、操作类型、结果）

### 技术实现

- 后端用 `golang.org/x/crypto/ssh` 建立 SFTP 会话（复用现有 SSH 连接逻辑和凭证解密）
- 目录浏览和文件操作走 REST API（JSON 请求/响应）
- 大文件上传走 HTTP multipart（`/api/v1/cmdb/file/upload`），后端通过 SFTP 写入远程
- 文件下载走 HTTP 流式响应（`/api/v1/cmdb/file/download`），后端通过 SFTP 读取远程
- 批量分发使用 goroutine 并行，复用 SSH 连接池
- 前端用 Element Plus `el-tree` + `el-table`，文本编辑用 Monaco Editor

### 数据模型

```go
// FileOperationLog 文件操作审计日志
type FileOperationLog struct {
    ID         uint      `gorm:"primaryKey"`
    UserID     uint      `gorm:"index"`                // 操作人
    HostID     uint      `gorm:"index"`                // 目标主机
    SessionID  string    `gorm:"index"`                // 会话ID
    OpType     string    // ls/upload/download/delete/rename/mkdir/chmod/edit
    FilePath   string    // 目标文件路径
    FileSize   int64     // 文件大小
    Result     string    // success/failed
    ErrorMsg   string    // 错误信息
    ClientIP   string    // 客户端IP
    CreatedAt  time.Time `gorm:"index"`
}
```

### API 设计

```
GET    /api/v1/cmdb/file/browse/:hostId?path=/     # 浏览目录
POST   /api/v1/cmdb/file/upload/:hostId             # 上传文件
GET    /api/v1/cmdb/file/download/:hostId            # 下载文件
DELETE /api/v1/cmdb/file/:hostId                     # 删除文件/目录
PUT    /api/v1/cmdb/file/rename/:hostId              # 重命名
POST   /api/v1/cmdb/file/mkdir/:hostId               # 新建目录
PUT    /api/v1/cmdb/file/chmod/:hostId               # 修改权限
POST   /api/v1/cmdb/file/distribute                  # 批量分发
GET    /api/v1/cmdb/file/preview/:hostId             # 预览文件内容
PUT    /api/v1/cmdb/file/edit/:hostId                # 编辑文件内容
GET    /api/v1/cmdb/file/audit                        # 操作审计日志
```

### 权限

复用主机权限模型 — 用户需要对目标主机有 `terminal` 或 `admin` 级别权限才能操作文件。

### 与竞品差异

AutoOps 没有文件浏览器，只能通过 SSH 终端操作文件。我们有完整的 SFTP GUI + 操作审计。

---

## 模块二：任务执行中心

### 场景

运维需要批量在多台服务器上执行脚本/命令，替代 Ansible Tower / SaltStack。

### 功能清单

- **脚本模板**：创建/管理 Shell、Python 脚本模板，支持参数变量（`{{host}}`、`{{date}}`、自定义变量）
- **即时执行**：选主机 + 选脚本 → 立即执行 → 实时查看输出（WebSocket 推送）
- **定时任务**：cron 表达式调度，支持单次/重复，启用/禁用
- **批量执行**：多台主机并行执行，实时输出聚合展示（按主机分组显示）
- **执行历史**：每次执行的日志、状态、耗时、操作人
- **执行回放**：复用 asciicast 录制，执行过程可回放

### 技术实现

- 后端通过 SSH 连接执行远程命令（复用凭证解密和连接逻辑）
- 使用 `robfig/cron/v3` 管理定时任务
- WebSocket 实时推送执行输出（每条输出带主机标识和时间戳）
- 执行日志存储为 asciicast v2 格式（复用录制基础设施）
- 前端用 xterm.js 展示实时输出 + 历史回放

### 数据模型

```go
// TaskTemplate 脚本模板
type TaskTemplate struct {
    ID          uint      `gorm:"primaryKey"`
    Name        string    `gorm:"size:200;not null"`
    Description string    `gorm:"size:500"`
    ScriptType  string    `gorm:"size:20;not null"`   // shell/python
    Content     string    `gorm:"type:text;not null"`  // 脚本内容
    Parameters  string    `gorm:"type:text"`           // JSON: 参数定义
    CreatedBy   uint      `gorm:"index"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// TaskExecution 任务执行记录
type TaskExecution struct {
    ID         uint      `gorm:"primaryKey"`
    TemplateID uint      `gorm:"index"`
    Name       string    `gorm:"size:200"`
    HostIDs    string    `gorm:"type:text"`           // JSON: [1,2,3]
    Params     string    `gorm:"type:text"`           // JSON: 参数值
    Type       string    `gorm:"size:20"`             // immediate/scheduled
    CronExpr   string    `gorm:"size:100"`            // cron 表达式
    Status     string    `gorm:"size:20;index"`       // pending/running/completed/failed/cancelled
    CreatedBy  uint      `gorm:"index"`
    StartedAt  *time.Time
    FinishedAt *time.Time
    CreatedAt  time.Time
}

// TaskExecutionHost 主机级执行结果
type TaskExecutionHost struct {
    ID           uint      `gorm:"primaryKey"`
    ExecutionID  uint      `gorm:"index"`
    HostID       uint      `gorm:"index"`
    Status       string    `gorm:"size:20;index"`     // pending/running/success/failed/timeout
    OutputFile   string    // asciicast 录制文件路径
    ExitCode     int
    Duration     int       // 秒
    ErrorMessage string
    StartedAt    *time.Time
    FinishedAt   *time.Time
}
```

### API 设计

```
# 脚本模板
GET    /api/v1/cmdb/task/templates           # 模板列表
POST   /api/v1/cmdb/task/templates           # 创建模板
GET    /api/v1/cmdb/task/templates/:id       # 模板详情
PUT    /api/v1/cmdb/task/templates/:id       # 更新模板
DELETE /api/v1/cmdb/task/templates/:id       # 删除模板

# 任务执行
POST   /api/v1/cmdb/task/execute             # 即时执行
POST   /api/v1/cmdb/task/schedule            # 创建定时任务
GET    /api/v1/cmdb/task/executions          # 执行历史列表
GET    /api/v1/cmdb/task/executions/:id      # 执行详情
GET    /api/v1/cmdb/task/executions/:id/hosts # 主机级执行结果
WS     /api/v1/cmdb/task/executions/:id/ws   # 实时输出 WebSocket
DELETE /api/v1/cmdb/task/executions/:id      # 取消执行
PUT    /api/v1/cmdb/task/schedule/:id/toggle # 启用/禁用定时任务
GET    /api/v1/cmdb/task/schedules           # 定时任务列表
DELETE /api/v1/cmdb/task/schedules/:id       # 删除定时任务
```

### 权限

复用主机权限模型 — 用户只能对有 `terminal` 或 `admin` 权限的主机执行任务。执行记录对有查看权限的用户可见。

### 与竞品差异

AutoOps 有脚本任务但没有执行回放，没有参数变量模板，没有与 CMDB 权限体系深度集成。我们的任务执行自动继承主机权限 + 执行过程可回放。

---

## 模块三：数据库管理

### 场景

DBA/运维需要管理 MySQL/PgSQL/Redis 等实例，直接在平台内执行 SQL 查询。

### 功能清单

- **实例注册**：注册数据库实例（类型、地址、端口、凭证），凭证复用 CMDB 凭证体系
- **连接管理**：连接池 + 连接测试 + 连接状态监控 + 最大并发连接数控制
- **Web SQL 查询**：SQL 编辑器（语法高亮 + 自动补全）+ 结果表格展示 + 分页
- **SQL 审计**：所有 SQL 查询自动记录（操作人、时间、目标库、SQL 内容、影响行数、耗时）
- **慢查询**：展示 MySQL slow_query_log Top 列表
- **权限控制**：哪些用户能查询哪些库，复用主机权限模型扩展

### 技术实现

- 后端用 `database/sql` + 各数据库驱动（go-sql-driver/mysql、pgx、go-redis）
- SQL 查询走 WebSocket（大结果集流式推送）
- SQL 审计存入专用表，支持导出
- 前端 SQL 编辑器用 Monaco Editor（SQL 语法高亮 + 关键词自动补全）
- 查询结果用 Element Plus `el-table` 展示，支持排序、筛选、导出 CSV

### 数据模型

```go
// DatabaseInstance 数据库实例
type DatabaseInstance struct {
    ID           uint      `gorm:"primaryKey"`
    Name         string    `gorm:"size:200;not null"`
    Type         string    `gorm:"size:20;not null"`   // mysql/postgresql/redis/elasticsearch/mongodb
    Host         string    `gorm:"size:200;not null"`
    Port         int       `gorm:"not null"`
    CredentialID uint      `gorm:"index"`              // 复用 CMDB 凭证
    Database     string    `gorm:"size:200"`           // 默认数据库
    MaxConns     int       `gorm:"default:5"`          // 最大并发连接
    Status       string    `gorm:"size:20"`            // online/offline
    Description  string    `gorm:"size:500"`
    CreatedBy    uint      `gorm:"index"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// SQLAuditLog SQL 审计日志
type SQLAuditLog struct {
    ID         uint      `gorm:"primaryKey"`
    UserID     uint      `gorm:"index"`
    InstanceID uint      `gorm:"index"`
    Database   string    `gorm:"size:200"`
    SQL        string    `gorm:"type:text"`
    SQLType    string    `gorm:"size:20"`           // SELECT/INSERT/UPDATE/DELETE/DDL
    Rows       int64                              // 影响行数
    Duration   int                                 // 毫秒
    Status     string    `gorm:"size:20"`           // success/failed
    ErrorMsg   string    `gorm:"type:text"`
    ClientIP   string    `gorm:"size:50"`
    CreatedAt  time.Time `gorm:"index"`
}
```

### API 设计

```
# 实例管理
GET    /api/v1/cmdb/db/instances             # 实例列表
POST   /api/v1/cmdb/db/instances             # 注册实例
GET    /api/v1/cmdb/db/instances/:id         # 实例详情
PUT    /api/v1/cmdb/db/instances/:id         # 更新实例
DELETE /api/v1/cmdb/db/instances/:id         # 删除实例
POST   /api/v1/cmdb/db/instances/:id/test    # 连接测试

# SQL 查询
WS     /api/v1/cmdb/db/query/:instanceId     # SQL 查询 WebSocket
GET    /api/v1/cmdb/db/schemas/:instanceId   # 获取数据库/表结构

# SQL 审计
GET    /api/v1/cmdb/db/audit                 # SQL 审计日志
GET    /api/v1/cmdb/db/slow-queries/:instanceId # 慢查询列表
```

### 安全限制

- 默认禁止 DDL（CREATE/DROP/ALTER）和 DML（INSERT/UPDATE/DELETE）操作，仅允许 SELECT
- 可通过权限配置开放写权限（需要 `admin` 级别）
- 单次查询最大返回行数可配置（默认 1000）
- 查询超时限制（默认 30 秒）

### 与竞品差异

AutoOps 有数据库管理但没有 SQL 审计录制，没有与主机权限联动。我们的 DB 管理继承 CMDB 权限 + 自带审计。

---

## 模块四：监控告警

### 场景

运维需要实时了解主机状态，异常时收到告警。

### 功能清单

- **主机监控**：CPU、内存、磁盘、网络指标的实时曲线 + 历史趋势
- **Agent 管理**：在主机上部署/卸载/升级监控 Agent（复用 SSH 连接 + 文件管理）
- **告警规则**：基于 PromQL 的告警规则配置（CPU > 80% 持续 5 分钟等）
- **告警通知**：支持企业微信/钉钉/邮件/webhook 通知
- **告警历史**：告警记录、确认、静默
- **Dashboard**：主机资源概览大盘

### 技术实现

- 部署 Prometheus + Pushgateway（Docker Compose 扩展）
- Go Agent 定期上报主机指标到 Pushgateway（CPU/内存/磁盘/网络/TCP连接数/进程数）
- Agent 二进制通过 SFTP 分发到主机，通过 SSH 启动/停止
- 后端查询 Prometheus HTTP API 获取指标数据
- 告警规则通过 Prometheus Alertmanager 管理
- 告警通知通过 webhook 回调后端，后端转发到企业微信/钉钉/邮件
- 前端用 ECharts 展示监控图表

### 数据模型

```go
// MonitorAgent 监控 Agent
type MonitorAgent struct {
    ID         uint      `gorm:"primaryKey"`
    HostID     uint      `gorm:"uniqueIndex"`
    Status     string    `gorm:"size:20"`          // running/stopped/error
    Version    string    `gorm:"size:50"`
    LastBeat   time.Time // 最后心跳
    Port       int       `gorm:"default:9100"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

// AlertRule 告警规则
type AlertRule struct {
    ID          uint      `gorm:"primaryKey"`
    Name        string    `gorm:"size:200;not null"`
    PromQL      string    `gorm:"type:text;not null"`
    Duration    string    `gorm:"size:50"`           // 持续时间 "5m"
    Severity    string    `gorm:"size:20"`           // critical/warning/info
    Labels      string    `gorm:"type:text"`         // JSON
    Annotations string    `gorm:"type:text"`         // JSON: 描述信息
    NotifyType  string    `gorm:"size:100"`          // wecom/dingtalk/email/webhook
    NotifyTarget string   `gorm:"type:text"`         // 通知目标
    Enabled     bool      `gorm:"default:true"`
    CreatedBy   uint      `gorm:"index"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// AlertRecord 告警记录
type AlertRecord struct {
    ID         uint      `gorm:"primaryKey"`
    RuleID     uint      `gorm:"index"`
    RuleName   string    `gorm:"size:200"`
    Severity   string    `gorm:"size:20"`
    Message    string    `gorm:"type:text"`
    Status     string    `gorm:"size:20;index"`     // firing/resolved/acknowledged/silenced
    FiredAt    time.Time `gorm:"index"`
    ResolvedAt *time.Time
    AckedBy    uint      `gorm:"index"`
    AckedAt    *time.Time
    Labels     string    `gorm:"type:text"`
}
```

### API 设计

```
# Agent 管理
GET    /api/v1/cmdb/monitor/agents            # Agent 列表
POST   /api/v1/cmdb/monitor/agents/deploy     # 部署 Agent（选主机）
POST   /api/v1/cmdb/monitor/agents/:id/start  # 启动
POST   /api/v1/cmdb/monitor/agents/:id/stop   # 停止
DELETE /api/v1/cmdb/monitor/agents/:id         # 卸载

# 监控数据
GET    /api/v1/cmdb/monitor/metrics/:hostId   # 主机指标（query Prometheus）
GET    /api/v1/cmdb/monitor/dashboard          # Dashboard 汇总数据

# 告警规则
GET    /api/v1/cmdb/monitor/alert-rules        # 规则列表
POST   /api/v1/cmdb/monitor/alert-rules        # 创建规则
PUT    /api/v1/cmdb/monitor/alert-rules/:id    # 更新规则
DELETE /api/v1/cmdb/monitor/alert-rules/:id    # 删除规则
PUT    /api/v1/cmdb/monitor/alert-rules/:id/toggle # 启用/禁用

# 告警记录
GET    /api/v1/cmdb/monitor/alerts             # 告警列表
PUT    /api/v1/cmdb/monitor/alerts/:id/ack     # 确认告警
PUT    /api/v1/cmdb/monitor/alerts/:id/silence # 静默告警
```

### 与竞品差异

AutoOps 的监控是基础的主机指标，我们的监控与 CMDB 深度联动：自动发现监控目标、告警关联到资产、权限控制谁看什么数据、Agent 部署复用 SSH + SFTP 能力。

---

## 实现顺序

1. **文件管理（SFTP）** — 最简单，复用度最高，快速交付价值
2. **任务执行中心** — 核心运维能力，替代 Ansible Tower
3. **数据库管理** — DBA 刚需，与现有凭证体系天然融合
4. **监控告警** — 依赖 Agent 部署，需要文件管理模块支持

## UI 风格

经典运维风格：左树右表、标签页导航、批量操作按钮。与现有 CMDB 页面保持一致。

## 总结

通过这 4 个模块的扩展，平台将从"CMDB + 终端审计"升级为"全能运维平台"，每一项操作都自带审计 + 权限管控。用户不再需要：
- 堡垒机（终端审计 + 文件管理已覆盖）
- Ansible Tower（任务执行中心已覆盖）
- Navicat/DBeaver（数据库管理已覆盖）
- Zabbix/Prometheus AlertManager（监控告警已覆盖）
