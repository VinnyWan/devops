<script setup lang="ts">
import { h, ref, watch, onMounted } from 'vue'
import {
  NCard,
  NSpace,
  NDataTable,
  NButton,
  useMessage,
  NPagination,
  NModal,
  NCode,
  NPopconfirm,
} from 'naive-ui'
import ClusterSelector from '@/components/ClusterSelector.vue'
import { useCluster } from '@/composables/useCluster'
import {
  k8sK8sConfigmapListPost,
  k8sK8sConfigmapDeletePost,
} from '@/api/generated/k8s-resource.api'

const message = useMessage()
const { currentClusterId } = useCluster()
const loading = ref(false)
const configmaps = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

// YAML 详情弹窗
const showYamlModal = ref(false)
const yamlContent = ref('')
const yamlTitle = ref('')

const columns = [
  { title: 'ConfigMap', key: 'name' },
  { title: '命名空间', key: 'namespace' },
  { title: '数据项', key: 'dataCount' },
  { title: '创建时间', key: 'createdAt' },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row: any) =>
      h(NSpace, { size: 'small' }, () => [
        h(
          NButton,
          {
            size: 'small',
            quaternary: true,
            type: 'info',
            onClick: () => showYaml(row),
          },
          { default: () => '详情' }
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDelete(row) },
          {
            trigger: () =>
              h(
                NButton,
                { size: 'small', quaternary: true, type: 'error' },
                { default: () => '删除' }
              ),
            default: () => `确认删除 ConfigMap「${row.name}」？`,
          }
        ),
      ]),
  },
]

async function fetchData() {
  if (!currentClusterId.value) return

  loading.value = true
  try {
    const res = await k8sK8sConfigmapListPost({ clusterId: currentClusterId.value })
    const data = res.data.data as any
    if (Array.isArray(data)) {
      configmaps.value = data
      total.value = data.length
    } else if (data?.items) {
      configmaps.value = data.items
      total.value = data.total || data.items.length
    }
  } catch (error: any) {
    message.error(error.message || '获取配置失败')
  } finally {
    loading.value = false
  }
}

function showYaml(row: any) {
  yamlTitle.value = `ConfigMap: ${row.name}`
  yamlContent.value = row.yaml || generateYaml(row)
  showYamlModal.value = true
}

function generateYaml(row: any): string {
  const dataEntries = row.data
    ? Object.entries(row.data)
        .map(([k, v]) => `  ${k}: |-\n    ${v}`)
        .join('\n')
    : '  {}'

  return `apiVersion: v1
kind: ConfigMap
metadata:
  name: ${row.name}
  namespace: ${row.namespace}
data:
${dataEntries}`
}

async function handleDelete(row: any) {
  try {
    await k8sK8sConfigmapDeletePost({
      clusterId: currentClusterId.value!,
      namespace: row.namespace,
      name: row.name,
    })
    message.success('删除 ConfigMap 成功')
    await fetchData()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

watch(currentClusterId, fetchData)
onMounted(fetchData)
</script>

<template>
  <NCard title="配置管理">
    <template #header-extra>
      <NSpace>
        <ClusterSelector />
        <NButton @click="fetchData">刷新</NButton>
      </NSpace>
    </template>
    <NSpace vertical :size="16">
      <NDataTable :columns="columns" :data="configmaps" :loading="loading" />
      <NPagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :item-count="total"
        :page-sizes="[10, 20, 50, 100]"
        show-size-picker
      />
    </NSpace>
  </NCard>

  <!-- YAML 详情弹窗 -->
  <NModal
    v-model:show="showYamlModal"
    preset="card"
    :title="yamlTitle"
    style="width: 800px; max-width: calc(100vw - 32px)"
  >
    <NCode :code="yamlContent" language="yaml" />
  </NModal>
</template>
