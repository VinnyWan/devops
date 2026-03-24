import type { LoginParams, LoginResult, UserInfo } from '@/types/auth'
import { userInfoPost } from '@/api/generated/user.api'
import { userLoginPost, userLogoutPost } from '@/api/generated/auth.api'
import { unwrapResponseData } from '@/api/service'

export function login(data: LoginParams) {
  return unwrapResponseData<LoginResult>(userLoginPost(data as Parameters<typeof userLoginPost>[0]))
}

export function logout() {
  return unwrapResponseData<unknown>(userLogoutPost())
}

export function getCurrentUser() {
  return unwrapResponseData<UserInfo>(userInfoPost())
}
