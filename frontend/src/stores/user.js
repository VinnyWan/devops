import { defineStore } from 'pinia'
import { ref } from 'vue'
import request from '../api/request'

export const useUserStore = defineStore('user', () => {
  const token = ref(sessionStorage.getItem('token') || '')
  const userInfo = ref(null)

  const setToken = (newToken) => {
    token.value = newToken
    sessionStorage.setItem('token', newToken)
  }

  const setUserInfo = (info) => {
    userInfo.value = info
    sessionStorage.setItem('userInfo', JSON.stringify(info))
  }

  const loadUserInfo = () => {
    const stored = sessionStorage.getItem('userInfo')
    if (stored) {
      userInfo.value = JSON.parse(stored)
    }
  }

  const logout = () => {
    token.value = ''
    userInfo.value = null
    sessionStorage.removeItem('token')
    sessionStorage.removeItem('userInfo')
  }

  const hasPermission = (permission) => {
    return userInfo.value?.permissions?.includes(permission) || false
  }

  const fetchPermissions = async () => {
    if (!token.value) return
    try {
      const res = await request.get('/user/permissions')
      if (res.data) {
        userInfo.value = { ...userInfo.value, permissions: res.data }
        sessionStorage.setItem('userInfo', JSON.stringify(userInfo.value))
      }
    } catch (e) {
      // 权限加载失败不阻塞页面
    }
  }

  loadUserInfo()

  return { token, userInfo, setToken, setUserInfo, loadUserInfo, fetchPermissions, logout, hasPermission }
})
