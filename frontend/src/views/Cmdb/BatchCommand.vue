<template>
  <div class="page-container">
    <div class="page-header">
      <h3>批量命令</h3>
    </div>

    <el-row :gutter="16">
      <!-- Left: Host selection + command input -->
      <el-col :span="10">
        <el-card shadow="never">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>选择主机</span>
              <el-button type="primary" size="small" @click="showHostSelector = true">
                添加主机 (已选 {{ selectedHosts.length }})
              </el-button>
            </div>
          </template>

          <div class="selected-hosts" v-if="selectedHosts.length">
            <el-tag
              v-for="host in selectedHosts"
              :key="host.id"
              closable
              @close="removeHost(host)"
              style="margin: 2px;"
            >
              {{ host.hostname || host.ip }}
            </el-tag>
          </div>
          <el-empty v-else description="请选择主机" :image-size="40" />
        </el-card>

        <el-card shadow="never" style="margin-top: 16px;">
          <template #header><span>命令</span></template>
          <el-input
            v-model="command"
            type="textarea"
            :rows="6"
            placeholder="输入要执行的命令..."
            :disabled="executing"
            style="font-family: monospace;"
          />
          <div style="margin-top: 12px; display: flex; justify-content: space-between; align-items: center;">
            <el-input-number v-model="timeout" :min="5" :max="300" :step="5" size="small" style="width: 160px;" />
            <span style="font-size: 12px; color: #909399; margin-left: 8px;">超时(秒)</span>
            <el-button
              type="primary"
              @click="executeBatch"
              :loading="executing"
              :disabled="!command || !selectedHosts.length"
            >
              {{ executing ? '执行中...' : '执行' }}
            </el-button>
          </div>
        </el-card>
      </el-col>

      <!-- Right: Results -->
      <el-col :span="14">
        <el-card shadow="never">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>执行结果</span>
              <span v-if="results.length" style="font-size: 12px; color: #909399;">
                {{ successCount }}/{{ results.length }} 成功
              </span>
            </div>
          </template>
          <div class="results-panel">
            <div v-for="r in results" :key="r.hostId" class="result-item">
              <div class="result-header">
                <el-tag :type="r.status === 'success' ? 'success' : r.status === 'running' ? 'warning' : 'danger'" size="small">
                  {{ r.hostName || r.hostIp }}
                </el-tag>
                <span class="result-status">{{ statusText(r.status) }}</span>
              </div>
              <pre class="result-output">{{ r.output || r.error || '等待中...' }}</pre>
            </div>
            <el-empty v-if="!results.length && !executing" description="执行命令后查看结果" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Host Selector Dialog -->
    <el-dialog v-model="showHostSelector" title="选择主机" width="70%">
      <el-table
        ref="hostTableRef"
        :data="hosts"
        @selection-change="handleSelectionChange"
        stripe
        max-height="400"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="hostname" label="主机名" width="150" />
        <el-table-column prop="ip" label="IP" width="130" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'info'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="osName" label="系统" />
      </el-table>
      <template #footer>
        <el-button @click="showHostSelector = false">取消</el-button>
        <el-button type="primary" @click="confirmHostSelection">确认 ({{ tempSelection.length }})</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getHostList } from '@/api/cmdb/host'
import { getBatchCommandWsUrl } from '@/api/cmdb/batch_command'

const command = ref('')
const timeout = ref(30)
const executing = ref(false)
const selectedHosts = ref([])
const tempSelection = ref([])
const results = ref([])
const showHostSelector = ref(false)
const hosts = ref([])
const hostTableRef = ref(null)

const successCount = computed(() => results.value.filter(r => r.status === 'success').length)

const statusText = (s) => {
  const map = { running: '执行中...', success: '成功', failed: '失败', timeout: '超时' }
  return map[s] || s
}

const fetchHosts = async () => {
  try {
    const res = await getHostList({ page: 1, pageSize: 1000 })
    hosts.value = res.data?.list || res.data || []
  } catch (e) {
    ElMessage.error('获取主机列表失败')
  }
}

const handleSelectionChange = (selection) => {
  tempSelection.value = selection
}

const confirmHostSelection = () => {
  selectedHosts.value = [...tempSelection.value]
  showHostSelector.value = false
}

const syncTableSelection = async () => {
  if (!showHostSelector.value) return
  await nextTick()
  const table = hostTableRef.value
  if (!table) return

  table.clearSelection()
  const selectedIdSet = new Set(selectedHosts.value.map(host => host.id))
  hosts.value.forEach((host) => {
    if (selectedIdSet.has(host.id)) {
      table.toggleRowSelection(host, true)
    }
  })
  tempSelection.value = hosts.value.filter(host => selectedIdSet.has(host.id))
}

const removeHost = (host) => {
  selectedHosts.value = selectedHosts.value.filter(h => h.id !== host.id)
}

const executeBatch = () => {
  if (!command.value || !selectedHosts.value.length) return

  executing.value = true
  results.value = selectedHosts.value.map(h => ({
    hostId: h.id,
    hostName: h.hostname || h.ip,
    hostIp: h.ip,
    status: 'running',
    output: '',
    error: ''
  }))

  const ws = new WebSocket(getBatchCommandWsUrl())

  ws.onopen = () => {
    ws.send(JSON.stringify({
      hostIds: selectedHosts.value.map(h => h.id),
      command: command.value,
      timeout: timeout.value
    }))
  }

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data)

    if (msg.type === 'host_result') {
      const idx = results.value.findIndex(r => r.hostId === msg.data.hostId)
      if (idx >= 0) {
        results.value[idx] = { ...results.value[idx], ...msg.data }
      }
    } else if (msg.type === 'complete') {
      executing.value = false
      ElMessage.success(`执行完成: ${msg.success} 成功, ${msg.failed} 失败`)
      ws.close()
    } else if (msg.type === 'error') {
      executing.value = false
      ElMessage.error(msg.message)
      ws.close()
    }
  }

  ws.onerror = () => {
    executing.value = false
    ElMessage.error('WebSocket 连接失败')
  }

  ws.onclose = () => {
    executing.value = false
  }
}

watch(showHostSelector, (visible) => {
  if (visible) {
    syncTableSelection()
  }
})

onMounted(fetchHosts)
</script>

<style scoped>
.page-container { padding: 24px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h3 { margin: 0; font-size: 18px; font-weight: 500; }

.selected-hosts { max-height: 120px; overflow-y: auto; }

.results-panel { max-height: 65vh; overflow-y: auto; }
.result-item { margin-bottom: 12px; border: 1px solid #ebeef5; border-radius: 4px; overflow: hidden; }
.result-header { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: #f5f7fa; }
.result-status { font-size: 12px; color: #909399; }
.result-output { margin: 0; padding: 10px 12px; font-family: 'Consolas', 'Monaco', monospace; font-size: 12px; background: #1e1e1e; color: #d4d4d4; max-height: 200px; overflow-y: auto; white-space: pre-wrap; word-break: break-all; }
</style>
