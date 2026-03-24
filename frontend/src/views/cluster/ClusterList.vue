<script setup lang="ts">
import { h, ref, computed, onMounted } from 'vue'
import {
  NCard,
  NDataTable,
  NButton,
  NTag,
  NSpace,
  NInput,
  NSelect,
  NModal,
  NForm,
  NFormItem,
  NPopconfirm,
  NRadioGroup,
  NRadio,
  useMessage,
} from 'naive-ui'
import { useRouter } from 'vue-router'
import {
  getClusterList,
  deleteCluster,
  updateCluster,
  createCluster,
  checkClusterHealth,
} from '@/api/cluster'
import type { Cluster, ClusterForm } from '@/types/cluster'
import StatusTag from '@/components/StatusTag.vue'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const message = useMessage()
const { hasPermission } = useAuth()
const loading = ref(false)
const clusters = ref<Cluster[]>([])
const canCreateCluster = computed(() => hasPermission('cluster:create'))
const canUpdateCluster = computed(() => hasPermission('cluster:update'))
const canDeleteCluster = computed(() => hasPermission('cluster:delete'))
const canHealthCheckCluster = computed(() => hasPermission('cluster:list'))

// 搜索条件
const searchName = ref('')
const searchStatus = ref<string | null>(null)

const statusOptions = [
  { label: '全部', value: '' },
  { label: 'healthy', value: 'healthy' },
  { label: 'unhealthy', value: 'unhealthy' },
  { label: 'unknown', value: 'unknown' },
]

// 新建/编辑弹窗
const showModal = ref(false)
const modalLoading = ref(false)
const isEdit = ref(false)
const editId = ref(0)

const defaultForm = (): ClusterForm => ({
  authType: 'kubeconfig',
  caData: '',
  env: '',
  kubeconfig: '',
  labels: '',
  name: '',
  remark: '',
  token: '',
  url: '',
})

const form = ref<ClusterForm>(defaultForm())

// 按 ID 升序排序 + 搜索过滤
const filteredData = computed(() => {
  let list = [...clusters.value].sort((a, b) => a.id - b.id)
  if (searchName.value) {
    const keyword = searchName.value.toLowerCase()
    list = list.filter((c) => c.name.toLowerCase().includes(keyword))
  }
  if (searchStatus.value) {
    list = list.filter((c) => c.status === searchStatus.value)
  }
  return list
})

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '名称', key: 'name', width: 160 },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: Cluster) => h(StatusTag, { status: row.status }),
  },
  { title: '认证方式', key: 'authType', width: 100 },
  {
    title: '环境',
    key: 'env',
    width: 80,
    render: (row: Cluster) =>
      h(
        NTag,
        { size: 'small', type: row.env === 'prod' ? 'error' : 'info' },
        { default: () => row.env },
      ),
  },
  { title: 'K8s 版本', key: 'k8sVersion', width: 110 },
  { title: '节点数', key: 'nodeCount', width: 80 },
  { title: 'API Server', key: 'url', ellipsis: { tooltip: true } },
  { title: '备注', key: 'remark', ellipsis: { tooltip: true } },
  {
    title: '操作',
    key: 'actions',
    width: 300,
    render: (row: Cluster) => {
      const actions = [
        h(
          NButton,
          { size: 'small', type: 'info', onClick: () => router.push(`/cluster/${row.id}`) },
          { default: () => '详情' },
        ),
      ]
      if (canUpdateCluster.value) {
        actions.push(
          h(
            NButton,
            { size: 'small', type: 'warning', onClick: () => openEdit(row) },
            { default: () => '编辑' },
          ),
        )
      }
      if (canHealthCheckCluster.value) {
        actions.push(
          h(
            NButton,
            { size: 'small', type: 'success', onClick: () => handleHealthCheck(row) },
            { default: () => '连通性' },
          ),
        )
      }
      if (canDeleteCluster.value) {
        actions.push(
          h(
            NPopconfirm,
            { onPositiveClick: () => handleDelete(row) },
            {
              trigger: () =>
                h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
              default: () => `确认删除集群「${row.name}」？`,
            },
          ),
        )
      }
      return h(NSpace, { size: 'small' }, () => actions)
    },
  },
]

async function fetchData() {
  loading.value = true
  try {
    clusters.value = await getClusterList({ page: 1, pageSize: 100 })
  } finally {
    loading.value = false
  }
}

// 新建集群
function openCreate() {
  isEdit.value = false
  editId.value = 0
  form.value = defaultForm()
  showModal.value = true
}

// 编辑集群
function openEdit(row: Cluster) {
  isEdit.value = true
  editId.value = row.id
  form.value = {
    authType: row.authType,
    caData: '',
    env: row.env,
    kubeconfig: '',
    labels: row.labels,
    name: row.name,
    remark: row.remark,
    token: '',
    url: row.url,
  }
  showModal.value = true
}

async function handleSubmit() {
  modalLoading.value = true
  try {
    if (isEdit.value) {
      await updateCluster({ id: editId.value, ...form.value })
      message.success('更新成功')
    } else {
      await createCluster(form.value)
      message.success('创建成功')
    }
    showModal.value = false
    await fetchData()
  } catch (e: unknown) {
    message.error((e as Error).message || (isEdit.value ? '更新失败' : '创建失败'))
  } finally {
    modalLoading.value = false
  }
}

// 测试连通性 - 直接显示结果，不显示 loading 弹框
async function handleHealthCheck(row: Cluster) {
  try {
    const health = await checkClusterHealth(row.id)
    if (health.healthy) {
      message.success(`集群「${row.name}」连通正常，状态：${health.status}`)
    } else {
      message.error(`集群「${row.name}」连通异常：${health.error || health.status}`)
    }
    await fetchData()
  } catch (e: unknown) {
    message.error((e as Error).message || '检测失败')
  }
}

// 删除集群
async function handleDelete(row: Cluster) {
  loading.value = true
  try {
    await deleteCluster(row.id)
    message.success('删除成功')
    await fetchData()
  } catch (e: unknown) {
    message.error((e as Error).message || '删除失败')
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <n-card title="集群管理">
    <template #header-extra>
      <n-space>
        <n-input
          v-model:value="searchName"
          placeholder="搜索集群名称"
          clearable
          style="width: 200px"
        />
        <n-select
          v-model:value="searchStatus"
          :options="statusOptions"
          placeholder="集群状态"
          clearable
          style="width: 150px"
        />
        <n-button v-if="canCreateCluster" type="primary" @click="openCreate">新建集群</n-button>
      </n-space>
    </template>
    <n-data-table
      :columns="columns"
      :data="filteredData"
      :loading="loading"
      :bordered="false"
      :row-key="(row: Cluster) => row.id"
    />
  </n-card>

  <!-- 新建/编辑集群弹窗 -->
  <n-modal
    v-model:show="showModal"
    preset="dialog"
    :title="isEdit ? '编辑集群' : '新建集群'"
    positive-text="保存"
    negative-text="取消"
    :loading="modalLoading"
    style="width: 640px"
    @positive-click="handleSubmit"
  >
    <n-form label-placement="left" label-width="90">
      <n-form-item label="名称">
        <n-input v-model:value="form.name" placeholder="集群名称" />
      </n-form-item>
      <n-form-item label="认证方式">
        <n-radio-group v-model:value="form.authType">
          <n-radio value="kubeconfig">KubeConfig</n-radio>
          <n-radio value="token">Token</n-radio>
        </n-radio-group>
      </n-form-item>
      <n-form-item label="环境">
        <n-input v-model:value="form.env" placeholder="如 prod / test / dev" />
      </n-form-item>
      <n-form-item label="API Server">
        <n-input v-model:value="form.url" placeholder="https://1.2.3.4:6443" />
      </n-form-item>
      <n-form-item v-if="form.authType === 'kubeconfig'" label="KubeConfig">
        <n-input
          v-model:value="form.kubeconfig"
          type="textarea"
          :rows="4"
          placeholder="粘贴 kubeconfig 内容"
        />
      </n-form-item>
      <n-form-item v-if="form.authType === 'token'" label="Token">
        <n-input v-model:value="form.token" type="textarea" :rows="3" placeholder="Bearer Token" />
      </n-form-item>
      <n-form-item label="CA 证书">
        <n-input
          v-model:value="form.caData"
          type="textarea"
          :rows="3"
          placeholder="CA 证书数据（可选）"
        />
      </n-form-item>
      <n-form-item label="标签">
        <n-input v-model:value="form.labels" placeholder='{"region":"shanghai"}' />
      </n-form-item>
      <n-form-item label="备注">
        <n-input v-model:value="form.remark" type="textarea" :rows="2" placeholder="备注信息" />
      </n-form-item>
    </n-form>
  </n-modal>
</template>
