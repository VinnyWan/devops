<template>
  <div class="page-container">
    <div class="page-header">
      <h3>凭据管理</h3>
      <el-button type="primary" @click="showCreateDialog">新增凭据</el-button>
    </div>

    <div style="margin-bottom: 16px;">
      <el-input v-model="keyword" placeholder="搜索凭据名称/用户名" style="width: 300px;" clearable @clear="fetchData" @keyup.enter="fetchData">
        <template #append>
          <el-button @click="fetchData"><el-icon><Search /></el-icon></el-button>
        </template>
      </el-input>
    </div>

    <el-table :data="tableData" stripe v-loading="loading" style="width: 100%">
      <el-table-column prop="name" label="凭据名称" min-width="180" />
      <el-table-column prop="type" label="类型" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.type === 'password' ? '密码' : '密钥' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="username" label="用户名" min-width="140" />
      <el-table-column prop="description" label="描述" min-width="180" />
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-wrap">
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next" @current-change="fetchData" @size-change="fetchData" />
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑凭据' : '新增凭据'" width="640px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="90px">
        <el-form-item label="凭据名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio label="password">密码</el-radio>
            <el-radio label="key">密钥</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="用户名" prop="username"><el-input v-model="form.username" /></el-form-item>
        <el-form-item v-if="form.type === 'password'" label="密码" prop="password"><el-input v-model="form.password" type="password" show-password /></el-form-item>
        <template v-else>
          <el-form-item label="私钥" prop="privateKey"><el-input v-model="form.privateKey" type="textarea" :rows="8" /></el-form-item>
          <el-form-item label="密钥密码"><el-input v-model="form.passphrase" type="password" show-password /></el-form-item>
        </template>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getCredentialList, createCredential, updateCredential, deleteCredential } from '@/api/cmdb/credential'
import { required } from '@/utils/validate'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref()
const form = ref({ name: '', type: 'password', username: '', password: '', privateKey: '', passphrase: '', description: '' })
const rules = {
  name: [required('请输入凭据名称')],
  type: [required('请选择凭据类型')],
  username: [required('请输入用户名')]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getCredentialList({ page: page.value, pageSize: pageSize.value, keyword: keyword.value })
    tableData.value = res.data || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  form.value = { name: '', type: 'password', username: '', password: '', privateKey: '', passphrase: '', description: '' }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { id: row.id, name: row.name, type: row.type, username: row.username, password: '', privateKey: '', passphrase: '', description: row.description || '' }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    if (form.value.type === 'password' && !isEdit.value && !form.value.password) {
      ElMessage.error('请输入密码')
      return
    }
    if (form.value.type === 'key' && !isEdit.value && !form.value.privateKey) {
      ElMessage.error('请输入私钥')
      return
    }

    if (isEdit.value) {
      await updateCredential(form.value)
      ElMessage.success('更新成功')
    } else {
      await createCredential(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除凭据 "${row.name}"？`, '提示', { type: 'warning' })
  try {
    await deleteCredential({ id: row.id })
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '删除失败')
  }
}

watch(() => form.value.type, (val) => {
  if (val === 'password') {
    form.value.privateKey = ''
    form.value.passphrase = ''
  } else {
    form.value.password = ''
  }
})

onMounted(fetchData)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
