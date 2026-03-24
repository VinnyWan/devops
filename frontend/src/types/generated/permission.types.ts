import type { ModelPermission, ServiceCreatePermissionRequest } from './common.types'

export type PermissionAllPostParams = Record<string, never>

export type PermissionAllPostRequest = Record<string, never>

export type PermissionAllPostResponse = Record<string, unknown>

export type PermissionCreatePostParams = Record<string, never>

export type PermissionCreatePostRequest = ServiceCreatePermissionRequest

export type PermissionCreatePostResponse = Record<string, unknown>

export type PermissionDeletePostParams = { id: number }

export type PermissionDeletePostRequest = Record<string, never>

export type PermissionDeletePostResponse = Record<string, unknown>

export type PermissionDetailPostParams = { id: number }

export type PermissionDetailPostRequest = Record<string, never>

export type PermissionDetailPostResponse = Record<string, unknown>

export type PermissionListPostParams = {
  page?: number
  pageSize?: number
  resource?: string
  keyword?: string
}

export type PermissionListPostRequest = Record<string, never>

export type PermissionListPostResponse = Record<string, unknown>

export type PermissionUpdatePostParams = Record<string, never>

export type PermissionUpdatePostRequest = ModelPermission

export type PermissionUpdatePostResponse = Record<string, unknown>
