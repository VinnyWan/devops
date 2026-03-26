<script setup lang="ts">
import { ref, computed, watch, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NSpace, NDataTable, NButton, useMessage, NPagination, NTabs, NTabPane, NSelect, NModal, NInputNumber, NTag } from 'naive-ui'
import ClusterSelector from '@/components/ClusterSelector.vue'
import YamlTerminalModal from '@/components/YamlTerminalModal.vue'
import { useCluster } from '@/composables/useCluster'
import { useYamlModal, resourceTypeNames } from '@/composables/useYamlModal'
import {
  k8sK8sDeploymentListPost,
  k8sK8sDeploymentScalePost,
  k8sK8sDeploymentYamlPost,
  k8sK8sDeploymentYamlUpdatePost,
  k8sK8sDeploymentDeletePost,
  k8sK8sNamespacesListPost
} from '@/api/generated/k8s-resource.api'
import {
  k8sK8sStatefulSetListPost,
  k8sK8sStatefulSetScalePost,
  k8sK8sStatefulSetYamlPost,
  k8sK8sStatefulSetYamlUpdatePost,
  k8sK8sStatefulSetDeletePost,
  k8sK8sDaemonSetListPost,
  k8sK8sDaemonSetYamlPost,
  k8sK8sDaemonSetYamlUpdatePost,
  k8sK8sDaemonSetDeletePost
} from '@/api/generated/k8s-workload.api'

const message = useMessage()
const { currentClusterId } = useCluster()
const loading = ref(false)
const workloads = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

// 资源类型
const resourceType = ref<'deployment' | 'statefulset' | 'daemonset'>('deployment')

// 命名空间
const selectedNamespace = ref<string>('all')
const namespaceOptions = ref<Array<{ label: string; value: string }>>([
  { label: '所有命名空间', value: 'all' }
])

// 伸缩弹窗
const showScaleModal = ref(false)
const scaleReplicas = ref(0)
const currentWorkload = ref<any>(null)

// 删除确认弹窗
const showDeleteModal = ref(false)

// YAML 弹窗 - 使用 composable
const yamlModal = useYamlModal({
  onSaveSuccess: () => fetchData()
})

const router = useRouter()

const columns = computed(() => [
  { title: '名称', key: 'name', width: 180 },
  {
    title: '标签',
    key: 'labels',
    width: 200,
    render: (row: any) => {
      const labels = Object.entries(row.labels || {}).slice(0, 2).map(([k, v]) => `${k}=${v}`)
      return h(NSpace, {}, {
        default: () => labels.map(l => h(NTag, { size: 'small' }, { default: () => l }))
      })
    }
  },
  {
    title: '容器组',
    key: 'podCount',
    width: 100,
    render: (row: any) => `${row.readyReplicas || 0}/${row.replicas || 0}`
  },
  {
    title: 'Request/Limits',
    key: 'resources',
    width: 200,
    render: (row: any) => {
      const rs = row.resourceSummary || {}
      return `CPU: ${rs.cpuRequest || '-'}/${rs.cpuLimit || '-'}\nMem: ${rs.memoryRequest || '-'}/${rs.memoryLimit || '-'}`
    }
  },
  {
    title: '镜像',
    key: 'images',
    width: 250,
    ellipsis: { tooltip: true },
    render: (row: any) => row.containers?.map((c: any) => c.image).join(', ') || '-'
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 180,
    render: (row: any) => new Date(row.createdAt).toLocaleString()
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    fixed: 'right' as const,
    render: (row: any) => {
      return h(NSpace, {}, {
        default: () => [
          h(NButton, {
            size: 'small',
            type: 'info',
            onClick: () => viewDetail(row)
          }, { default: () => '详情' }),
          resourceType.value !== 'daemonset' && h(NButton, {
            size: 'small',
            onClick: () => openScaleModal(row)
          }, { default: () => '伸缩' }),
          h(NButton, {
            size: 'small',
            onClick: () => openYamlModal(row)
          }, { default: () => 'YAML' }),
          h(NButton, {
            size: 'small',
            type: 'error',
            onClick: () => openDeleteModal(row)
          }, { default: () => '删除' })
        ]
      })
    }
  }
])

async function fetchNamespaces() {
  if (!currentClusterId.value) return
  try {
    const res = await k8sK8sNamespacesListPost({ clusterId: currentClusterId.value })
    const namespaces = (res.data as any).data || []
    namespaceOptions.value = [
      { label: '所有命名空间', value: 'all' },
      ...namespaces.map((ns: any) => ({ label: ns.name, value: ns.name }))
    ]
  } catch (error: any) {
    message.error('获取命名空间列表失败')
  }
}

async function fetchData() {
  if (!currentClusterId.value) return

  loading.value = true
  try {
    const params: any = {
      clusterId: currentClusterId.value,
      namespace: selectedNamespace.value === 'all' ? undefined : selectedNamespace.value,
      page: page.value,
      pageSize: pageSize.value
    }

    let res
    switch (resourceType.value) {
      case 'deployment':
        res = await k8sK8sDeploymentListPost(params)
        break
      case 'statefulset':
        res = await k8sK8sStatefulSetListPost(params)
        break
      case 'daemonset':
        res = await k8sK8sDaemonSetListPost(params)
        break
    }

    const data = res.data.data as any
    if (data?.items) {
      workloads.value = data.items
      total.value = data.total || data.items.length
    }
  } catch (error: any) {
    message.error(error.message || '获取工作负载失败')
  } finally {
    loading.value = false
  }
}

function openScaleModal(row: any) {
  currentWorkload.value = row
  scaleReplicas.value = row.replicas || 0
  showScaleModal.value = true
}

async function scaleWorkload() {
  try {
    const data = {
      clusterId: currentClusterId.value!,
      namespace: currentWorkload.value.namespace,
      name: currentWorkload.value.name,
      replicas: scaleReplicas.value
    }

    if (resourceType.value === 'deployment') {
      await k8sK8sDeploymentScalePost(data)
    } else if (resourceType.value === 'statefulset') {
      await k8sK8sStatefulSetScalePost(data)
    }

    message.success('伸缩成功')
    showScaleModal.value = false
    fetchData()
  } catch (error: any) {
    message.error('伸缩失败: ' + error.message)
  }
}

// 打开 YAML 弹窗
async function openYamlModal(row: any) {
  currentWorkload.value = row

  // 定义获取 YAML 的函数
  const fetchYaml = async () => {
    const params = {
      clusterId: currentClusterId.value!,
      namespace: row.namespace,
      name: row.name
    }

    let res
    switch (resourceType.value) {
      case 'deployment':
        res = await k8sK8sDeploymentYamlPost(params)
        break
      case 'statefulset':
        res = await k8sK8sStatefulSetYamlPost(params)
        break
      case 'daemonset':
        res = await k8sK8sDaemonSetYamlPost(params)
        break
    }
    return res?.data.data?.yaml || ''
  }

  // 使用 composable 打开弹窗
  await yamlModal.openWithFetch(
    fetchYaml,
    { type: resourceTypeNames[resourceType.value], namespace: row.namespace, name: row.name }
  )
}

// 保存 YAML
async function saveYaml() {
  // 定义保存函数
  const saveFunc = async (yaml: string) => {
    const params = {
      clusterId: currentClusterId.value!,
      namespace: currentWorkload.value.namespace,
      name: currentWorkload.value.name
    }
    const data = { yaml }

    switch (resourceType.value) {
      case 'deployment':
        await k8sK8sDeploymentYamlUpdatePost(params, data)
        break
      case 'statefulset':
        await k8sK8sStatefulSetYamlUpdatePost(params, data)
        break
      case 'daemonset':
        await k8sK8sDaemonSetYamlUpdatePost(params, data)
        break
    }
  }

  await yamlModal.save(saveFunc)
}

function openDeleteModal(row: any) {
  currentWorkload.value = row
  showDeleteModal.value = true
}

async function confirmDelete() {
  try {
    const data = {
      clusterId: currentClusterId.value!,
      namespace: currentWorkload.value.namespace,
      name: currentWorkload.value.name
    }

    switch (resourceType.value) {
      case 'deployment':
        await k8sK8sDeploymentDeletePost(data)
        break
      case 'statefulset':
        await k8sK8sStatefulSetDeletePost(data)
        break
      case 'daemonset':
        await k8sK8sDaemonSetDeletePost(data)
        break
    }

    message.success('删除成功')
    showDeleteModal.value = false
    fetchData()
  } catch (error: any) {
    message.error('删除失败: ' + error.message)
  }
}

function handlePageSizeChange(newSize: number) {
  pageSize.value = newSize
  page.value = 1
  fetchData()
}

function viewDetail(row: any) {
  router.push(`/workload/${resourceType.value}/${row.namespace}/${row.name}`)
}

function handlePageChange(newPage: number) {
  page.value = newPage
  fetchData()
}

watch(currentClusterId, () => {
  fetchNamespaces()
  fetchData()
})
watch(resourceType, fetchData)
watch(selectedNamespace, fetchData)

onMounted(() => {
  if (currentClusterId.value) {
    fetchNamespaces()
    fetchData()
  }
})
</script>

<template>
  <NCard title="工作负载">
    <template #header-extra>
      <NSpace>
        <NSelect
          v-model:value="selectedNamespace"
          :options="namespaceOptions"
          placeholder="选择命名空间"
          style="width: 200px"
        />
        <ClusterSelector />
        <NButton @click="fetchData">刷新</NButton>
      </NSpace>
    </template>

    <NSpace vertical :size="16">
      <NTabs v-model:value="resourceType" type="segment">
        <NTabPane name="deployment" tab="Deployment" />
        <NTabPane name="statefulset" tab="StatefulSet" />
        <NTabPane name="daemonset" tab="DaemonSet" />
      </NTabs>

      <NDataTable
        :columns="columns"
        :data="workloads"
        :loading="loading"
        :scroll-x="1400"
      />

      <div style="display: flex; justify-content: flex-end; margin-top: 16px">
        <NPagination
          v-model:page="page"
          v-model:page-size="pageSize"
          :item-count="total"
          :page-sizes="[10, 20, 50, 100]"
          show-size-picker
          show-quick-jumper
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </div>
    </NSpace>

    <!-- 伸缩弹窗 -->
    <NModal v-model:show="showScaleModal" preset="card" title="伸缩副本数" style="width: 500px">
      <NSpace vertical>
        <div>当前副本数: {{ currentWorkload?.replicas || 0 }}</div>
        <NInputNumber
          v-model:value="scaleReplicas"
          :min="0"
          placeholder="副本数"
          style="width: 100%"
        />
      </NSpace>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showScaleModal = false">取消</NButton>
          <NButton type="primary" @click="scaleWorkload">确认</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- YAML 弹窗 - 使用 composable 状态 -->
    <YamlTerminalModal
      v-model:show="yamlModal.show.value"
      v-model:content="yamlModal.content.value"
      :title="yamlModal.title.value"
      :loading="yamlModal.loading.value"
      @save="saveYaml"
    />

    <!-- 删除确认弹窗 -->
    <NModal v-model:show="showDeleteModal" preset="card" title="确认删除" style="width: 500px">
      <p>确定要删除 <strong>{{ currentWorkload?.name }}</strong> 吗？此操作不可逆。</p>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showDeleteModal = false">取消</NButton>
          <NButton type="error" @click="confirmDelete">确认删除</NButton>
        </NSpace>
      </template>
    </NModal>
  </NCard>
</template>
