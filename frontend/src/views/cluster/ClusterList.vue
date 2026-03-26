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
  NIcon,
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
  { label: '健康', value: 'healthy' },
  { label: '异常', value: 'unhealthy' },
  { label: '未知', value: 'unknown' },
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
  {
    title: '名称',
    key: 'name',
    width: 160,
    render: (row: Cluster) => h('span', { class: 'cluster-name' }, row.name),
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: Cluster) => h(StatusTag, { status: row.status }),
  },
  {
    title: '认证方式',
    key: 'authType',
    width: 110,
    render: (row: Cluster) =>
      h(NTag, { size: 'small', bordered: false, type: 'info' }, { default: () => row.authType }),
  },
  {
    title: '环境',
    key: 'env',
    width: 90,
    render: (row: Cluster) =>
      h(
        NTag,
        {
          size: 'small',
          bordered: false,
          type: row.env === 'prod' ? 'error' : row.env === 'test' ? 'warning' : 'info',
        },
        { default: () => row.env || '-' }
      ),
  },
  { title: 'K8s 版本', key: 'k8sVersion', width: 110 },
  { title: '节点数', key: 'nodeCount', width: 80 },
  { title: 'API Server', key: 'url', ellipsis: { tooltip: true } },
  { title: '备注', key: 'remark', ellipsis: { tooltip: true } },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    fixed: 'right' as const,
    render: (row: Cluster) => {
      const actions = [
        h(
          NButton,
          {
            size: 'small',
            quaternary: true,
            type: 'info',
            onClick: () => router.push(`/cluster/${row.id}`),
          },
          {
            icon: () =>
              h(NIcon, null, {
                default: () =>
                  h('svg', { viewBox: '0 0 24 24', fill: 'none' }, [
                    h('path', {
                      d: 'M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z',
                      fill: 'currentColor',
                    }),
                  ]),
              }),
            default: () => '详情',
          }
        ),
      ]
      if (canUpdateCluster.value) {
        actions.push(
          h(
            NButton,
            {
              size: 'small',
              quaternary: true,
              type: 'warning',
              onClick: () => openEdit(row),
            },
            {
              icon: () =>
                h(NIcon, null, {
                  default: () =>
                    h('svg', { viewBox: '0 0 24 24', fill: 'none' }, [
                      h('path', {
                        d: 'M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z',
                        fill: 'currentColor',
                      }),
                    ]),
                }),
              default: () => '编辑',
            }
          )
        )
      }
      if (canHealthCheckCluster.value) {
        actions.push(
          h(
            NButton,
            {
              size: 'small',
              quaternary: true,
              type: 'success',
              onClick: () => handleHealthCheck(row),
            },
            {
              icon: () =>
                h(NIcon, null, {
                  default: () =>
                    h('svg', { viewBox: '0 0 24 24', fill: 'none' }, [
                      h('path', {
                        d: 'M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z',
                        fill: 'currentColor',
                      }),
                    ]),
                }),
              default: () => '连通性',
            }
          )
        )
      }
      if (canDeleteCluster.value) {
        actions.push(
          h(
            NPopconfirm,
            { onPositiveClick: () => handleDelete(row) },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: 'small', quaternary: true, type: 'error' },
                  {
                    icon: () =>
                      h(NIcon, null, {
                        default: () =>
                          h('svg', { viewBox: '0 0 24 24', fill: 'none' }, [
                            h('path', {
                              d: 'M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z',
                              fill: 'currentColor',
                            }),
                          ]),
                      }),
                    default: () => '删除',
                  }
                ),
              default: () => `确认删除集群「${row.name}」？`,
            }
          )
        )
      }
      return h(NSpace, { size: 4 }, () => actions)
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

// 测试连通性
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
  <div class="cluster-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">集群管理</h1>
        <p class="page-subtitle">管理 Kubernetes 集群连接配置</p>
      </div>
      <div class="header-actions">
        <NButton v-if="canCreateCluster" type="primary" @click="openCreate">
          <template #icon>
            <NIcon>
              <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" fill="currentColor" />
              </svg>
            </NIcon>
          </template>
          新建集群
        </NButton>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-item">
        <label class="filter-label">搜索</label>
        <NInput
          v-model:value="searchName"
          placeholder="搜索集群名称"
          clearable
          class="filter-input"
        >
          <template #prefix>
            <NIcon>
              <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                  d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"
                  fill="currentColor"
                />
              </svg>
            </NIcon>
          </template>
        </NInput>
      </div>
      <div class="filter-item">
        <label class="filter-label">状态</label>
        <NSelect
          v-model:value="searchStatus"
          :options="statusOptions"
          placeholder="全部状态"
          clearable
          class="filter-select"
        />
      </div>
      <div class="filter-item filter-item--action">
        <NButton quaternary @click="fetchData">
          <template #icon>
            <NIcon>
              <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                  d="M17.65 6.35C16.2 4.9 14.21 4 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4-3.31 0-6-2.69-6-6s2.69-6 6-6c2.66 0 4.87 1.69 5.65 4zM12 14c-3.31 0-6-2.69-6-6s2.69-6 6-6 6 2.69 6 6-2.69 6-6-6z"
                  fill="currentColor"
                />
              </svg>
            </NIcon>
          </template>
          刷新
        </NButton>
      </div>
    </div>

    <!-- 数据表格 -->
    <NCard class="data-card" :bordered="false">
      <NDataTable
        :columns="columns"
        :data="filteredData"
        :loading="loading"
        :bordered="false"
        :row-key="(row: Cluster) => row.id"
        :scroll-x="1300"
      />
    </NCard>

    <!-- 新建/编辑集群弹窗 -->
    <NModal
      v-model:show="showModal"
      preset="card"
      :title="isEdit ? '编辑集群' : '新建集群'"
      style="width: 640px; max-width: calc(100vw - 32px)"
      :mask-closable="false"
    >
      <NForm label-placement="left" label-width="100" class="cluster-form">
        <NFormItem label="名称">
          <NInput v-model:value="form.name" placeholder="请输入集群名称" />
        </NFormItem>
        <NFormItem label="认证方式">
          <NRadioGroup v-model:value="form.authType">
            <NRadio value="kubeconfig">KubeConfig</NRadio>
            <NRadio value="token">Token</NRadio>
          </NRadioGroup>
        </NFormItem>
        <NFormItem label="环境">
          <NInput v-model:value="form.env" placeholder="如 prod / test / dev" />
        </NFormItem>
        <NFormItem label="API Server">
          <NInput v-model:value="form.url" placeholder="https://1.2.3.4:6443" />
        </NFormItem>
        <NFormItem v-if="form.authType === 'kubeconfig'" label="KubeConfig">
          <NInput
            v-model:value="form.kubeconfig"
            type="textarea"
            :rows="4"
            placeholder="粘贴 kubeconfig 内容"
          />
        </NFormItem>
        <NFormItem v-if="form.authType === 'token'" label="Token">
          <NInput
            v-model:value="form.token"
            type="textarea"
            :rows="3"
            placeholder="Bearer Token"
          />
        </NFormItem>
        <NFormItem label="CA 证书">
          <NInput
            v-model:value="form.caData"
            type="textarea"
            :rows="3"
            placeholder="CA 证书数据（可选）"
          />
        </NFormItem>
        <NFormItem label="标签">
          <NInput v-model:value="form.labels" placeholder='{"region":"shanghai"}' />
        </NFormItem>
        <NFormItem label="备注">
          <NInput
            v-model:value="form.remark"
            type="textarea"
            :rows="2"
            placeholder="备注信息"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">取消</NButton>
          <NButton type="primary" :loading="modalLoading" @click="handleSubmit">
            保存
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.cluster-page {
  padding: var(--spacing-lg);
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-xl);
}

.header-content {
  flex: 1;
}

.page-title {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  margin: 0 0 var(--spacing-xs);
}

.page-subtitle {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  margin: 0;
}

/* 筛选栏 */
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  margin-bottom: var(--spacing-lg);
  align-items: flex-end;
  border: 1px solid var(--border-light);
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.filter-item--action {
  margin-left: auto;
}

.filter-label {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  font-weight: var(--font-weight-medium);
}

.filter-input {
  width: 220px;
}

.filter-select {
  width: 160px;
}

/* 数据卡片 */
.data-card {
  border-radius: var(--radius-lg);
  overflow: hidden;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.data-card :deep(.n-card__content) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.data-card :deep(.n-data-table) {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.data-card :deep(.n-data-table-wrapper) {
  flex: 1;
  overflow: auto;
}

/* 集群名称样式 */
.cluster-name {
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
}

/* 表单样式 */
.cluster-form {
  padding: var(--spacing-md) 0;
}

/* 响应式 */
@media (max-width: 768px) {
  .cluster-page {
    padding: var(--spacing-md);
  }

  .page-header {
    flex-direction: column;
    gap: var(--spacing-md);
  }

  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-item {
    width: 100%;
  }

  .filter-input,
  .filter-select {
    width: 100%;
  }

  .filter-item--action {
    margin-left: 0;
    margin-top: var(--spacing-sm);
  }
}
</style>
