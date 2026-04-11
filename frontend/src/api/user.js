import request from './request'

export const login = (data) => request.post('/user/login', data)

export const getUserInfo = () => request.get('/user/info')

export const logout = () => request.post('/user/logout')
