<template>
  <div class="page-container">
    <div class="page-header">
      <h3>云账号管理</h3>
      <el-button type="primary" @click="showCreateDialog">添加云账号</el-button>
    </div>
    <div class="toolbar">
      <el-select v-model="filterStatus" placeholder="全部状态" clearable style="width: 150px" @change="fetchData">
        <el-option label="正常" value="active" />
        <el-option label="错误" value="error" />
      </el-select>
    </div>
    <el-table :data="tableData" stripe v-loading="loading">
      <el-table-column prop="name" label="账号名称" min-width="150" />
      <el-table-column prop="provider" label="云厂商" width="100">
        <template #default="{ row }">{{ providerLabel(row.provider) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
            {{ row.status === 'active' ? '正常' : '错误' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="最后同步" width="180">
        <template #default="{ row }">{{ formatTime(row.lastSyncAt) }}</template>
      </el-table-column>
      <el-table-column label="同步间隔" width="100">
        <template #default="{ row }">{{ row.syncInterval }}分钟</template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleSync(row)" :loading="syncingId === row.id">同步</el-button>
          <el-button link type="primary" size="small" @click="showResources(row)">资源</el-button>
          <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
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
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </div>

    <!-- Create/Edit dialog -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑云账号' : '添加云账号'" width="500px" destroy-on-close>
      <el-form :model="form" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="账号名称" prop="name">
          <el-input v-model="form.name" placeholder="输入账号名称" />
        </el-form-item>
        <el-form-item label="SecretId" prop="secretId">
          <el-input v-model="form.secretId" :placeholder="isEdit ? '不修改则留空' : '输入 SecretId'" />
        </el-form-item>
        <el-form-item label="SecretKey" prop="secretKey">
          <el-input v-model="form.secretKey" type="password" show-password :placeholder="isEdit ? '不修改则留空' : '输入 SecretKey'" />
        </el-form-item>
        <el-form-item label="同步间隔" prop="syncInterval">
          <el-input-number v-model="form.syncInterval" :min="5" :max="1440" />
          <span style="margin-left: 8px; color: #909399">分钟</span>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- Resources dialog -->
    <el-dialog v-model="resourceDialogVisible" :title="`云资源 - ${resourceAccountName}`" width="80%" top="5vh" destroy-on-close>
      <div class="resource-toolbar">
        <span class="resource-summary">
          当前显示 {{ resourceData.length }} / 共 {{ resourceTotal }} 条
          <template v-if="resourceTotal > resourcePageSize">
            （第 {{ resourcePage }} 页）
          </template>
        </span>
      </div>
      <el-tabs v-model="resourceType" @tab-change="handleResourceTypeChange">
        <el-tab-pane label="CVM" name="cvm" />
        <el-tab-pane label="VPC" name="vpc" />
        <el-tab-pane label="子网" name="subnet" />
        <el-tab-pane label="安全组" name="security_group" />
        <el-tab-pane label="云硬盘" name="cbs" />
      </el-tabs>
      <el-table :data="resourceData" stripe v-loading="resourceLoading" max-height="400">
        <el-table-column prop="resourceId" label="资源 ID" min-width="180" show-overflow-tooltip />
        <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="region" label="地域" width="120" />
        <el-table-column prop="zone" label="可用区" width="120" />
        <el-table-column prop="state" label="状态" width="100" />
        <el-table-column label="规格" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ formatSpec(row.spec) }}</template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrap resource-pagination">
        <el-pagination
          v-model:current-page="resourcePage"
          v-model:page-size="resourcePageSize"
          :total="resourceTotal"
          :page-sizes="[50, 100, 200]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchResources"
          @current-change="fetchResources"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCloudAccountList, createCloudAccount, updateCloudAccount, deleteCloudAccount, syncCloudAccount, getCloudResources } from '@/api/cmdb/cloud'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const filterStatus = ref('')
const syncingId = ref(0)

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()

const form = reactive({
  id: 0,
  name: '',
  secretId: '',
  secretKey: '',
  syncInterval: 60,
  description: ''
})

const formRules = {
  name: [{ required: true, message: '请输入账号名称', trigger: 'blur' }]
}

const resourceDialogVisible = ref(false)
const resourceAccountName = ref('')
const resourceAccountId = ref(0)
const resourceType = ref('cvm')
const resourceData = ref([])
const resourceLoading = ref(false)
const resourcePage = ref(1)
const resourcePageSize = ref(100)
const resourceTotal = ref(0)

const providerLabel = (p) => {
  const map = { tencent: '腾讯云', aliyun: '阿里云', aws: 'AWS' }
  return map[p] || p
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = { page: page.value, pageSize: pageSize.value }
    if (filterStatus.value) params.status = filterStatus.value
    const res = await getCloudAccountList(params)
    tableData.value = res.data || []
    total.value = res.total || 0
  } catch (e) {
    ElMessage.error('获取云账号列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  form.id = 0
  form.name = ''
  form.secretId = ''
  form.secretKey = ''
  form.syncInterval = 60
  form.description = ''
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.id = row.id
  form.name = row.name
  form.secretId = ''
  form.secretKey = ''
  form.syncInterval = row.syncInterval
  form.description = row.description
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try { await formRef.value.validate() } catch { return }
  submitting.value = true
  try {
    if (isEdit.value) {
      const payload = { id: form.id, name: form.name, syncInterval: form.syncInterval, description: form.description }
      if (form.secretId) payload.secretId = form.secretId
      if (form.secretKey) payload.secretKey = form.secretKey
      await updateCloudAccount(payload)
      ElMessage.success('更新成功')
    } else {
      await createCloudAccount({
        name: form.name,
        secretId: form.secretId,
        secretKey: form.secretKey,
        syncInterval: form.syncInterval,
        description: form.description
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  try { await ElMessageBox.confirm('确定删除该云账号？关联的云资源也会被删除。', '确认', { type: 'warning' }) } catch { return }
  try {
    await deleteCloudAccount({ id: row.id })
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '删除失败')
  }
}

const handleSync = async (row) => {
  syncingId.value = row.id
  try {
    await syncCloudAccount({ id: row.id })
    ElMessage.success('同步成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '同步失败')
  } finally {
    syncingId.value = 0
  }
}

const showResources = (row) => {
  resourceAccountId.value = row.id
  resourceAccountName.value = row.name
  resourceType.value = 'cvm'
  resourcePage.value = 1
  resourcePageSize.value = 100
  resourceDialogVisible.value = true
  fetchResources()
}

const handleResourceTypeChange = () => {
  resourcePage.value = 1
  fetchResources()
}

const fetchResources = async () => {
  resourceLoading.value = true
  try {
    const res = await getCloudResources({
      cloudAccountId: resourceAccountId.value,
      resourceType: resourceType.value,
      page: resourcePage.value,
      pageSize: resourcePageSize.value
    })
    resourceData.value = res.data?.list || res.data || []
    resourceTotal.value = res.data?.total || res.total || resourceData.value.length
  } catch {
    resourceData.value = []
    resourceTotal.value = 0
  } finally {
    resourceLoading.value = false
  }
}

const formatSpec = (spec) => {
  if (!spec) return '-'
  try {
    const obj = typeof spec === 'string' ? JSON.parse(spec) : spec
    return Object.entries(obj).map(([k, v]) => `${k}: ${v}`).join(', ')
  } catch {
    return spec
  }
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.resource-toolbar { display: flex; justify-content: flex-end; margin-bottom: 12px; }
.resource-summary { font-size: 12px; color: #909399; }
.resource-pagination { margin-top: 12px; }
</style>
