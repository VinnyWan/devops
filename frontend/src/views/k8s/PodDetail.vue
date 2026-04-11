<template>
  <div class="page-container">
    <el-breadcrumb separator="/">
      <el-breadcrumb-item :to="{ path: '/k8s/workload' }">工作负载</el-breadcrumb-item>
      <el-breadcrumb-item>Pod 详情</el-breadcrumb-item>
    </el-breadcrumb>

    <el-tabs v-model="activeTab" style="margin-top: 16px">
      <el-tab-pane label="基本信息" name="info">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ podInfo.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ podInfo.namespace }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <StatusTag :status="podInfo.status" />
          </el-descriptions-item>
          <el-descriptions-item label="IP">{{ podInfo.ip }}</el-descriptions-item>
          <el-descriptions-item label="节点">{{ podInfo.node }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ podInfo.createdAt }}</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>

      <el-tab-pane label="容器列表" name="containers">
        <el-table :data="podInfo.containers" stripe>
          <el-table-column prop="name" label="容器名称" />
          <el-table-column prop="image" label="镜像" />
          <el-table-column prop="status" label="状态">
            <template #default="{ row }">
              <StatusTag :status="row.status" />
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="日志" name="logs">
        <LogViewer :logs="logs" />
      </el-tab-pane>

      <el-tab-pane label="YAML" name="yaml">
        <YamlEditor v-model="yaml" :readonly="true" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import StatusTag from '@/components/K8s/StatusTag.vue'
import LogViewer from '@/components/K8s/LogViewer.vue'
import YamlEditor from '@/components/K8s/YamlEditor.vue'
import { getPodDetail, getPodLogs } from '@/api/workload'

const route = useRoute()
const activeTab = ref('info')
const podInfo = ref({ containers: [] })
const logs = ref('')
const yaml = ref('')

const fetchPodDetail = async () => {
  const res = await getPodDetail({ podName: route.params.name })
  podInfo.value = res.data || { containers: [] }
  yaml.value = res.data?.yaml || ''
}

const fetchLogs = async () => {
  const res = await getPodLogs({ podName: route.params.name })
  logs.value = res.data || ''
}

onMounted(() => {
  fetchPodDetail()
  fetchLogs()
})
</script>

