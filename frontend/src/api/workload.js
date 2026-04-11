import request from './request'

export const getPodList = (params) => request.get('/k8s/pod/list', { params })

export const getPodDetail = (params) => request.get('/k8s/pod/detail', { params })

export const deletePod = (data) => request.delete('/k8s/pod/delete', { data })

export const getPodLogs = (params) => request.get('/k8s/pod/logs', { params })

export const getPodEvents = (params) => request.get('/k8s/pod/events', { params })

export const getPodTerminal = (data) => request.post('/k8s/pod/terminal', data)
