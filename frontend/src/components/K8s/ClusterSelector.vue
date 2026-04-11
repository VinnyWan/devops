<template>
  <el-select v-model="selectedCluster" placeholder="选择集群" @change="handleChange" style="width: 200px">
    <el-option v-for="cluster in clusters" :key="cluster.name" :label="cluster.name" :value="cluster.name">
      <span>{{ cluster.name }}</span>
      <el-tag :type="cluster.status === 'healthy' ? 'success' : 'danger'" size="small" style="margin-left: 8px">
        {{ cluster.status }}
      </el-tag>
    </el-option>
  </el-select>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getClusterList } from '@/api/cluster'

const selectedCluster = defineModel()
const emit = defineEmits(['change'])
const clusters = ref([])

const fetchClusters = async () => {
  const res = await getClusterList()
  clusters.value = res.data || []
  if (clusters.value.length > 0 && !selectedCluster.value) {
    selectedCluster.value = clusters.value[0].name
    emit('change', clusters.value[0].name)
  }
}

const handleChange = (val) => {
  selectedCluster.value = val
  emit('change', val)
}

onMounted(fetchClusters)
</script>
