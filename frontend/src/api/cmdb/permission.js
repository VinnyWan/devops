import request from '../request'

export const getPermissionList = (params) => request.get('/cmdb/permission/list', { params })
export const createPermission = (data) => request.post('/cmdb/permission/create', data)
export const updatePermission = (data) => request.post('/cmdb/permission/update', data)
export const deletePermission = (data) => request.post('/cmdb/permission/delete', data)
export const getMyHosts = () => request.get('/cmdb/permission/my-hosts')
export const checkPermission = (params) => request.get('/cmdb/permission/check', { params })
export const getGroupHostCount = (params) => request.get('/cmdb/permission/group-host-count', { params })
