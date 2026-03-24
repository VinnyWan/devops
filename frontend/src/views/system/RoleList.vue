<script setup lang="ts">
import { ref, h, computed } from 'vue'
import { NCard, NButton, NModal, NSpace, NTree, useMessage, useDialog } from 'naive-ui'
import type { TreeOption } from 'naive-ui'
import CrudTable from '@/components/CrudTable.vue'
import CrudForm from '@/components/CrudForm.vue'
import SearchBar from '@/components/SearchBar.vue'
import { roleListPost, roleCreatePost, roleUpdatePost, roleDeletePost, roleAssignPermissionsPost } from '@/api/generated/role.api'
import { permissionAllPost } from '@/api/generated/permission.api'

const message = useMessage()
const dialog = useDialog()
const tableRef = ref()
const showModal = ref(false)
const showPermModal = ref(false)
const editingRole = ref<any>(null)
const currentRoleId = ref<number>(0)
const permissions = ref<any[]>([])
const selectedPermissions = ref<number[]>([])
const searchParams = ref<any>({})

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '角色名称', key: 'name' },
  { title: '类型', key: 'type' },
  { title: '描述', key: 'description' },
  {
    title: '权限',
    key: 'permissions',
    render: (row: any) => {
      return h(NButton, { size: 'small', onClick: () => handleAssignPermissions(row) }, { default: () => '分配权限' })
    }
  }
]

const formFields = [
  { name: 'name', label: '角色名称', type: 'text' as const, required: true },
  { name: 'type', label: '类型', type: 'select' as const, required: true, options: [
    { label: '系统角色', value: 'system' },
    { label: '自定义角色', value: 'custom' }
  ]},
  { name: 'description', label: '描述', type: 'textarea' as const }
]

const fetchData = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  const res = await roleListPost({ page, pageSize, ...searchParams.value })
  return { data: (res.data.data?.list || []) as any[], total: (res.data.data?.total || 0) as number }
}

const handleSearch = async (params: any) => {
  searchParams.value = params
  tableRef.value?.loadData()
}

const handleCreate = () => {
  editingRole.value = null
  showModal.value = true
}

const handleEdit = (row: any) => {
  editingRole.value = row
  showModal.value = true
}

const handleDelete = (id: number) => {
  dialog.warning({
    title: '确认删除',
    content: '确定要删除该角色吗？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      await roleDeletePost({ id })
      message.success('删除成功')
      tableRef.value?.loadData()
    }
  })
}

const handleSubmit = async (data: any) => {
  if (editingRole.value) {
    await roleUpdatePost({ ...data, id: editingRole.value.id })
    message.success('更新成功')
  } else {
    await roleCreatePost(data)
    message.success('创建成功')
  }
  showModal.value = false
  tableRef.value?.loadData()
}

const handleAssignPermissions = (row: any) => {
  currentRoleId.value = row.id
  selectedPermissions.value = row.permission_ids || []
  showPermModal.value = true
}

const handlePermSubmit = async () => {
  // 只提交叶子节点（权限ID），过滤掉父节点
  const permIds = selectedPermissions.value.filter((key: any) => typeof key === 'number')
  await roleAssignPermissionsPost({ role_id: currentRoleId.value, permission_ids: permIds })
  message.success('权限分配成功')
  showPermModal.value = false
  tableRef.value?.loadData()
}

const loadPermissions = async () => {
  const res = await permissionAllPost()
  permissions.value = (res.data.data?.list || []) as any[]
}

// 构建权限树状结构
const permissionTree = computed<TreeOption[]>(() => {
  const grouped = new Map<string, any[]>()

  permissions.value.forEach((perm: any) => {
    const resource = perm.resource || '其他'
    if (!grouped.has(resource)) {
      grouped.set(resource, [])
    }
    grouped.get(resource)!.push(perm)
  })

  return Array.from(grouped.entries()).map(([resource, perms]) => ({
    key: `resource-${resource}`,
    label: resource,
    children: perms.map((perm: any) => ({
      key: perm.id,
      label: `${perm.name} (${perm.action})`
    }))
  }))
})

loadPermissions()
</script>

<template>
  <NCard title="角色管理">
    <template #header-extra>
      <NButton type="primary" @click="handleCreate">新增角色</NButton>
    </template>
    <NSpace vertical :size="16">
      <SearchBar :filters="[]" @search="handleSearch" />
      <CrudTable
        ref="tableRef"
        :columns="columns"
        :fetch-data="fetchData"
        :on-edit="handleEdit"
        :on-delete="handleDelete"
        :permissions="{ update: true, delete: true }"
      />
    </NSpace>

    <NModal v-model:show="showModal" preset="card" :title="editingRole ? '编辑角色' : '新增角色'" style="width: 600px">
      <CrudForm :fields="formFields" :initial-data="editingRole" :on-submit="handleSubmit" />
    </NModal>

    <NModal v-model:show="showPermModal" preset="card" title="分配权限" style="width: 600px">
      <NSpace vertical>
        <NTree
          :data="permissionTree"
          checkable
          cascade
          :checked-keys="selectedPermissions"
          @update:checked-keys="selectedPermissions = $event"
          :default-expand-all="true"
        />
        <NButton type="primary" @click="handlePermSubmit">确定</NButton>
      </NSpace>
    </NModal>
  </NCard>
</template>

