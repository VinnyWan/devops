<template>
  <div class="page-container">
    <div class="search-bar">
      <ClusterSelector v-model="clusterName" />
      <NamespaceSelector v-model="namespace" :cluster-name="clusterName" style="margin-left: 12px" />
      <el-input v-model="keyword" placeholder="搜索" style="width: 200px; margin-left: 12px" clearable />
      <el-button type="primary" @click="fetchData" style="margin-left: 12px">查询</el-button>
      <el-button type="success" @click="handleCreate" style="margin-left: 12px">创建</el-button>
    </div>

    <el-tabs v-model="activeTab" @tab-change="handleTabChange" style="margin-top: 16px">
      <el-tab-pane label="Service" name="service" />
      <el-tab-pane label="Ingress" name="ingress" />
    </el-tabs>

    <!-- Service 表格 -->
    <el-table v-if="activeTab === 'service' && (tableData.length || loading)" :data="tableData" stripe v-loading="loading">
      <el-table-column label="名称" min-width="160">
        <template #default="{ row }">
          <el-link type="primary">{{ row.name }}</el-link>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="120" />
      <el-table-column prop="clusterIP" label="ClusterIP" width="140" />
      <el-table-column label="端口" min-width="140">
        <template #default="{ row }">{{ (row.ports || []).join(', ') }}</template>
      </el-table-column>
      <el-table-column label="目标端口" min-width="120">
        <template #default="{ row }">{{ (row.targetPort || []).join(', ') || '-' }}</template>
      </el-table-column>
      <el-table-column label="Endpoints" min-width="200">
        <template #default="{ row }">{{ (row.endpoints || []).join(', ') || '-' }}</template>
      </el-table-column>
      <el-table-column label="选择器" min-width="160">
        <template #default="{ row }">
          <template v-if="row.selector && Object.keys(row.selector).length">
            <el-tag v-for="(value, key) in row.selector" :key="key" size="small" style="margin: 2px">{{ key }}={{ value }}</el-tag>
          </template>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleYaml(row)">YAML</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Ingress 表格 -->
    <el-table v-if="activeTab === 'ingress' && (tableData.length || loading)" :data="tableData" stripe v-loading="loading">
      <el-table-column label="名称" min-width="160">
        <template #default="{ row }">
          <el-link type="primary">{{ row.name }}</el-link>
        </template>
      </el-table-column>
      <el-table-column label="IngressClass" width="120">
        <template #default="{ row }">{{ row.ingressClass || '-' }}</template>
      </el-table-column>
      <el-table-column label="地址" width="140">
        <template #default="{ row }">{{ row.address || '-' }}</template>
      </el-table-column>
      <el-table-column label="主机" min-width="160">
        <template #default="{ row }">{{ (row.hosts || []).join(', ') || '-' }}</template>
      </el-table-column>
      <el-table-column label="路径" width="100">
        <template #default="{ row }">{{ (row.paths || []).join(', ') || '-' }}</template>
      </el-table-column>
      <el-table-column label="后端Service" width="140">
        <template #default="{ row }">
          <span v-if="row.backendService">{{ row.backendService }}:{{ row.backendPort }}</span>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleYaml(row)">YAML</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty
      v-if="!loading && !tableData.length"
      :description="activeTab === 'service' ? '暂无 Service 数据' : '暂无 Ingress 数据'"
      style="margin-top: 16px"
    />

    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      @current-change="fetchData"
      style="margin-top: 16px; justify-content: flex-end"
    />

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
import { ref, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ClusterSelector from '@/components/K8s/ClusterSelector.vue'
import NamespaceSelector from '@/components/K8s/NamespaceSelector.vue'
import YamlEditor from '@/components/K8s/YamlEditor.vue'
import { getServiceList, createService, deleteService, getServiceYAML, updateServiceByYAML } from '@/api/service'
import { getIngressList, createIngress, deleteIngress, getIngressYAML, updateIngressByYAML } from '@/api/ingress'
import { formatTime } from '@/utils/format'

const clusterName = ref('')
const namespace = ref('')
const keyword = ref('')
const activeTab = ref('service')
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const loading = ref(false)

// YAML 弹窗
const yamlVisible = ref(false)
const yamlTitle = ref('YAML')
const yamlContent = ref('')
const yamlReadonly = ref(false)
const yamlMode = ref('')
const yamlEditorRef = ref(null)
const yamlCopyText = ref('复制')
const currentRow = ref(null)

// 自动加载：集群变化时触发
watch(clusterName, (val) => {
  if (val) {
    page.value = 1
    fetchData()
  }
})

// 自动加载：命名空间变化时触发
watch(namespace, () => {
  if (clusterName.value) {
    page.value = 1
    fetchData()
  }
})

// 页面加载时若集群已就绪则自动获取
onMounted(() => {
  if (clusterName.value) fetchData()
})

const fetchData = async () => {
  if (!clusterName.value) {
    tableData.value = []
    total.value = 0
    return
  }

  loading.value = true
  try {
    const params = {
      clusterName: clusterName.value,
      namespace: namespace.value,
      keyword: keyword.value,
      page: page.value,
      pageSize: pageSize.value
    }
    const res = activeTab.value === 'service'
      ? await getServiceList(params)
      : await getIngressList(params)
    tableData.value = res.data?.items || res.data || []
    total.value = res.data?.total || res.total || 0
  } catch {
    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  page.value = 1
  fetchData()
}

// 删除
const handleDelete = async (row) => {
  await ElMessageBox.confirm('确认删除?', '提示', { type: 'warning' })
  const data = { clusterName: clusterName.value, namespace: row.namespace, name: row.name }
  activeTab.value === 'service' ? await deleteService(data) : await deleteIngress(data)
  ElMessage.success('删除成功')
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

// 打开 YAML 编辑
const handleYaml = async (row) => {
  currentRow.value = row
  yamlMode.value = 'edit'
  yamlReadonly.value = false
  yamlTitle.value = `YAML - ${row.name}`
  try {
    const res = await (activeTab.value === 'service'
      ? getServiceYAML({ resourceType: 'service', clusterName: clusterName.value, namespace: row.namespace, name: row.name })
      : getIngressYAML({ resourceType: 'ingress', clusterName: clusterName.value, namespace: row.namespace, name: row.name }))
    yamlContent.value = res.data?.yaml || ''
  } catch {
    yamlContent.value = ''
  }
  yamlVisible.value = true
}

// 创建
const handleCreate = () => {
  currentRow.value = { namespace: namespace.value }
  yamlMode.value = 'create'
  yamlReadonly.value = false
  yamlTitle.value = `新建 ${activeTab.value === 'service' ? 'Service' : 'Ingress'}`
  yamlContent.value = ''
  yamlVisible.value = true
}

// YAML 保存
const handleYamlSave = async () => {
  if (!yamlContent.value.trim()) {
    ElMessage.warning('YAML 内容不能为空')
    return
  }

  if (yamlMode.value === 'create') {
    const ns = currentRow.value.namespace || namespace.value || 'default'
    if (activeTab.value === 'service') {
      await createService({ yaml: yamlContent.value, clusterName: clusterName.value, namespace: ns })
    } else {
      await createIngress({ yaml: yamlContent.value, clusterName: clusterName.value, namespace: ns })
    }
    ElMessage.success('创建成功')
  } else {
    const data = {
      clusterName: clusterName.value,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      yaml: yamlContent.value
    }
    if (activeTab.value === 'service') {
      await updateServiceByYAML(data)
    } else {
      await updateIngressByYAML(data)
    }
    ElMessage.success('保存成功')
  }
  yamlVisible.value = false
  fetchData()
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
