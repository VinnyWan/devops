import request from './request'

// StorageClass
export const getStorageClassList = (params) => request.get('/k8s/storageclass/list', { params })
export const updateStorageClassYAML = (data) => request.post('/k8s/storageclass/yaml/update', data)

// PV
export const getPVList = (params) => request.get('/k8s/pv/list', { params })
export const updatePVYAML = (data) => request.post('/k8s/pv/yaml/update', data)

// PVC
export const getPVCList = (params) => request.get('/k8s/pvc/list', { params })
export const updatePVCYAML = (data) => request.post('/k8s/pvc/yaml/update', data)
export const deletePVC = (data) => request.post('/k8s/pvc/delete', data)
