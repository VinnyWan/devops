import request from '../request'

export const getCredentialList = (params) => request.get('/cmdb/credential/list', { params })
export const getCredentialDetail = (params) => request.get('/cmdb/credential/detail', { params })
export const createCredential = (data) => request.post('/cmdb/credential/create', data)
export const updateCredential = (data) => request.post('/cmdb/credential/update', data)
export const deleteCredential = (data) => request.post('/cmdb/credential/delete', data)
