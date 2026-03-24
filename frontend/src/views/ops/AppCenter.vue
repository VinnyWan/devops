<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NButton, NCard, NDataTable, NInput, NSpace, NTabs, NTabPane } from 'naive-ui'
import {
  getAppDeployments,
  getApps,
  getAppTemplates,
  getAppTopology,
  getAppVersions,
} from '@/api/ops'
import type {
  AppTemplate,
  Application,
  ApplicationDeployment,
  ApplicationVersion,
} from '@/types/ops'

const loading = ref(false)
const keyword = ref('')
const environment = ref('')
const apps = ref<Application[]>([])
const templates = ref<AppTemplate[]>([])
const deployments = ref<ApplicationDeployment[]>([])
const versions = ref<ApplicationVersion[]>([])
const topologySummary = ref('未查询')

const appColumns = [
  { title: '应用ID', key: 'id', width: 90 },
  { title: '应用名', key: 'name', width: 180 },
  { title: '命名空间', key: 'namespace', width: 140 },
  { title: '状态', key: 'status', width: 100 },
  { title: '更新时间', key: 'updatedAt', width: 180 },
]

const templateColumns = [
  { title: '模板ID', key: 'id', width: 90 },
  { title: '模板名', key: 'name', width: 180 },
  { title: '类型', key: 'type', width: 120 },
  { title: '描述', key: 'description', ellipsis: { tooltip: true } },
]

const deploymentColumns = [
  { title: '记录ID', key: 'id', width: 90 },
  { title: '应用', key: 'appName', width: 160 },
  { title: '集群', key: 'cluster', width: 120 },
  { title: '环境', key: 'environment', width: 120 },
  { title: '版本', key: 'version', width: 130 },
  { title: '状态', key: 'status', width: 100 },
  { title: '时间', key: 'createdAt', width: 180 },
]

const versionColumns = [
  { title: '版本ID', key: 'id', width: 90 },
  { title: '版本', key: 'version', width: 120 },
  { title: '镜像', key: 'image', ellipsis: { tooltip: true } },
  { title: '环境', key: 'environment', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '时间', key: 'createdAt', width: 180 },
]

async function fetchData() {
  loading.value = true
  try {
    const [appsRes, templateRes, deploymentRes, versionRes] = await Promise.all([
      getApps(),
      getAppTemplates({ keyword: keyword.value.trim() || undefined }),
      getAppDeployments({ environment: environment.value.trim() || undefined, limit: 50 }),
      getAppVersions({ limit: 50 }),
    ])
    apps.value = appsRes
    templates.value = templateRes.items
    deployments.value = deploymentRes.items
    versions.value = versionRes.items
    const firstApp = apps.value[0]
    if (firstApp) {
      const topology = await getAppTopology({
        appId: firstApp.id,
        environment: environment.value.trim() || undefined,
      })
      topologySummary.value = `${topology.appName} 节点 ${topology.nodes.length} / 边 ${topology.edges.length}`
    } else {
      topologySummary.value = '暂无应用'
    }
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <n-card :title="`应用管理（拓扑：${topologySummary}）`">
    <template #header-extra>
      <n-space>
        <n-input
          v-model:value="keyword"
          placeholder="模板关键词"
          clearable
          style="width: 180px"
          @keyup.enter="fetchData"
        />
        <n-input
          v-model:value="environment"
          placeholder="环境过滤"
          clearable
          style="width: 160px"
          @keyup.enter="fetchData"
        />
        <n-button type="primary" @click="fetchData">查询</n-button>
      </n-space>
    </template>
    <n-tabs type="line" animated>
      <n-tab-pane name="apps" tab="应用">
        <n-data-table
          :columns="appColumns"
          :data="apps"
          :loading="loading"
          :bordered="false"
          :row-key="(row: Application) => row.id"
        />
      </n-tab-pane>
      <n-tab-pane name="templates" tab="模板">
        <n-data-table
          :columns="templateColumns"
          :data="templates"
          :loading="loading"
          :bordered="false"
          :row-key="(row: AppTemplate) => row.id"
        />
      </n-tab-pane>
      <n-tab-pane name="deployments" tab="部署记录">
        <n-data-table
          :columns="deploymentColumns"
          :data="deployments"
          :loading="loading"
          :bordered="false"
          :row-key="(row: ApplicationDeployment) => row.id"
        />
      </n-tab-pane>
      <n-tab-pane name="versions" tab="版本">
        <n-data-table
          :columns="versionColumns"
          :data="versions"
          :loading="loading"
          :bordered="false"
          :row-key="(row: ApplicationVersion) => row.id"
        />
      </n-tab-pane>
    </n-tabs>
  </n-card>
</template>
