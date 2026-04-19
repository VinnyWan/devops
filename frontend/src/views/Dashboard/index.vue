<template>
  <div class="page-container">
    <div class="page-header">
      <h3>运维仪表盘</h3>
      <el-button @click="fetchAllData" :loading="loading"><el-icon><Refresh /></el-icon>刷新</el-button>
    </div>

    <!-- Stats Cards -->
    <el-row :gutter="12" style="margin-bottom: 20px;">
      <el-col :span="4" v-for="card in statCards" :key="card.key">
        <el-card shadow="hover" class="stat-card" :body-style="{ padding: '16px' }">
          <div class="stat-label">{{ card.title }}</div>
          <div class="stat-value" :style="{ color: card.color }">{{ card.value }}</div>
          <div class="stat-sub">{{ card.sub }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Charts Row -->
    <el-row :gutter="16" style="margin-bottom: 20px;">
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>主机状态分布</span></template>
          <div ref="hostChartRef" style="height: 240px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>分组主机分布</span></template>
          <div ref="groupChartRef" style="height: 240px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>最近活动</span></template>
          <div class="activity-feed">
            <div v-for="event in cmdbData?.activity || []" :key="event.id + event.type" class="activity-item">
              <el-tag :type="activityTagType(event.type)" size="small" class="activity-tag">
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

    <!-- Bottom Row: K8s Clusters + My Hosts + Quick Actions -->
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card shadow="never">
          <template #header><span>集群状态</span></template>
          <el-table :data="clusters" stripe v-loading="k8sLoading" style="width: 100%" size="small">
            <el-table-column prop="name" label="集群" width="160">
              <template #default="{ row }">
                <router-link :to="`/k8s/cluster/${row.name}`" class="link">{{ row.name }}</router-link>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 'healthy' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'healthy' ? '健康' : '异常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="节点" width="60"><template #default="{ row }">{{ row.nodeCount || 0 }}</template></el-table-column>
            <el-table-column label="Pod" width="60"><template #default="{ row }">{{ row.podCount || 0 }}</template></el-table-column>
            <el-table-column label="CPU" width="120">
              <template #default="{ row }"><el-progress :percentage="row.cpuUsage || 0" :color="progressColor(row.cpuUsage)" :stroke-width="8" /></template>
            </el-table-column>
            <el-table-column label="内存" width="120">
              <template #default="{ row }"><el-progress :percentage="row.memoryUsage || 0" :color="progressColor(row.memoryUsage)" :stroke-width="8" /></template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span>我的主机</span></template>
          <div class="my-hosts-grid" v-if="cmdbData?.myHosts?.length">
            <div v-for="host in cmdbData.myHosts" :key="host.id" class="host-card" @click="openTerminal(host)">
              <div class="host-status" :class="'status-' + host.status"></div>
              <div class="host-name">{{ host.hostname }}</div>
              <div class="host-ip">{{ host.ip }}</div>
            </div>
          </div>
          <el-empty v-else description="暂无访问记录" :image-size="60" />
        </el-card>
      </el-col>

      <el-col :span="4">
        <el-card shadow="never">
          <template #header><span>快捷操作</span></template>
          <div class="quick-actions">
            <router-link to="/cmdb/hosts" class="action-btn primary">
              <el-icon><Monitor /></el-icon>
              <span>主机列表</span>
            </router-link>
            <router-link to="/cmdb/terminal/sessions" class="action-btn warning">
              <el-icon><VideoCamera /></el-icon>
              <span>终端审计</span>
            </router-link>
            <router-link to="/cmdb/files" class="action-btn success">
              <el-icon><FolderOpened /></el-icon>
              <span>文件管理</span>
            </router-link>
            <router-link to="/cmdb/cloud-accounts" class="action-btn info">
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
const k8sLoading = ref(false)
const clusters = ref([])
const cmdbData = ref(null)

const hostChartRef = ref(null)
const groupChartRef = ref(null)
let hostChart = null
let groupChart = null

const statCards = computed(() => {
  const d = cmdbData.value
  if (!d) return []
  return [
    { key: 'hosts', title: '主机总数', value: d.stats.hosts.total, sub: d.stats.hosts.online + ' 在线', color: '#67C23A' },
    { key: 'terminals', title: '活跃终端', value: d.stats.terminals.activeCount, sub: d.stats.terminals.onlineUsers + ' 用户在线', color: '#E6A23C' },
    { key: 'todaySessions', title: '今日会话', value: d.stats.terminals.todayCount, sub: '', color: '#409EFF' },
    { key: 'cloud', title: '云实例', value: d.stats.cloud.instanceCount, sub: d.stats.cloud.lastSyncAt ? '同步于 ' + d.stats.cloud.lastSyncAt : '未同步', color: '#9B59B6' },
    { key: 'fileOps', title: '今日文件操作', value: d.stats.files.todayOps, sub: '', color: '#3498DB' },
    { key: 'clusters', title: 'K8s 集群', value: clusters.value.length, sub: clusters.value.reduce((s, c) => s + (c.podCount || 0), 0) + ' pods', color: '#F56C6C' }
  ]
})

const progressColor = (p) => { if (p >= 90) return '#F56C6C'; if (p >= 70) return '#E6A23C'; return '#67C23A' }

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
  if (!hostChart) {
    hostChart = echarts.init(hostChartRef.value)
  }
  const h = cmdbData.value.stats.hosts
  hostChart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, textStyle: { fontSize: 11 } },
    series: [{
      type: 'pie',
      radius: ['40%', '65%'],
      center: ['50%', '45%'],
      label: { show: false },
      data: [
        { value: h.online, name: '在线', itemStyle: { color: '#67C23A' } },
        { value: h.warning, name: '告警', itemStyle: { color: '#E6A23C' } },
        { value: h.offline, name: '离线', itemStyle: { color: '#F56C6C' } },
        { value: h.unknown, name: '未知', itemStyle: { color: '#909399' } }
      ].filter(d => d.value > 0)
    }]
  })
}

const renderGroupChart = () => {
  if (!groupChartRef.value || !cmdbData.value) return
  if (!groupChart) {
    groupChart = echarts.init(groupChartRef.value)
  }
  const groups = cmdbData.value.stats.hosts.byGroup || []
  groupChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: 80, right: 20, top: 10, bottom: 30 },
    xAxis: { type: 'value' },
    yAxis: {
      type: 'category',
      data: groups.map(g => g.groupName).reverse(),
      axisLabel: { fontSize: 11, width: 70, overflow: 'truncate' }
    },
    series: [{
      type: 'bar',
      data: groups.map(g => g.count).reverse(),
      itemStyle: { color: '#409EFF', borderRadius: [0, 4, 4, 0] },
      barWidth: 16,
      label: { show: true, position: 'right', fontSize: 11 }
    }]
  })
}

const fetchCmdbData = async () => {
  try {
    const res = await getCmdbDashboard()
    cmdbData.value = res.data
    await nextTick()
    renderHostChart()
    renderGroupChart()
  } catch (e) {
    ElMessage.error('获取仪表盘数据失败')
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
.page-container { padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }

.stat-card { text-align: center; }
.stat-label { font-size: 12px; color: #909399; margin-bottom: 6px; }
.stat-value { font-size: 28px; font-weight: 700; line-height: 1.2; }
.stat-sub { font-size: 11px; color: #b0b5bd; margin-top: 4px; }

.activity-feed { max-height: 240px; overflow-y: auto; }
.activity-item { display: flex; align-items: center; gap: 8px; padding: 6px 0; border-bottom: 1px solid #f0f0f0; font-size: 12px; }
.activity-tag { flex-shrink: 0; width: 36px; text-align: center; }
.activity-text { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.activity-time { color: #909399; flex-shrink: 0; font-size: 11px; }

.my-hosts-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.host-card { background: #f5f7fa; border-radius: 6px; padding: 10px; cursor: pointer; transition: background 0.2s; }
.host-card:hover { background: #ecf5ff; }
.host-status { display: inline-block; width: 6px; height: 6px; border-radius: 50%; margin-right: 4px; }
.status-online { background: #67C23A; }
.status-offline { background: #F56C6C; }
.status-warning { background: #E6A23C; }
.status-unknown { background: #909399; }
.host-name { font-size: 13px; font-weight: 600; }
.host-ip { font-size: 11px; color: #909399; }

.quick-actions { display: flex; flex-direction: column; gap: 8px; }
.action-btn { display: flex; align-items: center; gap: 6px; padding: 10px 14px; border-radius: 6px; color: #fff; font-size: 13px; text-decoration: none; transition: opacity 0.2s; }
.action-btn:hover { opacity: 0.85; color: #fff; }
.action-btn.primary { background: #409EFF; }
.action-btn.warning { background: #E6A23C; }
.action-btn.success { background: #67C23A; }
.action-btn.info { background: #909399; }

.link { color: var(--el-color-primary); text-decoration: none; }
.link:hover { text-decoration: underline; }
</style>
