import request from '../request'

const joinBasePath = (basePath, path) => {
  const normalizedBase = (basePath || '').replace(/\/$/, '')
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${normalizedBase}${normalizedPath}`
}

export const getTerminalWsBaseUrl = () => {
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || ''

  if (apiBaseUrl.startsWith('https://')) {
    return `wss://${apiBaseUrl.slice('https://'.length)}`
  }

  if (apiBaseUrl.startsWith('http://')) {
    return `ws://${apiBaseUrl.slice('http://'.length)}`
  }

  return apiBaseUrl
}

export const getTerminalSessionList = (params) => request.get('/cmdb/terminal/list', { params })
export const getTerminalSessionDetail = (params) => request.get('/cmdb/terminal/detail', { params })
export const getTerminalRecording = (params) => request.get('/cmdb/terminal/recording', { params })
export const getTerminalConnectWsUrl = (hostId) => joinBasePath(getTerminalWsBaseUrl(), `/cmdb/terminal/connect?hostId=${encodeURIComponent(hostId)}`)
export const addSessionTag = (data) => request.post('/cmdb/terminal/tag/add', data)
export const removeSessionTag = (data) => request.post('/cmdb/terminal/tag/remove', data)
export const getSessionTags = (params) => request.get('/cmdb/terminal/tag/list', { params })
export const getAvailableTags = () => request.get('/cmdb/terminal/tags')
export const searchSessionsByTag = (params) => request.get('/cmdb/terminal/tag/search', { params })
