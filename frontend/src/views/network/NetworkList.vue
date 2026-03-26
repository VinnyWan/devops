<script setup lang="ts">
import { h, ref, watch, onMounted } from 'vue'
import {
  NCard,
  NSpace,
  NDataTable,
  NButton,
  NTabs,
  NTabPane,
  useMessage,
  NPagination,
  NModal,
  NCode,
  NPopconfirm,
} from 'naive-ui'
import ClusterSelector from '@/components/ClusterSelector.vue'
import { useCluster } from '@/composables/useCluster'
import {
  k8sK8sServiceListPost,
  k8sK8sIngressListPost,
  k8sK8sServiceDeletePost,
  k8sK8sIngressDeletePost,
} from '@/api/generated/k8s-resource.api'

const message = useMessage()
const { currentClusterId } = useCluster()
const loading = ref(false)
const services = ref<any[]>([])
const ingresses = ref<any[]>([])
const svcTotal = ref(0)
const ingTotal = ref(0)
const page = ref(1)
const pageSize = ref(10)

// YAML 详情弹窗
const showYamlModal = ref(false)
const yamlContent = ref('')
const yamlTitle = ref('')
const yamlType = ref<'service' | 'ingress'>('service')

const serviceColumns = [
  { title: 'Service', key: 'name' },
  { title: '命名空间', key: 'namespace' },
  { title: '类型', key: 'type' },
  { title: 'ClusterIP', key: 'clusterIP' },
  { title: '端口', key: 'ports' },
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
            onClick: () => showYaml(row, 'service'),
          },
          { default: () => '详情' }
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDeleteService(row) },
          {
            trigger: () =>
              h(
                NButton,
                { size: 'small', quaternary: true, type: 'error' },
                { default: () => '删除' }
              ),
            default: () => `确认删除 Service「${row.name}」？`,
          }
        ),
      ]),
  },
]

const ingressColumns = [
  { title: 'Ingress', key: 'name' },
  { title: '命名空间', key: 'namespace' },
  { title: 'Hosts', key: 'hosts' },
  { title: '路径', key: 'paths' },
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
            onClick: () => showYaml(row, 'ingress'),
          },
          { default: () => '详情' }
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDeleteIngress(row) },
          {
            trigger: () =>
              h(
                NButton,
                { size: 'small', quaternary: true, type: 'error' },
                { default: () => '删除' }
              ),
            default: () => `确认删除 Ingress「${row.name}」？`,
          }
        ),
      ]),
  },
]

async function fetchData() {
  if (!currentClusterId.value) return

  loading.value = true
  try {
    const [svcRes, ingRes] = await Promise.all([
      k8sK8sServiceListPost({ clusterId: currentClusterId.value }),
      k8sK8sIngressListPost({ clusterId: currentClusterId.value }),
    ])
    const svcData = svcRes.data.data as any
    const ingData = ingRes.data.data as any

    services.value = Array.isArray(svcData) ? svcData : svcData?.items || []
    ingresses.value = Array.isArray(ingData) ? ingData : ingData?.items || []
    svcTotal.value = svcData?.total || services.value.length
    ingTotal.value = ingData?.total || ingresses.value.length
  } catch (error: any) {
    message.error(error.message || '获取网络资源失败')
  } finally {
    loading.value = false
  }
}

function showYaml(row: any, type: 'service' | 'ingress') {
  yamlTitle.value = `${type === 'service' ? 'Service' : 'Ingress'}: ${row.name}`
  yamlType.value = type
  yamlContent.value = row.yaml || generateYaml(row, type)
  showYamlModal.value = true
}

function generateYaml(row: any, type: 'service' | 'ingress'): string {
  if (type === 'service') {
    return `apiVersion: v1
kind: Service
metadata:
  name: ${row.name}
  namespace: ${row.namespace}
  labels:
    app: ${row.name}
spec:
  type: ${row.type || 'ClusterIP'}
  clusterIP: ${row.clusterIP || 'None'}
  ports:
    - port: 80
      targetPort: 8080
      protocol: TCP
  selector:
    app: ${row.name}`
  } else {
    return `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ${row.name}
  namespace: ${row.namespace}
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
    - host: ${row.hosts || 'example.com'}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: ${row.name}
                port:
                  number: 80`
  }
}

async function handleDeleteService(row: any) {
  try {
    await k8sK8sServiceDeletePost({
      clusterId: currentClusterId.value!,
      namespace: row.namespace,
      name: row.name,
    })
    message.success('删除 Service 成功')
    await fetchData()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

async function handleDeleteIngress(row: any) {
  try {
    await k8sK8sIngressDeletePost({
      clusterId: currentClusterId.value!,
      namespace: row.namespace,
      name: row.name,
    })
    message.success('删除 Ingress 成功')
    await fetchData()
  } catch (error: any) {
    message.error(error.message || '删除失败')
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
