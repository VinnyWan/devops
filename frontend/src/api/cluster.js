import request from './request'

export const getClusterList = (params) => request.get('/k8s/cluster/list', { params })
export const getClusterDetail = (name) => request.get('/k8s/cluster/detail', { params: { name } })
export const createCluster = (data) => request.post('/k8s/cluster/create', data)
export const updateCluster = (data) => request.post('/k8s/cluster/update', data)
export const deleteCluster = (id) => request.post('/k8s/cluster/delete', { id })
export const getClusterNetworkStats = (name) => request.get('/k8s/cluster/stats/network', { params: { name } })
export const getClusterStorageStats = (name) => request.get('/k8s/cluster/stats/storage', { params: { name } })
export const getClusterWorkloadStats = (name) => request.get('/k8s/cluster/stats/workload', { params: { name } })
export const getClusterEvents = (params) => request.get('/k8s/cluster/events', { params })
