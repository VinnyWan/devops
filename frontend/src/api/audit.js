import request from './request'

export const getAuditList = (params) => request.get('/audit/list', { params })
