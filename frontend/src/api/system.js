import request from './request'

export const getUserList = (params) => request.get('/system/users', { params })
export const createUser = (data) => request.post('/system/users', data)
export const updateUser = (data) => request.put(`/system/users/${data.id}`, data)
export const deleteUser = (id) => request.delete(`/system/users/${id}`)
export const getRoleList = (params) => request.get('/system/roles', { params })
export const getDepartmentList = () => request.get('/system/departments/tree')
