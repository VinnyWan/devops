# 系统管理页面与会话持久化实施计划

**设计文档**: docs/plans/2026-03-16-system-management-session-design.md
**创建时间**: 2026-03-16
**状态**: 待执行

## 任务列表

### 阶段1：会话持久化（后端）

#### Task 1.1: 修改Login方法集成Redis存储

**文件**: `backend/internal/modules/auth/service/auth_service.go`

**步骤**:
1. 导入Redis包和uuid包
2. 在Login方法中，验证成功后生成session_id (使用uuid)
3. 构造session数据结构 (userID, username, authSource, loginTime)
4. 使用Redis.Set存储session数据，key为"session:{id}"，过期时间24小时
5. 返回session_id给调用方

**验证**:
- 运行 `go build` 确保编译通过
- 检查Login方法返回值包含session_id

#### Task 1.2: 修改Login API返回Cookie

**文件**: `backend/internal/modules/auth/api/auth_api.go`

**步骤**:
1. 在Login handler中获取service返回的session_id
2. 使用gin的SetCookie方法设置Cookie
3. Cookie配置: Name="session_id", HttpOnly=true, Path="/", MaxAge=86400
4. 同时在响应body中也返回session_id（兼容性）

**验证**:
- 运行后端服务
- 使用curl测试登录接口，检查响应头包含Set-Cookie
- 检查Cookie属性正确

#### Task 1.3: 修改SessionAuth中间件从Redis读取

**文件**: `backend/internal/middleware/session.go`

**步骤**:
1. 在SessionAuth中间件中，先尝试从Cookie读取session_id
2. 如果Cookie中没有，再从Authorization header读取（兼容旧方式）
3. 使用Redis.Get查询session数据，key为"session:{id}"
4. 如果session不存在或过期，返回401错误
5. 解析session数据，将userID、username、authSource写入context
6. 刷新session过期时间（Redis.Expire延长24小时）

**验证**:
- 运行后端服务
- 登录后访问需要认证的接口，检查请求成功
- 检查日志中user字段有值
- 等待几秒后再次请求，验证session自动刷新

#### Task 1.4: 配置前端axios携带Cookie

**文件**: `frontend/src/api/service.ts`

**步骤**:
1. 在axios实例配置中添加 `withCredentials: true`
2. 确保baseURL配置正确

**验证**:
- 运行前端服务
- 打开浏览器开发者工具Network面板
- 登录后检查请求头包含Cookie
- 刷新页面，验证不需要重新登录

---

### 阶段2：通用组件开发（前端）

#### Task 2.1: 创建CrudTable组件

**文件**: `frontend/src/components/CrudTable.vue`

**步骤**:
1. 创建组件文件，定义Props接口（columns, fetchData, onEdit, onDelete, permissions）
2. 实现表格渲染（使用NDataTable）
3. 实现分页功能（page, pageSize, total）
4. 添加操作列（编辑/删除按钮，根据permissions控制显示）
5. 添加批量选择功能（checkbox列）
6. 添加加载状态和空状态展示
7. 实现数据自动加载（onMounted调用fetchData）

**验证**:
- 创建测试页面使用该组件
- 验证表格正常渲染
- 验证分页功能正常
- 验证操作按钮点击触发回调

#### Task 2.2: 创建CrudForm组件

**文件**: `frontend/src/components/CrudForm.vue`

**步骤**:
1. 创建组件文件，定义Props接口（fields, initialData, onSubmit）
2. 实现动态字段渲染（根据field.type渲染不同组件）
3. 支持字段类型：text, email, select, radio, textarea
4. 实现表单验证（required, email, custom rules）
5. 实现提交处理（调用onSubmit回调）
6. 添加重置功能
7. 添加加载状态（提交时禁用按钮）

**验证**:
- 创建测试页面使用该组件
- 验证各种字段类型正常渲染
- 验证表单验证正常工作
- 验证提交回调正常触发

#### Task 2.3: 创建SearchBar组件

**文件**: `frontend/src/components/SearchBar.vue`

**步骤**:
1. 创建组件文件，定义Props接口（filters, onSearch）
2. 实现关键词搜索输入框
3. 实现过滤器渲染（根据filter.type渲染select/date等）
4. 实现搜索按钮（触发onSearch回调）
5. 实现重置按钮（清空所有条件）
6. 使用NSpace布局组件

**验证**:
- 创建测试页面使用该组件
- 验证搜索功能正常
- 验证重置功能正常
- 验证回调参数正确

---

### 阶段3：系统管理页面（前端）

#### Task 3.1: 实现用户管理页面

**文件**: `frontend/src/views/system/UserList.vue`

**步骤**:
1. 导入CrudTable、CrudForm、SearchBar组件
2. 配置表格列（id, username, email, department, status, roles）
3. 配置表单字段（username, email, password, departmentId, status）
4. 配置搜索过滤器（keyword, departmentId, status）
5. 实现fetchData函数（调用getUserList API）
6. 实现新增/编辑Modal（使用CrudForm）
7. 实现删除确认（使用NPopconfirm）
8. 实现用户-角色分配Modal（多选角色列表）
9. 添加批量删除功能
10. 添加权限控制（create, update, delete按钮）

**验证**:
- 运行前端服务，访问/system/users
- 验证列表正常加载
- 验证搜索功能正常
- 验证新增用户功能
- 验证编辑用户功能
- 验证删除用户功能
- 验证角色分配功能

#### Task 3.2: 实现部门管理页面

**文件**: `frontend/src/views/system/DepartmentList.vue`

**步骤**:
1. 导入CrudTable、CrudForm、SearchBar组件
2. 配置表格列（id, name, parentName, memberCount）
3. 配置表单字段（name, parentId, description）
4. 配置搜索过滤器（keyword）
5. 实现fetchData函数（调用getDepartmentList API）
6. 实现新增/编辑Modal
7. 实现删除确认（检查是否有子部门）
8. 添加权限控制

**验证**:
- 访问/system/departments
- 验证列表正常加载
- 验证CRUD功能正常

#### Task 3.3: 实现角色管理页面

**文件**: `frontend/src/views/system/RoleList.vue`

**步骤**:
1. 导入CrudTable、CrudForm、SearchBar组件
2. 配置表格列（id, name, type, description, permissionCount）
3. 配置表单字段（name, type, description）
4. 配置搜索过滤器（keyword, type）
5. 实现fetchData函数（调用getRoleList API）
6. 实现新增/编辑Modal
7. 实现删除确认
8. 实现角色-权限分配Modal（树形权限列表，按resource分组）
9. 添加权限控制

**验证**:
- 访问/system/roles
- 验证列表正常加载
- 验证CRUD功能正常
- 验证权限分配功能

#### Task 3.4: 实现权限管理页面

**文件**: `frontend/src/views/system/PermissionList.vue`

**步骤**:
1. 导入CrudTable、SearchBar组件
2. 配置表格列（id, name, resource, action, description）
3. 配置搜索过滤器（keyword, resource）
4. 实现fetchData函数（调用getPermissionList API）
5. 只读展示，不提供新增/编辑/删除功能

**验证**:
- 访问/system/permissions
- 验证列表正常加载
- 验证搜索功能正常

---

### 阶段4：导航调整与优化

#### Task 4.1: 调整导航结构

**文件**: `frontend/src/layouts/MainLayout.vue`

**步骤**:
1. 修改allMenuOptions数组
2. 将"平台能力"下的7个子页面提升到顶级
3. 删除"平台能力"父级菜单项
4. 保持其他菜单结构不变

**验证**:
- 运行前端服务
- 检查侧边栏导航结构正确
- 验证所有页面可正常访问

#### Task 4.2: 实现批量删除功能

**文件**: 修改各个List页面

**步骤**:
1. 在CrudTable组件中添加selectedRows状态
2. 添加批量删除按钮（在表格上方）
3. 实现批量删除逻辑（循环调用delete API）
4. 显示删除进度
5. 完成后刷新列表

**验证**:
- 在用户管理页面勾选多个用户
- 点击批量删除
- 验证删除成功

#### Task 4.3: 实现数据导出功能

**文件**: 修改各个List页面

**步骤**:
1. 添加导出按钮（在表格上方）
2. 实现导出逻辑（调用API获取全部数据）
3. 使用xlsx库生成Excel文件
4. 触发浏览器下载

**验证**:
- 点击导出按钮
- 验证下载Excel文件
- 打开文件检查数据正确

#### Task 4.4: 集成测试

**步骤**:
1. 清空Redis缓存
2. 重启后端服务
3. 重启前端服务
4. 完整测试登录流程
5. 测试页面刷新不丢失登录状态
6. 测试所有系统管理页面功能
7. 测试权限控制正常
8. 测试修改权限后刷新页面获取最新权限
9. 记录并修复发现的bug

**验证**:
- 所有功能正常工作
- 无明显bug
- 用户体验良好

## 注意事项

1. **最小化改动**: 每个任务只修改必要的代码，避免过度工程
2. **增量验证**: 每完成一个任务立即验证，不要累积问题
3. **保持简洁**: 组件设计要简单实用，不要添加不必要的功能
4. **错误处理**: 所有API调用都要有错误处理
5. **权限控制**: 所有操作按钮都要根据权限显示/隐藏
6. **用户体验**: 加载状态、空状态、错误提示都要完善

## 成功标准

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
