import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  K8sClusterCreatePostRequest,
  K8sClusterCreatePostResponse,
  K8sClusterDefaultPostResponse,
  K8sClusterDeletePostRequest,
  K8sClusterDeletePostResponse,
  K8sClusterDetailPostParams,
  K8sClusterDetailPostResponse,
  K8sClusterEventsPostParams,
  K8sClusterEventsPostResponse,
  K8sClusterHealthPostRequest,
  K8sClusterHealthPostResponse,
  K8sClusterListPostParams,
  K8sClusterListPostResponse,
  K8sClusterNodesPostParams,
  K8sClusterNodesPostResponse,
  K8sClusterSearchPostParams,
  K8sClusterSearchPostResponse,
  K8sClusterSetDefaultPostRequest,
  K8sClusterSetDefaultPostResponse,
  K8sClusterStatsNetworkPostParams,
  K8sClusterStatsNetworkPostResponse,
  K8sClusterStatsStoragePostParams,
  K8sClusterStatsStoragePostResponse,
  K8sClusterStatsWorkloadPostParams,
  K8sClusterStatsWorkloadPostResponse,
  K8sClusterUpdatePostRequest,
  K8sClusterUpdatePostResponse,
} from '@/types/generated/cluster.types'

export function k8sClusterCreatePost(data: K8sClusterCreatePostRequest) {
  return http.post<ApiResponse<K8sClusterCreatePostResponse>>('/v1/k8s/cluster/create', data)
}

export function k8sClusterDefaultPost() {
  return http.get<ApiResponse<K8sClusterDefaultPostResponse>>('/v1/k8s/cluster/default')
}

export function k8sClusterDeletePost(data: K8sClusterDeletePostRequest) {
  return http.post<ApiResponse<K8sClusterDeletePostResponse>>('/v1/k8s/cluster/delete', data)
}

export function k8sClusterDetailPost(params: K8sClusterDetailPostParams) {
  return http.get<ApiResponse<K8sClusterDetailPostResponse>>('/v1/k8s/cluster/detail', {
    params,
  })
}

export function k8sClusterEventsPost(params: K8sClusterEventsPostParams) {
  return http.get<ApiResponse<K8sClusterEventsPostResponse>>('/v1/k8s/cluster/events', {
    params,
  })
}

export function k8sClusterHealthPost(data: K8sClusterHealthPostRequest) {
  return http.get<ApiResponse<K8sClusterHealthPostResponse>>('/v1/k8s/cluster/health', {
    params: data,
  })
}

export function k8sClusterListPost(params: K8sClusterListPostParams) {
  return http.get<ApiResponse<K8sClusterListPostResponse>>('/v1/k8s/cluster/list', {
    params,
  })
}

export function k8sClusterNodesPost(params: K8sClusterNodesPostParams) {
  return http.get<ApiResponse<K8sClusterNodesPostResponse>>('/v1/k8s/cluster/nodes', {
    params,
  })
}

export function k8sClusterSearchPost(params: K8sClusterSearchPostParams) {
  return http.get<ApiResponse<K8sClusterSearchPostResponse>>('/v1/k8s/cluster/search', {
    params,
  })
}

export function k8sClusterSetDefaultPost(data: K8sClusterSetDefaultPostRequest) {
  return http.post<ApiResponse<K8sClusterSetDefaultPostResponse>>(
    '/v1/k8s/cluster/set-default',
    data,
  )
}

export function k8sClusterStatsNetworkPost(params: K8sClusterStatsNetworkPostParams) {
  return http.get<ApiResponse<K8sClusterStatsNetworkPostResponse>>(
    '/v1/k8s/cluster/stats/network',
    { params },
  )
}

export function k8sClusterStatsStoragePost(params: K8sClusterStatsStoragePostParams) {
  return http.get<ApiResponse<K8sClusterStatsStoragePostResponse>>(
    '/v1/k8s/cluster/stats/storage',
    { params },
  )
}

export function k8sClusterStatsWorkloadPost(params: K8sClusterStatsWorkloadPostParams) {
  return http.get<ApiResponse<K8sClusterStatsWorkloadPostResponse>>(
    '/v1/k8s/cluster/stats/workload',
    { params },
  )
}

export function k8sClusterUpdatePost(data: K8sClusterUpdatePostRequest) {
  return http.post<ApiResponse<K8sClusterUpdatePostResponse>>('/v1/k8s/cluster/update', data)
}
