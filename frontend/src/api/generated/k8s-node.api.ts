import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  K8sK8sNodeCordonPostRequest,
  K8sK8sNodeCordonPostResponse,
  K8sK8sNodeDetailPostParams,
  K8sK8sNodeDetailPostResponse,
  K8sK8sNodeDrainPostRequest,
  K8sK8sNodeDrainPostResponse,
  K8sK8sNodeEventsPostParams,
  K8sK8sNodeEventsPostResponse,
  K8sK8sNodeLabelsPostRequest,
  K8sK8sNodeLabelsPostResponse,
  K8sK8sNodeTaintsPostRequest,
  K8sK8sNodeTaintsPostResponse,
  K8sK8sNodesPostParams,
  K8sK8sNodesPostResponse,
} from '@/types/generated/k8s-node.types'

export function k8sK8sNodeCordonPost(data: K8sK8sNodeCordonPostRequest) {
  return http.post<ApiResponse<K8sK8sNodeCordonPostResponse>>('/v1/k8s/node/cordon', data)
}

export function k8sK8sNodeDetailPost(params: K8sK8sNodeDetailPostParams) {
  return http.get<ApiResponse<K8sK8sNodeDetailPostResponse>>('/v1/k8s/node/detail', {
    params,
  })
}

export function k8sK8sNodeDrainPost(data: K8sK8sNodeDrainPostRequest) {
  return http.post<ApiResponse<K8sK8sNodeDrainPostResponse>>('/v1/k8s/node/drain', data)
}

export function k8sK8sNodeEventsPost(params: K8sK8sNodeEventsPostParams) {
  return http.get<ApiResponse<K8sK8sNodeEventsPostResponse>>('/v1/k8s/node/events', {
    params,
  })
}

export function k8sK8sNodeLabelsPost(data: K8sK8sNodeLabelsPostRequest) {
  return http.post<ApiResponse<K8sK8sNodeLabelsPostResponse>>('/v1/k8s/node/labels', data)
}

export function k8sK8sNodeTaintsPost(data: K8sK8sNodeTaintsPostRequest) {
  return http.post<ApiResponse<K8sK8sNodeTaintsPostResponse>>('/v1/k8s/node/taints', data)
}

export function k8sK8sNodesPost(params: K8sK8sNodesPostParams) {
  return http.get<ApiResponse<K8sK8sNodesPostResponse>>('/v1/k8s/nodes', { params })
}
