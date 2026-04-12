<template>
  <div class="page-container">
    <div class="search-bar">
      <ClusterSelector v-model="clusterName" />
      <NamespaceSelector v-model="namespace" :cluster-name="clusterName" style="margin-left: 12px" />
      <el-input v-model="keyword" placeholder="搜索关键词" style="width: 200px; margin-left: 12px" clearable />
      <el-button type="primary" @click="fetchData" style="margin-left: 12px">查询</el-button>
      <el-button type="success" @click="handleCreate" style="margin-left: 12px">新建</el-button>
    </div>

    <el-tabs v-model="activeTab" @tab-change="handleTabChange" style="margin-top: 16px">
      <el-tab-pane label="Deployment" name="deployment" />
      <el-tab-pane label="StatefulSet" name="statefulset" />
      <el-tab-pane label="DaemonSet" name="daemonset" />
      <el-tab-pane label="Job" name="job" />
      <el-tab-pane label="CronJob" name="cronjob" />
    </el-tabs>

    <!-- Deployment / StatefulSet / DaemonSet 表格 -->
    <el-table v-if="!isJobTab && !isCronJobTab" :data="tableData" stripe>
      <el-table-column label="名称" min-width="160">
        <template #default="{ row }">
          <el-link type="primary" @click="goDetail(row)">{{ row.name }}</el-link>
        </template>
      </el-table-column>
      <el-table-column prop="namespace" label="命名空间" v-if="!namespace" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <StatusTag :status="row.status" />
        </template>
      </el-table-column>
      <el-table-column label="镜像" min-width="200">
        <template #default="{ row }">{{ formatImages(row.containers) }}</template>
      </el-table-column>
      <el-table-column label="READY" width="100">
        <template #default="{ row }">
          <template v-if="activeTab === 'daemonset'">{{ row.readyNumber }}/{{ row.desiredNumber }}</template>
          <template v-else>{{ row.readyReplicas ?? 0 }}/{{ row.replicas ?? 0 }}</template>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="goDetail(row)">详情</el-button>
          <el-button link type="primary" size="small" @click="handleYaml(row)">YAML</el-button>
          <el-button link type="primary" size="small" @click="handleRestart(row)" v-if="activeTab === 'deployment'">重启</el-button>
          <el-button link type="primary" size="small" @click="handleScale(row)" v-if="activeTab !== 'daemonset'">扩缩容</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Job 表格 -->
    <el-table v-if="isJobTab" :data="tableData" stripe>
      <el-table-column label="名称" min-width="160">
        <template #default="{ row }">
          <el-link type="primary" @click="goDetail(row)">{{ row.name }}</el-link>
        </template>
      </el-table-column>
      <el-table-column prop="namespace" label="命名空间" v-if="!namespace" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <StatusTag :status="row.status" />
        </template>
      </el-table-column>
      <el-table-column label="完成度" width="100">
        <template #default="{ row }">{{ row.succeeded ?? 0 }}/{{ row.completions ?? '-' }}</template>
      </el-table-column>
      <el-table-column label="镜像" min-width="200">
        <template #default="{ row }">{{ formatImages(row.containers) }}</template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="goDetail(row)">详情</el-button>
          <el-button link type="primary" size="small" @click="handleYaml(row)">YAML</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- CronJob 表格 -->
    <el-table v-if="isCronJobTab" :data="tableData" stripe>
      <el-table-column label="名称" min-width="160">
        <template #default="{ row }">
          <el-link type="primary" @click="goDetail(row)">{{ row.name }}</el-link>
        </template>
      </el-table-column>
      <el-table-column prop="namespace" label="命名空间" v-if="!namespace" />
      <el-table-column prop="schedule" label="调度规则" width="140" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <StatusTag :status="row.status" />
        </template>
      </el-table-column>
      <el-table-column label="暂停" width="80">
        <template #default="{ row }">{{ row.suspend ? '是' : '否' }}</template>
      </el-table-column>
      <el-table-column label="镜像" min-width="200">
        <template #default="{ row }">{{ formatImages(row.containers) }}</template>
      </el-table-column>
      <el-table-column label="上次调度" width="180">
        <template #default="{ row }">{{ row.lastSchedule ? formatTime(row.lastSchedule) : '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="goDetail(row)">详情</el-button>
          <el-button link type="primary" size="small" @click="handleYaml(row)">YAML</el-button>
          <el-button link type="primary" size="small" @click="handleSuspend(row, !row.suspend)">{{ row.suspend ? '恢复' : '暂停' }}</el-button>
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

    <!-- 扩缩容弹窗 -->
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

    <!-- YAML 编辑弹窗 -->
    <el-dialog v-model="yamlVisible" :title="yamlTitle" width="900px" destroy-on-close top="3vh">
      <YamlEditor ref="yamlEditorRef" v-model="yamlContent" :readonly="yamlReadonly" :show-copy="false" min-height="600px" />
      <template #footer>
        <div style="display: flex; justify-content: space-between; width: 100%">
          <el-button @click="handleYamlCopy">{{ yamlCopyText }}</el-button>
          <div>
            <el-button @click="yamlVisible = false">取消</el-button>
            <el-button type="primary" @click="handleYamlSave" v-if="!yamlReadonly">保存</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ClusterSelector from '@/components/K8s/ClusterSelector.vue'
import NamespaceSelector from '@/components/K8s/NamespaceSelector.vue'
import StatusTag from '@/components/K8s/StatusTag.vue'
import YamlEditor from '@/components/K8s/YamlEditor.vue'
import { getDeploymentList, createDeployment, restartDeployment, scaleDeployment, deleteDeployment } from '@/api/deployment'
import { getStatefulSetList, createStatefulSet, deleteStatefulSet, scaleStatefulSet } from '@/api/statefulset'
import { getDaemonSetList, createDaemonSet, deleteDaemonSet } from '@/api/daemonset'
import { getJobList, createJob, deleteJob } from '@/api/job'
import { getCronJobList, createCronJob, updateCronJobYAML, suspendCronJob, deleteCronJob } from '@/api/cronjob'
import { formatTime } from '@/utils/format'
import request from '@/api/request'

const router = useRouter()

const clusterName = ref('')
const namespace = ref('')
const keyword = ref('')
const activeTab = ref('deployment')
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 扩缩容
const scaleVisible = ref(false)
const scaleReplicas = ref(1)
const currentRow = ref(null)

// YAML 弹窗
const yamlVisible = ref(false)
const yamlTitle = ref('YAML')
const yamlContent = ref('')
const yamlReadonly = ref(false)
const yamlMode = ref('') // 'view' | 'edit' | 'create'
const yamlEditorRef = ref(null)
const yamlCopyText = ref('复制')

const isJobTab = computed(() => activeTab.value === 'job')
const isCronJobTab = computed(() => activeTab.value === 'cronjob')

// 集群变化时自动加载数据
watch(clusterName, (val) => {
  if (val) fetchData()
})

// 页面加载时若集群已就绪则自动获取
onMounted(() => {
  if (clusterName.value) fetchData()
})

// 格式化镜像列表
const formatImages = (containers) => {
  if (!containers || !Array.isArray(containers)) return '-'
  return containers.map(c => c.image).join(', ')
}

// 路由跳转详情
const goDetail = (row) => {
  router.push(`/k8s/workload/${activeTab.value}/${clusterName.value}/${row.namespace}/${row.name}`)
}

// 拉取数据
const fetchData = async () => {
  const params = {
    clusterName: clusterName.value,
    namespace: namespace.value,
    keyword: keyword.value,
    page: page.value,
    pageSize: pageSize.value
  }
  let res
  switch (activeTab.value) {
    case 'deployment':
      res = await getDeploymentList(params)
      break
    case 'statefulset':
      res = await getStatefulSetList(params)
      break
    case 'daemonset':
      res = await getDaemonSetList(params)
      break
    case 'job':
      res = await getJobList(params)
      break
    case 'cronjob':
      res = await getCronJobList(params)
      break
  }
  tableData.value = res.data?.items || []
  total.value = res.data?.total || 0
}

const handleTabChange = () => {
  page.value = 1
  fetchData()
}

// 重启 Deployment
const handleRestart = async (row) => {
  await ElMessageBox.confirm('确认重启该 Deployment?', '提示', { type: 'warning' })
  await restartDeployment({ clusterName: clusterName.value, namespace: row.namespace, name: row.name })
  ElMessage.success('重启成功')
  fetchData()
}

// 扩缩容
const handleScale = (row) => {
  currentRow.value = row
  scaleReplicas.value = row.replicas || 1
  scaleVisible.value = true
}

const confirmScale = async () => {
  const data = {
    clusterName: clusterName.value,
    namespace: currentRow.value.namespace,
    name: currentRow.value.name,
    replicas: scaleReplicas.value
  }
  if (activeTab.value === 'deployment') await scaleDeployment(data)
  else if (activeTab.value === 'statefulset') await scaleStatefulSet(data)
  ElMessage.success('操作成功')
  scaleVisible.value = false
  fetchData()
}

// 删除
const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除?', '提示', { type: 'warning' })
  const data = { clusterName: clusterName.value, namespace: row.namespace, name: row.name }
  switch (activeTab.value) {
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
  }
  ElMessage.success('删除成功')
  fetchData()
}

// 暂停/恢复 CronJob
const handleSuspend = async (row, suspend) => {
  const action = suspend ? '暂停' : '恢复'
  await ElMessageBox.confirm(`确认${action}?`, '提示', { type: 'warning' })
  await suspendCronJob({ clusterName: clusterName.value, namespace: row.namespace, name: row.name, suspend })
  ElMessage.success(`${action}成功`)
  fetchData()
}

// YAML 复制
const handleYamlCopy = async () => {
  try {
    await navigator.clipboard.writeText(yamlContent.value)
    yamlCopyText.value = '已复制'
    setTimeout(() => { yamlCopyText.value = '复制' }, 2000)
  } catch {
    ElMessage.error('复制失败')
  }
}

// YAML 查看
const handleYaml = async (row) => {
  currentRow.value = row
  yamlMode.value = 'edit'
  yamlReadonly.value = false
  yamlTitle.value = `YAML - ${row.name}`
  try {
    const res = await request.get(`/k8s/${activeTab.value}/yaml`, {
      params: { clusterName: clusterName.value, namespace: row.namespace, name: row.name }
    })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch {
    yamlContent.value = ''
  }
  // Job 的 YAML 只读
  if (activeTab.value === 'job') {
    yamlReadonly.value = true
  }
  yamlVisible.value = true
}

// YAML 保存（编辑或新建）
const handleYamlSave = async () => {
  if (!yamlContent.value.trim()) {
    ElMessage.warning('YAML 内容不能为空')
    return
  }

  if (yamlMode.value === 'create') {
    // 新建模式
    const ns = currentRow.value.namespace || namespace.value
    switch (activeTab.value) {
      case 'deployment':
        await createDeployment({ yaml: yamlContent.value, clusterName: clusterName.value, namespace: ns })
        break
      case 'statefulset':
        await createStatefulSet({ yaml: yamlContent.value, clusterName: clusterName.value, namespace: ns })
        break
      case 'daemonset':
        await createDaemonSet({ yaml: yamlContent.value, clusterName: clusterName.value, namespace: ns })
        break
      case 'job':
        await createJob({ yaml: yamlContent.value }, { namespace: ns, clusterName: clusterName.value })
        break
      case 'cronjob':
        await createCronJob({ yaml: yamlContent.value }, { namespace: ns, clusterName: clusterName.value })
        break
    }
    ElMessage.success('创建成功')
  } else {
    // 编辑模式
    const data = {
      clusterName: clusterName.value,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      yaml: yamlContent.value
    }
    if (activeTab.value === 'cronjob') {
      await updateCronJobYAML(data)
    } else {
      await request.post(`/k8s/${activeTab.value}/yaml/update`, data)
    }
    ElMessage.success('保存成功')
  }
  yamlVisible.value = false
  fetchData()
}

// 新建
const handleCreate = () => {
  currentRow.value = { namespace: namespace.value }
  yamlMode.value = 'create'
  yamlReadonly.value = false
  yamlTitle.value = `新建 ${activeTab.value}`
  yamlContent.value = ''
  yamlVisible.value = true
}
</script>

<style scoped>
.page-container {
  padding: 20px;
}
.search-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}
</style>
