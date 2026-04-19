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

  loadUserInfo()

  return { token, userInfo, setToken, setUserInfo, loadUserInfo, logout, hasPermission }
})
