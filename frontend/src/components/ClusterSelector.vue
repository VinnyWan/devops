<script setup lang="ts">
import { onMounted } from 'vue'
import { NSelect } from 'naive-ui'
import { useCluster } from '@/composables/useCluster'

const { clusters, currentClusterId, fetchClusters, setCurrentCluster } = useCluster()

onMounted(() => {
  if (clusters.value.length === 0) {
    fetchClusters()
  }
})

function handleChange(value: number) {
  setCurrentCluster(value)
}
</script>

<template>
  <n-select
    :value="currentClusterId"
    :options="clusters.map((c) => ({ label: c.name, value: c.id }))"
    placeholder="选择集群"
    style="width: 200px"
    @update:value="handleChange"
  />
</template>
