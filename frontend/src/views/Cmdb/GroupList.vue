<template>
  <div class="page-container">
    <div class="page-header">
      <h3>分组管理</h3>
      <el-button type="primary" @click="showCreateDialog(1, 0)">新增业务分组</el-button>
    </div>

    <el-empty v-if="!treeData.length" description="暂无分组数据，请新增业务分组" />

    <el-tree
      v-else
      :data="treeData"
      node-key="id"
      default-expand-all
      highlight-current
      :current-node-key="currentNodeId"
      :props="{ label: 'name', children: 'children' }"
      :expand-on-click-node="false"
      class="group-tree"
      @current-change="handleCurrentChange"
    >
      <template #default="{ data }">
        <div class="tree-node" @mouseenter="hoveredId = data.id" @mouseleave="hoveredId = null">
          <div class="node-left">
            <span class="level-dot" :class="'level-' + data.level"></span>
            <span class="node-name">{{ data.name }}</span>
            <el-tag size="small" round :type="levelTagType(data.level)" class="level-tag">
              {{ levelText(data.level) }}
            </el-tag>
          </div>
          <div class="node-actions" :class="{ visible: hoveredId === data.id || currentNodeId === data.id }">
            <el-button v-if="data.level < 3" link type="primary" size="small" @click.stop="showCreateDialog(data.level + 1, data.id)">
              <el-icon><Plus /></el-icon>新增下级
            </el-button>
            <el-button link type="primary" size="small" @click.stop="handleEdit(data)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button link type="danger" size="small" @click.stop="handleDelete(data)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
          </div>
        </div>
      </template>
    </el-tree>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑分组' : '新增分组'" width="480px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="90px">
        <el-form-item label="分组名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="层级">
          <el-tag round :type="levelTagType(form.level)">{{ levelText(form.level) }}</el-tag>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sortOrder" :min="0" style="width:100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import { getGroupTree, createGroup, updateGroup, deleteGroup } from '@/api/cmdb/group'
import { required } from '@/utils/validate'

const treeData = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const hoveredId = ref(null)
const currentNodeId = ref(null)
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
  formRef.value?.clearValidate()
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { id: row.id, name: row.name, level: row.level, parentId: row.parentId, sortOrder: row.sortOrder || 0 }
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

const handleCurrentChange = (data) => {
  currentNodeId.value = data?.id || null
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
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
  } finally {
    submitting.value = false
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
const levelTagType = (level) => ({ 1: '', 2: 'success', 3: 'warning' }[level] || 'info')

onMounted(fetchData)
</script>

<style scoped>
.group-tree {
  background: transparent;
}

.group-tree :deep(.el-tree-node__content) {
  height: 44px;
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
  padding-right: var(--spacing-sm);
}

.group-tree :deep(.el-tree-node__content:hover) {
  background: var(--color-bg-muted);
}

.tree-node {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-sm);
}

.node-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  min-width: 0;
}

.level-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.level-dot.level-1 { background: var(--color-primary); }
.level-dot.level-2 { background: var(--color-success); }
.level-dot.level-3 { background: var(--color-warning); }

.node-name {
  font-size: var(--font-size-base);
  font-weight: 500;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.level-tag {
  flex-shrink: 0;
}

.node-actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity var(--transition-fast);
  flex-shrink: 0;
}

.node-actions.visible {
  opacity: 1;
}

/* Show actions on tree node hover too */
.group-tree :deep(.el-tree-node__content:hover) .node-actions {
  opacity: 1;
}

@media (max-width: 768px) {
  .node-actions {
    opacity: 1;
  }
}
</style>
