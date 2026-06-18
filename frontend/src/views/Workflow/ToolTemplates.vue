<template>
  <div class="page-container">
    <div class="page-header">
      <h3>工具模板</h3>
      <el-button type="primary" @click="showCreateTemplate">添加模板</el-button>
    </div>

    <!-- Category filter -->
    <div class="toolbar">
      <el-radio-group v-model="category" @change="fetchTemplates">
        <el-radio-button value="">全部</el-radio-button>
        <el-radio-button value="database">数据库</el-radio-button>
        <el-radio-button value="middleware">中间件</el-radio-button>
        <el-radio-button value="monitoring">监控</el-radio-button>
        <el-radio-button value="web">Web</el-radio-button>
        <el-radio-button value="cicd">CI/CD</el-radio-button>
        <el-radio-button value="logging">日志</el-radio-button>
      </el-radio-group>
    </div>

    <!-- Template list -->
    <el-table :data="templates" stripe v-loading="loading">
      <el-table-column prop="name" label="名称" width="150" />
      <el-table-column prop="category" label="分类" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ categoryLabel(row.category) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="showVersions(row)">版本</el-button>
          <el-button link type="primary" size="small" @click="editTemplate(row)">编辑</el-button>
          <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && !templates.length" description="暂无模板" />

    <!-- Template Edit Dialog -->
    <el-dialog v-model="editDialogVisible" :title="isEdit ? '编辑模板' : '添加模板'" width="500px">
      <el-form ref="templateFormRef" :model="templateForm" :rules="templateRules" label-width="80px">
        <el-form-item label="名称" prop="name"><el-input v-model="templateForm.name" /></el-form-item>
        <el-form-item label="分类" prop="category">
          <el-select v-model="templateForm.category" style="width: 100%">
            <el-option label="数据库" value="database" />
            <el-option label="中间件" value="middleware" />
            <el-option label="监控" value="monitoring" />
            <el-option label="Web" value="web" />
            <el-option label="CI/CD" value="cicd" />
            <el-option label="日志" value="logging" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="templateForm.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitTemplate" :loading="submitting">保存</el-button>
      </template>
    </el-dialog>

    <!-- Version Management Dialog -->
    <el-dialog v-model="versionDialogVisible" :title="`版本管理: ${selectedTemplate?.name || ''}`" width="800px">
      <el-button type="primary" size="small" @click="showCreateVersion" style="margin-bottom: 12px">添加版本</el-button>
      <el-table :data="versions" stripe v-loading="versionLoading">
        <el-table-column prop="version" label="版本" width="120" />
        <el-table-column label="推荐" width="80">
          <template #default="{ row: v }">
            <el-tag v-if="v.isRecommended" type="success" size="small">推荐</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="安装脚本" min-width="200" show-overflow-tooltip>
          <template #default="{ row: v }">{{ v.installScript?.substring(0, 80) }}...</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row: v }">
            <el-button link type="danger" size="small" @click="deleteVersion(v)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Version Edit Dialog -->
      <el-dialog v-model="versionFormVisible" title="添加版本" width="600px" append-to-body>
        <el-form ref="versionFormRef" :model="versionForm" :rules="versionRules" label-width="100px">
          <el-form-item label="版本号" prop="version"><el-input v-model="versionForm.version" placeholder="如 8.0.36" /></el-form-item>
          <el-form-item label="安装脚本" prop="installScript"><el-input v-model="versionForm.installScript" type="textarea" :rows="10" /></el-form-item>
          <el-form-item label="验证脚本"><el-input v-model="versionForm.verifyScript" type="textarea" :rows="3" /></el-form-item>
          <el-form-item label="设为推荐"><el-switch v-model="versionForm.isRecommended" /></el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="versionFormVisible = false">取消</el-button>
          <el-button type="primary" @click="submitVersion" :loading="versionSubmitting">保存</el-button>
        </template>
      </el-dialog>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listTemplates, getTemplate, saveTemplate, updateTemplate, deleteTemplate, listTemplateVersions, saveTemplateVersion, deleteTemplateVersion } from '@/api/tool'

const loading = ref(false)
const templates = ref([])
const category = ref('')

const editDialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const templateFormRef = ref()
const templateForm = reactive({ id: 0, name: '', category: 'database', description: '' })
const templateRules = { name: [{ required: true, message: '必填' }], category: [{ required: true, message: '必填' }] }

const versionDialogVisible = ref(false)
const versionLoading = ref(false)
const versions = ref([])
const selectedTemplate = ref(null)
const versionFormVisible = ref(false)
const versionSubmitting = ref(false)
const versionFormRef = ref()
const versionForm = reactive({ version: '', installScript: '', verifyScript: '', isRecommended: false })
const versionRules = { version: [{ required: true, message: '必填' }], installScript: [{ required: true, message: '必填' }] }

const categoryLabel = (cat) => {
  const map = { database: '数据库', middleware: '中间件', monitoring: '监控', web: 'Web', cicd: 'CI/CD', logging: '日志', other: '其他' }
  return map[cat] || cat
}

const fetchTemplates = async () => {
  loading.value = true
  try {
    const res = await listTemplates({ page: 1, pageSize: 100, category: category.value })
    templates.value = res.data || []
  } catch { ElMessage.error('获取模板失败') } finally { loading.value = false }
}

const showCreateTemplate = () => { isEdit.value = false; Object.assign(templateForm, { id: 0, name: '', category: 'database', description: '' }); editDialogVisible.value = true }
const editTemplate = (row) => { isEdit.value = true; Object.assign(templateForm, { ...row }); editDialogVisible.value = true }

const submitTemplate = async () => {
  const valid = await templateFormRef.value.validate().catch(() => false)
  if (!valid) return; submitting.value = true
  try {
    if (isEdit.value) { await updateTemplate(templateForm.id, templateForm) } else { await saveTemplate(templateForm) }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功'); editDialogVisible.value = false; fetchTemplates()
  } catch { ElMessage.error('保存失败') } finally { submitting.value = false }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该模板及所有版本？', '确认删除', { type: 'warning' })
  try { await deleteTemplate(row.id); ElMessage.success('已删除'); fetchTemplates() } catch { /* */ }
}

const showVersions = async (row) => {
  selectedTemplate.value = row; versionDialogVisible.value = true; versionLoading.value = true
  try { const res = await listTemplateVersions(row.id); versions.value = res.data || [] } catch { ElMessage.error('获取版本失败') } finally { versionLoading.value = false }
}

const showCreateVersion = () => { Object.assign(versionForm, { version: '', installScript: '', verifyScript: '', isRecommended: false }); versionFormVisible.value = true }

const submitVersion = async () => {
  const valid = await versionFormRef.value.validate().catch(() => false)
  if (!valid) return; versionSubmitting.value = true
  try {
    await saveTemplateVersion(selectedTemplate.value.id, versionForm)
    ElMessage.success('添加成功'); versionFormVisible.value = false; showVersions(selectedTemplate.value)
  } catch { ElMessage.error('保存失败') } finally { versionSubmitting.value = false }
}

const deleteVersion = async (v) => {
  await ElMessageBox.confirm('确定删除该版本？', '确认删除', { type: 'warning' })
  try { await deleteTemplateVersion(selectedTemplate.value.id, v.id); ElMessage.success('已删除'); showVersions(selectedTemplate.value) } catch { /* */ }
}

onMounted(fetchTemplates)
</script>
