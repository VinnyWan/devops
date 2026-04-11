import request from './request'

export const getDeploymentList = (params) => request.get('/k8s/deployment/list', { params })

export const createDeployment = (data) => request.post('/k8s/deployment/create', data)

export const updateDeployment = (data) => request.put('/k8s/deployment/update', data)

export const deleteDeployment = (data) => request.delete('/k8s/deployment/delete', { data })

export const restartDeployment = (data) => request.post('/k8s/deployment/restart', data)

export const scaleDeployment = (data) => request.post('/k8s/deployment/scale', data)
