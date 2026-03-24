<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import { NDataTable, NButton, NSpace, NPagination } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'

interface Props {
  columns: DataTableColumns
  fetchData: (params: { page: number; pageSize: number }) => Promise<{ data: any[]; total: number }>
  onEdit?: (row: any) => void
  onDelete?: (id: number) => void
  permissions?: {
    update?: boolean
    delete?: boolean
  }
}

const props = withDefaults(defineProps<Props>(), {
  permissions: () => ({ update: true, delete: true })
})

const data = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const selectedRowKeys = ref<number[]>([])

const loadData = async () => {
  loading.value = true
  try {
    const result = await props.fetchData({ page: page.value, pageSize: pageSize.value })
    data.value = result.data
    total.value = result.total
  } finally {
    loading.value = false
  }
}

const handlePageChange = (newPage: number) => {
  page.value = newPage
  loadData()
}

const handleEdit = (row: any) => {
  props.onEdit?.(row)
}

const handleDelete = (id: number) => {
  props.onDelete?.(id)
}

const actionColumn = {
  title: '操作',
  key: 'actions',
  render: (row: any) => {
    return h(NSpace, null, {
      default: () => [
        props.permissions?.update && h(NButton, { size: 'small', onClick: () => handleEdit(row) }, { default: () => '编辑' }),
        props.permissions?.delete && h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row.id) }, { default: () => '删除' })
      ].filter(Boolean)
    })
  }
}

const tableColumns = computed(() => [...props.columns, actionColumn])

onMounted(() => {
  loadData()
})

defineExpose({ loadData })
</script>

<template>
  <div>
    <NDataTable
      :columns="tableColumns"
      :data="data"
      :loading="loading"
      :row-key="(row: any) => row.id"
      v-model:checked-row-keys="selectedRowKeys"
    />
    <div style="margin-top: 16px; display: flex; justify-content: flex-end">
      <NPagination
        v-model:page="page"
        :page-size="pageSize"
        :item-count="total"
        @update:page="handlePageChange"
      />
    </div>
  </div>
</template>
