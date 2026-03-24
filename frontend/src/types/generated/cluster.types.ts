import type {
  ApiClusterCreateResponse,
  ApiClusterDeleteResponse,
  ApiClusterHealthResponse,
  ApiClusterListResponse,
  ApiClusterResponse,
  ApiResponse,
  ServiceCreateRequest,
  ServiceUpdateRequest,
} from './common.types'

export type K8sClusterCreatePostParams = Record<string, never>

export type K8sClusterCreatePostRequest = ServiceCreateRequest

export type K8sClusterCreatePostResponse = ApiClusterCreateResponse

export type K8sClusterDefaultPostParams = Record<string, never>

export type K8sClusterDefaultPostRequest = Record<string, never>

export type K8sClusterDefaultPostResponse = ApiClusterResponse

export type K8sClusterDeletePostParams = Record<string, never>

export type K8sClusterDeletePostRequest = Record<string, unknown>

export type K8sClusterDeletePostResponse = ApiClusterDeleteResponse

export type K8sClusterDetailPostParams = { id: number }

export type K8sClusterDetailPostRequest = Record<string, never>

export type K8sClusterDetailPostResponse = ApiClusterResponse

export type K8sClusterEventsPostParams = { id: number; page?: number; pageSize?: number }

export type K8sClusterEventsPostRequest = Record<string, never>

export type K8sClusterEventsPostResponse = ApiResponse

export type K8sClusterHealthPostParams = Record<string, never>

export type K8sClusterHealthPostRequest = Record<string, unknown>

export type K8sClusterHealthPostResponse = ApiClusterHealthResponse

export type K8sClusterListPostParams = {
  page?: number
  pageSize?: number
  env?: string
  keyword?: string
  name?: string
}

export type K8sClusterListPostRequest = Record<string, never>

export type K8sClusterListPostResponse = ApiClusterListResponse

export type K8sClusterNodesPostParams = {
  id: number
  page?: number
  pageSize?: number
  name?: string
}

export type K8sClusterNodesPostRequest = Record<string, never>

export type K8sClusterNodesPostResponse = ApiResponse

export type K8sClusterSearchPostParams = {
  name?: string
  env?: string
  page?: number
  pageSize?: number
}

export type K8sClusterSearchPostRequest = Record<string, never>

export type K8sClusterSearchPostResponse = ApiClusterListResponse

export type K8sClusterSetDefaultPostParams = Record<string, never>

export type K8sClusterSetDefaultPostRequest = Record<string, unknown>

export type K8sClusterSetDefaultPostResponse = Record<string, unknown>

export type K8sClusterStatsNetworkPostParams = { id: number }

export type K8sClusterStatsNetworkPostRequest = Record<string, never>

export type K8sClusterStatsNetworkPostResponse = ApiResponse

export type K8sClusterStatsStoragePostParams = { id: number }

export type K8sClusterStatsStoragePostRequest = Record<string, never>

export type K8sClusterStatsStoragePostResponse = ApiResponse

export type K8sClusterStatsWorkloadPostParams = { id: number }

export type K8sClusterStatsWorkloadPostRequest = Record<string, never>

export type K8sClusterStatsWorkloadPostResponse = ApiResponse

export type K8sClusterUpdatePostParams = Record<string, never>

export type K8sClusterUpdatePostRequest = ServiceUpdateRequest

export type K8sClusterUpdatePostResponse = ApiClusterCreateResponse
