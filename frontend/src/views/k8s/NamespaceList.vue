<template>
  <div class="page-container">
    <div class="search-bar">
      <ClusterSelector v-model="clusterName" @change="onClusterChange" />
      <el-input v-model="keyword" placeholder="搜索命名空间" style="width: 200px; margin-left: 12px" clearable @keyup.enter="fetchData" />
      <el-button type="primary" @click="fetchData" style="margin-left: 12px">查询</el-button>
      <el-button type="success" @click="showCreateDialog" style="margin-left: 12px">创建命名空间</el-button>
    </div>

    <el-table :data="tableData" stripe v-loading="loading" style="margin-top: 16px">
      <el-table-column prop="name" label="名称" />
      <el-table-column label="状态">
        <template #default="{ row }">
          <StatusTag :status="row.status" />
        </template>
      </el-table-column>
      <el-table-column label="创建时间">
        <template #default="{ row }">{{ formatTime(row.creationTimestamp) }}</template>
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

    <el-dialog v-model="dialogVisible" title="创建命名空间" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="请输入命名空间名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ClusterSelector from '@/components/K8s/ClusterSelector.vue'
import StatusTag from '@/components/K8s/StatusTag.vue'
import { getNamespaceList, createNamespace, deleteNamespace } from '@/api/namespace'
import { formatTime } from '@/utils/format'

const clusterName = ref('')
const keyword = ref('')
const loading = ref(false)
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const form = ref({ name: '' })

const fetchData = async () => {
  if (!clusterName.value) return
  loading.value = true
  try {
    const res = await getNamespaceList({
      clusterName: clusterName.value,
      keyword: keyword.value,
      page: page.value,
      pageSize: pageSize.value
    })
    tableData.value = res.data?.items || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error(e.message || '获取列表失败')
  } finally {
    loading.value = false
  }
}

const onClusterChange = () => {
  page.value = 1
  fetchData()
}

const showCreateDialog = () => {
  form.value = { name: '' }
  dialogVisible.value = true
}

const handleCreate = async () => {
  if (!form.value.name) {
    ElMessage.warning('请输入命名空间名称')
    return
  }
  try {
    await createNamespace({ clusterName: clusterName.value, name: form.value.name })
    ElMessage.success('创建成功')
    dialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '创建失败')
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除命名空间 "${row.name}"？该操作不可恢复。`, '删除确认', { type: 'warning' })
    await deleteNamespace({ clusterName: clusterName.value, name: row.name })
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error(e.message || '删除失败')
    }
  }
}

onMounted(fetchData)
</script>
