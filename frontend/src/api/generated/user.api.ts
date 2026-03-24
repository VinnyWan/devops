import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  UserAssignRolesPostRequest,
  UserAssignRolesPostResponse,
  UserChangePasswordPostRequest,
  UserChangePasswordPostResponse,
  UserDeletePostParams,
  UserDeletePostResponse,
  UserDetailPostParams,
  UserDetailPostResponse,
  UserInfoPostResponse,
  UserListPostParams,
  UserListPostResponse,
  UserLockPostParams,
  UserLockPostResponse,
  UserPermissionsPostResponse,
  UserRegisterPostRequest,
  UserRegisterPostResponse,
  UserResetPasswordPostRequest,
  UserResetPasswordPostResponse,
  UserUnlockPostParams,
  UserUnlockPostResponse,
  UserUpdatePostRequest,
  UserUpdatePostResponse,
} from '@/types/generated/user.types'

export function userAssignRolesPost(data: UserAssignRolesPostRequest) {
  return http.post<ApiResponse<UserAssignRolesPostResponse>>('/v1/user/assign-roles', data)
}

export function userChangePasswordPost(data: UserChangePasswordPostRequest) {
  return http.post<ApiResponse<UserChangePasswordPostResponse>>('/v1/user/change-password', data)
}

export function userDeletePost(params: UserDeletePostParams) {
  return http.post<ApiResponse<UserDeletePostResponse>>('/v1/user/delete', undefined, { params })
}

export function userDetailPost(params: UserDetailPostParams) {
  return http.get<ApiResponse<UserDetailPostResponse>>('/v1/user/detail', { params })
}

export function userInfoPost() {
  return http.get<ApiResponse<UserInfoPostResponse>>('/v1/user/info')
}

export function userListPost(params: UserListPostParams) {
  return http.get<ApiResponse<UserListPostResponse>>('/v1/user/list', { params })
}

export function userLockPost(params: UserLockPostParams) {
  return http.post<ApiResponse<UserLockPostResponse>>('/v1/user/lock', undefined, { params })
}

export function userPermissionsPost() {
  return http.post<ApiResponse<UserPermissionsPostResponse>>('/v1/user/permissions')
}

export function userRegisterPost(data: UserRegisterPostRequest) {
  return http.post<ApiResponse<UserRegisterPostResponse>>('/v1/user/register', data)
}

export function userResetPasswordPost(data: UserResetPasswordPostRequest) {
  return http.post<ApiResponse<UserResetPasswordPostResponse>>('/v1/user/reset-password', data)
}

export function userUnlockPost(params: UserUnlockPostParams) {
  return http.post<ApiResponse<UserUnlockPostResponse>>('/v1/user/unlock', undefined, { params })
}

export function userUpdatePost(data: UserUpdatePostRequest) {
  return http.post<ApiResponse<UserUpdatePostResponse>>('/v1/user/update', data)
}
