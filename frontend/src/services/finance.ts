import { get, post, put, del } from './http'

// ============================================================================
// 类型定义
// ============================================================================

// 合同相关类型
export interface Contract {
  id: number
  contract_code: string
  case_id?: number
  client_id: number
  contract_amount: number
  currency: string
  billing_cycle: string
  payment_terms: string
  start_date?: string
  end_date?: string
  status: 'draft' | 'active' | 'suspended' | 'completed' | 'cancelled'
  contract_type: 'original' | 'supplementary'
  parent_contract_id?: number
  signed_at?: string
  document_id?: number
  created_at: string
  updated_at: string
  client?: {
    id: number
    name: string
  }
  case?: {
    id: number
    title: string
  }
  milestones?: PaymentMilestone[]
  supplementary_contracts?: Contract[]
}

export interface CreateContractRequest {
  contract_code: string
  case_id?: number
  client_id: number
  contract_amount: number
  currency: string
  billing_cycle: string
  payment_terms?: string
  start_date?: string
  end_date?: string
  contract_type: 'original' | 'supplementary'
  parent_contract_id?: number
  document_id?: number
  milestones?: CreateMilestoneRequest[]
}

export interface UpdateContractRequest {
  contract_amount?: number
  currency?: string
  billing_cycle?: string
  payment_terms?: string
  start_date?: string
  end_date?: string
  document_id?: number
}

export interface ContractStats {
  total_contracts: number
  draft_contracts: number
  active_contracts: number
  suspended_contracts: number
  completed_contracts: number
  cancelled_contracts: number
  total_contract_amount: number
  new_contracts_this_month: number
}

// 付款计划相关类型
export interface PaymentMilestone {
  id: number
  contract_id: number
  name: string
  sequence: number
  amount: number
  percentage: number
  due_date?: string
  condition: string
  status: 'pending' | 'billed' | 'partial_paid' | 'paid' | 'overdue'
  invoice_id?: number
  paid_amount: number
}

export interface CreateMilestoneRequest {
  contract_id: number
  name: string
  amount: number
  percentage: number
  due_date?: string
  condition?: string
}

export interface UpdateMilestoneRequest {
  name?: string
  amount?: number
  percentage?: number
  due_date?: string
  condition?: string
}

// 发票相关类型
export interface Invoice {
  id: number
  invoice_code: string
  contract_id?: number
  milestone_id?: number
  client_id: number
  amount: number
  tax_rate: number
  tax_amount: number
  total_amount: number
  // 客户开票信息
  client_name: string
  client_tax_id: string
  client_address: string
  client_bank_name: string
  client_bank_account: string
  // 发票类型
  invoice_type: 'normal' | 'credit'
  original_invoice_id?: number
  refund_reason: string
  write_off_amount: number
  // 状态
  status: 'draft' | 'submitted' | 'approved' | 'issued' | 'received' | 'cancelled'
  submitted_at?: string
  approved_by_finance_at?: string
  issued_at?: string
  received_at?: string
  // 电子发票
  electronic_invoice_url: string
  electronic_invoice_code: string
  electronic_invoice_number: string
  // 审批信息
  created_by: number
  submitted_by?: number
  approved_by?: number
  created_at: string
  updated_at: string
  // 关联数据
  client?: {
    id: number
    name: string
  }
  contract?: {
    id: number
    contract_code: string
    contract_amount: number
  }
  milestone?: {
    id: number
    name: string
    amount: number
  }
  payments?: PaymentSummary[]
  total_paid_amount: number
  remaining_amount: number
}

export interface CreateInvoiceRequest {
  invoice_code: string
  contract_id?: number
  milestone_id?: number
  client_id: number
  amount: number
  tax_rate: number
  invoice_type: 'normal' | 'credit'
  client_name: string
  client_tax_id?: string
  client_address?: string
  client_bank_name?: string
  client_bank_account?: string
  original_invoice_id?: number
  refund_reason?: string
}

export interface UpdateInvoiceRequest {
  amount?: number
  tax_rate?: number
  client_name?: string
  client_tax_id?: string
  client_address?: string
  client_bank_name?: string
  client_bank_account?: string
}

export interface InvoiceStats {
  total_invoices: number
  draft_invoices: number
  submitted_invoices: number
  approved_invoices: number
  issued_invoices: number
  received_invoices: number
  total_invoice_amount: number
  pending_invoice_amount: number
  overdue_amount: number
}

// 回款相关类型
export interface Payment {
  id: number
  payment_code: string
  invoice_id: number
  amount: number
  payment_date: string
  payment_method: 'bank_transfer' | 'cash' | 'other'
  reference_no: string
  payer_name: string
  payer_account: string
  attachment_id?: number
  confirmed_by: number
  confirmed_at?: string
  status: 'pending' | 'confirmed' | 'rejected'
  remark: string
  created_at: string
  invoice?: {
    id: number
    invoice_code: string
    total_amount: number
    client_name: string
  }
}

export interface CreatePaymentRequest {
  invoice_id: number
  amount: number
  payment_date: string
  payment_method: 'bank_transfer' | 'cash' | 'other'
  reference_no?: string
  payer_name?: string
  payer_account?: string
  attachment_id?: number
  remark?: string
}

export interface UpdatePaymentRequest {
  amount?: number
  payment_date?: string
  payment_method?: 'bank_transfer' | 'cash' | 'other'
  reference_no?: string
  payer_name?: string
  payer_account?: string
  attachment_id?: number
  remark?: string
}

export interface PaymentStats {
  total_payments: number
  pending_payments: number
  confirmed_payments: number
  rejected_payments: number
  total_payment_amount: number
  month_payment_amount: number
  pending_amount: number
}

// 坏账核销相关类型
export interface BadDebt {
  id: number
  contract_id: number
  invoice_id?: number
  original_amount: number
  write_off_amount: number
  remaining_amount: number
  reason: string
  reason_type: 'bankruptcy' | 'dispute' | 'uncollectible' | 'other'
  status: 'pending' | 'approved' | 'rejected'
  approved_by?: number
  approved_at?: string
  approval_notes: string
  attachment_ids: number[]
  created_at: string
  updated_at: string
  contract?: {
    id: number
    contract_code: string
    contract_amount: number
  }
  invoice?: {
    id: number
    invoice_code: string
    total_amount: number
    client_name: string
  }
}

export interface CreateBadDebtRequest {
  contract_id: number
  invoice_id?: number
  write_off_amount: number
  reason: string
  reason_type: 'bankruptcy' | 'dispute' | 'uncollectible' | 'other'
  attachment_ids?: number[]
}

// 提成相关类型
export interface Commission {
  id: number
  commission_code: string
  contract_id: number
  payment_id: number
  case_id?: number
  beneficiary_id: number
  beneficiary_role: 'source' | 'lawyer' | 'assistant'
  payment_amount: number
  cost_deduction: number
  commission_base: number
  commission_rate: number
  commission_amount: number
  status: 'pending' | 'calculated' | 'paid' | 'cancelled'
  paid_date?: string
  payment_voucher: string
  calculated_at?: string
  created_at: string
  updated_at: string
  contract?: {
    id: number
    contract_code: string
    contract_amount: number
  }
  payment?: {
    id: number
    payment_code: string
    amount: number
  }
  case?: {
    id: number
    title: string
  }
  beneficiary?: {
    id: number
    name: string
  }
}

export interface CalculateCommissionRequest {
  payment_id: number
  cost_deduction?: number
}

export interface CommissionStats {
  total_commissions: number
  pending_commissions: number
  calculated_commissions: number
  paid_commissions: number
  cancelled_commissions: number
  total_commission_amount: number
  pending_commission_amount: number
  month_commission_amount: number
}

// 财务概览
export interface FinanceOverview {
  contract_stats: ContractStats
  invoice_stats: InvoiceStats
  payment_stats: PaymentStats
  commission_stats: CommissionStats
}

// 分页响应类型
export interface PaginatedResponse<T> {
  data: T[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
    has_next: boolean
    has_prev: boolean
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
// 合同API
// ============================================================================

/**
 * 获取合同列表
 */
export const getContracts = (params?: {
  page?: number
  page_size?: number
  status?: string
  contract_type?: string
  client_id?: number
  case_id?: number
  search?: string
  start_date_from?: string
  start_date_to?: string
  end_date_from?: string
  end_date_to?: string
}): Promise<APIResponse<Contract[]>> => {
  return get<APIResponse<Contract[]>>('/finance/contracts', params)
}

/**
 * 获取合同详情
 */
export const getContract = (id: number): Promise<APIResponse<Contract>> => {
  return get<APIResponse<Contract>>(`/finance/contracts/${id}`)
}

/**
 * 创建合同
 */
export const createContract = (data: CreateContractRequest): Promise<APIResponse<Contract>> => {
  return post<APIResponse<Contract>>('/finance/contracts', data)
}

/**
 * 更新合同
 */
export const updateContract = (id: number, data: UpdateContractRequest): Promise<APIResponse<Contract>> => {
  return put<APIResponse<Contract>>(`/finance/contracts/${id}`, data)
}

/**
 * 删除合同
 */
export const deleteContract = (id: number): Promise<APIResponse<{ message: string }>> => {
  return del<APIResponse<{ message: string }>>(`/finance/contracts/${id}`)
}

/**
 * 激活合同
 */
export const activateContract = (id: number): Promise<APIResponse<Contract>> => {
  return post<APIResponse<Contract>>(`/finance/contracts/${id}/activate`, {})
}

/**
 * 暂停合同
 */
export const suspendContract = (id: number): Promise<APIResponse<Contract>> => {
  return post<APIResponse<Contract>>(`/finance/contracts/${id}/suspend`, {})
}

/**
 * 完成合同
 */
export const completeContract = (id: number): Promise<APIResponse<Contract>> => {
  return post<APIResponse<Contract>>(`/finance/contracts/${id}/complete`, {})
}

/**
 * 获取合同统计
 */
export const getContractStats = (): Promise<APIResponse<ContractStats>> => {
  return get<APIResponse<ContractStats>>('/finance/contracts/stats')
}

/**
 * 获取客户的合同列表
 */
export const getContractsByClient = (clientId: number): Promise<APIResponse<Contract[]>> => {
  return get<APIResponse<Contract[]>>(`/finance/contracts?client_id=${clientId}`)
}

// ============================================================================
// 付款计划API
// ============================================================================

/**
 * 获取合同的付款计划
 */
export const getContractMilestones = (contractId: number): Promise<APIResponse<PaymentMilestone[]>> => {
  return get<APIResponse<PaymentMilestone[]>>(`/finance/contracts/${contractId}/milestones`)
}

/**
 * 创建付款计划
 */
export const createMilestone = (data: CreateMilestoneRequest): Promise<APIResponse<PaymentMilestone>> => {
  return post<APIResponse<PaymentMilestone>>('/finance/milestones', data)
}

/**
 * 更新付款计划
 */
export const updateMilestone = (id: number, data: UpdateMilestoneRequest): Promise<APIResponse<PaymentMilestone>> => {
  return put<APIResponse<PaymentMilestone>>(`/finance/milestones/${id}`, data)
}

/**
 * 删除付款计划
 */
export const deleteMilestone = (id: number): Promise<APIResponse<{ message: string }>> => {
  return del<APIResponse<{ message: string }>>(`/finance/milestones/${id}`)
}

// ============================================================================
// 发票API
// ============================================================================

/**
 * 获取发票列表
 */
export const getInvoices = (params?: {
  page?: number
  page_size?: number
  status?: string
  invoice_type?: string
  client_id?: number
  contract_id?: number
  search?: string
  date_from?: string
  date_to?: string
}): Promise<APIResponse<Invoice[]>> => {
  return get<APIResponse<Invoice[]>>('/finance/invoices', params)
}

/**
 * 获取发票详情
 */
export const getInvoice = (id: number): Promise<APIResponse<Invoice>> => {
  return get<APIResponse<Invoice>>(`/finance/invoices/${id}`)
}

/**
 * 创建发票
 */
export const createInvoice = (data: CreateInvoiceRequest): Promise<APIResponse<Invoice>> => {
  return post<APIResponse<Invoice>>('/finance/invoices', data)
}

/**
 * 更新发票
 */
export const updateInvoice = (id: number, data: UpdateInvoiceRequest): Promise<APIResponse<Invoice>> => {
  return put<APIResponse<Invoice>>(`/finance/invoices/${id}`, data)
}

/**
 * 删除发票
 */
export const deleteInvoice = (id: number): Promise<APIResponse<{ message: string }>> => {
  return del<APIResponse<{ message: string }>>(`/finance/invoices/${id}`)
}

/**
 * 提交发票审批
 */
export const submitInvoice = (id: number): Promise<APIResponse<Invoice>> => {
  return post<APIResponse<Invoice>>(`/finance/invoices/${id}/submit`, {})
}

/**
 * 审批通过发票
 */
export const approveInvoice = (id: number): Promise<APIResponse<Invoice>> => {
  return post<APIResponse<Invoice>>(`/finance/invoices/${id}/approve`, {})
}

/**
 * 审批拒绝发票
 */
export const rejectInvoice = (id: number): Promise<APIResponse<Invoice>> => {
  return post<APIResponse<Invoice>>(`/finance/invoices/${id}/reject`, {})
}

/**
 * 开票
 */
export const issueInvoice = (id: number, data: {
  electronic_url: string
  code: string
  number: string
}): Promise<APIResponse<Invoice>> => {
  return post<APIResponse<Invoice>>(`/finance/invoices/${id}/issue`, data)
}

/**
 * 客户签收
 */
export const confirmInvoiceReceipt = (id: number): Promise<APIResponse<Invoice>> => {
  return post<APIResponse<Invoice>>(`/finance/invoices/${id}/confirm-receipt`, {})
}

/**
 * 作废发票
 */
export const cancelInvoice = (id: number): Promise<APIResponse<Invoice>> => {
  return post<APIResponse<Invoice>>(`/finance/invoices/${id}/cancel`, {})
}

/**
 * 获取发票统计
 */
export const getInvoiceStats = (): Promise<APIResponse<InvoiceStats>> => {
  return get<APIResponse<InvoiceStats>>('/finance/invoices/stats')
}

/**
 * 获取发票的回款记录
 */
export const getInvoicePayments = (invoiceId: number): Promise<APIResponse<Payment[]>> => {
  return get<APIResponse<Payment[]>>(`/finance/invoices/${invoiceId}/payments`)
}

// ============================================================================
// 回款API
// ============================================================================

/**
 * 获取回款列表
 */
export const getPayments = (params?: {
  page?: number
  page_size?: number
  status?: string
  invoice_id?: number
  client_id?: number
  search?: string
  date_from?: string
  date_to?: string
}): Promise<APIResponse<Payment[]>> => {
  return get<APIResponse<Payment[]>>('/finance/payments', params)
}

/**
 * 获取回款详情
 */
export const getPayment = (id: number): Promise<APIResponse<Payment>> => {
  return get<APIResponse<Payment>>(`/finance/payments/${id}`)
}

/**
 * 创建回款记录
 */
export const createPayment = (data: CreatePaymentRequest): Promise<APIResponse<Payment>> => {
  return post<APIResponse<Payment>>('/finance/payments', data)
}

/**
 * 更新回款记录
 */
export const updatePayment = (id: number, data: UpdatePaymentRequest): Promise<APIResponse<Payment>> => {
  return put<APIResponse<Payment>>(`/finance/payments/${id}`, data)
}

/**
 * 删除回款记录
 */
export const deletePayment = (id: number): Promise<APIResponse<{ message: string }>> => {
  return del<APIResponse<{ message: string }>>(`/finance/payments/${id}`)
}

/**
 * 确认回款
 */
export const confirmPayment = (id: number): Promise<APIResponse<Payment>> => {
  return post<APIResponse<Payment>>(`/finance/payments/${id}/confirm`, {})
}

/**
 * 拒绝回款
 */
export const rejectPayment = (id: number): Promise<APIResponse<Payment>> => {
  return post<APIResponse<Payment>>(`/finance/payments/${id}/reject`, {})
}

/**
 * 获取回款统计
 */
export const getPaymentStats = (): Promise<APIResponse<PaymentStats>> => {
  return get<APIResponse<PaymentStats>>('/finance/payments/stats')
}

// ============================================================================
// 坏账核销API
// ============================================================================

/**
 * 获取坏账核销列表
 */
export const getBadDebts = (params?: {
  page?: number
  page_size?: number
  status?: string
  contract_id?: number
  reason_type?: string
}): Promise<APIResponse<BadDebt[]>> => {
  return get<APIResponse<BadDebt[]>>('/finance/bad-debts', params)
}

/**
 * 获取待审批的坏账核销
 */
export const getPendingBadDebts = (): Promise<APIResponse<BadDebt[]>> => {
  return get<APIResponse<BadDebt[]>>('/finance/bad-debts/pending')
}

/**
 * 创建坏账核销申请
 */
export const createBadDebt = (data: CreateBadDebtRequest): Promise<APIResponse<BadDebt>> => {
  return post<APIResponse<BadDebt>>('/finance/bad-debts', data)
}

/**
 * 获取坏账核销详情
 */
export const getBadDebt = (id: number): Promise<APIResponse<BadDebt>> => {
  return get<APIResponse<BadDebt>>(`/finance/bad-debts/${id}`)
}

/**
 * 审批通过坏账核销
 */
export const approveBadDebt = (id: number, notes?: string): Promise<APIResponse<BadDebt>> => {
  return post<APIResponse<BadDebt>>(`/finance/bad-debts/${id}/approve`, { notes })
}

/**
 * 审批拒绝坏账核销
 */
export const rejectBadDebt = (id: number, notes: string): Promise<APIResponse<BadDebt>> => {
  return post<APIResponse<BadDebt>>(`/finance/bad-debts/${id}/reject`, { notes })
}

// ============================================================================
// 提成API
// ============================================================================

/**
 * 获取提成列表
 */
export const getCommissions = (params?: {
  page?: number
  page_size?: number
  status?: string
  beneficiary_id?: number
  beneficiary_role?: string
  contract_id?: number
  case_id?: number
  date_from?: string
  date_to?: string
}): Promise<APIResponse<Commission[]>> => {
  return get<APIResponse<Commission[]>>('/finance/commissions', params)
}

/**
 * 计算提成
 */
export const calculateCommissions = (data: CalculateCommissionRequest): Promise<APIResponse<Commission[]>> => {
  return post<APIResponse<Commission[]>>('/finance/commissions/calculate', data)
}

/**
 * 获取提成详情
 */
export const getCommission = (id: number): Promise<APIResponse<Commission>> => {
  return get<APIResponse<Commission>>(`/finance/commissions/${id}`)
}

/**
 * 标记提成已支付
 */
export const markCommissionAsPaid = (id: number, data: {
  paid_date: string
  voucher: string
}): Promise<APIResponse<Commission>> => {
  return post<APIResponse<Commission>>(`/finance/commissions/${id}/mark-paid`, data)
}

/**
 * 取消提成
 */
export const cancelCommission = (id: number): Promise<APIResponse<Commission>> => {
  return post<APIResponse<Commission>>(`/finance/commissions/${id}/cancel`, {})
}

/**
 * 获取受益人的提成记录
 */
export const getCommissionsByBeneficiary = (beneficiaryId: number): Promise<APIResponse<Commission[]>> => {
  return get<APIResponse<Commission[]>>(`/finance/commissions/beneficiary/${beneficiaryId}`)
}

/**
 * 获取提成统计
 */
export const getCommissionStats = (): Promise<APIResponse<CommissionStats>> => {
  return get<APIResponse<CommissionStats>>('/finance/commissions/stats')
}

// ============================================================================
// 财务概览API
// ============================================================================

/**
 * 获取财务概览
 */
export const getFinanceOverview = (): Promise<APIResponse<FinanceOverview>> => {
  return get<APIResponse<FinanceOverview>>('/finance/overview')
}

// ============================================================================
// 辅助函数
// ============================================================================

/**
 * 发票状态文本映射
 */
export const invoiceStatusMap: Record<string, { text: string; color: string }> = {
  draft: { text: '草稿', color: 'default' },
  submitted: { text: '待审批', color: 'processing' },
  approved: { text: '已审批', color: 'success' },
  issued: { text: '已开票', color: 'warning' },
  received: { text: '已签收', color: 'success' },
  cancelled: { text: '已作废', color: 'default' },
}

/**
 * 合同状态文本映射
 */
export const contractStatusMap: Record<string, { text: string; color: string }> = {
  draft: { text: '草稿', color: 'default' },
  active: { text: '生效中', color: 'success' },
  suspended: { text: '已暂停', color: 'warning' },
  completed: { text: '已完成', color: 'default' },
  cancelled: { text: '已取消', color: 'default' },
}

/**
 * 回款状态文本映射
 */
export const paymentStatusMap: Record<string, { text: string; color: string }> = {
  pending: { text: '待确认', color: 'warning' },
  confirmed: { text: '已确认', color: 'success' },
  rejected: { text: '已拒绝', color: 'error' },
}

/**
 * 提成状态文本映射
 */
export const commissionStatusMap: Record<string, { text: string; color: string }> = {
  pending: { text: '待计算', color: 'default' },
  calculated: { text: '已计算', color: 'processing' },
  paid: { text: '已支付', color: 'success' },
  cancelled: { text: '已取消', color: 'default' },
}

/**
 * 格式化金额
 */
export const formatAmount = (amount: number): string => {
  return `¥${amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

/**
 * 格式化日期
 */
export const formatDate = (dateStr?: string): string => {
  if (!dateStr) return '-'
  return dateStr.substring(0, 10)
}

export interface PaymentSummary {
  id: number
  payment_code: string
  amount: number
  payment_date: string
  status: string
}
