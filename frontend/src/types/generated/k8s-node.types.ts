import type {
  ApiCordonNodeRequest,
  ApiDrainNodeRequest,
  ApiResponse,
  ApiUpdateLabelsRequest,
  ApiUpdateTaintsRequest,
  ServiceNodeDetail,
  ServiceNodeListResponse,
} from './common.types'

export type K8sK8sNodeCordonPostParams = Record<string, never>

export type K8sK8sNodeCordonPostRequest = ApiCordonNodeRequest

export type K8sK8sNodeCordonPostResponse = ApiResponse

export type K8sK8sNodeDetailPostParams = { clusterId: number; name: string }

export type K8sK8sNodeDetailPostRequest = Record<string, never>

export type K8sK8sNodeDetailPostResponse = ServiceNodeDetail

export type K8sK8sNodeDrainPostParams = Record<string, never>

export type K8sK8sNodeDrainPostRequest = ApiDrainNodeRequest

export type K8sK8sNodeDrainPostResponse = ApiResponse

export type K8sK8sNodeEventsPostParams = { clusterId: number; name: string }

export type K8sK8sNodeEventsPostRequest = Record<string, never>

export type K8sK8sNodeEventsPostResponse = ApiResponse

export type K8sK8sNodeLabelsPostParams = Record<string, never>

export type K8sK8sNodeLabelsPostRequest = ApiUpdateLabelsRequest

export type K8sK8sNodeLabelsPostResponse = ApiResponse

export type K8sK8sNodeTaintsPostParams = Record<string, never>

export type K8sK8sNodeTaintsPostRequest = ApiUpdateTaintsRequest

export type K8sK8sNodeTaintsPostResponse = ApiResponse

export type K8sK8sNodesPostParams = {
  clusterId: number
  page?: number
  pageSize?: number
  name?: string
  status?: string
  role?: string
}

export type K8sK8sNodesPostRequest = Record<string, never>

export type K8sK8sNodesPostResponse = ServiceNodeListResponse
