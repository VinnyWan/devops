import request from './request'

export const getCronJobList = (params) => request.get('/k8s/cronjob/list', { params })
export const getCronJobDetail = (params) => request.get('/k8s/cronjob/detail', { params })
export const getCronJobPods = (params) => request.get('/k8s/cronjob/pods', { params })
export const getCronJobYAML = (params) => request.get('/k8s/cronjob/yaml', { params })
export const createCronJob = (data, params) => request.post('/k8s/cronjob/create' + (params?.namespace ? '?namespace=' + params.namespace : ''), data)
export const updateCronJobYAML = (data) => request.post('/k8s/cronjob/yaml/update', data)
export const suspendCronJob = (data) => request.post('/k8s/cronjob/suspend', data)
export const deleteCronJob = (data) => request.post('/k8s/cronjob/delete', data)
