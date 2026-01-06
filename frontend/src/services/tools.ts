import { get, post } from './http'

// 诉讼费计算
export interface LitigationFeeParams {
  amount: number
}

export interface LitigationFeeResult {
  amount: number
  fee: number
  calculationTime: number
}

export function calculateLitigationFee(params: LitigationFeeParams) {
  return post('/tools/litigation-fee', params)
}

// 利息计算器
export interface InterestCalculatorParams {
  principal: number
  rate: number
  days: number
  type: 'simple' | 'compound' | 'penalty'
}

export interface InterestCalculatorResult {
  principal: number
  rate: number
  days: number
  type: string
  interest: number
  total: number
}

export function calculateInterest(params: InterestCalculatorParams) {
  return post('/tools/interest-calculator', params)
}

// 工期计算器
export interface DeadlineCalculatorParams {
  startDate: string
  days: number
  excludeWeekends: boolean
  excludeHolidays: boolean
}

export interface DeadlineCalculatorResult {
  startDate: string
  days: number
  excludeWeekends: boolean
  excludeHolidays: boolean
  endDate: string
  workDays: number
}

export function calculateDeadline(params: DeadlineCalculatorParams) {
  return post('/tools/deadline-calculator', params)
}

// =============================================================================
// 法条查询相关类型定义
// =============================================================================

export interface LawItem {
  id: number
  statuteNumber: string
  title: string
  content: string
  lawName: string
  chapter?: string
  section?: string
  part?: string
  effectiveDate?: string
  expiryDate?: string
  publishingAuthority?: string
  status: string
  hierarchyLevel: number
  parentStatuteId?: number
  orderInHierarchy?: number
  tags: string[]
  keywords: string[]
  createdAt: string
  updatedAt: string
  category?: LawCategory
  isFavorited?: boolean
  viewCount?: number
  favoriteCount?: number
  fullPath?: string
  isActive?: boolean
}

export interface LawCategory {
  id: number
  name: string
  code: string
  parentId?: number
  level: number
  description?: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface LawTag {
  id: number
  name: string
  color: string
  description?: string
  usageCount: number
  createdAt: string
}

export interface LegalSearchRequest {
  query?: string
  categoryId?: number
  lawName?: string
  status?: string
  effectiveFrom?: string
  effectiveTo?: string
  tags?: string[]
  includeInactive?: boolean
  sortBy?: 'relevance' | 'date' | 'title'
  sortOrder?: 'asc' | 'desc'
  page?: number
  pageSize?: number
}

export interface LegalSearchResponse {
  total: number
  page: number
  pageSize: number
  totalPages: number
  statutes: LawItem[]
  categories?: CategoryStat[]
  suggestions?: string[]
  searchTime: number
}

export interface CategoryStat {
  id: number
  name: string
  code: string
  level: number
  description: string
  statuteCount: number
}

export interface SearchHistory {
  id: number
  userId?: number
  searchQuery: string
  searchFilters: Record<string, any>
  resultCount: number
  searchDuration?: number
  createdAt: string
}

export interface UserFavorite {
  id: number
  statuteId: number
  userId: number
  createdAt: string
  statute?: LawItem
}

// =============================================================================
// 法条查询 API 函数 - 修正为后端实际路由
// =============================================================================

/**
 * 搜索法条 - 使用后端正确的路由
 * GET /api/v1/legal/statutes/search
 */
export function searchLaws(params: LegalSearchRequest) {
  return get<{
    data: LegalSearchResponse
  }>('/legal/statutes/search', { params })
}

/**
 * 获取法条详情
 * GET /api/v1/legal/statutes/:id
 */
export function getLawById(id: number) {
  return get<{
    data: LawItem
  }>(`/legal/statutes/${id}`)
}

/**
 * 根据编号获取法条
 * 注意：后端可能没有此端点，使用搜索代替
 */
export function getLawByNumber(number: string) {
  return searchLaws({ query: number, pageSize: 1 })
}

/**
 * 获取法条列表 - 使用搜索 API
 */
export function getLaws(params?: LegalSearchRequest) {
  return get<{
    data: LegalSearchResponse
  }>('/legal/statutes/search', { params })
}

/**
 * 获取法条分类
 * GET /api/v1/legal/categories
 */
export function getLawCategories() {
  return get<{
    data: LawCategory[]
  }>('/legal/categories')
}

/**
 * 获取分类树
 * GET /api/v1/legal/categories/tree
 */
export function getLawCategoryTree() {
  return get<{
    data: LawCategory[]
  }>('/legal/categories/tree')
}

/**
 * 获取法条标签
 * GET /api/v1/legal/tags
 */
export function getLawTags() {
  return get<{
    data: LawTag[]
  }>('/legal/tags')
}

/**
 * 获取热门标签
 * 注意：后端可能未实现此端点，返回 getLawTags 的结果
 */
export function getPopularLawTags(limit?: number) {
  return get<{
    data: LawTag[]
  }>('/legal/tags', { params: { limit } })
}

/**
 * 获取搜索建议
 * GET /api/v1/legal/search/suggestions
 */
export function getLawSuggestions(query: string) {
  return get<{
    data: string[]
  }>('/legal/search/suggestions', { params: { query } })
}

/**
 * 获取相关法条
 * 注意：后端可能未实现此端点
 */
export function getRelatedLaws(id: number, limit?: number) {
  return get<{
    data: LawItem[]
  }>(`/legal/statutes/${id}/related`, { params: { limit } })
}

/**
 * 获取热门法条
 * 注意：后端可能未实现此端点，使用搜索代替
 */
export function getPopularLaws(limit?: number) {
  return get<{
    data: LawItem[]
  }>('/legal/statutes/search', {
    params: { page: 1, pageSize: limit || 10, sortBy: 'relevance', sortOrder: 'desc' },
  })
}

/**
 * 获取最新法条
 * 注意：后端可能未实现此端点，使用搜索代替
 */
export function getRecentLaws(days?: number) {
  return get<{
    data: LawItem[]
  }>('/legal/statutes/search', {
    params: { page: 1, pageSize: 20, sortBy: 'date', sortOrder: 'desc' },
  })
}

/**
 * 获取分类统计
 * 注意：后端可能未实现此端点
 */
export function getCategoryStats() {
  return get<{
    data: CategoryStat[]
  }>('/legal/stats/categories')
}

/**
 * 获取热门搜索
 * 注意：后端可能未实现此端点，返回空数组
 */
export function getPopularSearches(limit?: number) {
  return get<{
    data: string[]
  }>('/legal/popular-searches', { params: { limit } })
}

// =============================================================================
// 用户收藏相关
// =============================================================================

export function addToFavorites(statuteId: number) {
  return post<{
    data: UserFavorite
  }>('/legal/favorites', { statuteId })
}

export function removeFromFavorites(statuteId: number) {
  return fetch(`/api/v1/legal/favorites/${statuteId}`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
  })
}

export function getUserFavorites(params?: { page?: number; pageSize?: number }) {
  return get<{
    data: UserFavorite[]
  }>('/legal/favorites', params)
}

export function getSearchHistory(limit?: number) {
  return get<{
    data: SearchHistory[]
  }>('/legal/search-history', { params: { limit } })
}

// =============================================================================
// 批量导入
// =============================================================================

export interface LegalStatuteImportItem {
  statute_number: string
  title: string
  content: string
  category_code: string
  law_name: string
  chapter?: string
  section?: string
  part?: string
  effective_date?: string
  expiry_date?: string
  publishing_authority?: string
  status?: string
  tags?: string[]
  keywords?: string[]
}

export interface LegalStatuteImportRequest {
  statutes: LegalStatuteImportItem[]
}

export interface LegalStatuteImportResponse {
  total_count: number
  success_count: number
  failure_count: number
  success_numbers: string[]
  failure_numbers: string[]
  errors: Array<{
    statute_number: string
    message: string
  }>
  processing_time_ms: number
}

export function bulkImportStatutes(data: LegalStatuteImportRequest) {
  return post<{
    data: LegalStatuteImportResponse
  }>('/legal/statutes/import', data)
}

// =============================================================================
// 管理员功能
// =============================================================================

export function syncToElasticsearch() {
  return post('/legal/admin/sync-elasticsearch')
}

export function rebuildSearchIndex() {
  return post('/legal/admin/rebuild-index')
}

// =============================================================================
// 合同模板
// =============================================================================

export interface ContractTemplate {
  id: number
  name: string
  category: string
  description: string
  downloadUrl: string
  updateTime: string
}

export function getContractTemplates() {
  return get('/tools/contract-templates')
}

// =============================================================================
// 文档转换
// =============================================================================

export interface DocumentConvertParams {
  sourceFormat: string
  targetFormat: string
}

export interface DocumentConvertResult {
  sourceFormat: string
  targetFormat: string
  status: string
  message: string
  downloadUrl: string
}

export function convertDocument(params: DocumentConvertParams) {
  return post('/tools/document-convert', params)
}

// =============================================================================
// 翻译服务
// =============================================================================

export interface TranslateParams {
  text: string
  targetLang: string
}

export interface TranslateResult {
  originalText: string
  targetLang: string
  translatedText: string
  sourceLang: string
}

export function translate(params: TranslateParams) {
  return post('/tools/translate', params)
}
