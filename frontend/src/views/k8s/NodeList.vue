<template>
  <div class="page-container">
    <div class="search-bar">
      <ClusterSelector v-model="clusterName" @change="onClusterChange" />
      <el-input v-model="keyword" placeholder="搜索节点" style="width: 200px; margin-left: 12px" clearable @keyup.enter="fetchData" />
      <el-select v-model="statusFilter" placeholder="状态" style="width: 120px; margin-left: 12px" @change="fetchData">
        <el-option label="全部" value="" />
        <el-option label="Ready" value="Ready" />
        <el-option label="NotReady" value="NotReady" />
      </el-select>
    </div>

    <el-table v-if="tableData.length || loading" :data="tableData" stripe v-loading="loading" style="margin-top: 16px">
      <el-table-column prop="name" label="节点名称" width="150"/>
      <el-table-column label="状态">
        <template #default="{ row }">
          <StatusTag :status="row.status" />
        </template>
      </el-table-column>
      <el-table-column prop="role" label="角色" />
      <el-table-column prop="k8sVersion" label="版本" width="120"/>
      <el-table-column prop="ip" label="IP" />
      <el-table-column label="CPU" min-width="140">
        <template #default="{ row }">
          <span class="nowrap">{{ formatCPU(row.cpuUsage) }} / {{ formatCPU(row.cpuCapacity) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="内存" min-width="140">
        <template #default="{ row }">
          <span class="nowrap">{{ formatMemory(row.memoryUsage) }} / {{ formatMemory(row.memoryCapacity) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="podCount" label="Pod数" />
      <el-table-column prop="age" label="运行时间" />
      <el-table-column label="操作" width="240">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleViewDetail(row)">详情</el-button>
          <el-button link type="warning" size="small" @click="handleCordon(row)">隔离</el-button>
          <el-button link type="danger" size="small" @click="handleDrain(row)">驱逐</el-button>
          <el-button link type="success" size="small" @click="handleLabel(row)">标签</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-else description="暂无节点数据" style="margin-top: 16px" />

    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      @current-change="fetchData"
      style="margin-top: 16px; justify-content: flex-end"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ClusterSelector from '@/components/K8s/ClusterSelector.vue'
import StatusTag from '@/components/K8s/StatusTag.vue'
import { getNodeList, cordonNode, drainNode } from '@/api/node'

const router = useRouter()
const clusterName = ref('')
const loading = ref(false)
const keyword = ref('')
const statusFilter = ref('')
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const fetchData = async () => {
  if (!clusterName.value) return
  loading.value = true
  try {
    const res = await getNodeList({ clusterName: clusterName.value, name: keyword.value, status: statusFilter.value, page: page.value, pageSize: pageSize.value })
    tableData.value = res.data?.items || []
    total.value = res.data?.total || 0
  } finally {
    loading.value = false
  }
}

const onClusterChange = () => {
  page.value = 1
  fetchData()
}

const handleViewDetail = (row) => {
  router.push(`/k8s/node/${encodeURIComponent(clusterName.value)}/${encodeURIComponent(row.name)}`)
}

const handleCordon = async (row) => {
  await ElMessageBox.confirm('确认隔离该节点?', '提示')
  await cordonNode({ clusterName: clusterName.value, name: row.name })
  ElMessage.success('操作成功')
  fetchData()
}

const handleDrain = async (row) => {
  await ElMessageBox.confirm('确认驱逐该节点?', '提示')
  await drainNode({ clusterName: clusterName.value, name: row.name })
  ElMessage.success('操作成功')
  fetchData()
}

const handleLabel = (row) => {
  ElMessage.info('标签编辑功能待实现')
}

const formatCPU = (val) => {
  if (val === null || val === undefined || val === '') return '-'

  const text = String(val).trim()
  const normalized = text.toLowerCase()
  const match = normalized.match(/^([0-9]+(?:\.[0-9]+)?)\s*(m|millicores?|c|core|cores)?$/)

  if (!match) return text

  const amount = Number(match[1])
  const unit = match[2] || 'c'
  const cores = unit.startsWith('m') ? amount / 1000 : amount

  return `${trimTrailingZeros(cores)}C`
}

const formatMemory = (val) => {
  if (val === null || val === undefined || val === '') return '-'

  const text = String(val).trim()
  const normalized = text.replace(/\s+/g, '')
  const match = normalized.match(/^([0-9]+(?:\.[0-9]+)?)(Ki|Mi|Gi|Ti|Pi|K|M|G|T|P|KB|MB|GB|TB|PB|B)?$/i)

  if (!match) return text

  const amount = Number(match[1])
  const unit = (match[2] || 'G').toUpperCase()
  const unitMap = {
    B: 1 / 1024 ** 3,
    KI: 1 / 1024 ** 2,
    K: 1 / 1024 ** 2,
    KB: 1 / 1024 ** 2,
    MI: 1 / 1024,
    M: 1 / 1024,
    MB: 1 / 1024,
    GI: 1,
    G: 1,
    GB: 1,
    TI: 1024,
    T: 1024,
    TB: 1024,
    PI: 1024 ** 2,
    P: 1024 ** 2,
    PB: 1024 ** 2
  }

  const valueInG = amount * (unitMap[unit] || 1)
  return `${trimTrailingZeros(valueInG)}G`
}

const trimTrailingZeros = (num) => {
  if (!Number.isFinite(num)) return '-'
  return Number(num.toFixed(num >= 100 ? 0 : num >= 10 ? 1 : 2)).toString()
}

onMounted(fetchData)
</script>

<style scoped>
.nowrap { white-space: nowrap; }
</style>

