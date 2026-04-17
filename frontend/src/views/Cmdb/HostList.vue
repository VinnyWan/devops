<template>
  <div class="page-container">
    <div class="page-header">
      <h3>主机管理</h3>
      <div style="display: flex; gap: 8px;">
        <el-button @click="showBatchDialog">批量导入</el-button>
        <el-button type="primary" @click="showCreateDialog">新增主机</el-button>
      </div>
    </div>

    <div class="toolbar">
      <el-input v-model="keyword" placeholder="搜索主机名/IP" style="width: 240px;" clearable @clear="fetchData" @keyup.enter="fetchData">
        <template #append>
          <el-button @click="fetchData"><el-icon><Search /></el-icon></el-button>
        </template>
      </el-input>
      <el-select v-model="groupId" placeholder="分组" clearable style="width: 180px;" @change="fetchData">
        <el-option v-for="item in flatGroups" :key="item.id" :label="item.label" :value="item.id" />
      </el-select>
      <el-select v-model="status" placeholder="状态" clearable style="width: 140px;" @change="fetchData">
        <el-option label="在线" value="online" />
        <el-option label="离线" value="offline" />
        <el-option label="未知" value="unknown" />
      </el-select>
    </div>

    <el-table :data="tableData" stripe v-loading="loading" style="width: 100%">
      <el-table-column prop="hostname" label="主机名" min-width="160" />
      <el-table-column prop="ip" label="IP" width="160" />
      <el-table-column prop="port" label="端口" width="80" />
      <el-table-column prop="osName" label="操作系统" min-width="140" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="最近活跃" width="180">
        <template #default="{ row }">{{ formatTime(row.lastActiveAt) }}</template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleTest(row)">测试连接</el-button>
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-wrap">
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next" @current-change="fetchData" @size-change="fetchData" />
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑主机' : '新增主机'" width="720px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="主机名" prop="hostname"><el-input v-model="form.hostname" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="IP" prop="ip"><el-input v-model="form.ip" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="端口"><el-input-number v-model="form.port" :min="1" :max="65535" style="width:100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="分组"><el-select v-model="form.groupId" clearable style="width:100%"><el-option v-for="item in flatGroups" :key="item.id" :label="item.label" :value="item.id" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="凭据"><el-select v-model="form.credentialId" clearable style="width:100%"><el-option v-for="item in credentials" :key="item.id" :label="item.name" :value="item.id" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="操作系统"><el-input v-model="form.osName" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="CPU"><el-input-number v-model="form.cpuCores" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="内存(MB)"><el-input-number v-model="form.memoryTotal" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="磁盘(GB)"><el-input-number v-model="form.diskTotal" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="标签"><el-input v-model="form.labels" placeholder='例如: {"env":"prod"}' /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchDialogVisible" title="批量导入主机" width="760px">
      <el-alert type="info" :closable="false" style="margin-bottom: 12px;" title='请输入 JSON 数组，例如：[{"hostname":"web-01","ip":"192.168.1.10","port":22}]' />
      <el-input v-model="batchText" type="textarea" :rows="14" placeholder="请输入 JSON 数组" />
      <template #footer>
        <el-button @click="batchDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBatchSubmit">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getHostList, createHost, updateHost, deleteHost, testHost, batchCreateHost } from '@/api/cmdb/host'
import { getGroupTree } from '@/api/cmdb/group'
import { getCredentialList } from '@/api/cmdb/credential'
import { required, ipAddress } from '@/utils/validate'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const groupId = ref('')
const status = ref('')
const dialogVisible = ref(false)
const batchDialogVisible = ref(false)
const isEdit = ref(false)
const batchText = ref('')
const formRef = ref()
const groupTree = ref([])
const credentials = ref([])
const form = ref({ hostname: '', ip: '', port: 22, groupId: '', credentialId: '', osName: '', cpuCores: 0, memoryTotal: 0, diskTotal: 0, labels: '', description: '' })
const rules = {
  hostname: [required('请输入主机名')],
  ip: [required('请输入 IP'), ipAddress()]
}

const flatGroups = computed(() => {
  const result = []
  const walk = (nodes, prefix = '') => {
    nodes.forEach(node => {
      const label = prefix ? `${prefix} / ${node.name}` : node.name
      result.push({ id: node.id, label })
      if (node.children?.length) walk(node.children, label)
    })
  }
  walk(groupTree.value)
  return result
})

const fetchGroups = async () => {
  const res = await getGroupTree()
  groupTree.value = res.data || []
}

const fetchCredentials = async () => {
  const res = await getCredentialList({ page: 1, pageSize: 100 })
  credentials.value = res.data || []
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getHostList({ page: page.value, pageSize: pageSize.value, keyword: keyword.value, groupId: groupId.value || undefined, status: status.value || undefined })
    tableData.value = res.data || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  form.value = { hostname: '', ip: '', port: 22, groupId: '', credentialId: '', osName: '', cpuCores: 0, memoryTotal: 0, diskTotal: 0, labels: '', description: '' }
  dialogVisible.value = true
}

const showBatchDialog = () => {
  batchText.value = ''
  batchDialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { id: row.id, hostname: row.hostname, ip: row.ip, port: row.port, groupId: row.groupId || '', credentialId: row.credentialId || '', osName: row.osName || '', cpuCores: row.cpuCores || 0, memoryTotal: row.memoryTotal || 0, diskTotal: row.diskTotal || 0, labels: row.labels || '', description: row.description || '' }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value) {
      await updateHost(form.value)
      ElMessage.success('更新成功')
    } else {
      await createHost(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  }
}

const handleBatchSubmit = async () => {
  try {
    const data = JSON.parse(batchText.value)
    await batchCreateHost(data)
    ElMessage.success('导入成功')
    batchDialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '导入失败，请检查 JSON 格式')
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除主机 "${row.hostname}"？`, '提示', { type: 'warning' })
  try {
    await deleteHost({ id: row.id })
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '删除失败')
  }
}

const handleTest = async (row) => {
  try {
    const res = await testHost({ id: row.id })
    ElMessage.success(res.message || '连接成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '连接失败')
  }
}

const statusTagType = (val) => ({ online: 'success', offline: 'danger', unknown: 'info' }[val] || 'info')
const statusText = (val) => ({ online: '在线', offline: '离线', unknown: '未知' }[val] || val || '-')

onMounted(async () => {
  await Promise.all([fetchGroups(), fetchCredentials()])
  fetchData()
})
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
