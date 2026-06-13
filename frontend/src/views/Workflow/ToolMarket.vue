<template>
  <div class="page-container">
    <div class="page-header">
      <h3>运维工具市场</h3>
    </div>
    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <el-tab-pane label="可用工具" name="tools" />
      <el-tab-pane label="安装记录" name="installations" />
    </el-tabs>

    <!-- Tool list -->
    <div v-if="activeTab === 'tools'" v-loading="loading">
      <el-row :gutter="16">
        <el-col v-for="tool in tools" :key="tool.id" :xs="24" :sm="12" :md="8" :lg="6">
          <el-card class="tool-card" shadow="hover">
            <template #header>
              <div class="tool-header">
                <span class="tool-name">{{ tool.displayName }}</span>
                <el-tag size="small" type="info">{{ tool.category }}</el-tag>
              </div>
            </template>
            <p class="tool-desc">{{ tool.description }}</p>
            <div class="tool-actions">
              <el-button size="small" type="primary" @click="showInstallDialog(tool)">安装</el-button>
              <el-button size="small" @click="showCheckDialog(tool)">检查</el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- Installations -->
    <div v-if="activeTab === 'installations'">
      <el-table :data="installations" stripe v-loading="instLoading">
        <el-table-column label="工具" width="150">
          <template #default="{ row }">{{ toolLabel(row.toolId) }}</template>
        </el-table-column>
        <el-table-column prop="hostIp" label="主机" width="160" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="instStatusTag(row.status)" size="small">{{ instStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column label="安装时间" width="180">
          <template #default="{ row }">{{ formatTime(row.installedAt) }}</template>
        </el-table-column>
        <el-table-column prop="log" label="日志" min-width="200" show-overflow-tooltip />
      </el-table>
    </div>

    <!-- Install dialog -->
    <el-dialog v-model="installVisible" title="安装工具" width="480px" destroy-on-close>
      <el-form :model="installForm" label-width="100px">
        <el-form-item label="工具">
          <span>{{ installTarget?.displayName }}</span>
        </el-form-item>
        <el-form-item label="目标主机 IP" required>
          <el-input v-model="installForm.hostIp" placeholder="192.168.1.100" />
        </el-form-item>
        <el-form-item label="SSH 端口">
          <el-input-number v-model="installForm.sshPort" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="SSH 用户" required>
          <el-input v-model="installForm.sshUser" placeholder="root" />
        </el-form-item>
        <el-form-item label="SSH 密码">
          <el-input v-model="installForm.sshPassword" type="password" show-password placeholder="密码或密钥二选一" />
        </el-form-item>
        <el-form-item label="SSH 密钥">
          <el-input v-model="installForm.sshKey" type="textarea" :rows="3" placeholder="或粘贴私钥内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="installVisible = false">取消</el-button>
        <el-button type="primary" @click="handleInstall" :loading="installing">安装</el-button>
      </template>
    </el-dialog>

    <!-- Check dialog -->
    <el-dialog v-model="checkVisible" title="检查状态" width="480px" destroy-on-close>
      <el-form :model="checkForm" label-width="100px">
        <el-form-item label="工具">
          <span>{{ checkTarget?.displayName }}</span>
        </el-form-item>
        <el-form-item label="目标主机 IP" required>
          <el-input v-model="checkForm.hostIp" placeholder="192.168.1.100" />
        </el-form-item>
        <el-form-item label="SSH 端口">
          <el-input-number v-model="checkForm.sshPort" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="SSH 用户" required>
          <el-input v-model="checkForm.sshUser" placeholder="root" />
        </el-form-item>
        <el-form-item label="SSH 密码">
          <el-input v-model="checkForm.sshPassword" type="password" show-password />
        </el-form-item>
        <el-form-item label="SSH 密钥">
          <el-input v-model="checkForm.sshKey" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="checkVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCheck" :loading="checking">检查</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getToolList, installTool, checkToolStatus, getInstallations } from '@/api/tool'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const tools = ref([])
const activeTab = ref('tools')

const installVisible = ref(false)
const installing = ref(false)
const installTarget = ref(null)
const installForm = reactive({ hostIp: '', sshPort: 22, sshUser: 'root', sshPassword: '', sshKey: '' })

const checkVisible = ref(false)
const checking = ref(false)
const checkTarget = ref(null)
const checkForm = reactive({ hostIp: '', sshPort: 22, sshUser: 'root', sshPassword: '', sshKey: '' })

const installations = ref([])
const instLoading = ref(false)

const toolLabel = (id) => {
  const t = tools.value.find(t => t.id === id)
  return t ? t.displayName : `#${id}`
}

const instStatusLabel = (s) => {
  const map = { installed: '已安装', installing: '安装中', not_installed: '未安装', failed: '失败' }
  return map[s] || s
}

const instStatusTag = (s) => {
  const map = { installed: 'success', installing: 'warning', not_installed: 'info', failed: 'danger' }
  return map[s] || 'info'
}

const fetchTools = async () => {
  loading.value = true
  try {
    const res = await getToolList()
    tools.value = res.data || []
  } catch {
    ElMessage.error('获取工具列表失败')
  } finally {
    loading.value = false
  }
}

const fetchInstallations = async () => {
  instLoading.value = true
  try {
    const res = await getInstallations()
    installations.value = res.data || []
  } catch {
    installations.value = []
  } finally {
    instLoading.value = false
  }
}

const onTabChange = (tab) => {
  if (tab === 'installations') fetchInstallations()
}

const showInstallDialog = (tool) => {
  installTarget.value = tool
  installForm.hostIp = ''
  installForm.sshPort = 22
  installForm.sshUser = 'root'
  installForm.sshPassword = ''
  installForm.sshKey = ''
  installVisible.value = true
}

const handleInstall = async () => {
  if (!installForm.hostIp || !installForm.sshUser) return ElMessage.warning('请填写主机 IP 和 SSH 用户')
  installing.value = true
  try {
    await installTool(installTarget.value.id, { ...installForm })
    ElMessage.success('安装成功')
    installVisible.value = false
    fetchInstallations()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '安装失败')
  } finally {
    installing.value = false
  }
}

const showCheckDialog = (tool) => {
  checkTarget.value = tool
  checkForm.hostIp = ''
  checkForm.sshPort = 22
  checkForm.sshUser = 'root'
  checkForm.sshPassword = ''
  checkForm.sshKey = ''
  checkVisible.value = true
}

const handleCheck = async () => {
  if (!checkForm.hostIp || !checkForm.sshUser) return ElMessage.warning('请填写主机 IP 和 SSH 用户')
  checking.value = true
  try {
    await checkToolStatus(checkTarget.value.id, { ...checkForm })
    ElMessage.success('检查完成')
    checkVisible.value = false
    fetchInstallations()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '检查失败')
  } finally {
    checking.value = false
  }
}

onMounted(fetchTools)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.tool-card { margin-bottom: 16px; }
.tool-header { display: flex; justify-content: space-between; align-items: center; }
.tool-name { font-weight: 600; font-size: 15px; }
.tool-desc { color: #606266; font-size: 13px; line-height: 1.5; min-height: 36px; }
.tool-actions { display: flex; gap: 8px; margin-top: 8px; }
</style>
