// 系统管理相关类型
export interface User {
  id: number
  username: string
  nickname: string
  email: string
  phone: string
  departmentId: number
  departmentName: string
  status: number
  roles: string[]
  createdAt: string
}

export interface Department {
  id: number
  name: string
  parentId: number
  sort: number
  leader: string
  children?: Department[]
}

export interface Role {
  id: number
  name: string
  code: string
  description: string
  status: number
  permissions: number[]
}

export interface Permission {
  id: number
  name: string
  code: string
  type: string
  parentId: number
  path: string
  icon: string
  sort: number
  children?: Permission[]
}

export interface AuditLog {
  id: number
  userId: number
  username: string
  operation: string
  method: string
  path: string
  params: string
  ip: string
  status: number
  latency: number
  retentionDays: number
  createdAt: string
}
