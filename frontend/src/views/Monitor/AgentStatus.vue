<template>
  <div class="page-container">
    <div class="page-header">
      <h3>Agent 状态</h3>
    </div>

    <div class="toolbar">
      <el-select v-model="configId" placeholder="数据源" style="width: 200px">
        <el-option v-for="c in configs" :key="c.id" :label="c.name" :value="c.id" />
      </el-select>
      <el-input
        v-model="hostIps"
        placeholder="主机 IP（逗号分隔）"
        style="width: 300px"
        clearable
      />
      <el-button type="primary" @click="fetchStatus">查询</el-button>
    </div>

    <el-table :data="tableData" stripe v-loading="loading" style="margin-top: 16px">
      <el-table-column prop="ip" label="主机 IP" width="180" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag
            :type="
              row.status === 'online'
                ? 'success'
                : row.status === 'offline'
                  ? 'danger'
                  : 'info'
            "
          >
            {{
              row.status === 'online'
                ? '在线'
                : row.status === 'offline'
                  ? '离线'
                  : '未部署'
            }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listPrometheusConfigs, queryAgentStatus } from '@/api/monitor'

const configs = ref([])
const configId = ref(0)
const hostIps = ref('')
const tableData = ref([])
const loading = ref(false)

const fetchStatus = async () => {
  if (!configId.value || !hostIps.value) return
  loading.value = true
  try {
    const ips = hostIps.value
      .split(',')
      .map(s => s.trim())
      .filter(Boolean)
    const res = await queryAgentStatus({ configId: configId.value, hostIps: ips })
    const data = res.data || {}
    tableData.value = ips.map(ip => ({ ip, status: data[ip] || 'unknown' }))
  } catch {
    ElMessage.error('查询失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const res = await listPrometheusConfigs({ page: 1, pageSize: 100 })
    configs.value = res.data || []
    if (configs.value.length) {
      configId.value = configs.value[0].id
    }
  } catch {
    /* ignore */
  }
})
</script>

<style scoped>
.page-container {
  padding: 20px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
}
.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}
</style>
