// 节点相关类型
export interface NodeInfo {
  name: string
  status: string
  roles: string
  kubeletVersion: string
  internalIP: string
  externalIP: string
  os: string
  arch: string
  containerRuntime: string
  age: string
  cpu: ResourceUsage
  memory: ResourceUsage
  pods: PodCount
  schedulable: boolean
  taints: Taint[]
}

export interface ResourceUsage {
  capacity: string
  allocatable: string
  used: string
  usagePercent: number
}

export interface PodCount {
  current: number
  capacity: number
}

export interface Taint {
  key: string
  value: string
  effect: string
}

export interface NodeLabel {
  key: string
  value: string
}
