<template>
  <div class="page-container">
    <div class="page-header">
      <h3>SQL 操作审计</h3>
    </div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="数据库连接" name="connections" />
      <el-tab-pane label="SQL 执行" name="execute" />
      <el-tab-pane label="审计记录" name="records" />
    </el-tabs>

    <!-- Connections -->
    <div v-if="activeTab === 'connections'" v-loading="connLoading">
      <div style="margin-bottom:16px">
        <el-button type="primary" @click="showConnDialog">添加连接</el-button>
      </div>
      <el-table :data="connections" stripe>
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column label="地址" min-width="200">
          <template #default="{ row }">{{ row.host }}:{{ row.port }}/{{ row.database }}</template>
        </el-table-column>
        <el-table-column label="模式" width="100">
          <template #default="{ row }">
            <el-tag :type="row.mode === 'read_only' ? 'warning' : ''" size="small">{{ row.mode === 'read_only' ? '只读' : '读写' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleTest(row)">测试</el-button>
            <el-button link type="danger" size="small" @click="handleDeleteConn(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- SQL Execute -->
    <div v-if="activeTab === 'execute'">
      <el-form :model="sqlForm" label-width="100px">
        <el-form-item label="数据库连接">
          <el-select v-model="sqlForm.connectionId" placeholder="选择数据库连接" style="width: 100%">
            <el-option v-for="c in connections" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="SQL 语句">
          <el-input v-model="sqlForm.sql" type="textarea" :rows="6" placeholder="输入 SQL 语句..." />
        </el-form-item>
      </el-form>
      <el-button type="primary" @click="handleExecute" :loading="executing">执行</el-button>
      <el-alert v-if="sqlError" :title="sqlError" type="error" show-icon style="margin-top:12px" closable @close="sqlError=''" />

      <!-- Query result -->
      <div v-if="sqlResult" style="margin-top:16px">
        <el-tag size="small" style="margin-bottom:8px">耗时 {{ sqlResult.duration }}ms · {{ sqlResult.rowsAffected }} 行</el-tag>
        <el-table :data="sqlResult.rows" stripe border max-height="400" v-if="sqlResult.columns">
          <el-table-column v-for="(col, i) in sqlResult.columns" :key="col" :prop="String(i)" :label="col" min-width="120" show-overflow-tooltip />
        </el-table>
        <el-alert v-else :title="'影响行数: ' + sqlResult.rowsAffected" type="success" />
      </div>
    </div>

    <!-- Audit records -->
    <div v-if="activeTab === 'records'" v-loading="recordLoading">
      <el-table :data="records" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="SQL" min-width="250" show-overflow-tooltip>
          <template #default="{ row }">
            <span :style="{ color: row.sensitive ? '#f56c6c' : '' }">{{ row.sql }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="database" label="数据库" width="120" />
        <el-table-column label="风险" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.riskLevel !== 'none'" :type="riskTag(row.riskLevel)" size="small">{{ row.riskLevel }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="耗时" width="80">
          <template #default="{ row }">{{ row.duration }}ms</template>
        </el-table-column>
        <el-table-column label="时间" width="180">
          <template #default="{ row }">{{ formatTime(row.executedAt) }}</template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="recordPage"
          v-model:page-size="recordPageSize"
          :total="recordTotal"
          layout="total, prev, pager, next"
          @current-change="fetchRecords"
        />
      </div>
    </div>

    <!-- Add connection dialog -->
    <el-dialog v-model="connVisible" title="添加数据库连接" width="480px" destroy-on-close>
      <el-form :model="connForm" label-width="100px">
        <el-form-item label="名称" required><el-input v-model="connForm.name" /></el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="connForm.type" style="width:100%">
            <el-option label="MySQL" value="mysql" />
            <el-option label="PostgreSQL" value="postgresql" />
          </el-select>
        </el-form-item>
        <el-form-item label="主机" required><el-input v-model="connForm.host" /></el-form-item>
        <el-form-item label="端口"><el-input-number v-model="connForm.port" :min="1" :max="65535" /></el-form-item>
        <el-form-item label="数据库"><el-input v-model="connForm.database" /></el-form-item>
        <el-form-item label="用户名" required><el-input v-model="connForm.username" /></el-form-item>
        <el-form-item label="密码" required><el-input v-model="connForm.password" type="password" show-password /></el-form-item>
        <el-form-item label="模式">
          <el-radio-group v-model="connForm.mode">
            <el-radio value="read_write">读写</el-radio>
            <el-radio value="read_only">只读</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="connVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateConn" :loading="connSubmitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getConnections, createConnection, testConnection, deleteConnection, executeSQL, getSqlRecords } from '@/api/sqlaudit'
import { formatTime } from '@/utils/format'

const activeTab = ref('connections')

// Connections
const connLoading = ref(false)
const connections = ref([])
const connVisible = ref(false)
const connSubmitting = ref(false)
const connForm = reactive({ name: '', type: 'mysql', host: '', port: 3306, database: '', username: '', password: '', mode: 'read_write' })

// SQL execution
const sqlForm = reactive({ connectionId: 0, sql: '' })
const executing = ref(false)
const sqlResult = ref(null)
const sqlError = ref('')

// Records
const recordLoading = ref(false)
const records = ref([])
const recordTotal = ref(0)
const recordPage = ref(1)
const recordPageSize = ref(20)

const riskTag = (level) => {
  const map = { low: 'info', medium: 'warning', high: 'danger' }
  return map[level] || ''
}

const fetchConnections = async () => {
  connLoading.value = true
  try {
    const res = await getConnections()
    connections.value = res.data || []
  } catch {
    connections.value = []
  } finally {
    connLoading.value = false
  }
}

const showConnDialog = () => {
  connForm.name = ''; connForm.type = 'mysql'; connForm.host = ''; connForm.port = 3306
  connForm.database = ''; connForm.username = ''; connForm.password = ''; connForm.mode = 'read_write'
  connVisible.value = true
}

const handleCreateConn = async () => {
  if (!connForm.name || !connForm.host || !connForm.username || !connForm.password) {
    return ElMessage.warning('请填写必填项')
  }
  connSubmitting.value = true
  try {
    await createConnection({ ...connForm })
    ElMessage.success('添加成功')
    connVisible.value = false
    fetchConnections()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '添加失败')
  } finally {
    connSubmitting.value = false
  }
}

const handleTest = async (row) => {
  try {
    await testConnection(row.id)
    ElMessage.success('连接正常')
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '连接失败')
  }
}

const handleDeleteConn = async (row) => {
  try { await ElMessageBox.confirm('确定删除该连接？', '确认', { type: 'warning' }) } catch { return }
  try {
    await deleteConnection(row.id)
    ElMessage.success('已删除')
    fetchConnections()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '删除失败')
  }
}

const handleExecute = async () => {
  if (!sqlForm.connectionId || !sqlForm.sql) return ElMessage.warning('请选择连接并输入 SQL')
  executing.value = true
  sqlError.value = ''
  sqlResult.value = null
  try {
    const res = await executeSQL({ ...sqlForm })
    sqlResult.value = res.data
  } catch (e) {
    sqlError.value = e.response?.data?.message || '执行失败'
  } finally {
    executing.value = false
  }
}

const fetchRecords = async () => {
  recordLoading.value = true
  try {
    const res = await getSqlRecords({ page: recordPage.value, pageSize: recordPageSize.value })
    const data = res.data || {}
    records.value = data.items || []
    recordTotal.value = data.total || 0
  } catch {
    records.value = []
  } finally {
    recordLoading.value = false
  }
}

onMounted(fetchConnections)
</script>

<style scoped>
.page-container { background: #fff; border-radius: 4px; padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
