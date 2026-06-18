import request from './request'

export const getToolList = (category) => request.get('/tools', { params: { category } })
export const getToolDetail = (id) => request.get(`/tools/${id}`)
export const installTool = (id, data) => request.post(`/tools/${id}/install`, data)
export const checkToolStatus = (id, data) => request.post(`/tools/${id}/check`, data)
export const getInstallations = (hostId) => request.get('/tools/installations', { params: { hostId } })

// Templates
export const listTemplates = (params) => request.get('/tools/templates', { params })
export const getTemplate = (id) => request.get(`/tools/templates/${id}`)
export const saveTemplate = (data) => request.post('/tools/templates', data)
export const updateTemplate = (id, data) => request.put(`/tools/templates/${id}`, data)
export const deleteTemplate = (id) => request.delete(`/tools/templates/${id}`)

// Versions
export const listTemplateVersions = (id) => request.get(`/tools/templates/${id}/versions`)
export const saveTemplateVersion = (id, data) => request.post(`/tools/templates/${id}/versions`, data)
export const deleteTemplateVersion = (templateId, versionId) => request.delete(`/tools/templates/${templateId}/versions/${versionId}`)
