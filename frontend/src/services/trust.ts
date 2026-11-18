import { get, post } from './http'

// ============================================================================
// 类型定义
// ============================================================================

// 代管款账户状态
export type AccountStatus = 'active' | 'frozen' | 'closed'

// 币种
export type Currency = 'CNY' | 'USD' | 'EUR'

// 交易类型
export type TransactionType = 'deposit' | 'deposit_refund' | 'withdraw' | 'transfer'

// 交易状态
export type TransactionStatus = 'pending' | 'completed' | 'cancelled'

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
  available_balance?: number
  status: AccountStatus
  purpose_restriction?: string
  authorized_uses?: string[]
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
  purpose_restriction?: string
  authorized_uses?: string[]
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
  purpose_code?: string
  case_id?: number
  recipient_name?: string
  recipient_bank_account?: string
  recipient_bank_name?: string
  created_by: number
  created_by_name?: string
  approved_by?: number
  approved_by_name?: string
  approved_at?: string
  created_at: string
  updated_at: string
  // 关联信息
  account?: {
    id: number
    account_code: string
    balance: number
  }
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
  purpose_code?: string
  case_id?: number
  recipient_name?: string
  recipient_bank_account?: string
  recipient_bank_name?: string
  attachment_id?: number
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

const responseMeta = {
  timestamp: new Date().toISOString(),
  version: 'v1',
  server: 'law-oa-go',
  environment: import.meta.env.MODE,
}

function wrapResponse<T>(payload: T | APIResponse<T>): APIResponse<T> {
  if (
    payload &&
    typeof payload === 'object' &&
    'success' in payload &&
    'data' in payload
  ) {
    return payload as APIResponse<T>
  }

  return {
    success: true,
    data: payload as T,
    meta: responseMeta,
  }
}

function normalizeAuthorizedUses(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.map(String)
  }

  if (value && typeof value === 'object' && 'uses' in value) {
    const uses = (value as { uses?: unknown }).uses
    return Array.isArray(uses) ? uses.map(String) : []
  }

  return []
}

function normalizeAccount(raw: any): TrustAccount {
  const balance = Number(raw?.balance || 0)
  const frozenAmount = Number(raw?.frozen_amount || 0)
  const available = Number(raw?.available_amount ?? raw?.available_balance ?? balance - frozenAmount)

  return {
    ...raw,
    client_name: raw?.client_name || raw?.client?.name,
    balance,
    frozen_amount: frozenAmount,
    available_amount: available,
    available_balance: available,
    authorized_uses: normalizeAuthorizedUses(raw?.authorized_uses),
  }
}

function normalizeTransaction(raw: any): TrustTransaction {
  return {
    ...raw,
    account_code: raw?.account_code || raw?.account?.account_code,
    balance_after: Number(raw?.balance_after ?? raw?.account?.balance ?? 0),
  }
}

function normalizeAccountsResponse(payload: ListAccountsResponse | APIResponse<ListAccountsResponse>): APIResponse<ListAccountsResponse> {
  const wrapped = wrapResponse<ListAccountsResponse>(payload)
  const data = wrapped.data || { accounts: [], pagination: { page: 1, page_size: 10, total: 0 } }

  return {
    ...wrapped,
    data: {
      ...data,
      accounts: (data.accounts || []).map(normalizeAccount),
    },
  }
}

function normalizeTransactionsResponse(
  payload: ListTransactionsResponse | APIResponse<ListTransactionsResponse>,
): APIResponse<ListTransactionsResponse> {
  const wrapped = wrapResponse<ListTransactionsResponse>(payload)
  const data = wrapped.data || { transactions: [], pagination: { page: 1, page_size: 10, total: 0 } }

  return {
    ...wrapped,
    data: {
      ...data,
      transactions: (data.transactions || []).map(normalizeTransaction),
    },
  }
}

// ============================================================================
// 账户API
// ============================================================================

/**
 * 获取账户列表
 */
export const getTrustAccounts = async (
  params?: ListAccountsRequest,
): Promise<APIResponse<ListAccountsResponse>> => {
  return normalizeAccountsResponse(await get<ListAccountsResponse>('/trust/accounts', params))
}

/**
 * 获取账户详情
 */
export const getTrustAccount = async (id: number): Promise<APIResponse<TrustAccount>> => {
  const response = wrapResponse<TrustAccount>(await get<TrustAccount>(`/trust/accounts/${id}`))
  return { ...response, data: response.data ? normalizeAccount(response.data) : response.data }
}

/**
 * 创建账户
 */
export const createTrustAccount = async (
  data: CreateAccountRequest,
): Promise<APIResponse<TrustAccount>> => {
  const response = wrapResponse<TrustAccount>(await post<TrustAccount>('/trust/accounts', data))
  return { ...response, data: response.data ? normalizeAccount(response.data) : response.data }
}

/**
 * 冻结账户
 */
export const freezeTrustAccount = async (id: number): Promise<APIResponse<TrustAccount>> => {
  const response = wrapResponse<TrustAccount>(await post<TrustAccount>(`/trust/accounts/${id}/freeze`, {}))
  return { ...response, data: response.data ? normalizeAccount(response.data) : response.data }
}

/**
 * 解冻账户
 */
export const unfreezeTrustAccount = async (id: number): Promise<APIResponse<TrustAccount>> => {
  const response = wrapResponse<TrustAccount>(await post<TrustAccount>(`/trust/accounts/${id}/unfreeze`, {}))
  return { ...response, data: response.data ? normalizeAccount(response.data) : response.data }
}

/**
 * 关闭账户
 */
export const closeTrustAccount = async (id: number): Promise<APIResponse<TrustAccount>> => {
  const response = wrapResponse<TrustAccount>(await post<TrustAccount>(`/trust/accounts/${id}/close`, {}))
  return { ...response, data: response.data ? normalizeAccount(response.data) : response.data }
}

/**
 * 获取账户交易记录
 */
export const getAccountTransactions = (
  accountId: number,
  params?: Pick<ListTransactionsRequest, 'page' | 'page_size'>
): Promise<APIResponse<ListTransactionsResponse>> => {
  return get<ListTransactionsResponse>(`/trust/accounts/${accountId}/transactions`, params).then(
    normalizeTransactionsResponse,
  )
}

/**
 * 获取账户统计
 */
export const getAccountStats = async (): Promise<APIResponse<AccountStats>> => {
  return wrapResponse<AccountStats>(await get<AccountStats>('/trust/stats'))
}

// ============================================================================
// 交易API
// ============================================================================

/**
 * 获取交易列表
 */
export const getTrustTransactions = async (
  params?: ListTransactionsRequest,
): Promise<APIResponse<ListTransactionsResponse>> => {
  return normalizeTransactionsResponse(await get<ListTransactionsResponse>('/trust/transactions', params))
}

/**
 * 获取交易详情
 */
export const getTrustTransaction = async (id: number): Promise<APIResponse<TrustTransaction>> => {
  const response = wrapResponse<TrustTransaction>(await get<TrustTransaction>(`/trust/transactions/${id}`))
  return { ...response, data: response.data ? normalizeTransaction(response.data) : response.data }
}

/**
 * 创建交易
 */
export const createTrustTransaction = async (
  data: CreateTransactionRequest,
): Promise<APIResponse<TrustTransaction>> => {
  const response = wrapResponse<TrustTransaction>(await post<TrustTransaction>('/trust/transactions', data))
  return { ...response, data: response.data ? normalizeTransaction(response.data) : response.data }
}

/**
 * 审批通过交易
 */
export const approveTrustTransaction = async (id: number): Promise<APIResponse<TrustTransaction>> => {
  const response = wrapResponse<TrustTransaction>(await post<TrustTransaction>(`/trust/transactions/${id}/approve`, {}))
  return { ...response, data: response.data ? normalizeTransaction(response.data) : response.data }
}

/**
 * 审批拒绝交易
 */
export const rejectTrustTransaction = async (id: number): Promise<APIResponse<TrustTransaction>> => {
  const response = wrapResponse<TrustTransaction>(await post<TrustTransaction>(`/trust/transactions/${id}/reject`, {}))
  return { ...response, data: response.data ? normalizeTransaction(response.data) : response.data }
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
  deposit_refund: { text: '退回存入', icon: '↩', color: 'warning' },
  withdraw: { text: '支取', icon: '↑', color: 'warning' },
  transfer: { text: '转账', icon: '→', color: 'processing' },
}

/**
 * 交易状态文本映射
 */
export const transactionStatusMap: Record<TransactionStatus, { text: string; color: string }> = {
  pending: { text: '待审批', color: 'warning' },
  completed: { text: '已完成', color: 'success' },
  cancelled: { text: '已取消', color: 'error' },
}

/**
 * 币种符号映射
 */
export const currencySymbolMap: Record<Currency, string> = {
  CNY: '¥',
  USD: '$',
  EUR: '€',
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
