import request from './request'

export const getConfigMapList = (params) => request.get('/k8s/configmap/list', { params })

export const getConfigMapYAML = (params) => request.get('/k8s/configmap/yaml', { params })

export const createConfigMap = (data) => request.post('/k8s/configmap/create', data)

export const updateConfigMapByYAML = (data) => request.post('/k8s/configmap/yaml/update', data)

export const deleteConfigMap = (data) => request.post('/k8s/configmap/delete', data)
