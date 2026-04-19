import request from '../request'

export const getSnippetList = (params) => request.get('/cmdb/snippet/list', { params })
export const searchSnippets = (keyword) => request.get('/cmdb/snippet/list', { params: { keyword } })
export const createSnippet = (data) => request.post('/cmdb/snippet/create', data)
export const updateSnippet = (data) => request.post('/cmdb/snippet/update', data)
export const deleteSnippet = (params) => request.post('/cmdb/snippet/delete', null, { params })
