<template>
  <div class="page-container">
    <div class="page-header">
      <h3>终端审计</h3>
    </div>

    <div class="toolbar">
      <el-input
        v-model="keyword"
        placeholder="搜索主机名/IP"
        style="width: 240px;"
        clearable
        @clear="fetchData"
        @keyup.enter="fetchData"
      >
        <template #append>
          <el-button @click="fetchData"><el-icon><Search /></el-icon></el-button>
        </template>
      </el-input>
      <el-input
        v-model="username"
        placeholder="用户名"
        style="width: 180px;"
        clearable
        @clear="fetchData"
        @keyup.enter="fetchData"
      />
      <el-select v-model="status" placeholder="状态" clearable style="width: 140px;" @change="fetchData">
        <el-option label="活跃" value="active" />
        <el-option label="已关闭" value="closed" />
        <el-option label="已中断" value="interrupted" />
      </el-select>
    </div>

    <el-table :data="tableData" stripe v-loading="loading" style="width: 100%">
      <el-table-column prop="hostName" label="主机名" min-width="160" />
      <el-table-column prop="hostIp" label="主机 IP" width="160" />
      <el-table-column prop="username" label="用户名" width="140" />
      <el-table-column prop="clientIp" label="客户端 IP" width="160" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="开始时间" width="180">
        <template #default="{ row }">{{ formatTime(row.startedAt) }}</template>
      </el-table-column>
      <el-table-column label="结束时间" width="180">
        <template #default="{ row }">{{ formatTime(row.finishedAt) }}</template>
      </el-table-column>
      <el-table-column label="时长(秒)" width="100">
        <template #default="{ row }">{{ row.duration ?? 0 }}</template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="handleReplay(row)">回放</el-button>
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
        @size-change="fetchData"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search } from '@element-plus/icons-vue'
import { getTerminalSessionList } from '@/api/cmdb/terminal'
import { formatTime } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const username = ref('')
const status = ref('')

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getTerminalSessionList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value || undefined,
      username: username.value || undefined,
      status: status.value || undefined
    })
    tableData.value = res.data || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

const handleReplay = (row) => {
  router.push(`/cmdb/terminal/replay/${row.id}`)
}

const statusTagType = (val) => ({ active: 'success', closed: 'info', interrupted: 'danger' }[val] || 'info')
const statusText = (val) => ({ active: '活跃', closed: '已关闭', interrupted: '已中断' }[val] || val || '-')

onMounted(() => {
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
