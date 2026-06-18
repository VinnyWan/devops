import request from './request'

export const listLogSources = (params) => request.get('/log/sources', { params })
export const saveLogSource = (data) => request.post('/log/sources', data)
export const updateLogSource = (id, data) => request.put(`/log/sources/${id}`, data)
export const deleteLogSource = (id) => request.delete(`/log/sources/${id}`)
export const testLogSourceConnection = (id) => request.post(`/log/sources/${id}/test`)

export const searchLogs = (data) => request.post('/log/search', data)
export const exportLogs = (data) => request.post('/log/export', data, { responseType: 'blob' })
