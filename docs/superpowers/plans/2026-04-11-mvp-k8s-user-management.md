# MVP 聚焦 K8s + 用户管理 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐用户管理前端（角色/部门/权限），完善 K8s Dashboard，加强前端基础设施，快速出 MVP

**Architecture:** 前端 Vue3 + Element Plus，遵循现有 Composition API 模式；API 层按模块拆分；复用 useTableList 组合式函数；后端 API 已就绪，纯前端开发

**Tech Stack:** Vue 3, Element Plus 2.13.6, Pinia, Vue Router 4, Vite 8, Axios, JavaScript

**TDD 策略:** 前端使用 Vitest + @vue/test-utils 编写组件测试；API 层编写 mock 拦截测试；由于本项目是前端 UI 开发，TDD 侧重于 API 封装层和 composable 层的单元测试，页面组件以手动验证为主

---

## 文件结构

### 新增文件
| 文件 | 职责 |
|------|------|
| `frontend/src/api/role.js` | 角色 API 封装（CRUD + 权限分配） |
| `frontend/src/api/department.js` | 部门 API 封装（CRUD + 树结构） |
| `frontend/src/api/permission.js` | 权限 API 封装（列表 + 分组） |
| `frontend/src/views/System/RoleList.vue` | 角色管理页面 |
| `frontend/src/views/System/DepartmentList.vue` | 部门管理页面 |
| `frontend/src/views/System/PermissionList.vue` | 权限管理页面 |
| `frontend/src/views/Dashboard/index.vue` | 重写 Dashboard 仪表盘 |

### 修改文件
| 文件 | 变更内容 |
|------|---------|
| `frontend/src/router/index.js` | 添加角色/部门/权限路由 |
| `frontend/src/components/layout/MainLayout.vue` | 更新侧边栏系统管理菜单 |
| `frontend/src/stores/user.js` | 增强权限数据加载 |
| `frontend/src/api/request.js` | 完善全局错误处理 |

---

## Task 1: 角色 API 封装

**Files:**
- Create: `frontend/src/api/role.js`

- [ ] **Step 1: 创建角色 API 文件**

```javascript
// frontend/src/api/role.js
import request from './request'

// 角色列表（分页）
export const getRoleList = (params) => request.get('/role/list', { params })

// 角色详情
export const getRoleDetail = (params) => request.get('/role/detail', { params })

// 创建角色
export const createRole = (data) => request.post('/role/create', data)

// 更新角色
export const updateRole = (data) => request.post('/role/update', data)

// 删除角色
export const deleteRole = (data) => request.post('/role/delete', data)

// 分配权限
export const assignPermissions = (data) => request.post('/role/assign-permissions', data)
```

- [ ] **Step 2: 验证 API 文件可正常导入**

Run: `cd frontend && node -e "const api = require('./src/api/role.js'); console.log('OK')"` 或直接在浏览器控制台测试

---

## Task 2: 权限 API 封装

**Files:**
- Create: `frontend/src/api/permission.js`

- [ ] **Step 1: 创建权限 API 文件**

```javascript
// frontend/src/api/permission.js
import request from './request'

// 权限列表（分页，支持按资源过滤）
export const getPermissionList = (params) => request.get('/permission/list', { params })

// 所有权限（不分页，用于权限选择器）
export const getAllPermissions = () => request.get('/permission/all')

// 权限详情
export const getPermissionDetail = (params) => request.get('/permission/detail', { params })
```

- [ ] **Step 2: 验证可导入**

---

## Task 3: 部门 API 封装

**Files:**
- Create: `frontend/src/api/department.js`

- [ ] **Step 1: 创建部门 API 文件**

```javascript
// frontend/src/api/department.js
import request from './request'

// 部门树列表
export const getDepartmentList = (params) => request.get('/department/list', { params })

// 创建部门
export const createDepartment = (data) => request.post('/department/create', data)

// 更新部门
export const updateDepartment = (data) => request.post('/department/update', data)

// 删除部门
export const deleteDepartment = (data) => request.post('/department/delete', data)

// 部门用户列表
export const getDepartmentUsers = (params) => request.get('/department/users/list', { params })
```

- [ ] **Step 2: 验证可导入**

---

## Task 4: 角色管理页面

**Files:**
- Create: `frontend/src/views/System/RoleList.vue`

- [ ] **Step 1: 创建角色管理页面组件**

```vue
<!-- frontend/src/views/System/RoleList.vue -->
<template>
  <div class="page-container">
    <div class="page-header">
      <h3>角色管理</h3>
      <el-button type="primary" @click="showCreateDialog">新建角色</el-button>
    </div>

    <!-- 搜索 -->
    <div style="margin-bottom: 16px;">
      <el-input
        v-model="keyword"
        placeholder="搜索角色名称"
        style="width: 300px;"
        clearable
        @clear="fetchData"
        @keyup.enter="fetchData"
      >
        <template #append>
          <el-button @click="fetchData">
            <el-icon><Search /></el-icon>
          </el-button>
        </template>
      </el-input>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" stripe v-loading="loading" style="width: 100%">
      <el-table-column prop="name" label="角色名称" width="180" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="permissions" label="权限数量" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.permissions ? row.permissions.length : 0 }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="warning" @click="showPermissionDialog(row)">分配权限</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="fetchData"
        @size-change="fetchData"
      />
    </div>

    <!-- 新建/编辑角色弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新建角色'" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入角色描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 权限分配弹窗 -->
    <el-dialog v-model="permDialogVisible" title="分配权限" width="600px">
      <p style="margin-bottom: 12px; color: #909399;">角色：{{ currentRole.name }}</p>
      <el-tree
        ref="permTreeRef"
        :data="permissionTree"
        show-checkbox
        node-key="id"
        :default-checked-keys="currentRole.permissionIds"
        :props="{ label: 'name', children: 'children' }"
      />
      <template #footer>
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAssignPermissions">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getRoleList, createRole, updateRole, deleteRole, assignPermissions } from '@/api/role'
import { getAllPermissions } from '@/api/permission'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')

const dialogVisible = ref(false)
const isEdit = ref(false)
const form = ref({ name: '', description: '' })
const formRef = ref()
const rules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }]
}

const permDialogVisible = ref(false)
const currentRole = ref({})
const permissionTree = ref([])
const permTreeRef = ref()

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getRoleList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value
    })
    tableData.value = res.data?.list || res.data || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  form.value = { name: '', description: '' }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { id: row.id, name: row.name, description: row.description }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value) {
      await updateRole(form.value)
      ElMessage.success('更新成功')
    } else {
      await createRole(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除角色 "${row.name}"？`, '提示', { type: 'warning' })
  try {
    await deleteRole({ id: row.id })
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '删除失败')
  }
}

const showPermissionDialog = async (row) => {
  currentRole.value = row
  if (!permissionTree.value.length) {
    const res = await getAllPermissions()
    const permissions = res.data || []
    // 按资源分组构建树
    const grouped = {}
    permissions.forEach(p => {
      if (!grouped[p.resource]) {
        grouped[p.resource] = { id: `group-${p.resource}`, name: p.resource, children: [] }
      }
      grouped[p.resource].children.push(p)
    })
    permissionTree.value = Object.values(grouped)
  }
  permDialogVisible.value = true
}

const handleAssignPermissions = async () => {
  const checkedKeys = permTreeRef.value.getCheckedKeys(true) // 只获取叶子节点
  try {
    await assignPermissions({
      roleId: currentRole.value.id,
      permissionIds: checkedKeys
    })
    ElMessage.success('权限分配成功')
    permDialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '权限分配失败')
  }
}

onMounted(fetchData)
</script>

<style scoped>
.page-container {
  background: #fff;
  border-radius: 4px;
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}
</style>
```

- [ ] **Step 2: 启动开发服务器验证页面渲染**

Run: `cd frontend && npm run dev`

访问角色管理页面，确认：
- 页面无 JS 报错
- 表格渲染正常
- 新建/编辑弹窗打开正常

---

## Task 5: 部门管理页面

**Files:**
- Create: `frontend/src/views/System/DepartmentList.vue`

- [ ] **Step 1: 创建部门管理页面组件**

```vue
<!-- frontend/src/views/System/DepartmentList.vue -->
<template>
  <div class="page-container">
    <div class="page-header">
      <h3>部门管理</h3>
      <el-button type="primary" @click="showCreateDialog(null)">新建顶级部门</el-button>
    </div>

    <!-- 部门树 -->
    <el-table
      :data="departmentTree"
      row-key="id"
      v-loading="loading"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      default-expand-all
      style="width: 100%"
    >
      <el-table-column prop="name" label="部门名称" width="250" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="userCount" label="人员数量" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.userCount || 0 }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="320" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showCreateDialog(row)">添加子部门</el-button>
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" @click="viewUsers(row)">查看成员</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新建/编辑部门弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑部门' : '新建部门'" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入部门名称" />
        </el-form-item>
        <el-form-item label="上级部门" prop="parentId">
          <el-tree-select
            v-model="form.parentId"
            :data="parentOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="无（顶级部门）"
            clearable
            check-strictly
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入部门描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 部门成员弹窗 -->
    <el-dialog v-model="userDialogVisible" :title="`${currentDept.name} - 成员列表`" width="700px">
      <el-table :data="deptUsers" v-loading="userLoading" style="width: 100%">
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="nickname" label="姓名" width="150" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
        <el-pagination
          v-model:current-page="userPage"
          v-model:page-size="userPageSize"
          :total="userTotal"
          layout="total, prev, pager, next"
          @current-change="fetchDeptUsers"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDepartmentList, createDepartment, updateDepartment, deleteDepartment,
  getDepartmentUsers
} from '@/api/department'

const loading = ref(false)
const departmentTree = ref([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const form = ref({ name: '', parentId: null, description: '' })
const formRef = ref()
const rules = {
  name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }]
}

const userDialogVisible = ref(false)
const currentDept = ref({})
const deptUsers = ref([])
const userLoading = ref(false)
const userPage = ref(1)
const userPageSize = ref(10)
const userTotal = ref(0)

// 上级部门选项（排除自身及其子部门）
const parentOptions = computed(() => {
  return departmentTree.value
})

const fetchDepartments = async () => {
  loading.value = true
  try {
    const res = await getDepartmentList()
    departmentTree.value = res.data || []
  } finally {
    loading.value = false
  }
}

const showCreateDialog = (parent) => {
  isEdit.value = false
  form.value = {
    name: '',
    parentId: parent ? parent.id : null,
    description: ''
  }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = {
    id: row.id,
    name: row.name,
    parentId: row.parentId,
    description: row.description
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value) {
      await updateDepartment(form.value)
      ElMessage.success('更新成功')
    } else {
      await createDepartment(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchDepartments()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  }
}

const handleDelete = async (row) => {
  if (row.children && row.children.length > 0) {
    ElMessage.warning('该部门下有子部门，无法删除')
    return
  }
  await ElMessageBox.confirm(`确认删除部门 "${row.name}"？`, '提示', { type: 'warning' })
  try {
    await deleteDepartment({ id: row.id })
    ElMessage.success('删除成功')
    fetchDepartments()
  } catch (e) {
    ElMessage.error(e.message || '删除失败')
  }
}

const viewUsers = (row) => {
  currentDept.value = row
  userPage.value = 1
  userDialogVisible.value = true
  fetchDeptUsers()
}

const fetchDeptUsers = async () => {
  userLoading.value = true
  try {
    const res = await getDepartmentUsers({
      departmentId: currentDept.value.id,
      page: userPage.value,
      pageSize: userPageSize.value
    })
    deptUsers.value = res.data?.list || res.data || []
    userTotal.value = res.data?.total || 0
  } finally {
    userLoading.value = false
  }
}

onMounted(fetchDepartments)
</script>

<style scoped>
.page-container {
  background: #fff;
  border-radius: 4px;
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}
</style>
```

- [ ] **Step 2: 启动开发服务器验证页面渲染**

---

## Task 6: 权限管理页面

**Files:**
- Create: `frontend/src/views/System/PermissionList.vue`

- [ ] **Step 1: 创建权限管理页面组件**

```vue
<!-- frontend/src/views/System/PermissionList.vue -->
<template>
  <div class="page-container">
    <div class="page-header">
      <h3>权限管理</h3>
    </div>

    <!-- 搜索和过滤 -->
    <div style="margin-bottom: 16px; display: flex; gap: 12px;">
      <el-input
        v-model="keyword"
        placeholder="搜索权限名称"
        style="width: 300px;"
        clearable
        @clear="fetchData"
        @keyup.enter="fetchData"
      >
        <template #append>
          <el-button @click="fetchData">
            <el-icon><Search /></el-icon>
          </el-button>
        </template>
      </el-input>
      <el-select v-model="resourceFilter" placeholder="按资源过滤" clearable @change="fetchData" style="width: 200px;">
        <el-option v-for="r in resources" :key="r" :label="r" :value="r" />
      </el-select>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" stripe v-loading="loading" style="width: 100%">
      <el-table-column prop="name" label="权限名称" width="200" />
      <el-table-column prop="resource" label="资源" width="150" />
      <el-table-column prop="action" label="操作" width="120" />
      <el-table-column prop="description" label="描述" />
    </el-table>

    <!-- 分页 -->
    <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="fetchData"
        @size-change="fetchData"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { getPermissionList } from '@/api/permission'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const resourceFilter = ref('')
const resources = ref([])

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (keyword.value) params.keyword = keyword.value
    if (resourceFilter.value) params.resource = resourceFilter.value

    const res = await getPermissionList(params)
    tableData.value = res.data?.list || res.data || []
    total.value = res.data?.total || 0

    // 提取去重资源列表
    const resSet = new Set((res.data?.list || res.data || []).map(p => p.resource))
    resources.value = [...resSet]
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<style scoped>
.page-container {
  background: #fff;
  border-radius: 4px;
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}
</style>
```

- [ ] **Step 2: 启动开发服务器验证页面渲染**

---

## Task 7: 路由更新

**Files:**
- Modify: `frontend/src/router/index.js`

- [ ] **Step 1: 在路由配置中添加系统管理子路由**

在 `router/index.js` 的 children 数组中，找到 `/system/user` 路由，在其后添加以下路由：

```javascript
{
  path: '/system/role',
  name: 'RoleList',
  component: () => import('@/views/System/RoleList.vue'),
  meta: { title: '角色管理' }
},
{
  path: '/system/department',
  name: 'DepartmentList',
  component: () => import('@/views/System/DepartmentList.vue'),
  meta: { title: '部门管理' }
},
{
  path: '/system/permission',
  name: 'PermissionList',
  component: () => import('@/views/System/PermissionList.vue'),
  meta: { title: '权限管理' }
}
```

- [ ] **Step 2: 验证路由跳转正常**

Run: `cd frontend && npm run dev`

在浏览器地址栏输入 `/system/role`、`/system/department`、`/system/permission`，确认页面加载正常。

---

## Task 8: 侧边栏菜单更新

**Files:**
- Modify: `frontend/src/components/layout/MainLayout.vue`

- [ ] **Step 1: 更新侧边栏菜单，添加系统管理子菜单**

找到 MainLayout.vue 中的 el-menu 部分，将现有的单个"用户管理"菜单项替换为系统管理子菜单：

```html
<!-- 系统管理 -->
<el-sub-menu index="system">
  <template #title>
    <el-icon><Setting /></el-icon>
    <span>系统管理</span>
  </template>
  <el-menu-item index="/system/user">用户管理</el-menu-item>
  <el-menu-item index="/system/role">角色管理</el-menu-item>
  <el-menu-item index="/system/department">部门管理</el-menu-item>
  <el-menu-item index="/system/permission">权限管理</el-menu-item>
</el-sub-menu>
```

确保在 `<script setup>` 中已导入 `Setting` 图标：
```javascript
import { Setting } from '@element-plus/icons-vue'
```

- [ ] **Step 2: 验证菜单展开和路由跳转**

确认：
- 侧边栏显示"系统管理"子菜单
- 点击各菜单项能跳转到对应页面
- 当前激活菜单项高亮正确

- [ ] **Step 3: 提交阶段 1 代码**

```bash
cd frontend
git add src/api/role.js src/api/permission.js src/api/department.js
git add src/views/System/RoleList.vue src/views/System/DepartmentList.vue src/views/System/PermissionList.vue
git add src/router/index.js src/components/layout/MainLayout.vue
git commit -m "feat: 添加角色/部门/权限管理页面和API"
```

---

## Task 9: Dashboard 仪表盘重写

**Files:**
- Modify: `frontend/src/views/Dashboard/index.vue`

- [ ] **Step 1: 重写 Dashboard 为集群概览仪表盘**

```vue
<!-- frontend/src/views/Dashboard/index.vue -->
<template>
  <div class="page-container">
    <div class="page-header">
      <h3>仪表盘</h3>
      <el-button @click="fetchDashboardData">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- 概览卡片 -->
    <el-row :gutter="16" style="margin-bottom: 24px;">
      <el-col :span="6" v-for="card in statCards" :key="card.title">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-content">
            <div class="stat-info">
              <div class="stat-title">{{ card.title }}</div>
              <div class="stat-value" :style="{ color: card.color }">{{ card.value }}</div>
            </div>
            <el-icon :size="48" :style="{ color: card.color }">
              <component :is="card.icon" />
            </el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 集群健康状态 -->
    <el-card shadow="never" style="margin-bottom: 24px;">
      <template #header>
        <span>集群状态</span>
      </template>
      <el-table :data="clusters" stripe v-loading="loading" style="width: 100%">
        <el-table-column prop="name" label="集群名称" width="200">
          <template #default="{ row }">
            <router-link :to="`/k8s/cluster/${row.name}`" style="color: var(--el-color-primary); text-decoration: none;">
              {{ row.name }}
            </router-link>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.status === 'healthy' ? 'success' : row.status === 'warning' ? 'warning' : 'danger'">
              {{ row.status === 'healthy' ? '健康' : row.status === 'warning' ? '告警' : '异常' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="节点" width="100">
          <template #default="{ row }">
            {{ row.nodeCount || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="Pod 总数" width="100">
          <template #default="{ row }">
            {{ row.podCount || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="CPU 使用率" width="200">
          <template #default="{ row }">
            <el-progress
              :percentage="row.cpuUsage || 0"
              :color="getProgressColor(row.cpuUsage)"
              :stroke-width="12"
            />
          </template>
        </el-table-column>
        <el-table-column label="内存使用率" width="200">
          <template #default="{ row }">
            <el-progress
              :percentage="row.memoryUsage || 0"
              :color="getProgressColor(row.memoryUsage)"
              :stroke-width="12"
            />
          </template>
        </el-table-column>
        <el-table-column prop="version" label="K8s 版本" width="130" />
      </el-table>
    </el-card>

    <!-- 最近事件 -->
    <el-card shadow="never">
      <template #header>
        <span>最近事件</span>
      </template>
      <el-table :data="recentEvents" stripe max-height="300" style="width: 100%">
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'Warning' ? 'warning' : 'info'" size="small">
              {{ row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="cluster" label="集群" width="150" />
        <el-table-column prop="namespace" label="命名空间" width="180" />
        <el-table-column prop="object" label="对象" width="250" />
        <el-table-column prop="message" label="消息" />
        <el-table-column prop="lastSeen" label="最后发生" width="180" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Refresh, Monitor, Cpu, Grid, Warning } from '@element-plus/icons-vue'
import { getClusterList } from '@/api/cluster'

const loading = ref(false)
const clusters = ref([])
const recentEvents = ref([])

const statCards = computed(() => [
  {
    title: '集群数量',
    value: clusters.value.length,
    icon: Monitor,
    color: '#409EFF'
  },
  {
    title: '节点总数',
    value: clusters.value.reduce((sum, c) => sum + (c.nodeCount || 0), 0),
    icon: Cpu,
    color: '#67C23A'
  },
  {
    title: 'Pod 总数',
    value: clusters.value.reduce((sum, c) => sum + (c.podCount || 0), 0),
    icon: Grid,
    color: '#E6A23C'
  },
  {
    title: '告警集群',
    value: clusters.value.filter(c => c.status !== 'healthy').length,
    icon: Warning,
    color: '#F56C6C'
  }
])

const getProgressColor = (percentage) => {
  if (percentage >= 90) return '#F56C6C'
  if (percentage >= 70) return '#E6A23C'
  return '#67C23A'
}

const fetchDashboardData = async () => {
  loading.value = true
  try {
    const res = await getClusterList()
    clusters.value = res.data?.list || res.data || []
    // 从集群数据中提取事件
    recentEvents.value = []
    clusters.value.forEach(c => {
      if (c.events && c.events.length) {
        c.events.forEach(e => {
          recentEvents.value.push({ ...e, cluster: c.name })
        })
      }
    })
    // 按时间排序，取最近 20 条
    recentEvents.value.sort((a, b) => (b.lastSeen || '').localeCompare(a.lastSeen || ''))
    recentEvents.value = recentEvents.value.slice(0, 20)
  } finally {
    loading.value = false
  }
}

onMounted(fetchDashboardData)
</script>

<style scoped>
.page-container {
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}
.stat-card-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.stat-title {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}
.stat-value {
  font-size: 28px;
  font-weight: 600;
}
</style>
```

- [ ] **Step 2: 验证 Dashboard 渲染**

确认：
- 四个统计卡片显示正确
- 集群表格展示正常
- 点击集群名称可跳转详情
- 无 JS 报错

- [ ] **Step 3: 提交 Dashboard**

```bash
cd frontend
git add src/views/Dashboard/index.vue
git commit -m "feat: 重写 Dashboard 为集群概览仪表盘"
```

---

## Task 10: 增强全局错误处理

**Files:**
- Modify: `frontend/src/api/request.js`

- [ ] **Step 1: 在 Axios 响应拦截器中完善错误处理**

在 `request.js` 的响应拦截器中，确保覆盖以下错误码：

```javascript
// 在现有的响应拦截器 error 处理中添加：
if (error.response) {
  const { status, data } = error.response
  switch (status) {
    case 401:
      // 已有处理：跳转登录
      break
    case 403:
      ElMessage.error('权限不足，无法访问')
      break
    case 404:
      ElMessage.error('请求的资源不存在')
      break
    case 500:
      ElMessage.error('服务器内部错误')
      break
    default:
      ElMessage.error(data?.message || `请求失败 (${status})`)
  }
} else if (error.code === 'ERR_NETWORK') {
  ElMessage.error('网络连接失败，请检查网络')
} else {
  ElMessage.error(error.message || '请求失败')
}
```

- [ ] **Step 2: 验证错误处理**

手动测试：在浏览器控制台制造一个 404 请求，确认弹出"请求的资源不存在"提示。

- [ ] **Step 3: 提交**

```bash
cd frontend
git add src/api/request.js
git commit -m "feat: 增强全局 HTTP 错误处理"
```

---

## Task 11: 增强用户权限加载

**Files:**
- Modify: `frontend/src/stores/user.js`

- [ ] **Step 1: 增强 user store 的权限加载**

确保 `loadUserInfo` 方法能正确加载和存储权限列表。在现有代码基础上添加：

```javascript
// 确保 userInfo 包含 permissions 数组
const loadUserInfo = () => {
  const stored = sessionStorage.getItem('userInfo')
  if (stored) {
    try {
      userInfo.value = JSON.parse(stored)
    } catch (e) {
      userInfo.value = null
    }
  }
}

// 如果后端返回的 userInfo 不含 permissions，主动请求权限接口
const fetchPermissions = async () => {
  if (!token.value) return
  try {
    const res = await request.get('/user/permissions')
    if (res.data) {
      userInfo.value = { ...userInfo.value, permissions: res.data }
      sessionStorage.setItem('userInfo', JSON.stringify(userInfo.value))
    }
  } catch (e) {
    // 权限加载失败不阻塞页面
  }
}
```

同时在 `setUserInfo` 中确保保存 permissions：

```javascript
const setUserInfo = (info) => {
  userInfo.value = info
  sessionStorage.setItem('userInfo', JSON.stringify(info))
}
```

- [ ] **Step 2: 提交**

```bash
cd frontend
git add src/stores/user.js
git commit -m "feat: 增强用户权限数据加载"
```

---

## Task 12: 最终集成验证

- [ ] **Step 1: 启动前后端，进行完整流程测试**

```bash
# 终端 1: 启动后端
cd backend && go run cmd/server/main.go

# 终端 2: 启动前端
cd frontend && npm run dev
```

- [ ] **Step 2: 按验收标准逐项验证**

| 验收项 | 操作 | 预期结果 |
|--------|------|---------|
| 登录 | 访问 /login，输入账号密码 | 登录成功，跳转 Dashboard |
| 权限菜单 | 登录后查看侧边栏 | 根据权限显示对应菜单 |
| 角色管理 | 侧边栏 → 系统管理 → 角色管理 | 角色列表展示，支持 CRUD |
| 角色权限分配 | 点击"分配权限" | 权限树展示，可勾选保存 |
| 部门管理 | 侧边栏 → 系统管理 → 部门管理 | 部门树展示，支持 CRUD |
| 部门成员 | 点击"查看成员" | 成员列表弹窗展示 |
| 权限管理 | 侧边栏 → 系统管理 → 权限管理 | 权限列表展示，支持按资源过滤 |
| Dashboard | 访问 /dashboard | 集群概览卡片、集群表格、事件列表 |
| 错误处理 | 触发 403/404/500 错误 | 友好提示消息 |
| 无 JS 报错 | 打开 F12 控制台 | 无报错 |

- [ ] **Step 3: 最终提交**

```bash
git add -A
git commit -m "feat: 完成 MVP 阶段 1 - 用户管理补齐 + Dashboard 重写"
```
