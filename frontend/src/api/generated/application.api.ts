import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  AppDeployPostRequest,
  AppDeployPostResponse,
  AppDeploymentListPostParams,
  AppDeploymentListPostResponse,
  AppListPostResponse,
  AppRollbackPostRequest,
  AppRollbackPostResponse,
  AppTemplateListPostParams,
  AppTemplateListPostResponse,
  AppTemplateSavePostRequest,
  AppTemplateSavePostResponse,
  AppTopologyPostParams,
  AppTopologyPostResponse,
  AppVersionListPostParams,
  AppVersionListPostResponse,
} from '@/types/generated/application.types'

export function appDeployPost(data: AppDeployPostRequest) {
  return http.post<ApiResponse<AppDeployPostResponse>>('/v1/app/deploy', data)
}

export function appDeploymentListPost(params: AppDeploymentListPostParams) {
  return http.post<ApiResponse<AppDeploymentListPostResponse>>(
    '/v1/app/deployment/list',
    undefined,
    { params },
  )
}

export function appListPost() {
  return http.post<ApiResponse<AppListPostResponse>>('/v1/app/list')
}

export function appRollbackPost(data: AppRollbackPostRequest) {
  return http.post<ApiResponse<AppRollbackPostResponse>>('/v1/app/rollback', data)
}

export function appTemplateListPost(params: AppTemplateListPostParams) {
  return http.post<ApiResponse<AppTemplateListPostResponse>>('/v1/app/template/list', undefined, {
    params,
  })
}

export function appTemplateSavePost(data: AppTemplateSavePostRequest) {
  return http.post<ApiResponse<AppTemplateSavePostResponse>>('/v1/app/template/save', data)
}

export function appTopologyPost(params: AppTopologyPostParams) {
  return http.post<ApiResponse<AppTopologyPostResponse>>('/v1/app/topology', undefined, { params })
}

export function appVersionListPost(params: AppVersionListPostParams) {
  return http.post<ApiResponse<AppVersionListPostResponse>>('/v1/app/version/list', undefined, {
    params,
  })
}
