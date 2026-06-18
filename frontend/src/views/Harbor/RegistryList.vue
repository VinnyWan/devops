<template>
  <div class="page-container">
    <div class="page-header">
      <h3>镜像仓库</h3>
      <el-button type="primary" @click="showConfigDialog">管理仓库</el-button>
    </div>

    <!-- Harbor server selector -->
    <div class="toolbar">
      <el-select v-model="configId" placeholder="选择 Harbor 仓库" style="width: 220px" @change="fetchProjects">
        <el-option v-for="c in configs" :key="c.id" :label="c.name" :value="c.id" />
      </el-select>
      <el-input v-model="keyword" placeholder="搜索项目" style="width: 200px" clearable @change="fetchProjects" />
      <el-button type="primary" @click="fetchProjects">查询</el-button>
    </div>

    <!-- Project list -->
    <el-table :data="projects" stripe v-loading="loading" @row-click="showRepos">
      <el-table-column prop="name" label="项目名称" min-width="200" />
      <el-table-column label="访问级别" width="100">
        <template #default="{ row }">
          <el-tag :type="row.public ? 'success' : 'warning'">{{ row.public ? '公开' : '私有' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="repoCount" label="仓库数" width="100" />
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click.stop="showRepos(row)">查看仓库</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && !projects.length" description="暂无项目" />

    <!-- Repository Dialog -->
    <el-dialog v-model="repoDialogVisible" :title="`仓库: ${selectedProject?.name || ''}`" width="800px">
      <el-table :data="repos" stripe v-loading="repoLoading" max-height="400" @row-click="showArtifacts">
        <el-table-column prop="name" label="仓库名称" min-width="200" />
        <el-table-column prop="artifactCount" label="镜像数" width="100" />
        <el-table-column prop="pullCount" label="拉取次数" width="100" />
        <el-table-column label="操作" width="120">
          <template #default="{ row: r }">
            <el-button link type="primary" size="small" @click.stop="showArtifacts(r)">查看镜像</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- Artifact Dialog -->
    <el-dialog v-model="artifactDialogVisible" :title="`镜像: ${selectedRepo?.name || ''}`" width="900px">
      <el-table :data="artifacts" stripe v-loading="artifactLoading" max-height="400">
        <el-table-column label="Tag" min-width="150">
          <template #default="{ row: a }">
            <el-tag v-for="tag in (a.tags || [])" :key="tag.id" size="small" style="margin-right: 4px">{{ tag.name }}</el-tag>
            <span v-if="!a.tags?.length" style="color: #999">无 tag</span>
          </template>
        </el-table-column>
        <el-table-column label="大小" width="120">
          <template #default="{ row: a }">{{ formatSize(a.size) }}</template>
        </el-table-column>
        <el-table-column label="摘要" min-width="200">
          <template #default="{ row: a }">{{ a.digest?.substring(0, 19) }}...</template>
        </el-table-column>
        <el-table-column label="推送时间" width="180">
          <template #default="{ row: a }">{{ a.pushTime }}</template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ row: a }">
            <el-button link type="danger" size="small" @click="handleDeleteArtifact(a)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- Config Management Dialog -->
    <el-dialog v-model="configDialogVisible" title="Harbor 仓库管理" width="700px">
      <el-table :data="configs" stripe max-height="300">
        <el-table-column prop="name" label="名称" width="150" />
        <el-table-column prop="url" label="地址" min-width="200" />
        <el-table-column label="状态" width="100">
          <template #default="{ row: c }">
            <el-tag :type="c.status === 'connected' ? 'success' : 'danger'">{{ c.status === 'connected' ? '已连接' : '异常' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{ row: c }">
            <el-button link type="primary" size="small" @click="editConfig(c)">编辑</el-button>
            <el-button link type="danger" size="small" @click="deleteConfig(c)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button type="primary" style="margin-top: 12px" @click="showCreateConfig">添加仓库</el-button>
    </el-dialog>

    <!-- Config Edit Dialog -->
    <el-dialog v-model="formDialogVisible" :title="isEditConfig ? '编辑仓库' : '添加仓库'" width="500px" append-to-body>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="URL" prop="url"><el-input v-model="form.url" placeholder="https://harbor.example.com" /></el-form-item>
        <el-form-item label="用户名" prop="username"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="密码" prop="password"><el-input v-model="form.password" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitConfig" :loading="submitting">保存并测试</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listHarborConfigs, saveHarborConfig, updateHarborConfig, deleteHarborConfig, testHarborConnection, listProjects, listRepositories, listArtifacts, deleteArtifact } from '@/api/harbor'

// Config state
const configs = ref([])
const configId = ref(0)
const configDialogVisible = ref(false)
const formDialogVisible = ref(false)
const isEditConfig = ref(false)
const submitting = ref(false)
const formRef = ref()
const form = reactive({ id: 0, name: '', url: '', username: '', password: '' })
const rules = { name: [{ required: true, message: '必填' }], url: [{ required: true, message: '必填' }], username: [{ required: true, message: '必填' }], password: [{ required: true, message: '必填' }] }

// Project state
const loading = ref(false)
const keyword = ref('')
const projects = ref([])

// Repository state
const repoDialogVisible = ref(false)
const repoLoading = ref(false)
const repos = ref([])
const selectedProject = ref(null)

// Artifact state
const artifactDialogVisible = ref(false)
const artifactLoading = ref(false)
const artifacts = ref([])
const selectedRepo = ref(null)

const fetchConfigs = async () => {
  try { const res = await listHarborConfigs({ page: 1, pageSize: 100 }); configs.value = res.data || [] } catch { /* */ }
}

const fetchProjects = async () => {
  if (!configId.value) return
  loading.value = true
  try { const res = await listProjects({ configId: configId.value, keyword: keyword.value, page: 1, pageSize: 50 }); projects.value = res.data || [] } catch { ElMessage.error('获取项目列表失败') } finally { loading.value = false }
}

const showRepos = async (row) => {
  selectedProject.value = row
  repoDialogVisible.value = true; repoLoading.value = true
  try { const res = await listRepositories(row.name, { configId: configId.value, page: 1, pageSize: 50 }); repos.value = res.data || [] } catch { ElMessage.error('获取仓库列表失败') } finally { repoLoading.value = false }
}

const showArtifacts = async (row) => {
  selectedRepo.value = row
  artifactDialogVisible.value = true; artifactLoading.value = true
  try { const res = await listArtifacts(selectedProject.value.name, row.name, { configId: configId.value, page: 1, pageSize: 50 }); artifacts.value = res.data || [] } catch { ElMessage.error('获取镜像列表失败') } finally { artifactLoading.value = false }
}

const handleDeleteArtifact = async (a) => {
  const ref = a.tags?.[0]?.name || a.digest
  await ElMessageBox.confirm(`确定删除 ${ref}？`, '确认删除', { type: 'warning' })
  try {
    await deleteArtifact(selectedProject.value.name, selectedRepo.value.name, { reference: ref })
    ElMessage.success('已删除'); showArtifacts(selectedRepo.value)
  } catch { /* */ }
}

const showConfigDialog = () => { fetchConfigs(); configDialogVisible.value = true }
const showCreateConfig = () => { isEditConfig.value = false; Object.assign(form, { id: 0, name: '', url: '', username: '', password: '' }); formDialogVisible.value = true }
const editConfig = (c) => { isEditConfig.value = true; Object.assign(form, { ...c, password: '' }); formDialogVisible.value = true }

const submitConfig = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return; submitting.value = true
  try {
    if (isEditConfig.value) { await updateHarborConfig(form.id, form) } else { await saveHarborConfig(form) }
    ElMessage.success(isEditConfig.value ? '更新成功' : '创建成功'); formDialogVisible.value = false; fetchConfigs()
  } catch { ElMessage.error('保存失败') } finally { submitting.value = false }
}

const deleteConfig = async (c) => {
  await ElMessageBox.confirm('确定删除？', '确认删除', { type: 'warning' })
  try { await deleteHarborConfig(c.id); ElMessage.success('已删除'); fetchConfigs() } catch { /* */ }
}

const formatSize = (bytes) => {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + ' MB'
  return (bytes / 1073741824).toFixed(2) + ' GB'
}

onMounted(async () => { await fetchConfigs(); if (configs.value.length) { configId.value = configs.value[0].id; fetchProjects() } })
</script>
