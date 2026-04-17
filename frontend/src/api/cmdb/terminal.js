import request from '../request'

const joinBasePath = (basePath, path) => {
  const normalizedBase = (basePath || '').replace(/\/$/, '')
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${normalizedBase}${normalizedPath}`
}

export const getTerminalSessionList = (params) => request.get('/cmdb/terminal/list', { params })
export const getTerminalSessionDetail = (params) => request.get('/cmdb/terminal/detail', { params })
export const getTerminalRecording = (params) => request.get('/cmdb/terminal/recording', { params })
export const getTerminalConnectWsUrl = (hostId) => joinBasePath(import.meta.env.VITE_API_BASE_URL, `/cmdb/terminal/connect?hostId=${hostId}`)
