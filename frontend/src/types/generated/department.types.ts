import type {
  ServiceCreateDepartmentRequest,
  ServiceCreateDeptUserRequest,
  ServiceTransferUserDepartmentRequest,
  ServiceUpdateDepartmentRequest,
  ServiceUpdateDeptUserRequest,
} from './common.types'

export type DepartmentAssignRolesPostParams = Record<string, never>

export type DepartmentAssignRolesPostRequest = Record<string, unknown>

export type DepartmentAssignRolesPostResponse = Record<string, unknown>

export type DepartmentCreatePostParams = Record<string, never>

export type DepartmentCreatePostRequest = ServiceCreateDepartmentRequest

export type DepartmentCreatePostResponse = Record<string, unknown>

export type DepartmentDeletePostParams = { id: number }

export type DepartmentDeletePostRequest = Record<string, never>

export type DepartmentDeletePostResponse = Record<string, unknown>

export type DepartmentListPostParams = { keyword?: string }

export type DepartmentListPostRequest = Record<string, never>

export type DepartmentListPostResponse = Record<string, unknown>

export type DepartmentUpdatePostParams = Record<string, never>

export type DepartmentUpdatePostRequest = ServiceUpdateDepartmentRequest

export type DepartmentUpdatePostResponse = Record<string, unknown>

export type DepartmentUsersCreatePostParams = Record<string, never>

export type DepartmentUsersCreatePostRequest = ServiceCreateDeptUserRequest

export type DepartmentUsersCreatePostResponse = Record<string, unknown>

export type DepartmentUsersDeletePostParams = { id: number }

export type DepartmentUsersDeletePostRequest = Record<string, never>

export type DepartmentUsersDeletePostResponse = Record<string, unknown>

export type DepartmentUsersListPostParams = {
  departmentId?: number
  page?: number
  pageSize?: number
  keyword?: string
}

export type DepartmentUsersListPostRequest = Record<string, never>

export type DepartmentUsersListPostResponse = Record<string, unknown>

export type DepartmentUsersTransferPostParams = Record<string, never>

export type DepartmentUsersTransferPostRequest = ServiceTransferUserDepartmentRequest

export type DepartmentUsersTransferPostResponse = Record<string, unknown>

export type DepartmentUsersUpdatePostParams = Record<string, never>

export type DepartmentUsersUpdatePostRequest = ServiceUpdateDeptUserRequest

export type DepartmentUsersUpdatePostResponse = Record<string, unknown>
