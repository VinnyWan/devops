import request from './request'

export const getPermissionList = (params) => request.get('/system/permissions', { params })
export const getAllPermissions = () => request.get('/system/permissions/all')
export const getPermissionDetail = (params) => request.get(`/system/permissions/${params.id}`)
