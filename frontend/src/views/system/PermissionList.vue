<script setup lang="ts">
import { ref } from 'vue'
import { NCard, NSpace } from 'naive-ui'
import CrudTable from '@/components/CrudTable.vue'
import SearchBar from '@/components/SearchBar.vue'
import { permissionListPost } from '@/api/generated/permission.api'

const tableRef = ref()
const searchParams = ref<any>({})

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '权限名称', key: 'name' },
  { title: '资源', key: 'resource' },
  { title: '操作', key: 'action' },
  { title: '描述', key: 'description' }
]

const fetchData = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  const res = await permissionListPost({ page, pageSize, ...searchParams.value })
  return { data: (res.data.data?.list || []) as any[], total: (res.data.data?.total || 0) as number }
}

const handleSearch = async (params: any) => {
  searchParams.value = params
  tableRef.value?.loadData()
}
</script>

<template>
  <NCard title="权限管理">
    <NSpace vertical :size="16">
      <SearchBar :filters="[]" @search="handleSearch" />
      <CrudTable
        ref="tableRef"
        :columns="columns"
        :fetch-data="fetchData"
        :permissions="{ update: false, delete: false }"
      />
    </NSpace>
  </NCard>
</template>
