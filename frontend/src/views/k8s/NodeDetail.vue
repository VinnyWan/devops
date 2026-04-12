<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <div>
        <h3>{{ detail?.name || '节点详情' }}</h3>
        <div class="page-subtitle">{{ clusterName }} / {{ nodeName }}</div>
      </div>
      <el-button @click="router.back()">返回</el-button>
    </div>

    <el-alert v-if="errorText" :title="errorText" type="error" show-icon :closable="false" />

    <template v-else-if="detail">
      <el-row :gutter="16" class="summary-row">
        <el-col :xs="24" :md="8">
          <el-card shadow="never" class="summary-card">
            <template #header>基础信息</template>
            <div class="summary-list">
              <div class="summary-item">
                <span class="label">状态</span>
                <StatusTag :status="detail.status || '-'" />
              </div>
              <div class="summary-item">
                <span class="label">角色</span>
                <span class="value">{{ formatValue(detail.role) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">内部 IP</span>
                <span class="value">{{ formatValue(detail.ip) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">外部 IP</span>
                <span class="value">{{ formatValue(detail.externalIP) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">K8s 版本</span>
                <span class="value">{{ formatValue(detail.k8sVersion || detail.kubeletVersion) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">运行时长</span>
                <span class="value">{{ formatValue(detail.age) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="8">
          <el-card shadow="never" class="summary-card">
            <template #header>资源使用</template>
            <div class="summary-list">
              <div class="summary-item">
                <span class="label">CPU</span>
                <span class="value">{{ formatUsage(detail.cpuUsage, detail.cpuCapacity) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">内存</span>
                <span class="value">{{ formatUsage(detail.memoryUsage, detail.memoryCapacity) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">CPU Requests</span>
                <span class="value">{{ formatAllocated(detail.allocatedResources?.cpuRequests, detail.allocatedResources?.cpuRequestsPercentage) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">CPU Limits</span>
                <span class="value">{{ formatAllocated(detail.allocatedResources?.cpuLimits, detail.allocatedResources?.cpuLimitsPercentage) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">内存 Requests</span>
                <span class="value">{{ formatAllocated(detail.allocatedResources?.memoryRequests, detail.allocatedResources?.memoryRequestsPercentage) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">内存 Limits</span>
                <span class="value">{{ formatAllocated(detail.allocatedResources?.memoryLimits, detail.allocatedResources?.memoryLimitsPercentage) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="8">
          <el-card shadow="never" class="summary-card">
            <template #header>快速指标</template>
            <div class="summary-list">
              <div class="summary-item">
                <span class="label">Pod 数</span>
                <span class="value">{{ podItems.length }} / {{ formatValue(detail.podCapacity) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">是否可调度</span>
                <span class="value">{{ detail.unschedulable ? '否' : '是' }}</span>
              </div>
              <div class="summary-item">
                <span class="label">Pod CIDR</span>
                <span class="value">{{ formatValue(detail.podCIDR) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">Provider ID</span>
                <span class="value wrap">{{ formatValue(detail.providerID) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">Lease Renew</span>
                <span class="value">{{ formatTime(detail.lease?.renewTime) }}</span>
              </div>
              <div class="summary-item">
                <span class="label">创建时间</span>
                <span class="value">{{ formatTime(detail.createdAt || detail.creationTimestamp) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="标签与污点" name="labels">
          <el-row :gutter="16">
            <el-col :xs="24" :lg="12">
              <el-card shadow="never">
                <template #header>Labels</template>
                <el-empty v-if="!objectEntries(detail.labels).length" description="暂无 Labels" />
                <el-table v-else :data="objectEntries(detail.labels)" stripe>
                  <el-table-column prop="key" label="Key" min-width="220" />
                  <el-table-column prop="value" label="Value" min-width="220" />
                </el-table>
              </el-card>
            </el-col>
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="detail-card">
                <template #header>Taints</template>
                <el-empty v-if="!detail.taints?.length" description="暂无 Taints" />
                <el-table v-else :data="detail.taints" stripe>
                  <el-table-column label="Key" min-width="180">
                    <template #default="{ row }">{{ formatValue(row.key) }}</template>
                  </el-table-column>
                  <el-table-column label="Value" min-width="160">
                    <template #default="{ row }">{{ formatValue(row.value) }}</template>
                  </el-table-column>
                  <el-table-column label="Effect" min-width="140">
                    <template #default="{ row }">{{ formatValue(row.effect) }}</template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>

          <el-row :gutter="16" class="tab-row">
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="detail-card">
                <template #header>Annotations</template>
                <el-empty v-if="!objectEntries(detail.annotations).length" description="暂无 Annotations" />
                <el-table v-else :data="objectEntries(detail.annotations)" stripe>
                  <el-table-column prop="key" label="Key" min-width="220" />
                  <el-table-column prop="value" label="Value" min-width="220" />
                </el-table>
              </el-card>
            </el-col>
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="detail-card">
                <template #header>Conditions</template>
                <el-empty v-if="!detail.conditions?.length" description="暂无 Conditions" />
                <el-table v-else :data="detail.conditions" stripe>
                  <el-table-column label="类型" min-width="160">
                    <template #default="{ row }">{{ formatValue(row.type) }}</template>
                  </el-table-column>
                  <el-table-column label="状态" width="120">
                    <template #default="{ row }">{{ formatValue(row.status) }}</template>
                  </el-table-column>
                  <el-table-column label="原因" min-width="160">
                    <template #default="{ row }">{{ formatValue(row.reason) }}</template>
                  </el-table-column>
                  <el-table-column label="最后心跳" min-width="180">
                    <template #default="{ row }">{{ formatTime(row.lastHeartbeatTime) }}</template>
                  </el-table-column>
                </el-table>
              </el-card>
            </el-col>
          </el-row>
        </el-tab-pane>

        <el-tab-pane label="系统信息" name="system">
          <el-row :gutter="16">
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="detail-card">
                <template #header>系统信息</template>
                <el-descriptions :column="1" border>
                  <el-descriptions-item label="OS Image">{{ formatValue(detail.systemInfo?.osImage || detail.osImage) }}</el-descriptions-item>
                  <el-descriptions-item label="Kernel Version">{{ formatValue(detail.systemInfo?.kernelVersion || detail.kernelVersion) }}</el-descriptions-item>
                  <el-descriptions-item label="Kubelet Version">{{ formatValue(detail.systemInfo?.kubeletVersion || detail.kubeletVersion) }}</el-descriptions-item>
                  <el-descriptions-item label="Container Runtime">{{ formatValue(detail.systemInfo?.containerRuntimeVersion) }}</el-descriptions-item>
                  <el-descriptions-item label="Architecture">{{ formatValue(detail.systemInfo?.architecture) }}</el-descriptions-item>
                  <el-descriptions-item label="Machine ID">{{ formatValue(detail.systemInfo?.machineID) }}</el-descriptions-item>
                </el-descriptions>
              </el-card>
            </el-col>
            <el-col :xs="24" :lg="12">
              <el-card shadow="never" class="detail-card">
                <template #header>容量与地址</template>
                <el-descriptions :column="1" border>
                  <el-descriptions-item label="Capacity CPU">{{ formatValue(detail.capacity?.cpu) }}</el-descriptions-item>
                  <el-descriptions-item label="Capacity Memory">{{ formatValue(detail.capacity?.memory) }}</el-descriptions-item>
                  <el-descriptions-item label="Allocatable CPU">{{ formatValue(detail.allocatable?.cpu) }}</el-descriptions-item>
                  <el-descriptions-item label="Allocatable Memory">{{ formatValue(detail.allocatable?.memory) }}</el-descriptions-item>
                  <el-descriptions-item label="Pod CIDRs">{{ formatArray(detail.podCIDRs) }}</el-descriptions-item>
                  <el-descriptions-item label="地址">{{ formatAddresses(detail.addresses) }}</el-descriptions-item>
                  <el-descriptions-item label="Lease Acquire">{{ formatTime(detail.lease?.acquireTime) }}</el-descriptions-item>
                  <el-descriptions-item label="Lease Holder">{{ formatValue(detail.lease?.holderIdentity) }}</el-descriptions-item>
                </el-descriptions>
              </el-card>
            </el-col>
          </el-row>
        </el-tab-pane>

        <el-tab-pane :label="`运行的 Pods (${podItems.length})`" name="pods">
          <el-empty v-if="!podItems.length" description="暂无运行中的 Pods" />
          <template v-else>
            <el-table :data="pagedPods" stripe>
              <el-table-column prop="name" label="Pod 名称" min-width="220" />
              <el-table-column prop="namespace" label="命名空间" min-width="140" />
              <el-table-column prop="status" label="状态" width="120">
                <template #default="{ row }">
                  <StatusTag :status="row.status || '-'" />
                </template>
              </el-table-column>
              <el-table-column prop="cpuRequest" label="CPU 请求" width="110" />
              <el-table-column prop="cpuLimit" label="CPU 限制" width="110" />
              <el-table-column prop="memoryRequest" label="内存请求" width="120" />
              <el-table-column prop="memoryLimit" label="内存限制" width="120" />
              <el-table-column prop="restartCount" label="重启次数" width="100" />
              <el-table-column prop="age" label="运行时长" width="120" />
            </el-table>

            <el-pagination
              v-model:current-page="podPage"
              v-model:page-size="podPageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="podItems.length"
              layout="total, sizes, prev, pager, next"
              class="pagination"
            />
          </template>
        </el-tab-pane>

        <el-tab-pane label="事件" name="events">
          <el-alert
            v-if="eventErrorText"
            :title="eventErrorText"
            type="warning"
            show-icon
            :closable="false"
            style="margin-bottom: 16px"
          />

          <el-table v-loading="eventLoading" :data="events" stripe>
            <el-table-column prop="time" label="时间" min-width="180">
              <template #default="{ row }">{{ formatTime(row.time) }}</template>
            </el-table-column>
            <el-table-column prop="type" label="类型" width="100">
              <template #default="{ row }">
                <el-tag :type="row.type === 'Normal' ? 'success' : 'warning'" size="small">
                  {{ row.type || '-' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="reason" label="原因" min-width="140" />
            <el-table-column prop="object" label="对象" min-width="220" />
            <el-table-column prop="message" label="消息" min-width="320" show-overflow-tooltip />
          </el-table>

          <el-empty v-if="!eventLoading && !events.length && !eventErrorText" description="暂无事件" />
        </el-tab-pane>
      </el-tabs>
    </template>

    <el-empty v-else description="暂无节点详情数据" />
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import StatusTag from '@/components/K8s/StatusTag.vue'
import { getNodeDetail, getNodeEvents } from '@/api/node'

const route = useRoute()
const router = useRouter()

const clusterName = computed(() => String(route.params.clusterName || ''))
const nodeName = computed(() => String(route.params.nodeName || ''))

const loading = ref(false)
const eventLoading = ref(false)
const detail = ref(null)
const events = ref([])
const activeTab = ref('labels')
const errorText = ref('')
const eventErrorText = ref('')
const podPage = ref(1)
const podPageSize = ref(10)
const eventLoaded = ref(false)

const podItems = computed(() => detail.value?.pods?.items || [])
const pagedPods = computed(() => {
  const start = (podPage.value - 1) * podPageSize.value
  return podItems.value.slice(start, start + podPageSize.value)
})

const objectEntries = (value) => Object.entries(value || {}).map(([key, val]) => ({ key, value: val }))

const formatValue = (value) => {
  if (value === null || value === undefined) return '-'
  const text = String(value).trim()
  return text ? text : '-'
}

const formatUsage = (used, total) => `${formatValue(used)} / ${formatValue(total)}`

const formatAllocated = (value, percentage) => {
  const base = formatValue(value)
  const rate = formatValue(percentage)
  return rate === '-' ? base : `${base} (${rate})`
}

const formatArray = (value) => (Array.isArray(value) && value.length ? value.join(', ') : '-')

const formatAddresses = (value) => {
  if (!Array.isArray(value) || !value.length) return '-'
  return value.map((item) => `${formatValue(item.type)}: ${formatValue(item.address)}`).join('；')
}

const formatTime = (value) => {
  if (!value || value === '-' || value === '0001-01-01T00:00:00Z') return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return formatValue(value)
  return date.toLocaleString('zh-CN')
}

const resetPageState = () => {
  detail.value = null
  events.value = []
  errorText.value = ''
  eventErrorText.value = ''
  activeTab.value = 'labels'
  podPage.value = 1
  eventLoaded.value = false
}

const fetchDetail = async () => {
  loading.value = true
  errorText.value = ''
  detail.value = null
  try {
    const res = await getNodeDetail({ clusterName: clusterName.value, name: nodeName.value })
    detail.value = res.data || null
  } catch (error) {
    errorText.value = error.response?.data?.message || '节点详情加载失败'
  } finally {
    loading.value = false
  }
}

const fetchEvents = async () => {
  if (eventLoading.value || eventLoaded.value) return
  eventLoading.value = true
  eventErrorText.value = ''
  try {
    const res = await getNodeEvents({ clusterName: clusterName.value, name: nodeName.value })
    events.value = res.data || []
    eventLoaded.value = true
  } catch (error) {
    eventErrorText.value = error.response?.data?.message || '节点事件加载失败'
    ElMessage.error(eventErrorText.value)
  } finally {
    eventLoading.value = false
  }
}

watch([clusterName, nodeName], () => {
  resetPageState()
  fetchDetail()
}, { immediate: true })

watch(activeTab, (tab) => {
  if (tab === 'events') {
    fetchEvents()
  }
})

watch(podPageSize, () => {
  podPage.value = 1
})
</script>

<style scoped>
.page-container {
  background: #fff;
  border-radius: 4px;
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

.page-subtitle {
  margin-top: 6px;
  color: #909399;
  font-size: 13px;
}

.summary-row,
.tab-row {
  margin-bottom: 16px;
}

.summary-card,
.detail-card {
  height: 100%;
}

.summary-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.summary-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  line-height: 1.5;
}

.label {
  color: #606266;
  flex-shrink: 0;
}

.value {
  color: #303133;
  text-align: right;
}

.wrap {
  word-break: break-all;
}

.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>
