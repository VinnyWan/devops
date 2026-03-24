import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  AuthOidcCallbackGetParams,
  AuthOidcCallbackGetResponse,
  AuthOidcLoginGetResponse,
  UserLoginPostRequest,
  UserLoginPostResponse,
  UserLogoutPostResponse,
} from '@/types/generated/auth.types'

export function authOidcCallbackGet(params: AuthOidcCallbackGetParams) {
  return http.get<ApiResponse<AuthOidcCallbackGetResponse>>('/v1/auth/oidc/callback', { params })
}

export function authOidcLoginGet() {
  return http.get<ApiResponse<AuthOidcLoginGetResponse>>('/v1/auth/oidc/login')
}

export function userLoginPost(data: UserLoginPostRequest) {
  return http.post<ApiResponse<UserLoginPostResponse>>('/v1/user/login', data)
}

export function userLogoutPost() {
  return http.post<ApiResponse<UserLogoutPostResponse>>('/v1/user/logout')
}
