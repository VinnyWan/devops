import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  PermissionAllPostResponse,
  PermissionCreatePostRequest,
  PermissionCreatePostResponse,
  PermissionDeletePostParams,
  PermissionDeletePostResponse,
  PermissionDetailPostParams,
  PermissionDetailPostResponse,
  PermissionListPostParams,
  PermissionListPostResponse,
  PermissionUpdatePostRequest,
  PermissionUpdatePostResponse,
} from '@/types/generated/permission.types'

export function permissionAllPost() {
  return http.get<ApiResponse<PermissionAllPostResponse>>('/v1/permission/all')
}

export function permissionCreatePost(data: PermissionCreatePostRequest) {
  return http.post<ApiResponse<PermissionCreatePostResponse>>('/v1/permission/create', data)
}

export function permissionDeletePost(params: PermissionDeletePostParams) {
  return http.post<ApiResponse<PermissionDeletePostResponse>>('/v1/permission/delete', undefined, {
    params,
  })
}

export function permissionDetailPost(params: PermissionDetailPostParams) {
  return http.get<ApiResponse<PermissionDetailPostResponse>>('/v1/permission/detail', {
    params,
  })
}

export function permissionListPost(params: PermissionListPostParams) {
  return http.get<ApiResponse<PermissionListPostResponse>>('/v1/permission/list', {
    params,
  })
}

export function permissionUpdatePost(data: PermissionUpdatePostRequest) {
  return http.post<ApiResponse<PermissionUpdatePostResponse>>('/v1/permission/update', data)
}
