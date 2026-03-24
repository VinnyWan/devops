import type {
  ServiceChangePasswordRequest,
  ServiceRegisterRequest,
  ServiceUpdateUserRequest,
} from './common.types'

export type UserAssignRolesPostParams = Record<string, never>

export type UserAssignRolesPostRequest = Record<string, unknown>

export type UserAssignRolesPostResponse = Record<string, unknown>

export type UserChangePasswordPostParams = Record<string, never>

export type UserChangePasswordPostRequest = ServiceChangePasswordRequest

export type UserChangePasswordPostResponse = Record<string, unknown>

export type UserDeletePostParams = { id: number }

export type UserDeletePostRequest = Record<string, never>

export type UserDeletePostResponse = Record<string, unknown>

export type UserDetailPostParams = { id: number }

export type UserDetailPostRequest = Record<string, never>

export type UserDetailPostResponse = Record<string, unknown>

export type UserInfoPostParams = Record<string, never>

export type UserInfoPostRequest = Record<string, never>

export type UserInfoPostResponse = Record<string, unknown>

export type UserListPostParams = { page?: number; pageSize?: number; keyword?: string }

export type UserListPostRequest = Record<string, never>

export type UserListPostResponse = Record<string, unknown>

export type UserLockPostParams = { id: number }

export type UserLockPostRequest = Record<string, never>

export type UserLockPostResponse = Record<string, unknown>

export type UserPermissionsPostParams = Record<string, never>

export type UserPermissionsPostRequest = Record<string, never>

export type UserPermissionsPostResponse = Record<string, unknown>

export type UserRegisterPostParams = Record<string, never>

export type UserRegisterPostRequest = ServiceRegisterRequest

export type UserRegisterPostResponse = Record<string, unknown>

export type UserResetPasswordPostParams = Record<string, never>

export type UserResetPasswordPostRequest = Record<string, unknown>

export type UserResetPasswordPostResponse = Record<string, unknown>

export type UserUnlockPostParams = { id: number }

export type UserUnlockPostRequest = Record<string, never>

export type UserUnlockPostResponse = Record<string, unknown>

export type UserUpdatePostParams = Record<string, never>

export type UserUpdatePostRequest = ServiceUpdateUserRequest

export type UserUpdatePostResponse = Record<string, unknown>
