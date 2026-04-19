<template>
  <div class="page-container">
    <!-- Page Header -->
    <div class="page-header">
      <h3>集群详情</h3>
      <el-button @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon>返回
      </el-button>
    </div>

    <!-- Cluster Basic Info -->
    <el-descriptions :column="3" border size="default" style="margin-bottom: 16px">
      <el-descriptions-item label="集群名称">{{ clusterInfo.name }}</el-descriptions-item>
      <el-descriptions-item label="K8s 版本">{{ clusterInfo.k8sVersion }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <span :style="{ color: clusterInfo.status === 'healthy' ? '#67c23a' : '#f56c6c' }">
          ● {{ clusterInfo.status === 'healthy' ? '运行中' : clusterInfo.status }}
        </span>
      </el-descriptions-item>
      <el-descriptions-item label="环境">
        <el-tag :type="envTagType" size="small">{{ envLabel }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="节点数">{{ clusterInfo.nodeCount }}</el-descriptions-item>
      <el-descriptions-item label="认证方式">{{ clusterInfo.authType }}</el-descriptions-item>
      <el-descriptions-item label="API 地址" :span="2">{{ clusterInfo.url }}</el-descriptions-item>
      <el-descriptions-item label="创建时间">{{ formatTime(clusterInfo.createdAt) }}</el-descriptions-item>
      <el-descriptions-item label="标签" :span="3">
        <el-tag v-for="(value, key) in clusterInfo.labels" :key="key" size="small" style="margin-right: 4px">
          {{ key }}={{ value }}
        </el-tag>
        <span v-if="!clusterInfo.labels || Object.keys(clusterInfo.labels).length === 0">-</span>
      </el-descriptions-item>
    </el-descriptions>

    <!-- Stats Cards -->
    <el-row :gutter="16" style="margin-bottom: 16px">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value" style="color: #409eff">{{ workloadStats.deployment || 0 }}</div>
          <div class="stat-label">Deployments</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value" style="color: #67c23a">{{ podCount }}</div>
          <div class="stat-label">Pods</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value" style="color: #e6a23c">{{ networkStats.service || 0 }}</div>
          <div class="stat-label">Services</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value" style="color: #f56c6c">{{ storageStats.pv || 0 }} / {{ storageStats.pvc || 0 }}</div>
          <div class="stat-label">PV / PVC</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Tabs -->
    <el-tabs v-model="activeTab" type="border-card" @tab-change="handleTabChange">
      <!-- Overview Tab: Events -->
      <el-tab-pane label="概览" name="overview">
        <el-table :data="events" stripe v-loading="eventsLoading">
          <el-table-column prop="time" label="时间" width="180">
            <template #default="{ row }">{{ formatTime(row.time) }}</template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag :type="row.type === 'Normal' ? 'success' : 'warning'" size="small">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="reason" label="原因" width="150" />
          <el-table-column prop="object" label="对象" width="300" />
          <el-table-column prop="message" label="消息" />
        </el-table>
        <el-pagination
          v-model:current-page="eventPage"
          v-model:page-size="eventPageSize"
          :total="eventTotal"
          @current-change="fetchEvents"
          class="pagination-wrap"
        />
      </el-tab-pane>

      <!-- Nodes Tab -->
      <el-tab-pane label="节点" name="nodes">
        <el-table :data="nodes" stripe v-loading="nodesLoading">
          <el-table-column prop="name" label="节点名称" width="150" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }"><StatusTag :status="row.status" /></template>
          </el-table-column>
          <el-table-column prop="role" label="角色" width="100" />
          <el-table-column prop="k8sVersion" label="版本" width="120" />
          <el-table-column prop="ip" label="IP" width="140" />
          <el-table-column label="CPU" width="140">
            <template #default="{ row }">{{ row.cpuUsage }} / {{ row.cpuCapacity }}</template>
          </el-table-column>
          <el-table-column label="内存" width="160">
            <template #default="{ row }">{{ row.memoryUsage }} / {{ row.memoryCapacity }}</template>
          </el-table-column>
          <el-table-column prop="podCount" label="Pod数" width="80" />
          <el-table-column prop="age" label="运行时间" />
        </el-table>
        <el-pagination
          v-model:current-page="nodePage"
          v-model:page-size="nodePageSize"
          :total="nodeTotal"
          @current-change="fetchNodes"
          class="pagination-wrap"
        />
      </el-tab-pane>

      <!-- Workloads Tab -->
      <el-tab-pane label="工作负载" name="workloads">
        <el-table :data="workloads" stripe v-loading="workloadsLoading">
          <el-table-column prop="name" label="名称" width="200" />
          <el-table-column prop="namespace" label="命名空间" width="150" />
          <el-table-column label="类型" width="120">
            <template #default="{ row }">
              <el-tag size="small">{{ row.kind }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="副本数" width="120">
            <template #default="{ row }">{{ row.readyReplicas || 0 }} / {{ row.replicas || 0 }}</template>
          </el-table-column>
          <el-table-column prop="images" label="镜像" show-overflow-tooltip />
          <el-table-column label="创建时间" width="180">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
        </el-table>
        <el-pagination
          v-model:current-page="workloadPage"
          v-model:page-size="workloadPageSize"
          :total="workloadTotal"
          @current-change="fetchWorkloads"
          class="pagination-wrap"
        />
      </el-tab-pane>

      <!-- Network Tab -->
      <el-tab-pane label="网络" name="network">
        <el-table :data="networkResources" stripe v-loading="networkLoading">
          <el-table-column prop="name" label="名称" width="200" />
          <el-table-column prop="namespace" label="命名空间" width="150" />
          <el-table-column label="类型" width="120">
            <template #default="{ row }">
              <el-tag :type="row.kind === 'Service' ? '' : 'warning'" size="small">{{ row.kind }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="clusterIP" label="ClusterIP" width="150" />
          <el-table-column label="端口" width="200">
            <template #default="{ row }">{{ row.ports || '-' }}</template>
          </el-table-column>
          <el-table-column prop="selector" label="选择器" show-overflow-tooltip />
        </el-table>
        <el-pagination
          v-model:current-page="networkPage"
          v-model:page-size="networkPageSize"
          :total="networkTotal"
          @current-change="fetchNetworkResources"
          class="pagination-wrap"
        />
      </el-tab-pane>

      <!-- Storage Tab -->
      <el-tab-pane label="存储" name="storage">
        <el-table :data="storageResources" stripe v-loading="storageLoading">
          <el-table-column prop="name" label="名称" width="200" />
          <el-table-column prop="namespace" label="命名空间" width="150">
            <template #default="{ row }">{{ row.namespace || '-' }}</template>
          </el-table-column>
          <el-table-column label="类型" width="130">
            <template #default="{ row }">
              <el-tag :type="storageTypeTag(row.kind)" size="small">{{ row.kind }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="capacity" label="容量" width="100">
            <template #default="{ row }">{{ row.capacity || '-' }}</template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }"><StatusTag :status="row.status" /></template>
          </el-table-column>
          <el-table-column prop="storageClass" label="存储类" width="150">
            <template #default="{ row }">{{ row.storageClass || '-' }}</template>
          </el-table-column>
          <el-table-column label="访问模式" width="150">
            <template #default="{ row }">{{ (row.accessModes || []).join(', ') || '-' }}</template>
          </el-table-column>
          <el-table-column label="创建时间" width="180">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import StatusTag from '@/components/K8s/StatusTag.vue'
import { formatTime } from '@/utils/format'
import {
  getClusterDetail,
  getClusterNetworkStats,
  getClusterStorageStats,
  getClusterWorkloadStats,
  getClusterEvents
} from '@/api/cluster'
import { getNodeList } from '@/api/node'
import { getDeploymentList } from '@/api/deployment'
import { getStatefulSetList } from '@/api/statefulset'
import { getDaemonSetList } from '@/api/daemonset'
import { getServiceList } from '@/api/service'
import { getIngressList } from '@/api/ingress'
import { getPVList, getPVCList } from '@/api/storage'

const route = useRoute()
const clusterName = route.params.name

// ---- Cluster Info ----
const clusterInfo = ref({})
const envTagType = computed(() => {
  const map = { prod: 'danger', test: 'warning', dev: 'info' }
  return map[clusterInfo.value.env] || 'info'
})
const envLabel = computed(() => {
  const map = { prod: '生产环境', test: '测试环境', dev: '开发环境' }
  return map[clusterInfo.value.env] || clusterInfo.value.env
})

// ---- Stats ----
const workloadStats = ref({})
const networkStats = ref({})
const storageStats = ref({})
const podCount = computed(() => {
  return (workloadStats.value.deployment || 0) +
    (workloadStats.value.statefulset || 0) +
    (workloadStats.value.daemonset || 0)
})

// ---- Tab state ----
const activeTab = ref('overview')
const loadedTabs = ref(new Set(['overview']))

// ---- Events Tab ----
const events = ref([])
const eventsLoading = ref(false)
const eventPage = ref(1)
const eventPageSize = ref(10)
const eventTotal = ref(0)

// ---- Nodes Tab ----
const nodes = ref([])
const nodesLoading = ref(false)
const nodePage = ref(1)
const nodePageSize = ref(10)
const nodeTotal = ref(0)

// ---- Workloads Tab ----
const workloads = ref([])
const workloadsLoading = ref(false)
const workloadPage = ref(1)
const workloadPageSize = ref(10)
const workloadTotal = ref(0)

// ---- Network Tab ----
const networkResources = ref([])
const networkLoading = ref(false)
const networkPage = ref(1)
const networkPageSize = ref(10)
const networkTotal = ref(0)

// ---- Storage Tab ----
const storageResources = ref([])
const storageLoading = ref(false)

// ---- Fetch functions ----
const fetchClusterInfo = async () => {
  const res = await getClusterDetail(clusterName)
  clusterInfo.value = res.data || {}
}

const fetchStats = async () => {
  const [workload, network, storage] = await Promise.all([
    getClusterWorkloadStats(clusterName),
    getClusterNetworkStats(clusterName),
    getClusterStorageStats(clusterName)
  ])
  workloadStats.value = workload.data || {}
  networkStats.value = network.data || {}
  storageStats.value = storage.data || {}
}

const fetchEvents = async () => {
  eventsLoading.value = true
  try {
    const res = await getClusterEvents({ name: clusterName, page: eventPage.value, pageSize: eventPageSize.value })
    events.value = res.data?.items || []
    eventTotal.value = res.data?.total || 0
  } finally {
    eventsLoading.value = false
  }
}

const fetchNodes = async () => {
  nodesLoading.value = true
  try {
    const res = await getNodeList({ clusterName, page: nodePage.value, pageSize: nodePageSize.value })
    nodes.value = res.data?.items || []
    nodeTotal.value = res.data?.total || 0
  } finally {
    nodesLoading.value = false
  }
}

const fetchWorkloads = async () => {
  workloadsLoading.value = true
  try {
    const [depRes, stsRes, dsRes] = await Promise.all([
      getDeploymentList({ clusterName, page: workloadPage.value, pageSize: 100 }),
      getStatefulSetList({ clusterName, page: workloadPage.value, pageSize: 100 }),
      getDaemonSetList({ clusterName, page: workloadPage.value, pageSize: 100 })
    ])
    const depItems = (depRes.data?.items || []).map(i => ({ ...i, kind: 'Deployment' }))
    const stsItems = (stsRes.data?.items || []).map(i => ({ ...i, kind: 'StatefulSet' }))
    const dsItems = (dsRes.data?.items || []).map(i => ({ ...i, kind: 'DaemonSet' }))
    const all = [...depItems, ...stsItems, ...dsItems]
    workloadTotal.value = all.length
    const start = (workloadPage.value - 1) * workloadPageSize.value
    workloads.value = all.slice(start, start + workloadPageSize.value)
  } finally {
    workloadsLoading.value = false
  }
}

const fetchNetworkResources = async () => {
  networkLoading.value = true
  try {
    const [svcRes, ingRes] = await Promise.all([
      getServiceList({ clusterName, page: networkPage.value, pageSize: 100 }),
      getIngressList({ clusterName, page: networkPage.value, pageSize: 100 })
    ])
    const svcItems = (svcRes.data?.items || []).map(i => ({ ...i, kind: 'Service' }))
    const ingItems = (ingRes.data?.items || []).map(i => ({ ...i, kind: 'Ingress' }))
    const all = [...svcItems, ...ingItems]
    networkTotal.value = all.length
    const start = (networkPage.value - 1) * networkPageSize.value
    networkResources.value = all.slice(start, start + networkPageSize.value)
  } finally {
    networkLoading.value = false
  }
}

const fetchStorageResources = async () => {
  storageLoading.value = true
  try {
    const [pvRes, pvcRes] = await Promise.all([
      getPVList({ clusterName }),
      getPVCList({ clusterName })
    ])
    const pvItems = (pvRes.data || []).map(i => ({ ...i, kind: 'PV' }))
    const pvcItems = (pvcRes.data || []).map(i => ({ ...i, kind: 'PVC' }))
    storageResources.value = [...pvItems, ...pvcItems]
  } finally {
    storageLoading.value = false
  }
}

const handleTabChange = (tab) => {
  if (loadedTabs.value.has(tab)) return
  loadedTabs.value.add(tab)
  if (tab === 'nodes') fetchNodes()
  else if (tab === 'workloads') fetchWorkloads()
  else if (tab === 'network') fetchNetworkResources()
  else if (tab === 'storage') fetchStorageResources()
}

const storageTypeTag = (kind) => {
  if (kind === 'PV') return 'danger'
  if (kind === 'PVC') return 'warning'
  return 'info'
}

// ---- Init ----
onMounted(async () => {
  await Promise.all([fetchClusterInfo(), fetchStats()])
  fetchEvents()
})
</script>

<style scoped>
.stat-card {
  text-align: center;
  cursor: default;
}
.stat-card :deep(.el-card__body) {
  padding: 16px;
}
.stat-value {
  font-size: 28px;
  font-weight: 600;
  line-height: 1.2;
}
.stat-label {
  color: #909399;
  font-size: 13px;
  margin-top: 4px;
}
</style>
