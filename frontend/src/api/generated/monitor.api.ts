import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  MonitorConfigPostResponse,
  MonitorConfigUpsertPostRequest,
  MonitorConfigUpsertPostResponse,
  MonitorQueryPostParams,
  MonitorQueryPostResponse,
} from '@/types/generated/monitor.types'

export function monitorConfigPost() {
  return http.post<ApiResponse<MonitorConfigPostResponse>>('/v1/monitor/config')
}

export function monitorConfigUpsertPost(data: MonitorConfigUpsertPostRequest) {
  return http.post<ApiResponse<MonitorConfigUpsertPostResponse>>('/v1/monitor/config/upsert', data)
}

export function monitorQueryPost(params: MonitorQueryPostParams) {
  return http.post<ApiResponse<MonitorQueryPostResponse>>('/v1/monitor/query', undefined, {
    params,
  })
}
