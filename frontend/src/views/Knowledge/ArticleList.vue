<template>
  <div class="page-container" style="display: flex; gap: 20px">
    <!-- Category sidebar -->
    <div style="width: 220px; flex-shrink: 0; border-right: 1px solid var(--color-border); padding-right: 16px">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px">
        <strong>分类</strong>
        <el-button link type="primary" size="small" @click="showCreateCategory">+ 新建</el-button>
      </div>
      <el-tree :data="categoryTree" :props="{ children: 'children', label: 'name' }" node-key="id" highlight-current @node-click="handleCategoryClick" default-expand-all />
    </div>

    <!-- Main content -->
    <div style="flex: 1; min-width: 0">
      <div class="page-header">
        <h3>知识库</h3>
        <el-button type="primary" @click="showCreateArticle">新建文章</el-button>
      </div>

      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索文章" style="width: 250px" clearable @change="fetchArticles" />
        <el-button type="primary" @click="fetchArticles">搜索</el-button>
      </div>

      <el-table :data="articles" stripe v-loading="loading" @row-click="viewArticle">
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column prop="categoryName" label="分类" width="120" />
        <el-table-column prop="viewCount" label="阅读" width="80" />
        <el-table-column label="更新时间" width="180">
          <template #default="{ row }">{{ new Date(row.updatedAt).toLocaleString() }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click.stop="editArticle(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click.stop="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && !articles.length" description="暂无文章" />

      <div class="pagination-wrap" v-if="total > 0">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next" @size-change="fetchArticles" @current-change="fetchArticles" />
      </div>
    </div>

    <!-- Article View Dialog -->
    <el-dialog v-model="viewDialogVisible" :title="viewArticleData?.title" width="800px">
      <div v-html="viewArticleData?.contentHtml || ''" style="line-height: 1.8; max-height: 500px; overflow: auto" />
    </el-dialog>

    <!-- Article Edit Dialog -->
    <el-dialog v-model="editDialogVisible" :title="isEdit ? '编辑文章' : '新建文章'" width="900px" top="40px">
      <el-form ref="articleFormRef" :model="articleForm" :rules="articleRules" label-width="80px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="articleForm.title" placeholder="文章标题" />
        </el-form-item>
        <el-form-item label="分类">
          <el-tree-select v-model="articleForm.categoryId" :data="categoryTree" :props="{ children: 'children', label: 'name', value: 'id' }" check-strictly placeholder="选择分类" clearable style="width: 100%" />
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <div style="display: flex; gap: 12px; width: 100%">
            <el-input v-model="articleForm.content" type="textarea" :rows="20" placeholder="Markdown 内容" style="flex: 1" />
            <div style="flex: 1; border: 1px solid var(--color-border); border-radius: 6px; padding: 12px; max-height: 500px; overflow: auto; background: #fafafa" v-html="markdownPreview" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitArticle" :loading="submitting">保存</el-button>
      </template>
    </el-dialog>

    <!-- Category Edit Dialog -->
    <el-dialog v-model="catDialogVisible" :title="catParentId ? '添加子分类' : (editCatId ? '编辑分类' : '添加分类')" width="450px" append-to-body>
      <el-form ref="catFormRef" :model="catForm" :rules="{ name: [{ required: true, message: '必填' }] }" label-width="80px">
        <el-form-item label="名称" prop="name"><el-input v-model="catForm.name" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="catForm.sortOrder" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="catDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCategory">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listCategories, createCategory, updateCategory, deleteCategory, listArticles, getArticle, createArticle, updateArticle, deleteArticle } from '@/api/knowledge'

const categoryTree = ref([])
const keyword = ref('')
const selectedCategoryId = ref(null)
const articles = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

// View dialog
const viewDialogVisible = ref(false)
const viewArticleData = ref(null)

// Edit dialog
const editDialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const articleFormRef = ref()
const articleForm = reactive({ id: 0, title: '', categoryId: null, content: '' })
const articleRules = { title: [{ required: true, message: '必填' }], content: [{ required: true, message: '必填' }] }

// Simple markdown preview (backend handles markdown→HTML rendering for actual display)
const markdownPreview = computed(() => {
  if (!articleForm.content) return '<span style="color:#999">预览</span>'
  return `<pre style="white-space:pre-wrap;font-family:inherit;margin:0">${articleForm.content}</pre>`
})

// Category dialog
const catDialogVisible = ref(false)
const catParentId = ref(null)
const editCatId = ref(null)
const catFormRef = ref()
const catForm = reactive({ name: '', sortOrder: 0, parentId: null })

const fetchCategories = async () => {
  try { const res = await listCategories(); categoryTree.value = res.data || [] } catch { /* */ }
}

const fetchArticles = async () => {
  loading.value = true
  try {
    const params = { page: page.value, pageSize: pageSize.value, keyword: keyword.value }
    if (selectedCategoryId.value) params.categoryId = selectedCategoryId.value
    const res = await listArticles(params)
    articles.value = res.data || []
    total.value = res.total || 0
  } catch { ElMessage.error('获取文章失败') } finally { loading.value = false }
}

const handleCategoryClick = (data) => {
  selectedCategoryId.value = data.id; page.value = 1; fetchArticles()
}

const viewArticle = async (row) => {
  try {
    const res = await getArticle(row.id)
    viewArticleData.value = res.data
    viewDialogVisible.value = true
  } catch { ElMessage.error('获取文章内容失败') }
}

const showCreateArticle = () => { isEdit.value = false; Object.assign(articleForm, { id: 0, title: '', categoryId: selectedCategoryId.value, content: '' }); editDialogVisible.value = true }
const editArticle = async (row) => {
  try {
    const res = await getArticle(row.id)
    const a = res.data
    isEdit.value = true
    Object.assign(articleForm, { id: a.id, title: a.title, categoryId: a.categoryId, content: a.content })
    editDialogVisible.value = true
  } catch { ElMessage.error('获取文章失败') }
}

const submitArticle = async () => {
  const valid = await articleFormRef.value.validate().catch(() => false)
  if (!valid) return; submitting.value = true
  try {
    const data = { title: articleForm.title, categoryId: articleForm.categoryId, content: articleForm.content }
    if (isEdit.value) { await updateArticle(articleForm.id, data) } else { await createArticle(data) }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功'); editDialogVisible.value = false; fetchArticles()
  } catch { ElMessage.error('保存失败') } finally { submitting.value = false }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该文章？', '确认删除', { type: 'warning' })
  try { await deleteArticle(row.id); ElMessage.success('已删除'); fetchArticles() } catch { /* */ }
}

// Category actions
const showCreateCategory = () => { catParentId.value = null; editCatId.value = null; Object.assign(catForm, { name: '', sortOrder: 0, parentId: null }); catDialogVisible.value = true }

const submitCategory = async () => {
  const valid = await catFormRef.value.validate().catch(() => false)
  if (!valid) return
  try {
    const data = { name: catForm.name, sortOrder: catForm.sortOrder, parentId: catParentId.value || null }
    if (editCatId.value) { await updateCategory(editCatId.value, data) } else { await createCategory(data) }
    ElMessage.success('保存成功'); catDialogVisible.value = false; fetchCategories()
  } catch { ElMessage.error('保存失败') }
}

onMounted(async () => { await fetchCategories(); fetchArticles() })
</script>
