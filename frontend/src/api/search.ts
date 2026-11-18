import { request } from '@/utils/request'

/**
 * 全文检索API
 */

// 全文搜索
export function fullTextSearch(params: { keyword: string; pageNum?: number; pageSize?: number }) {
  return request({
    url: '/api/search/fulltext',
    method: 'get',
    params,
  })
}

// 高级搜索
export function advancedSearch(data: {
  keyword?: string
  caseType?: string
  caseStatus?: string
  responsibleLawyer?: string
  startDate?: string
  endDate?: string
  minAmount?: number
  maxAmount?: number
  pageNum?: number
  pageSize?: number
}) {
  return request({
    url: '/api/search/advanced',
    method: 'post',
    data,
  })
}

// 搜索建议
export function searchSuggestions(params: { prefix: string; size?: number }) {
  return request({
    url: '/api/search/suggestions',
    method: 'get',
    params,
  })
}

// 同步单个案例
export function syncCase(caseId: number) {
  return request({
    url: `/api/search/sync/${caseId}`,
    method: 'post',
  })
}

// 批量同步案例
export function batchSyncCases(caseIds: number[]) {
  return request({
    url: '/api/search/sync/batch',
    method: 'post',
    data: caseIds,
  })
}

// 同步所有案例
export function syncAllCases() {
  return request({
    url: '/api/search/sync/all',
    method: 'post',
  })
}

// 从索引中删除案例
export function deleteCaseFromIndex(caseId: number) {
  return request({
    url: `/api/search/sync/${caseId}`,
    method: 'delete',
  })
}
