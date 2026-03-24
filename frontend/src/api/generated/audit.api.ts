import http from '@/api/index'

import type { ApiResponse } from '@/types/api'

import type {
  AuditCleanupPostResponse,
  AuditExportGetParams,
  AuditExportGetResponse,
  AuditListGetParams,
  AuditListGetResponse,
} from '@/types/generated/audit.types'

export function auditCleanupPost() {
  return http.post<ApiResponse<AuditCleanupPostResponse>>('/v1/audit/cleanup')
}

export function auditExportGet(params: AuditExportGetParams) {
  return http.get<ApiResponse<AuditExportGetResponse>>('/v1/audit/export', { params })
}

export function auditListGet(params: AuditListGetParams) {
  return http.get<ApiResponse<AuditListGetResponse>>('/v1/audit/list', { params })
}
