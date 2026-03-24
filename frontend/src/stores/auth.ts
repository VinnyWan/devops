import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserInfo, AuthType } from '@/types/auth'
import * as authApi from '@/api/auth'
import { getUserPermissions } from '@/api/user-permission'
import { getItem, setItem, removeItem } from '@/utils/storage'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserInfo | null>(null)
  const token = ref<string | null>(getItem<string>('token'))
  const permissions = ref<string[]>([])
  const permissionsLoaded = ref(false)
  const isLoggedIn = computed(() => !!token.value)

  async function login(username: string, password: string, authType: AuthType = 'local') {
    const loginResult = await authApi.login({ username, password, authType })
    token.value = loginResult.token
    user.value = loginResult.user
    permissionsLoaded.value = false
    permissions.value = []
    setItem('token', loginResult.token)
  }

  async function fetchUser() {
    try {
      user.value = await authApi.getCurrentUser()
    } catch (error) {
      clearAuth()
      throw error
    }
  }

  async function fetchPermissions() {
    try {
      permissions.value = (await getUserPermissions()) ?? []
      permissionsLoaded.value = true
    } catch (error) {
      clearAuth()
      throw error
    }
  }

  function hasPermission(permission: string) {
    if (!permission) {
      return true
    }
    if (user.value?.isAdmin) {
      return true
    }
    return permissions.value.includes(permission)
  }

  async function logout() {
    await authApi.logout()
    clearAuth()
  }

  function clearAuth() {
    user.value = null
    token.value = null
    permissions.value = []
    permissionsLoaded.value = false
    removeItem('token')
  }

  return {
    user,
    token,
    permissions,
    permissionsLoaded,
    isLoggedIn,
    login,
    fetchUser,
    fetchPermissions,
    hasPermission,
    logout,
    clearAuth,
  }
})
