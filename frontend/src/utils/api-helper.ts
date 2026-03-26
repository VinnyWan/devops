export function parseId(id: string | number): number {
  const parsed = typeof id === 'string' ? parseInt(id, 10) : id
  if (isNaN(parsed)) {
    throw new Error('无效的ID')
  }
  return parsed
}

export function parsePagination(page?: number, pageSize?: number) {
  return {
    page: page && page > 0 ? page : 1,
    pageSize: pageSize && pageSize > 0 ? pageSize : 10
  }
}
