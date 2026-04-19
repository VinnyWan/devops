<template>
  <div class="page-container">
    <div class="page-header">
      <h3>操作日志</h3>
    </div>

    <div class="toolbar">
      <el-input v-model="filters.username" placeholder="用户账号" clearable style="width: 160px;" @keyup.enter="handleSearch" @clear="handleSearch" />
      <el-select v-model="filters.method" placeholder="请求方式" clearable style="width: 120px;" @change="handleSearch">
        <el-option label="GET" value="GET" />
        <el-option label="POST" value="POST" />
        <el-option label="PUT" value="PUT" />
        <el-option label="DELETE" value="DELETE" />
      </el-select>
      <el-input v-model="filters.operation" placeholder="操作描述" clearable style="width: 200px;" @keyup.enter="handleSearch" @clear="handleSearch" />
      <el-date-picker
        v-model="dateRange"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        value-format="YYYY-MM-DDTHH:mm:ssZ"
        style="width: 360px;"
        @change="handleSearch"
      />
      <el-button type="primary" @click="handleSearch">搜索</el-button>
    </div>

    <el-table :data="tableData" stripe v-loading="loading" style="width: 100%">
      <el-table-column prop="username" label="用户账号" width="120" />
      <el-table-column label="请求方式" width="100">
        <template #default="{ row }">
          <el-tag :type="methodTagType(row.method)" size="small">{{ row.method }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="ip" label="登录IP" width="140" />
      <el-table-column prop="path" label="请求URL" min-width="200" show-overflow-tooltip />
      <el-table-column prop="operation" label="操作描述" min-width="150" show-overflow-tooltip />
      <el-table-column label="操作时间" width="170">
        <template #default="{ row }">{{ formatTime(row.requestAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="showDetail(row)">详情</el-button>
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
        @current-change="fetchData"
        @size-change="handlePageSizeChange"
      />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="操作详情" width="600px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="用户账号">{{ detail.username }}</el-descriptions-item>
        <el-descriptions-item label="请求方式">{{ detail.method }}</el-descriptions-item>
        <el-descriptions-item label="请求IP">{{ detail.ip }}</el-descriptions-item>
        <el-descriptions-item label="HTTP状态">{{ detail.status }}</el-descriptions-item>
        <el-descriptions-item label="响应耗时">{{ detail.latency }} ms</el-descriptions-item>
        <el-descriptions-item label="操作时间">{{ formatTime(detail.requestAt) }}</el-descriptions-item>
        <el-descriptions-item label="请求路径" :span="2">{{ detail.path }}</el-descriptions-item>
      </el-descriptions>
      <div v-if="detail.params" style="margin-top: 16px;">
        <div class="detail-label">请求参数</div>
        <el-input type="textarea" :model-value="formatJSON(detail.params)" :rows="4" readonly />
      </div>
      <div v-if="detail.result" style="margin-top: 16px;">
        <div class="detail-label">返回结果</div>
        <el-input type="textarea" :model-value="formatJSON(detail.result)" :rows="4" readonly />
      </div>
      <div v-if="detail.errorMessage" style="margin-top: 16px;">
        <div class="detail-label error-label">错误信息</div>
        <el-input type="textarea" :model-value="detail.errorMessage" :rows="2" readonly />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getAuditList } from '@/api/audit'
import dayjs from 'dayjs'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const dateRange = ref(null)

const filters = reactive({
  username: '',
  method: '',
  operation: ''
})

const detailVisible = ref(false)
const detail = ref({})

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (filters.username) params.username = filters.username
    if (filters.operation) params.operation = filters.operation
    if (filters.method) params.method = filters.method
    if (dateRange.value && dateRange.value.length === 2) {
      params.startAt = dateRange.value[0]
      params.endAt = dateRange.value[1]
    }
    const res = await getAuditList(params)
    tableData.value = res.data?.list || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  fetchData()
}

const handlePageSizeChange = () => {
  page.value = 1
  fetchData()
}

const showDetail = (row) => {
  detail.value = row
  detailVisible.value = true
}

const methodTagType = (method) => {
  const map = { GET: 'success', POST: 'warning', DELETE: 'danger', PUT: '', PATCH: 'info' }
  return map[method] || 'info'
}

const formatTime = (val) => {
  if (!val) return '-'
  return dayjs(val).format('YYYY-MM-DD HH:mm:ss')
}

const formatJSON = (str) => {
  if (!str) return ''
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

onMounted(fetchData)
</script>

<style scoped>
.detail-label {
  font-weight: 500;
  margin-bottom: var(--spacing-sm);
  color: var(--color-text);
}
.error-label {
  color: var(--color-danger);
}
</style>
