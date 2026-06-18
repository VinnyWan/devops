<template>
  <div class="page-container">
    <div class="page-header">
      <h3>Prometheus 数据源</h3>
      <el-button type="primary" @click="showCreateDialog">添加数据源</el-button>
    </div>

    <el-table :data="tableData" stripe v-loading="loading">
      <el-table-column prop="name" label="名称" min-width="150" />
      <el-table-column prop="endpoint" label="地址" min-width="250" />
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

    <el-empty v-if="!loading && !tableData.length" description="暂无数据源" />

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑数据源' : '添加数据源'" width="550px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如：生产环境 Prometheus" />
        </el-form-item>
        <el-form-item label="地址" prop="endpoint">
          <el-input v-model="form.endpoint" placeholder="http://prometheus:9090" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="可选" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="可选" show-password />
        </el-form-item>
        <el-form-item label="超时(秒)">
          <el-input-number v-model="form.timeoutSeconds" :min="5" :max="120" />
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
import {
  listPrometheusConfigs,
  savePrometheusConfig,
  updatePrometheusConfig,
  deletePrometheusConfig,
  testPrometheusConnection
} from '@/api/monitor'

const loading = ref(false)
const tableData = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()
const form = reactive({
  id: 0,
  name: '',
  endpoint: '',
  username: '',
  password: '',
  timeoutSeconds: 15
})
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  endpoint: [{ required: true, message: '请输入地址', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await listPrometheusConfigs({ page: 1, pageSize: 100 })
    tableData.value = res.data || []
  } catch {
    ElMessage.error('获取数据失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  Object.assign(form, {
    id: 0,
    name: '',
    endpoint: '',
    username: '',
    password: '',
    timeoutSeconds: 15
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  Object.assign(form, { ...row, password: '' })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const data = { ...form, password: form.password || undefined }
    if (isEdit.value) {
      await updatePrometheusConfig(form.id, data)
    } else {
      await savePrometheusConfig(data)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    dialogVisible.value = false
    fetchData()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    submitting.value = false
  }
}

const handleTest = async (row) => {
  try {
    await testPrometheusConnection({
      endpoint: row.endpoint,
      username: row.username,
      password: ''
    })
    ElMessage.success('连接成功')
  } catch {
    ElMessage.error('连接失败')
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该数据源？', '确认删除', { type: 'warning' })
  try {
    await deletePrometheusConfig(row.id)
    ElMessage.success('已删除')
    fetchData()
  } catch {
    /* cancelled */
  }
}

onMounted(fetchData)
</script>

<style scoped>
.page-container {
  padding: 20px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
}
</style>
