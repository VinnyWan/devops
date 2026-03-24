import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  K8sK8sConfigmapCreatePostParams,
  K8sK8sConfigmapCreatePostRequest,
  K8sK8sConfigmapCreatePostResponse,
  K8sK8sConfigmapDeletePostRequest,
  K8sK8sConfigmapDeletePostResponse,
  K8sK8sConfigmapDetailPostParams,
  K8sK8sConfigmapDetailPostResponse,
  K8sK8sConfigmapListPostParams,
  K8sK8sConfigmapListPostResponse,
  K8sK8sConfigmapUpdatePostParams,
  K8sK8sConfigmapUpdatePostRequest,
  K8sK8sConfigmapUpdatePostResponse,
  K8sK8sDeploymentCreatePostParams,
  K8sK8sDeploymentCreatePostRequest,
  K8sK8sDeploymentCreatePostResponse,
  K8sK8sDeploymentDeletePostRequest,
  K8sK8sDeploymentDeletePostResponse,
  K8sK8sDeploymentDetailPostParams,
  K8sK8sDeploymentDetailPostResponse,
  K8sK8sDeploymentListPostParams,
  K8sK8sDeploymentListPostResponse,
  K8sK8sDeploymentPodsPostParams,
  K8sK8sDeploymentPodsPostResponse,
  K8sK8sDeploymentRestartPostRequest,
  K8sK8sDeploymentRestartPostResponse,
  K8sK8sDeploymentScalePostRequest,
  K8sK8sDeploymentScalePostResponse,
  K8sK8sDeploymentUpdatePostParams,
  K8sK8sDeploymentUpdatePostRequest,
  K8sK8sDeploymentUpdatePostResponse,
  K8sK8sDeploymentYamlPostParams,
  K8sK8sDeploymentYamlPostResponse,
  K8sK8sDeploymentYamlUpdatePostParams,
  K8sK8sDeploymentYamlUpdatePostRequest,
  K8sK8sDeploymentYamlUpdatePostResponse,
  K8sK8sIngressCreatePostParams,
  K8sK8sIngressCreatePostRequest,
  K8sK8sIngressCreatePostResponse,
  K8sK8sIngressDeletePostRequest,
  K8sK8sIngressDeletePostResponse,
  K8sK8sIngressDetailPostParams,
  K8sK8sIngressDetailPostResponse,
  K8sK8sIngressListPostParams,
  K8sK8sIngressListPostResponse,
  K8sK8sIngressUpdatePostParams,
  K8sK8sIngressUpdatePostRequest,
  K8sK8sIngressUpdatePostResponse,
  K8sK8sNamespacesListPostParams,
  K8sK8sNamespacesListPostResponse,
  K8sK8sPodCreatePostParams,
  K8sK8sPodCreatePostRequest,
  K8sK8sPodCreatePostResponse,
  K8sK8sPodDeletePostRequest,
  K8sK8sPodDeletePostResponse,
  K8sK8sPodDescribePostParams,
  K8sK8sPodDescribePostResponse,
  K8sK8sPodDetailPostParams,
  K8sK8sPodDetailPostResponse,
  K8sK8sPodListByOwnerPostParams,
  K8sK8sPodListByOwnerPostResponse,
  K8sK8sPodListPostParams,
  K8sK8sPodListPostResponse,
  K8sK8sPodTerminalGetParams,
  K8sK8sPodTerminalGetResponse,
  K8sK8sPodLogsGetParams,
  K8sK8sPodLogsGetResponse,
  K8sK8sPodYamlGetParams,
  K8sK8sPodYamlGetResponse,
  K8sK8sPodYamlUpdatePostParams,
  K8sK8sPodYamlUpdatePostRequest,
  K8sK8sPodYamlUpdatePostResponse,
  K8sK8sPodUpdatePostParams,
  K8sK8sPodUpdatePostRequest,
  K8sK8sPodUpdatePostResponse,
  K8sK8sServiceCreatePostParams,
  K8sK8sServiceCreatePostRequest,
  K8sK8sServiceCreatePostResponse,
  K8sK8sServiceDeletePostRequest,
  K8sK8sServiceDeletePostResponse,
  K8sK8sServiceDetailPostParams,
  K8sK8sServiceDetailPostResponse,
  K8sK8sServiceListPostParams,
  K8sK8sServiceListPostResponse,
  K8sK8sServiceUpdatePostParams,
  K8sK8sServiceUpdatePostRequest,
  K8sK8sServiceUpdatePostResponse,
} from '@/types/generated/k8s-resource.types'

export function k8sK8sConfigmapCreatePost(
  params: K8sK8sConfigmapCreatePostParams,
  data: K8sK8sConfigmapCreatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sConfigmapCreatePostResponse>>(
    '/v1/k8s/configmap/create',
    data,
    { params },
  )
}

export function k8sK8sConfigmapDeletePost(data: K8sK8sConfigmapDeletePostRequest) {
  return http.post<ApiResponse<K8sK8sConfigmapDeletePostResponse>>('/v1/k8s/configmap/delete', data)
}

export function k8sK8sConfigmapDetailPost(params: K8sK8sConfigmapDetailPostParams) {
  return http.get<ApiResponse<K8sK8sConfigmapDetailPostResponse>>(
    '/v1/k8s/configmap/detail',
    { params },
  )
}

export function k8sK8sConfigmapListPost(params: K8sK8sConfigmapListPostParams) {
  return http.get<ApiResponse<K8sK8sConfigmapListPostResponse>>(
    '/v1/k8s/configmap/list',
    { params },
  )
}

export function k8sK8sConfigmapUpdatePost(
  params: K8sK8sConfigmapUpdatePostParams,
  data: K8sK8sConfigmapUpdatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sConfigmapUpdatePostResponse>>(
    '/v1/k8s/configmap/update',
    data,
    { params },
  )
}

export function k8sK8sDeploymentCreatePost(
  params: K8sK8sDeploymentCreatePostParams,
  data: K8sK8sDeploymentCreatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sDeploymentCreatePostResponse>>(
    '/v1/k8s/deployment/create',
    data,
    { params },
  )
}

export function k8sK8sDeploymentDeletePost(data: K8sK8sDeploymentDeletePostRequest) {
  return http.post<ApiResponse<K8sK8sDeploymentDeletePostResponse>>(
    '/v1/k8s/deployment/delete',
    data,
  )
}

export function k8sK8sDeploymentDetailPost(params: K8sK8sDeploymentDetailPostParams) {
  return http.get<ApiResponse<K8sK8sDeploymentDetailPostResponse>>(
    '/v1/k8s/deployment/detail',
    { params },
  )
}

export function k8sK8sDeploymentListPost(params: K8sK8sDeploymentListPostParams) {
  return http.get<ApiResponse<K8sK8sDeploymentListPostResponse>>(
    '/v1/k8s/deployment/list',
    { params },
  )
}

export function k8sK8sDeploymentPodsPost(params: K8sK8sDeploymentPodsPostParams) {
  return http.get<ApiResponse<K8sK8sDeploymentPodsPostResponse>>(
    '/v1/k8s/deployment/pods',
    { params },
  )
}

export function k8sK8sDeploymentRestartPost(data: K8sK8sDeploymentRestartPostRequest) {
  return http.post<ApiResponse<K8sK8sDeploymentRestartPostResponse>>(
    '/v1/k8s/deployment/restart',
    data,
  )
}

export function k8sK8sDeploymentScalePost(data: K8sK8sDeploymentScalePostRequest) {
  return http.post<ApiResponse<K8sK8sDeploymentScalePostResponse>>('/v1/k8s/deployment/scale', data)
}

export function k8sK8sDeploymentUpdatePost(
  params: K8sK8sDeploymentUpdatePostParams,
  data: K8sK8sDeploymentUpdatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sDeploymentUpdatePostResponse>>(
    '/v1/k8s/deployment/update',
    data,
    { params },
  )
}

export function k8sK8sDeploymentYamlPost(params: K8sK8sDeploymentYamlPostParams) {
  return http.get<ApiResponse<K8sK8sDeploymentYamlPostResponse>>(
    '/v1/k8s/deployment/yaml',
    { params },
  )
}

export function k8sK8sDeploymentYamlUpdatePost(
  params: K8sK8sDeploymentYamlUpdatePostParams,
  data: K8sK8sDeploymentYamlUpdatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sDeploymentYamlUpdatePostResponse>>(
    '/v1/k8s/deployment/yaml/update',
    data,
    { params },
  )
}

export function k8sK8sIngressCreatePost(
  params: K8sK8sIngressCreatePostParams,
  data: K8sK8sIngressCreatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sIngressCreatePostResponse>>('/v1/k8s/ingress/create', data, {
    params,
  })
}

export function k8sK8sIngressDeletePost(data: K8sK8sIngressDeletePostRequest) {
  return http.post<ApiResponse<K8sK8sIngressDeletePostResponse>>('/v1/k8s/ingress/delete', data)
}

export function k8sK8sIngressDetailPost(params: K8sK8sIngressDetailPostParams) {
  return http.get<ApiResponse<K8sK8sIngressDetailPostResponse>>(
    '/v1/k8s/ingress/detail',
    { params },
  )
}

export function k8sK8sIngressListPost(params: K8sK8sIngressListPostParams) {
  return http.get<ApiResponse<K8sK8sIngressListPostResponse>>('/v1/k8s/ingress/list', {
    params,
  })
}

export function k8sK8sIngressUpdatePost(
  params: K8sK8sIngressUpdatePostParams,
  data: K8sK8sIngressUpdatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sIngressUpdatePostResponse>>('/v1/k8s/ingress/update', data, {
    params,
  })
}

export function k8sK8sNamespacesListPost(params: K8sK8sNamespacesListPostParams) {
  return http.get<ApiResponse<K8sK8sNamespacesListPostResponse>>(
    '/v1/k8s/namespaces/list',
    { params },
  )
}

export function k8sK8sPodCreatePost(
  params: K8sK8sPodCreatePostParams,
  data: K8sK8sPodCreatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sPodCreatePostResponse>>('/v1/k8s/pod/create', data, { params })
}

export function k8sK8sPodDeletePost(data: K8sK8sPodDeletePostRequest) {
  return http.post<ApiResponse<K8sK8sPodDeletePostResponse>>('/v1/k8s/pod/delete', data)
}

export function k8sK8sPodDescribePost(params: K8sK8sPodDescribePostParams) {
  return http.get<ApiResponse<K8sK8sPodDescribePostResponse>>('/v1/k8s/pod/describe', {
    params,
  })
}

export function k8sK8sPodDetailPost(params: K8sK8sPodDetailPostParams) {
  return http.get<ApiResponse<K8sK8sPodDetailPostResponse>>('/v1/k8s/pod/detail', {
    params,
  })
}

export function k8sK8sPodListPost(params: K8sK8sPodListPostParams) {
  return http.get<ApiResponse<K8sK8sPodListPostResponse>>('/v1/k8s/pod/list', {
    params,
  })
}

export function k8sK8sPodListByOwnerPost(params: K8sK8sPodListByOwnerPostParams) {
  return http.get<ApiResponse<K8sK8sPodListByOwnerPostResponse>>(
    '/v1/k8s/pod/list_by_owner',
    { params },
  )
}

export function k8sK8sPodTerminalGet(params: K8sK8sPodTerminalGetParams) {
  return http.get<ApiResponse<K8sK8sPodTerminalGetResponse>>('/v1/k8s/pod/terminal', { params })
}

export function k8sK8sPodLogsGet(params: K8sK8sPodLogsGetParams) {
  return http.get<ApiResponse<K8sK8sPodLogsGetResponse>>('/v1/k8s/pod/logs', { params })
}

export function k8sK8sPodYamlGet(params: K8sK8sPodYamlGetParams) {
  return http.get<ApiResponse<K8sK8sPodYamlGetResponse>>('/v1/k8s/pod/yaml', { params })
}

export function k8sK8sPodYamlUpdatePost(
  params: K8sK8sPodYamlUpdatePostParams,
  data: K8sK8sPodYamlUpdatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sPodYamlUpdatePostResponse>>('/v1/k8s/pod/yaml/update', data, { params })
}

export function k8sK8sPodUpdatePost(
  params: K8sK8sPodUpdatePostParams,
  data: K8sK8sPodUpdatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sPodUpdatePostResponse>>('/v1/k8s/pod/update', data, { params })
}

export function k8sK8sServiceCreatePost(
  params: K8sK8sServiceCreatePostParams,
  data: K8sK8sServiceCreatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sServiceCreatePostResponse>>('/v1/k8s/service/create', data, {
    params,
  })
}

export function k8sK8sServiceDeletePost(data: K8sK8sServiceDeletePostRequest) {
  return http.post<ApiResponse<K8sK8sServiceDeletePostResponse>>('/v1/k8s/service/delete', data)
}

export function k8sK8sServiceDetailPost(params: K8sK8sServiceDetailPostParams) {
  return http.get<ApiResponse<K8sK8sServiceDetailPostResponse>>(
    '/v1/k8s/service/detail',
    { params },
  )
}

export function k8sK8sServiceListPost(params: K8sK8sServiceListPostParams) {
  return http.get<ApiResponse<K8sK8sServiceListPostResponse>>('/v1/k8s/service/list', {
    params,
  })
}

export function k8sK8sServiceUpdatePost(
  params: K8sK8sServiceUpdatePostParams,
  data: K8sK8sServiceUpdatePostRequest,
) {
  return http.post<ApiResponse<K8sK8sServiceUpdatePostResponse>>('/v1/k8s/service/update', data, {
    params,
  })
}
