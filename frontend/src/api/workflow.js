import request from './request'

export const getOrderList = (params) => request.get('/orders', { params })
export const getOrderDetail = (id) => request.get(`/orders/${id}`)
export const createOrder = (data) => request.post('/orders', data)
export const submitOrder = (id) => request.post(`/orders/${id}/submit`)
export const approveOrder = (id, comment) => request.post(`/orders/${id}/approve`, { comment })
export const rejectOrder = (id, comment) => request.post(`/orders/${id}/reject`, { comment })
export const executeOrder = (id) => request.post(`/orders/${id}/execute`)
