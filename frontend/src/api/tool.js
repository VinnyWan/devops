import request from './request'

export const getToolList = (category) => request.get('/tools', { params: { category } })
export const getToolDetail = (id) => request.get(`/tools/${id}`)
export const installTool = (id, data) => request.post(`/tools/${id}/install`, data)
export const checkToolStatus = (id, data) => request.post(`/tools/${id}/check`, data)
export const getInstallations = (hostId) => request.get('/tools/installations', { params: { hostId } })
