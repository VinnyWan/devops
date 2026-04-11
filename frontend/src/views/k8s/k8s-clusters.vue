<template>
  <div class="page-container">
    <div class="page-header">
      <h3>集群管理</h3>
      <el-button type="primary" @click="showCreateDialog">新建集群</el-button>
    </div>

    <el-table :data="tableData" stripe>
      <el-table-column prop="name" label="集群名称" width="150"/>
      <el-table-column prop="url" label="API Server" width="220"/>
      <el-table-column prop="env" label="环境" width="60"/>
      <el-table-column prop="k8sVersion" label="版本" width="150"/>
      <el-table-column prop="nodeCount" label="节点数" />
      <el-table-column label="状态">
        <template #default="{ row }">
          <el-tag :type="row.status === 'healthy' ? 'success' : 'danger'" size="small">
            {{ row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="默认">
        <template #default="{ row }">
          <el-tag v-if="row.isDefault" type="success" size="small">是</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="200">
        <template #default="{ row }">
          {{ formatTime(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleDetail(row.name)">详情</el-button>
          <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row.id)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑集群' : '创建集群'" width="900px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="集群名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="API Server">
          <el-input v-model="form.url" placeholder="https://xxx:6443" />
        </el-form-item>
        <el-form-item label="认证类型">
          <el-select v-model="form.authType" style="width: 100%">
            <el-option label="Kubeconfig" value="kubeconfig" />
            <el-option label="Token" value="token" />
          </el-select>
        </el-form-item>
        <el-form-item label="Kubeconfig" v-if="form.authType === 'kubeconfig'">
          <el-input v-model="form.kubeconfig" type="textarea" :rows="6" />
        </el-form-item>
        <el-form-item label="Token" v-if="form.authType === 'token'">
          <el-input v-model="form.token" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="环境">
          <el-select v-model="form.env" style="width: 100%">
            <el-option label="开发" value="dev" />
            <el-option label="测试" value="test" />
            <el-option label="生产" value="prod" />
          </el-select>
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="form.isDefault" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getClusterList, createCluster, updateCluster, deleteCluster } from '../../api/cluster'
import { formatTime } from '../../utils/format'

const router = useRouter()
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const isEdit = ref(false)
const form = ref({
  name: '',
  url: '',
  authType: 'kubeconfig',
  kubeconfig: '',
  token: '',
  env: 'prod',
  isDefault: false,
  remark: ''
})

const fetchData = async () => {
  const res = await getClusterList({ page: page.value, pageSize: pageSize.value })
  tableData.value = res.data || []
  total.value = res.total || 0
}

const showCreateDialog = () => {
  isEdit.value = false
  form.value = { name: '', url: '', authType: 'kubeconfig', kubeconfig: '', token: '', env: 'prod', isDefault: false, remark: '' }
  dialogVisible.value = true
}

const handleDetail = (name) => {
  router.push(`/k8s/cluster/${encodeURIComponent(name)}`)
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { ...row }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const data = { ...form.value }
  if (isEdit.value) {
    if (!data.kubeconfig) delete data.kubeconfig
    if (!data.token) delete data.token
    await updateCluster(data)
    ElMessage.success('更新成功')
  } else {
    await createCluster(data)
    ElMessage.success('创建成功')
  }
  dialogVisible.value = false
  fetchData()
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该集群?', '提示')
  await deleteCluster(id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
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
</style>

