import request from './request'

export const getDepartmentList = (params) => request.get('/system/departments/tree', { params })
export const createDepartment = (data) => request.post('/system/departments', data)
export const updateDepartment = (data) => request.put(`/system/departments/${data.id}`, data)
export const deleteDepartment = (data) => request.delete(`/system/departments/${data.id}`)
export const getDepartmentUsers = (params) => request.get(`/system/departments/${params.departmentId}/users`, { params: { page: params.page, pageSize: params.pageSize, keyword: params.keyword } })
