import request from './request'

// 登录
export const login = (data) =>
  request.post('/auth/login', data)

// 登出
export const logout = () =>
  request.post('/auth/logout')

// 获取当前用户全部权限（菜单/按钮/字段/API 四合一）
export const getPermissions = () =>
  request.get('/auth/permissions')
