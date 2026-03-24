import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  AlertChannelUpsertPostRequest,
  AlertChannelUpsertPostResponse,
  AlertChannelsPostParams,
  AlertChannelsPostResponse,
  AlertConfigPostResponse,
  AlertConfigUpsertPostRequest,
  AlertConfigUpsertPostResponse,
  AlertHistoryPostParams,
  AlertHistoryPostResponse,
  AlertRuleTogglePostRequest,
  AlertRuleTogglePostResponse,
  AlertRulesPostParams,
  AlertRulesPostResponse,
  AlertSilenceUpsertPostRequest,
  AlertSilenceUpsertPostResponse,
  AlertSilencesPostParams,
  AlertSilencesPostResponse,
} from '@/types/generated/alert.types'

export function alertChannelUpsertPost(data: AlertChannelUpsertPostRequest) {
  return http.post<ApiResponse<AlertChannelUpsertPostResponse>>('/v1/alert/channel/upsert', data)
}

export function alertChannelsPost(params: AlertChannelsPostParams) {
  return http.post<ApiResponse<AlertChannelsPostResponse>>('/v1/alert/channels', undefined, {
    params,
  })
}

export function alertConfigPost() {
  return http.post<ApiResponse<AlertConfigPostResponse>>('/v1/alert/config')
}

export function alertConfigUpsertPost(data: AlertConfigUpsertPostRequest) {
  return http.post<ApiResponse<AlertConfigUpsertPostResponse>>('/v1/alert/config/upsert', data)
}

export function alertHistoryPost(params: AlertHistoryPostParams) {
  return http.post<ApiResponse<AlertHistoryPostResponse>>('/v1/alert/history', undefined, {
    params,
  })
}

export function alertRuleTogglePost(data: AlertRuleTogglePostRequest) {
  return http.post<ApiResponse<AlertRuleTogglePostResponse>>('/v1/alert/rule/toggle', data)
}

export function alertRulesPost(params: AlertRulesPostParams) {
  return http.post<ApiResponse<AlertRulesPostResponse>>('/v1/alert/rules', undefined, { params })
}

export function alertSilenceUpsertPost(data: AlertSilenceUpsertPostRequest) {
  return http.post<ApiResponse<AlertSilenceUpsertPostResponse>>('/v1/alert/silence/upsert', data)
}

export function alertSilencesPost(params: AlertSilencesPostParams) {
  return http.post<ApiResponse<AlertSilencesPostResponse>>('/v1/alert/silences', undefined, {
    params,
  })
}
