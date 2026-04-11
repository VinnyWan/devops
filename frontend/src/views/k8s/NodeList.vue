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

    <el-table :data="tableData" stripe v-loading="loading" style="margin-top: 16px">
      <el-table-column prop="name" label="节点名称" width="150"/>
      <el-table-column label="状态">
        <template #default="{ row }">
          <StatusTag :status="row.status" />
        </template>
      </el-table-column>
      <el-table-column prop="role" label="角色" />
      <el-table-column prop="k8sVersion" label="版本" width="120"/>
      <el-table-column prop="ip" label="IP" />
      <el-table-column label="CPU">
        <template #default="{ row }">
          {{ row.cpuUsage }} / {{ row.cpuCapacity }}
        </template>
      </el-table-column>
      <el-table-column label="内存">
        <template #default="{ row }">
          {{ row.memoryUsage }} / {{ row.memoryCapacity }}
        </template>
      </el-table-column>
      <el-table-column prop="podCount" label="Pod数" />
      <el-table-column prop="age" label="运行时间" />
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleCordon(row)">隔离</el-button>
          <el-button link type="primary" size="small" @click="handleDrain(row)">驱逐</el-button>
          <el-button link type="primary" size="small" @click="handleLabel(row)">标签</el-button>
        </template>
      </el-table-column>
    </el-table>

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
import { ElMessage, ElMessageBox } from 'element-plus'
import ClusterSelector from '@/components/K8s/ClusterSelector.vue'
import StatusTag from '@/components/K8s/StatusTag.vue'
import { getNodeList, cordonNode, drainNode } from '@/api/node'

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

const handleCordon = async (row) => {
  await ElMessageBox.confirm('确认隔离该节点?', '提示')
  await cordonNode({ clusterName: clusterName.value, nodeName: row.name })
  ElMessage.success('操作成功')
  fetchData()
}

const handleDrain = async (row) => {
  await ElMessageBox.confirm('确认驱逐该节点?', '提示')
  await drainNode({ clusterName: clusterName.value, nodeName: row.name })
  ElMessage.success('操作成功')
  fetchData()
}

const handleLabel = (row) => {
  ElMessage.info('标签编辑功能待实现')
}

onMounted(fetchData)
</script>

