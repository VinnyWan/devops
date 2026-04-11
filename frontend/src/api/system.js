import request from './request'

export const getUserList = (params) => request.get('/user/list', { params })

export const createUser = (data) => request.post('/user/create', data)

export const updateUser = (data) => request.put('/user/update', data)

export const deleteUser = (id) => request.delete(`/user/delete/${id}`)

export const getRoleList = (params) => request.get('/role/list', { params })

export const getDepartmentList = () => request.get('/department/list')
