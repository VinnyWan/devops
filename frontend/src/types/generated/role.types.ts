import type { ServiceCreateRoleRequest, ServiceUpdateRoleRequest } from './common.types'

export type RoleAssignDepartmentsPostParams = Record<string, never>

export type RoleAssignDepartmentsPostRequest = Record<string, unknown>

export type RoleAssignDepartmentsPostResponse = Record<string, unknown>

export type RoleAssignPermissionsPostParams = Record<string, never>

export type RoleAssignPermissionsPostRequest = Record<string, unknown>

export type RoleAssignPermissionsPostResponse = Record<string, unknown>

export type RoleAssignUsersPostParams = Record<string, never>

export type RoleAssignUsersPostRequest = Record<string, unknown>

export type RoleAssignUsersPostResponse = Record<string, unknown>

export type RoleCreatePostParams = Record<string, never>

export type RoleCreatePostRequest = ServiceCreateRoleRequest

export type RoleCreatePostResponse = Record<string, unknown>

export type RoleDeletePostParams = { id: number }

export type RoleDeletePostRequest = Record<string, never>

export type RoleDeletePostResponse = Record<string, unknown>

export type RoleDepartmentsPostParams = { id: number }

export type RoleDepartmentsPostRequest = Record<string, never>

export type RoleDepartmentsPostResponse = Record<string, unknown>

export type RoleDetailPostParams = { id: number }

export type RoleDetailPostRequest = Record<string, never>

export type RoleDetailPostResponse = Record<string, unknown>

export type RoleListPostParams = { page?: number; pageSize?: number; keyword?: string }

export type RoleListPostRequest = Record<string, never>

export type RoleListPostResponse = Record<string, unknown>

export type RoleUpdatePostParams = Record<string, never>

export type RoleUpdatePostRequest = ServiceUpdateRoleRequest

export type RoleUpdatePostResponse = Record<string, unknown>

export type RoleUsersPostParams = { id: number }

export type RoleUsersPostRequest = Record<string, never>

export type RoleUsersPostResponse = Record<string, unknown>
