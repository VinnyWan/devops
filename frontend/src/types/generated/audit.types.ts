export type AuditCleanupPostParams = Record<string, never>

export type AuditCleanupPostRequest = Record<string, never>

export type AuditCleanupPostResponse = Record<string, unknown>

export type AuditExportGetParams = {
  userId?: number
  username?: string
  operation?: string
  resource?: string
  keyword?: string
  startAt?: string
  endAt?: string
  limit?: number
}

export type AuditExportGetRequest = Record<string, never>

export type AuditExportGetResponse = string

export type AuditListGetParams = {
  userId?: number
  username?: string
  operation?: string
  resource?: string
  keyword?: string
  startAt?: string
  endAt?: string
  page?: number
  pageSize?: number
}

export type AuditListGetRequest = Record<string, never>

export type AuditListGetResponse = Record<string, unknown>
