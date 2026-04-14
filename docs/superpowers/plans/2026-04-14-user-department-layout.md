# 用户管理页面增加部门关联 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将用户管理页面改造为左侧部门树 + 右侧用户表格的主从布局，支持按部门筛选用户。

**Architecture:** 后端在现有 `GET /user/list` API 上增加 `departmentId` 查询参数，复用已有的 `ListByDepartmentInTenant` repo 方法。前端重写 `UserList.vue`，使用 Element Plus 的 `el-tree` + `el-table` + `el-pagination` 构建主从布局。

**Tech Stack:** Go (Gin + GORM) / Vue 3 + Element Plus + dayjs

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `backend/internal/modules/user/api/user.go:134-156` | Modify | `List` handler 增加 `departmentId` 参数解析 |
| `backend/internal/modules/user/service/user_service.go:110-119` | Modify | `ListUsers` 增加 `departmentID` 参数，按需路由 |
| `frontend/src/views/System/UserList.vue` | Rewrite | 左右主从布局、部门树、用户表格、搜索分页、编辑弹窗加部门选择器 |

---

### Task 1: 后端 — 增加 departmentId 筛选参数

**Files:**
- Modify: `backend/internal/modules/user/api/user.go:134-156`
- Modify: `backend/internal/modules/user/service/user_service.go:110-119`

- [ ] **Step 1: 修改 user service 的 ListUsers 方法签名**

在 `backend/internal/modules/user/service/user_service.go` 中，修改 `ListUsers` 方法，增加 `departmentID *uint` 参数。当 `departmentID` 有值时，直接调用已有的 `ListByDepartmentInTenant`。

```go
// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, tenantID uint, operatorID uint, page, pageSize int, keyword string, departmentID *uint) ([]model.User, int64, error) {
	// 如果指定了部门ID，直接按部门筛选
	if departmentID != nil {
		return s.userRepo.ListByDepartmentInTenant(tenantID, *departmentID, page, pageSize, keyword)
	}

	scope, err := s.scopeSvc.Resolve(ctx, tenantID, operatorID)
	if err != nil {
		return nil, 0, err
	}
	if scope.AllowsAll() {
		return s.userRepo.ListInTenant(tenantID, page, pageSize, keyword)
	}
	return s.userRepo.ListByDepartmentIDsInTenant(tenantID, scope.DepartmentIDs, page, pageSize, keyword)
}
```

- [ ] **Step 2: 修改 user handler 的 List 方法**

在 `backend/internal/modules/user/api/user.go` 的 `List` 函数中，解析 `departmentId` 查询参数并传递给 service。

```go
func List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	tenantID := GetCurrentTenantID(c)
	operatorID := GetCurrentUserID(c)

	// 解析可选的部门ID筛选参数
	var departmentID *uint
	if deptIDStr := c.Query("departmentId"); deptIDStr != "" {
		if id, err := strconv.ParseUint(deptIDStr, 10, 32); err == nil {
			uid := uint(id)
			departmentID = &uid
		}
	}

	users, total, err := getService().ListUsers(c.Request.Context(), tenantID, operatorID, page, pageSize, keyword, departmentID)
	if err != nil {
		logger.Log.Error("Failed to list users", zap.Error(err))
		writeModuleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  users,
			"total": total,
			"page":  page,
		},
	})
}
```

- [ ] **Step 3: 检查其他 ListUsers 调用点并修复编译**

搜索代码库中所有调用 `ListUsers` 的地方，确保都加上了新的 `departmentID` 参数（传 `nil` 即可保持原行为）。

Run: `cd D:/owner/code/devops/backend && grep -rn "ListUsers(" --include="*.go"`

修复所有调用点后验证编译：
Run: `cd D:/owner/code/devops/backend && go build ./...`
Expected: 编译成功，无错误

- [ ] **Step 4: Commit**

```bash
cd D:/owner/code/devops
git add backend/internal/modules/user/api/user.go backend/internal/modules/user/service/user_service.go
git commit -m "feat: add departmentId filter to user list API"
```

---

### Task 2: 前端 — 重写 UserList.vue

**Files:**
- Modify: `frontend/src/views/System/UserList.vue`

- [ ] **Step 1: 重写 UserList.vue 完整文件**

用以下内容替换 `frontend/src/views/System/UserList.vue` 的全部内容：

```vue
<template>
  <div class="user-page">
    <div class="page-header">
      <h3>用户管理</h3>
    </div>

    <div class="page-body">
      <!-- 左侧部门树 -->
      <div class="dept-panel">
        <div class="dept-tree-header">部门列表</div>
        <el-tree
          ref="treeRef"
          :data="deptTree"
          node-key="id"
          :props="{ label: 'name', children: 'children' }"
          default-expand-all
          highlight-current
          :expand-on-click-node="false"
          @node-click="handleDeptClick"
        >
          <template #default="{ node, data }">
            <span class="dept-node">
              <span>{{ data.name }}</span>
              <span class="dept-count">{{ data.userCount || 0 }}</span>
            </span>
          </template>
        </el-tree>
        <div class="dept-all" :class="{ active: !selectedDeptId }" @click="handleShowAll">
          全部用户
        </div>
      </div>

      <!-- 右侧用户表格 -->
      <div class="user-panel">
        <div class="user-toolbar">
          <el-input
            v-model="keyword"
            placeholder="搜索用户名/邮箱"
            clearable
            style="width: 240px"
            @input="handleSearch"
          />
          <el-button type="primary" @click="showDialog()">新建用户</el-button>
        </div>

        <el-table :data="tableData" v-loading="loading" stripe>
          <el-table-column prop="username" label="用户名" min-width="120" />
          <el-table-column prop="email" label="邮箱" min-width="180" />
          <el-table-column label="部门" min-width="120">
            <template #default="{ row }">{{ row.department?.name || '-' }}</template>
          </el-table-column>
          <el-table-column label="角色" min-width="140">
            <template #default="{ row }">
              <template v-if="row.roles && row.roles.length">
                <el-tag v-for="role in row.roles" :key="role.id" size="small" style="margin-right:4px">{{ role.name }}</el-tag>
              </template>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
                {{ row.status === 'active' ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="170">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column label="最后登录" width="170">
            <template #default="{ row }">{{ formatTime(row.lastLoginAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="showDialog(row)">编辑</el-button>
              <el-button link type="danger" @click="handleDelete(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrap">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            @size-change="fetchUsers"
            @current-change="fetchUsers"
          />
        </div>
      </div>
    </div>

    <!-- 创建/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑用户' : '创建用户'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="部门">
          <el-tree-select
            v-model="form.departmentId"
            :data="deptTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择部门"
            clearable
            check-strictly
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!form.id">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getUserList, createUser, updateUser, deleteUser, getDepartmentList } from '../../api/system'
import { ElMessage, ElMessageBox } from 'element-plus'
import { required, email } from '../../utils/validate'
import dayjs from 'dayjs'

// --- 部门树 ---
const treeRef = ref()
const deptTree = ref([])

const fetchDepartments = async () => {
  try {
    const res = await getDepartmentList()
    deptTree.value = res.data || []
  } catch {
    deptTree.value = []
  }
}

// --- 用户列表 ---
const tableData = ref([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const selectedDeptId = ref(null)

let searchTimer = null
const handleSearch = () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchUsers()
  }, 300)
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value,
    }
    if (keyword.value) params.keyword = keyword.value
    if (selectedDeptId.value) params.departmentId = selectedDeptId.value
    const res = await getUserList(params)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } finally {
    loading.value = false
  }
}

const handleDeptClick = (data) => {
  selectedDeptId.value = data.id
  page.value = 1
  fetchUsers()
}

const handleShowAll = () => {
  selectedDeptId.value = null
  if (treeRef.value) treeRef.value.setCurrentKey(null)
  page.value = 1
  fetchUsers()
}

// --- 创建/编辑 ---
const dialogVisible = ref(false)
const form = ref({})
const formRef = ref()
const saving = ref(false)

const rules = {
  username: [required('请输入用户名')],
  email: [required('请输入邮箱'), email()],
  password: [required('请输入密码')]
}

const showDialog = (row) => {
  form.value = row ? { ...row } : {}
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

const handleSave = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    if (form.value.id) {
      await updateUser(form.value)
    } else {
      await createUser(form.value)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetchUsers()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该用户吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteUser(id)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// --- 工具 ---
const formatTime = (val) => {
  if (!val) return '-'
  return dayjs(val).format('YYYY-MM-DD HH:mm')
}

onMounted(() => {
  fetchDepartments()
  fetchUsers()
})
</script>

<style scoped>
.user-page {
  background: #fff;
  border-radius: 4px;
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}
.page-body {
  display: flex;
  gap: 16px;
}
.dept-panel {
  width: 240px;
  flex-shrink: 0;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  display: flex;
  flex-direction: column;
}
.dept-tree-header {
  padding: 12px 16px;
  font-weight: 500;
  border-bottom: 1px solid #e4e7ed;
  background: #f5f7fa;
  border-radius: 4px 4px 0 0;
}
.dept-panel :deep(.el-tree) {
  padding: 8px;
}
.dept-node {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex: 1;
  font-size: 13px;
}
.dept-count {
  color: #909399;
  font-size: 12px;
}
.dept-all {
  padding: 8px 16px;
  cursor: pointer;
  font-size: 13px;
  border-top: 1px solid #e4e7ed;
  color: #606266;
}
.dept-all:hover,
.dept-all.active {
  color: #409eff;
  background: #ecf5ff;
}
.user-panel {
  flex: 1;
  min-width: 0;
}
.user-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
```

- [ ] **Step 2: 验证前端编译**

Run: `cd D:/owner/code/devops/frontend && npx vite build --mode development 2>&1 | tail -5`
Expected: 构建成功，无错误

- [ ] **Step 3: Commit**

```bash
cd D:/owner/code/devops
git add frontend/src/views/System/UserList.vue
git commit -m "feat: rewrite UserList with department tree panel, search, and pagination"
```

---

### Task 3: 验收测试

- [ ] **Step 1: 启动后端确认编译通过**

Run: `cd D:/owner/code/devops/backend && go build ./...`
Expected: 编译成功

- [ ] **Step 2: 浏览器手动验证**

1. 打开 `http://localhost:3000/system/user`
2. 确认左侧显示部门树，右侧显示用户表格
3. 点击某个部门 → 确认右侧只显示该部门用户
4. 点击「全部用户」→ 确认显示所有用户
5. 在搜索框输入关键词 → 确认搜索生效
6. 点击「新建用户」→ 确认弹窗中有部门选择器
7. 编辑某个用户 → 确认部门选择器回显正确值
8. 确认分页正常工作
