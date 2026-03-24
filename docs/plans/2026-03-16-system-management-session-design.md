# 系统管理页面与会话持久化设计方案

**日期**: 2026-03-16
**状态**: 设计中
**优先级**: 高

## 1. 目标

解决两个关键问题：
1. **会话持久化**：修复页面刷新导致重新登录的问题
2. **系统管理页面**：完成用户、部门、角色、权限管理的完整功能

## 2. 问题分析

### 2.1 会话持久化问题

**当前状态：**
- Session存储在内存中
- 页面刷新后session丢失
- 用户需要重新登录

**影响：**
- 用户体验差
- 权限更新后无法立即生效
- 开发调试困难

### 2.2 系统管理页面现状

**当前状态：**
- 4个管理页面都是空白占位
- 缺少CRUD功能
- 没有搜索和批量操作

**需求：**
- 基础CRUD操作
- 搜索和过滤
- 批量操作
- 数据导出
- 权限分配UI（角色-权限、用户-角色）

## 3. 技术方案

### 3.1 会话持久化方案

**架构设计：**
```
登录流程：
用户登录 → 验证成功 → 生成session_id
        → 存储到Redis (key: session:{id}, value: user_data)
        → 返回session_id到前端
        → 前端存储到Cookie

请求验证：
请求 → SessionAuth中间件 → 读取Cookie中的session_id
     → 从Redis查询session数据 → 验证有效期
     → 将用户信息写入context → 继续处理请求
```

**核心改动：**

1. **后端 - Session存储层**
   - 修改 `internal/modules/auth/service/auth_service.go`
   - Login方法：生成session_id，存储到Redis
   - 设置Cookie：HttpOnly, Secure, SameSite

2. **后端 - 中间件层**
   - 修改 `internal/middleware/session.go`
   - SessionAuth：从Cookie读取session_id，从Redis查询数据
   - 验证有效期，刷新过期时间

3. **前端 - API层**
   - 修改 `src/api/service.ts`
   - 配置axios携带Cookie（withCredentials: true）

### 3.2 系统管理页面方案

**组件化架构：**

```
通用组件层：
├── CrudTable.vue      # 通用CRUD表格
│   ├── 分页
│   ├── 加载状态
│   ├── 操作列（编辑/删除）
│   └── 批量选择
├── CrudForm.vue       # 通用表单
│   ├── 动态字段渲染
│   ├── 表单验证
│   └── 提交处理
├── SearchBar.vue      # 搜索栏
│   ├── 关键词搜索
│   ├── 条件过滤
│   └── 重置功能
└── BatchActions.vue   # 批量操作
    ├── 批量删除
    ├── 批量导出
    └── 自定义操作

页面层（配置驱动）：
├── UserList.vue       # 用户管理
├── DepartmentList.vue # 部门管理
├── RoleList.vue       # 角色管理 + 权限分配
└── PermissionList.vue # 权限管理
```

**组件接口设计：**

```typescript
// CrudTable Props
interface CrudTableProps {
  columns: ColumnConfig[]
  fetchData: (params: any) => Promise<any>
  onEdit?: (row: any) => void
  onDelete?: (id: number) => void
  permissions: {
    create?: string
    update?: string
    delete?: string
  }
}

// CrudForm Props
interface CrudFormProps {
  fields: FieldConfig[]
  initialData?: any
  onSubmit: (data: any) => void
}

// SearchBar Props
interface SearchBarProps {
  filters: FilterConfig[]
  onSearch: (params: any) => void
}
```

## 4. 数据流设计

### 4.1 会话持久化流程

```
登录成功：
前端 → POST /api/v1/user/login
     → 后端验证 → 生成session_id
     → Redis.Set("session:{id}", userData, 24h)
     → 返回 Set-Cookie: session_id=xxx
     → 前端自动存储Cookie

后续请求：
前端 → 请求携带Cookie
     → SessionAuth中间件 → Redis.Get("session:{id}")
     → 验证有效 → c.Set("userID", xxx)
     → 继续处理

权限更新：
管理员修改权限 → 更新数据库
              → Redis.Del("user:perms:{id}")
              → 用户下次请求重新加载权限
```

### 4.2 CRUD操作流程

```
列表查询：
页面加载 → fetchData(page, filters)
        → API请求 → 后端分页查询
        → 返回数据 → 表格渲染

新增/编辑：
点击按钮 → 打开Modal → 填写表单
        → 提交 → API请求 → 后端验证
        → 成功 → 关闭Modal → 刷新列表

批量操作：
勾选行 → 点击批量按钮 → 确认
      → 循环调用API → 显示进度
      → 完成 → 刷新列表
```

## 5. 导航结构调整

**当前结构：**
```
- 仪表盘
- 资产管理
- 容器管理 (7个子页面)
- 平台能力 (7个子页面)
  - 告警中心
  - 日志检索
  - 监控配置
  - Harbor管理
  - CI/CD流水线
  - 应用管理
  - 审计日志
- 系统管理 (4个子页面)
```

**调整后：**
```
- 仪表盘
- 资产管理
- 容器管理 (7个子页面)
- 告警中心 (提升到顶级)
- 日志检索 (提升到顶级)
- 监控配置 (提升到顶级)
- Harbor管理 (提升到顶级)
- CI/CD流水线 (提升到顶级)
- 应用管理 (提升到顶级)
- 审计日志 (提升到顶级)
- 系统管理 (4个子页面)
```

## 6. 错误处理

### 6.1 后端错误处理

**Session相关：**
- Session不存在：返回401，前端跳转登录
- Session过期：返回401，前端跳转登录
- Redis连接失败：降级到内存session（临时方案）

**权限相关：**
- 权限不足：返回403，显示提示信息
- 资源不存在：返回404

### 6.2 前端错误处理

**API错误：**
- 401：清除本地状态，跳转登录页
- 403：显示权限不足提示
- 500：显示服务器错误提示

**表单验证：**
- 必填项检查
- 格式验证（邮箱、手机号）
- 自定义规则验证

**网络错误：**
- 超时重试（最多3次）
- 显示网络错误提示

## 7. 实施计划

### 阶段1：会话持久化（后端）

**任务1.1：修改Session存储**
- 修改 auth_service.go 的 Login 方法
- 集成Redis存储session
- 设置Cookie返回

**任务1.2：修改SessionAuth中间件**
- 从Cookie读取session_id
- 从Redis查询session数据
- 验证并刷新过期时间

**任务1.3：前端配置**
- 配置axios withCredentials
- 测试Cookie携带

### 阶段2：通用组件开发（前端）

**任务2.1：创建CrudTable组件**
- 表格渲染
- 分页功能
- 操作列
- 批量选择

**任务2.2：创建CrudForm组件**
- 动态字段渲染
- 表单验证
- 提交处理

**任务2.3：创建SearchBar组件**
- 搜索输入
- 过滤器
- 重置功能

### 阶段3：系统管理页面（前端）

**任务3.1：用户管理页面**
- 配置字段和API
- 实现用户-角色分配
- 测试CRUD功能

**任务3.2：部门管理页面**
- 配置字段和API
- 树形结构展示
- 测试功能

**任务3.3：角色管理页面**
- 配置字段和API
- 实现角色-权限分配
- 测试功能

**任务3.4：权限管理页面**
- 配置字段和API
- 只读展示
- 测试功能

### 阶段4：导航调整与优化

**任务4.1：调整导航结构**
- 修改 MainLayout.vue
- 提升"平台能力"子页面到顶级

**任务4.2：批量操作与导出**
- 实现批量删除
- 实现数据导出

**任务4.3：集成测试**
- 测试完整流程
- 修复bug

## 8. 成功标准

- [ ] 页面刷新不会丢失登录状态
- [ ] 权限更新后刷新页面可获取最新权限
- [ ] 用户管理页面完整CRUD功能
- [ ] 部门管理页面完整CRUD功能
- [ ] 角色管理页面完整CRUD功能 + 权限分配
- [ ] 权限管理页面展示功能
- [ ] 搜索和过滤功能正常
- [ ] 批量操作功能正常
- [ ] 数据导出功能正常
- [ ] 导航结构调整完成
- [ ] 所有功能通过测试
