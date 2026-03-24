import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const pendingRequestCount = ref(0)
  const isRequestLoading = ref(false)

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function startRequestLoading() {
    pendingRequestCount.value += 1
    isRequestLoading.value = pendingRequestCount.value > 0
  }

  function finishRequestLoading() {
    pendingRequestCount.value = Math.max(0, pendingRequestCount.value - 1)
    isRequestLoading.value = pendingRequestCount.value > 0
  }

  return {
    sidebarCollapsed,
    pendingRequestCount,
    isRequestLoading,
    toggleSidebar,
    startRequestLoading,
    finishRequestLoading,
  }
})
