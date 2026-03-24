import type { ApiResponse, PageParams } from '@/types/api'
import type { Permission } from '@/types/system'
import {
  permissionAllPost,
  permissionCreatePost,
  permissionDeletePost,
  permissionDetailPost,
  permissionListPost,
  permissionUpdatePost,
} from '@/api/generated/permission.api'

export function getPermissionList(params: PageParams & { keyword?: string }) {
  return permissionListPost(
    params as Parameters<typeof permissionListPost>[0],
  ) as unknown as Promise<{
    data: { code: number; message: string; data: Permission[]; total?: number }
  }>
}

export function getAllPermissions() {
  return permissionAllPost() as unknown as Promise<{
    data: { code: number; message: string; data: Permission[] }
  }>
}

export function getPermissionDetail(id: number) {
  return permissionDetailPost({ id } as Parameters<
    typeof permissionDetailPost
  >[0]) as unknown as Promise<{
    data: { code: number; message: string; data: Permission }
  }>
}

export function createPermission(data: Partial<Permission>) {
  return permissionCreatePost(data as Parameters<typeof permissionCreatePost>[0]) as Promise<{
    data: ApiResponse<unknown>
  }>
}

export function updatePermission(data: Partial<Permission>) {
  return permissionUpdatePost(data as Parameters<typeof permissionUpdatePost>[0]) as Promise<{
    data: ApiResponse<unknown>
  }>
}

export function deletePermission(id: number) {
  return permissionDeletePost({ id } as Parameters<typeof permissionDeletePost>[0])
}
