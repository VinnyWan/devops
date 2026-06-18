import request from './request'

// Jenkins Config
export const listJenkinsConfigs = (params) => request.get('/cicd/jenkins', { params })
export const saveJenkinsConfig = (data) => request.post('/cicd/jenkins', data)
export const updateJenkinsConfig = (id, data) => request.put(`/cicd/jenkins/${id}`, data)
export const deleteJenkinsConfig = (id) => request.delete(`/cicd/jenkins/${id}`)
export const testJenkinsConnection = (data) => request.post('/cicd/jenkins/test', data)

// Jobs
export const listJobs = (configId, params) => request.get(`/cicd/jenkins/${configId}/jobs`, { params })
export const triggerBuild = (configId, data) => request.post(`/cicd/jenkins/${configId}/build`, data)
export const listBuilds = (configId, params) => request.get(`/cicd/jenkins/${configId}/builds`, { params })
export const getBuildLog = (configId, params) => request.get(`/cicd/jenkins/${configId}/build-log`, { params })

// Pipelines
export const listPipelines = (params) => request.get('/cicd/pipelines', { params })
export const savePipeline = (data) => request.post('/cicd/pipelines', data)
export const updatePipeline = (id, data) => request.put(`/cicd/pipelines/${id}`, data)
export const deletePipeline = (id) => request.delete(`/cicd/pipelines/${id}`)
