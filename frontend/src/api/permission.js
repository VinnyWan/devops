import request from './request'

// 权限列表（分页，支持按资源过滤）
export const getPermissionList = (params) => request.get('/permission/list', { params })

// 所有权限（不分页，用于权限选择器）
export const getAllPermissions = () => request.get('/permission/all')

// 权限详情
export const getPermissionDetail = (params) => request.get('/permission/detail', { params })
