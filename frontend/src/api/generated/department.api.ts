import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  DepartmentAssignRolesPostRequest,
  DepartmentAssignRolesPostResponse,
  DepartmentCreatePostRequest,
  DepartmentCreatePostResponse,
  DepartmentDeletePostParams,
  DepartmentDeletePostResponse,
  DepartmentListPostParams,
  DepartmentListPostResponse,
  DepartmentUpdatePostRequest,
  DepartmentUpdatePostResponse,
  DepartmentUsersCreatePostRequest,
  DepartmentUsersCreatePostResponse,
  DepartmentUsersDeletePostParams,
  DepartmentUsersDeletePostResponse,
  DepartmentUsersListPostParams,
  DepartmentUsersListPostResponse,
  DepartmentUsersTransferPostRequest,
  DepartmentUsersTransferPostResponse,
  DepartmentUsersUpdatePostRequest,
  DepartmentUsersUpdatePostResponse,
} from '@/types/generated/department.types'

export function departmentAssignRolesPost(data: DepartmentAssignRolesPostRequest) {
  return http.post<ApiResponse<DepartmentAssignRolesPostResponse>>(
    '/v1/department/assign-roles',
    data,
  )
}

export function departmentCreatePost(data: DepartmentCreatePostRequest) {
  return http.post<ApiResponse<DepartmentCreatePostResponse>>('/v1/department/create', data)
}

export function departmentDeletePost(params: DepartmentDeletePostParams) {
  return http.post<ApiResponse<DepartmentDeletePostResponse>>('/v1/department/delete', undefined, {
    params,
  })
}

export function departmentListPost(params: DepartmentListPostParams) {
  return http.get<ApiResponse<DepartmentListPostResponse>>('/v1/department/list', {
    params,
  })
}

export function departmentUpdatePost(data: DepartmentUpdatePostRequest) {
  return http.post<ApiResponse<DepartmentUpdatePostResponse>>('/v1/department/update', data)
}

export function departmentUsersCreatePost(data: DepartmentUsersCreatePostRequest) {
  return http.post<ApiResponse<DepartmentUsersCreatePostResponse>>(
    '/v1/department/users/create',
    data,
  )
}

export function departmentUsersDeletePost(params: DepartmentUsersDeletePostParams) {
  return http.post<ApiResponse<DepartmentUsersDeletePostResponse>>(
    '/v1/department/users/delete',
    undefined,
    { params },
  )
}

export function departmentUsersListPost(params: DepartmentUsersListPostParams) {
  return http.post<ApiResponse<DepartmentUsersListPostResponse>>(
    '/v1/department/users/list',
    undefined,
    { params },
  )
}

export function departmentUsersTransferPost(data: DepartmentUsersTransferPostRequest) {
  return http.post<ApiResponse<DepartmentUsersTransferPostResponse>>(
    '/v1/department/users/transfer',
    data,
  )
}

export function departmentUsersUpdatePost(data: DepartmentUsersUpdatePostRequest) {
  return http.post<ApiResponse<DepartmentUsersUpdatePostResponse>>(
    '/v1/department/users/update',
    data,
  )
}
