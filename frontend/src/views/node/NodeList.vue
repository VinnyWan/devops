<script setup lang="ts">
import { ref, onMounted, h, computed } from 'vue'
import {
  NSpace,
  NDataTable,
  NButton,
  useMessage,
  NPagination,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NCheckbox,
  NInputNumber,
  NTag,
  NPopconfirm,
  NDescriptions,
  NDescriptionsItem,
} from 'naive-ui'
import ClusterSelector from '@/components/ClusterSelector.vue'
import { useCluster } from '@/composables/useCluster'
import {
  k8sK8sNodesPost,
  k8sK8sNodeDetailPost,
  k8sK8sNodeCordonPost,
  k8sK8sNodeDrainPost,
  k8sK8sNodeLabelsPost,
  k8sK8sNodeTaintsPost,
} from '@/api/generated/k8s-node.api'

const message = useMessage()
const { currentClusterId } = useCluster()
const loading = ref(false)
const nodes = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

// Filter state
const filters = ref({
  name: '',
  status: '',
  role: '',
})

const statusOptions = [
  { label: '全部', value: '' },
  { label: '运行中', value: 'Ready' },
  { label: '异常', value: 'NotReady' },
  { label: '调度中', value: 'Scheduling' },
]

const roleOptions = [
  { label: '全部', value: '' },
  { label: 'Master', value: 'master' },
  { label: 'Worker', value: 'worker' },
  { label: 'Control-plane', value: 'control-plane' },
]

// 详情弹窗
const showDetailModal = ref(false)
const nodeDetail = ref<any>(null)
const detailLoading = ref(false)
const podPage = ref(1)
const podPageSize = ref(10)

// 污点弹窗
const showTaintModal = ref(false)
const currentNode = ref<any>(null)
const taints = ref<any[]>([])
const newTaint = ref({ key: '', value: '', effect: 'NoSchedule' })

// 标签弹窗
const showLabelModal = ref(false)
const labels = ref<Record<string, string>>({})
const newLabel = ref({ key: '', value: '' })

// 驱逐弹窗
const showDrainModal = ref(false)
const drainOptions = ref({
  force: false,
  ignoreDaemonSets: true,
  deleteLocalData: false,
  gracePeriodSeconds: 30,
})

const taintEffects = [
  { label: '禁止调度 (NoSchedule)', value: 'NoSchedule' },
  { label: '禁止执行 (NoExecute)', value: 'NoExecute' },
  { label: '尽量避免调度 (PreferNoSchedule)', value: 'PreferNoSchedule' },
]

// Filtered nodes
const filteredNodes = computed(() => {
  let result = nodes.value

  if (filters.value.name) {
    result = result.filter((node: any) =>
      node.name.toLowerCase().includes(filters.value.name.toLowerCase())
    )
  }

  if (filters.value.status) {
    result = result.filter((node: any) => node.status === filters.value.status)
  }

  if (filters.value.role) {
    result = result.filter((node: any) =>
      node.role?.toLowerCase().includes(filters.value.role.toLowerCase())
    )
  }

  return result
})

// Reset filters
function resetFilters() {
  filters.value = {
    name: '',
    status: '',
    role: '',
  }
}

const columns = [
  { title: '节点名称', key: 'name', width: 200 },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: any) => {
      const statusMap: Record<
        string,
        { type: 'success' | 'error' | 'info' | 'warning' | 'primary'; text: string }
      > = {
        Ready: { type: 'success', text: '运行中' },
        NotReady: { type: 'error', text: '异常' },
        Scheduling: { type: 'info', text: '调度中' },
      }
      const status = statusMap[row.status] || { type: 'info', text: row.status }
      return h(
        NTag,
        {
          type: status.type,
          bordered: false,
          size: 'small',
        },
        { default: () => status.text }
      )
    },
  },
  { title: 'IP地址', key: 'ip', width: 140 },
  { title: '角色', key: 'role', width: 120 },
  { title: 'CPU容量', key: 'cpuCapacity', width: 100 },
  { title: '内存容量', key: 'memoryCapacity', width: 120 },
  {
    title: '操作',
    key: 'actions',
    width: 320,
    fixed: 'right' as const,
    render: (row: any) => {
      return h(
        NSpace,
        { size: 'small' },
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                quaternary: true,
                onClick: () => openDetailModal(row),
              },
              { default: () => '详情' }
            ),
            h(
              NButton,
              {
                size: 'small',
                quaternary: true,
                onClick: () => openTaintModal(row),
              },
              { default: () => '污点' }
            ),
            h(
              NButton,
              {
                size: 'small',
                quaternary: true,
                onClick: () => openLabelModal(row),
              },
              { default: () => '标签' }
            ),
            h(
              NButton,
              {
                size: 'small',
                type: row.unschedulable ? 'success' : 'warning',
                onClick: () => toggleCordon(row),
              },
              { default: () => (row.unschedulable ? '恢复调度' : '禁止调度') }
            ),
            h(
              NPopconfirm,
              {
                onPositiveClick: () => openDrainModal(row),
              },
              {
                default: () => '确定要驱逐此节点上的所有Pod吗？',
                trigger: () =>
                  h(
                    NButton,
                    { size: 'small', type: 'error', quaternary: true },
                    { default: () => '驱逐' }
                  ),
              }
            ),
          ],
        }
      )
    },
  },
]

async function fetchNodes() {
  if (!currentClusterId.value) {
    message.warning('请先选择集群')
    return
  }

  loading.value = true
  try {
    const res = await k8sK8sNodesPost({
      clusterId: currentClusterId.value,
      page: page.value,
      pageSize: pageSize.value,
    })
    const data = res.data.data
    nodes.value = data?.items || []
    total.value = data?.total || 0
  } catch (error: any) {
    message.error(error.message || '获取节点列表失败')
  } finally {
    loading.value = false
  }
}

// 详情功能
async function openDetailModal(node: any) {
  currentNode.value = node
  showDetailModal.value = true
  detailLoading.value = true

  try {
    const res = await k8sK8sNodeDetailPost({
      clusterId: currentClusterId.value!,
      name: node.name,
    })
    nodeDetail.value = res.data.data
  } catch (error: any) {
    message.error('获取节点详情失败: ' + error.message)
  } finally {
    detailLoading.value = false
  }
}

function calculatePercentage(used: string, capacity: string): string {
  const usedNum = parseFloat(used) || 0
  const capNum = parseFloat(capacity) || 1
  return ((usedNum / capNum) * 100).toFixed(1) + '%'
}

// 污点功能
async function openTaintModal(node: any) {
  currentNode.value = node
  taints.value = node.taints || []
  showTaintModal.value = true
}

function addTaint() {
  if (!newTaint.value.key) {
    message.warning('请输入污点键名')
    return
  }
  taints.value.push({ ...newTaint.value })
  newTaint.value = { key: '', value: '', effect: 'NoSchedule' }
}

function removeTaint(index: number) {
  taints.value.splice(index, 1)
}

async function saveTaints() {
  try {
    await k8sK8sNodeTaintsPost({
      clusterId: currentClusterId.value!,
      name: currentNode.value.name,
      taints: taints.value,
    })
    message.success('更新污点成功')
    showTaintModal.value = false
    fetchNodes()
  } catch (error: any) {
    message.error('更新污点失败: ' + error.message)
  }
}

// 标签功能
async function openLabelModal(node: any) {
  currentNode.value = node
  labels.value = node.labels ? { ...node.labels } : {}
  showLabelModal.value = true
}

function addLabel() {
  if (!newLabel.value.key) {
    message.warning('请输入标签键')
    return
  }
  labels.value[newLabel.value.key] = newLabel.value.value
  newLabel.value = { key: '', value: '' }
}

function removeLabel(key: string) {
  delete labels.value[key]
}

async function saveLabels() {
  try {
    await k8sK8sNodeLabelsPost({
      clusterId: currentClusterId.value!,
      name: currentNode.value.name,
      labels: labels.value,
    })
    message.success('更新标签成功')
    showLabelModal.value = false
    fetchNodes()
  } catch (error: any) {
    message.error('更新标签失败: ' + error.message)
  }
}

// 禁止调度功能
async function toggleCordon(node: any) {
  try {
    await k8sK8sNodeCordonPost({
      clusterId: currentClusterId.value!,
      name: node.name,
      cordon: !node.unschedulable,
    })
    message.success(node.unschedulable ? '恢复调度成功' : '禁止调度成功')
    fetchNodes()
  } catch (error: any) {
    message.error('操作失败: ' + error.message)
  }
}

// 驱逐功能
function openDrainModal(node: any) {
  currentNode.value = node
  showDrainModal.value = true
}

async function drainNode() {
  try {
    await k8sK8sNodeDrainPost({
      clusterId: currentClusterId.value!,
      name: currentNode.value.name,
      ...drainOptions.value,
    })
    message.success('节点驱逐成功')
    showDrainModal.value = false
    fetchNodes()
  } catch (error: any) {
    message.error('节点驱逐失败: ' + error.message)
  }
}

function handlePageChange(newPage: number) {
  page.value = newPage
  fetchNodes()
}

function handlePageSizeChange(newPageSize: number) {
  pageSize.value = newPageSize
  page.value = 1
  fetchNodes()
}

onMounted(() => {
  if (currentClusterId.value) {
    fetchNodes()
  }
})
</script>

<template>
  <div class="node-list-page">
    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="filter-item">
        <span class="filter-label">节点名称</span>
        <n-input
          v-model:value="filters.name"
          placeholder="请输入节点名称"
          clearable
          style="width: 200px"
        />
      </div>

      <div class="filter-item">
        <span class="filter-label">状态</span>
        <n-select
          v-model:value="filters.status"
          :options="statusOptions"
          placeholder="请选择状态"
          clearable
          style="width: 140px"
        />
      </div>

      <div class="filter-item">
        <span class="filter-label">角色</span>
        <n-select
          v-model:value="filters.role"
          :options="roleOptions"
          placeholder="请选择角色"
          clearable
          style="width: 140px"
        />
      </div>

      <div class="filter-actions">
        <n-button type="primary" @click="fetchNodes">搜索</n-button>
        <n-button @click="resetFilters">重置</n-button>
      </div>

      <div class="filter-extra">
        <ClusterSelector @update:value="fetchNodes" />
        <n-button type="info" @click="() => message.info('监控仪表板功能开发中')">
          监控仪表板
        </n-button>
      </div>
    </div>

    <!-- Data Table -->
    <div class="table-container">
      <n-data-table
        :columns="columns"
        :data="filteredNodes"
        :loading="loading"
        :bordered="false"
        :row-key="(row: any) => row.name"
      />

      <div class="pagination-container">
        <n-pagination
          v-model:page="page"
          v-model:page-size="pageSize"
          :item-count="total"
          :page-sizes="[10, 20, 50, 100]"
          show-size-picker
          show-quick-jumper
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </div>
    </div>

    <!-- Detail Modal -->
    <n-modal
      v-model:show="showDetailModal"
      preset="card"
      title="节点详情"
      style="width: 900px"
      :trap-focus="false"
      :block-scroll="true"
    >
      <n-space vertical v-if="nodeDetail" :size="16">
        <n-descriptions bordered :column="2">
          <n-descriptions-item label="CPU使用率">
            {{ calculatePercentage(nodeDetail.cpuUsage, nodeDetail.cpuCapacity) }}
          </n-descriptions-item>
          <n-descriptions-item label="内存使用率">
            {{ calculatePercentage(nodeDetail.memoryUsage, nodeDetail.memoryCapacity) }}
          </n-descriptions-item>
          <n-descriptions-item label="Pod使用率">
            {{ calculatePercentage(String(nodeDetail.podCount), String(nodeDetail.podCapacity)) }}
          </n-descriptions-item>
        </n-descriptions>

        <div>
          <h4>Pod列表</h4>
          <n-data-table
            :columns="[
              { title: 'Pod名称', key: 'name' },
              { title: '命名空间', key: 'namespace' },
              { title: 'CPU请求', key: 'cpuRequest' },
              { title: '内存请求', key: 'memoryRequest' },
              { title: 'CPU限制', key: 'cpuLimit' },
              { title: '内存限制', key: 'memoryLimit' },
            ]"
            :data="nodeDetail.pods || []"
            :pagination="{ page: podPage, pageSize: podPageSize }"
            @update:page="podPage = $event"
          />
        </div>
      </n-space>
    </n-modal>

    <!-- Taint Modal -->
    <n-modal v-model:show="showTaintModal" preset="card" title="污点管理" style="width: 700px">
      <n-space vertical :size="16">
        <div>
          <h4>现有污点</h4>
          <n-space v-if="taints.length > 0">
            <n-tag
              v-for="(taint, index) in taints"
              :key="index"
              closable
              @close="removeTaint(index)"
            >
              {{ taint.key }}={{ taint.value }}:{{ taint.effect }}
            </n-tag>
          </n-space>
          <div v-else style="color: #999">暂无污点</div>
        </div>

        <n-form inline>
          <n-form-item label="键名">
            <n-input
              v-model:value="newTaint.key"
              placeholder="例如: node-role"
              style="width: 150px"
            />
          </n-form-item>
          <n-form-item label="值">
            <n-input
              v-model:value="newTaint.value"
              placeholder="例如: master"
              style="width: 120px"
            />
          </n-form-item>
          <n-form-item label="效果">
            <n-select
              v-model:value="newTaint.effect"
              :options="taintEffects"
              style="width: 200px"
            />
          </n-form-item>
          <n-button type="primary" @click="addTaint">添加</n-button>
        </n-form>
      </n-space>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showTaintModal = false">取消</n-button>
          <n-button type="primary" @click="saveTaints">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Label Modal -->
    <n-modal v-model:show="showLabelModal" preset="card" title="标签管理" style="width: 700px">
      <n-space vertical :size="16">
        <div>
          <h4>现有标签</h4>
          <n-space v-if="Object.keys(labels).length > 0">
            <n-tag
              v-for="(value, key) in labels"
              :key="key"
              closable
              @close="removeLabel(key)"
            >
              {{ key }}={{ value }}
            </n-tag>
          </n-space>
          <div v-else style="color: #999">暂无标签</div>
        </div>

        <n-form inline>
          <n-form-item label="键">
            <n-input
              v-model:value="newLabel.key"
              placeholder="例如: env"
              style="width: 150px"
            />
          </n-form-item>
          <n-form-item label="值">
            <n-input
              v-model:value="newLabel.value"
              placeholder="例如: production"
              style="width: 150px"
            />
          </n-form-item>
          <n-button type="primary" @click="addLabel">添加</n-button>
        </n-form>
      </n-space>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showLabelModal = false">取消</n-button>
          <n-button type="primary" @click="saveLabels">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Drain Modal -->
    <n-modal v-model:show="showDrainModal" preset="card" title="驱逐节点" style="width: 600px">
      <n-space vertical :size="16">
        <div style="color: #d03050">
          <strong>警告：</strong>此操作将驱逐节点上的所有Pod，请谨慎操作！
        </div>

        <n-form label-placement="left" label-width="140">
          <n-form-item label="强制驱逐">
            <n-checkbox v-model:checked="drainOptions.force">
              强制删除Pod（即使违反PDB）
            </n-checkbox>
          </n-form-item>
          <n-form-item label="忽略DaemonSet">
            <n-checkbox v-model:checked="drainOptions.ignoreDaemonSets">
              忽略DaemonSet管理的Pod
            </n-checkbox>
          </n-form-item>
          <n-form-item label="删除本地数据">
            <n-checkbox v-model:checked="drainOptions.deleteLocalData">
              删除使用emptyDir的Pod
            </n-checkbox>
          </n-form-item>
          <n-form-item label="优雅期限（秒）">
            <n-input-number
              v-model:value="drainOptions.gracePeriodSeconds"
              :min="0"
              :max="300"
            />
          </n-form-item>
        </n-form>
      </n-space>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showDrainModal = false">取消</n-button>
          <n-button type="error" @click="drainNode">确认驱逐</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.node-list-page {
  padding: var(--spacing-lg);
}

.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background: var(--card-bg);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-lg);
  align-items: flex-end;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.filter-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.filter-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.filter-extra {
  margin-left: auto;
  display: flex;
  gap: var(--spacing-md);
  align-items: center;
}

.table-container {
  background: var(--card-bg);
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-lg);
}
</style>
