<template>
  <div class="page-container">
    <div class="search-bar">
      <ClusterSelector v-model="clusterName" />
      <NamespaceSelector v-model="namespace" :cluster-name="clusterName" style="margin-left: 12px" />
      <el-input v-model="keyword" placeholder="搜索" style="width: 200px; margin-left: 12px" clearable />
      <el-button type="primary" @click="fetchData" style="margin-left: 12px">查询</el-button>
    </div>

    <el-tabs v-model="activeTab" @tab-change="handleTabChange" style="margin-top: 16px">
      <el-tab-pane label="Service" name="service" />
      <el-tab-pane label="Ingress" name="ingress" />
    </el-tabs>

    <el-table :data="tableData" stripe>
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="namespace" label="命名空间" v-if="!namespace" />
      <el-table-column prop="type" label="类型" v-if="activeTab === 'service'" />
      <el-table-column prop="clusterIP" label="ClusterIP" v-if="activeTab === 'service'" />
      <el-table-column prop="ports" label="端口" v-if="activeTab === 'service'" />
      <el-table-column prop="host" label="主机" v-if="activeTab === 'ingress'" />
      <el-table-column prop="path" label="路径" v-if="activeTab === 'ingress'" />
      <el-table-column label="创建时间">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
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
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ClusterSelector from '@/components/K8s/ClusterSelector.vue'
import NamespaceSelector from '@/components/K8s/NamespaceSelector.vue'
import { getServiceList, deleteService } from '@/api/service'
import { getIngressList, deleteIngress } from '@/api/ingress'
import { formatTime } from '@/utils/format'

const clusterName = ref('')
const namespace = ref('')
const keyword = ref('')
const activeTab = ref('service')
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const fetchData = async () => {
  const params = { clusterName: clusterName.value, namespace: namespace.value, keyword: keyword.value, page: page.value, pageSize: pageSize.value }
  const res = activeTab.value === 'service' ? await getServiceList(params) : await getIngressList(params)
  tableData.value = res.data || []
  total.value = res.total || 0
}

const handleTabChange = () => {
  page.value = 1
  fetchData()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除?', '提示')
  const data = { clusterName: clusterName.value, namespace: row.namespace, name: row.name }
  activeTab.value === 'service' ? await deleteService(data) : await deleteIngress(data)
  ElMessage.success('删除成功')
  fetchData()
}
</script>

