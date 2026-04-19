<template>
  <div class="page-container">
    <div class="page-header">
      <h3>文件管理</h3>
      <div style="display: flex; gap: 8px; align-items: center;">
        <el-select v-model="hostId" placeholder="选择主机" clearable filterable style="width: 240px;" @change="handleHostChange">
          <el-option v-for="h in hosts" :key="h.id" :label="`${h.hostname || h.ip} (${h.ip})`" :value="h.id" />
        </el-select>
        <el-button type="primary" @click="showUploadDialog" :disabled="!hostId">上传文件</el-button>
        <el-button @click="showMkdirDialog" :disabled="!hostId">新建目录</el-button>
        <el-button @click="showDistributeDialog" :disabled="!hostId">批量分发</el-button>
      </div>
    </div>

    <!-- Empty state -->
    <div v-if="!hostId" class="empty-state">
      <el-icon :size="64" color="#c0c4cc"><FolderOpened /></el-icon>
      <p>请先选择一台主机</p>
    </div>

    <!-- File browser -->
    <template v-else>
      <div class="path-breadcrumb">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item
            v-for="(segment, idx) in pathSegments"
            :key="idx"
            @click="navigateToBreadcrumb(idx)"
          >
            <span :class="{ 'path-link': idx < pathSegments.length - 1 }">
              {{ segment }}
            </span>
          </el-breadcrumb-item>
        </el-breadcrumb>
      </div>

      <el-table v-if="loading || files.length > 0" :data="files" stripe v-loading="loading" style="width: 100%" @row-dblclick="handleDblClick">
        <el-table-column label="名称" min-width="280">
          <template #default="{ row }">
            <span class="file-name">
              <el-icon :size="18" :color="row.isDir ? '#e6a23c' : '#909399'">
                <Folder v-if="row.isDir" />
                <Document v-else />
              </el-icon>
              <span>{{ row.name }}</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column label="大小" width="120">
          <template #default="{ row }">{{ row.isDir ? '-' : formatSize(row.size) }}</template>
        </el-table-column>
        <el-table-column prop="mode" label="权限" width="120" />
        <el-table-column label="修改时间" width="180">
          <template #default="{ row }">{{ row.modTime }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="handleDownload(row)" :disabled="row.isDir">下载</el-button>
            <el-button size="small" @click="showRenameDialog(row)">重命名</el-button>
            <el-button size="small" @click="showEditDialog(row)" :disabled="row.isDir || !isTextFile(row.name)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="当前目录暂无文件" :image-size="60" />
    </template>

    <!-- Upload dialog -->
    <el-dialog v-model="uploadDialogVisible" title="上传文件" width="520px">
      <el-form label-width="90px">
        <el-form-item label="远程路径">
          <el-input v-model="uploadPath" placeholder="目标目录路径，如 /tmp" />
        </el-form-item>
        <el-form-item label="选择文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            :on-exceed="() => ElMessage.warning('只能选择一个文件')"
          >
            <el-button>选择文件</el-button>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUpload" :loading="uploading">上传</el-button>
      </template>
    </el-dialog>

    <!-- Mkdir dialog -->
    <el-dialog v-model="mkdirDialogVisible" title="新建目录" width="480px">
      <el-form label-width="90px">
        <el-form-item label="目录路径">
          <el-input v-model="mkdirPath" placeholder="如 /tmp/newdir" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="mkdirDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleMkdir">创建</el-button>
      </template>
    </el-dialog>

    <!-- Rename dialog -->
    <el-dialog v-model="renameDialogVisible" title="重命名" width="480px">
      <el-form label-width="90px">
        <el-form-item label="新名称">
          <el-input v-model="renameNewName" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="renameDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRename">确定</el-button>
      </template>
    </el-dialog>

    <!-- Edit dialog -->
    <el-dialog v-model="editDialogVisible" title="编辑文件" width="80%" top="5vh">
      <el-input
        v-model="editContent"
        type="textarea"
        :rows="24"
        style="font-family: 'Courier New', Courier, monospace;"
      />
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditSave" :loading="editSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- Distribute dialog -->
    <el-dialog v-model="distributeDialogVisible" title="批量分发" width="640px">
      <el-form label-width="90px">
        <el-form-item label="选择文件">
          <el-upload
            ref="distributeUploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleDistributeFileChange"
            :on-exceed="() => ElMessage.warning('只能选择一个文件')"
          >
            <el-button>选择文件</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item label="目标路径">
          <el-input v-model="distributePath" placeholder="如 /tmp/deploy/" />
        </el-form-item>
        <el-form-item label="目标主机">
          <el-select v-model="distributeHostIds" multiple filterable placeholder="选择目标主机" style="width: 100%;">
            <el-option v-for="h in hosts" :key="h.id" :label="`${h.hostname || h.ip} (${h.ip})`" :value="h.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="distributeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleDistribute" :loading="distributing">分发</el-button>
      </template>
    </el-dialog>

    <!-- Distribute results dialog -->
    <el-dialog v-model="distributeResultVisible" title="分发结果" width="640px">
      <el-table :data="distributeResults" stripe style="width: 100%">
        <el-table-column prop="host" label="主机" min-width="160" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.success ? 'success' : 'danger'">{{ row.success ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="信息" min-width="200" />
      </el-table>
      <template #footer>
        <el-button @click="distributeResultVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Folder, Document, FolderOpened } from '@element-plus/icons-vue'
import { getHostList } from '@/api/cmdb/host'
import {
  browseFiles, uploadFile, deleteFile, renameFile,
  mkdir, previewFile, editFile, distributeFile, getDownloadUrl
} from '@/api/cmdb/file'

const hosts = ref([])
const hostId = ref('')
const currentPath = ref('/')
const files = ref([])
const loading = ref(false)

// Upload
const uploadDialogVisible = ref(false)
const uploadPath = ref('')
const uploadFileList = ref([])
const uploadRef = ref()
const uploading = ref(false)

// Mkdir
const mkdirDialogVisible = ref(false)
const mkdirPath = ref('')

// Rename
const renameDialogVisible = ref(false)
const renameRow = ref(null)
const renameNewName = ref('')

// Edit
const editDialogVisible = ref(false)
const editContent = ref('')
const editFilePath = ref('')
const editSaving = ref(false)

// Distribute
const distributeDialogVisible = ref(false)
const distributePath = ref('')
const distributeHostIds = ref([])
const distributeFile_ = ref(null)
const distributeUploadRef = ref()
const distributing = ref(false)
const distributeResultVisible = ref(false)
const distributeResults = ref([])

const pathSegments = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  return ['/', ...parts]
})

function isTextFile(name) {
  const exts = ['.txt', '.conf', '.cfg', '.ini', '.yaml', '.yml', '.json', '.xml', '.sh', '.py', '.js',
    '.ts', '.go', '.java', '.c', '.cpp', '.h', '.md', '.log', '.toml', '.env', '.sql', '.css', '.html', '.vue']
  return exts.some(ext => name.toLowerCase().endsWith(ext))
}

function formatSize(bytes) {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}

const fetchHosts = async () => {
  try {
    const res = await getHostList({ page: 1, pageSize: 500 })
    hosts.value = res.data || []
  } catch (e) {
    ElMessage.error(e.message || '获取主机列表失败')
  }
}

const fetchFiles = async () => {
  if (!hostId.value) return
  loading.value = true
  try {
    const res = await browseFiles({ hostId: hostId.value, path: currentPath.value })
    files.value = res.data || []
  } catch (e) {
    ElMessage.error(e.message || '获取文件列表失败')
  } finally {
    loading.value = false
  }
}

const handleHostChange = () => {
  currentPath.value = '/'
  fetchFiles()
}

const handleDblClick = (row) => {
  if (!row.isDir) return
  if (currentPath.value.endsWith('/')) {
    currentPath.value = currentPath.value + row.name
  } else {
    currentPath.value = currentPath.value + '/' + row.name
  }
  fetchFiles()
}

const navigateToBreadcrumb = (idx) => {
  if (idx === 0) {
    currentPath.value = '/'
  } else {
    const parts = currentPath.value.split('/').filter(Boolean)
    currentPath.value = '/' + parts.slice(0, idx).join('/')
  }
  fetchFiles()
}

// Upload
const showUploadDialog = () => {
  uploadPath.value = currentPath.value
  uploadFileList.value = []
  uploadDialogVisible.value = true
}

const handleFileChange = (file) => {
  uploadFileList.value = [file]
}

const handleUpload = async () => {
  if (!uploadFileList.value.length) {
    ElMessage.warning('请选择文件')
    return
  }
  uploading.value = true
  try {
    await uploadFile(hostId.value, uploadPath.value, uploadFileList.value[0].raw)
    ElMessage.success('上传成功')
    uploadDialogVisible.value = false
    fetchFiles()
  } catch (e) {
    ElMessage.error(e.message || '上传失败')
  } finally {
    uploading.value = false
  }
}

// Mkdir
const showMkdirDialog = () => {
  mkdirPath.value = currentPath.value.endsWith('/')
    ? currentPath.value
    : currentPath.value + '/'
  mkdirDialogVisible.value = true
}

const handleMkdir = async () => {
  if (!mkdirPath.value) {
    ElMessage.warning('请输入目录路径')
    return
  }
  try {
    await mkdir({ hostId: hostId.value, path: mkdirPath.value })
    ElMessage.success('创建成功')
    mkdirDialogVisible.value = false
    fetchFiles()
  } catch (e) {
    ElMessage.error(e.message || '创建失败')
  }
}

// Rename
const showRenameDialog = (row) => {
  renameRow.value = row
  renameNewName.value = row.name
  renameDialogVisible.value = true
}

const handleRename = async () => {
  if (!renameNewName.value) {
    ElMessage.warning('请输入新名称')
    return
  }
  const dirPath = currentPath.value.endsWith('/')
    ? currentPath.value
    : currentPath.value + '/'
  const oldPath = dirPath + renameRow.value.name
  const newPath = dirPath + renameNewName.value
  try {
    await renameFile({ hostId: hostId.value, oldPath, newPath })
    ElMessage.success('重命名成功')
    renameDialogVisible.value = false
    fetchFiles()
  } catch (e) {
    ElMessage.error(e.message || '重命名失败')
  }
}

// Edit
const showEditDialog = async (row) => {
  const dirPath = currentPath.value.endsWith('/')
    ? currentPath.value
    : currentPath.value + '/'
  const filePath = dirPath + row.name
  editFilePath.value = filePath
  try {
    const res = await previewFile({ hostId: hostId.value, path: filePath })
    editContent.value = res.data || ''
    editDialogVisible.value = true
  } catch (e) {
    ElMessage.error(e.message || '预览文件失败')
  }
}

const handleEditSave = async () => {
  editSaving.value = true
  try {
    await editFile({ hostId: hostId.value, path: editFilePath.value, content: editContent.value })
    ElMessage.success('保存成功')
    editDialogVisible.value = false
    fetchFiles()
  } catch (e) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    editSaving.value = false
  }
}

// Download
const handleDownload = (row) => {
  const dirPath = currentPath.value.endsWith('/')
    ? currentPath.value
    : currentPath.value + '/'
  const filePath = dirPath + row.name
  window.open(getDownloadUrl(hostId.value, filePath), '_blank')
}

// Delete
const handleDelete = async (row) => {
  const dirPath = currentPath.value.endsWith('/')
    ? currentPath.value
    : currentPath.value + '/'
  const filePath = dirPath + row.name
  await ElMessageBox.confirm(`确认删除 "${row.name}"？`, '提示', { type: 'warning' })
  try {
    await deleteFile({ hostId: hostId.value, path: filePath })
    ElMessage.success('删除成功')
    fetchFiles()
  } catch (e) {
    ElMessage.error(e.message || '删除失败')
  }
}

// Distribute
const showDistributeDialog = () => {
  distributePath.value = ''
  distributeHostIds.value = []
  distributeFile_.value = null
  distributeDialogVisible.value = true
}

const handleDistributeFileChange = (file) => {
  distributeFile_.value = file
}

const handleDistribute = async () => {
  if (!distributeFile_.value) {
    ElMessage.warning('请选择文件')
    return
  }
  if (!distributePath.value) {
    ElMessage.warning('请输入目标路径')
    return
  }
  if (!distributeHostIds.value.length) {
    ElMessage.warning('请选择目标主机')
    return
  }
  distributing.value = true
  try {
    const res = await distributeFile(distributeFile_.value.raw, distributePath.value, distributeHostIds.value)
    distributeResults.value = (res.data || []).map(r => ({
      host: r.host || r.hostname || r.ip || r.hostId,
      success: r.success || r.error === '' || !r.error,
      message: r.message || r.error || (r.success ? '成功' : '失败')
    }))
    distributeDialogVisible.value = false
    distributeResultVisible.value = true
  } catch (e) {
    ElMessage.error(e.message || '分发失败')
  } finally {
    distributing.value = false
  }
}

onMounted(() => {
  fetchHosts()
})
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.toolbar { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 120px 0;
  color: #909399;
}
.empty-state p {
  margin-top: 16px;
  font-size: 16px;
}

.path-breadcrumb {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 10px 16px;
  margin-bottom: 16px;
}

.path-link {
  cursor: pointer;
  color: #409eff;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 6px;
}
</style>
