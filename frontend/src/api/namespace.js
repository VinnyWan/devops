import request from './request'

export const getNamespaceList = (params) => request.get('/k8s/namespace/list', { params })

export const createNamespace = (data) => request.post('/k8s/namespace/create', data)

export const deleteNamespace = (data) => request.delete('/k8s/namespace/delete', { data })
