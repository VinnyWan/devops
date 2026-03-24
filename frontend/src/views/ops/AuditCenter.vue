<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NInput,
  NPagination,
  NPopconfirm,
  NSpace,
  useMessage,
} from 'naive-ui'
import { cleanupExpiredAuditLogs, exportAuditLogs, getAuditLogs } from '@/api/ops'
import type { AuditLogRecord } from '@/types/ops'

const message = useMessage()
const loading = ref(false)
const exportLoading = ref(false)
const cleanupLoading = ref(false)
const userId = ref<string>('')
const username = ref('')
const operation = ref('')
const resource = ref('')
const range = ref<[number, number] | null>(null)
const page = ref(1)
const pageSize = ref(20)
const data = ref<AuditLogRecord[]>([])
const total = ref(0)

const columns = [
  { title: '用户ID', key: 'userId', width: 90 },
  { title: '时间', key: 'createdAt', width: 180 },
  { title: '用户', key: 'username', width: 120 },
  { title: '操作', key: 'operation', width: 180 },
  { title: '方法', key: 'method', width: 90 },
  { title: '资源', key: 'path', width: 220, ellipsis: { tooltip: true } },
  { title: '状态码', key: 'status', width: 90 },
  { title: '耗时(ms)', key: 'latency', width: 100 },
  { title: 'IP', key: 'ip', width: 130 },
]

const cardTitle = computed(() => `审计日志（共 ${total.value} 条）`)

function buildParams() {
  const parsedUserId = Number.parseInt(userId.value.trim(), 10)
  return {
    userId: Number.isFinite(parsedUserId) ? parsedUserId : undefined,
    username: username.value.trim() || undefined,
    operation: operation.value.trim() || undefined,
    resource: resource.value.trim() || undefined,
    startAt: range.value ? new Date(range.value[0]).toISOString() : undefined,
    endAt: range.value ? new Date(range.value[1]).toISOString() : undefined,
    page: page.value,
    pageSize: pageSize.value,
  }
}

async function fetchData(resetPage = false) {
  if (resetPage) {
    page.value = 1
  }
  loading.value = true
  try {
    const result = await getAuditLogs(buildParams())
    data.value = result.list
    total.value = result.total
    page.value = result.page
    pageSize.value = result.pageSize
  } catch (error) {
    message.error(error instanceof Error ? error.message : '查询失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  void fetchData(true)
}

function handleReset() {
  userId.value = ''
  username.value = ''
  operation.value = ''
  resource.value = ''
  range.value = null
  void fetchData(true)
}

function handlePageChange(nextPage: number) {
  page.value = nextPage
  void fetchData()
}

function handlePageSizeChange(nextSize: number) {
  pageSize.value = nextSize
  page.value = 1
  void fetchData()
}

async function handleExport() {
  exportLoading.value = true
  try {
    const params = buildParams()
    const response = await exportAuditLogs({
      userId: params.userId,
      username: params.username,
      operation: params.operation,
      resource: params.resource,
      startAt: params.startAt,
      endAt: params.endAt,
      limit: 10000,
    })
    const contentType = response.headers['content-type']
    if (typeof contentType === 'string' && contentType.includes('application/json')) {
      const text = await response.data.text()
      const payload = JSON.parse(text) as { message?: string }
      throw new Error(payload.message || '导出失败')
    }
    const blob = new Blob([response.data], { type: 'text/csv;charset=utf-8' })
    const downloadUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    const stamp = new Date().toISOString().replace(/[-:]/g, '').slice(0, 15)
    link.download = `audit_logs_${stamp}.csv`
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(downloadUrl)
    message.success('导出成功')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '导出失败')
  } finally {
    exportLoading.value = false
  }
}

async function handleCleanup() {
  cleanupLoading.value = true
  try {
    const result = await cleanupExpiredAuditLogs()
    message.success(`清理完成，删除 ${result.cleaned} 条过期日志`)
    await fetchData()
  } catch (error) {
    message.error(error instanceof Error ? error.message : '清理失败')
  } finally {
    cleanupLoading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <n-card :title="cardTitle">
    <template #header-extra>
      <n-space>
        <n-input
          v-model:value="userId"
          placeholder="用户ID"
          clearable
          style="width: 120px"
          @keyup.enter="handleSearch"
        />
        <n-input
          v-model:value="username"
          placeholder="用户名"
          clearable
          style="width: 140px"
          @keyup.enter="handleSearch"
        />
        <n-input
          v-model:value="operation"
          placeholder="操作类型"
          clearable
          style="width: 160px"
          @keyup.enter="handleSearch"
        />
        <n-input
          v-model:value="resource"
          placeholder="资源路径"
          clearable
          style="width: 180px"
          @keyup.enter="handleSearch"
        />
        <n-date-picker v-model:value="range" type="datetimerange" clearable style="width: 300px" />
        <n-button type="primary" @click="handleSearch">查询</n-button>
        <n-button @click="handleReset">重置</n-button>
        <n-button :loading="exportLoading" @click="handleExport">CSV导出</n-button>
        <n-popconfirm @positive-click="handleCleanup">
          <template #trigger>
            <n-button type="warning" :loading="cleanupLoading">过期清理</n-button>
          </template>
          确认立即清理过期审计日志吗？
        </n-popconfirm>
      </n-space>
    </template>
    <n-data-table
      :columns="columns"
      :data="data"
      :loading="loading"
      :bordered="false"
      :row-key="(row: AuditLogRecord) => row.id"
    />
    <n-space justify="end" style="margin-top: 12px">
      <n-pagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :item-count="total"
        show-size-picker
        :page-sizes="[10, 20, 50, 100]"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </n-space>
  </n-card>
</template>
