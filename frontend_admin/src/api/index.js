import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'https://cj-api.wsky.fun/api',
  timeout: 10000,
})

// Export baseURL for use in components
export const baseURL = api.defaults.baseURL

export function getAdminToken() {
  return localStorage.getItem('admin_token') || ''
}

export function setAdminToken(token) {
  if (!token) {
    localStorage.removeItem('admin_token')
    return
  }
  localStorage.setItem('admin_token', token)
}

// Alias for compatibility
export const setToken = setAdminToken

api.interceptors.request.use((config) => {
  const token = getAdminToken()
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export async function adminSetup({ username, password, setupKey }) {
  const res = await api.post('/admin/setup', { username, password, setupKey })
  return res.data
}

export async function adminLogin({ username, password }) {
  const res = await api.post('/admin/login', { username, password })
  return res.data
}

export async function adminLogout() {
  const res = await api.post('/admin/logout')
  return res.data
}

export async function adminUserList() {
  const res = await api.get('/admin/users')
  return res.data
}

export async function adminUserCreate({ username, password }) {
  const res = await api.post('/admin/users', { username, password })
  return res.data
}

export async function adminUserSetPassword({ id, password }) {
  const res = await api.patch(`/admin/users/${id}/password`, { password })
  return res.data
}

export async function adminUserDelete(id) {
  const res = await api.delete(`/admin/users/${id}`)
  return res.data
}

// User Aliases
export const getUsers = adminUserList
export const createUser = adminUserCreate
export const updateUserPassword = adminUserSetPassword
export const deleteUser = adminUserDelete

export async function adminSiteCategoryList() {
  const res = await api.get('/admin/site-categories')
  return res.data
}

export async function adminSiteCategoryCreate({ name, sort }) {
  const res = await api.post('/admin/site-categories', { name, sort })
  return res.data
}

export async function adminSiteCategoryUpdate({ id, name, sort }) {
  const res = await api.patch(`/admin/site-categories/${id}`, { name, sort })
  return res.data
}

export async function adminSiteCategoryDelete(id) {
  const res = await api.delete(`/admin/site-categories/${id}`)
  return res.data
}

// Site Category Aliases
export const getSiteCategories = adminSiteCategoryList
export const createSiteCategory = adminSiteCategoryCreate
export const updateSiteCategory = adminSiteCategoryUpdate
export const deleteSiteCategory = adminSiteCategoryDelete

export async function adminSiteList({ categoryId } = {}) {
  const qs = categoryId ? `?categoryId=${encodeURIComponent(categoryId)}` : ''
  const res = await api.get(`/admin/sites${qs}`)
  return res.data
}

export async function adminSiteCreate(payload) {
  const res = await api.post('/admin/sites', payload)
  return res.data
}

export async function adminSiteUpdate({ id, ...payload }) {
  const res = await api.patch(`/admin/sites/${id}`, payload)
  return res.data
}

export async function adminSiteDelete(id) {
  const res = await api.delete(`/admin/sites/${id}`)
  return res.data
}

// Site Aliases
export const getSites = adminSiteList
export const createSite = adminSiteCreate
export const updateSite = adminSiteUpdate
export const deleteSite = adminSiteDelete

export async function adminBatchList({ type, createdAtStart, createdAtEnd } = {}) {
  const params = new URLSearchParams()
  if (type) params.set('type', type)
  if (createdAtStart) params.set('createdAtStart', createdAtStart)
  if (createdAtEnd) params.set('createdAtEnd', createdAtEnd)
  const qs = params.toString() ? `?${params.toString()}` : ''
  const res = await api.get(`/admin/batches${qs}`)
  return res.data
}

export async function adminBatchNewsList(batchId) {
  const res = await api.get(`/admin/batches/${batchId}/news`)
  return res.data
}

export async function adminBatchDelete(batchId) {
  const res = await api.delete(`/admin/batches/${batchId}`)
  return res.data
}

// Batch Aliases
export const getBatches = adminBatchList
export const getBatchNews = adminBatchNewsList
export const deleteBatch = adminBatchDelete

export async function adminNewsList({ batchId, keyword, createdAtStart, createdAtEnd } = {}) {
  const params = new URLSearchParams()
  if (batchId) params.set('batchId', String(batchId))
  if (keyword) params.set('keyword', keyword)
  if (createdAtStart) params.set('createdAtStart', createdAtStart)
  if (createdAtEnd) params.set('createdAtEnd', createdAtEnd)
  const qs = params.toString() ? `?${params.toString()}` : ''
  const res = await api.get(`/admin/news${qs}`)
  return res.data
}

export async function adminNewsCreate(payload) {
  const res = await api.post('/admin/news', payload)
  return res.data
}

export async function adminNewsUpdate({ id, ...payload }) {
  const res = await api.patch(`/admin/news/${id}`, payload)
  return res.data
}

export async function adminNewsDelete(id) {
  const res = await api.delete(`/admin/news/${id}`)
  return res.data
}

// News Aliases
export const getNews = adminNewsList
export const createNews = adminNewsCreate
export const updateNews = adminNewsUpdate
export const deleteNews = adminNewsDelete

export async function adminAnalysisList({ batchId, type, createdAtStart, createdAtEnd } = {}) {
  const params = new URLSearchParams()
  if (batchId) params.set('batchId', String(batchId))
  if (type) params.set('type', type)
  if (createdAtStart) params.set('createdAtStart', createdAtStart)
  if (createdAtEnd) params.set('createdAtEnd', createdAtEnd)
  const qs = params.toString() ? `?${params.toString()}` : ''
  const res = await api.get(`/admin/analysis${qs}`)
  return res.data
}

export async function adminAnalysisCreate(payload) {
  const res = await api.post('/admin/analysis', payload)
  return res.data
}

export async function adminAnalysisUpdate({ id, ...payload }) {
  const res = await api.patch(`/admin/analysis/${id}`, payload)
  return res.data
}

export async function adminAnalysisDelete(id) {
  const res = await api.delete(`/admin/analysis/${id}`)
  return res.data
}

// Analysis Aliases
export const getAnalysis = adminAnalysisList
export const createAnalysis = adminAnalysisCreate
export const updateAnalysis = adminAnalysisUpdate
export const deleteAnalysis = adminAnalysisDelete
