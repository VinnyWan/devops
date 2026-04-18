import request from '../request'

export const getCloudAccountList = (params) => request.get('/cmdb/cloud-account/list', { params })
export const getCloudAccountDetail = (params) => request.get('/cmdb/cloud-account/detail', { params })
export const createCloudAccount = (data) => request.post('/cmdb/cloud-account/create', data)
export const updateCloudAccount = (data) => request.post('/cmdb/cloud-account/update', data)
export const deleteCloudAccount = (data) => request.post('/cmdb/cloud-account/delete', data)
export const syncCloudAccount = (data) => request.post('/cmdb/cloud-account/sync', data)
export const getCloudResources = (params) => request.get('/cmdb/cloud-account/resources', { params })
