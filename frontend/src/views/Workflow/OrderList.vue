<template>
  <div class="page-container">
    <div class="page-header">
      <h3>变更工单</h3>
      <el-button type="primary" @click="showCreateDialog">创建工单</el-button>
    </div>
    <div class="toolbar">
      <el-select v-model="filterStatus" placeholder="全部状态" clearable style="width: 150px" @change="fetchData">
        <el-option label="草稿" value="draft" />
        <el-option label="审批中" value="pending_review" />
        <el-option label="已批准" value="approved" />
        <el-option label="执行中" value="executing" />
        <el-option label="已完成" value="completed" />
        <el-option label="已驳回" value="rejected" />
        <el-option label="失败" value="failed" />
      </el-select>
    </div>
    <el-table :data="tableData" stripe v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
      <el-table-column prop="type" label="类型" width="100" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="审批进度" width="120">
        <template #default="{ row }">
          {{ row.currentLevel }} / {{ row.approvalLevels }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="goDetail(row.id)">详情</el-button>
          <template v-if="row.status === 'draft'">
            <el-button link type="primary" size="small" @click="handleSubmit(row)">提交</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
          <template v-if="row.status === 'pending_review'">
            <el-button link type="success" size="small" @click="handleApprove(row)">批准</el-button>
            <el-button link type="danger" size="small" @click="handleReject(row)">驳回</el-button>
          </template>
          <template v-if="row.status === 'approved'">
            <el-button link type="primary" size="small" @click="handleExecute(row)">执行</el-button>
          </template>
        </template>
      </el-table-column>
    </el-table>
    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </div>

    <el-dialog v-model="dialogVisible" title="创建工单" width="500px" destroy-on-close>
      <el-form :model="form" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="输入工单标题" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" placeholder="选择工单类型" style="width: 100%">
            <el-option label="配置变更" value="config" />
            <el-option label="发布上线" value="deploy" />
            <el-option label="资源扩容" value="scale" />
            <el-option label="安全变更" value="security" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="审批级别">
          <el-input-number v-model="form.approvalLevels" :min="1" :max="3" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="4" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderList, createOrder, submitOrder, approveOrder, rejectOrder, executeOrder } from '@/api/workflow'
import { formatTime } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const filterStatus = ref('')

const dialogVisible = ref(false)
const submitting = ref(false)
const formRef = ref()
const form = reactive({
  title: '',
  type: 'config',
  description: '',
  approvalLevels: 1
})
const formRules = {
  title: [{ required: true, message: '请输入工单标题', trigger: 'blur' }],
  type: [{ required: true, message: '请选择工单类型', trigger: 'change' }]
}

const statusLabel = (s) => {
  const map = {
    draft: '草稿', pending_review: '审批中', approved: '已批准',
    executing: '执行中', completed: '已完成', rejected: '已驳回', failed: '失败'
  }
  return map[s] || s
}

const statusTag = (s) => {
  const map = {
    draft: 'info', pending_review: 'warning', approved: 'success',
    executing: '', completed: 'success', rejected: 'danger', failed: 'danger'
  }
  return map[s] || 'info'
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = { page: page.value, pageSize: pageSize.value }
    if (filterStatus.value) params.status = filterStatus.value
    const res = await getOrderList(params)
    const data = res.data || {}
    tableData.value = data.items || []
    total.value = data.total || 0
  } catch {
    ElMessage.error('获取工单列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  form.title = ''
  form.type = 'config'
  form.description = ''
  form.approvalLevels = 1
  dialogVisible.value = true
}

const handleCreate = async () => {
  try { await formRef.value.validate() } catch { return }
  submitting.value = true
  try {
    await createOrder({ ...form })
    ElMessage.success('创建成功')
    dialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  } finally {
    submitting.value = false
  }
}

const goDetail = (id) => router.push(`/workflow/orders/${id}`)

const handleSubmit = async (row) => {
  try {
    await submitOrder(row.id)
    ElMessage.success('已提交审批')
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '提交失败')
  }
}

const handleApprove = async (row) => {
  try {
    const { value: comment } = await ElMessageBox.prompt('审批意见', '批准工单', {
      confirmButtonText: '批准', inputPlaceholder: '输入审批意见（可选）'
    })
    await approveOrder(row.id, comment || '同意')
    ElMessage.success('已批准')
    fetchData()
  } catch { /* cancel */ }
}

const handleReject = async (row) => {
  try {
    const { value: comment } = await ElMessageBox.prompt('驳回原因', '驳回工单', {
      confirmButtonText: '驳回', inputPlaceholder: '输入驳回原因', inputType: 'textarea'
    })
    if (!comment) return
    await rejectOrder(row.id, comment)
    ElMessage.success('已驳回')
    fetchData()
  } catch { /* cancel */ }
}

const handleExecute = async (row) => {
  try { await ElMessageBox.confirm('确定执行该工单？', '确认', { type: 'warning' }) } catch { return }
  try {
    await executeOrder(row.id)
    ElMessage.success('执行已触发')
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '执行失败')
  }
}

const handleDelete = async (row) => {
  try { await ElMessageBox.confirm('确定删除该工单？', '确认', { type: 'warning' }) } catch { return }
  ElMessage.info('删除功能待实现')
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
