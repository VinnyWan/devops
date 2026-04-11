import request from './request'

export const getIngressList = (params) => request.get('/k8s/ingress/list', { params })

export const createIngress = (data) => request.post('/k8s/ingress/create', data)

export const deleteIngress = (data) => request.delete('/k8s/ingress/delete', { data })
