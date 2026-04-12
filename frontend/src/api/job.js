import request from './request'

export const getJobList = (params) => request.get('/k8s/job/list', { params })
export const getJobDetail = (params) => request.get('/k8s/job/detail', { params })
export const getJobPods = (params) => request.get('/k8s/job/pods', { params })
export const getJobYAML = (params) => request.get('/k8s/job/yaml', { params })
export const createJob = (data, params) => request.post('/k8s/job/create' + (params?.namespace ? '?namespace=' + params.namespace : ''), data)
export const deleteJob = (data) => request.post('/k8s/job/delete', data)
