<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NButton, NCard, NDataTable, NInput, NSpace, NTabs, NTabPane } from 'naive-ui'
import {
  getCICDConfig,
  getCICDLogs,
  getCICDPipelines,
  getCICDRuns,
  getCICDTemplates,
} from '@/api/ops'
import type { Pipeline, PipelineLog, PipelineRun, PipelineTemplate } from '@/types/ops'

const loading = ref(false)
const status = ref('')
const keyword = ref('')
const endpoint = ref('')
const pipelines = ref<Pipeline[]>([])
const runs = ref<PipelineRun[]>([])
const templates = ref<PipelineTemplate[]>([])
const logs = ref<PipelineLog[]>([])

const pipelineColumns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '流水线', key: 'name', width: 180 },
  { title: '状态', key: 'status', width: 100 },
  { title: '分支', key: 'branch', width: 120 },
  { title: '最近运行', key: 'lastRunAt', width: 180 },
]

const runColumns = [
  { title: '运行ID', key: 'id', width: 90 },
  { title: '流水线', key: 'pipeline', width: 160 },
  { title: '环境', key: 'environment', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '操作人', key: 'operator', width: 120 },
  { title: '时间', key: 'createdAt', width: 180 },
]

const templateColumns = [
  { title: '模板ID', key: 'id', width: 90 },
  { title: '模板名', key: 'name', width: 180 },
  { title: '来源', key: 'source', width: 140 },
  { title: '描述', key: 'description', ellipsis: { tooltip: true } },
]

const logColumns = [
  { title: '日志ID', key: 'id', width: 90 },
  { title: '阶段', key: 'stage', width: 140 },
  { title: '级别', key: 'level', width: 100 },
  { title: '内容', key: 'message', ellipsis: { tooltip: true } },
  { title: '时间', key: 'createdAt', width: 180 },
]

async function fetchData() {
  loading.value = true
  try {
    const [configRes, pipelineRes, runRes, templateRes, logRes] = await Promise.all([
      getCICDConfig(),
      getCICDPipelines({
        status: status.value.trim() || undefined,
        keyword: keyword.value.trim() || undefined,
      }),
      getCICDRuns({ limit: 50 }),
      getCICDTemplates({ keyword: keyword.value.trim() || undefined }),
      getCICDLogs({ limit: 50 }),
    ])
    endpoint.value = configRes.endpoint
    pipelines.value = pipelineRes.items
    runs.value = runRes.items
    templates.value = templateRes.items
    logs.value = logRes.items
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <n-card :title="`CI/CD流水线（${endpoint || '-'}）`">
    <template #header-extra>
      <n-space>
        <n-input
          v-model:value="status"
          placeholder="状态"
          clearable
          style="width: 120px"
          @keyup.enter="fetchData"
        />
        <n-input
          v-model:value="keyword"
          placeholder="关键词"
          clearable
          style="width: 180px"
          @keyup.enter="fetchData"
        />
        <n-button type="primary" @click="fetchData">查询</n-button>
      </n-space>
    </template>
    <n-tabs type="line" animated>
      <n-tab-pane name="pipelines" tab="流水线">
        <n-data-table
          :columns="pipelineColumns"
          :data="pipelines"
          :loading="loading"
          :bordered="false"
          :row-key="(row: Pipeline) => row.id"
        />
      </n-tab-pane>
      <n-tab-pane name="runs" tab="运行记录">
        <n-data-table
          :columns="runColumns"
          :data="runs"
          :loading="loading"
          :bordered="false"
          :row-key="(row: PipelineRun) => row.id"
        />
      </n-tab-pane>
      <n-tab-pane name="templates" tab="模板">
        <n-data-table
          :columns="templateColumns"
          :data="templates"
          :loading="loading"
          :bordered="false"
          :row-key="(row: PipelineTemplate) => row.id"
        />
      </n-tab-pane>
      <n-tab-pane name="logs" tab="日志">
        <n-data-table
          :columns="logColumns"
          :data="logs"
          :loading="loading"
          :bordered="false"
          :row-key="(row: PipelineLog) => row.id"
        />
      </n-tab-pane>
    </n-tabs>
  </n-card>
</template>
