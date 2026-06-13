<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <el-button link type="primary" @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon> 返回
        </el-button>
        <h3>工单详情</h3>
      </div>
      <div class="header-actions">
        <template v-if="order.status === 'draft'">
          <el-button type="primary" @click="handleSubmit">提交审批</el-button>
        </template>
        <template v-if="order.status === 'pending_review'">
          <el-button type="success" @click="handleApprove">批准</el-button>
          <el-button type="danger" @click="handleReject">驳回</el-button>
        </template>
        <template v-if="order.status === 'approved'">
          <el-button type="primary" @click="handleExecute">执行</el-button>
        </template>
      </div>
    </div>

    <el-descriptions v-if="order.id" :column="2" border>
      <el-descriptions-item label="工单编号">{{ order.id }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="statusTag(order.status)" size="small">{{ statusLabel(order.status) }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="标题" :span="2">{{ order.title }}</el-descriptions-item>
      <el-descriptions-item label="类型">{{ order.type }}</el-descriptions-item>
      <el-descriptions-item label="审批进度">{{ order.currentLevel }} / {{ order.approvalLevels }}</el-descriptions-item>
      <el-descriptions-item label="创建时间">{{ formatTime(order.created_at) }}</el-descriptions-item>
      <el-descriptions-item label="更新时间">{{ formatTime(order.updated_at) }}</el-descriptions-item>
      <el-descriptions-item label="描述" :span="2">{{ order.description || '-' }}</el-descriptions-item>
      <el-descriptions-item v-if="order.callback_module" label="回调模块">{{ order.callback_module }}</el-descriptions-item>
      <el-descriptions-item v-if="order.callback_action" label="回调动作">{{ order.callback_action }}</el-descriptions-item>
      <el-descriptions-item v-if="order.callback_payload" label="回调参数" :span="2">
        <pre style="margin:0;white-space:pre-wrap;font-size:12px">{{ order.callback_payload }}</pre>
      </el-descriptions-item>
    </el-descriptions>

    <el-card v-if="order.id" class="timeline-card" shadow="never">
      <template #header><span>审批记录</span></template>
      <el-timeline v-if="approvals.length">
        <el-timeline-item
          v-for="item in approvals"
          :key="item.id"
          :timestamp="formatTime(item.approved_at)"
          :type="item.status === 'approved' ? 'success' : 'danger'"
        >
          <p>第 {{ item.level }} 级审批 · 审批人 ID: {{ item.approver_id }}</p>
          <p v-if="item.comment" style="color:#909399;font-size:13px">{{ item.comment }}</p>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-else description="暂无审批记录" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderDetail, submitOrder, approveOrder, rejectOrder, executeOrder } from '@/api/workflow'
import { formatTime } from '@/utils/format'

const route = useRoute()
const order = ref({})
const approvals = ref([])

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

const fetchDetail = async () => {
  try {
    const res = await getOrderDetail(route.params.id)
    order.value = res.data || {}
    approvals.value = order.value.approvals || []
  } catch {
    ElMessage.error('获取工单详情失败')
  }
}

const refresh = async () => { await fetchDetail() }

const handleSubmit = async () => {
  try {
    await submitOrder(order.value.id)
    ElMessage.success('已提交审批')
    refresh()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '提交失败')
  }
}

const handleApprove = async () => {
  try {
    const { value: comment } = await ElMessageBox.prompt('审批意见', '批准工单', {
      confirmButtonText: '批准', inputPlaceholder: '输入审批意见（可选）'
    })
    await approveOrder(order.value.id, comment || '同意')
    ElMessage.success('已批准')
    refresh()
  } catch { /* cancel */ }
}

const handleReject = async () => {
  try {
    const { value: comment } = await ElMessageBox.prompt('驳回原因', '驳回工单', {
      confirmButtonText: '驳回', inputPlaceholder: '输入驳回原因', inputType: 'textarea'
    })
    if (!comment) return
    await rejectOrder(order.value.id, comment)
    ElMessage.success('已驳回')
    refresh()
  } catch { /* cancel */ }
}

const handleExecute = async () => {
  try { await ElMessageBox.confirm('确定执行该工单？', '确认', { type: 'warning' }) } catch { return }
  try {
    await executeOrder(order.value.id)
    ElMessage.success('执行已触发')
    refresh()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '执行失败')
  }
}

onMounted(fetchDetail)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.header-left { display: flex; align-items: center; gap: 16px; }
.header-left h3 { margin: 0; font-size: 18px; font-weight: 500; }
.header-actions { display: flex; gap: 8px; }
.timeline-card { margin-top: 24px; }
.timeline-card p { margin: 0; }
</style>
