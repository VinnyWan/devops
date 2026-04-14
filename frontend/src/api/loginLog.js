import request from './request'

export const getLoginLogList = (params) => request.get('/login-log/list', { params })
