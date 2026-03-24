import type { PageParams } from '@/types/api'
import type { User } from '@/types/system'
import {
  userAssignRolesPost,
  userChangePasswordPost,
  userDeletePost,
  userDetailPost,
  userListPost,
  userLockPost,
  userResetPasswordPost,
  userUnlockPost,
  userUpdatePost,
} from '@/api/generated/user.api'
import { unwrapResponseData } from '@/api/service'

export function getUserList(params: PageParams & { departmentId?: number; keyword?: string }) {
  return unwrapResponseData<User[]>(userListPost(params as Parameters<typeof userListPost>[0]))
}

export function getUserDetail(id: number) {
  return unwrapResponseData<User>(userDetailPost({ id } as Parameters<typeof userDetailPost>[0]))
}

export function updateUser(data: Partial<User>) {
  return unwrapResponseData<unknown>(userUpdatePost(data as Parameters<typeof userUpdatePost>[0]))
}

export function deleteUser(id: number) {
  return unwrapResponseData<unknown>(userDeletePost({ id } as Parameters<typeof userDeletePost>[0]))
}

export function changePassword(data: any) {
  return unwrapResponseData<unknown>(
    userChangePasswordPost(data as Parameters<typeof userChangePasswordPost>[0]),
  )
}

export function resetPassword(id: number) {
  return unwrapResponseData<unknown>(
    userResetPasswordPost({ id } as Parameters<typeof userResetPasswordPost>[0]),
  )
}

export function assignRoles(userId: number, roleIds: number[]) {
  return unwrapResponseData<unknown>(
    userAssignRolesPost({ userId, roleIds } as Parameters<typeof userAssignRolesPost>[0]),
  )
}

export function lockUser(id: number) {
  return unwrapResponseData<unknown>(userLockPost({ id } as Parameters<typeof userLockPost>[0]))
}

export function unlockUser(id: number) {
  return unwrapResponseData<unknown>(userUnlockPost({ id } as Parameters<typeof userUnlockPost>[0]))
}
