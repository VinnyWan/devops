<template>
  <div class="page-container">
    <div class="page-header">
      <h3>登录日志</h3>
    </div>

    <div style="margin-bottom: 16px; display: flex; gap: 12px; flex-wrap: wrap;">
      <el-input v-model="filters.username" placeholder="用户账号" clearable style="width: 200px;" @keyup.enter="handleSearch" @clear="handleSearch" />
      <el-select v-model="filters.status" placeholder="登录状态" clearable style="width: 120px;" @change="handleSearch">
        <el-option label="成功" value="success" />
        <el-option label="失败" value="failed" />
      </el-select>
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
      <el-table-column prop="ip" label="登录IP" width="140" />
      <el-table-column prop="location" label="登录地点" width="150" />
      <el-table-column prop="browser" label="浏览器" width="140" />
      <el-table-column prop="os" label="操作系统" width="140" />
      <el-table-column label="登录状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
            {{ row.status === 'success' ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="message" label="提示消息" min-width="150" show-overflow-tooltip />
      <el-table-column label="访问时间" width="170">
        <template #default="{ row }">{{ formatTime(row.loginAt) }}</template>
      </el-table-column>
    </el-table>

    <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="fetchData"
        @size-change="fetchData"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getLoginLogList } from '@/api/loginLog'
import dayjs from 'dayjs'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const dateRange = ref(null)

const filters = reactive({
  username: '',
  status: ''
})

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (filters.username) params.username = filters.username
    if (filters.status) params.status = filters.status
    if (dateRange.value && dateRange.value.length === 2) {
      params.startAt = dateRange.value[0]
      params.endAt = dateRange.value[1]
    }
    const res = await getLoginLogList(params)
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

const formatTime = (val) => {
  if (!val) return '-'
  return dayjs(val).format('YYYY-MM-DD HH:mm:ss')
}

onMounted(fetchData)
</script>

<style scoped>
.page-container {
  background: #fff;
  border-radius: 4px;
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}
</style>
