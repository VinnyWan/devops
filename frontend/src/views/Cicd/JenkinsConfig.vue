<template>
  <div class="page-container">
    <div class="page-header">
      <h3>Jenkins 服务器</h3>
      <el-button type="primary" @click="showCreateDialog">添加服务器</el-button>
    </div>

    <el-table :data="tableData" stripe v-loading="loading">
      <el-table-column prop="name" label="名称" min-width="150" />
      <el-table-column prop="url" label="地址" min-width="250" />
      <el-table-column prop="username" label="用户" width="120" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'connected' ? 'success' : 'danger'">
            {{ row.status === 'connected' ? '已连接' : '异常' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleTest(row)">测试</el-button>
          <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && !tableData.length" description="暂无 Jenkins 服务器" />

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑服务器' : '添加服务器'" width="550px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如：生产 Jenkins" />
        </el-form-item>
        <el-form-item label="URL" prop="url">
          <el-input v-model="form.url" placeholder="http://jenkins.example.com:8080" />
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="admin" />
        </el-form-item>
        <el-form-item label="API Token" prop="apiToken">
          <el-input v-model="form.apiToken" type="password" show-password placeholder="Jenkins API Token" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">保存并测试</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listJenkinsConfigs, saveJenkinsConfig, updateJenkinsConfig, deleteJenkinsConfig, testJenkinsConnection } from '@/api/cicd'

const loading = ref(false)
const tableData = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()
const form = reactive({ id: 0, name: '', url: '', username: '', apiToken: '' })
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  url: [{ required: true, message: '请输入 URL', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  apiToken: [{ required: true, message: '请输入 API Token', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await listJenkinsConfigs({ page: 1, pageSize: 100 })
    tableData.value = res.data || []
  } catch { ElMessage.error('获取数据失败') } finally { loading.value = false }
}

const showCreateDialog = () => {
  isEdit.value = false
  Object.assign(form, { id: 0, name: '', url: '', username: '', apiToken: '' })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  Object.assign(form, { ...row, apiToken: '' })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const data = { ...form, apiToken: form.apiToken || undefined }
    if (isEdit.value) {
      await updateJenkinsConfig(form.id, data)
    } else {
      await saveJenkinsConfig(data)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    dialogVisible.value = false
    fetchData()
  } catch { ElMessage.error('保存失败') } finally { submitting.value = false }
}

const handleTest = async (row) => {
  try {
    await testJenkinsConnection({ url: row.url, username: row.username, apiToken: row.apiToken || '' })
    ElMessage.success('连接成功')
  } catch { ElMessage.error('连接失败') }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该 Jenkins 服务器？', '确认删除', { type: 'warning' })
  try {
    await deleteJenkinsConfig(row.id)
    ElMessage.success('已删除')
    fetchData()
  } catch { /* cancelled */ }
}

onMounted(fetchData)
</script>
