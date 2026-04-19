import request from '../request'

export function browseFiles(params) {
  return request.get('/cmdb/file/browse', { params })
}

export function uploadFile(hostId, path, file, onProgress) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('path', path)
  return request.post(`/cmdb/file/upload/${hostId}`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: onProgress
  })
}

export function deleteFile(data) {
  return request.post('/cmdb/file/delete', data)
}

export function renameFile(data) {
  return request.post('/cmdb/file/rename', data)
}

export function mkdir(data) {
  return request.post('/cmdb/file/mkdir', data)
}

export function previewFile(params) {
  return request.get('/cmdb/file/preview', { params })
}

export function editFile(data) {
  return request.post('/cmdb/file/edit', data)
}

export function distributeFile(file, path, hostIds, onProgress) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('path', path)
  formData.append('hostIds', hostIds.join(','))
  return request.post('/cmdb/file/distribute', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: onProgress
  })
}

export function getDownloadUrl(hostId, filePath) {
  const token = sessionStorage.getItem('token')
  const baseURL = import.meta.env.VITE_API_BASE_URL || ''
  return `${baseURL}/cmdb/file/download?hostId=${hostId}&path=${encodeURIComponent(filePath)}&token=${token}`
}
