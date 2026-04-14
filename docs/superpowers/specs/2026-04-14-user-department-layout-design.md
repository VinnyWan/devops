# 用户管理页面增加部门关联展示

**日期**: 2026-04-14
**状态**: 已批准

## 背景

用户管理页面 (`/system/user`) 当前是一个简单的扁平用户表格，只显示用户名、邮箱、状态三列。用户与部门的关联关系在后端已完整存在（User 模型有 DepartmentID/Department 字段，部门树 API 已就绪），但前端完全没有展示和利用这些信息。

## 目标

将用户管理页面改造为左侧部门树 + 右侧用户表格的主从布局，使用户能按部门快速筛选和查看用户。

## 布局设计

```
┌─────────────────────────────────────────────────────┐
│ 用户管理                                 [新建用户]   │
├──────────┬──────────────────────────────────────────┤
│ 部门列表  │  [搜索框]                                 │
│          │  ┌────────────────────────────────────┐   │
│ ▸ 全部用户│  │ 用户名 | 邮箱 | 部门 | 角色 | 状态 | ... │   │
│ ▾ 总公司  │  │ 张三   | ...  | ...  | ...  | 启用  |   │
│   ▸ 技术部│  │ 李四   | ...  | ...  | ...  | 启用  |   │
│   ▸ 产品部│  │                                    │   │
│   ▸ 运营部│  └────────────────────────────────────┘   │
│          │  [分页: 1 2 3 ...]                         │
└──────────┴──────────────────────────────────────────┘
```

- **左侧面板**（约 240px）：`el-tree` 渲染部门树，顶部有「全部用户」节点
- **右侧面板**（flex:1）：搜索框 + 用户表格 + 分页

## 交互行为

1. 点击左侧部门节点 → 右侧表格筛选该部门直属用户
2. 点击「全部用户」→ 显示所有用户（取消筛选）
3. 搜索框输入关键词 → 按用户名/邮箱模糊搜索（防抖 300ms）
4. 创建/编辑弹窗中增加 `el-tree-select` 用于选择所属部门

## 后端改动

### `GET /user/list` 增加 `departmentId` 查询参数

**文件**: `backend/internal/modules/user/api/user.go` — `List` handler
- 读取 `departmentId` query 参数（可选，uint）
- 传递给 service 层

**文件**: `backend/internal/modules/user/service/user_service.go` — `ListUsers` 方法
- 增加 `departmentID *uint` 参数
- 当 `departmentID` 有值时，按部门 ID 过滤
- 当 `departmentID` 为空时，保持现有行为（返回所有用户）

**文件**: `backend/internal/modules/user/repository/user_repo.go`
- `ListInTenant` 增加可选 `departmentID *uint` 参数
- 有值时追加 `WHERE department_id = ?` 条件

## 前端改动

### 仅修改 `frontend/src/views/System/UserList.vue`

**数据加载**:
- 页面加载时调用 `getDepartmentList()` 获取部门树
- 调用 `getUserList()` 获取用户列表（现有逻辑）

**左侧部门树**:
- `el-tree` 组件，`node-key="id"`，`:default-expand-all="true"`
- 顶部增加一个虚拟节点「全部用户」，点击取消筛选
- `@node-click` 事件更新 `selectedDeptId`，重新请求用户列表

**右侧用户表格**:
- 列：用户名、邮箱、部门名称（`row.department?.name`）、角色（`row.roles?.map(r=>r.name).join('、')`）、状态、创建时间（dayjs 格式化）、最后登录（dayjs 格式化）
- 操作列：编辑、删除（保持现有逻辑）

**搜索 + 分页**:
- `el-input` 搜索框，`@input` 防抖 300ms，传 `keyword` 参数
- `el-pagination` 底部分页，传 `page`/`pageSize` 参数

**创建/编辑弹窗**:
- 新增 `el-form-item` 部门字段
- 使用 `el-tree-select`，数据源同左侧部门树
- 编辑时回显当前用户的 `departmentId`

### API 层无需改动

`getDepartmentList` 和 `getUserList` 已存在于 `api/system.js`。`getUserList` 已支持 `params` 传递查询参数。

## 不做的事情

- 不新建独立的部门树组件（直接在 UserList.vue 内联）
- 不修改部门管理页面（DepartmentList.vue）
- 不修改后端部门相关 API
- 不增加部门树的 CRUD 操作（用户管理页只读浏览部门）
