<script setup lang="ts">
import { h, ref, onMounted } from 'vue'
import {
  NCard,
  NGrid,
  NGi,
  NDescriptions,
  NDescriptionsItem,
  NDataTable,
  NTag,
  NButton,
  NSpace,
  NSpin,
  NDivider,
  NInput,
  NPagination,
} from 'naive-ui'
import { useRoute, useRouter } from 'vue-router'
import {
  getClusterDetail,
  getWorkloadStats,
  getNetworkStats,
  getStorageStats,
  getClusterEvents,
} from '@/api/cluster'
import type {
  Cluster,
  WorkloadCounts,
  NetworkCounts,
  StorageCounts,
  EventInfo,
} from '@/types/cluster'
import StatusTag from '@/components/StatusTag.vue'

const route = useRoute()
const router = useRouter()
const clusterId = Number(route.params.id)

const loading = ref(true)
const eventsLoading = ref(false)
const cluster = ref<Cluster | null>(null)
const workload = ref<WorkloadCounts | null>(null)
const network = ref<NetworkCounts | null>(null)
const storage = ref<StorageCounts | null>(null)
const events = ref<EventInfo[]>([])
const eventTotal = ref(0)
const eventPage = ref(1)
const eventPageSize = ref(10)
const eventKeyword = ref('')

const eventColumns = [
  {
    title: '类型',
    key: 'type',
    width: 80,
    render: (row: EventInfo) => {
      const type = row.type === 'Warning' ? 'warning' : 'success'
      return h(NTag, { size: 'small', type }, { default: () => row.type })
    },
  },
  { title: '对象', key: 'object', width: 200 },
  { title: '原因', key: 'reason', width: 120 },
  { title: '信息', key: 'message', ellipsis: { tooltip: true } },
  { title: '时间', key: 'time', width: 180 },
]

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const eventData = await getClusterEvents(clusterId, {
      page: eventPage.value,
      pageSize: eventPageSize.value,
      keyword: eventKeyword.value,
    })
    events.value = eventData?.items || []
    eventTotal.value = eventData?.total || 0
  } finally {
    eventsLoading.value = false
  }
}

async function fetchAll() {
  loading.value = true
  try {
    const [detailRes, workloadRes, networkRes, storageRes] = await Promise.all([
      getClusterDetail(clusterId),
      getWorkloadStats(clusterId),
      getNetworkStats(clusterId),
      getStorageStats(clusterId),
    ])
    cluster.value = detailRes
    workload.value = workloadRes
    network.value = networkRes
    storage.value = storageRes
    await fetchEvents()
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  eventPage.value = page
  fetchEvents()
}

function handleSearch() {
  eventPage.value = 1
  fetchEvents()
}

onMounted(fetchAll)
</script>

<template>
  <n-spin :show="loading">
    <n-space vertical :size="12">
      <!-- 顶部返回 -->
      <n-space align="center">
        <n-button quaternary @click="router.push('/cluster')">← 返回集群列表</n-button>
        <span v-if="cluster" style="font-size: 18px; font-weight: 600">{{ cluster.name }}</span>
      </n-space>

      <!-- 基本信息 -->
      <n-card v-if="cluster" title="基本信息" size="small">
        <n-descriptions :column="4" label-placement="left" size="small" bordered>
          <n-descriptions-item label="集群 ID">{{ cluster.id }}</n-descriptions-item>
          <n-descriptions-item label="名称">{{ cluster.name }}</n-descriptions-item>
          <n-descriptions-item label="状态">
            <StatusTag :status="cluster.status" />
          </n-descriptions-item>
          <n-descriptions-item label="认证方式">{{ cluster.authType }}</n-descriptions-item>
          <n-descriptions-item label="环境">{{ cluster.env }}</n-descriptions-item>
          <n-descriptions-item label="K8s 版本">{{ cluster.k8sVersion }}</n-descriptions-item>
          <n-descriptions-item label="API Server">{{ cluster.url }}</n-descriptions-item>
          <n-descriptions-item label="节点数">{{ cluster.nodeCount }}</n-descriptions-item>
          <n-descriptions-item label="默认集群">{{
            cluster.isDefault ? '是' : '否'
          }}</n-descriptions-item>
          <n-descriptions-item label="标签">{{ cluster.labels || '-' }}</n-descriptions-item>
          <n-descriptions-item label="备注">{{ cluster.remark || '-' }}</n-descriptions-item>
          <n-descriptions-item label="创建时间">{{ cluster.createdAt }}</n-descriptions-item>
        </n-descriptions>
      </n-card>

      <!-- 统计信息 -->
      <n-card v-if="workload && network && storage" size="small" title="资源统计">
        <n-grid :cols="24" :x-gap="24">
          <n-gi :span="10">
            <div class="stat-group">
              <div class="stat-title">工作负载</div>
              <n-data-table
                size="small"
                :bordered="false"
                :single-line="false"
                :columns="[
                  { title: 'Deploy', key: 'deployment', align: 'center' },
                  { title: 'STS', key: 'statefulset', align: 'center' },
                  { title: 'DS', key: 'daemonset', align: 'center' },
                  { title: 'Job', key: 'job', align: 'center' },
                  { title: 'CronJob', key: 'cronjob', align: 'center' },
                ]"
                :data="[workload]"
              />
            </div>
          </n-gi>
          <n-gi :span="1">
            <n-divider vertical style="height: 100%" />
          </n-gi>
          <n-gi :span="6">
            <div class="stat-group">
              <div class="stat-title">网络</div>
              <n-data-table
                size="small"
                :bordered="false"
                :single-line="false"
                :columns="[
                  { title: 'Service', key: 'service', align: 'center' },
                  { title: 'Ingress', key: 'ingress', align: 'center' },
                ]"
                :data="[network]"
              />
            </div>
          </n-gi>
          <n-gi :span="1">
            <n-divider vertical style="height: 100%" />
          </n-gi>
          <n-gi :span="6">
            <div class="stat-group">
              <div class="stat-title">存储</div>
              <n-data-table
                size="small"
                :bordered="false"
                :single-line="false"
                :columns="[
                  { title: 'PV', key: 'pv', align: 'center' },
                  { title: 'PVC', key: 'pvc', align: 'center' },
                ]"
                :data="[storage]"
              />
            </div>
          </n-gi>
        </n-grid>
      </n-card>

      <!-- 事件列表 -->
      <n-card size="small">
        <template #header>
          <n-space justify="space-between" align="center">
            <span>事件列表（{{ eventTotal }}）</span>
            <n-input
              v-model:value="eventKeyword"
              placeholder="搜索事件..."
              size="small"
              style="width: 200px"
              @keyup.enter="handleSearch"
            >
              <template #suffix>🔍</template>
            </n-input>
          </n-space>
        </template>

        <n-data-table
          :loading="eventsLoading"
          :columns="eventColumns"
          :data="events"
          :bordered="false"
          size="small"
          :max-height="400"
          :row-key="
            (row: EventInfo) =>
              `${row.time || ''}-${row.object || ''}-${row.reason || ''}-${row.message || ''}`
          "
        />
        <n-space justify="end" style="margin-top: 12px">
          <n-pagination
            v-model:page="eventPage"
            v-model:page-size="eventPageSize"
            :item-count="eventTotal"
            size="small"
            show-size-picker
            :page-sizes="[10, 20, 50, 100]"
            @update:page="handlePageChange"
            @update:page-size="handleSearch"
          />
        </n-space>
      </n-card>
    </n-space>
  </n-spin>
</template>

<style scoped>
.stat-group {
  height: 100%;
}
.stat-title {
  font-size: 14px;
  font-weight: 500;
  color: #666;
  margin-bottom: 8px;
}
</style>
