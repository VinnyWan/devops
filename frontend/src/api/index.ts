import axios from 'axios'
import type { ApiResponse } from '@/types/api'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { getItem } from '@/utils/storage'
import router from '@/router'

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
  withCredentials: true,
})

// 请求拦截器：附加 Token
http.interceptors.request.use((config) => {
  const appStore = useAppStore()
  const skipLoading = Boolean((config as typeof config & { skipLoading?: boolean }).skipLoading)
  if (!skipLoading) {
    appStore.startRequestLoading()
    ;(config as typeof config & { __trackedLoading?: boolean }).__trackedLoading = true
  }
  const token = getItem<string>('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器
http.interceptors.response.use(
  (response) => {
    const appStore = useAppStore()
    if (
      (response.config as typeof response.config & { __trackedLoading?: boolean }).__trackedLoading
    ) {
      appStore.finishRequestLoading()
    }
    const res = response.data as ApiResponse
    if (res.code !== 200) {
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return response
  },
  (error) => {
    const appStore = useAppStore()
    if ((error.config as typeof error.config & { __trackedLoading?: boolean }).__trackedLoading) {
      appStore.finishRequestLoading()
    }
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.clearAuth()
      router.push('/login')
    }
    return Promise.reject(error)
  },
)

export default http
