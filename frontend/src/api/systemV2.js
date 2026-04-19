import request from './request'

// ==================== 用户管理（新路由 /system/users） ====================

// 用户列表
export const getUserListV2 = (params) =>
  request.get('/system/users', { params })

// 用户详情
export const getUserDetailV2 = (id) =>
  request.get(`/system/users/${id}`)

// 创建用户
export const createUserV2 = (data) =>
  request.post('/system/users', data)

// 更新用户
export const updateUserV2 = (id, data) =>
  request.put(`/system/users/${id}`, data)

// 删除用户
export const deleteUserV2 = (id) =>
  request.delete(`/system/users/${id}`)

// ==================== 角色管理（新路由 /system/roles） ====================

// 角色列表
export const getRoleListV2 = (params) =>
  request.get('/system/roles', { params })

// 角色详情
export const getRoleDetailV2 = (id) =>
  request.get(`/system/roles/${id}`)

// 创建角色
export const createRoleV2 = (data) =>
  request.post('/system/roles', data)

// 更新角色
export const updateRoleV2 = (id, data) =>
  request.put(`/system/roles/${id}`, data)

// 删除角色
export const deleteRoleV2 = (id) =>
  request.delete(`/system/roles/${id}`)

// 分配权限给角色
export const assignRolePermissionsV2 = (id, data) =>
  request.post(`/system/roles/${id}/permissions`, data)

// ==================== 权限管理（新路由 /system/permissions） ====================

// 权限列表
export const getPermissionListV2 = (params) =>
  request.get('/system/permissions', { params })

// 创建权限
export const createPermissionV2 = (data) =>
  request.post('/system/permissions', data)

// 更新权限
export const updatePermissionV2 = (id, data) =>
  request.put(`/system/permissions/${id}`, data)

// 删除权限
export const deletePermissionV2 = (id) =>
  request.delete(`/system/permissions/${id}`)

// ==================== 部门管理（新路由 /system/departments） ====================

// 部门树列表
export const getDepartmentTreeV2 = (params) =>
  request.get('/system/departments', { params })

// 创建部门
export const createDepartmentV2 = (data) =>
  request.post('/system/departments', data)

// 更新部门
export const updateDepartmentV2 = (id, data) =>
  request.put(`/system/departments/${id}`, data)

// 删除部门
export const deleteDepartmentV2 = (id) =>
  request.delete(`/system/departments/${id}`)
