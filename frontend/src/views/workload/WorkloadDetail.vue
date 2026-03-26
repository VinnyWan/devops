<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  NCard,
  NSpace,
  NDataTable,
  NButton,
  NTag,
  NModal,
  useMessage,
  NIcon,
  NSpin,
} from 'naive-ui'
import K8sTerminal from '@/components/K8sTerminal.vue'
import YamlTerminalModal from '@/components/YamlTerminalModal.vue'
import { useCluster } from '@/composables/useCluster'
import {
  k8sK8sDeploymentDetailPost,
  k8sK8sDeploymentPodsPost,
  k8sK8sDeploymentYamlPost,
  k8sK8sPodLogsGet,
  k8sK8sPodDetectShellGet,
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
const terminalDetectingShell = ref(false)

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

// Info item type
interface InfoItem {
  label?: string
  value?: string | number | unknown | Record<string, string>
  divider?: boolean
  isLabel?: boolean
  isLabelTags?: boolean
}
const podColumns = computed(() => [
  {
    title: 'Pod名称',
    key: 'name',
    width: 200,
    render: (row: any) => {
      return h('div', { class: 'pod-name-cell' }, [
        h('span', { class: 'pod-name-text' }, row.name),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          class: 'pod-copy-btn',
          onClick: () => copyToClipboard(row.name)
        }, {
          icon: () => h(NIcon, { size: 14 }, {
            default: () => h('svg', { viewBox: '0 0 24 24', fill: 'none', xmlns: 'http://www.w3.org/2000/svg' }, [
              h('path', { d: 'M16 1H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V3c0-1.1-.9-2-2-2zm-8 12l-4-4 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 8z', fill: 'currentColor' })
            ])
          })
        })
      ])
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 90,
    render: (row: any) => {
      const statusMap: Record<string, { type: 'success' | 'error' | 'warning' | 'info'; text: string }> = {
        Running: { type: 'success', text: '运行中' },
        Starting: { type: 'info', text: '启动中' },
        Pending: { type: 'warning', text: '等待中' },
        Failed: { type: 'error', text: '失败' },
        Error: { type: 'error', text: '错误' },
        Succeeded: { type: 'success', text: '已完成' },
        Unknown: { type: 'info', text: '未知' },
        Terminating: { type: 'warning', text: '终止中' },
        CrashLoopBackOff: { type: 'error', text: '崩溃重启' },
        ImagePullError: { type: 'error', text: '镜像拉取失败' },
      }
      const status = statusMap[row.status] || { type: 'info' as const, text: row.status }
      return h(NTag, { type: status.type, bordered: false, size: 'small' }, { default: () => status.text })
    }
  },
  {
    title: '重启次数',
    key: 'restartCount',
    width: 85,
  },
  {
    title: '监控',
    key: 'monitoring',
    width: 60,
    render: () => {
      return h(NIcon, { size: 18, color: '#4285F4' }, {
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
    width: 120,
    render: (row: any) => row.node || '-'
  },
  {
    title: 'Pod IP',
    key: 'ip',
    width: 100,
    render: (row: any) => row.ip || '-'
  },
  {
    title: '运行时间',
    key: 'age',
    width: 120,
    render: (row: any) => row.age || '-'
  },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    fixed: 'right' as const,
    render: (row: any) => {
      return h(NSpace, { size: 'small' }, {
        default: () => [
          h(NButton, {
            size: 'small',
            quaternary: true,
            onClick: () => viewPodYaml(row)
          }, { default: () => '查看YAML' }),
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
const basicInfoCards = computed(() => {
  const labels = workloadDetail.value?.labels || {}

  return [
    {
      title: '基本信息',
      items: [
        { label: '名称', value: workloadDetail.value?.name || '-' },
        { label: '命名空间', value: workloadDetail.value?.namespace || '-' },
        { label: '类型', value: typeDisplayNames[workloadType.value] || '-' },
        { label: '副本数', value: `${workloadDetail.value?.readyReplicas || 0}/${workloadDetail.value?.replicas || 0}` },
        { divider: true },
        { label: '标签', value: labels, isLabelTags: true } as InfoItem & { isLabelTags?: boolean; value: Record<string, string> },
      ] as InfoItem[]
    }
  ]
})

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
    ] as InfoItem[]
  }
])

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).then(() => {
    message.success('已复制到剪贴板')
  })
}

function copyLabel(key: string, value: string) {
  const labelStr = `${key}=${value}`
  copyToClipboard(labelStr)
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

  terminalDetectingShell.value = true
  showTerminalModal.value = true
  selectedPod.value = row
  terminalWsUrl.value = ''

  try {
    // 先检测容器支持的shell
    const container = row.containers?.[0]?.name || ''
    const detectRes = await k8sK8sPodDetectShellGet({
      clusterId: currentClusterId.value,
      namespace: row.namespace,
      pod: row.name,
      container: container
    })

    // 响应结构是 { code: 200, data: { availableShells: [...], recommendedShell: "..." } }
    const detectData = (detectRes as any)?.data?.data
    const recommendedShell = detectData?.recommendedShell || 'bash'

    console.log('检测到的Shell:', recommendedShell, '可用Shell:', detectData?.availableShells)

    // Construct WebSocket URL - use detected shell
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const backendHost = window.location.hostname + ':8000'
    const wsUrl = `${protocol}//${backendHost}/api/v1/k8s/pod/terminal?clusterId=${currentClusterId.value}&namespace=${row.namespace}&pod=${row.name}&container=${container}&shell=${recommendedShell}`

    terminalWsUrl.value = wsUrl
  } catch (error: any) {
    console.error('检测Shell失败:', error)
    message.warning('检测Shell失败，使用默认bash')
    // 即使检测失败，也使用默认bash打开终端
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const backendHost = window.location.hostname + ':8000'
    const container = row.containers?.[0]?.name || ''
    const wsUrl = `${protocol}//${backendHost}/api/v1/k8s/pod/terminal?clusterId=${currentClusterId.value}&namespace=${row.namespace}&pod=${row.name}&container=${container}&shell=bash`
    terminalWsUrl.value = wsUrl
  } finally {
    terminalDetectingShell.value = false
  }
}

async function viewPodYaml(row: any) {
  // Show Pod YAML
  currentPodForYaml.value = row
  podYamlLoading.value = true
  showPodYamlModal.value = true

  try {
    const res = await k8sK8sPodYamlGet({
      clusterId: currentClusterId.value!,
      namespace: row.namespace,
      name: row.name
    })
    // Response structure: { code: 200, data: { yaml: "..." } }
    podYamlContent.value = (res as any)?.data?.data?.yaml || ''
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
  // 导航回工作负载列表页面，而不是使用router.back()避免返回到登录页
  router.push('/workload')
}

function refresh() {
  fetchWorkloadDetail()
  fetchPods()
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
                  d="M14 2H6C4.9 2 4 2.9 4 4V20C4 21.1 4.89 22 5.99 22H18C19.1 22 20 21.1 20 20V8L14 2ZM18 20H6V4H13V9H18V20ZM9 13V19H7V13H9ZM15 15V19H17V15H15ZM11 11V19H13V11H11Z"
                  fill="currentColor"
                />
              </svg>
            </NIcon>
          </template>
          查看YAML
        </NButton>
      </NSpace>
    </div>

    <!-- Info Cards -->
    <div class="info-cards" v-if="workloadDetail">
      <div class="info-card" v-for="(card, index) in [...basicInfoCards, ...statusInfoCards]" :key="index">
        <h3 class="card-title">{{ card.title }}</h3>
        <div class="card-content">
          <template v-for="(item, idx) in card.items" :key="idx">
            <div v-if="item.divider" class="info-divider"></div>
            <div v-else-if="item.isLabelTags" class="info-item info-labels">
              <span class="info-label">{{ item.label }}:</span>
              <div class="label-tags">
                <NTag
                  v-for="(val, key) in Object.entries(item.value as Record<string, string>)"
                  :key="key"
                  type="info"
                  :bordered="false"
                  size="small"
                  class="label-tag"
                  @click="copyLabel(val[0], val[1])"
                >
                  {{ val[0] }}={{ val[1] }}
                </NTag>
                <span v-if="Object.keys(item.value as Record<string, string>).length === 0" class="info-value">-</span>
              </div>
            </div>
            <div v-else class="info-item">
              <span class="info-label">{{ item.label }}:</span>
              <span class="info-value">{{ item.value }}</span>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- Pod List -->
    <NCard class="pods-card">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <h3 style="margin: 0">容器组 (Pod)</h3>
          <NButton size="small" @click="refresh">
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
        </div>
      </template>
      <NSpin :show="loading" size="large">
        <NDataTable
          :columns="podColumns"
          :data="pods"
          :loading="loading"
          :bordered="false"
          :row-key="(row: any) => row.name"
        />
      </NSpin>
    </NCard>

    <!-- YAML Modal -->
    <YamlTerminalModal
      v-model:show="showYamlModal"
      :title="`${typeDisplayNames[workloadType]} / ${workloadName}`"
      :content="yamlContent"
      :loading="yamlLoading"
      readonly
    />

    <!-- Terminal Modal (Full Screen) -->
    <NModal
      v-model:show="showTerminalModal"
      :style="{ width: '100vw', height: '100vh' }"
      preset="card"
      :title="`终端 - ${selectedPod?.name || ''}`"
      :mask-closable="false"
      :closable="true"
      :auto-focus="false"
    >
      <div style="width: 100%; height: calc(100vh - 120px)">
        <NSpin v-if="terminalDetectingShell" size="large" style="height: 100%; display: flex; align-items: center; justify-content: center">
          <template #description>
            正在检测容器可用Shell...
          </template>
        </NSpin>
        <K8sTerminal v-else-if="terminalWsUrl" :ws-url="terminalWsUrl" />
        <div v-else class="terminal-loading">
          <NSpin size="large">正在连接终端...</NSpin>
        </div>
      </div>
      <template #footer>
        <NSpace justify="end">
          <NButton type="error" @click="showTerminalModal = false">关闭终端</NButton>
        </NSpace>
      </template>
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
    <YamlTerminalModal
      v-model:show="showPodYamlModal"
      :title="`Pod / ${currentPodForYaml?.name || ''}`"
      :content="podYamlContent"
      :loading="podYamlLoading"
      readonly
    />

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
          <NButton type="error" @click="confirmDeletePod">确认删除</NButton>
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

.info-item.info-labels {
  align-items: flex-start;
}

.info-labels .label-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
}

.info-labels .label-tag {
  cursor: pointer;
  font-family: 'Consolas', 'Monaco', monospace;
  transition: all 0.2s;
  background: #3b82f6;
  color: white;
  border: none;
}

.info-labels .label-tag:hover {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.4);
  background: #2563eb;
}

.info-divider {
  height: 1px;
  background: var(--border-color);
  margin: var(--spacing-md) 0;
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

.pods-card {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
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

.pod-name-cell {
  display: flex;
  align-items: center;
  gap: 4px;
}

.pod-name-text {
  font-size: 13px;
  font-family: 'Consolas', 'Monaco', monospace;
  word-break: break-all;
}

.pod-copy-btn {
  flex-shrink: 0;
  margin-left: 4px;
}
</style>
