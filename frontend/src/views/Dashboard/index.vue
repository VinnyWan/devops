<template>
  <div class="page-container">
    <div class="page-header">
      <h3>运维仪表盘</h3>
      <el-button @click="fetchAllData" :loading="loading" round><el-icon><Refresh /></el-icon>刷新</el-button>
    </div>

    <!-- Stats Cards -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4" v-for="card in statCards" :key="card.key">
        <div class="stat-card" :style="{ borderTopColor: card.color }">
          <div class="stat-label">{{ card.title }}</div>
          <div class="stat-value">{{ card.value }}</div>
          <div class="stat-sub">{{ card.sub }}</div>
        </div>
      </el-col>
    </el-row>

    <!-- Charts Row -->
    <el-row :gutter="16" class="chart-row">
      <el-col :xs="24" :md="8">
        <el-card shadow="never">
          <template #header><span>主机状态分布</span></template>
          <div class="chart-panel">
            <div v-if="hostChartStatus === 'loading'" class="chart-placeholder">
              <el-skeleton :rows="4" animated />
            </div>
            <el-empty v-else-if="hostChartStatus === 'error'" class="chart-placeholder" description="图表加载失败" :image-size="60" />
            <el-empty v-else-if="hostChartStatus === 'empty'" class="chart-placeholder" description="暂无图表数据" :image-size="60" />
            <div ref="hostChartRef" class="chart-canvas" :class="{ 'is-hidden': hostChartStatus !== 'ready' }"></div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card shadow="never">
          <template #header><span>分组主机分布</span></template>
          <div class="chart-panel">
            <div v-if="groupChartStatus === 'loading'" class="chart-placeholder">
              <el-skeleton :rows="4" animated />
            </div>
            <el-empty v-else-if="groupChartStatus === 'error'" class="chart-placeholder" description="图表加载失败" :image-size="60" />
            <el-empty v-else-if="groupChartStatus === 'empty'" class="chart-placeholder" description="暂无图表数据" :image-size="60" />
            <div ref="groupChartRef" class="chart-canvas" :class="{ 'is-hidden': groupChartStatus !== 'ready' }"></div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card shadow="never">
          <template #header><span>最近活动</span></template>
          <div class="activity-feed">
            <div v-for="event in cmdbData?.activity || []" :key="event.id + event.type" class="activity-item">
              <el-tag :type="activityTagType(event.type)" size="small" round>
                {{ activityLabel(event.type) }}
              </el-tag>
              <span class="activity-text">{{ event.message }}</span>
              <span class="activity-time">{{ relativeTime(event.timestamp) }}</span>
            </div>
            <el-empty v-if="!cmdbData?.activity?.length" description="暂无活动" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Bottom Row -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card shadow="never">
          <template #header><span>集群状态</span></template>
          <el-table :data="clusters" stripe v-loading="k8sLoading" style="width: 100%" size="small">
            <el-table-column prop="name" label="集群" width="160">
              <template #default="{ row }">
                <router-link :to="`/k8s/cluster/${row.name}`" class="cluster-link">{{ row.name }}</router-link>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <span class="status-dot" :class="row.status === 'healthy' ? 'is-healthy' : 'is-error'"></span>
                {{ row.status === 'healthy' ? '健康' : '异常' }}
              </template>
            </el-table-column>
            <el-table-column label="节点" width="60"><template #default="{ row }">{{ row.nodeCount || 0 }}</template></el-table-column>
            <el-table-column label="Pod" width="60"><template #default="{ row }">{{ row.podCount || 0 }}</template></el-table-column>
            <el-table-column label="CPU" width="120">
              <template #default="{ row }"><el-progress :percentage="row.cpuUsage || 0" :color="progressColor(row.cpuUsage)" :stroke-width="6" /></template>
            </el-table-column>
            <el-table-column label="内存" width="120">
              <template #default="{ row }"><el-progress :percentage="row.memoryUsage || 0" :color="progressColor(row.memoryUsage)" :stroke-width="6" /></template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :xs="24" :md="8">
        <el-card shadow="never">
          <template #header><span>我的主机</span></template>
          <div class="my-hosts-grid" v-if="cmdbData?.myHosts?.length">
            <div v-for="host in cmdbData.myHosts" :key="host.id" class="host-card" @click="openTerminal(host)">
              <span class="host-status-dot" :class="'status-' + host.status"></span>
              <div class="host-info">
                <div class="host-name">{{ host.hostname }}</div>
                <div class="host-ip">{{ host.ip }}</div>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无访问记录" :image-size="60" />
        </el-card>
      </el-col>

      <el-col :xs="12" :md="4">
        <el-card shadow="never">
          <template #header><span>快捷操作</span></template>
          <div class="quick-actions">
            <router-link to="/cmdb/hosts" class="action-btn">
              <el-icon><Monitor /></el-icon>
              <span>主机列表</span>
            </router-link>
            <router-link to="/cmdb/terminal/sessions" class="action-btn">
              <el-icon><VideoCamera /></el-icon>
              <span>终端审计</span>
            </router-link>
            <router-link to="/cmdb/files" class="action-btn">
              <el-icon><FolderOpened /></el-icon>
              <span>文件管理</span>
            </router-link>
            <router-link to="/cmdb/cloud-accounts" class="action-btn">
              <el-icon><Cloudy /></el-icon>
              <span>云资源</span>
            </router-link>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh, Monitor, VideoCamera, FolderOpened, Cloudy } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { getClusterList } from '@/api/cluster'
import { getCmdbDashboard } from '@/api/cmdb/dashboard'
import { getTerminalConnectWsUrl } from '@/api/cmdb/terminal'

const router = useRouter()
const loading = ref(false)
const cmdbLoading = ref(false)
const k8sLoading = ref(false)
const clusters = ref([])
const cmdbData = ref(null)
const hostChartStatus = ref('loading')
const groupChartStatus = ref('loading')

const hostChartRef = ref(null)
const groupChartRef = ref(null)
let hostChart = null
let groupChart = null

const COLORS = {
  primary: '#2563EB',
  primaryLight: '#3B82F6',
  success: '#10B981',
  warning: '#F59E0B',
  danger: '#EF4444',
  muted: '#94A3B8',
  info: '#6366F1',
  teal: '#14B8A6',
  amber: '#F59E0B'
}

const statCards = computed(() => {
  const d = cmdbData.value
  if (!d) return []
  return [
    { key: 'hosts', title: '主机总数', value: d.stats.hosts.total, sub: d.stats.hosts.online + ' 在线', color: COLORS.success },
    { key: 'terminals', title: '活跃终端', value: d.stats.terminals.activeCount, sub: d.stats.terminals.onlineUsers + ' 用户在线', color: COLORS.amber },
    { key: 'todaySessions', title: '今日会话', value: d.stats.terminals.todayCount, sub: '', color: COLORS.primary },
    { key: 'cloud', title: '云实例', value: d.stats.cloud.instanceCount, sub: d.stats.cloud.lastSyncAt ? '同步于 ' + d.stats.cloud.lastSyncAt : '未同步', color: COLORS.info },
    { key: 'fileOps', title: '今日文件操作', value: d.stats.files.todayOps, sub: '', color: COLORS.teal },
    { key: 'clusters', title: 'K8s 集群', value: clusters.value.length, sub: clusters.value.reduce((s, c) => s + (c.podCount || 0), 0) + ' pods', color: COLORS.danger }
  ]
})

const progressColor = (p) => {
  if (p >= 90) return COLORS.danger
  if (p >= 70) return COLORS.warning
  return COLORS.success
}

const activityTagType = (type) => {
  const map = { terminal: 'danger', file: 'warning', sync: 'success', host: '' }
  return map[type] || 'info'
}

const activityLabel = (type) => {
  const map = { terminal: '终端', file: '文件', sync: '同步', host: '主机' }
  return map[type] || type
}

const relativeTime = (ts) => {
  if (!ts) return ''
  const diff = Date.now() - new Date(ts).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return mins + '分钟前'
  const hours = Math.floor(mins / 60)
  if (hours < 24) return hours + '小时前'
  return Math.floor(hours / 24) + '天前'
}

const openTerminal = (host) => {
  router.push({ path: '/cmdb/hosts', query: { terminalHostId: host.id } })
}

const renderHostChart = () => {
  if (!hostChartRef.value || !cmdbData.value) return

  const h = cmdbData.value.stats.hosts
  const chartData = [
    { value: h.online, name: '在线', itemStyle: { color: COLORS.success } },
    { value: h.warning, name: '告警', itemStyle: { color: COLORS.warning } },
    { value: h.offline, name: '离线', itemStyle: { color: COLORS.danger } },
    { value: h.unknown, name: '未知', itemStyle: { color: COLORS.muted } }
  ].filter(d => d.value > 0)

  if (!chartData.length) {
    hostChart?.clear()
    hostChartStatus.value = 'empty'
    return
  }

  if (!hostChart) {
    hostChart = echarts.init(hostChartRef.value)
  }
  hostChart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, textStyle: { fontSize: 11, color: '#64748B' } },
    series: [{
      type: 'pie',
      radius: ['40%', '65%'],
      center: ['50%', '45%'],
      label: { show: false },
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      data: chartData
    }]
  })
  hostChartStatus.value = 'ready'
}

const renderGroupChart = () => {
  if (!groupChartRef.value || !cmdbData.value) return

  const groups = cmdbData.value.stats.hosts.byGroup || []
  const chartData = groups.map(g => g.count).reverse().filter(value => value > 0)

  if (!chartData.length) {
    groupChart?.clear()
    groupChartStatus.value = 'empty'
    return
  }

  if (!groupChart) {
    groupChart = echarts.init(groupChartRef.value)
  }
  groupChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: 80, right: 20, top: 10, bottom: 30 },
    xAxis: { type: 'value', axisLine: { lineStyle: { color: '#E2E8F0' } }, splitLine: { lineStyle: { color: '#F1F5F9' } } },
    yAxis: {
      type: 'category',
      data: groups.filter(g => g.count > 0).map(g => g.groupName).reverse(),
      axisLabel: { fontSize: 11, width: 70, overflow: 'truncate', color: '#64748B' },
      axisLine: { lineStyle: { color: '#E2E8F0' } }
    },
    series: [{
      type: 'bar',
      data: chartData,
      itemStyle: { color: COLORS.primary, borderRadius: [0, 6, 6, 0] },
      barWidth: 16,
      label: { show: true, position: 'right', fontSize: 11, color: '#64748B' }
    }]
  })
  groupChartStatus.value = 'ready'
}

const fetchCmdbData = async () => {
  cmdbLoading.value = true
  hostChartStatus.value = 'loading'
  groupChartStatus.value = 'loading'
  try {
    const res = await getCmdbDashboard()
    cmdbData.value = res.data
    await nextTick()
    renderHostChart()
    renderGroupChart()
  } catch (e) {
    cmdbData.value = null
    hostChart?.clear()
    groupChart?.clear()
    hostChartStatus.value = 'error'
    groupChartStatus.value = 'error'
    ElMessage.error('获取仪表盘数据失败')
  } finally {
    cmdbLoading.value = false
  }
}

const fetchK8sData = async () => {
  k8sLoading.value = true
  try {
    const res = await getClusterList()
    clusters.value = res.data?.list || res.data || []
  } finally {
    k8sLoading.value = false
  }
}

const fetchAllData = async () => {
  loading.value = true
  try {
    await Promise.all([fetchCmdbData(), fetchK8sData()])
  } finally {
    loading.value = false
  }
}

const handleResize = () => {
  hostChart?.resize()
  groupChart?.resize()
}

onMounted(fetchAllData)
onMounted(() => window.addEventListener('resize', handleResize))
onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  hostChart?.dispose()
  groupChart?.dispose()
})
</script>

<style scoped>
.stat-row {
  margin-bottom: var(--spacing-lg);
}

.stat-card {
  background: var(--color-bg-white);
  border: 1px solid var(--color-border-light);
  border-top: 3px solid var(--color-primary);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  text-align: center;
  transition: all var(--transition-fast);
}
.stat-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}

.stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
  margin-bottom: var(--spacing-xs);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}
.stat-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
  color: var(--color-text);
}
.stat-sub {
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
  margin-top: 4px;
}

.chart-row {
  margin-bottom: var(--spacing-lg);
}

.chart-panel {
  position: relative;
  min-height: 240px;
}

.chart-canvas {
  height: 240px;
}

.chart-canvas.is-hidden {
  visibility: hidden;
}

.chart-placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: var(--color-bg-white);
}

.activity-feed {
  max-height: 240px;
  overflow-y: auto;
}
.activity-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: 8px 0;
  border-bottom: 1px solid var(--color-border-light);
  font-size: var(--font-size-sm);
}
.activity-item:last-child {
  border-bottom: none;
}
.activity-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text);
}
.activity-time {
  color: var(--color-text-tertiary);
  flex-shrink: 0;
  font-size: var(--font-size-xs);
}

.my-hosts-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-sm);
}
.host-card {
  background: var(--color-bg-muted);
  border-radius: var(--radius-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  transition: all var(--transition-fast);
  border: 1px solid transparent;
}
.host-card:hover {
  background: var(--color-primary-lightest);
  border-color: var(--color-primary-lighter);
}

.host-status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.host-status-dot.status-online { background: var(--color-success); }
.host-status-dot.status-offline { background: var(--color-danger); }
.host-status-dot.status-warning { background: var(--color-warning); }
.host-status-dot.status-unknown { background: var(--color-text-tertiary); }

.host-info { min-width: 0; }
.host-name { font-size: var(--font-size-sm); font-weight: 600; color: var(--color-text); }
.host-ip { font-size: var(--font-size-xs); color: var(--color-text-tertiary); }

.quick-actions {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}
.action-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  text-decoration: none;
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border-light);
  transition: all var(--transition-fast);
  font-weight: 500;
}
.action-btn:hover {
  background: var(--color-primary-lightest);
  color: var(--color-primary);
  border-color: var(--color-primary-lighter);
}

.cluster-link {
  color: var(--color-primary);
  text-decoration: none;
  font-weight: 500;
  transition: color var(--transition-fast);
}
.cluster-link:hover {
  color: var(--color-primary-dark);
}

.status-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 4px;
}
.status-dot.is-healthy { background: var(--color-success); }
.status-dot.is-error { background: var(--color-danger); }
</style>
