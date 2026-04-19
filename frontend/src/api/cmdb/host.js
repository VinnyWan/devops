import request from '../request'

const normalizeHostPayload = (data) => ({
  ...data,
  groupId: data.groupId || undefined,
  credentialId: data.credentialId || undefined
})

export const getHostList = (params) => request.get('/cmdb/host/list', { params })
export const getHostDetail = (params) => request.get('/cmdb/host/detail', { params })
export const createHost = (data) => request.post('/cmdb/host/create', normalizeHostPayload(data))
export const batchCreateHost = (data) => request.post('/cmdb/host/batch', data)
export const updateHost = (data) => request.post('/cmdb/host/update', normalizeHostPayload(data))
export const deleteHost = (data) => request.post('/cmdb/host/delete', data)
export const testHost = (data) => request.post('/cmdb/host/test', data)

export function checkHostPermission(hostId, action) {
  return request({
    url: '/cmdb/permissions/check',
    method: 'get',
    params: { host_id: hostId, action }
  })
}
