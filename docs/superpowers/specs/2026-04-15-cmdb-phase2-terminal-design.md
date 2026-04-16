# CMDB 资产管理模块 - 第二阶段：SSH Web 终端与终端审计

> 依赖前提：第一阶段（主机管理、分组管理、凭据管理）已完成。
> 
> 本文档为第二阶段实施规格，覆盖可直接落地的首版范围，不包含第三阶段的云账号同步和主机级细粒度授权。

## 1. 目标

第二阶段只解决一个问题：让用户可以从已纳管主机直接进入浏览器终端，并且整个连接过程可审计、可回放。

首版必须形成完整闭环：

1. 从主机列表进入 SSH Web 终端
2. 连接过程创建终端会话记录
3. 完整记录输入输出流
4. 会话结束后生成可回放录像
5. 在终端审计页面按条件查看和回放历史会话

## 2. 明确边界

### 本阶段包含

- 从主机管理页直接发起终端连接
- WebSocket 到 SSH 的双向桥接
- 终端会话元信息存储
- 完整输入输出录制
- 审计列表页
- 录像回放页
- 全局终端权限控制
- 本地文件系统录像存储

### 本阶段不包含

- 按主机或按分组的细粒度授权
- 云账号接入和云主机同步
- 多凭据选择
- 输入内容脱敏编辑
- 对象存储或外部文件服务
- 跨租户共享会话或录像

## 3. 已确认的产品决策

### 3.1 首版范围

第二阶段首版直接做到完整闭环：Web SSH + 会话记录 + 录像回放。

### 3.2 权限模型

第二阶段先只做全局终端权限，不做主机级或分组级授权。

权限资源定义：

- `cmdb:terminal connect`
- `cmdb:terminal list`
- `cmdb:terminal get`
- `cmdb:terminal replay`

第三阶段再补主机/分组级授权，把全局权限升级为“是否可用终端功能”，细粒度权限升级为“能连哪些主机”。

### 3.3 入口位置

终端入口放在 `frontend/src/views/Cmdb/HostList.vue` 的操作列。

每条主机记录新增一个主操作按钮：

- 终端连接

“终端审计”作为独立菜单和独立页面存在，不做“终端中心”。

### 3.4 凭据来源

首版只允许使用主机上已经绑定的凭据。

规则：

- 主机未绑定凭据，不允许发起终端连接
- 连接时不弹凭据选择框
- 连接时不允许临时指定其他凭据

这样可以保证终端连接、审计记录、主机资产之间关系稳定，避免首版把凭据选择和审计归属做复杂。

### 3.5 录像内容

首版记录完整输入输出流：

- 用户输入记入录像
- SSH 输出记入录像
- 回放时按原始顺序还原

### 3.6 存储方式

录像首版落本地文件系统。

后续如果录制量变大，再迁移到对象存储。首版不为对象存储预留额外抽象层。

## 4. 前端设计

## 4.1 页面与路由

新增页面：

- `/cmdb/terminal/sessions`，终端审计列表页
- `/cmdb/terminal/replay/:id`，终端回放页

在资产管理菜单下新增：

- 终端审计

不增加“终端连接”单独菜单项，因为终端连接是主机上下文动作，不是独立业务对象。

## 4.2 主机列表中的终端入口

文件：`frontend/src/views/Cmdb/HostList.vue`

在操作列中新增按钮，顺序调整为：

- 终端连接
- 测试连接
- 编辑
- 删除

点击“终端连接”后的行为：

1. 如果主机未绑定 `credentialId`，前端直接提示“请先为主机绑定凭据”
2. 如果主机已绑定凭据，打开终端弹窗
3. 弹窗中挂载通用终端组件并建立 WebSocket 连接

首版终端体验采用 `el-dialog`，不跳转独立连接页。原因是这更贴合主机管理的工作流，用户连接结束后可以直接继续管理主机。

## 4.3 终端组件复用

复用 `frontend/src/components/K8s/Terminal.vue` 的 xterm.js 能力，但要改造成通用终端组件，而不是继续绑定 K8s 语义。

组件应保留以下输入输出：

- `wsUrl`
- `visible`
- `fit()` 暴露方法
- 处理 `stdout` / `stderr` / `error` / `closed`

这样 K8s Pod 终端和 CMDB SSH 终端可以共用一套前端终端外壳。

## 4.4 终端审计列表页

文件：`frontend/src/views/Cmdb/TerminalSessionList.vue`

展示字段：

- 主机名
- 主机 IP
- 用户名
- 客户端 IP
- 开始时间
- 持续时长
- 状态
- 操作

筛选项首版保留：

- 主机关键字
- 用户关键字
- 状态
- 时间范围

操作列只保留：

- 回放

## 4.5 回放页

文件：`frontend/src/views/Cmdb/TerminalReplay.vue`

结构分为三块：

1. 顶部会话元信息
2. 中间只读终端区域
3. 底部播放控制条

控制条首版支持：

- 播放 / 暂停
- 重新播放
- 倍速切换：1x / 2x / 4x
- 当前进度 / 总时长显示

首版不做关键帧定位、命令检索、输入高亮、下载录像。

## 5. 后端设计

## 5.1 目录结构与职责

所有代码继续放在 CMDB 模块内，不新建平级大模块。

建议目录：

```text
backend/internal/modules/cmdb/
├── model/
│   └── terminal.go
├── repository/
│   └── terminal.go
├── service/
│   └── terminal.go
├── api/
│   └── terminal.go
└── terminal/
    ├── ssh.go
    ├── bridge.go
    ├── recorder.go
    └── replay.go
```

职责划分：

- `model/terminal.go`：终端会话模型定义
- `repository/terminal.go`：会话列表、详情、状态更新
- `service/terminal.go`：主机、凭据、会话编排逻辑
- `api/terminal.go`：WebSocket 升级、列表接口、详情接口、录像读取接口
- `terminal/ssh.go`：SSH client/session/pty 建立与 resize
- `terminal/bridge.go`：WebSocket 与 SSH 双向数据桥接
- `terminal/recorder.go`：asciinema v2 写入
- `terminal/replay.go`：读取 cast 文件并返回回放数据

## 5.2 路由设计

继续沿用当前 CMDB 的非 REST 风格。

路由前缀：`/api/v1/cmdb/terminal`

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET（WebSocket） | `/connect?hostId=1` | 建立 SSH 终端连接 | `cmdb:terminal connect` |
| GET | `/list` | 会话列表 | `cmdb:terminal list` |
| GET | `/detail?id=1` | 会话详情 | `cmdb:terminal get` |
| GET | `/recording?id=1` | 获取会话录像内容 | `cmdb:terminal replay` |

不提供 create/update/delete，会话创建和关闭由终端连接生命周期驱动。

## 5.3 WebSocket 协议

客户端发送：

```json
{"operation":"stdin","data":"ls\n"}
{"operation":"resize","cols":120,"rows":30}
```

服务端发送：

```json
{"operation":"stdout","data":"total 8\r\n"}
{"operation":"stderr","data":"permission denied\r\n"}
{"operation":"error","data":"ssh auth failed"}
{"operation":"closed","data":"connection closed"}
```

协议保持和现有 K8s 终端风格一致，便于复用前端通用终端组件。

## 5.4 建连流程

```text
用户在主机列表点击“终端连接”
    ↓
前端建立 WebSocket: /api/v1/cmdb/terminal/connect?hostId=X
    ↓
后端完成 Session/Bearer 认证
    ↓
校验 cmdb:terminal connect 权限
    ↓
按租户查询主机
    ↓
校验主机已绑定 credentialId
    ↓
按租户查询凭据并解密
    ↓
建立 SSH 连接并申请 PTY
    ↓
创建 TerminalSession(status=active)
    ↓
创建 cast 文件并写入 header
    ↓
启动双向桥接与录制
```

关键规则：

- SSH 建立成功后才创建 `TerminalSession`
- 连接前失败不产生会话记录
- 连接前失败只进入普通审计日志

## 5.5 关闭流程

```text
浏览器主动关闭 / 网络中断 / SSH 断开 / 后端结束连接
    ↓
停止双向桥接
    ↓
关闭 recorder
    ↓
关闭 SSH session/client
    ↓
更新 TerminalSession
      - finishedAt
      - duration
      - fileSize
      - status=closed 或 interrupted
    ↓
向前端发送 closed 消息（若连接仍可写）
```

状态规则：

- 用户主动正常关闭，记为 `closed`
- 网络异常、远端断开、非预期关闭，记为 `interrupted`

## 6. 数据模型

文件：`backend/internal/modules/cmdb/model/terminal.go`

### 6.1 TerminalSession

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| user_id | uint | 用户 ID |
| username | string(100) | 用户名冗余 |
| host_id | uint | 主机 ID |
| host_ip | string(45) | 主机 IP 冗余 |
| host_name | string(255) | 主机名冗余 |
| credential_id | uint | 使用的凭据 ID |
| client_ip | string(45) | 客户端 IP |
| started_at | timestamp | 开始时间 |
| finished_at | timestamp nullable | 结束时间 |
| duration | int | 持续秒数 |
| recording_path | string(500) | 录像文件路径 |
| file_size | int64 | 文件大小 |
| status | string(20) | active / closed / interrupted |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
| deleted_at | gorm.DeletedAt | 软删 |

索引：

- `idx_cmdb_terminal_tenant_started` (`tenant_id`, `started_at`)
- `idx_cmdb_terminal_tenant_host` (`tenant_id`, `host_id`)
- `idx_cmdb_terminal_tenant_user` (`tenant_id`, `user_id`)
- `idx_cmdb_terminal_tenant_status` (`tenant_id`, `status`)

首版不建“逐条命令表”或“逐帧事件表”，录像内容只存在文件系统。

## 7. 录像格式与存储

## 7.1 录像格式

使用 asciinema v2 格式。

header 示例：

```json
{"version":2,"width":120,"height":30,"timestamp":1713136800}
```

事件示例：

```json
[0.532,"i","ls\n"]
[0.913,"o","total 8\r\n"]
```

规则：

- 用户输入记为 `i`
- 服务端输出记为 `o`
- 时间使用相对会话开始时间的秒偏移

## 7.2 存储目录

首版录像目录：

```text
backend/data/recordings/cmdb-terminal/YYYYMMDD/
```

文件名：

```text
terminal_<sessionId>.cast
```

不把 `userId`、`hostId` 拼入文件名。文件与业务关系由数据库维护，文件名保持简单，后续迁移更容易。

## 7.3 回放读取方式

首版不直接把 cast 文件原样下发给前端播放器。

回放接口由后端读取 cast 文件，解析为前端可直接消费的数据结构返回。这样后续如果要增加：

- 脱敏
- 水印
- 权限控制扩展
- 录像下载控制

都可以在服务端统一处理。

## 8. 认证、权限与安全

### 8.1 认证

继续复用当前 session / Bearer 认证方式，不新增终端专用认证模型。

### 8.2 权限

第二阶段只校验全局终端权限，不校验主机级授权。

### 8.3 租户隔离

所有主机、凭据、会话查询都必须按租户范围读取。

连接终端时：

- 只能读取租户内主机
- 只能读取该主机关联的租户内凭据
- 只能查看租户内终端会话和录像

### 8.4 敏感信息保护

- 凭据只在后端解密
- 凭据永不返回前端
- 录像文件只通过受控接口读取
- 未授权用户不能直接读文件路径

### 8.5 超时控制

首版必须有两个保护：

- 最大会话时长
- 空闲超时

建议配置：

```yaml
terminal:
  recording_dir: "./data/recordings/cmdb-terminal"
  max_session_duration: 86400
  idle_timeout: 300
```

## 9. 审计规则

### 9.1 TerminalSession 的含义

`TerminalSession` 只代表“真正建立成功的终端会话”。

以下情况不创建 TerminalSession：

- 主机不存在
- 主机未绑定凭据
- 凭据不存在
- SSH 认证失败
- 目标不可达
- PTY 建立失败

### 9.2 普通审计日志

连接前失败仍要进入普通审计日志，便于排查和追责。

因此第二阶段存在两条审计线：

1. 普通审计日志，记录连接尝试、失败、查看录像等动作
2. TerminalSession，记录真正建立成功的终端会话与录像

## 10. 错误处理

### 10.1 连接前错误

前端收到明确提示，连接直接失败：

- 无权限
- 主机未绑定凭据
- 凭据不存在
- SSH 认证失败
- 目标不可达

### 10.2 连接中错误

如果连接已建立，中途断开：

- 页面提示连接关闭
- 会话标记为 `closed` 或 `interrupted`
- 已写入的录像保持可回放

### 10.3 录像读取错误

- 文件不存在，返回“录像不存在或已损坏”
- 解析失败，返回“录像解析失败”
- 会话不存在或越权访问，返回 404 或 403

## 11. 与现有代码的集成点

### 前端

- 复用 `frontend/src/components/K8s/Terminal.vue`
- 在 `frontend/src/views/Cmdb/HostList.vue` 增加入口
- 在 `frontend/src/components/Layout/MainLayout.vue` 增加“终端审计”菜单
- 在 `frontend/src/components/Layout/Breadcrumb.vue` 增加对应面包屑

### 后端

- 参考 `backend/internal/modules/k8s/api/terminal.go` 的 WebSocket 生命周期和同源校验写法
- 在 `backend/routers/v1/cmdb.go` 注册终端相关路由
- 在 `backend/internal/bootstrap/db.go` 增加 TerminalSession 模型迁移和权限种子
- 复用 Phase 1 的主机、凭据、租户隔离逻辑

## 12. 测试策略

第二阶段验收不能只靠 build，至少覆盖四层。

### 12.1 Service 单测

- 只有租户内主机能被读取
- 只有租户内凭据能被读取
- 主机未绑定凭据时拒绝连接
- 成功连接后正确创建会话
- 关闭时正确写回状态、时长、文件大小

### 12.2 Recorder 单测

- cast header 正确
- 输入输出事件格式正确
- 文件关闭后内容可解析

### 12.3 API 集成测试

- 未登录不能连接
- 无权限不能连接
- 未绑定凭据不能连接
- 会话列表可查
- 会话详情可查
- 回放接口可读

### 12.4 手工联调

- 从主机管理页打开终端
- 执行数条命令
- 正常关闭
- 到终端审计页查到记录
- 打开回放页可重播刚才会话

## 13. 实施顺序

实现顺序按下面四块推进：

1. **后端基础层**
   - TerminalSession 模型
   - 仓储、列表、详情、录像读取接口
   - 路由和权限种子
2. **终端桥接层**
   - SSH 建连
   - WebSocket 双向桥接
   - recorder 写 cast
   - 会话关闭收尾
3. **前端接入层**
   - 主机列表“终端连接”入口
   - 通用终端组件改造
   - 审计列表页
4. **回放与联调层**
   - 回放页
   - 冒烟测试
   - 联调修正

## 14. 推迟到第三阶段的内容

以下内容明确推迟到第三阶段：

- 按主机或分组的细粒度终端授权
- 终端权限和主机权限统一校验
- 云主机同步后自动接入终端权限体系
- 录像脱敏和高级审计分析

这是刻意的分阶段边界，不是遗漏。
