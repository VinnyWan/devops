<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { NCard, NSpace, NDataTable, NButton, NTabs, NTabPane, useMessage, NPagination } from 'naive-ui'
import ClusterSelector from '@/components/ClusterSelector.vue'
import { useCluster } from '@/composables/useCluster'
import { k8sK8sServiceListPost, k8sK8sIngressListPost } from '@/api/generated/k8s-resource.api'

const message = useMessage()
const { currentClusterId } = useCluster()
const loading = ref(false)
const services = ref<any[]>([])
const ingresses = ref<any[]>([])
const svcTotal = ref(0)
const ingTotal = ref(0)
const page = ref(1)
const pageSize = ref(10)

const serviceColumns = [
  { title: 'Service', key: 'name' },
  { title: '命名空间', key: 'namespace' },
  { title: '类型', key: 'type' },
  { title: 'ClusterIP', key: 'clusterIP' },
]

const ingressColumns = [
  { title: 'Ingress', key: 'name' },
  { title: '命名空间', key: 'namespace' },
  { title: 'Hosts', key: 'hosts' },
]

async function fetchData() {
  if (!currentClusterId.value) return

  loading.value = true
  try {
    const [svcRes, ingRes] = await Promise.all([
      k8sK8sServiceListPost({ clusterId: currentClusterId.value }),
      k8sK8sIngressListPost({ clusterId: currentClusterId.value })
    ])
    const svcData = svcRes.data.data as any
    const ingData = ingRes.data.data as any

    services.value = Array.isArray(svcData) ? svcData : (svcData?.items || [])
    ingresses.value = Array.isArray(ingData) ? ingData : (ingData?.items || [])
    svcTotal.value = svcData?.total || services.value.length
    ingTotal.value = ingData?.total || ingresses.value.length
  } catch (error: any) {
    message.error(error.message || '获取网络资源失败')
  } finally {
    loading.value = false
  }
}

watch(currentClusterId, fetchData)
onMounted(fetchData)
</script>

<template>
  <NCard title="网络管理">
    <template #header-extra>
      <NSpace>
        <ClusterSelector />
        <NButton @click="fetchData">刷新</NButton>
      </NSpace>
    </template>
    <NTabs type="line">
      <NTabPane name="service" tab="Service">
        <NSpace vertical :size="16">
          <NDataTable :columns="serviceColumns" :data="services" :loading="loading" />
          <NPagination
            v-model:page="page"
            v-model:page-size="pageSize"
            :item-count="svcTotal"
            :page-sizes="[10, 20, 50, 100]"
            show-size-picker
          />
        </NSpace>
      </NTabPane>
      <NTabPane name="ingress" tab="Ingress">
        <NSpace vertical :size="16">
          <NDataTable :columns="ingressColumns" :data="ingresses" :loading="loading" />
          <NPagination
            v-model:page="page"
            v-model:page-size="pageSize"
            :item-count="ingTotal"
            :page-sizes="[10, 20, 50, 100]"
            show-size-picker
          />
        </NSpace>
      </NTabPane>
    </NTabs>
  </NCard>
</template>

