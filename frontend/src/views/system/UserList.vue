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
