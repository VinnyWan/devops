import type { ApiK8sObject, ApiResponse } from './common.types'

export type K8sK8sConfigmapCreatePostParams = { clusterId?: number; namespace: string }

export type K8sK8sConfigmapCreatePostRequest = ApiK8sObject

export type K8sK8sConfigmapCreatePostResponse = ApiResponse

export type K8sK8sConfigmapDeletePostParams = Record<string, never>

export type K8sK8sConfigmapDeletePostRequest = Record<string, unknown>

export type K8sK8sConfigmapDeletePostResponse = ApiResponse

export type K8sK8sConfigmapDetailPostParams = {
  clusterId?: number
  namespace: string
  name: string
}

export type K8sK8sConfigmapDetailPostRequest = Record<string, never>

export type K8sK8sConfigmapDetailPostResponse = ApiResponse

export type K8sK8sConfigmapListPostParams = {
  clusterId?: number
  namespace?: string
  page?: number
  pageSize?: number
  keyword?: string
}

export type K8sK8sConfigmapListPostRequest = Record<string, never>

export type K8sK8sConfigmapListPostResponse = ApiResponse

export type K8sK8sConfigmapUpdatePostParams = { clusterId?: number; namespace: string }

export type K8sK8sConfigmapUpdatePostRequest = ApiK8sObject

export type K8sK8sConfigmapUpdatePostResponse = ApiResponse

export type K8sK8sDeploymentCreatePostParams = { clusterId?: number; namespace: string }

export type K8sK8sDeploymentCreatePostRequest = ApiK8sObject

export type K8sK8sDeploymentCreatePostResponse = ApiResponse

export type K8sK8sDeploymentDeletePostParams = Record<string, never>

export type K8sK8sDeploymentDeletePostRequest = Record<string, unknown>

export type K8sK8sDeploymentDeletePostResponse = ApiResponse

export type K8sK8sDeploymentDetailPostParams = {
  clusterId?: number
  namespace: string
  name: string
}

export type K8sK8sDeploymentDetailPostRequest = Record<string, never>

export type K8sK8sDeploymentDetailPostResponse = ApiResponse

export type K8sK8sDeploymentListPostParams = {
  clusterId?: number
  namespace?: string
  page?: number
  pageSize?: number
  keyword?: string
}

export type K8sK8sDeploymentListPostRequest = Record<string, never>

export type K8sK8sDeploymentListPostResponse = ApiResponse

export type K8sK8sDeploymentPodsPostParams = { clusterId?: number; namespace: string; name: string }

export type K8sK8sDeploymentPodsPostRequest = Record<string, never>

export type K8sK8sDeploymentPodsPostResponse = ApiResponse

export type K8sK8sDeploymentRestartPostParams = Record<string, never>

export type K8sK8sDeploymentRestartPostRequest = Record<string, unknown>

export type K8sK8sDeploymentRestartPostResponse = ApiResponse

export type K8sK8sDeploymentScalePostParams = Record<string, never>

export type K8sK8sDeploymentScalePostRequest = Record<string, unknown>

export type K8sK8sDeploymentScalePostResponse = ApiResponse

export type K8sK8sDeploymentUpdatePostParams = { clusterId?: number; namespace: string }

export type K8sK8sDeploymentUpdatePostRequest = ApiK8sObject

export type K8sK8sDeploymentUpdatePostResponse = ApiResponse

export type K8sK8sDeploymentYamlPostParams = { clusterId?: number; namespace: string; name: string }

export type K8sK8sDeploymentYamlPostRequest = Record<string, never>

export type K8sK8sDeploymentYamlPostResponse = ApiResponse

export type K8sK8sDeploymentYamlUpdatePostParams = {
  clusterId?: number
  namespace: string
  name: string
}

export type K8sK8sDeploymentYamlUpdatePostRequest = Record<string, unknown>

export type K8sK8sDeploymentYamlUpdatePostResponse = ApiResponse

export type K8sK8sIngressCreatePostParams = { clusterId?: number; namespace: string }

export type K8sK8sIngressCreatePostRequest = ApiK8sObject

export type K8sK8sIngressCreatePostResponse = ApiResponse

export type K8sK8sIngressDeletePostParams = Record<string, never>

export type K8sK8sIngressDeletePostRequest = Record<string, unknown>

export type K8sK8sIngressDeletePostResponse = ApiResponse

export type K8sK8sIngressDetailPostParams = { clusterId?: number; namespace: string; name: string }

export type K8sK8sIngressDetailPostRequest = Record<string, never>

export type K8sK8sIngressDetailPostResponse = ApiResponse

export type K8sK8sIngressListPostParams = {
  clusterId?: number
  namespace?: string
  page?: number
  pageSize?: number
  keyword?: string
}

export type K8sK8sIngressListPostRequest = Record<string, never>

export type K8sK8sIngressListPostResponse = ApiResponse

export type K8sK8sIngressUpdatePostParams = { clusterId?: number; namespace: string }

export type K8sK8sIngressUpdatePostRequest = ApiK8sObject

export type K8sK8sIngressUpdatePostResponse = ApiResponse

export type K8sK8sNamespacesListPostParams = { clusterId?: number }

export type K8sK8sNamespacesListPostRequest = Record<string, never>

export type K8sK8sNamespacesListPostResponse = ApiResponse

export type K8sK8sPodCreatePostParams = { clusterId?: number; namespace: string }

export type K8sK8sPodCreatePostRequest = ApiK8sObject

export type K8sK8sPodCreatePostResponse = ApiResponse

export type K8sK8sPodDeletePostParams = Record<string, never>

export type K8sK8sPodDeletePostRequest = Record<string, unknown>

export type K8sK8sPodDeletePostResponse = ApiResponse

export type K8sK8sPodDescribePostParams = {
  clusterId?: number
  namespace: string
  ownerType: string
  ownerName: string
  name: string
}

export type K8sK8sPodDescribePostRequest = Record<string, never>

export type K8sK8sPodDescribePostResponse = ApiResponse

export type K8sK8sPodDetailPostParams = { clusterId?: number; namespace: string; name: string }

export type K8sK8sPodDetailPostRequest = Record<string, never>

export type K8sK8sPodDetailPostResponse = ApiResponse

export type K8sK8sPodListPostParams = {
  clusterId?: number
  namespace?: string
  page?: number
  pageSize?: number
  keyword?: string
}

export type K8sK8sPodListPostRequest = Record<string, never>

export type K8sK8sPodListPostResponse = ApiResponse

export type K8sK8sPodListByOwnerPostParams = {
  clusterId?: number
  namespace: string
  ownerType: string
  ownerName: string
  name: string
}

export type K8sK8sPodListByOwnerPostRequest = Record<string, never>

export type K8sK8sPodListByOwnerPostResponse = ApiResponse

export type K8sK8sPodTerminalGetParams = {
  clusterId?: number
  namespace: string
  pod: string
  container?: string
  shell?: string
  cols?: number
  rows?: number
}

export type K8sK8sPodTerminalGetRequest = Record<string, never>

export type K8sK8sPodTerminalGetResponse = unknown

export type K8sK8sPodUpdatePostParams = { clusterId?: number; namespace: string }

export type K8sK8sPodUpdatePostRequest = ApiK8sObject

export type K8sK8sPodUpdatePostResponse = ApiResponse

export type K8sK8sServiceCreatePostParams = { clusterId?: number; namespace: string }

export type K8sK8sServiceCreatePostRequest = ApiK8sObject

export type K8sK8sServiceCreatePostResponse = ApiResponse

export type K8sK8sServiceDeletePostParams = Record<string, never>

export type K8sK8sServiceDeletePostRequest = Record<string, unknown>

export type K8sK8sServiceDeletePostResponse = ApiResponse

export type K8sK8sServiceDetailPostParams = { clusterId?: number; namespace: string; name: string }

export type K8sK8sServiceDetailPostRequest = Record<string, never>

export type K8sK8sServiceDetailPostResponse = ApiResponse

export type K8sK8sServiceListPostParams = {
  clusterId?: number
  namespace?: string
  page?: number
  pageSize?: number
  keyword?: string
}

export type K8sK8sServiceListPostRequest = Record<string, never>

export type K8sK8sServiceListPostResponse = ApiResponse

export type K8sK8sServiceUpdatePostParams = { clusterId?: number; namespace: string }

export type K8sK8sServiceUpdatePostRequest = ApiK8sObject

export type K8sK8sServiceUpdatePostResponse = ApiResponse
