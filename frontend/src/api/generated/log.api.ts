import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type { LogSearchPostParams, LogSearchPostResponse } from '@/types/generated/log.types'

export function logSearchPost(params: LogSearchPostParams) {
  return http.post<ApiResponse<LogSearchPostResponse>>('/v1/log/search', undefined, { params })
}
