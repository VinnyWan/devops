<template>
  <div class="page-container">
    <div class="search-bar">
      <ClusterSelector v-model="clusterName" />
      <NamespaceSelector v-model="namespace" :cluster-name="clusterName" style="margin-left: 12px" />
      <el-input v-model="keyword" placeholder="搜索" style="width: 200px; margin-left: 12px" clearable />
      <el-button type="primary" @click="fetchData" style="margin-left: 12px">查询</el-button>
    </div>

    <el-table :data="tableData" stripe style="margin-top: 16px">
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="namespace" label="命名空间" v-if="!namespace" />
      <el-table-column prop="dataCount" label="数据项数量" />
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
import ClusterSelector from '@/components/K8s/ClusterSelector.vue'
import NamespaceSelector from '@/components/K8s/NamespaceSelector.vue'
import { getConfigMapList, deleteConfigMap } from '@/api/configmap'
import { formatTime } from '@/utils/format'

const clusterName = ref('')
const namespace = ref('')
const keyword = ref('')
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const fetchData = async () => {
  const res = await getConfigMapList({ clusterName: clusterName.value, namespace: namespace.value, keyword: keyword.value, page: page.value, pageSize: pageSize.value })
  tableData.value = res.data || []
  total.value = res.total || 0
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除?', '提示')
  await deleteConfigMap({ clusterName: clusterName.value, namespace: row.namespace, name: row.name })
  ElMessage.success('删除成功')
  fetchData()
}
</script>

