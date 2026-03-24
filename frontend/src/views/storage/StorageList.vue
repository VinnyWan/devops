<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { NCard, NSpace, NDataTable, NButton, useMessage, NPagination } from 'naive-ui'
import ClusterSelector from '@/components/ClusterSelector.vue'
import { useCluster } from '@/composables/useCluster'

const message = useMessage()
const { currentClusterId } = useCluster()
const loading = ref(false)
const volumes = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const columns = [
  { title: 'PV名称', key: 'name' },
  { title: '容量', key: 'capacity' },
  { title: '状态', key: 'status' },
  { title: '存储类', key: 'storageClass' },
]

async function fetchData() {
  if (!currentClusterId.value) {
    message.warning('请先选择集群')
    return
  }
  // TODO: 实现存储卷API调用
  message.info('存储管理功能开发中')
}

watch(currentClusterId, fetchData)
onMounted(fetchData)
</script>

<template>
  <NCard title="存储管理">
    <template #header-extra>
      <NSpace>
        <ClusterSelector />
        <NButton @click="fetchData">刷新</NButton>
      </NSpace>
    </template>
    <NSpace vertical :size="16">
      <NDataTable :columns="columns" :data="volumes" :loading="loading" />
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

