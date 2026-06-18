import request from './request'

export const listHarborConfigs = (params) => request.get('/harbor/configs', { params })
export const saveHarborConfig = (data) => request.post('/harbor/configs', data)
export const updateHarborConfig = (id, data) => request.put(`/harbor/configs/${id}`, data)
export const deleteHarborConfig = (id) => request.delete(`/harbor/configs/${id}`)
export const testHarborConnection = (data) => request.post('/harbor/configs/test', data)

export const listProjects = (params) => request.get('/harbor/projects', { params })
export const listRepositories = (projectName, params) => request.get(`/harbor/projects/${encodeURIComponent(projectName)}/repos`, { params })
export const listArtifacts = (projectName, repoName, params) => request.get(`/harbor/projects/${encodeURIComponent(projectName)}/repos/${encodeURIComponent(repoName)}/artifacts`, { params })
export const deleteArtifact = (projectName, repoName, data) => request.delete(`/harbor/projects/${encodeURIComponent(projectName)}/repos/${encodeURIComponent(repoName)}/artifacts`, { data })
