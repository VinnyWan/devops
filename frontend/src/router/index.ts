import { createRouter, createWebHistory } from 'vue-router'
import routes from './routes'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  // 不需要认证的页面直接放行
  if (to.meta.requiresAuth === false) {
    return true
  }

  // 未登录跳转到登录页
  if (!authStore.isLoggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  // 已登录但没有用户信息，尝试获取
  if (!authStore.user) {
    try {
      await authStore.fetchUser()
    } catch {
      authStore.clearAuth()
      return { path: '/login', query: { redirect: to.fullPath } }
    }
  }

  // 获取权限（失败不影响导航）
  if (!authStore.permissionsLoaded) {
    try {
      await authStore.fetchPermissions()
    } catch (error) {
      console.error('Failed to fetch permissions:', error)
      authStore.permissionsLoaded = true
    }
  }

  const routePermissions = to.meta.permissions
  const requiredPermissions = Array.isArray(routePermissions) ? routePermissions : []
  if (requiredPermissions.length > 0) {
    const hasAccess = requiredPermissions.every((permission) => authStore.hasPermission(permission))
    if (!hasAccess) {
      return { path: '/dashboard' }
    }
  }

  return true
})

// 设置页面标题
router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} - DevOps` : 'DevOps 管理平台'
})

export default router
