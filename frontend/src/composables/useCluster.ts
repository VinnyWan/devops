import { computed } from 'vue'
import { useClusterStore } from '@/stores/cluster'

export function useCluster() {
  const store = useClusterStore()

  const clusters = computed(() => store.clusters)
  const currentClusterId = computed(() => store.currentClusterId)
  const currentCluster = computed(() => store.clusters.find((c) => c.id === store.currentClusterId))

  return {
    clusters,
    currentClusterId,
    currentCluster,
    fetchClusters: store.fetchClusters,
    setCurrentCluster: store.setCurrentCluster,
  }
}
