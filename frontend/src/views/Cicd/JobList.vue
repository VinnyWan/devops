<template>
  <div class="page-container">
    <div class="page-header">
      <h3>Jenkins Job 管理</h3>
    </div>

    <div class="toolbar">
      <el-select v-model="configId" placeholder="选择 Jenkins 服务器" style="width: 220px" @change="fetchJobs">
        <el-option v-for="c in configs" :key="c.id" :label="c.name" :value="c.id" />
      </el-select>
      <el-input v-model="keyword" placeholder="搜索 Job 名称" style="width: 200px" clearable @change="fetchJobs" />
      <el-button type="primary" @click="fetchJobs">查询</el-button>
    </div>

    <el-table :data="jobs" stripe v-loading="loading" @row-click="showBuilds">
      <el-table-column prop="name" label="Job 名称" min-width="250" />
      <el-table-column prop="displayName" label="显示名称" min-width="150" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.color === 'blue' ? 'success' : row.color === 'red' ? 'danger' : row.color === 'disabled' ? 'info' : 'warning'">
            {{ row.color === 'blue' ? '正常' : row.color === 'red' ? '失败' : row.color === 'disabled' ? '禁用' : '异常' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button link type="primary" size="small" :disabled="!row.buildable" @click.stop="handleTrigger(row)">构建</el-button>
          <el-button link type="primary" size="small" @click.stop="showBuilds(row)">历史</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && !jobs.length" description="暂无 Job" />

    <!-- Build History Dialog -->
    <el-dialog v-model="buildDialogVisible" :title="`构建历史: ${selectedJob?.name || ''}`" width="800px">
      <el-table :data="builds" stripe v-loading="buildLoading" max-height="400">
        <el-table-column prop="number" label="构建号" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="{ row: b }">
            <el-tag :type="b.result === 'SUCCESS' ? 'success' : b.result === 'FAILURE' ? 'danger' : b.building ? 'warning' : 'info'">
              {{ b.building ? '构建中' : (b.result || 'UNKNOWN') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="120">
          <template #default="{ row: b }">{{ formatDuration(b.duration) }}</template>
        </el-table-column>
        <el-table-column label="时间" width="180">
          <template #default="{ row: b }">{{ formatTime(b.timestamp) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row: b }">
            <el-button link type="primary" size="small" @click="showLog(b)">日志</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- Build Log Dialog -->
    <el-dialog v-model="logDialogVisible" :title="`构建日志 #${selectedBuild?.number || ''}`" width="900px">
      <div style="background: #1e1e1e; color: #d4d4d4; padding: 16px; max-height: 500px; overflow: auto; font-family: monospace; font-size: 13px; white-space: pre-wrap; border-radius: 6px;">{{ logText || '加载中...' }}</div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listJenkinsConfigs, listJobs, triggerBuild, listBuilds, getBuildLog } from '@/api/cicd'

const configs = ref([])
const configId = ref(0)
const keyword = ref('')
const jobs = ref([])
const loading = ref(false)

const buildDialogVisible = ref(false)
const buildLoading = ref(false)
const builds = ref([])
const selectedJob = ref(null)

const logDialogVisible = ref(false)
const logText = ref('')
const selectedBuild = ref(null)

const fetchJobs = async () => {
  if (!configId.value) return
  loading.value = true
  try {
    const res = await listJobs(configId.value, { keyword: keyword.value })
    jobs.value = res.data || []
  } catch { ElMessage.error('获取 Job 列表失败') } finally { loading.value = false }
}

const handleTrigger = async (row) => {
  try {
    await triggerBuild(configId.value, { jobName: row.name })
    ElMessage.success('构建已触发')
  } catch { ElMessage.error('触发构建失败') }
}

const showBuilds = async (row) => {
  selectedJob.value = row
  buildDialogVisible.value = true
  buildLoading.value = true
  try {
    const res = await listBuilds(configId.value, { jobName: row.name })
    builds.value = res.data || []
  } catch { ElMessage.error('获取构建历史失败') } finally { buildLoading.value = false }
}

const showLog = async (build) => {
  selectedBuild.value = build
  logDialogVisible.value = true
  logText.value = ''
  try {
    const res = await getBuildLog(configId.value, { jobName: selectedJob.value.name, buildNumber: build.number })
    logText.value = res.data?.text || '(无内容)'
  } catch { logText.value = '获取日志失败' }
}

const formatDuration = (ms) => {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`
}

const formatTime = (ts) => {
  if (!ts) return '-'
  return new Date(ts).toLocaleString()
}

onMounted(async () => {
  try {
    const res = await listJenkinsConfigs({ page: 1, pageSize: 100 })
    configs.value = res.data || []
    if (configs.value.length) { configId.value = configs.value[0].id; fetchJobs() }
  } catch { /* */ }
})
</script>
