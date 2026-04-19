import request from './request'

// ==================== 租户管理 ====================

// 租户列表
export const getTenantList = (params) =>
  request.get('/platform/tenants', { params })

// 租户详情
export const getTenantDetail = (id) =>
  request.get(`/platform/tenants/${id}`)

// 创建租户
export const createTenant = (data) =>
  request.post('/platform/tenants', data)

// 更新租户
export const updateTenant = (id, data) =>
  request.put(`/platform/tenants/${id}`, data)

// 删除租户
export const deleteTenant = (id) =>
  request.delete(`/platform/tenants/${id}`)
