import request from '../request'

export const getGroupTree = () => request.get('/cmdb/group/tree')
export const getGroupDetail = (params) => request.get('/cmdb/group/detail', { params })
export const createGroup = (data) => request.post('/cmdb/group/create', data)
export const updateGroup = (data) => request.post('/cmdb/group/update', data)
export const deleteGroup = (data) => request.post('/cmdb/group/delete', data)
