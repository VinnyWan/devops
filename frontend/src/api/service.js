import request from './request'

export const getServiceList = (params) => request.get('/k8s/service/list', { params })

export const createService = (data) => request.post('/k8s/service/create', data)

export const deleteService = (data) => request.post('/k8s/service/delete', data)

export const getServiceYAML = (params) => request.get('/k8s/resource/yaml', { params })

export const updateServiceByYAML = (data) => request.post('/k8s/service/yaml/update', data)
