export type LogSearchPostParams = {
  keyword?: string
  source?: string
  level?: string
  start?: string
  end?: string
  page?: number
  pageSize?: number
}

export type LogSearchPostRequest = Record<string, never>

export type LogSearchPostResponse = Record<string, unknown>
