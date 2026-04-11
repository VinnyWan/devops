import request from './request'

// 角色列表(分页)
export const getRoleList = (params) => request.get('/role/list', { params })

// 角色详情
export const getRoleDetail = (params) => request.get('/role/detail', { params })

// 创建角色
export const createRole = (data) => request.post('/role/create', data)

// 更新角色
export const updateRole = (data) => request.post('/role/update', data)

// 删除角色
export const deleteRole = (data) => request.post('/role/delete', data)

// 分配权限
export const assignPermissions = (data) => request.post('/role/assign-permissions', data)
