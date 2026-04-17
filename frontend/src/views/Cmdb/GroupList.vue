<template>
  <div class="page-container">
    <div class="page-header">
      <h3>分组管理</h3>
      <el-button type="primary" @click="showCreateDialog(1, 0)">新增业务分组</el-button>
    </div>

    <el-tree :data="treeData" node-key="id" default-expand-all :props="{ label: 'name', children: 'children' }">
      <template #default="{ data }">
        <div class="tree-node">
          <span>
            {{ data.name }}
            <el-tag size="small" style="margin-left: 8px;">{{ levelText(data.level) }}</el-tag>
          </span>
          <span>
            <el-button v-if="data.level < 3" size="small" @click.stop="showCreateDialog(data.level + 1, data.id)">新增下级</el-button>
            <el-button size="small" @click.stop="handleEdit(data)">编辑</el-button>
            <el-button size="small" type="danger" @click.stop="handleDelete(data)">删除</el-button>
          </span>
        </div>
      </template>
    </el-tree>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑分组' : '新增分组'" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="90px">
        <el-form-item label="分组名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="层级"><el-input :model-value="levelText(form.level)" disabled /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sortOrder" :min="0" style="width:100%" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getGroupTree, createGroup, updateGroup, deleteGroup } from '@/api/cmdb/group'
import { required } from '@/utils/validate'

const treeData = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref()
const form = ref({ name: '', level: 1, parentId: 0, sortOrder: 0 })
const rules = { name: [required('请输入分组名称')] }

const fetchData = async () => {
  const res = await getGroupTree()
  treeData.value = res.data || []
}

const showCreateDialog = (level, parentId) => {
  isEdit.value = false
  form.value = { name: '', level, parentId, sortOrder: 0 }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { id: row.id, name: row.name, level: row.level, parentId: row.parentId, sortOrder: row.sortOrder || 0 }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    if (isEdit.value) {
      await updateGroup(form.value)
      ElMessage.success('更新成功')
    } else {
      await createGroup(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除分组 "${row.name}"？`, '提示', { type: 'warning' })
  try {
    await deleteGroup({ id: row.id })
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.message || '删除失败')
  }
}

const levelText = (level) => ({ 1: '业务', 2: '环境', 3: '地域/机房' }[level] || `L${level}`)

onMounted(fetchData)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.tree-node { width: 100%; display: flex; justify-content: space-between; align-items: center; padding-right: 12px; }
</style>
