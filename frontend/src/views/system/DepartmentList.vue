<template>
  <div class="page-container">
    <div class="page-header">
      <h3>部门管理</h3>
      <el-button type="primary" @click="showCreateDialog(null)">新建顶级部门</el-button>
    </div>

    <el-table :data="departmentTree" row-key="id" v-loading="loading" :tree-props="{ children: 'children', hasChildren: 'hasChildren' }" default-expand-all style="width: 100%">
      <el-table-column prop="name" label="部门名称" width="250" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="userCount" label="人员数量" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.userCount || 0 }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="320" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showCreateDialog(row)">添加子部门</el-button>
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" @click="viewUsers(row)">查看成员</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑部门' : '新建部门'" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="部门名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入部门名称" />
        </el-form-item>
        <el-form-item label="上级部门" prop="parentId">
          <el-tree-select v-model="form.parentId" :data="parentOptions" :props="{ label: 'name', value: 'id', children: 'children' }" placeholder="无（顶级部门）" clearable check-strictly style="width: 100%" />
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

    <el-dialog v-model="userDialogVisible" :title="`${currentDept.name} - 成员列表`" width="700px">
      <el-table :data="deptUsers" v-loading="userLoading" style="width: 100%">
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="nickname" label="姓名" width="150" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'">{{ row.status === 'active' ? '正常' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
        <el-pagination v-model:current-page="userPage" v-model:page-size="userPageSize" :total="userTotal" layout="total, prev, pager, next" @current-change="fetchDeptUsers" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDepartmentList, createDepartment, updateDepartment, deleteDepartment, getDepartmentUsers } from '@/api/department'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const departmentTree = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const form = ref({ name: '', parentId: null, description: '' })
const formRef = ref()
const rules = { name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }] }
const userDialogVisible = ref(false)
const currentDept = ref({})
const deptUsers = ref([])
const userLoading = ref(false)
const userPage = ref(1)
const userPageSize = ref(10)
const userTotal = ref(0)

const parentOptions = computed(() => departmentTree.value)

const fetchDepartments = async () => {
  loading.value = true
  try { const res = await getDepartmentList(); departmentTree.value = res.data || [] }
  finally { loading.value = false }
}

const showCreateDialog = (parent) => {
  isEdit.value = false
  form.value = { name: '', parentId: parent ? parent.id : null, description: '' }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { id: row.id, name: row.name, parentId: row.parentId, description: row.description }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value) { await updateDepartment(form.value); ElMessage.success('更新成功') }
    else { await createDepartment(form.value); ElMessage.success('创建成功') }
    dialogVisible.value = false; fetchDepartments()
  } catch (e) { ElMessage.error(e.message || '操作失败') }
}

const handleDelete = async (row) => {
  if (row.children && row.children.length > 0) { ElMessage.warning('该部门下有子部门，无法删除'); return }
  await ElMessageBox.confirm(`确认删除部门 "${row.name}"？`, '提示', { type: 'warning' })
  try { await deleteDepartment({ id: row.id }); ElMessage.success('删除成功'); fetchDepartments() }
  catch (e) { ElMessage.error(e.message || '删除失败') }
}

const viewUsers = (row) => { currentDept.value = row; userPage.value = 1; userDialogVisible.value = true; fetchDeptUsers() }

const fetchDeptUsers = async () => {
  userLoading.value = true
  try {
    const res = await getDepartmentUsers({ departmentId: currentDept.value.id, page: userPage.value, pageSize: userPageSize.value })
    deptUsers.value = res.data?.list || res.data || []
    userTotal.value = res.data?.total || 0
  } finally { userLoading.value = false }
}

onMounted(fetchDepartments)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
</style>
