import type { ApiResponse } from '@/types/api'
import type { Role, User, Department } from '@/types/system'
import {
  roleAssignDepartmentsPost,
  roleAssignPermissionsPost,
  roleAssignUsersPost,
  roleCreatePost,
  roleDeletePost,
  roleDepartmentsPost,
  roleDetailPost,
  roleListPost,
  roleUpdatePost,
  roleUsersPost,
} from '@/api/generated/role.api'

export function getRoleList(params?: { keyword?: string }) {
  return roleListPost((params || {}) as Parameters<typeof roleListPost>[0]) as unknown as Promise<{
    data: { code: number; message: string; data: Role[] }
  }>
}

export function getRoleDetail(id: number) {
  return roleDetailPost({ id } as Parameters<typeof roleDetailPost>[0]) as unknown as Promise<{
    data: { code: number; message: string; data: Role }
  }>
}

export function createRole(data: Partial<Role>) {
  return roleCreatePost(data as Parameters<typeof roleCreatePost>[0]) as Promise<{
    data: ApiResponse<unknown>
  }>
}

export function updateRole(data: Partial<Role>) {
  return roleUpdatePost(data as Parameters<typeof roleUpdatePost>[0]) as Promise<{
    data: ApiResponse<unknown>
  }>
}

export function deleteRole(id: number) {
  return roleDeletePost({ id } as Parameters<typeof roleDeletePost>[0])
}

export function assignRolePermissions(roleId: number, permissionIds: number[]) {
  return roleAssignPermissionsPost({ roleId, permissionIds } as Parameters<
    typeof roleAssignPermissionsPost
  >[0])
}

export function assignRoleUsers(roleId: number, userIds: number[]) {
  return roleAssignUsersPost({ roleId, userIds } as Parameters<typeof roleAssignUsersPost>[0])
}

export function assignRoleDepartments(roleId: number, departmentIds: number[]) {
  return roleAssignDepartmentsPost({ roleId, departmentIds } as Parameters<
    typeof roleAssignDepartmentsPost
  >[0])
}

export function getRoleUsers(roleId: number) {
  return roleUsersPost({ id: roleId } as Parameters<
    typeof roleUsersPost
  >[0]) as unknown as Promise<{
    data: { code: number; message: string; data: User[] }
  }>
}

export function getRoleDepartments(roleId: number) {
  return roleDepartmentsPost({ id: roleId } as Parameters<
    typeof roleDepartmentsPost
  >[0]) as unknown as Promise<{
    data: { code: number; message: string; data: Department[] }
  }>
}
