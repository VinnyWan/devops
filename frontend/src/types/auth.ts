// 认证相关类型
export type AuthType = 'local' | 'ldap' | 'oauth2'

export interface LoginParams {
  username: string
  password: string
  authType: AuthType
}

export interface LoginResult {
  token: string
  user: UserInfo
}

export interface Department {
  id: number
  name: string
  parentId: number | null
  createdAt: string
  updatedAt: string
}

export interface UserInfo {
  ID: number
  username: string
  name: string
  email: string
  externalId: string
  authType: AuthType
  status: string
  isAdmin: boolean
  isLocked: boolean
  departmentId: number
  department: Department
  roles: RoleInfo[]
  lastLoginAt: string
  createdAt: string
  updatedAt: string
}

export interface RoleInfo {
  id: number
  name: string
  code: string
}

export interface ChangePasswordParams {
  oldPassword: string
  newPassword: string
}
