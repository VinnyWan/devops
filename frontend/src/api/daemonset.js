import request from './request'

export const getDaemonSetList = (params) => request.get('/k8s/daemonset/list', { params })

export const createDaemonSet = (data) => request.post('/k8s/daemonset/create', data)

export const deleteDaemonSet = (data) => request.post('/k8s/daemonset/delete', data)
