import request from './request'

export const getNodeList = (params) => request.get('/k8s/nodes', { params })

export const cordonNode = (data) => request.post('/k8s/node/cordon', data)

export const drainNode = (data) => request.post('/k8s/node/drain', data)

export const updateNodeLabel = (data) => request.post('/k8s/node/label', data)

export const updateNodeTaint = (data) => request.post('/k8s/node/taint', data)
