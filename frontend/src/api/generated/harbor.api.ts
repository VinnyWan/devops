import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  HarborHarborConfigPostResponse,
  HarborHarborConfigUpsertPostRequest,
  HarborHarborConfigUpsertPostResponse,
  HarborHarborImagesPostParams,
  HarborHarborImagesPostResponse,
  HarborHarborListPostParams,
  HarborHarborListPostResponse,
} from '@/types/generated/harbor.types'

export function harborHarborConfigPost() {
  return http.post<ApiResponse<HarborHarborConfigPostResponse>>('/v1/harbor/config')
}

export function harborHarborConfigUpsertPost(data: HarborHarborConfigUpsertPostRequest) {
  return http.post<ApiResponse<HarborHarborConfigUpsertPostResponse>>(
    '/v1/harbor/config/upsert',
    data,
  )
}

export function harborHarborImagesPost(params: HarborHarborImagesPostParams) {
  return http.post<ApiResponse<HarborHarborImagesPostResponse>>('/v1/harbor/images', undefined, {
    params,
  })
}

export function harborHarborListPost(params: HarborHarborListPostParams) {
  return http.post<ApiResponse<HarborHarborListPostResponse>>('/v1/harbor/list', undefined, {
    params,
  })
}
