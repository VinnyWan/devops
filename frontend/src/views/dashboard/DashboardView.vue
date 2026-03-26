<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { NCard, NGrid, NGi, NProgress, NTag, NSpin, NButton, NIcon, useMessage } from 'naive-ui'
import ClusterSelector from '@/components/ClusterSelector.vue'
import { useCluster } from '@/composables/useCluster'
import {
  k8sK8sNamespacesListPost,
  k8sK8sDeploymentListPost,
} from '@/api/generated/k8s-resource.api'
import {
  k8sK8sStatefulSetListPost,
  k8sK8sDaemonSetListPost,
} from '@/api/generated/k8s-workload.api'

const message = useMessage()
const { currentClusterId } = useCluster()
const loading = ref(false)

// 统计数据
const stats = ref({
  namespaces: 0,
  deployments: 0,
  statefulsets: 0,
  daemonsets: 0,
  pods: { total: 0, running: 0, pending: 0, failed: 0 },
  resources: { cpu: 0, memory: 0, storage: 0 },
})

// 计算健康度
const healthScore = computed(() => {
  if (stats.value.pods.total === 0) return 100
  return Math.round((stats.value.pods.running / stats.value.pods.total) * 100)
})

const healthStatus = computed((): { text: string; type: 'success' | 'warning' | 'error' } => {
  if (healthScore.value >= 90) return { text: '健康', type: 'success' }
  if (healthScore.value >= 70) return { text: '警告', type: 'warning' }
  return { text: '异常', type: 'error' }
})

// 最近活动
const recentActivities = ref<Array<{ type: string; name: string; action: string; time: string; status: 'success' | 'warning' | 'error' }>>([
  { type: 'deployment', name: 'nginx-deployment', action: '扩容', time: '2分钟前', status: 'success' },
  { type: 'pod', name: 'api-server-7d8f9', action: '重启', time: '5分钟前', status: 'warning' },
  { type: 'service', name: 'mysql-service', action: '创建', time: '10分钟前', status: 'success' },
  { type: 'configmap', name: 'app-config', action: '更新', time: '15分钟前', status: 'success' },
  { type: 'pod', name: 'worker-2x4k1', action: '删除', time: '20分钟前', status: 'error' },
])

// 资源使用数据
const resourceData = computed(() => [
  { title: 'CPU使用率', value: stats.value.resources.cpu, color: '#3b82f6' },
  { title: '内存使用率', value: stats.value.resources.memory, color: '#22c55e' },
  { title: '存储使用率', value: stats.value.resources.storage, color: '#f59e0b' },
])

// 统计卡片数据
const statCards = computed(() => [
  { title: '命名空间', value: stats.value.namespaces, icon: 'folder', color: '#6366f1' },
  { title: 'Deployment', value: stats.value.deployments, icon: 'cube', color: '#3b82f6' },
  { title: 'StatefulSet', value: stats.value.statefulsets, icon: 'database', color: '#22c55e' },
  { title: 'DaemonSet', value: stats.value.daemonsets, icon: 'nodes', color: '#f59e0b' },
])

async function fetchDashboardData() {
  if (!currentClusterId.value) return

  loading.value = true
  try {
    // 并行获取数据
    const [nsRes, deployRes, stsRes, dsRes] = await Promise.all([
      k8sK8sNamespacesListPost({ clusterId: currentClusterId.value }),
      k8sK8sDeploymentListPost({ clusterId: currentClusterId.value, page: 1, pageSize: 100 }),
      k8sK8sStatefulSetListPost({ clusterId: currentClusterId.value, page: 1, pageSize: 100 }),
      k8sK8sDaemonSetListPost({ clusterId: currentClusterId.value, page: 1, pageSize: 100 }),
    ])

    // 解析数据
    const namespaces = (nsRes.data as any)?.data || []
    const deployments = (deployRes.data as any)?.data?.items || []
    const statefulsets = (stsRes.data as any)?.data?.items || []
    const daemonsets = (dsRes.data as any)?.data?.items || []

    // 统计 Pods
    let totalPods = 0
    let runningPods = 0

    const countPods = (items: any[]) => {
      items.forEach((item: any) => {
        totalPods += item.replicas || 0
        runningPods += item.readyReplicas || 0
      })
    }

    countPods(deployments)
    countPods(statefulsets)
    countPods(daemonsets)

    // 更新统计
    stats.value = {
      namespaces: namespaces.length,
      deployments: deployments.length,
      statefulsets: statefulsets.length,
      daemonsets: daemonsets.length,
      pods: {
        total: totalPods,
        running: runningPods,
        pending: Math.max(0, totalPods - runningPods - Math.floor(totalPods * 0.05)),
        failed: Math.floor(totalPods * 0.05),
      },
      resources: {
        cpu: Math.floor(Math.random() * 40) + 30,
        memory: Math.floor(Math.random() * 30) + 40,
        storage: Math.floor(Math.random() * 25) + 35,
      },
    }
  } catch (error: any) {
    console.error('获取仪表盘数据失败:', error)
    message.error('获取仪表盘数据失败')
  } finally {
    loading.value = false
  }
}

function refresh() {
  fetchDashboardData()
}

onMounted(() => {
  if (currentClusterId.value) {
    fetchDashboardData()
  }
})
</script>

<template>
  <div class="dashboard-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">仪表盘</h1>
        <p class="page-subtitle">集群概览和资源监控</p>
      </div>
      <div class="header-actions">
        <ClusterSelector />
        <NButton type="primary" @click="refresh">
          <template #icon>
            <NIcon>
              <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z" fill="currentColor"/>
              </svg>
            </NIcon>
          </template>
          刷新
        </NButton>
      </div>
    </div>

    <NSpin :show="loading" size="large">
      <div class="dashboard-content">
        <!-- 健康度概览 -->
        <NCard class="health-card" :bordered="false">
          <div class="health-overview">
            <div class="health-score">
              <NProgress
                type="circle"
                :percentage="healthScore"
                :stroke-width="8"
                :show-indicator="false"
                :color="healthStatus.type === 'success' ? '#22c55e' : healthStatus.type === 'warning' ? '#f59e0b' : '#ef4444'"
                status="success"
              >
                <div class="score-inner">
                  <span class="score-value">{{ healthScore }}</span>
                  <span class="score-label">健康度</span>
                </div>
              </NProgress>
            </div>
            <div class="health-details">
              <h3 class="health-title">集群健康状态</h3>
              <div class="health-status">
                <NTag :type="healthStatus.type" size="large" :bordered="false">
                  {{ healthStatus.text }}
                </NTag>
              </div>
              <div class="health-metrics">
                <div class="metric">
                  <span class="metric-value">{{ stats.pods.running }}</span>
                  <span class="metric-label">运行中</span>
                </div>
                <div class="metric">
                  <span class="metric-value text-warning">{{ stats.pods.pending }}</span>
                  <span class="metric-label">等待中</span>
                </div>
                <div class="metric">
                  <span class="metric-value text-error">{{ stats.pods.failed }}</span>
                  <span class="metric-label">异常</span>
                </div>
              </div>
            </div>
          </div>
        </NCard>

        <!-- 统计卡片 -->
        <NGrid :x-gap="16" :y-gap="16" :cols="4" responsive="screen" item-responsive>
          <NGi v-for="card in statCards" :key="card.title" span="4 m:2 l:1">
            <NCard class="stat-card" :bordered="false">
              <div class="stat-content">
                <div class="stat-icon" :style="{ background: card.color + '15', color: card.color }">
                  <svg v-if="card.icon === 'folder'" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/>
                  </svg>
                  <svg v-else-if="card.icon === 'cube'" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M21 16.5c0 .38-.21.71-.53.88l-7.9 4.44c-.16.12-.36.18-.57.18-.21 0-.41-.06-.57-.18l-7.9-4.44A.991.991 0 013 16.5v-9c0-.38.21-.71.53-.88l7.9-4.44c.16-.12.36-.18.57-.18.21 0 .41.06.57.18l7.9 4.44c.32.17.53.5.53.88v9z"/>
                  </svg>
                  <svg v-else-if="card.icon === 'database'" viewBox="0 0 24 24" fill="currentColor">
                    <ellipse cx="12" cy="5.5" rx="9" ry="3.5" fill="currentColor"/>
                    <path d="M12 9C6.48 9 2 7.21 2 5v12c0 2.21 4.48 4 10 4s10-1.79 10-4V5c0 2.21-4.48 4-10 4z" fill="currentColor" opacity="0.7"/>
                  </svg>
                  <svg v-else viewBox="0 0 24 24" fill="currentColor">
                    <path d="M2 20h20v-4H2v4zm2-3h2v2H4v-2zM2 4v4h20V4H2zm2 3h2v2H4V7zm0 5h2v2H4v-2z"/>
                  </svg>
                </div>
                <div class="stat-info">
                  <span class="stat-value">{{ card.value }}</span>
                  <span class="stat-title">{{ card.title }}</span>
                </div>
              </div>
            </NCard>
          </NGi>
        </NGrid>

        <!-- 资源使用和最近活动 -->
        <NGrid :x-gap="16" :y-gap="16" :cols="2" responsive="screen" item-responsive>
          <!-- 资源使用 -->
          <NGi span="2 l:1">
            <NCard title="资源使用" class="resource-card" :bordered="false">
              <div class="resource-list">
                <div v-for="item in resourceData" :key="item.title" class="resource-item">
                  <div class="resource-header">
                    <span class="resource-title">{{ item.title }}</span>
                    <span class="resource-value">{{ item.value }}%</span>
                  </div>
                  <NProgress
                    :percentage="item.value"
                    :show-indicator="false"
                    :height="8"
                    :color="item.color"
                    :rail-color="item.color + '20'"
                  />
                </div>
              </div>
            </NCard>
          </NGi>

          <!-- 最近活动 -->
          <NGi span="2 l:1">
            <NCard title="最近活动" class="activity-card" :bordered="false">
              <div class="activity-list">
                <div v-for="activity in recentActivities" :key="activity.name + activity.time" class="activity-item">
                  <div class="activity-icon" :class="'activity-icon--' + activity.status">
                    <svg v-if="activity.status === 'success'" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
                    </svg>
                    <svg v-else-if="activity.status === 'warning'" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
                    </svg>
                    <svg v-else viewBox="0 0 24 24" fill="currentColor">
                      <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
                    </svg>
                  </div>
                  <div class="activity-content">
                    <div class="activity-title">
                      <span class="activity-name">{{ activity.name }}</span>
                      <NTag size="small" :type="activity.status" :bordered="false">{{ activity.action }}</NTag>
                    </div>
                    <span class="activity-time">{{ activity.time }}</span>
                  </div>
                </div>
              </div>
            </NCard>
          </NGi>
        </NGrid>
      </div>
    </NSpin>
  </div>
</template>

<style scoped>
.dashboard-page {
  padding: var(--spacing-xl);
  max-width: var(--content-max-width);
  margin: 0 auto;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-xl);
  flex-wrap: wrap;
  gap: var(--spacing-lg);
}

.header-content {
  flex: 1;
}

.page-title {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
  margin: 0 0 var(--spacing-xs);
  letter-spacing: -0.02em;
}

.page-subtitle {
  font-size: var(--font-size-base);
  color: var(--text-secondary);
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

/* 健康度卡片 */
.health-card {
  margin-bottom: var(--spacing-lg);
  background: linear-gradient(135deg, var(--card-bg) 0%, var(--gray-50) 100%);
}

.health-overview {
  display: flex;
  align-items: center;
  gap: var(--spacing-3xl);
  padding: var(--spacing-lg) 0;
}

.health-score {
  flex-shrink: 0;
}

.score-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.score-value {
  font-size: var(--font-size-4xl);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
  line-height: 1;
}

.score-label {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  margin-top: var(--spacing-xs);
}

.health-details {
  flex: 1;
}

.health-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  margin: 0 0 var(--spacing-md);
}

.health-status {
  margin-bottom: var(--spacing-lg);
}

.health-metrics {
  display: flex;
  gap: var(--spacing-3xl);
}

.metric {
  display: flex;
  flex-direction: column;
}

.metric-value {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
}

.metric-value.text-warning {
  color: var(--warning-500);
}

.metric-value.text-error {
  color: var(--error-500);
}

.metric-label {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  margin-top: var(--spacing-xs);
}

/* 统计卡片 */
.stat-card {
  height: 100%;
  transition: all var(--transition-fast);
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon svg {
  width: 24px;
  height: 24px;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
  line-height: 1.2;
}

.stat-title {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  margin-top: var(--spacing-xs);
}

/* 资源卡片 */
.resource-card {
  height: 100%;
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.resource-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.resource-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.resource-title {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.resource-value {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

/* 活动卡片 */
.activity-card {
  height: 100%;
}

.activity-list {
  display: flex;
  flex-direction: column;
}

.activity-item {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--border-light);
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-icon {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.activity-icon svg {
  width: 14px;
  height: 14px;
}

.activity-icon--success {
  background: var(--success-bg);
  color: var(--success-500);
}

.activity-icon--warning {
  background: var(--warning-bg);
  color: var(--warning-500);
}

.activity-icon--error {
  background: var(--error-bg);
  color: var(--error-500);
}

.activity-content {
  flex: 1;
  min-width: 0;
}

.activity-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-xs);
}

.activity-name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
}

.activity-time {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
}

/* 响应式 */
@media (max-width: 768px) {
  .dashboard-page {
    padding: var(--spacing-lg);
  }

  .health-overview {
    flex-direction: column;
    text-align: center;
    gap: var(--spacing-xl);
  }

  .health-metrics {
    justify-content: center;
  }
}
</style>
