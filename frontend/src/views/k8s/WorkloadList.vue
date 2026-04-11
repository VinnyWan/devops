<template>
  <div class="page-container">
    <div class="search-bar">
      <ClusterSelector v-model="clusterName" />
      <NamespaceSelector v-model="namespace" :cluster-name="clusterName" style="margin-left: 12px" />
      <el-input v-model="keyword" placeholder="搜索" style="width: 200px; margin-left: 12px" clearable />
      <el-button type="primary" @click="fetchData" style="margin-left: 12px">查询</el-button>
    </div>

    <el-tabs v-model="activeTab" @tab-change="handleTabChange" style="margin-top: 16px">
      <el-tab-pane label="Deployment" name="deployment" />
      <el-tab-pane label="StatefulSet" name="statefulset" />
      <el-tab-pane label="DaemonSet" name="daemonset" />
      <el-tab-pane label="Pod" name="pod" />
    </el-tabs>

    <el-table :data="tableData" stripe>
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="namespace" label="命名空间" v-if="!namespace" />
      <el-table-column prop="replicas" label="副本数" v-if="activeTab !== 'pod'" />
      <el-table-column prop="image" label="镜像" v-if="activeTab !== 'pod'" />
      <el-table-column prop="status" label="状态" v-if="activeTab === 'pod'">
        <template #default="{ row }">
          <StatusTag :status="row.status" />
        </template>
      </el-table-column>
      <el-table-column prop="ip" label="IP" v-if="activeTab === 'pod'" />
      <el-table-column prop="node" label="节点" v-if="activeTab === 'pod'" />
      <el-table-column label="创建时间">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleRestart(row)" v-if="activeTab === 'deployment'">重启</el-button>
          <el-button link type="primary" size="small" @click="handleScale(row)" v-if="activeTab !== 'pod'">扩缩容</el-button>
          <el-button link type="primary" size="small" @click="handleLogs(row)" v-if="activeTab === 'pod'">日志</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="scaleVisible" title="扩缩容" width="400px">
      <el-form label-width="80px">
        <el-form-item label="副本数">
          <el-input-number v-model="scaleReplicas" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scaleVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmScale">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="logVisible" title="Pod 日志" width="800px">
      <LogViewer :logs="logs" />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ClusterSelector from '@/components/K8s/ClusterSelector.vue'
import NamespaceSelector from '@/components/K8s/NamespaceSelector.vue'
import StatusTag from '@/components/K8s/StatusTag.vue'
import LogViewer from '@/components/K8s/LogViewer.vue'
import { getDeploymentList, restartDeployment, scaleDeployment, deleteDeployment } from '@/api/deployment'
import { getStatefulSetList, scaleStatefulSet, deleteStatefulSet } from '@/api/statefulset'
import { getDaemonSetList, deleteDaemonSet } from '@/api/daemonset'
import { getPodList, deletePod, getPodLogs } from '@/api/workload'
import { formatTime } from '@/utils/format'

const clusterName = ref('')
const namespace = ref('')
const keyword = ref('')
const activeTab = ref('deployment')
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const scaleVisible = ref(false)
const scaleReplicas = ref(1)
const currentRow = ref(null)
const logVisible = ref(false)
const logs = ref('')

const fetchData = async () => {
  const params = { clusterName: clusterName.value, namespace: namespace.value, keyword: keyword.value, page: page.value, pageSize: pageSize.value }
  let res
  if (activeTab.value === 'deployment') res = await getDeploymentList(params)
  else if (activeTab.value === 'statefulset') res = await getStatefulSetList(params)
  else if (activeTab.value === 'daemonset') res = await getDaemonSetList(params)
  else res = await getPodList(params)
  tableData.value = res.data || []
  total.value = res.total || 0
}

const handleTabChange = () => {
  page.value = 1
  fetchData()
}

const handleRestart = async (row) => {
  await ElMessageBox.confirm('确认重启?', '提示')
  await restartDeployment({ clusterName: clusterName.value, namespace: row.namespace, name: row.name })
  ElMessage.success('重启成功')
  fetchData()
}

const handleScale = (row) => {
  currentRow.value = row
  scaleReplicas.value = row.replicas || 1
  scaleVisible.value = true
}

const confirmScale = async () => {
  const data = { clusterName: clusterName.value, namespace: currentRow.value.namespace, name: currentRow.value.name, replicas: scaleReplicas.value }
  if (activeTab.value === 'deployment') await scaleDeployment(data)
  else await scaleStatefulSet(data)
  ElMessage.success('操作成功')
  scaleVisible.value = false
  fetchData()
}

const handleLogs = async (row) => {
  const res = await getPodLogs({ clusterName: clusterName.value, namespace: row.namespace, podName: row.name })
  logs.value = res.data || ''
  logVisible.value = true
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除?', '提示')
  const data = { clusterName: clusterName.value, namespace: row.namespace, name: row.name }
  if (activeTab.value === 'deployment') await deleteDeployment(data)
  else if (activeTab.value === 'statefulset') await deleteStatefulSet(data)
  else if (activeTab.value === 'daemonset') await deleteDaemonSet(data)
  else await deletePod(data)
  ElMessage.success('删除成功')
  fetchData()
}
</script>

