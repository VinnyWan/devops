export interface ApiClusterCreateResponse {
  data?: { authType?: string; id?: number; name?: string; status?: string; url?: string }
  message?: string
}

export interface ApiClusterDeleteResponse {
  message?: string
}

export interface ApiClusterHealthResponse {
  data?: { error?: string; healthy?: boolean; status?: string }
  message?: string
}

export interface ApiClusterListResponse {
  data?: Array<ApiClusterResponse>
  message?: string
  page?: number
  pageSize?: number
  total?: number
}

export interface ApiClusterResponse {
  authType?: string
  createdAt?: string
  env?: string
  id?: number
  isDefault?: boolean
  k8sVersion?: string
  labels?: string
  name?: string
  nodeCount?: number
  remark?: string
  status?: string
  updatedAt?: string
  url?: string
}

export interface ApiCordonNodeRequest {
  clusterId: number
  cordon?: boolean
  name: string
}

export interface ApiDrainNodeRequest {
  clusterId: number
  deleteLocalData?: boolean
  force?: boolean
  gracePeriodSeconds?: number
  ignoreDaemonSets?: boolean
  name: string
}

export type ApiK8sObject = Record<string, unknown>

export interface ApiResponse {
  code?: number
  data?: unknown
  message?: string
}

export interface ApiUpdateLabelsRequest {
  clusterId: number
  labels: Record<string, string>
  name: string
}

export interface ApiUpdateTaintsRequest {
  clusterId: number
  name: string
  taints: Array<unknown>
}

export interface ModelPermission {
  action?: string
  createdAt?: string
  description?: string
  id?: number
  name?: string
  resource?: string
  updatedAt?: string
}

export interface ServiceChangePasswordRequest {
  newPassword: string
  oldPassword: string
}

export interface ServiceCreateDepartmentRequest {
  name: string
  parentId?: number
}

export interface ServiceCreateDeptUserRequest {
  departmentId?: number
  email: string
  name?: string
  password: string
  status?: string
  username: string
}

export interface ServiceCreatePermissionRequest {
  action: string
  description?: string
  name: string
  resource: string
}

export interface ServiceCreateRequest {
  authType?: string
  caData?: string
  env?: string
  kubeconfig?: string
  labels?: string
  name?: string
  remark?: string
  token?: string
  url?: string
}

export interface ServiceCreateRoleRequest {
  description?: string
  displayName?: string
  name: string
}

export interface ServiceLoginRequest {
  authType?: string
  password: string
  username: string
}

export interface ServiceNodeDetail {
  addresses?: Array<unknown>
  age?: string
  conditions?: Array<unknown>
  cpuCapacity?: string
  cpuUsage?: string
  createdAt?: string
  creationTimestamp?: string
  externalIP?: string
  images?: Array<unknown>
  ip?: string
  k8sVersion?: string
  kernelVersion?: string
  kubeletVersion?: string
  labels?: Record<string, string>
  memoryCapacity?: string
  memoryUsage?: string
  name?: string
  osImage?: string
  podCapacity?: number
  podCount?: number
  role?: string
  status?: string
  systemInfo?: unknown
  taints?: Array<unknown>
  unschedulable?: boolean
}

export interface ServiceNodeListItem {
  age?: string
  cpuCapacity?: string
  cpuUsage?: string
  createdAt?: string
  creationTimestamp?: string
  externalIP?: string
  ip?: string
  k8sVersion?: string
  kernelVersion?: string
  kubeletVersion?: string
  labels?: Record<string, string>
  memoryCapacity?: string
  memoryUsage?: string
  name?: string
  osImage?: string
  podCapacity?: number
  podCount?: number
  role?: string
  status?: string
  taints?: Array<unknown>
  unschedulable?: boolean
}

export interface ServiceNodeListResponse {
  items?: Array<ServiceNodeListItem>
  total?: number
}

export interface ServiceRegisterRequest {
  email: string
  name?: string
  password: string
  username: string
}

export interface ServiceTransferUserDepartmentRequest {
  toDepartmentId: number
  userId: number
}

export interface ServiceUpdateDepartmentRequest {
  id: number
  name?: string
  parentId?: number
}

export interface ServiceUpdateDeptUserRequest {
  email?: string
  id: number
  name?: string
  status?: string
}

export interface ServiceUpdateRequest {
  caData?: string
  env?: string
  id?: number
  kubeconfig?: string
  labels?: string
  name?: string
  remark?: string
  token?: string
  url?: string
}

export interface ServiceUpdateRoleRequest {
  description?: string
  displayName?: string
  id: number
  name?: string
  permissionIds?: Array<number>
}

export interface ServiceUpdateUserRequest {
  departmentId?: number
  email?: string
  id: number
  name?: string
  status?: string
  username?: string
}
