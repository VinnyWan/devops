# MVP 设计方案：聚焦 K8s 管理 + 用户管理

**日期**: 2026-04-11
**状态**: 已批准
**目标**: 快速出 MVP，聚焦 K8s 核心管理功能，补齐用户管理前端

## 背景

### 项目现状
- 后端架构成熟，user/k8s/app 模块基本完成
- 前端经历了一次从 TypeScript 到 JavaScript 的重构，当前约 2000 行代码
- K8s 模块前端完成度约 40%，用户管理前端仅用户列表页面
- App/CICD/Harbor/Log/Monitor 模块前端完全缺失

### MVP 范围选择
选择**聚焦 K8s 管理**，理由：
1. K8s 模块后端 90% 完成，前端 40% 完成，最接近可用
2. 用户管理是平台基础，缺角色/部门/权限管理前端
3. 工作量最小，能最快出可用产品

### 不在范围内
- App 应用管理前端
- CICD/Harbor/Log/Monitor 模块
- 不引入新技术栈

## 技术方案

### 技术栈（保持现有）
- Vue 3 + Composition API (`<script setup>`)
- Element Plus 2.13.6
- Pinia 状态管理
- Vue Router 4
- Vite 8
- Axios HTTP 请求
- JavaScript（非 TypeScript）

### 代码模式（遵循现有）
- 复用 `useTableList` 组合式函数处理列表页
- API 层按模块拆分到 `src/api/`
- 状态管理用 Pinia stores
- 组件命名大驼峰（PascalCase）
- 路由懒加载

## 实施阶段

### 阶段 1：用户管理补齐

#### 1.1 角色管理页面
- **文件**: `src/views/System/RoleList.vue`
- **API**: `src/api/role.js`（新增）
- **功能**:
  - 角色列表（表格展示：角色名、描述、权限数量、创建时间）
  - 新建角色（弹窗表单：名称、描述、权限勾选）
  - 编辑角色（同新建表单，回填数据）
  - 删除角色（确认对话框）
  - 权限分配（树形权限选择器）
- **后端 API**:
  - `GET /api/v1/roles` - 角色列表
  - `POST /api/v1/roles` - 创建角色
  - `PUT /api/v1/roles/:id` - 更新角色
  - `DELETE /api/v1/roles/:id` - 删除角色
  - `GET /api/v1/permissions` - 权限列表

#### 1.2 部门管理页面
- **文件**: `src/views/System/DepartmentList.vue`
- **API**: `src/api/department.js`（新增）
- **功能**:
  - 部门树形展示（支持展开/折叠）
  - 新建部门（弹窗表单：名称、父部门选择、负责人）
  - 编辑部门
  - 删除部门（检查是否有子部门/用户）
  - 查看部门下用户列表
- **后端 API**:
  - `GET /api/v1/departments` - 部门树
  - `POST /api/v1/departments` - 创建部门
  - `PUT /api/v1/departments/:id` - 更新部门
  - `DELETE /api/v1/departments/:id` - 删除部门
  - `GET /api/v1/departments/:id/users` - 部门用户

#### 1.3 权限管理页面
- **文件**: `src/views/System/PermissionList.vue`
- **API**: 复用 `src/api/role.js`
- **功能**:
  - 权限列表展示
  - 按模块分组查看权限

#### 1.4 路由更新
- 在 `src/router/index.js` 添加：
  - `/system/role` - 角色管理
  - `/system/department` - 部门管理
  - `/system/permission` - 权限管理

#### 1.5 侧边栏菜单更新
- 在 MainLayout 中添加"系统管理"子菜单：
  - 用户管理
  - 角色管理
  - 部门管理
  - 权限管理

### 阶段 2：K8s 功能完善

#### 2.1 Dashboard 仪表盘
- **文件**: `src/views/Dashboard/index.vue`（重写）
- **功能**:
  - 集群概览卡片（集群数量、节点总数、Pod 总数、告警数）
  - 资源使用率图表（CPU/内存使用率，用进度条展示）
  - 最近事件列表
  - 各集群健康状态一览
- **API**: 复用 `src/api/cluster.js`

#### 2.2 存储管理页面
- **文件**: `src/views/k8s/StorageList.vue`（新增）
- **API**: `src/api/storage.js`（新增，需后端支持）
- **功能**:
  - PV（持久卷）列表
  - PVC（持久卷声明）列表
  - StorageClass 列表
  - Tab 切换三种资源
- **注意**: 需确认后端是否已有存储相关 API

#### 2.3 集群详情页增强
- **文件**: `src/views/k8s/ClusterDetail.vue`（增强）
- **新增功能**:
  - 节点资源使用率图表（CPU/内存/磁盘）
  - 实时事件流展示
  - 集群资源配额展示

#### 2.4 工作负载详情页
- **文件**: `src/views/k8s/WorkloadDetail.vue`（新增）
- **功能**:
  - 工作负载基本信息展示
  - YAML 查看/编辑
  - 关联 Pod 列表
  - 事件列表
  - 容器日志查看

### 阶段 3：前端基础设施加强

#### 3.1 全局错误处理
- 在 Axios 拦截器中统一处理业务错误码
- 网络错误友好提示
- 403 权限不足提示

#### 3.2 动态权限菜单
- 登录后根据用户权限动态生成侧边栏菜单
- 无权限的菜单项隐藏（而非点击后报错）
- 利用现有 `stores/user.js` 的 `hasPermission` 方法

#### 3.3 响应式布局优化
- 侧边栏折叠/展开
- 表格自适应列宽
- 移动端基础适配

## 文件变更清单

### 新增文件
| 文件 | 说明 |
|------|------|
| `src/api/role.js` | 角色 API |
| `src/api/department.js` | 部门 API |
| `src/api/storage.js` | 存储 API |
| `src/views/System/RoleList.vue` | 角色管理页 |
| `src/views/System/DepartmentList.vue` | 部门管理页 |
| `src/views/System/PermissionList.vue` | 权限管理页 |
| `src/views/k8s/StorageList.vue` | 存储管理页 |
| `src/views/k8s/WorkloadDetail.vue` | 工作负载详情页 |

### 修改文件
| 文件 | 变更内容 |
|------|---------|
| `src/router/index.js` | 添加新页面路由 |
| `src/components/layout/MainLayout.vue` | 更新侧边栏菜单 |
| `src/views/Dashboard/index.vue` | 重写为仪表盘 |
| `src/views/k8s/ClusterDetail.vue` | 增强功能 |
| `src/stores/user.js` | 增强权限数据加载 |
| `src/api/request.js` | 完善错误处理 |

## 验收标准

1. 用户能登录系统并看到基于权限的菜单
2. 能完成角色 CRUD 和权限分配
3. 能完成部门 CRUD 和树形展示
4. Dashboard 展示 K8s 集群概览信息
5. 能管理 K8s 存储资源（PV/PVC/StorageClass）
6. 能查看工作负载详情和 YAML
7. 所有页面无 JS 报错
