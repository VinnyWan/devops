import request from './request'

export const getConnections = (type) => request.get('/db-connections', { params: { type } })
export const createConnection = (data) => request.post('/db-connections', data)
export const testConnection = (id) => request.post(`/db-connections/${id}/test`)
export const deleteConnection = (id) => request.delete(`/db-connections/${id}`)
export const executeSQL = (data) => request.post('/sql/execute', data)
export const getSqlRecords = (params) => request.get('/sql/records', { params })
