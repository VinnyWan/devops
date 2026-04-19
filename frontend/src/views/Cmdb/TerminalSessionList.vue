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
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      >
        <template #append>
          <el-button @click="handleSearch"><el-icon><Search /></el-icon></el-button>
        </template>
      </el-input>
      <el-input
        v-model="username"
        placeholder="用户名"
        style="width: 180px;"
        clearable
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      />
      <el-select v-model="status" placeholder="状态" clearable style="width: 140px;" @change="handleSearch">
        <el-option label="活跃" value="active" />
        <el-option label="已关闭" value="closed" />
        <el-option label="已中断" value="interrupted" />
        <el-option label="空闲超时" value="idle_timeout" />
        <el-option label="时长超限" value="max_duration" />
      </el-select>
      <el-select v-model="tagFilter" placeholder="标签筛选" clearable size="default" @change="handleSearch" style="width: 150px;">
        <el-option v-for="tag in availableTags" :key="tag" :label="tag" :value="tag" />
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
      <el-table-column prop="hostName" label="主机名" min-width="160" />
      <el-table-column prop="hostIp" label="主机 IP" width="160" />
      <el-table-column prop="username" label="用户名" width="140" />
      <el-table-column prop="clientIp" label="客户端 IP" width="160" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="标签" width="200">
        <template #default="{ row }">
          <el-tag
            v-for="tag in (row.tags ? row.tags.split(',') : [])"
            :key="tag"
            size="small"
            closable
            @close="handleRemoveTag(row.id, tag)"
            style="margin-right: 2px;"
          >{{ tag }}</el-tag>
          <el-button size="small" link type="primary" @click="openTagDialog(row)">+标签</el-button>
        </template>
      </el-table-column>
      <el-table-column label="开始时间" width="180">
        <template #default="{ row }">{{ formatTime(row.startedAt) }}</template>
      </el-table-column>
      <el-table-column label="结束时间" width="180">
        <template #default="{ row }">{{ formatTime(row.finishedAt) }}</template>
      </el-table-column>
      <el-table-column label="持续时长" width="120">
        <template #default="{ row }">{{ formatDuration(row.duration) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="handleReplay(row)" :disabled="row.status === 'active'">回放</el-button>
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

    <el-dialog v-model="tagDialog" title="添加标签" width="350px">
      <div style="margin-bottom: 12px;">
        <span style="font-size: 12px; color: #909399;">常用标签:</span>
        <el-tag
          v-for="tag in availableTags.slice(0, 10)"
          :key="tag"
          size="small"
          style="margin: 2px; cursor: pointer;"
          @click="newTag = tag"
        >{{ tag }}</el-tag>
      </div>
      <el-input v-model="newTag" placeholder="输入标签名" @keyup.enter="handleAddTag">
        <template #append>
          <el-button @click="handleAddTag">添加</el-button>
        </template>
      </el-input>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search } from '@element-plus/icons-vue'
import { getTerminalSessionList, addSessionTag, removeSessionTag, getAvailableTags, searchSessionsByTag } from '@/api/cmdb/terminal'
import { formatTime } from '@/utils/format'
import { ElMessage } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keyword = ref('')
const username = ref('')
const status = ref('')
const dateRange = ref(null)
const tagFilter = ref('')
const availableTags = ref([])
const tagDialog = ref(false)
const tagSession = ref(null)
const newTag = ref('')

const formatDuration = (duration) => {
  const totalSeconds = Number(duration || 0)
  if (totalSeconds <= 0) return '0 秒'
  if (totalSeconds < 60) return `${totalSeconds} 秒`
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return seconds ? `${minutes} 分 ${seconds} 秒` : `${minutes} 分`
}

const fetchAvailableTags = async () => {
  try {
    const res = await getAvailableTags()
    availableTags.value = res.data || []
  } catch (e) { /* ignore */ }
}

const fetchData = async () => {
  loading.value = true
  try {
    let res
    if (tagFilter.value) {
      res = await searchSessionsByTag({
        tag: tagFilter.value,
        page: page.value,
        pageSize: pageSize.value
      })
    } else {
      res = await getTerminalSessionList({
        page: page.value,
        pageSize: pageSize.value,
        keyword: keyword.value || undefined,
        username: username.value || undefined,
        status: status.value || undefined,
        startAt: dateRange.value?.[0] || undefined,
        endAt: dateRange.value?.[1] || undefined
      })
    }
    tableData.value = res.data || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  fetchData()
}

const handleReplay = (row) => {
  router.push(`/cmdb/terminal/replay/${row.id}`)
}

const openTagDialog = (row) => {
  tagSession.value = row
  newTag.value = ''
  tagDialog.value = true
}

const handleAddTag = async () => {
  if (!newTag.value || !tagSession.value) return
  try {
    await addSessionTag({ sessionId: tagSession.value.id, tag: newTag.value })
    ElMessage.success('标签已添加')
    newTag.value = ''
    fetchData()
    fetchAvailableTags()
  } catch (e) {
    ElMessage.error('添加失败')
  }
}

const handleRemoveTag = async (sessionId, tag) => {
  try {
    await removeSessionTag({ sessionId, tag })
    fetchData()
  } catch (e) { /* ignore */ }
}

const statusTagType = (val) => ({ active: 'success', closed: 'info', interrupted: 'danger', idle_timeout: 'warning', max_duration: 'warning' }[val] || 'info')
const statusText = (val) => ({ active: '活跃', closed: '已关闭', interrupted: '已中断', idle_timeout: '空闲超时', max_duration: '时长超限' }[val] || val || '-')

onMounted(() => {
  fetchData()
  fetchAvailableTags()
})
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
