<template>
  <div class="page-container">
    <div class="page-header">
      <h3>角色管理</h3>
      <el-button type="primary" @click="showCreateDialog">新建角色</el-button>
    </div>

    <div style="margin-bottom: 16px;">
      <el-input v-model="keyword" placeholder="搜索角色名称" style="width: 300px;" clearable @clear="fetchData" @keyup.enter="fetchData">
        <template #append>
          <el-button @click="fetchData"><el-icon><Search /></el-icon></el-button>
        </template>
      </el-input>
    </div>

    <el-table :data="tableData" stripe v-loading="loading" style="width: 100%">
      <el-table-column prop="name" label="角色名称" width="180" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="permissions" label="权限数量" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.permissions ? row.permissions.length : 0 }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="warning" @click="showPermissionDialog(row)">分配权限</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next" @current-change="fetchData" @size-change="fetchData" />
    </div>

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

    <el-dialog v-model="permDialogVisible" title="分配权限" width="600px">
      <p style="margin-bottom: 12px; color: #909399;">角色：{{ currentRole.name }}</p>
      <el-tree ref="permTreeRef" :data="permissionTree" show-checkbox node-key="id" :default-checked-keys="currentRole.permissionIds" :props="{ label: 'name', children: 'children' }" />
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
import { formatTime } from '@/utils/format'

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
const rules = { name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }] }
const permDialogVisible = ref(false)
const currentRole = ref({})
const permissionTree = ref([])
const permTreeRef = ref()

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getRoleList({ page: page.value, pageSize: pageSize.value, keyword: keyword.value })
    tableData.value = res.data?.list || res.data || []
    total.value = res.data?.total || 0
  } finally { loading.value = false }
}

const showCreateDialog = () => { isEdit.value = false; form.value = { name: '', description: '' }; dialogVisible.value = true }
const handleEdit = (row) => { isEdit.value = true; form.value = { id: row.id, name: row.name, description: row.description }; dialogVisible.value = true }

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value) { await updateRole(form.value); ElMessage.success('更新成功') }
    else { await createRole(form.value); ElMessage.success('创建成功') }
    dialogVisible.value = false; fetchData()
  } catch (e) { ElMessage.error(e.message || '操作失败') }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除角色 "${row.name}"？`, '提示', { type: 'warning' })
  try { await deleteRole({ id: row.id }); ElMessage.success('删除成功'); fetchData() }
  catch (e) { ElMessage.error(e.message || '删除失败') }
}

const showPermissionDialog = async (row) => {
  currentRole.value = row
  if (!permissionTree.value.length) {
    const res = await getAllPermissions()
    const permissions = res.data || []
    const grouped = {}
    permissions.forEach(p => {
      if (!grouped[p.resource]) grouped[p.resource] = { id: `group-${p.resource}`, name: p.resource, children: [] }
      grouped[p.resource].children.push(p)
    })
    permissionTree.value = Object.values(grouped)
  }
  permDialogVisible.value = true
}

const handleAssignPermissions = async () => {
  const checkedKeys = permTreeRef.value.getCheckedKeys(true)
  try {
    await assignPermissions({ roleId: currentRole.value.id, permissionIds: checkedKeys })
    ElMessage.success('权限分配成功'); permDialogVisible.value = false; fetchData()
  } catch (e) { ElMessage.error(e.message || '权限分配失败') }
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
</style>
