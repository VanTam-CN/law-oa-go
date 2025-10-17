import { get, post } from './http';

// 诉讼费计算
export interface LitigationFeeParams {
  amount: number;
}

export interface LitigationFeeResult {
  amount: number;
  fee: number;
  calculationTime: number;
}

export function calculateLitigationFee(params: LitigationFeeParams) {
  return post('/tools/litigation-fee', params);
}

// 利息计算器
export interface InterestCalculatorParams {
  principal: number;
  rate: number;
  days: number;
  type: 'simple' | 'compound' | 'penalty';
}

export interface InterestCalculatorResult {
  principal: number;
  rate: number;
  days: number;
  type: string;
  interest: number;
  total: number;
}

export function calculateInterest(params: InterestCalculatorParams) {
  return post('/tools/interest-calculator', params);
}

// 工期计算器
export interface DeadlineCalculatorParams {
  startDate: string;
  days: number;
  excludeWeekends: boolean;
  excludeHolidays: boolean;
}

export interface DeadlineCalculatorResult {
  startDate: string;
  days: number;
  excludeWeekends: boolean;
  excludeHolidays: boolean;
  endDate: string;
  workDays: number;
}

export function calculateDeadline(params: DeadlineCalculatorParams) {
  return post('/tools/deadline-calculator', params);
}

// 法条查询
export interface LawItem {
  id: number;
  statuteNumber: string;
  title: string;
  content: string;
  lawName: string;
  chapter?: string;
  section?: string;
  part?: string;
  effectiveDate?: string;
  expiryDate?: string;
  publishingAuthority?: string;
  status: string;
  hierarchyLevel: number;
  parentStatuteId?: number;
  orderInHierarchy?: number;
  tags: string[];
  keywords: string[];
  createdAt: string;
  updatedAt: string;
  category?: LawCategory;
  isFavorited?: boolean;
  viewCount?: number;
  favoriteCount?: number;
  fullPath?: string;
  isActive?: boolean;
}

// 法条分类
export interface LawCategory {
  id: number;
  name: string;
  code: string;
  parentId?: number;
  level: number;
  description?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

// 法条标签
export interface LawTag {
  id: number;
  name: string;
  color: string;
  description?: string;
  usageCount: number;
  createdAt: string;
}

// 搜索请求参数
export interface LegalSearchRequest {
  query?: string;
  categoryId?: number;
  lawName?: string;
  status?: string;
  effectiveFrom?: string;
  effectiveTo?: string;
  tags?: string[];
  includeInactive?: boolean;
  sortBy?: 'relevance' | 'date' | 'title';
  sortOrder?: 'asc' | 'desc';
  page?: number;
  pageSize?: number;
}

// 搜索响应
export interface LegalSearchResponse {
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  statutes: LawItem[];
  categories?: CategoryStat[];
  suggestions?: string[];
  searchTime: number;
}

// 分类统计
export interface CategoryStat {
  id: number;
  name: string;
  code: string;
  level: number;
  description: string;
  statuteCount: number;
}

// 搜索历史
export interface SearchHistory {
  id: number;
  userId?: number;
  searchQuery: string;
  searchFilters: Record<string, any>;
  resultCount: number;
  searchDuration?: number;
  createdAt: string;
}

// 用户收藏
export interface UserFavorite {
  id: number;
  statuteId: number;
  userId: number;
  createdAt: string;
  statute?: LawItem;
}

// 创建法条请求
export interface CreateStatuteRequest {
  statuteNumber: string;
  title: string;
  content: string;
  categoryId: number;
  lawName: string;
  chapter?: string;
  section?: string;
  part?: string;
  effectiveDate?: string;
  expiryDate?: string;
  publishingAuthority?: string;
  status?: string;
  hierarchyLevel?: number;
  parentStatuteId?: number;
  orderInHierarchy?: number;
  tags?: string[];
  keywords?: string[];
}

// 更新法条请求
export interface UpdateStatuteRequest {
  title?: string;
  content?: string;
  categoryId?: number;
  chapter?: string;
  section?: string;
  part?: string;
  effectiveDate?: string;
  expiryDate?: string;
  publishingAuthority?: string;
  status?: string;
  orderInHierarchy?: number;
  tags?: string[];
  keywords?: string[];
  changeDescription?: string;
}

// 收藏请求
export interface FavoriteRequest {
  statuteId: number;
}

// API函数
export function getLaws(params?: LegalSearchRequest) {
  return get('/tools/laws', params);
}

export function getLawById(id: number) {
  return get(`/tools/laws/${id}`);
}

export function getLawByNumber(number: string) {
  return get(`/tools/laws/number/${number}`);
}

export function searchLaws(params: LegalSearchRequest) {
  return get('/api/v1/legal/statutes/search', params);
}

export function getLawCategories() {
  return get('/tools/laws/categories');
}

export function getLawCategoryTree() {
  return get('/api/v1/legal/categories/tree');
}

export function getLawTags() {
  return get('/tools/laws/tags');
}

export function getPopularLawTags(limit?: number) {
  return get('/api/v1/legal/tags/popular', { limit });
}

export function getLawSuggestions(query: string) {
  return get('/api/v1/legal/statutes/suggestions', { query });
}

export function getRelatedLaws(id: number, limit?: number) {
  return get(`/api/v1/legal/statutes/${id}/related`, { limit });
}

export function getPopularLaws(limit?: number) {
  return get('/api/v1/legal/stats/popular', { limit });
}

export function getRecentLaws(days?: number) {
  return get('/api/v1/legal/stats/recent', { days });
}

export function getCategoryStats() {
  return get('/api/v1/legal/stats/categories');
}

export function getPopularSearches(limit?: number) {
  return get('/api/v1/legal/popular-searches', { limit });
}

// 用户收藏相关
export function addToFavorites(statuteId: number) {
  return post('/api/v1/legal/favorites', { statuteId });
}

export function removeFromFavorites(statuteId: number) {
  return fetch(`/api/v1/legal/favorites/${statuteId}`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
  });
}

export function getUserFavorites(params?: { page?: number; pageSize?: number }) {
  return get('/api/v1/legal/favorites', params);
}

export function getSearchHistory(limit?: number) {
  return get('/api/v1/legal/search-history', { limit });
}

// 管理员功能
export function syncToElasticsearch() {
  return post('/api/v1/legal/admin/sync-elasticsearch');
}

export function rebuildSearchIndex() {
  return post('/api/v1/legal/admin/rebuild-index');
}

// 合同模板
export interface ContractTemplate {
  id: number;
  name: string;
  category: string;
  description: string;
  downloadUrl: string;
  updateTime: string;
}

export function getContractTemplates() {
  return get('/tools/contract-templates');
}

// 文档转换
export interface DocumentConvertParams {
  sourceFormat: string;
  targetFormat: string;
}

export interface DocumentConvertResult {
  sourceFormat: string;
  targetFormat: string;
  status: string;
  message: string;
  downloadUrl: string;
}

export function convertDocument(params: DocumentConvertParams) {
  return post('/tools/document-convert', params);
}

// 翻译服务
export interface TranslateParams {
  text: string;
  targetLang: string;
}

export interface TranslateResult {
  originalText: string;
  targetLang: string;
  translatedText: string;
  sourceLang: string;
}

export function translate(params: TranslateParams) {
  return post('/tools/translate', params);
}