import type { AxiosResponse } from 'axios'
import type { ApiResponse } from '@/types/api'

export async function unwrapResponseData<T>(
  request: Promise<AxiosResponse<ApiResponse<unknown>>>,
): Promise<T> {
  const response = await request
  return response.data.data as T
}
