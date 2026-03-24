<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NButton, NCard, NDataTable, NInput, NSpace, NTabs, NTabPane } from 'naive-ui'
import { getHarborConfig, getHarborImages, getHarborProjects } from '@/api/ops'
import type { HarborProject, RepositoryImage } from '@/types/ops'

const loading = ref(false)
const keyword = ref('')
const projectName = ref('')
const harborEndpoint = ref('')
const projects = ref<HarborProject[]>([])
const images = ref<RepositoryImage[]>([])

const projectColumns = [
  { title: '项目ID', key: 'id', width: 100 },
  { title: '项目名', key: 'name' },
  {
    title: '公开',
    key: 'public',
    width: 100,
    render: (row: HarborProject) => (row.public ? '是' : '否'),
  },
  { title: '更新时间', key: 'updatedAt', width: 180 },
]

const imageColumns = [
  { title: '项目', key: 'projectName', width: 160 },
  { title: '仓库', key: 'repository', width: 220 },
  { title: '标签', key: 'tag', width: 120 },
  { title: 'Digest', key: 'digest', ellipsis: { tooltip: true } },
  { title: '大小(Byte)', key: 'size', width: 130 },
  { title: '推送时间', key: 'pushedAt', width: 180 },
]

async function fetchData() {
  loading.value = true
  try {
    const [configRes, projectsRes, imagesRes] = await Promise.all([
      getHarborConfig(),
      getHarborProjects({ keyword: keyword.value.trim() || undefined }),
      getHarborImages({ projectName: projectName.value.trim() || undefined }),
    ])
    harborEndpoint.value = configRes.endpoint
    projects.value = projectsRes.items
    images.value = imagesRes.items
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <n-card :title="`Harbor管理（${harborEndpoint || '-'}）`">
    <template #header-extra>
      <n-space>
        <n-input
          v-model:value="keyword"
          placeholder="项目关键词"
          clearable
          style="width: 160px"
          @keyup.enter="fetchData"
        />
        <n-input
          v-model:value="projectName"
          placeholder="项目名过滤镜像"
          clearable
          style="width: 160px"
          @keyup.enter="fetchData"
        />
        <n-button type="primary" @click="fetchData">查询</n-button>
      </n-space>
    </template>
    <n-tabs type="line" animated>
      <n-tab-pane name="project" tab="项目">
        <n-data-table
          :columns="projectColumns"
          :data="projects"
          :loading="loading"
          :bordered="false"
          :row-key="(row: HarborProject) => row.id"
        />
      </n-tab-pane>
      <n-tab-pane name="image" tab="镜像">
        <n-data-table
          :columns="imageColumns"
          :data="images"
          :loading="loading"
          :bordered="false"
          :row-key="(row: RepositoryImage) => row.id"
        />
      </n-tab-pane>
    </n-tabs>
  </n-card>
</template>
