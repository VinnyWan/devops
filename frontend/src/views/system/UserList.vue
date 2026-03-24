<script setup lang="ts">
import { ref, h } from 'vue'
import { NCard, NButton, NModal, NSpace, useMessage, useDialog } from 'naive-ui'
import CrudTable from '@/components/CrudTable.vue'
import CrudForm from '@/components/CrudForm.vue'
import SearchBar from '@/components/SearchBar.vue'
import { userListPost, userRegisterPost, userUpdatePost, userDeletePost, userAssignRolesPost } from '@/api/generated/user.api'
import { roleListPost } from '@/api/generated/role.api'
import { departmentListPost } from '@/api/generated/department.api'

const message = useMessage()
const dialog = useDialog()
const tableRef = ref()
const showModal = ref(false)
const showRoleModal = ref(false)
const editingUser = ref<any>(null)
const currentUserId = ref<number>(0)

const departments = ref<any[]>([])
const roles = ref<any[]>([])
const selectedRoles = ref<number[]>([])

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '用户名', key: 'username' },
  { title: '邮箱', key: 'email' },
  { title: '部门', key: 'department_name' },
  { title: '状态', key: 'status' },
  {
    title: '角色',
    key: 'roles',
    render: (row: any) => {
      return h(NButton, { size: 'small', onClick: () => handleAssignRoles(row) }, { default: () => '分配角色' })
    }
  }
]

const formFields = [
  { name: 'username', label: '用户名', type: 'text' as const, required: true },
  { name: 'email', label: '邮箱', type: 'text' as const, required: true },
  { name: 'password', label: '密码', type: 'password' as const, required: true },
  { name: 'department_id', label: '部门', type: 'select' as const, required: true, options: [] as any[] },
  { name: 'status', label: '状态', type: 'select' as const, required: true, options: [
    { label: '激活', value: 'active' },
    { label: '锁定', value: 'locked' }
  ]}
]

const fetchData = async ({ page, pageSize }: { page: number; pageSize: number }) => {
  const res = await userListPost({ page, pageSize })
  return { data: (res.data.data?.list || []) as any[], total: (res.data.data?.total || 0) as number }
}

const handleSearch = async (params: any) => {
  const res = await userListPost({ page: 1, pageSize: 10, keyword: params.keyword })
  return { data: res.data.data?.list || [], total: res.data.data?.total || 0 }
}

const handleCreate = () => {
  editingUser.value = null
  showModal.value = true
}

const handleEdit = (row: any) => {
  editingUser.value = row
  showModal.value = true
}

const handleDelete = (id: number) => {
  dialog.warning({
    title: '确认删除',
    content: '确定要删除该用户吗？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      await userDeletePost({ id })
      message.success('删除成功')
      tableRef.value?.loadData()
    }
  })
}

const handleSubmit = async (data: any) => {
  if (editingUser.value) {
    await userUpdatePost({ ...data, id: editingUser.value.id })
    message.success('更新成功')
  } else {
    await userRegisterPost(data)
    message.success('创建成功')
  }
  showModal.value = false
  tableRef.value?.loadData()
}

const handleAssignRoles = (row: any) => {
  currentUserId.value = row.id
  selectedRoles.value = row.role_ids || []
  showRoleModal.value = true
}

const handleRoleSubmit = async () => {
  await userAssignRolesPost({ user_id: currentUserId.value, role_ids: selectedRoles.value })
  message.success('角色分配成功')
  showRoleModal.value = false
  tableRef.value?.loadData()
}

const loadOptions = async () => {
  const [deptRes, roleRes] = await Promise.all([
    departmentListPost({}),
    roleListPost({})
  ])
  const deptList = (deptRes.data.data?.list || []) as any[]
  const roleList = (roleRes.data.data?.list || []) as any[]
  departments.value = deptList.map((d: any) => ({ label: d.name, value: d.id }))
  roles.value = roleList.map((r: any) => ({ label: r.name, value: r.id }))
  formFields[3].options = departments.value
}

loadOptions()
</script>

<template>
  <NCard title="用户管理">
    <template #header-extra>
      <NButton type="primary" @click="handleCreate">新增用户</NButton>
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

    <NModal v-model:show="showModal" preset="card" :title="editingUser ? '编辑用户' : '新增用户'" style="width: 600px">
      <CrudForm :fields="formFields" :initial-data="editingUser" :on-submit="handleSubmit" />
    </NModal>

    <NModal v-model:show="showRoleModal" preset="card" title="分配角色" style="width: 500px">
      <NSpace vertical>
        <div v-for="role in roles" :key="role.value">
          <label>
            <input type="checkbox" :value="role.value" v-model="selectedRoles" />
            {{ role.label }}
          </label>
        </div>
        <NButton type="primary" @click="handleRoleSubmit">确定</NButton>
      </NSpace>
    </NModal>
  </NCard>
</template>


