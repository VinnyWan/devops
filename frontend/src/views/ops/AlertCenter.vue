<script setup lang="ts">
import { h, onMounted, ref, computed } from 'vue'
import { NButton, NCard, NDataTable, NInput, NSpace, NTag, useMessage } from 'naive-ui'
import {
  getAlertChannels,
  getAlertHistory,
  getAlertRules,
  getAlertSilences,
  toggleAlertRule,
} from '@/api/ops'
import type { AlertRule } from '@/types/ops'
import { useAuth } from '@/composables/useAuth'

const message = useMessage()
const loading = ref(false)
const keyword = ref('')
const data = ref<AlertRule[]>([])
const historyTotal = ref(0)
const silenceTotal = ref(0)
const channelTotal = ref(0)
const { hasPermission } = useAuth()
const canToggleAlertRule = computed(() => hasPermission('alert:update'))

const columns = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '规则名', key: 'name', width: 180 },
  { title: '级别', key: 'severity', width: 90 },
  { title: '集群', key: 'cluster', width: 120 },
  { title: '表达式', key: 'expr', ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'enabled',
    width: 100,
    render: (row: AlertRule) =>
      h(
        NTag,
        { type: row.enabled ? 'success' : 'warning' },
        { default: () => (row.enabled ? '启用' : '禁用') },
      ),
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render: (row: AlertRule) =>
      canToggleAlertRule.value
        ? h(
            NButton,
            { size: 'small', onClick: () => handleToggle(row) },
            { default: () => (row.enabled ? '停用' : '启用') },
          )
        : '-',
  },
]

async function fetchData() {
  loading.value = true
  try {
    const [rulesRes, historyRes, silenceRes, channelRes] = await Promise.all([
      getAlertRules({ keyword: keyword.value.trim() || undefined }),
      getAlertHistory(),
      getAlertSilences(),
      getAlertChannels(),
    ])
    data.value = rulesRes.items
    historyTotal.value = historyRes.total
    silenceTotal.value = silenceRes.total
    channelTotal.value = channelRes.total
  } finally {
    loading.value = false
  }
}

async function handleToggle(row: AlertRule) {
  try {
    await toggleAlertRule({ id: row.id, enabled: !row.enabled })
    message.success('状态已更新')
    await fetchData()
  } catch (error: unknown) {
    message.error((error as Error).message || '更新失败')
  }
}

onMounted(fetchData)
</script>

<template>
  <n-card title="告警中心">
    <template #header-extra>
      <n-space align="center">
        <n-tag type="info">历史 {{ historyTotal }}</n-tag>
        <n-tag type="warning">静默 {{ silenceTotal }}</n-tag>
        <n-tag type="success">渠道 {{ channelTotal }}</n-tag>
        <n-input
          v-model:value="keyword"
          placeholder="搜索规则关键词"
          clearable
          style="width: 220px"
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
      :row-key="(row: AlertRule) => row.id"
    />
  </n-card>
</template>
