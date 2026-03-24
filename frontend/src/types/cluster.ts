// 集群相关类型
export interface Cluster {
  id: number
  name: string
  authType: string
  env: string
  isDefault: boolean
  k8sVersion: string
  labels: string
  nodeCount: number
  remark: string
  status: string
  url: string
  createdAt: string
  updatedAt: string
}

// 创建/编辑集群表单
export interface ClusterForm {
  authType: string
  caData: string
  env: string
  kubeconfig: string
  labels: string
  name: string
  remark: string
  token: string
  url: string
}

// 工作负载统计
export interface WorkloadCounts {
  cronjob: number
  daemonset: number
  deployment: number
  job: number
  statefulset: number
}

// 网络统计
export interface NetworkCounts {
  ingress: number
  service: number
}

// 存储统计
export interface StorageCounts {
  pv: number
  pvc: number
}

// 事件信息
export interface EventInfo {
  message: string
  object: string
  reason: string
  time: string
  type: string
}

// 事件列表响应
export interface EventListResponse {
  items: EventInfo[]
  total: number
}
