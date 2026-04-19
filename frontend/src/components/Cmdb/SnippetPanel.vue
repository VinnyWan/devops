<template>
  <div class="snippet-panel" :class="{ collapsed: !expanded }">
    <div class="panel-toggle" @click="expanded = !expanded">
      <el-icon><component :is="expanded ? 'ArrowRight' : 'ArrowLeft'" /></el-icon>
      <span v-if="!expanded">片段</span>
    </div>
    <div v-if="expanded" class="panel-content">
      <div class="panel-header">
        <el-input v-model="keyword" placeholder="搜索..." size="small" clearable @input="handleSearch" style="margin-bottom: 8px;" />
        <el-button type="primary" size="small" @click="showCreate = true" style="width: 100%;">+ 新建</el-button>
      </div>
      <div class="snippet-list">
        <div v-for="s in snippets" :key="s.id" class="snippet-item" @click="insertSnippet(s)">
          <div class="snippet-name">{{ s.name }}</div>
          <div class="snippet-tags" v-if="s.tags">
            <el-tag v-for="tag in s.tags.split(',')" :key="tag" size="small" type="info" style="margin-right: 2px;">{{ tag.trim() }}</el-tag>
          </div>
          <pre class="snippet-preview">{{ s.content.substring(0, 80) }}{{ s.content.length > 80 ? '...' : '' }}</pre>
        </div>
        <el-empty v-if="!snippets.length" description="暂无片段" :image-size="40" />
      </div>
    </div>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreate" title="新建命令片段" width="400px" append-to-body>
      <el-form :model="form" label-width="60px">
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="命令"><el-input v-model="form.content" type="textarea" :rows="4" style="font-family: monospace;" /></el-form-item>
        <el-form-item label="标签"><el-input v-model="form.tags" placeholder="逗号分隔" /></el-form-item>
        <el-form-item label="可见性">
          <el-select v-model="form.visibility">
            <el-option label="仅自己" value="personal" />
            <el-option label="团队" value="team" />
            <el-option label="公开" value="public" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ArrowRight, ArrowLeft } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getSnippetList, searchSnippets, createSnippet } from '@/api/cmdb/snippet'

const emit = defineEmits(['insert'])

const expanded = ref(false)
const snippets = ref([])
const keyword = ref('')
const showCreate = ref(false)
const form = ref({ name: '', content: '', tags: '', visibility: 'personal' })

const fetchSnippets = async () => {
  try {
    const res = await getSnippetList({ page: 1, pageSize: 50 })
    snippets.value = res.data || []
  } catch (e) { /* ignore */ }
}

const handleSearch = async () => {
  if (!keyword.value) {
    await fetchSnippets()
    return
  }
  try {
    const res = await searchSnippets(keyword.value)
    snippets.value = res.data || []
  } catch (e) { /* ignore */ }
}

const insertSnippet = (snippet) => {
  emit('insert', snippet.content)
}

const handleCreate = async () => {
  if (!form.value.name || !form.value.content) {
    ElMessage.warning('名称和命令不能为空')
    return
  }
  try {
    await createSnippet(form.value)
    ElMessage.success('创建成功')
    showCreate.value = false
    form.value = { name: '', content: '', tags: '', visibility: 'personal' }
    await fetchSnippets()
  } catch (e) {
    ElMessage.error('创建失败')
  }
}

onMounted(fetchSnippets)
</script>

<style scoped>
.snippet-panel {
  display: flex;
  height: 100%;
  border-left: 1px solid #3c3c3c;
  background: #252526;
}

.snippet-panel.collapsed {
  width: 32px;
}

.panel-toggle {
  width: 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 12px;
  cursor: pointer;
  color: #888;
  font-size: 11px;
  writing-mode: vertical-rl;
}

.panel-toggle:hover { color: #fff; }

.panel-content {
  width: 220px;
  padding: 8px;
  overflow-y: auto;
}

.snippet-list { max-height: calc(100% - 80px); overflow-y: auto; }

.snippet-item {
  padding: 8px;
  border-radius: 4px;
  cursor: pointer;
  margin-bottom: 4px;
  background: #1e1e1e;
  transition: background 0.15s;
}
.snippet-item:hover { background: #2a2d2e; }

.snippet-name { font-size: 12px; font-weight: 600; color: #e0e0e0; margin-bottom: 2px; }
.snippet-tags { margin-bottom: 4px; }
.snippet-preview { margin: 0; font-size: 10px; color: #888; font-family: monospace; white-space: pre-wrap; }
</style>
