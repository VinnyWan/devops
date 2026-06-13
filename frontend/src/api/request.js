import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000
})

let isRefreshing = false
let pendingRequests = []

function handleRefreshQueue(error) {
  pendingRequests.forEach(cb => cb(error || null))
  pendingRequests = []
}

request.interceptors.request.use(
  config => {
    // Priority: JWT token > session_id (both stored in sessionStorage)
    const jwtToken = sessionStorage.getItem('jwt_token')
    if (jwtToken) {
      config.headers.Authorization = `Bearer ${jwtToken}`
    }
    return config
  },
  error => Promise.reject(error)
)

request.interceptors.response.use(
  response => response.data,
  async error => {
    if (error.response) {
      const { status, data, config } = error.response

      // Handle 401 with automatic JWT refresh
      if (status === 401 && !config._retry) {
        const refreshToken = sessionStorage.getItem('refresh_token')
        if (refreshToken) {
          if (!isRefreshing) {
            isRefreshing = true
            config._retry = true
            try {
              const res = await axios.post(
                `${import.meta.env.VITE_API_BASE_URL}/auth/refresh`,
                { refresh_token: refreshToken }
              )
              const newToken = res.data?.data?.token
              if (newToken) {
                sessionStorage.setItem('jwt_token', newToken)
                config.headers.Authorization = `Bearer ${newToken}`
                handleRefreshQueue()
                return request(config)
              }
            } catch {
              handleRefreshQueue(new Error('Token refresh failed'))
            } finally {
              isRefreshing = false
            }
          } else {
            return new Promise((resolve, reject) => {
              pendingRequests.push(err => {
                if (err) {
                  reject(err)
                } else {
                  config.headers.Authorization = `Bearer ${sessionStorage.getItem('jwt_token')}`
                  resolve(request(config))
                }
              })
            })
          }
        }

        // Clear auth state and redirect to login
        sessionStorage.removeItem('jwt_token')
        sessionStorage.removeItem('refresh_token')
        sessionStorage.removeItem('token')
        sessionStorage.removeItem('userInfo')
        ElMessage.error('登录已过期，请重新登录')
        window.location.href = '/login'
        return Promise.reject(error)
      }

      // Other errors
      switch (status) {
        case 403:
          ElMessage.error('权限不足，无法访问')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        default:
          ElMessage.error(data?.error || data?.message || `请求失败 (${status})`)
      }
    } else if (error.code === 'ERR_NETWORK') {
      ElMessage.error('网络连接失败，请检查网络')
    } else {
      ElMessage.error(error.message || '请求失败')
    }
    return Promise.reject(error)
  }
)

export default request
