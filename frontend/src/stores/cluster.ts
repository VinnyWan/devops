import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Cluster } from '@/types/cluster'
import { getClusterList } from '@/api/cluster'
import { getItem, setItem } from '@/utils/storage'

export const useClusterStore = defineStore('cluster', () => {
  const clusters = ref<Cluster[]>([])
  const currentClusterId = ref<number | null>(getItem<number>('currentClusterId'))

  async function fetchClusters() {
    clusters.value = await getClusterList()
    // 自动选中第一个集群
    if (!currentClusterId.value && clusters.value.length > 0) {
      const first = clusters.value[0]
      if (first) setCurrentCluster(first.id)
    }
  }

  function setCurrentCluster(id: number) {
    currentClusterId.value = id
    setItem('currentClusterId', id)
  }

  return { clusters, currentClusterId, fetchClusters, setCurrentCluster }
})
