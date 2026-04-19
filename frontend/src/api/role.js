import request from './request'

export const getRoleList = (params) => request.get('/system/roles', { params })
export const getRoleDetail = (params) => request.get(`/system/roles/${params.id}`)
export const createRole = (data) => request.post('/system/roles', data)
export const updateRole = (data) => request.put(`/system/roles/${data.id}`, data)
export const deleteRole = (data) => request.delete(`/system/roles/${data.id}`)
export const assignPermissions = (data) => request.put(`/system/roles/${data.roleId}/permissions`, { permissionIds: data.permissionIds })
