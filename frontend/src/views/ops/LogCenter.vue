<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { NButton, NCard, NDataTable, NInput, NSpace, NTag } from 'naive-ui'
import { searchLogs } from '@/api/ops'
import type { LogEntry } from '@/types/ops'

const loading = ref(false)
const keyword = ref('')
const source = ref('')
const level = ref('')
const data = ref<LogEntry[]>([])
const total = ref(0)

const columns = [
  { title: '时间', key: 'createdAt', width: 180 },
  { title: '集群', key: 'cluster', width: 120 },
  { title: '命名空间', key: 'namespace', width: 140 },
  { title: '来源', key: 'source', width: 120 },
  {
    title: '级别',
    key: 'level',
    width: 100,
    render: (row: LogEntry) =>
      h(NTag, { type: row.level === 'error' ? 'error' : 'info' }, { default: () => row.level }),
  },
  { title: '内容', key: 'message', ellipsis: { tooltip: true } },
]

async function fetchData() {
  loading.value = true
  try {
    const result = await searchLogs({
      keyword: keyword.value.trim() || undefined,
      source: source.value.trim() || undefined,
      level: level.value.trim() || undefined,
      page: 1,
      pageSize: 100,
    })
    data.value = result.items
    total.value = result.total
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <n-card :title="`日志检索（共 ${total} 条）`">
    <template #header-extra>
      <n-space>
        <n-input
          v-model:value="keyword"
          placeholder="关键词"
          clearable
          style="width: 180px"
          @keyup.enter="fetchData"
        />
        <n-input
          v-model:value="source"
          placeholder="来源"
          clearable
          style="width: 120px"
          @keyup.enter="fetchData"
        />
        <n-input
          v-model:value="level"
          placeholder="级别"
          clearable
          style="width: 120px"
          @keyup.enter="fetchData"
        />
        <n-button type="primary" @click="fetchData">查询</n-button>
      </n-space>
    </template>
    <n-data-table
      :columns="columns"
      :data="data"
      :loading="loading"
      :bordered="false"
      :row-key="(row: LogEntry) => row.id"
    />
  </n-card>
</template>
