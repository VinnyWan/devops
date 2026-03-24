<script setup lang="ts">
import { ref } from 'vue'
import { NCard, NButton, NModal, NSpace, useMessage, useDialog } from 'naive-ui'
import CrudTable from '@/components/CrudTable.vue'
import CrudForm from '@/components/CrudForm.vue'
import SearchBar from '@/components/SearchBar.vue'
import { departmentListPost, departmentCreatePost, departmentUpdatePost, departmentDeletePost } from '@/api/generated/department.api'

const message = useMessage()
const dialog = useDialog()
const tableRef = ref()
const showModal = ref(false)
const editingDept = ref<any>(null)
const departments = ref<any[]>([])
const searchParams = ref<any>({})

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '部门名称', key: 'name' },
  { title: '上级部门', key: 'parent_name' },
  { title: '成员数量', key: 'member_count' }
]

const formFields = [
  { name: 'name', label: '部门名称', type: 'text' as const, required: true },
  { name: 'parent_id', label: '上级部门', type: 'select' as const, options: [] as any[] },
  { name: 'description', label: '描述', type: 'textarea' as const }
]

const fetchData = async () => {
  const res = await departmentListPost(searchParams.value)
  return { data: (res.data.data?.list || []) as any[], total: (res.data.data?.total || 0) as number }
}

const handleSearch = async (params: any) => {
  searchParams.value = params
  tableRef.value?.loadData()
}

const handleCreate = () => {
  editingDept.value = null
  showModal.value = true
}

const handleEdit = (row: any) => {
  editingDept.value = row
  showModal.value = true
}

const handleDelete = (id: number) => {
  dialog.warning({
    title: '确认删除',
    content: '确定要删除该部门吗？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      await departmentDeletePost({ id })
      message.success('删除成功')
      tableRef.value?.loadData()
    }
  })
}

const handleSubmit = async (data: any) => {
  if (editingDept.value) {
    await departmentUpdatePost({ ...data, id: editingDept.value.id })
    message.success('更新成功')
  } else {
    await departmentCreatePost(data)
    message.success('创建成功')
  }
  showModal.value = false
  tableRef.value?.loadData()
}

const loadDepartments = async () => {
  const res = await departmentListPost({})
  const deptList = (res.data.data?.list || []) as any[]
  departments.value = deptList.map((d: any) => ({ label: d.name, value: d.id }))
  formFields[1].options = departments.value
}

loadDepartments()
</script>

<template>
  <NCard title="部门管理">
    <template #header-extra>
      <NButton type="primary" @click="handleCreate">新增部门</NButton>
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

    <NModal v-model:show="showModal" preset="card" :title="editingDept ? '编辑部门' : '新增部门'" style="width: 600px">
      <CrudForm :fields="formFields" :initial-data="editingDept" :on-submit="handleSubmit" />
    </NModal>
  </NCard>
</template>


