# CMDB 主机权限配置 - 实现规格

> **依赖前提**：Phase 1（主机管理、分组管理、凭据管理）已完成，Phase 2（终端审计）已完成。

## 概述

为 CMDB 模块增加主机级别的细粒度权限配置管理。本次仅实现权限规则的 CRUD 管理和查询接口，不在现有主机列表、终端连接等接口上强制权限校验（校验集成作为后续任务）。

## 功能范围

1. **权限规则 CRUD**：按分组为用户授予 view/terminal/admin 权限
2. **权限继承**：父分组授权自动继承到子分组及其中主机
3. **权限查询**：当前用户可访问的主机列表、权限检查接口
4. **前端管理页面**：分组树 + 权限规则表格的配置界面

## 数据模型

### HostPermission

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| tenant_id | uint | 租户 ID |
| user_id | uint | 用户 ID（外键 → users） |
| host_group_id | uint | 授权的分组 ID（外键 → cmdb_host_groups） |
| permission | string(20) | 权限：view / terminal / admin |
| created_by | uint | 创建者用户 ID |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

索引：
- `uk_perm_user_group` (tenant_id, user_id, host_group_id, permission) — 唯一约束
- `idx_perm_user` (user_id)
- `idx_perm_group` (host_group_id)

### 权限继承规则

- 授权到父级分组时，子分组及其中主机自动继承该权限
- admin 权限包含 terminal + view
- terminal 权限包含 view

继承实现方式：查询时向上遍历分组树，检查用户在主机所属分组及所有祖先分组上是否有足够权限。不使用冗余的物化记录。

## API 设计

路由前缀：`/api/v1/cmdb/permissions/`

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| GET | `/list` | 权限规则列表（按用户/分组/权限筛选，分页） | `cmdb:permission:list` |
| POST | `/create` | 授予权限（用户+分组+权限列表） | `cmdb:permission:create` |
| POST | `/update` | 更新权限规则 | `cmdb:permission:update` |
| POST | `/delete` | 删除权限规则 | `cmdb:permission:delete` |
| GET | `/my-hosts` | 当前用户可访问的主机列表（含权限级别） | 登录即可 |
| GET | `/check` | 检查用户对主机的权限，参数：host_id, action | 登录即可 |

### 请求/响应格式

**POST /create**
```json
{
  "user_id": 1,
  "host_group_id": 5,
  "permissions": ["view", "terminal"]
}
```

**GET /list**
查询参数：user_id, host_group_id, permission, page, page_size

**GET /my-hosts**
返回当前用户有权限的主机列表，每条包含权限级别。需解析继承：遍历用户所有权限规则，对每条规则展开该分组及其子分组下的主机。

**GET /check**
查询参数：host_id, action(view/terminal/admin)。查找主机所属分组，向上遍历分组树检查权限。

## 后端目录结构

```
backend/internal/modules/cmdb/
├── model/
│   └── permission.go         # HostPermission 模型 + GORM 注册
├── repository/
│   └── permission.go         # CRUD + 按用户/分组查询
├── service/
│   └── permission.go         # 继承解析 + 权限检查逻辑
└── api/
    └── permission.go         # HTTP Handler
```

## 前端设计

### 页面路由

```javascript
{ path: '/cmdb/permissions', component: PermissionList }
```

### PermissionList.vue

- 左侧面板：分组树（复用 Phase 1 HostGroup 树组件）
  - 点击分组时，右侧表格筛选该分组的权限规则
- 右侧面板：
  - 筛选条件：用户（下拉）+ 权限类型（下拉：view/terminal/admin）
  - [授予权限] 按钮
  - 权限规则表格：用户名、分组路径、权限（标签颜色区分）、创建时间、操作（编辑/删除）
- 授予权限弹窗：
  - 用户选择（搜索下拉）
  - 分组选择（树选择器，可搜索）
  - 权限多选（view/terminal/admin 标签）
  - 权限继承提示：显示「此权限将影响 N 台主机」（计算该分组及子分组下的主机数量）

### 侧边栏导航

在 MainLayout.vue 的 CMDB 菜单分组中新增：
```html
<el-menu-item index="/cmdb/permissions">权限配置</el-menu-item>
```

## 权限种子数据

在 `bootstrap/db.go` 的 `seedPermissions()` 中新增：

```go
{Name: "查看权限配置", Resource: "cmdb:permission", Action: "list", Description: "查看主机权限配置"},
{Name: "授予权限", Resource: "cmdb:permission", Action: "create", Description: "授予主机权限"},
{Name: "更新权限", Resource: "cmdb:permission", Action: "update", Description: "更新主机权限"},
{Name: "删除权限", Resource: "cmdb:permission", Action: "delete", Description: "删除主机权限"},
```

admin 角色自动拥有以上所有权限。

## 与现有模块的集成

- **复用分组管理**：权限按分组授予，查询时通过分组树解析继承
- **复用用户模块**：权限配置关联用户表
- **复用主机管理**：/my-hosts 接口返回主机列表，/check 接口通过主机找到所属分组
- **路由注册**：在 `api/cmdb.go` 的 cmdb router group 下注册 permissions 子路由

## 不在本期范围

- 在主机列表、终端连接等现有接口上强制权限校验（后续任务）
- 云账号管理（独立任务）
- 权限变更审计日志（后续任务）
