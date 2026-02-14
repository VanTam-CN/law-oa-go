import { get, post, put, del } from './http'

// ============================================================================
// 类型定义
// ============================================================================

// 代管款账户状态
export type AccountStatus = 'active' | 'frozen' | 'closed'

// 币种
export type Currency = 'CNY' | 'USD' | 'EUR' | 'HKD'

// 交易类型
export type TransactionType = 'deposit' | 'withdraw' | 'transfer_in' | 'transfer_out' | 'freeze' | 'unfreeze'

// 交易状态
export type TransactionStatus = 'pending' | 'approved' | 'rejected' | 'completed'

// 代管款账户
export interface TrustAccount {
  id: number
  account_code: string
  client_id: number
  client_name?: string
  currency: Currency
  balance: number
  frozen_amount: number
  available_amount: number
  status: AccountStatus
  description?: string
  created_at: string
  updated_at: string
  closed_at?: string
  // 关联交易统计
  transaction_count?: number
  last_transaction_at?: string
}

// 创建账户请求
export interface CreateAccountRequest {
  client_id: number
  currency: Currency
  description?: string
}

// 账户列表请求参数
export interface ListAccountsRequest {
  page?: number
  page_size?: number
  client_id?: number
  status?: AccountStatus
  currency?: Currency
  search?: string
}

// 账户列表响应
export interface ListAccountsResponse {
  accounts: TrustAccount[]
  pagination: {
    page: number
    page_size: number
    total: number
  }
}

// 账户统计
export interface AccountStats {
  total_accounts: number
  total_balance: number
  total_frozen: number
  active_accounts: number
  by_currency: {
    currency: Currency
    count: number
    balance: number
  }[]
}

// 代管款交易
export interface TrustTransaction {
  id: number
  transaction_code: string
  account_id: number
  account_code?: string
  transaction_type: TransactionType
  amount: number
  balance_after: number
  status: TransactionStatus
  description: string
  reference_no?: string
  related_case_id?: number
  related_contract_id?: number
  created_by: number
  created_by_name?: string
  approved_by?: number
  approved_by_name?: string
  approved_at?: string
  created_at: string
  updated_at: string
  // 关联信息
  case?: {
    id: number
    title: string
  }
  contract?: {
    id: number
    contract_code: string
  }
}

// 创建交易请求
export interface CreateTransactionRequest {
  account_id: number
  transaction_type: TransactionType
  amount: number
  description: string
  reference_no?: string
  related_case_id?: number
  related_contract_id?: number
}

// 交易列表请求参数
export interface ListTransactionsRequest {
  page?: number
  page_size?: number
  account_id?: number
  transaction_type?: TransactionType
  status?: TransactionStatus
  date_from?: string
  date_to?: string
}

// 交易列表响应
export interface ListTransactionsResponse {
  transactions: TrustTransaction[]
  pagination: {
    page: number
    page_size: number
    total: number
  }
}

// API响应类型
export interface APIResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: string
  }
  meta: {
    timestamp: string
    version: string
    server: string
    environment: string
  }
  pagination?: {
    page: number
    page_size: number
    total: number
    total_pages: number
    has_next: boolean
    has_prev: boolean
  }
}

// ============================================================================
// 账户API
// ============================================================================

/**
 * 获取账户列表
 */
export const getTrustAccounts = (params?: ListAccountsRequest): Promise<APIResponse<ListAccountsResponse>> => {
  return get<APIResponse<ListAccountsResponse>>('/trust/accounts', params)
}

/**
 * 获取账户详情
 */
export const getTrustAccount = (id: number): Promise<APIResponse<TrustAccount>> => {
  return get<APIResponse<TrustAccount>>(`/trust/accounts/${id}`)
}

/**
 * 创建账户
 */
export const createTrustAccount = (data: CreateAccountRequest): Promise<APIResponse<TrustAccount>> => {
  return post<APIResponse<TrustAccount>>('/trust/accounts', data)
}

/**
 * 冻结账户
 */
export const freezeTrustAccount = (id: number): Promise<APIResponse<TrustAccount>> => {
  return post<APIResponse<TrustAccount>>(`/trust/accounts/${id}/freeze`, {})
}

/**
 * 解冻账户
 */
export const unfreezeTrustAccount = (id: number): Promise<APIResponse<TrustAccount>> => {
  return post<APIResponse<TrustAccount>>(`/trust/accounts/${id}/unfreeze`, {})
}

/**
 * 关闭账户
 */
export const closeTrustAccount = (id: number): Promise<APIResponse<TrustAccount>> => {
  return post<APIResponse<TrustAccount>>(`/trust/accounts/${id}/close`, {})
}

/**
 * 获取账户交易记录
 */
export const getAccountTransactions = (
  accountId: number,
  params?: Pick<ListTransactionsRequest, 'page' | 'page_size'>
): Promise<APIResponse<ListTransactionsResponse>> => {
  return get<APIResponse<ListTransactionsResponse>>(`/trust/accounts/${accountId}/transactions`, params)
}

/**
 * 获取账户统计
 */
export const getAccountStats = (): Promise<APIResponse<AccountStats>> => {
  return get<APIResponse<AccountStats>>('/trust/stats')
}

// ============================================================================
// 交易API
// ============================================================================

/**
 * 获取交易列表
 */
export const getTrustTransactions = (params?: ListTransactionsRequest): Promise<APIResponse<ListTransactionsResponse>> => {
  return get<APIResponse<ListTransactionsResponse>>('/trust/transactions', params)
}

/**
 * 获取交易详情
 */
export const getTrustTransaction = (id: number): Promise<APIResponse<TrustTransaction>> => {
  return get<APIResponse<TrustTransaction>>(`/trust/transactions/${id}`)
}

/**
 * 创建交易
 */
export const createTrustTransaction = (data: CreateTransactionRequest): Promise<APIResponse<TrustTransaction>> => {
  return post<APIResponse<TrustTransaction>>('/trust/transactions', data)
}

/**
 * 审批通过交易
 */
export const approveTrustTransaction = (id: number): Promise<APIResponse<TrustTransaction>> => {
  return post<APIResponse<TrustTransaction>>(`/trust/transactions/${id}/approve`, {})
}

/**
 * 审批拒绝交易
 */
export const rejectTrustTransaction = (id: number): Promise<APIResponse<TrustTransaction>> => {
  return post<APIResponse<TrustTransaction>>(`/trust/transactions/${id}/reject`, {})
}

// ============================================================================
// 辅助函数
// ============================================================================

/**
 * 账户状态文本映射
 */
export const accountStatusMap: Record<AccountStatus, { text: string; color: string }> = {
  active: { text: '正常', color: 'success' },
  frozen: { text: '已冻结', color: 'error' },
  closed: { text: '已关闭', color: 'default' },
}

/**
 * 交易类型文本映射
 */
export const transactionTypeMap: Record<TransactionType, { text: string; icon: string; color: string }> = {
  deposit: { text: '存入', icon: '↓', color: 'success' },
  withdraw: { text: '支取', icon: '↑', color: 'warning' },
  transfer_in: { text: '转入', icon: '→', color: 'success' },
  transfer_out: { text: '转出', icon: '←', color: 'warning' },
  freeze: { text: '冻结', icon: '⚠', color: 'error' },
  unfreeze: { text: '解冻', icon: '✓', color: 'processing' },
}

/**
 * 交易状态文本映射
 */
export const transactionStatusMap: Record<TransactionStatus, { text: string; color: string }> = {
  pending: { text: '待审批', color: 'warning' },
  approved: { text: '已审批', color: 'processing' },
  rejected: { text: '已拒绝', color: 'error' },
  completed: { text: '已完成', color: 'success' },
}

/**
 * 币种符号映射
 */
export const currencySymbolMap: Record<Currency, string> = {
  CNY: '¥',
  USD: '$',
  EUR: '€',
  HKD: 'HK$',
}

/**
 * 格式化金额
 */
export const formatAmount = (amount: number, currency: Currency = 'CNY'): string => {
  const symbol = currencySymbolMap[currency]
  return `${symbol}${amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

/**
 * 格式化日期
 */
export const formatDate = (dateStr?: string): string => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
