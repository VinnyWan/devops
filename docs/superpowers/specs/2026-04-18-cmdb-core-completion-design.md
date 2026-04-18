# CMDB 核心补全 - 实现规格

> **依赖前提**：Phase 1（主机/分组/凭证）、Phase 2（终端审计）、Phase 3（权限配置、云账号）已完成。
> **实施顺序**：云同步增强 → 录像清理轮转 → 主机级权限接入。

---

## 一、云同步增强

### 1.1 启动接入

在 `backend/internal/bootstrap/` 中应用初始化阶段调用 `cloudSyncService.StartScheduledSync()`，确保服务启动时自动开始定时同步。

具体做法：在 bootstrap 完成数据库连接和模块初始化后，调用 `cloudSyncService.StartScheduledSync()` 启动后台 goroutine。

### 1.2 分页拉取

当前腾讯云 API 调用 `Limit` 硬编码为 100，超过 100 台实例的账户会同步不全。

改造方案：
- 每次请求 `Limit=100`，初始 `Offset=0`
- 循环拉取，每次 `Offset += Limit`
- 当返回结果数 < Limit 时结束
- 适用于全部 5 种资源：CVM、VPC、Subnet、SecurityGroup、CBS
- 提取公共分页函数 `paginateCloudResource` 放在 `service/cloud_sync.go` 内，避免 5 处重复代码

错误处理：
- 单次 API 调用失败记录日志继续（不影响其他资源类型的同步）
- 整页解析失败记录日志跳过该页继续下一页

### 1.3 测试覆盖

为 `cloud_sync.go`（561 行）补充测试文件 `service/cloud_sync_test.go`：
- Mock 腾讯云 SDK 的 DescribeInstances/DescribeVpcs 等方法
- 测试分页拉取逻辑（>100 台实例的多页场景）
- 测试同步自动创建 Host 记录
- 测试错误处理（API 调用失败不中断整体同步）

测试不使用 SQLite，基于 mock 接口进行。

---

## 二、录像文件清理轮转

### 2.1 配置

在 `backend/config/config.yaml` 中追加终端录像配置段（与现有 terminal 配置平级）：

```yaml
terminal:
  recording:
    max_age_days: 90    # 录像保留天数，0 表示不清理
    cleanup_hour: 3     # 每天凌晨 3 点执行清理（0-23）
```

### 2.2 清理逻辑

新增 `backend/internal/modules/cmdb/service/recording_cleanup.go`：
- `StartCleanupScheduler()` 启动定时任务，按 cleanup_hour 配置的每天执行
- `CleanupOldRecordings()` 核心清理函数：
  1. 查询数据库中 `created_at < now - max_age_days` 且 `status = closed` 的 terminal_session 记录
  2. 删除对应的 .cast 文件
  3. 将数据库记录标记为 `archived`（不物理删除，保留审计痕迹）
- 空的日期目录在清理后自动删除

### 2.3 边界处理

- 跳过 `status != closed` 的 session（避免删除正在录制的文件）
- 清理失败不影响主服务运行，仅记录错误日志
- max_age_days = 0 时跳过清理

### 2.4 启动接入

在 bootstrap 阶段与云同步定时任务一起启动。

---

## 三、主机级权限接入

### 3.1 终端连接权限校验

修改 `api/terminal.go` 的 WebSocket 连接处理：
- 在建立 SSH 连接之前，调用 `permissionService.CheckPermission(userID, hostID, "terminal")`
- 校验通过：继续建立连接
- 校验失败：返回 403 Forbidden，WebSocket 关闭，关闭原因记录为 `permission_denied`（新增 close_reason 枚举值，与现有的 closed/interrupted/idle_timeout/max_duration 并列）
- admin 角色跳过校验（通过 Casbin 全局 `cmdb:host:*` 权限判断，复用现有 isAdmin 机制）

### 3.2 主机列表权限过滤

修改 `api/host.go` 的 List 接口：
- 非 admin 用户：调用 `permissionService.GetUserHostIDs(userID)` 获取可见主机 ID 列表
- 在 GORM 查询中追加 `WHERE id IN (可见主机IDs)` 条件
- admin 角色看全部（不变）
- 无任何权限的用户返回空列表（不是 403）

### 3.3 主机详情权限校验

修改 `api/host.go` 的 Get 接口：
- 非 admin 用户：先调用 `permissionService.CheckPermission(userID, hostID, "view")`
- 无权限返回 404（不暴露主机存在性）
- admin 角色跳过校验

### 3.4 前端配合

- 主机列表：后端已做过滤，前端不需要额外改动
- 终端连接：WebSocket 连接被拒时（403），弹出「无该主机终端访问权限」提示
- 主机详情：404 时显示「主机不存在或无访问权限」

### 3.5 权限判断优先级

1. admin 角色（通过 Casbin 全局权限判断）→ 全部放行
2. 主机级权限（通过 HostPermission 表查询）→ 按具体规则判断
3. 无权限记录 → 拒绝

---

## 不在本期范围

- 权限变更审计日志
- 其他云厂商（AWS/阿里云）
- 终端录像下载、命令搜索、输入高亮
- 运维模块（告警/CI-CD/Harbor/日志/监控/应用管理）前端页面
