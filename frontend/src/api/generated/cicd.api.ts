import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  CICDCicdConfigPostResponse,
  CICDCicdConfigUpsertPostRequest,
  CICDCicdConfigUpsertPostResponse,
  CICDCicdListPostParams,
  CICDCicdListPostResponse,
  CICDCicdLogsPostParams,
  CICDCicdLogsPostResponse,
  CICDCicdOrchestrationPreviewPostParams,
  CICDCicdOrchestrationPreviewPostResponse,
  CICDCicdRunsPostParams,
  CICDCicdRunsPostResponse,
  CICDCicdStatusPostParams,
  CICDCicdStatusPostResponse,
  CICDCicdTemplateSavePostRequest,
  CICDCicdTemplateSavePostResponse,
  CICDCicdTemplatesPostParams,
  CICDCicdTemplatesPostResponse,
  CICDCicdTriggerPostRequest,
  CICDCicdTriggerPostResponse,
} from '@/types/generated/cicd.types'

export function cICDCicdConfigPost() {
  return http.post<ApiResponse<CICDCicdConfigPostResponse>>('/v1/cicd/config')
}

export function cICDCicdConfigUpsertPost(data: CICDCicdConfigUpsertPostRequest) {
  return http.post<ApiResponse<CICDCicdConfigUpsertPostResponse>>('/v1/cicd/config/upsert', data)
}

export function cICDCicdListPost(params: CICDCicdListPostParams) {
  return http.post<ApiResponse<CICDCicdListPostResponse>>('/v1/cicd/list', undefined, { params })
}

export function cICDCicdLogsPost(params: CICDCicdLogsPostParams) {
  return http.post<ApiResponse<CICDCicdLogsPostResponse>>('/v1/cicd/logs', undefined, { params })
}

export function cICDCicdOrchestrationPreviewPost(params: CICDCicdOrchestrationPreviewPostParams) {
  return http.post<ApiResponse<CICDCicdOrchestrationPreviewPostResponse>>(
    '/v1/cicd/orchestration/preview',
    undefined,
    { params },
  )
}

export function cICDCicdRunsPost(params: CICDCicdRunsPostParams) {
  return http.post<ApiResponse<CICDCicdRunsPostResponse>>('/v1/cicd/runs', undefined, { params })
}

export function cICDCicdStatusPost(params: CICDCicdStatusPostParams) {
  return http.post<ApiResponse<CICDCicdStatusPostResponse>>('/v1/cicd/status', undefined, {
    params,
  })
}

export function cICDCicdTemplateSavePost(data: CICDCicdTemplateSavePostRequest) {
  return http.post<ApiResponse<CICDCicdTemplateSavePostResponse>>('/v1/cicd/template/save', data)
}

export function cICDCicdTemplatesPost(params: CICDCicdTemplatesPostParams) {
  return http.post<ApiResponse<CICDCicdTemplatesPostResponse>>('/v1/cicd/templates', undefined, {
    params,
  })
}

export function cICDCicdTriggerPost(data: CICDCicdTriggerPostRequest) {
  return http.post<ApiResponse<CICDCicdTriggerPostResponse>>('/v1/cicd/trigger', data)
}
