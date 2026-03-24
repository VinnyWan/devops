import { userPermissionsPost } from '@/api/generated/user.api'
import { unwrapResponseData } from '@/api/service'

export function getUserPermissions() {
  return unwrapResponseData<string[]>(userPermissionsPost())
}
