import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  RoleAssignDepartmentsPostRequest,
  RoleAssignDepartmentsPostResponse,
  RoleAssignPermissionsPostRequest,
  RoleAssignPermissionsPostResponse,
  RoleAssignUsersPostRequest,
  RoleAssignUsersPostResponse,
  RoleCreatePostRequest,
  RoleCreatePostResponse,
  RoleDeletePostParams,
  RoleDeletePostResponse,
  RoleDepartmentsPostParams,
  RoleDepartmentsPostResponse,
  RoleDetailPostParams,
  RoleDetailPostResponse,
  RoleListPostParams,
  RoleListPostResponse,
  RoleUpdatePostRequest,
  RoleUpdatePostResponse,
  RoleUsersPostParams,
  RoleUsersPostResponse,
} from '@/types/generated/role.types'

export function roleAssignDepartmentsPost(data: RoleAssignDepartmentsPostRequest) {
  return http.post<ApiResponse<RoleAssignDepartmentsPostResponse>>(
    '/v1/role/assign-departments',
    data,
  )
}

export function roleAssignPermissionsPost(data: RoleAssignPermissionsPostRequest) {
  return http.post<ApiResponse<RoleAssignPermissionsPostResponse>>(
    '/v1/role/assign-permissions',
    data,
  )
}

export function roleAssignUsersPost(data: RoleAssignUsersPostRequest) {
  return http.post<ApiResponse<RoleAssignUsersPostResponse>>('/v1/role/assign-users', data)
}

export function roleCreatePost(data: RoleCreatePostRequest) {
  return http.post<ApiResponse<RoleCreatePostResponse>>('/v1/role/create', data)
}

export function roleDeletePost(params: RoleDeletePostParams) {
  return http.post<ApiResponse<RoleDeletePostResponse>>('/v1/role/delete', undefined, { params })
}

export function roleDepartmentsPost(params: RoleDepartmentsPostParams) {
  return http.get<ApiResponse<RoleDepartmentsPostResponse>>('/v1/role/departments', {
    params,
  })
}

export function roleDetailPost(params: RoleDetailPostParams) {
  return http.get<ApiResponse<RoleDetailPostResponse>>('/v1/role/detail', { params })
}

export function roleListPost(params: RoleListPostParams) {
  return http.get<ApiResponse<RoleListPostResponse>>('/v1/role/list', { params })
}

export function roleUpdatePost(data: RoleUpdatePostRequest) {
  return http.post<ApiResponse<RoleUpdatePostResponse>>('/v1/role/update', data)
}

export function roleUsersPost(params: RoleUsersPostParams) {
  return http.get<ApiResponse<RoleUsersPostResponse>>('/v1/role/users', { params })
}
