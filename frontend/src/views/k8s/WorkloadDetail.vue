<template>
  <div class="page-container" v-loading="loading">
    <!-- 头部区域 -->
    <div class="page-header">
      <div class="header-left">
        <el-button text @click="router.back()">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <el-tag class="kind-tag">{{ kindLabel }}</el-tag>
        <h3 class="header-name">{{ name }}</h3>
      </div>
      <div class="header-right">
        <el-button v-if="kind !== 'job'" type="primary" plain @click="openYamlEdit">YAML 编辑</el-button>
        <el-button v-if="canRestart" type="warning" plain @click="handleRestart">重启</el-button>
        <el-button v-if="canScale" type="success" plain @click="openScaleDialog">扩缩容</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
      </div>
    </div>

    <el-alert v-if="errorText" :title="errorText" type="error" show-icon :closable="false" style="margin-bottom: 16px" />

    <template v-if="detail">
      <!-- 基本信息 -->
      <el-card shadow="never" style="margin-bottom: 16px">
        <template #header>基本信息</template>
        <el-descriptions :column="4" border>
          <el-descriptions-item label="命名空间">{{ detail.namespace || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <StatusTag :status="detail.status || '-'" />
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(detail.createdAt || detail.creationTimestamp) }}</el-descriptions-item>
          <el-descriptions-item label="镜像">{{ formatImages(detail.containers) }}</el-descriptions-item>
          <el-descriptions-item label="标签" :span="4">
            <template v-if="detail.labels && Object.keys(detail.labels).length">
              <el-tag v-for="(val, key) in detail.labels" :key="key" size="small" class="label-tag">
                {{ key }}: {{ val }}
              </el-tag>
            </template>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- Pod 列表 -->
      <el-card shadow="never">
        <template #header>Pod 列表 ({{ pods.length }})</template>
        <el-table :data="pods" stripe size="small" v-if="pods.length">
          <el-table-column prop="name" label="名称" min-width="220" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="120">
            <template #default="{ row }">
              <StatusTag :status="row.status || '-'" />
            </template>
          </el-table-column>
          <el-table-column prop="node" label="节点" min-width="140" show-overflow-tooltip />
          <el-table-column prop="podIP" label="IP" min-width="130" show-overflow-tooltip />
          <el-table-column prop="restartCount" label="重启次数" width="100" />
          <el-table-column prop="createdAt" label="创建时间" min-width="170">
            <template #default="{ row }">{{ formatTime(row.createdAt || row.creationTimestamp) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="260" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="openPodYaml(row)">YAML</el-button>
              <el-button link type="primary" size="small" @click="openTerminal(row)">终端</el-button>
              <el-button link type="primary" size="small" @click="openPodLogs(row)">日志</el-button>
              <el-button link type="danger" size="small" @click="handleDeletePod(row)">重启</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无关联 Pod" />
      </el-card>
    </template>

    <el-empty v-else-if="!loading && !errorText" description="暂无工作负载详情数据" />

    <!-- Pod YAML 弹窗 -->
    <el-dialog v-model="podYamlVisible" title="Pod YAML" width="70%" destroy-on-close>
      <YamlEditor v-model="podYamlContent" :readonly="true" :show-copy="true" />
    </el-dialog>

    <!-- Pod 终端弹窗 -->
    <el-dialog v-model="terminalVisible" :title="`终端 - ${currentPod?.name || ''}`" width="80%" destroy-on-close @close="terminalVisible = false">
      <Terminal :ws-url="terminalWsUrl" :visible="terminalVisible" @error="onTerminalError" />
    </el-dialog>

    <!-- Pod 日志弹窗 -->
    <el-dialog v-model="logVisible" :title="`日志 - ${currentPod?.name || ''}`" width="70%" destroy-on-close>
      <div class="log-toolbar">
        <el-select v-model="logContainer" placeholder="选择容器" style="width: 200px" @change="fetchLogs">
          <el-option v-for="c in (currentPod?.containers || [])" :key="c.name" :label="c.name" :value="c.name" />
        </el-select>
        <el-input-number v-model="logTailLines" :min="10" :max="5000" :step="100" style="width: 160px; margin-left: 12px" />
        <el-button type="primary" style="margin-left: 12px" @click="fetchLogs">刷新</el-button>
      </div>
      <LogViewer :logs="logContent" :auto-scroll="true" />
    </el-dialog>

    <!-- 工作负载 YAML 编辑弹窗 -->
    <el-dialog v-model="workloadYamlVisible" title="编辑 YAML" width="70%" destroy-on-close>
      <YamlEditor v-model="workloadYamlContent" :readonly="false" :show-copy="true" />
      <template #footer>
        <el-button @click="workloadYamlVisible = false">取消</el-button>
        <el-button type="primary" :loading="yamlSaving" @click="confirmWorkloadYamlUpdate">确认更新</el-button>
      </template>
    </el-dialog>

    <!-- 扩缩容弹窗 -->
    <el-dialog v-model="scaleVisible" title="扩缩容" width="400px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="副本数">
          <el-input-number v-model="scaleReplicas" :min="0" :max="1000" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scaleVisible = false">取消</el-button>
        <el-button type="primary" :loading="scaleSaving" @click="confirmScale">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import StatusTag from '@/components/K8s/StatusTag.vue'
import YamlEditor from '@/components/K8s/YamlEditor.vue'
import LogViewer from '@/components/K8s/LogViewer.vue'
import Terminal from '@/components/K8s/Terminal.vue'
import { restartDeployment, scaleDeployment, deleteDeployment } from '@/api/deployment'
import { deleteStatefulSet, scaleStatefulSet } from '@/api/statefulset'
import { deleteDaemonSet } from '@/api/daemonset'
import { deleteJob } from '@/api/job'
import { updateCronJobYAML, deleteCronJob } from '@/api/cronjob'
import { getPodLogs, deletePod } from '@/api/workload'
import request from '@/api/request'
import { formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()

const kind = computed(() => route.params.kind)
const clusterName = computed(() => route.params.clusterName)
const namespace = computed(() => route.params.namespace)
const name = computed(() => route.params.name)

// kind 中文映射
const kindLabelMap = {
  deployment: 'Deployment',
  statefulset: 'StatefulSet',
  daemonset: 'DaemonSet',
  job: 'Job',
  cronjob: 'CronJob'
}
const kindLabel = computed(() => kindLabelMap[kind.value] || kind.value)

// 能力判断
const canRestart = computed(() => ['deployment', 'statefulset', 'daemonset'].includes(kind.value))
const canScale = computed(() => ['deployment', 'statefulset'].includes(kind.value))

// 加载状态
const loading = ref(false)
const errorText = ref('')
const detail = ref(null)
const pods = ref([])

// Pod YAML
const podYamlVisible = ref(false)
const podYamlContent = ref('')

// Pod 终端
const terminalVisible = ref(false)
const terminalWsUrl = ref('')
const currentPod = ref(null)
const detectedShell = ref('sh')

// Pod 日志
const logVisible = ref(false)
const logContent = ref('')
const logContainer = ref('')
const logTailLines = ref(500)

// 工作负载 YAML
const workloadYamlVisible = ref(false)
const workloadYamlContent = ref('')
const yamlSaving = ref(false)

// 扩缩容
const scaleVisible = ref(false)
const scaleReplicas = ref(1)
const scaleSaving = ref(false)

// 格式化镜像列表
const formatImages = (containers) => {
  if (!containers || !containers.length) return '-'
  return containers.map(c => c.image).join(', ')
}

// 加载详情和 Pod 列表
const loadDetail = async () => {
  loading.value = true
  errorText.value = ''
  detail.value = null
  pods.value = []
  const params = { clusterName: clusterName.value, namespace: namespace.value, name: name.value }
  try {
    const [detailRes, podsRes] = await Promise.all([
      request.get(`/k8s/${kind.value}/detail`, { params }),
      request.get(`/k8s/${kind.value}/pods`, { params })
    ])
    detail.value = detailRes.data
    pods.value = podsRes.data || []
    scaleReplicas.value = detail.value?.replicas ?? detail.value?.readyReplicas ?? 1
  } catch (error) {
    errorText.value = error.response?.data?.message || '工作负载详情加载失败'
  } finally {
    loading.value = false
  }
}

// --- 工作负载操作 ---

// 打开 YAML 编辑
const openYamlEdit = async () => {
  const params = { clusterName: clusterName.value, namespace: namespace.value, name: name.value }
  try {
    const res = await request.get(`/k8s/${kind.value}/yaml`, { params })
    workloadYamlContent.value = res.data || ''
    workloadYamlVisible.value = true
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '获取 YAML 失败')
  }
}

// 确认 YAML 更新
const confirmWorkloadYamlUpdate = async () => {
  yamlSaving.value = true
  const data = {
    clusterName: clusterName.value,
    namespace: namespace.value,
    name: name.value,
    yaml: workloadYamlContent.value
  }
  try {
    if (kind.value === 'cronjob') {
      await updateCronJobYAML(data)
    } else {
      await request.post(`/k8s/${kind.value}/yaml/update`, data)
    }
    ElMessage.success('更新成功')
    workloadYamlVisible.value = false
    loadDetail()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || 'YAML 更新失败')
  } finally {
    yamlSaving.value = false
  }
}

// 重启工作负载
const handleRestart = async () => {
  try {
    await ElMessageBox.confirm(`确定要重启 ${kindLabel.value} "${name.value}" 吗？`, '重启确认', { type: 'warning' })
  } catch { return }
  try {
    await request.post(`/k8s/${kind.value}/restart`, {
      clusterName: clusterName.value,
      namespace: namespace.value,
      name: name.value
    })
    ElMessage.success('重启指令已发送')
    loadDetail()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '重启失败')
  }
}

// 删除工作负载
const handleDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除 ${kindLabel.value} "${name.value}" 吗？此操作不可恢复。`, '删除确认', { type: 'error' })
  } catch { return }
  try {
    const data = { clusterName: clusterName.value, namespace: namespace.value, name: name.value }
    switch (kind.value) {
      case 'deployment':
        await deleteDeployment(data)
        break
      case 'statefulset':
        await deleteStatefulSet(data)
        break
      case 'daemonset':
        await deleteDaemonSet(data)
        break
      case 'job':
        await deleteJob(data)
        break
      case 'cronjob':
        await deleteCronJob(data)
        break
      default:
        await request.delete(`/k8s/${kind.value}/delete`, { data })
    }
    ElMessage.success('删除成功')
    router.back()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '删除失败')
  }
}

// 打开扩缩容弹窗
const openScaleDialog = () => {
  scaleReplicas.value = detail.value?.replicas ?? detail.value?.readyReplicas ?? 1
  scaleVisible.value = true
}

// 确认扩缩容
const confirmScale = async () => {
  scaleSaving.value = true
  const data = {
    clusterName: clusterName.value,
    namespace: namespace.value,
    name: name.value,
    replicas: scaleReplicas.value
  }
  try {
    if (kind.value === 'deployment') {
      await scaleDeployment(data)
    } else if (kind.value === 'statefulset') {
      await scaleStatefulSet(data)
    }
    ElMessage.success('扩缩容指令已发送')
    scaleVisible.value = false
    loadDetail()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '扩缩容失败')
  } finally {
    scaleSaving.value = false
  }
}

// --- Pod 操作 ---

// Pod YAML
const openPodYaml = async (row) => {
  try {
    const res = await request.get('/k8s/pod/yaml', {
      params: { clusterName: clusterName.value, namespace: namespace.value, name: row.name }
    })
    podYamlContent.value = res.data?.yaml || res.data || ''
    podYamlVisible.value = true
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '获取 Pod YAML 失败')
  }
}

// Pod 终端
const openTerminal = async (row) => {
  currentPod.value = row
  try {
    const shellRes = await request.get('/k8s/pod/detect-shell', {
      params: { clusterName: clusterName.value, namespace: namespace.value, pod: row.name }
    })
    detectedShell.value = shellRes.data?.recommendedShell || 'sh'
  } catch {
    detectedShell.value = 'sh'
  }
  terminalWsUrl.value = `/api/v1/k8s/pod/terminal?clusterName=${clusterName.value}&namespace=${namespace.value}&pod=${row.name}&shell=${detectedShell.value}`
  terminalVisible.value = true
}

const onTerminalError = (msg) => {
  ElMessage.error(msg || '终端连接异常')
}

// Pod 日志
const openPodLogs = (row) => {
  currentPod.value = row
  logContainer.value = row.containers?.[0]?.name || ''
  logTailLines.value = 500
  logContent.value = ''
  logVisible.value = true
  fetchLogs()
}

const fetchLogs = async () => {
  try {
    const res = await getPodLogs({
      clusterName: clusterName.value,
      namespace: namespace.value,
      name: currentPod.value.name,
      container: logContainer.value,
      tailLines: logTailLines.value
    })
    logContent.value = res.data?.logs || res.data || ''
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '获取日志失败')
  }
}

// Pod 重启（删除 Pod）
const handleDeletePod = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要重启 Pod "${row.name}" 吗？`, '重启确认', { type: 'warning' })
  } catch { return }
  try {
    await deletePod({
      clusterName: clusterName.value,
      namespace: namespace.value,
      name: row.name
    })
    ElMessage.success('Pod 已删除，将自动重建')
    loadDetail()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || 'Pod 重启失败')
  }
}

// 监听路由参数变化
watch([clusterName, namespace, name], () => { loadDetail() }, { immediate: true })
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
  margin-bottom: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-name {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

.kind-tag {
  font-weight: 500;
}

.header-right {
  display: flex;
  gap: 8px;
}

.label-tag {
  margin-right: 6px;
  margin-bottom: 4px;
}

.log-toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}
</style>
