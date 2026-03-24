// 通用 API 响应类型
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// 分页响应（page/pageSize/total 与 data 同级）
export interface PageResponse<T> {
  code: number
  message: string
  data: T[]
  page: number
  pageSize: number
  total: number
}

export interface PageParams {
  page?: number
  pageSize?: number
}
