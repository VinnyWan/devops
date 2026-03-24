import type { ApiResponse, PageParams } from '@/types/api'
import type { Department, User } from '@/types/system'
import {
  departmentAssignRolesPost,
  departmentCreatePost,
  departmentDeletePost,
  departmentListPost,
  departmentUpdatePost,
  departmentUsersCreatePost,
  departmentUsersDeletePost,
  departmentUsersListPost,
  departmentUsersTransferPost,
  departmentUsersUpdatePost,
} from '@/api/generated/department.api'

export function getDepartmentTree() {
  return departmentListPost({}) as unknown as Promise<{
    data: { code: number; message: string; data: Department[] }
  }>
}

export function createDepartment(data: Partial<Department>) {
  return departmentCreatePost(data as Parameters<typeof departmentCreatePost>[0]) as Promise<{
    data: ApiResponse<unknown>
  }>
}

export function updateDepartment(data: Partial<Department>) {
  return departmentUpdatePost(data as Parameters<typeof departmentUpdatePost>[0]) as Promise<{
    data: ApiResponse<unknown>
  }>
}

export function deleteDepartment(id: number) {
  return departmentDeletePost({ id } as Parameters<typeof departmentDeletePost>[0])
}

export function assignDepartmentRoles(departmentId: number, roleIds: number[]) {
  return departmentAssignRolesPost({ departmentId, roleIds } as Parameters<
    typeof departmentAssignRolesPost
  >[0])
}

export function getDepartmentUsers(params: PageParams & { departmentId: number }) {
  return departmentUsersListPost(
    params as Parameters<typeof departmentUsersListPost>[0],
  ) as unknown as Promise<{
    data: { code: number; message: string; data: User[]; total?: number }
  }>
}

export function createDepartmentUser(data: Partial<User> & { departmentId: number }) {
  return departmentUsersCreatePost(data as Parameters<typeof departmentUsersCreatePost>[0])
}

export function updateDepartmentUser(data: Partial<User>) {
  return departmentUsersUpdatePost(data as Parameters<typeof departmentUsersUpdatePost>[0])
}

export function deleteDepartmentUser(id: number) {
  return departmentUsersDeletePost({ id } as Parameters<typeof departmentUsersDeletePost>[0])
}

export function transferUser(userId: number, targetDepartmentId: number) {
  return departmentUsersTransferPost({
    userId,
    toDepartmentId: targetDepartmentId,
  } as Parameters<typeof departmentUsersTransferPost>[0])
}
