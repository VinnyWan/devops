<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { NCard, NSpace, NDataTable, NButton, useMessage, NPagination } from 'naive-ui'
import ClusterSelector from '@/components/ClusterSelector.vue'
import { useCluster } from '@/composables/useCluster'
import { k8sK8sNamespacesListPost } from '@/api/generated/k8s-resource.api'

const message = useMessage()
const { currentClusterId } = useCluster()
const loading = ref(false)
const namespaces = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const columns = [
  { title: '命名空间', key: 'name' },
  { title: '状态', key: 'status' },
  { title: '创建时间', key: 'createdAt' },
]

async function fetchData() {
  if (!currentClusterId.value) return

  loading.value = true
  try {
    const res = await k8sK8sNamespacesListPost({ clusterId: currentClusterId.value })
    const data = res.data.data as any
    if (Array.isArray(data)) {
      namespaces.value = data
      total.value = data.length
    } else if (data?.items) {
      namespaces.value = data.items
      total.value = data.total || data.items.length
    }
  } catch (error: any) {
    message.error(error.message || '获取命名空间失败')
  } finally {
    loading.value = false
  }
}

watch(currentClusterId, fetchData)
onMounted(fetchData)
</script>

<template>
  <NCard title="命名空间管理">
    <template #header-extra>
      <NSpace>
        <ClusterSelector />
        <NButton @click="fetchData">刷新</NButton>
      </NSpace>
    </template>
    <NSpace vertical :size="16">
      <NDataTable :columns="columns" :data="namespaces" :loading="loading" />
      <NPagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :item-count="total"
        :page-sizes="[10, 20, 50, 100]"
        show-size-picker
      />
    </NSpace>
  </NCard>
</template>

