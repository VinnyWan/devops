<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NButton, NCard, NDataTable, NInput, NSpace } from 'naive-ui'
import { getMonitorConfig, queryMonitor } from '@/api/ops'
import type { QuerySeries } from '@/types/ops'

const loading = ref(false)
const metric = ref('up')
const step = ref('1m')
const endpoint = ref('')
const rows = ref<Array<{ key: string; labels: string; points: number }>>([])

const columns = [
  { title: '序号', key: 'key', width: 90 },
  { title: '标签', key: 'labels' },
  { title: '点位数量', key: 'points', width: 120 },
]

function formatSeries(series: QuerySeries[]) {
  rows.value = series.map((item, index) => ({
    key: String(index + 1),
    labels: Object.entries(item.labels)
      .map(([k, v]) => `${k}=${v}`)
      .join(', '),
    points: item.points.length,
  }))
}

async function fetchData() {
  loading.value = true
  try {
    const [configRes, queryRes] = await Promise.all([
      getMonitorConfig(),
      queryMonitor({ metric: metric.value.trim(), step: step.value.trim() || undefined }),
    ])
    endpoint.value = configRes.endpoint
    formatSeries(queryRes.series)
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <n-card :title="`监控配置（Prometheus: ${endpoint || '-'}）`">
    <template #header-extra>
      <n-space>
        <n-input
          v-model:value="metric"
          placeholder="指标名，例如 up"
          clearable
          style="width: 220px"
          @keyup.enter="fetchData"
        />
        <n-input
          v-model:value="step"
          placeholder="步长，例如 1m"
          clearable
          style="width: 140px"
          @keyup.enter="fetchData"
        />
        <n-button type="primary" @click="fetchData">查询</n-button>
      </n-space>
    </template>
    <n-data-table
      :columns="columns"
      :data="rows"
      :loading="loading"
      :bordered="false"
      :row-key="(row) => row.key"
    />
  </n-card>
</template>
