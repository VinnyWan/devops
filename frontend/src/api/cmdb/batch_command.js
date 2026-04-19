import { getTerminalWsBaseUrl } from './terminal'

const joinBasePath = (basePath, path) => {
  const normalizedBase = (basePath || '').replace(/\/$/, '')
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${normalizedBase}${normalizedPath}`
}

export const getBatchCommandWsUrl = () => joinBasePath(getTerminalWsBaseUrl(), '/cmdb/terminal/batch')
