<template>
  <div class="page-container">
    <div class="page-header">
      <h3>集群详情</h3>
      <el-button @click="$router.back()">返回</el-button>
    </div>

    <el-row :gutter="16">
      <el-col :span="8">
        <el-card>
          <template #header>网络统计</template>
          <div v-if="networkStats" class="stats-content">
            <div class="stat-item" v-for="(value, key) in networkStats" :key="key">
              <span class="label">{{ formatLabel(key) }}:</span>
              <span class="value">{{ value }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card>
          <template #header>存储统计</template>
          <div v-if="storageStats" class="stats-content">
            <div class="stat-item" v-for="(value, key) in storageStats" :key="key">
              <span class="label">{{ formatLabel(key) }}:</span>
              <span class="value">{{ value }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card>
          <template #header>工作负载统计</template>
          <div v-if="workloadStats" class="stats-content">
            <div class="stat-item" v-for="(value, key) in workloadStats" :key="key">
              <span class="label">{{ formatLabel(key) }}:</span>
              <span class="value">{{ value }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top: 16px">
      <template #header>集群事件</template>
      <el-table :data="events" stripe>
        <el-table-column prop="time" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.time) }}
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'Normal' ? 'success' : 'warning'" size="small">
              {{ row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" width="150" />
        <el-table-column prop="object" label="对象" width="300" />
        <el-table-column prop="message" label="消息" />
      </el-table>

      <el-pagination
        v-model:current-page="eventPage"
        v-model:page-size="eventPageSize"
        :total="eventTotal"
        @current-change="fetchEvents"
        style="margin-top: 16px; justify-content: flex-end"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getClusterNetworkStats, getClusterStorageStats, getClusterWorkloadStats, getClusterEvents } from '../../api/cluster'

const route = useRoute()
const clusterName = route.params.name

const networkStats = ref(null)
const storageStats = ref(null)
const workloadStats = ref(null)
const events = ref([])
const eventPage = ref(1)
const eventPageSize = ref(10)
const eventTotal = ref(0)

const fetchStats = async () => {
  const [network, storage, workload] = await Promise.all([
    getClusterNetworkStats(clusterName),
    getClusterStorageStats(clusterName),
    getClusterWorkloadStats(clusterName)
  ])
  networkStats.value = network.data
  storageStats.value = storage.data
  workloadStats.value = workload.data
}

const fetchEvents = async () => {
  const res = await getClusterEvents({ name: clusterName, page: eventPage.value, pageSize: eventPageSize.value })
  events.value = res.data.items || []
  eventTotal.value = res.data.total || 0
}

const formatLabel = (key) => {
  const map = {
    services: '服务数',
    ingresses: '入口数',
    pvs: 'PV数量',
    pvcs: 'PVC数量',
    storageclasses: '存储类',
    deployments: 'Deployment',
    statefulsets: 'StatefulSet',
    daemonsets: 'DaemonSet',
    pods: 'Pod数量'
  }
  return map[key] || key
}

const formatTime = (time) => {
  if (!time || time === '0001-01-01T00:00:00Z') return '-'
  return new Date(time).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchStats()
  fetchEvents()
})
</script>

<style scoped>
.page-container {
  background: #fff;
  border-radius: 4px;
  padding: 24px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}
.stats-content {
  padding: 8px 0;
}
.stat-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}
.stat-item:last-child {
  border-bottom: none;
}
.stat-item .label {
  color: #666;
}
.stat-item .value {
  font-weight: 500;
}
</style>
