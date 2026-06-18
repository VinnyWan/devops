<template>
  <div class="page-container">
    <div class="page-header">
      <h3>主机监控</h3>
    </div>

    <div class="toolbar">
      <el-select v-model="configId" placeholder="数据源" style="width: 200px" @change="fetchMetrics">
        <el-option v-for="c in configs" :key="c.id" :label="c.name" :value="c.id" />
      </el-select>
      <el-input
        v-model="hostIp"
        placeholder="主机 IP"
        style="width: 180px"
        clearable
        @change="fetchMetrics"
      />
      <el-select v-model="metricType" style="width: 100px" @change="fetchMetrics">
        <el-option label="CPU" value="cpu" />
        <el-option label="内存" value="memory" />
        <el-option label="磁盘" value="disk" />
      </el-select>
      <el-select v-model="timeRange" style="width: 140px" @change="fetchMetrics">
        <el-option label="最近30分钟" value="30m" />
        <el-option label="最近1小时" value="1h" />
        <el-option label="最近6小时" value="6h" />
        <el-option label="最近24小时" value="24h" />
      </el-select>
      <el-button type="primary" @click="fetchMetrics">查询</el-button>
    </div>

    <div ref="chartRef" style="height: 400px; margin-top: 20px" v-show="hasData" />
    <el-empty v-if="!hasData" description="请选择数据源并输入主机 IP 查询" />
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { listPrometheusConfigs, queryHostMetrics } from '@/api/monitor'
import * as echarts from 'echarts'

const configs = ref([])
const configId = ref(0)
const hostIp = ref('')
const metricType = ref('cpu')
const timeRange = ref('1h')
const hasData = ref(false)
const chartRef = ref(null)
let chart = null

const fetchMetrics = async () => {
  if (!configId.value || !hostIp.value) return
  try {
    const rangeMap = { '30m': 1800, '1h': 3600, '6h': 21600, '24h': 86400 }
    const now = Math.floor(Date.now() / 1000)
    const start = now - (rangeMap[timeRange.value] || 3600)
    const res = await queryHostMetrics({
      configId: configId.value,
      hostIp: hostIp.value,
      metric: metricType.value,
      startTime: new Date(start * 1000).toISOString(),
      endTime: new Date(now * 1000).toISOString()
    })
    const results = res.data?.results || []
    hasData.value = results.length > 0
    await nextTick()
    if (!chart) {
      chart = echarts.init(chartRef.value)
    }
    const unitNames = { cpu: '%', memory: '%', disk: 'bytes' }
    chart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: results.map(s => Object.values(s.metric || {}).join(' ')) },
      xAxis: { type: 'time' },
      yAxis: { type: 'value', name: unitNames[metricType.value] || '' },
      series: results.map(s => ({
        name: Object.values(s.metric || {}).join(' '),
        type: 'line',
        smooth: true,
        data: (s.values || []).map(v => [v.timestamp * 1000, v.value])
      }))
    })
  } catch {
    ElMessage.error('获取指标失败')
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

onBeforeUnmount(() => {
  chart?.dispose()
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
