import request from './request'

// 部门树列表
export const getDepartmentList = (params) => request.get('/department/list', { params })

// 创建部门
export const createDepartment = (data) => request.post('/department/create', data)

// 更新部门
export const updateDepartment = (data) => request.post('/department/update', data)

// 删除部门
export const deleteDepartment = (data) => request.post('/department/delete', data)

// 部门用户列表
export const getDepartmentUsers = (params) => request.get('/department/users/list', { params })
