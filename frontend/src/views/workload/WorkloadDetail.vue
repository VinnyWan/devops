<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  NCard,
  NSpace,
  NDataTable,
  NButton,
  NTag,
  NTabs,
  NTabPane,
  NModal,
  useMessage,
  NIcon,
  NSpin,
} from 'naive-ui'
import K8sTerminal from '@/components/K8sTerminal.vue'
import { useCluster } from '@/composables/useCluster'
import {
  k8sK8sDeploymentDetailPost,
  k8sK8sDeploymentPodsPost,
  k8sK8sDeploymentYamlPost,
  k8sK8sPodLogsGet,
  k8sK8sPodYamlGet,
  k8sK8sPodDeletePost,
  k8sK8sPodListByOwnerPost,
} from '@/api/generated/k8s-resource.api'
import {
  k8sK8sStatefulSetYamlPost,
  k8sK8sDaemonSetYamlPost,
} from '@/api/generated/k8s-workload.api'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const { currentClusterId } = useCluster()

const loading = ref(false)
const detailLoading = ref(false)
const workloadDetail = ref<any>(null)
const pods = ref<any[]>([])
const yamlContent = ref('')
const yamlLoading = ref(false)

// Terminal modal
const showTerminalModal = ref(false)
const terminalWsUrl = ref('')
const selectedPod = ref<any>(null)

// YAML modal
const showYamlModal = ref(false)

// Delete confirm modal
const showDeleteModal = ref(false)
const deleteTarget = ref<{ type: string; name: string } | null>(null)

// Logs modal
const showLogsModal = ref(false)
const logsContent = ref('')
const logsLoading = ref(false)
const currentPodForLogs = ref<any>(null)

// Pod YAML modal
const showPodYamlModal = ref(false)
const podYamlContent = ref('')
const podYamlLoading = ref(false)
const currentPodForYaml = ref<any>(null)

// Active tab
const activeTab = ref('pods')

// Workload type from route
const workloadType = computed(() => route.params.type as 'deployment' | 'statefulset' | 'daemonset')
const workloadNamespace = computed(() => route.params.namespace as string)
const workloadName = computed(() => route.params.name as string)

// Workload type display name
const typeDisplayNames: Record<string, string> = {
  deployment: 'Deployment',
  statefulset: 'StatefulSet',
  daemonset: 'DaemonSet',
}

// Pod table columns
const podColumns = computed(() => [
  {
    title: 'Pod名称',
    key: 'name',
    width: 250,
    render: (row: any) => {
      return h(NSpace, { align: 'center' }, {
        default: () => [
          h('span', { class: 'text-base' }, row.name),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => copyToClipboard(row.name)
          }, {
            icon: () => h(NIcon, null, {
              default: () => h('svg', { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' }, [
                h('path', { d: 'M16 1H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V3c0-1.1-.9-2-2-2zm-8 12l-4-4 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 8z', fill: 'currentColor' })
              ])
            })
          })
        ]
      })
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: any) => {
      const statusMap: Record<string, { type: 'success' | 'error' | 'warning' | 'info'; text: string }> = {
        Running: { type: 'success', text: '运行中' },
        Pending: { type: 'warning', text: '等待中' },
        Failed: { type: 'error', text: '失败' },
        Succeeded: { type: 'info', text: '已完成' },
        Unknown: { type: 'info', text: '未知' },
      }
      const status = statusMap[row.status] || { type: 'info', text: row.status }
      return h(NTag, { type: status.type, bordered: false, size: 'small' }, { default: () => status.text })
    }
  },
  {
    title: '重启次数',
    key: 'restartCount',
    width: 100,
  },
  {
    title: '监控',
    key: 'monitoring',
    width: 80,
    render: () => {
      return h(NIcon, { size: 20, color: '#4285F4' }, {
        default: () => h('svg', { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' }, [
          h('circle', { cx: '12', cy: '12', r: '10', stroke: 'currentColor', 'stroke-width': '2' }),
          h('path', { d: 'M12 6v6l4 2', stroke: 'currentColor', 'stroke-width': '2', 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }),
        ])
      })
    }
  },
  {
    title: '所在节点',
    key: 'node',
    width: 150,
    render: (row: any) => row.nodeName || '-'
  },
  {
    title: 'Pod IP',
    key: 'podIp',
    width: 120,
    render: (row: any) => row.podIp || '-'
  },
  {
    title: '运行时间',
    key: 'age',
    width: 150,
    render: (row: any) => row.age || '-'
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    fixed: 'right' as const,
    render: (row: any) => {
      return h(NSpace, { size: 'small' }, {
        default: () => [
          h(NButton, {
            size: 'small',
            quaternary: true,
            onClick: () => viewPodDetail(row)
          }, { default: () => '详情' }),
          h(NButton, {
            size: 'small',
            quaternary: true,
            onClick: () => viewPodLogs(row)
          }, { default: () => '日志' }),
          h(NButton, {
            size: 'small',
            quaternary: true,
            onClick: () => openTerminal(row)
          }, { default: () => '终端' }),
          h(NButton, {
            size: 'small',
            type: 'error',
            quaternary: true,
            onClick: () => openDeletePodModal(row)
          }, { default: () => '删除' }),
        ]
      })
    }
  }
])

// Info cards data
const basicInfoCards = computed(() => [
  {
    title: '基本信息',
    items: [
      { label: '名称', value: workloadDetail.value?.name || '-' },
      { label: '命名空间', value: workloadDetail.value?.namespace || '-' },
      { label: '类型', value: typeDisplayNames[workloadType.value] || '-' },
      { label: '副本数', value: `${workloadDetail.value?.readyReplicas || 0}/${workloadDetail.value?.replicas || 0}` },
    ]
  },
  {
    title: '标签',
    items: [
      { label: '标签', value: formatLabels(workloadDetail.value?.labels) },
      { label: '选择器', value: formatSelectors(workloadDetail.value?.selector) },
    ]
  }
])

const statusInfoCards = computed(() => [
  {
    title: '状态信息',
    items: [
      { label: '创建时间', value: formatDate(workloadDetail.value?.createdAt) },
      { label: '运行时间', value: workloadDetail.value?.age || '-' },
      { label: '副本数', value: workloadDetail.value?.replicas || 0 },
      { label: '就绪副本', value: workloadDetail.value?.readyReplicas || 0 },
      { label: '更新副本数', value: workloadDetail.value?.updatedReplicas || 0 },
      { label: '可用副本', value: workloadDetail.value?.availableReplicas || 0 },
      { label: '不可用副本', value: workloadDetail.value?.unavailableReplicas || 0 },
    ]
  }
])

function formatLabels(labels: Record<string, string>): string {
  if (!labels || Object.keys(labels).length === 0) return '-'
  return Object.entries(labels).map(([k, v]) => `${k}=${v}`).join(', ')
}

function formatSelectors(selector: any): string {
  if (!selector) return '-'
  return JSON.stringify(selector)
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).then(() => {
    message.success('已复制到剪贴板')
  })
}

async function fetchWorkloadDetail() {
  if (!currentClusterId.value) return

  detailLoading.value = true
  try {
    const params = {
      clusterId: currentClusterId.value,
      namespace: workloadNamespace.value,
      name: workloadName.value
    }

    let res
    switch (workloadType.value) {
      case 'deployment':
        res = await k8sK8sDeploymentDetailPost(params)
        break
      case 'statefulset':
      case 'daemonset':
        // StatefulSet and DaemonSet don't have DetailPost, use list response
        // We'll get basic info from the list or set minimal info
        workloadDetail.value = {
          name: workloadName.value,
          namespace: workloadNamespace.value,
          replicas: 0,
          readyReplicas: 0,
          labels: {},
          selector: {},
          createdAt: new Date().toISOString()
        }
        return
      default:
        return
    }

    workloadDetail.value = res?.data?.data || null
  } catch (error: any) {
    message.error('获取工作负载详情失败: ' + error.message)
    // Set minimal info on error so page can still render
    workloadDetail.value = {
      name: workloadName.value,
      namespace: workloadNamespace.value,
      replicas: 0,
      readyReplicas: 0,
      labels: {},
      selector: {},
      createdAt: new Date().toISOString()
    }
  } finally {
    detailLoading.value = false
  }
}

async function fetchPods() {
  if (!currentClusterId.value) return

  loading.value = true
  try {
    let res
    if (workloadType.value === 'deployment') {
      res = await k8sK8sDeploymentPodsPost({
        clusterId: currentClusterId.value,
        namespace: workloadNamespace.value,
        name: workloadName.value
      })
      pods.value = (res as any)?.data?.data?.items || (res as any)?.data?.data || []
    } else {
      res = await k8sK8sPodListByOwnerPost({
        clusterId: currentClusterId.value,
        namespace: workloadNamespace.value,
        ownerType: workloadType.value === 'statefulset' ? 'StatefulSet' : 'DaemonSet',
        ownerName: workloadName.value,
        name: workloadName.value
      })
      pods.value = (res as any)?.data?.data?.items || (res as any)?.data?.data || []
    }
  } catch (error: any) {
    message.error('获取Pod列表失败: ' + error.message)
    pods.value = []
  } finally {
    loading.value = false
  }
}

async function viewYaml() {
  if (!currentClusterId.value) return

  yamlLoading.value = true
  try {
    const params = {
      clusterId: currentClusterId.value,
      namespace: workloadNamespace.value,
      name: workloadName.value
    }

    let res
    switch (workloadType.value) {
      case 'deployment':
        res = await k8sK8sDeploymentYamlPost(params)
        break
      case 'statefulset':
        res = await k8sK8sStatefulSetYamlPost(params)
        break
      case 'daemonset':
        res = await k8sK8sDaemonSetYamlPost(params)
        break
    }

    yamlContent.value = res?.data?.data?.yaml || ''
    showYamlModal.value = true
  } catch (error: any) {
    message.error('获取YAML失败: ' + error.message)
  } finally {
    yamlLoading.value = false
  }
}

async function openTerminal(row: any) {
  if (!currentClusterId.value) {
    message.warning('请先选择集群')
    return
  }

  try {
    // Construct WebSocket URL - use backend port (8000) not frontend port (3000)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const backendHost = window.location.hostname + ':8000'
    const container = row.containers?.[0]?.name || ''
    const wsUrl = `${protocol}//${backendHost}/api/v1/k8s/pod/terminal?clusterId=${currentClusterId.value}&namespace=${row.namespace}&pod=${row.name}&container=${container}&shell=/bin/sh`

    terminalWsUrl.value = wsUrl
    selectedPod.value = row
    showTerminalModal.value = true
  } catch (error: any) {
    message.error('获取终端连接失败: ' + error.message)
  }
}

async function viewPodDetail(row: any) {
  // Show Pod YAML instead of just a message
  currentPodForYaml.value = row
  podYamlLoading.value = true
  showPodYamlModal.value = true

  try {
    const res = await k8sK8sPodYamlGet({
      clusterId: currentClusterId.value!,
      namespace: row.namespace,
      name: row.name
    })
    podYamlContent.value = (res as any)?.data?.yaml || ''
  } catch (error: any) {
    message.error('获取Pod YAML失败: ' + error.message)
    podYamlContent.value = '获取失败: ' + error.message
  } finally {
    podYamlLoading.value = false
  }
}

async function viewPodLogs(row: any) {
  if (!currentClusterId.value) {
    message.warning('请先选择集群')
    return
  }

  currentPodForLogs.value = row
  logsLoading.value = true
  logsContent.value = ''
  showLogsModal.value = true

  try {
    const container = row.containers?.[0]?.name || ''
    const res = await k8sK8sPodLogsGet({
      clusterId: currentClusterId.value,
      namespace: row.namespace,
      name: row.name,
      container: container,
      tailLines: 100
    })

    // Extract logs from response - the structure is { code: 200, data: { logs: "..." } }
    const responseData = (res as any)?.data
    if (responseData?.logs !== undefined) {
      logsContent.value = responseData.logs
    } else {
      logsContent.value = '暂无日志'
    }
  } catch (error: any) {
    message.error('获取日志失败: ' + error.message)
    logsContent.value = '获取日志失败: ' + error.message
  } finally {
    logsLoading.value = false
  }
}

function openDeletePodModal(row: any) {
  deleteTarget.value = { type: 'pod', name: row.name }
  showDeleteModal.value = true
}

async function confirmDeletePod() {
  if (!deleteTarget.value || !currentClusterId.value) return

  try {
    await k8sK8sPodDeletePost({
      clusterId: currentClusterId.value,
      namespace: workloadNamespace.value,
      name: deleteTarget.value.name
    })

    message.success('删除Pod成功')
    showDeleteModal.value = false
    fetchPods()
    fetchWorkloadDetail()
  } catch (error: any) {
    message.error('删除Pod失败: ' + error.message)
  }
}

function goBack() {
  router.back()
}

function refresh() {
  fetchWorkloadDetail()
  fetchPods()
}

function openDeleteWorkloadModal() {
  deleteTarget.value = { type: workloadType.value, name: workloadName.value }
  showDeleteModal.value = true
}

async function confirmDeleteWorkload() {
  if (!deleteTarget.value || !currentClusterId.value) return

  try {
    message.info('删除工作负载功能开发中...')
    // TODO: Implement workload deletion
    showDeleteModal.value = false
  } catch (error: any) {
    message.error('删除工作负载失败: ' + error.message)
  }
}

onMounted(() => {
  if (currentClusterId.value) {
    fetchWorkloadDetail()
    fetchPods()
  }
})
</script>

<template>
  <div class="workload-detail-page">
    <!-- Header -->
    <div class="detail-header">
      <NSpace align="center">
        <NButton quaternary @click="goBack">
          <template #icon>
            <NIcon>
              <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M20 11H7.83l5.59-5.59L12 5l6.41 6.41L20 11z" fill="currentColor" />
              </svg>
            </NIcon>
          </template>
          返回
        </NButton>
        <h2 class="page-title">
          {{ typeDisplayNames[workloadType] }}详情 - {{ workloadName }}
        </h2>
        <NButton type="success" @click="refresh">
          <template #icon>
            <NIcon>
              <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                  d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c2.66 0 4.87 1.69 5.65 4zM12 14c-3.31 0-6-2.69-6-6s2.69-6 6-6 6 2.69 6 6-2.69 6-6-6z"
                  fill="currentColor"
                />
              </svg>
            </NIcon>
          </template>
          刷新
        </NButton>
        <NButton @click="viewYaml">
          <template #icon>
            <NIcon>
              <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                  d="M9.4 16.6L4.8 12l4.6-4.6L8 2l14 14-6 6-6.6z"
                  fill="currentColor"
                />
              </svg>
            </NIcon>
          </template>
          查看YAML
        </NButton>
        <NButton type="error" @click="openDeleteWorkloadModal">
          <template #icon>
            <NIcon>
              <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                  d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2H6c-1.1 0-2 .9-2 2v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14M8 13h8v-2H8v2z"
                  fill="currentColor"
                />
              </svg>
            </NIcon>
          </template>
          删除
        </NButton>
      </NSpace>
    </div>

    <!-- Info Cards -->
    <div class="info-cards" v-if="workloadDetail">
      <div class="info-card" v-for="(card, index) in [...basicInfoCards, ...statusInfoCards]" :key="index">
        <h3 class="card-title">{{ card.title }}</h3>
        <div class="card-content">
          <div class="info-item" v-for="(item, idx) in card.items" :key="idx">
            <span class="info-label">{{ item.label }}:</span>
            <span class="info-value">{{ item.value }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <NCard class="tabs-card">
      <NTabs v-model:value="activeTab" type="line">
        <NTabPane name="pods" tab="容器组">
          <NSpin :show="loading" size="large">
            <NDataTable
              :columns="podColumns"
              :data="pods"
              :loading="loading"
              :bordered="false"
              :row-key="(row: any) => row.name"
            />
          </NSpin>
        </NTabPane>
        <NTabPane name="history" tab="历史版本">
          <div class="empty-content">
            <p>历史版本功能开发中...</p>
          </div>
        </NTabPane>
        <NTabPane name="events" tab="事件">
          <div class="empty-content">
            <p>事件功能开发中...</p>
          </div>
        </NTabPane>
        <NTabPane name="logs" tab="日志">
          <div class="empty-content">
            <p>日志功能开发中...</p>
          </div>
        </NTabPane>
        <NTabPane name="scaling" tab="容器伸缩">
          <div class="empty-content">
            <p>容器伸缩功能开发中...</p>
          </div>
        </NTabPane>
      </NTabs>
    </NCard>

    <!-- YAML Modal -->
    <NModal
      v-model:show="showYamlModal"
      preset="card"
      title="YAML 配置"
      style="width: 800px"
    >
      <NSpin :show="yamlLoading">
        <pre class="yaml-content">{{ yamlContent }}</pre>
      </NSpin>
    </NModal>

    <!-- Terminal Modal -->
    <NModal
      v-model:show="showTerminalModal"
      preset="card"
      :title="`终端 - ${selectedPod?.name || ''}`"
      style="width: 900px"
    >
      <K8sTerminal v-if="terminalWsUrl" :ws-url="terminalWsUrl" />
      <div v-else class="terminal-loading">
        <NSpin size="large">正在连接终端...</NSpin>
      </div>
    </NModal>

    <!-- Logs Modal -->
    <NModal
      v-model:show="showLogsModal"
      preset="card"
      :title="`日志 - ${currentPodForLogs?.name || ''}`"
      style="width: 900px"
    >
      <div style="max-height: 500px; overflow-y: auto; background: #1e1e1e; padding: 12px; border-radius: 4px">
        <pre v-if="logsLoading" style="color: #888">加载中...</pre>
        <pre v-else style="color: #d4d4d4; font-family: 'Courier New', monospace; white-space: pre-wrap; word-wrap: break-word">{{ logsContent || '暂无日志' }}</pre>
      </div>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showLogsModal = false">关闭</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- Pod YAML Modal -->
    <NModal
      v-model:show="showPodYamlModal"
      preset="card"
      :title="`Pod YAML - ${currentPodForYaml?.name || ''}`"
      style="width: 900px"
    >
      <NSpin :show="podYamlLoading">
        <pre class="yaml-content">{{ podYamlContent }}</pre>
      </NSpin>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showPodYamlModal = false">关闭</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- Delete Confirm Modal -->
    <NModal
      v-model:show="showDeleteModal"
      preset="card"
      title="确认删除"
      style="width: 500px"
    >
      <p>
        确定要删除 <strong>{{ deleteTarget?.name }}</strong> 吗？此操作不可逆。
      </p>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showDeleteModal = false">取消</NButton>
          <NButton type="error" @click="deleteTarget?.type === 'pod' ? confirmDeletePod() : confirmDeleteWorkload()">确认删除</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.workload-detail-page {
  padding: var(--spacing-lg);
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-xl);
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.info-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-xl);
}

.info-card {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 var(--spacing-md) 0;
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.info-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-xs) 0;
}

.info-label {
  font-size: 12px;
  color: var(--text-secondary);
  min-width: 100px;
}

.info-value {
  font-size: 14px;
  color: var(--text-primary);
  word-break: break-all;
}

.tabs-card {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}

.yaml-content {
  background: var(--text-primary);
  color: var(--content-bg);
  padding: var(--spacing-md);
  border-radius: var(--radius-sm);
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  overflow-x: auto;
  max-height: 500px;
}

.empty-content {
  padding: var(--spacing-3xl) 0;
  text-align: center;
  color: var(--text-secondary);
}

.terminal-loading {
  height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
