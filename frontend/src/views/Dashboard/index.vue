<template>
  <div class="page-container">
    <div class="page-header">
      <h3>仪表盘</h3>
      <el-button @click="fetchDashboardData"><el-icon><Refresh /></el-icon>刷新</el-button>
    </div>

    <el-row :gutter="16" style="margin-bottom: 24px;">
      <el-col :span="6" v-for="card in statCards" :key="card.title">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-card-content">
            <div class="stat-info">
              <div class="stat-title">{{ card.title }}</div>
              <div class="stat-value" :style="{ color: card.color }">{{ card.value }}</div>
            </div>
            <el-icon :size="48" :style="{ color: card.color }"><component :is="card.icon" /></el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" style="margin-bottom: 24px;">
      <template #header><span>集群状态</span></template>
      <el-table :data="clusters" stripe v-loading="loading" style="width: 100%">
        <el-table-column prop="name" label="集群名称" width="200">
          <template #default="{ row }">
            <router-link :to="`/k8s/cluster/${row.name}`" style="color: var(--el-color-primary); text-decoration: none;">{{ row.name }}</router-link>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.status === 'healthy' ? 'success' : row.status === 'warning' ? 'warning' : 'danger'">{{ row.status === 'healthy' ? '健康' : row.status === 'warning' ? '告警' : '异常' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="节点" width="100"><template #default="{ row }">{{ row.nodeCount || 0 }}</template></el-table-column>
        <el-table-column label="Pod 总数" width="100"><template #default="{ row }">{{ row.podCount || 0 }}</template></el-table-column>
        <el-table-column label="CPU 使用率" width="200">
          <template #default="{ row }"><el-progress :percentage="row.cpuUsage || 0" :color="getProgressColor(row.cpuUsage)" :stroke-width="12" /></template>
        </el-table-column>
        <el-table-column label="内存使用率" width="200">
          <template #default="{ row }"><el-progress :percentage="row.memoryUsage || 0" :color="getProgressColor(row.memoryUsage)" :stroke-width="12" /></template>
        </el-table-column>
        <el-table-column prop="version" label="K8s 版本" width="130" />
      </el-table>
    </el-card>

    <el-card shadow="never">
      <template #header><span>最近事件</span></template>
      <el-table :data="recentEvents" stripe max-height="300" style="width: 100%">
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }"><el-tag :type="row.type === 'Warning' ? 'warning' : 'info'" size="small">{{ row.type }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="cluster" label="集群" width="150" />
        <el-table-column prop="namespace" label="命名空间" width="180" />
        <el-table-column prop="object" label="对象" width="250" />
        <el-table-column prop="message" label="消息" />
        <el-table-column label="最后发生" width="180">
          <template #default="{ row }">{{ formatTime(row.lastSeen) }}</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Refresh, Monitor, Cpu, Grid, Warning } from '@element-plus/icons-vue'
import { getClusterList } from '@/api/cluster'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const clusters = ref([])
const recentEvents = ref([])

const statCards = computed(() => [
  { title: '集群数量', value: clusters.value.length, icon: Monitor, color: '#409EFF' },
  { title: '节点总数', value: clusters.value.reduce((sum, c) => sum + (c.nodeCount || 0), 0), icon: Cpu, color: '#67C23A' },
  { title: 'Pod 总数', value: clusters.value.reduce((sum, c) => sum + (c.podCount || 0), 0), icon: Grid, color: '#E6A23C' },
  { title: '告警集群', value: clusters.value.filter(c => c.status !== 'healthy').length, icon: Warning, color: '#F56C6C' }
])

const getProgressColor = (p) => { if (p >= 90) return '#F56C6C'; if (p >= 70) return '#E6A23C'; return '#67C23A' }

const fetchDashboardData = async () => {
  loading.value = true
  try {
    const res = await getClusterList()
    clusters.value = res.data?.list || res.data || []
    recentEvents.value = []
    clusters.value.forEach(c => {
      if (c.events && c.events.length) c.events.forEach(e => recentEvents.value.push({ ...e, cluster: c.name }))
    })
    recentEvents.value.sort((a, b) => (b.lastSeen || '').localeCompare(a.lastSeen || ''))
    recentEvents.value = recentEvents.value.slice(0, 20)
  } finally { loading.value = false }
}

onMounted(fetchDashboardData)
</script>

<style scoped>
.page-container { padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.stat-card-content { display: flex; justify-content: space-between; align-items: center; }
.stat-title { font-size: 14px; color: #909399; margin-bottom: 8px; }
.stat-value { font-size: 28px; font-weight: 600; }
</style>
