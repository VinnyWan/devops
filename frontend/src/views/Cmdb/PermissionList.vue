<template>
  <div class="permission-page">
    <div class="left-panel">
      <div class="panel-header">
        <h3>分组</h3>
      </div>
      <div class="tree-wrap">
        <el-tree
          ref="treeRef"
          :data="treeData"
          node-key="id"
          highlight-current
          default-expand-all
          :props="{ label: 'name', children: 'children' }"
          @node-click="handleNodeClick"
        />
      </div>
    </div>
    <div class="right-panel">
      <div class="page-container">
        <div class="page-header">
          <h3>权限配置</h3>
          <el-button type="primary" @click="showCreateDialog">授予权限</el-button>
        </div>
        <div class="toolbar">
          <el-select v-model="filterUserId" placeholder="全部用户" clearable filterable style="width: 180px" @change="fetchData">
            <el-option v-for="u in userList" :key="u.id" :label="u.username" :value="u.id" />
          </el-select>
          <el-select v-model="filterPermission" placeholder="全部权限" clearable style="width: 150px" @change="fetchData">
            <el-option label="查看" value="view" />
            <el-option label="终端" value="terminal" />
            <el-option label="管理" value="admin" />
          </el-select>
        </div>
        <el-table :data="tableData" stripe v-loading="loading">
          <el-table-column prop="userId" label="用户 ID" width="100" />
          <el-table-column label="用户名" width="140">
            <template #default="{ row }">{{ getUsername(row.userId) }}</template>
          </el-table-column>
          <el-table-column label="分组" min-width="180">
            <template #default="{ row }">{{ getGroupName(row.hostGroupId) }}</template>
          </el-table-column>
          <el-table-column label="权限" width="120">
            <template #default="{ row }">
              <el-tag :type="permTagType(row.permission)" size="small">{{ permLabel(row.permission) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180" />
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-wrap">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            @size-change="fetchData"
            @current-change="fetchData"
          />
        </div>
      </div>

      <!-- Create dialog -->
      <el-dialog v-model="createDialogVisible" title="授予权限" width="500px" destroy-on-close>
        <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="100px">
          <el-form-item label="用户" prop="userId">
            <el-select v-model="createForm.userId" placeholder="选择用户" filterable style="width: 100%">
              <el-option v-for="u in userList" :key="u.id" :label="u.username" :value="u.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="分组" prop="hostGroupId">
            <el-tree-select
              v-model="createForm.hostGroupId"
              :data="treeData"
              :props="{ label: 'name', children: 'children', value: 'id' }"
              placeholder="选择分组"
              check-strictly
              filterable
              style="width: 100%"
              @change="handleGroupSelectChange"
            />
          </el-form-item>
          <el-form-item label="权限" prop="permissions">
            <el-checkbox-group v-model="createForm.permissions">
              <el-checkbox label="view">查看 (view)</el-checkbox>
              <el-checkbox label="terminal">终端 (terminal)</el-checkbox>
              <el-checkbox label="admin">管理 (admin)</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item v-if="groupHostCount >= 0" label="">
            <span class="host-count-tip">此权限将影响 {{ groupHostCount }} 台主机</span>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleCreateSubmit" :loading="submitting">确定</el-button>
        </template>
      </el-dialog>

      <!-- Edit dialog -->
      <el-dialog v-model="editDialogVisible" title="编辑权限" width="400px" destroy-on-close>
        <el-form :model="editForm" :rules="editRules" ref="editFormRef" label-width="80px">
          <el-form-item label="权限" prop="permission">
            <el-select v-model="editForm.permission" style="width: 100%">
              <el-option label="查看 (view)" value="view" />
              <el-option label="终端 (terminal)" value="terminal" />
              <el-option label="管理 (admin)" value="admin" />
            </el-select>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleEditSubmit" :loading="submitting">确定</el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPermissionList, createPermission, updatePermission, deletePermission, getGroupHostCount } from '@/api/cmdb/permission'
import { getGroupTree } from '@/api/cmdb/group'
import { getUserList } from '@/api/system'

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const filterUserId = ref('')
const filterPermission = ref('')
const selectedGroupId = ref(0)

const treeData = ref([])
const userList = ref([])
const groupFlatMap = ref({})

const createDialogVisible = ref(false)
const editDialogVisible = ref(false)
const submitting = ref(false)
const createFormRef = ref()
const editFormRef = ref()
const groupHostCount = ref(-1)

const createForm = reactive({
  userId: '',
  hostGroupId: '',
  permissions: []
})

const editForm = reactive({
  id: 0,
  permission: ''
})

const createRules = {
  userId: [{ required: true, message: '请选择用户', trigger: 'change' }],
  hostGroupId: [{ required: true, message: '请选择分组', trigger: 'change' }],
  permissions: [{ required: true, type: 'array', min: 1, message: '请至少选择一个权限', trigger: 'change' }]
}

const editRules = {
  permission: [{ required: true, message: '请选择权限', trigger: 'change' }]
}

const fetchTree = async () => {
  try {
    const res = await getGroupTree()
    treeData.value = res.data || []
    flattenTree(treeData.value)
  } catch (e) {
    console.error('fetch tree:', e)
  }
}

const flattenTree = (nodes, path = '') => {
  for (const node of nodes) {
    const currentPath = path ? `${path} / ${node.name}` : node.name
    groupFlatMap.value[node.id] = { ...node, path: currentPath }
    if (node.children && node.children.length) {
      flattenTree(node.children, currentPath)
    }
  }
}

const fetchUsers = async () => {
  try {
    const res = await getUserList({ page: 1, pageSize: 200 })
    userList.value = res.data?.list || res.data || []
  } catch (e) {
    console.error('fetch users:', e)
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (filterUserId.value) params.userId = filterUserId.value
    if (filterPermission.value) params.permission = filterPermission.value
    if (selectedGroupId.value) params.hostGroupId = selectedGroupId.value

    const res = await getPermissionList(params)
    tableData.value = res.data || []
    total.value = res.total || 0
  } catch (e) {
    ElMessage.error('获取权限列表失败')
  } finally {
    loading.value = false
  }
}

const handleNodeClick = (data) => {
  selectedGroupId.value = data.id
  page.value = 1
  fetchData()
}

const getUsername = (userId) => {
  const u = userList.value.find(u => u.id === userId)
  return u ? u.username : `用户${userId}`
}

const getGroupName = (groupId) => {
  const g = groupFlatMap.value[groupId]
  return g ? g.path : `分组${groupId}`
}

const permLabel = (p) => {
  const map = { view: '查看', terminal: '终端', admin: '管理' }
  return map[p] || p
}

const permTagType = (p) => {
  const map = { view: 'info', terminal: 'warning', admin: 'danger' }
  return map[p] || 'info'
}

const showCreateDialog = () => {
  createForm.userId = ''
  createForm.hostGroupId = selectedGroupId.value || ''
  createForm.permissions = []
  createDialogVisible.value = true
  if (createForm.hostGroupId) {
    handleGroupSelectChange(createForm.hostGroupId)
  } else {
    groupHostCount.value = -1
  }
}

const handleGroupSelectChange = async (val) => {
  if (!val) {
    groupHostCount.value = -1
    return
  }
  try {
    const res = await getGroupHostCount({ groupId: val })
    groupHostCount.value = res.data?.hostCount ?? -1
  } catch {
    groupHostCount.value = -1
  }
}

const handleCreateSubmit = async () => {
  try {
    await createFormRef.value.validate()
  } catch { return }

  submitting.value = true
  try {
    await createPermission({
      userId: createForm.userId,
      hostGroupId: createForm.hostGroupId,
      permissions: createForm.permissions
    })
    ElMessage.success('授权成功')
    createDialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '授权失败')
  } finally {
    submitting.value = false
  }
}

const handleEdit = (row) => {
  editForm.id = row.id
  editForm.permission = row.permission
  editDialogVisible.value = true
}

const handleEditSubmit = async () => {
  try {
    await editFormRef.value.validate()
  } catch { return }

  submitting.value = true
  try {
    await updatePermission({
      id: editForm.id,
      permission: editForm.permission
    })
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '更新失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该权限规则？', '确认', { type: 'warning' })
  } catch { return }

  try {
    await deletePermission({ id: row.id })
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '删除失败')
  }
}

onMounted(() => {
  fetchTree()
  fetchUsers()
  fetchData()
})
</script>

<style scoped>
.permission-page {
  display: flex;
  gap: 16px;
  height: calc(100vh - 120px);
}
.left-panel {
  width: 260px;
  min-width: 260px;
  background: #fff;
  border-radius: 4px;
  padding: 16px;
  overflow-y: auto;
}
.panel-header h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 500;
}
.tree-wrap {
  margin-top: 8px;
}
.right-panel {
  flex: 1;
  min-width: 0;
}
.page-container {
  background: #fff;
  border-radius: 4px;
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
.host-count-tip {
  color: #909399;
  font-size: 13px;
}
</style>
