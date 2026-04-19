<template>
  <div class="page-container">
    <div class="search-bar">
      <ClusterSelector v-model="clusterName" />
      <NamespaceSelector v-model="namespace" :cluster-name="clusterName" style="margin-left: 12px" />
      <el-input v-model="keyword" placeholder="搜索" style="width: 200px; margin-left: 12px" clearable />
      <el-button type="primary" @click="fetchData" style="margin-left: 12px">查询</el-button>
      <el-button type="success" @click="handleCreate" style="margin-left: 12px">创建</el-button>
    </div>

    <el-table v-if="tableData.length || loading" :data="tableData" stripe v-loading="loading" style="margin-top: 16px">
      <el-table-column prop="name" label="名称" min-width="220" show-overflow-tooltip />
      <el-table-column prop="namespace" label="命名空间" width="160" v-if="!namespace" />
      <el-table-column prop="dataCount" label="数据项数量" width="120" />
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
    <el-empty v-if="!loading && !tableData.length" description="暂无 ConfigMap 数据" style="margin-top: 16px" />

    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      @current-change="fetchData"
      style="margin-top: 16px; justify-content: flex-end"
    />

    <!-- YAML 编辑弹窗 -->
    <el-dialog v-model="yamlVisible" :title="yamlTitle" width="900px" destroy-on-close top="3vh">
      <YamlEditor ref="yamlEditorRef" v-model="yamlContent" :readonly="false" :show-copy="false" min-height="600px" />
      <template #footer>
        <div style="display: flex; justify-content: space-between; width: 100%">
          <el-button @click="handleYamlCopy">{{ yamlCopyText }}</el-button>
          <div>
            <el-button @click="yamlVisible = false">取消</el-button>
            <el-button type="primary" @click="handleYamlSave">保存</el-button>
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
import { getConfigMapList, getConfigMapYAML, createConfigMap, updateConfigMapByYAML, deleteConfigMap } from '@/api/configmap'
import { formatTime } from '@/utils/format'

const clusterName = ref('')
const namespace = ref('')
const keyword = ref('')
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const loading = ref(false)

// YAML 弹窗
const yamlVisible = ref(false)
const yamlTitle = ref('YAML')
const yamlContent = ref('')
const yamlMode = ref('')
const yamlEditorRef = ref(null)
const yamlCopyText = ref('复制')
const currentRow = ref(null)

// 自动加载
watch(clusterName, (val) => {
  if (val) {
    page.value = 1
    fetchData()
  }
})

watch(namespace, () => {
  if (clusterName.value) {
    page.value = 1
    fetchData()
  }
})

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
    const res = await getConfigMapList({
      clusterName: clusterName.value,
      namespace: namespace.value,
      keyword: keyword.value,
      page: page.value,
      pageSize: pageSize.value
    })
    tableData.value = res.data?.items || res.data || []
    total.value = res.data?.total || res.total || 0
  } catch {
    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 删除
const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除 ConfigMap "${row.name}" 吗？此操作不可恢复。`, '删除确认', { type: 'warning' })
  await deleteConfigMap({ clusterName: clusterName.value, namespace: row.namespace, name: row.name })
  ElMessage.success('删除成功')
  fetchData()
}

// 创建
const handleCreate = () => {
  currentRow.value = { namespace: namespace.value }
  yamlMode.value = 'create'
  yamlTitle.value = '新建 ConfigMap'
  yamlContent.value = `apiVersion: v1
kind: ConfigMap
metadata:
  name: ""
  namespace: ${namespace.value || 'default'}
data:
  key: value
`
  yamlVisible.value = true
}

// YAML 查看/编辑
const handleYaml = async (row) => {
  currentRow.value = row
  yamlMode.value = 'edit'
  yamlTitle.value = `YAML - ${row.name}`
  try {
    const res = await getConfigMapYAML({ clusterName: clusterName.value, namespace: row.namespace, name: row.name })
    yamlContent.value = res.data?.yaml || ''
  } catch {
    yamlContent.value = ''
  }
  yamlVisible.value = true
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

// YAML 保存
const handleYamlSave = async () => {
  if (!yamlContent.value.trim()) {
    ElMessage.warning('YAML 内容不能为空')
    return
  }

  if (yamlMode.value === 'create') {
    await createConfigMap({ yaml: yamlContent.value, clusterName: clusterName.value, namespace: currentRow.value.namespace || namespace.value || 'default' })
    ElMessage.success('创建成功')
  } else {
    await updateConfigMapByYAML({
      clusterName: clusterName.value,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      yaml: yamlContent.value
    })
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
