<template>
  <div class="page-container">
    <div class="search-bar">
      <ClusterSelector v-model="clusterName" />
      <NamespaceSelector v-model="namespace" :cluster-name="clusterName" style="margin-left: 12px" v-show="activeTab === 'pvc'" />
      <el-button type="primary" @click="fetchData" style="margin-left: 12px">查询</el-button>
    </div>

    <el-tabs v-model="activeTab" @tab-change="handleTabChange" style="margin-top: 16px">
      <el-tab-pane label="StorageClass" name="storageclass" />
      <el-tab-pane label="PersistentVolume" name="pv" />
      <el-tab-pane label="PersistentVolumeClaim" name="pvc" />
    </el-tabs>

    <!-- StorageClass 表格 -->
    <el-table v-if="activeTab === 'storageclass' && (tableData.length || loading)" :data="tableData" stripe v-loading="loading">
      <el-table-column prop="name" label="名称" min-width="180">
        <template #default="{ row }">
          {{ row.name }}
          <el-tag v-if="row.isDefault" size="small" type="warning" style="margin-left: 6px">default</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="provisioner" label="Provisioner" min-width="200" show-overflow-tooltip />
      <el-table-column prop="reclaimPolicy" label="回收策略" width="120" />
      <el-table-column prop="volumeBindingMode" label="绑定模式" width="180" />
      <el-table-column label="允许扩容" width="100">
        <template #default="{ row }">{{ row.allowVolumeExpansion ? 'true' : 'false' }}</template>
      </el-table-column>
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleYaml(row)">YAML</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- PV 表格 -->
    <el-table v-if="activeTab === 'pv' && (tableData.length || loading)" :data="tableData" stripe v-loading="loading">
      <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
      <el-table-column prop="capacity" label="容量" width="100" />
      <el-table-column label="访问模式" width="120">
        <template #default="{ row }">{{ (row.accessModes || []).join(', ') || '-' }}</template>
      </el-table-column>
      <el-table-column prop="reclaimPolicy" label="回收策略" width="120" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="pvStatusType(row.status)" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="claim" label="Claim" min-width="160" show-overflow-tooltip />
      <el-table-column prop="storageClass" label="StorageClass" min-width="140" show-overflow-tooltip />
      <el-table-column prop="reason" label="原因" min-width="120" show-overflow-tooltip />
      <el-table-column label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleYaml(row)">YAML</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- PVC 表格 -->
    <el-table v-if="activeTab === 'pvc' && (tableData.length || loading)" :data="tableData" stripe v-loading="loading">
      <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
      <el-table-column prop="namespace" label="命名空间" width="140" v-if="!namespace" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="pvcStatusType(row.status)" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="volume" label="Volume" min-width="140" show-overflow-tooltip />
      <el-table-column prop="capacity" label="容量" width="100" />
      <el-table-column label="访问模式" width="120">
        <template #default="{ row }">{{ (row.accessModes || []).join(', ') || '-' }}</template>
      </el-table-column>
      <el-table-column prop="storageClass" label="StorageClass" min-width="140" show-overflow-tooltip />
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
      :description="activeTab === 'storageclass' ? '暂无 StorageClass 数据' : activeTab === 'pv' ? '暂无 PersistentVolume 数据' : '暂无 PersistentVolumeClaim 数据'"
      style="margin-top: 16px"
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
import { getStorageClassList, updateStorageClassYAML } from '@/api/storage'
import { getPVList, updatePVYAML } from '@/api/storage'
import { getPVCList, updatePVCYAML, deletePVC } from '@/api/storage'
import { formatTime } from '@/utils/format'

const clusterName = ref('')
const namespace = ref('')
const activeTab = ref('storageclass')
const tableData = ref([])
const loading = ref(false)

// YAML 弹窗
const yamlVisible = ref(false)
const yamlTitle = ref('YAML')
const yamlContent = ref('')
const yamlEditorRef = ref(null)
const yamlCopyText = ref('复制')
const currentRow = ref(null)

watch(clusterName, (val) => {
  if (val) fetchData()
})

watch(namespace, () => {
  if (clusterName.value && activeTab.value === 'pvc') fetchData()
})

onMounted(() => {
  if (clusterName.value) fetchData()
})

const fetchData = async () => {
  if (!clusterName.value) {
    tableData.value = []
    return
  }

  loading.value = true
  try {
    const params = { clusterName: clusterName.value }
    if (activeTab.value === 'pvc' && namespace.value) {
      params.namespace = namespace.value
    }
    let res
    switch (activeTab.value) {
      case 'storageclass':
        res = await getStorageClassList(params)
        break
      case 'pv':
        res = await getPVList(params)
        break
      case 'pvc':
        res = await getPVCList(params)
        break
    }
    tableData.value = res.data || []
  } catch {
    tableData.value = []
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  tableData.value = []
  fetchData()
}

// PV status tag type
const pvStatusType = (status) => {
  const map = { Available: 'success', Bound: 'primary', Released: 'warning', Failed: 'danger' }
  return map[status] || 'info'
}

// PVC status tag type
const pvcStatusType = (status) => {
  const map = { Bound: 'success', Pending: 'warning', Lost: 'danger' }
  return map[status] || 'info'
}

// YAML 查看/编辑
const handleYaml = async (row) => {
  currentRow.value = row
  yamlTitle.value = `YAML - ${row.name}`
  yamlCopyText.value = '复制'

  let resourceType = activeTab.value
  if (resourceType === 'storageclass') resourceType = 'storageclass'

  try {
    const params = { resourceType, clusterName: clusterName.value, name: row.name }
    if (row.namespace) params.namespace = row.namespace
    const res = await import('@/api/request').then(m => m.default.get('/k8s/resource/yaml', { params }))
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

  const row = currentRow.value
  try {
    switch (activeTab.value) {
      case 'storageclass':
        await updateStorageClassYAML({ clusterName: clusterName.value, name: row.name, yaml: yamlContent.value })
        break
      case 'pv':
        await updatePVYAML({ clusterName: clusterName.value, name: row.name, yaml: yamlContent.value })
        break
      case 'pvc':
        await updatePVCYAML({ clusterName: clusterName.value, namespace: row.namespace, name: row.name, yaml: yamlContent.value })
        break
    }
    ElMessage.success('保存成功')
    yamlVisible.value = false
    fetchData()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '保存失败')
  }
}

// 删除 PVC
const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除 PVC "${row.name}" 吗？此操作不可恢复。`, '删除确认', { type: 'warning' })
  await deletePVC({ clusterName: clusterName.value, namespace: row.namespace, name: row.name })
  ElMessage.success('删除成功')
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
