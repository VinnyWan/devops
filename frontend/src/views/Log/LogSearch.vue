<template>
  <div class="page-container">
    <div class="page-header">
      <h3>日志中心</h3>
      <el-button @click="sourceDialogVisible = true">日志源管理</el-button>
    </div>

    <!-- Search bar -->
    <div class="toolbar">
      <el-select v-model="searchForm.sourceId" placeholder="日志源" style="width: 180px">
        <el-option v-for="s in sources" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
      <el-input v-model="searchForm.keywordInput" placeholder="关键词（空格分隔）" style="width: 250px" clearable />
      <el-select v-model="searchForm.level" placeholder="日志级别" style="width: 120px" clearable>
        <el-option label="ERROR" value="ERROR" />
        <el-option label="WARN" value="WARN" />
        <el-option label="INFO" value="INFO" />
        <el-option label="DEBUG" value="DEBUG" />
      </el-select>
      <el-input v-model="searchForm.service" placeholder="来源服务" style="width: 160px" clearable />
      <el-date-picker v-model="timeRange" type="datetimerange" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width: 360px" value-format="YYYY-MM-DDTHH:mm:ss.SSS[Z]" />
      <el-button type="primary" @click="doSearch">搜索</el-button>
      <el-button @click="doExport" :disabled="!entries.length">导出 CSV</el-button>
    </div>

    <!-- Results -->
    <el-table :data="entries" stripe v-loading="searching" style="margin-top: 16px" max-height="500">
      <el-table-column label="时间" width="180">
        <template #default="{ row }">{{ row.timestamp }}</template>
      </el-table-column>
      <el-table-column label="级别" width="80">
        <template #default="{ row }">
          <el-tag :type="row.level === 'ERROR' ? 'danger' : row.level === 'WARN' ? 'warning' : row.level === 'INFO' ? 'success' : 'info'" size="small">{{ row.level }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="service" label="服务" width="140" />
      <el-table-column prop="host" label="主机" width="140" />
      <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
    </el-table>

    <el-empty v-if="!searching && !entries.length" description="请设置搜索条件并点击搜索" />

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination v-model:current-page="searchForm.page" v-model:page-size="searchForm.pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next" @size-change="doSearch" @current-change="doSearch" />
    </div>

    <!-- Source Management Dialog -->
    <el-dialog v-model="sourceDialogVisible" title="日志源管理" width="800px">
      <el-table :data="sources" stripe max-height="300">
        <el-table-column prop="name" label="名称" width="150" />
        <el-table-column prop="type" label="类型" width="120" />
        <el-table-column prop="endpoint" label="地址" min-width="200" />
        <el-table-column label="状态" width="100">
          <template #default="{ row: s }">
            <el-tag :type="s.status === 'connected' ? 'success' : 'danger'">{{ s.status === 'connected' ? '已连接' : '异常' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row: s }">
            <el-button link type="primary" size="small" @click="testSource(s)">测试</el-button>
            <el-button link type="primary" size="small" @click="editSource(s)">编辑</el-button>
            <el-button link type="danger" size="small" @click="deleteSource(s)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button type="primary" style="margin-top: 12px" @click="showCreateSource">添加日志源</el-button>
    </el-dialog>

    <!-- Source Edit Dialog -->
    <el-dialog v-model="formDialogVisible" :title="isEditSource ? '编辑日志源' : '添加日志源'" width="500px" append-to-body>
      <el-form ref="sourceFormRef" :model="sourceForm" :rules="sourceRules" label-width="110px">
        <el-form-item label="名称" prop="name"><el-input v-model="sourceForm.name" /></el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="sourceForm.type" style="width: 100%"><el-option label="Elasticsearch" value="elasticsearch" /></el-select>
        </el-form-item>
        <el-form-item label="地址" prop="endpoint"><el-input v-model="sourceForm.endpoint" placeholder="http://es:9200" /></el-form-item>
        <el-form-item label="索引模式" prop="indexPattern"><el-input v-model="sourceForm.indexPattern" placeholder="app-logs-*" /></el-form-item>
        <el-form-item label="用户名"><el-input v-model="sourceForm.username" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="sourceForm.password" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitSource" :loading="submitting">保存并测试</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listLogSources, saveLogSource, updateLogSource, deleteLogSource, testLogSourceConnection, searchLogs, exportLogs } from '@/api/log'

// Sources
const sources = ref([])
const sourceDialogVisible = ref(false)
const formDialogVisible = ref(false)
const isEditSource = ref(false)
const submitting = ref(false)
const sourceFormRef = ref()
const sourceForm = reactive({ id: 0, name: '', type: 'elasticsearch', endpoint: '', indexPattern: 'app-logs-*', username: '', password: '' })
const sourceRules = { name: [{ required: true, message: '必填' }], endpoint: [{ required: true, message: '必填' }] }

// Search
const searching = ref(false)
const entries = ref([])
const total = ref(0)
const timeRange = ref([])
const searchForm = reactive({ sourceId: 0, keywordInput: '', level: '', service: '', page: 1, pageSize: 20 })

const fetchSources = async () => {
  try { const res = await listLogSources({ page: 1, pageSize: 100 }); sources.value = res.data || []; if (sources.value.length && !searchForm.sourceId) searchForm.sourceId = sources.value[0].id } catch { /* */ }
}

const doSearch = async () => {
  if (!searchForm.sourceId) { ElMessage.warning('请先选择日志源'); return }
  searching.value = true
  try {
    const keywords = searchForm.keywordInput ? searchForm.keywordInput.split(/\s+/).filter(Boolean) : []
    const startTime = timeRange.value?.[0] ? new Date(timeRange.value[0]).toISOString() : ''
    const endTime = timeRange.value?.[1] ? new Date(timeRange.value[1]).toISOString() : ''
    const res = await searchLogs({ sourceId: searchForm.sourceId, keywords, level: searchForm.level, service: searchForm.service, startTime, endTime, page: searchForm.page, pageSize: searchForm.pageSize })
    entries.value = res.data?.entries || []
    total.value = res.data?.total || 0
  } catch { ElMessage.error('搜索失败') } finally { searching.value = false }
}

const doExport = async () => {
  try {
    const keywords = searchForm.keywordInput ? searchForm.keywordInput.split(/\s+/).filter(Boolean) : []
    const res = await exportLogs({ sourceId: searchForm.sourceId, keywords, level: searchForm.level, service: searchForm.service })
    const blob = new Blob([res], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a'); a.href = url; a.download = 'logs.csv'; a.click()
    URL.revokeObjectURL(url); ElMessage.success('导出成功')
  } catch { ElMessage.error('导出失败') }
}

// Source actions
const showCreateSource = () => { isEditSource.value = false; Object.assign(sourceForm, { id: 0, name: '', type: 'elasticsearch', endpoint: '', indexPattern: 'app-logs-*', username: '', password: '' }); formDialogVisible.value = true }
const editSource = (s) => { isEditSource.value = true; Object.assign(sourceForm, { ...s, password: '' }); formDialogVisible.value = true }

const submitSource = async () => {
  const valid = await sourceFormRef.value.validate().catch(() => false)
  if (!valid) return; submitting.value = true
  try {
    if (isEditSource.value) { await updateLogSource(sourceForm.id, sourceForm) } else { await saveLogSource(sourceForm) }
    ElMessage.success(isEditSource.value ? '更新成功' : '创建成功'); formDialogVisible.value = false; fetchSources()
  } catch { ElMessage.error('保存失败') } finally { submitting.value = false }
}

const testSource = async (s) => {
  try { await testLogSourceConnection(s.id); ElMessage.success('连接成功') } catch { ElMessage.error('连接失败') }
}

const deleteSource = async (s) => {
  await ElMessageBox.confirm('确定删除？', '确认删除', { type: 'warning' })
  try { await deleteLogSource(s.id); ElMessage.success('已删除'); fetchSources() } catch { /* */ }
}

onMounted(fetchSources)
</script>
