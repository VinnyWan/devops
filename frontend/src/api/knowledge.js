import request from './request'

export const listCategories = () => request.get('/kb/categories')
export const createCategory = (data) => request.post('/kb/categories', data)
export const updateCategory = (id, data) => request.put(`/kb/categories/${id}`, data)
export const deleteCategory = (id) => request.delete(`/kb/categories/${id}`)

export const listArticles = (params) => request.get('/kb/articles', { params })
export const getArticle = (id) => request.get(`/kb/articles/${id}`)
export const createArticle = (data) => request.post('/kb/articles', data)
export const updateArticle = (id, data) => request.put(`/kb/articles/${id}`, data)
export const deleteArticle = (id) => request.delete(`/kb/articles/${id}`)
