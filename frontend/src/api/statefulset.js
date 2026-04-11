import request from './request'

export const getStatefulSetList = (params) => request.get('/k8s/statefulset/list', { params })

export const createStatefulSet = (data) => request.post('/k8s/statefulset/create', data)

export const deleteStatefulSet = (data) => request.delete('/k8s/statefulset/delete', { data })

export const scaleStatefulSet = (data) => request.post('/k8s/statefulset/scale', data)
