import request from './request'

export const getConfigMapList = (params) => request.get('/k8s/configmap/list', { params })

export const createConfigMap = (data) => request.post('/k8s/configmap/create', data)

export const updateConfigMap = (data) => request.put('/k8s/configmap/update', data)

export const deleteConfigMap = (data) => request.post('/k8s/configmap/delete', data)
