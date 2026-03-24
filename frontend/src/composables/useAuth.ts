import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

export function useAuth() {
  const authStore = useAuthStore()

  const user = computed(() => authStore.user)
  const isLoggedIn = computed(() => authStore.isLoggedIn)
  const permissions = computed(() => authStore.permissions)

  return {
    user,
    isLoggedIn,
    permissions,
    hasPermission: authStore.hasPermission,
    login: authStore.login,
    logout: authStore.logout,
  }
}
