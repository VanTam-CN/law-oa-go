import { del, get, post } from './http'

// 诉讼费计算
export interface LitigationFeeParams {
  amount: number
}

export interface LitigationFeeResult {
  amount: number
  fee: number
  calculationTime: number
}

export function calculateLitigationFee(params: LitigationFeeParams): Promise<LitigationFeeResult> {
  return post<LitigationFeeResult>('/tools/litigation-fee', params)
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

export function calculateInterest(
  params: InterestCalculatorParams,
): Promise<InterestCalculatorResult> {
  return post<InterestCalculatorResult>('/tools/interest-calculator', params)
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

export function calculateDeadline(
  params: DeadlineCalculatorParams,
): Promise<DeadlineCalculatorResult> {
  return post<DeadlineCalculatorResult>('/tools/deadline-calculator', params)
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

type ApiData<T> = {
  data: T
}

const wrapData = <T>(data: T): ApiData<T> => ({ data })

const normalizeCategory = (category: any): LawCategory => ({
  id: Number(category?.id || 0),
  name: category?.name || '',
  code: category?.code || '',
  parentId: category?.parentId ?? category?.parent_id,
  level: Number(category?.level || 0),
  description: category?.description || '',
  isActive: category?.isActive ?? category?.is_active ?? true,
  createdAt: category?.createdAt || category?.created_at || '',
  updatedAt: category?.updatedAt || category?.updated_at || '',
})

const normalizeTag = (tag: any): LawTag => ({
  id: Number(tag?.id || 0),
  name: tag?.name || '',
  color: tag?.color || 'blue',
  description: tag?.description || '',
  usageCount: Number(tag?.usageCount ?? tag?.usage_count ?? 0),
  createdAt: tag?.createdAt || tag?.created_at || '',
})

const normalizeLawItem = (item: any): LawItem => ({
  id: Number(item?.id || 0),
  statuteNumber: item?.statuteNumber || item?.statute_number || '',
  title: item?.title || '',
  content: item?.content || '',
  lawName: item?.lawName || item?.law_name || '',
  chapter: item?.chapter || undefined,
  section: item?.section || undefined,
  part: item?.part || undefined,
  effectiveDate: item?.effectiveDate || item?.effective_date || undefined,
  expiryDate: item?.expiryDate || item?.expiry_date || undefined,
  publishingAuthority: item?.publishingAuthority || item?.publishing_authority || undefined,
  status: item?.status || 'active',
  hierarchyLevel: Number(item?.hierarchyLevel ?? item?.hierarchy_level ?? 1),
  parentStatuteId: item?.parentStatuteId ?? item?.parent_statute_id,
  orderInHierarchy: item?.orderInHierarchy ?? item?.order_in_hierarchy,
  tags: Array.isArray(item?.tags) ? item.tags : [],
  keywords: Array.isArray(item?.keywords) ? item.keywords : [],
  createdAt: item?.createdAt || item?.created_at || '',
  updatedAt: item?.updatedAt || item?.updated_at || '',
  category: item?.category ? normalizeCategory(item.category) : undefined,
  isFavorited: item?.isFavorited ?? item?.is_favorited ?? false,
  viewCount: item?.viewCount ?? item?.view_count,
  favoriteCount: item?.favoriteCount ?? item?.favorite_count,
  fullPath: item?.fullPath || item?.full_path,
  isActive: item?.isActive ?? item?.is_active,
})

const normalizeCategoryStat = (category: any): CategoryStat => ({
  id: Number(category?.id || 0),
  name: category?.name || '',
  code: category?.code || '',
  level: Number(category?.level || 0),
  description: category?.description || '',
  statuteCount: Number(category?.statuteCount ?? category?.statute_count ?? 0),
})

const normalizeSearchResponse = (payload: any): LegalSearchResponse => {
  const raw = payload?.data && !payload?.statutes ? payload.data : payload || {}
  const pageSize = Number(raw.pageSize ?? raw.page_size ?? 20)
  const total = Number(raw.total || 0)

  return {
    total,
    page: Number(raw.page || 1),
    pageSize,
    totalPages: Number(raw.totalPages ?? raw.total_pages ?? (Math.ceil(total / pageSize) || 0)),
    statutes: Array.isArray(raw.statutes) ? raw.statutes.map(normalizeLawItem) : [],
    categories: Array.isArray(raw.categories)
      ? raw.categories.map(normalizeCategoryStat)
      : undefined,
    suggestions: Array.isArray(raw.suggestions) ? raw.suggestions : undefined,
    searchTime: Number(raw.searchTime ?? raw.search_time_ms ?? 0),
  }
}

const toLegalSearchParams = (params: LegalSearchRequest = {}) => {
  const queryParams: Record<string, string | number | boolean> = {}

  if (params.query?.trim()) queryParams.query = params.query.trim()
  if (params.categoryId) queryParams.category_id = params.categoryId
  if (params.lawName?.trim()) queryParams.law_name = params.lawName.trim()
  if (params.status) queryParams.status = params.status
  if (params.effectiveFrom) queryParams.effective_from = params.effectiveFrom
  if (params.effectiveTo) queryParams.effective_to = params.effectiveTo
  if (params.includeInactive !== undefined) queryParams.include_inactive = params.includeInactive
  if (params.sortBy) queryParams.sort_by = params.sortBy
  if (params.sortOrder) queryParams.sort_order = params.sortOrder
  if (params.page) queryParams.page = params.page
  if (params.pageSize) queryParams.page_size = params.pageSize
  if (params.tags?.length) queryParams.tags = params.tags.join(',')

  return queryParams
}

/**
 * 搜索法条 - 使用后端正确的路由
 * GET /api/v1/legal/statutes/search
 */
export async function searchLaws(params: LegalSearchRequest) {
  const response = await get<LegalSearchResponse>(
    '/legal/statutes/search',
    toLegalSearchParams(params),
  )
  return wrapData(normalizeSearchResponse(response))
}

/**
 * 获取法条详情
 * GET /api/v1/legal/statutes/:id
 */
export async function getLawById(id: number) {
  const response = await get<LawItem>(`/legal/statutes/${id}`)
  return wrapData(normalizeLawItem(response))
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
export async function getLaws(params?: LegalSearchRequest) {
  const response = await get<LegalSearchResponse>(
    '/legal/statutes/search',
    toLegalSearchParams(params),
  )
  return wrapData(normalizeSearchResponse(response))
}

/**
 * 获取法条分类
 * GET /api/v1/legal/categories
 */
export async function getLawCategories() {
  const response = await get<LawCategory[]>('/legal/categories')
  return wrapData((Array.isArray(response) ? response : []).map(normalizeCategory))
}

/**
 * 获取分类树
 * GET /api/v1/legal/categories/tree
 */
export async function getLawCategoryTree() {
  const response = await get<LawCategory[]>('/legal/categories/tree')
  return wrapData((Array.isArray(response) ? response : []).map(normalizeCategory))
}

/**
 * 获取法条标签
 * GET /api/v1/legal/tags
 */
export async function getLawTags() {
  const response = await get<LawTag[]>('/legal/tags')
  return wrapData((Array.isArray(response) ? response : []).map(normalizeTag))
}

export async function getPopularLawTags(limit?: number) {
  const response = await get<LawTag[]>('/legal/tags/popular', { limit })
  return wrapData((Array.isArray(response) ? response : []).map(normalizeTag))
}

/**
 * 获取搜索建议
 * GET /api/v1/legal/search/suggestions
 */
export async function getLawSuggestions(query: string) {
  const response = await get<string[]>('/legal/search/suggestions', { query })
  return wrapData(Array.isArray(response) ? response : [])
}

export async function getRelatedLaws(id: number, limit?: number) {
  const response = await get<LawItem[]>(`/legal/statutes/${id}/related`, { limit })
  return wrapData((Array.isArray(response) ? response : []).map(normalizeLawItem))
}

export async function getPopularLaws(limit?: number) {
  const response = await get<LawItem[]>('/legal/stats/popular', { limit })
  return wrapData((Array.isArray(response) ? response : []).map(normalizeLawItem))
}

export async function getRecentLaws(days?: number) {
  const response = await get<LawItem[]>('/legal/stats/recent', { days })
  return wrapData((Array.isArray(response) ? response : []).map(normalizeLawItem))
}

export async function getCategoryStats() {
  const response = await get<CategoryStat[]>('/legal/stats/categories')
  return wrapData((Array.isArray(response) ? response : []).map(normalizeCategoryStat))
}

export async function getPopularSearches(limit?: number) {
  const response = await get<string[]>('/legal/popular-searches', { limit })
  return wrapData(Array.isArray(response) ? response : [])
}

// =============================================================================
// 用户收藏相关
// =============================================================================

export async function addToFavorites(statuteId: number) {
  return post('/legal/favorites', { statute_id: statuteId })
}

export function removeFromFavorites(statuteId: number) {
  return del(`/legal/favorites/${statuteId}`)
}

export async function getUserFavorites(params?: { page?: number; pageSize?: number }) {
  const response = await get<LegalSearchResponse>('/legal/favorites', {
    page: params?.page,
    page_size: params?.pageSize,
  })
  return wrapData(normalizeSearchResponse(response))
}

export async function getSearchHistory(limit?: number) {
  const response = await get<SearchHistory[]>('/legal/search-history', { limit })
  return wrapData(Array.isArray(response) ? response : [])
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

export async function bulkImportStatutes(data: LegalStatuteImportRequest) {
  const response = await post<LegalStatuteImportResponse>('/legal/statutes/import', data)
  return wrapData(response)
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
