import { get, post } from './http'

// 发票接口
export interface Invoice {
  id: number
  invoice_number: string
  client_id: number
  client_name: string
  project_id: number
  project_name: string
  amount: number
  status: 'pending' | 'paid' | 'overdue'
  issue_date: string
  due_date: string
  paid_date?: string
  description: string
  created_at: string
}

// 费用接口
export interface Expense {
  id: number
  expense_number: string
  category: string
  description: string
  amount: number
  date: string
  applicant: string
  status: 'pending' | 'approved' | 'rejected'
  approve_date?: string
  approver?: string
  created_at: string
}

// 费用类别接口
export interface ExpenseCategory {
  id: number
  name: string
  code: string
}

// 财务统计接口
export interface FinanceStats {
  invoices: {
    total: number
    paid: number
    pending: number
    overdue: number
  }
  expenses: {
    total: number
    approved: number
    pending: number
  }
  netIncome: number
  profitMargin: number
}

/**
 * 获取发票列表
 * @param params 查询参数
 * @returns 发票列表
 */
export const getInvoices = (params?: {
  status?: string
  clientId?: number
  projectId?: number
}): Promise<Invoice[]> => {
  return get<Invoice[]>('/finance/invoices', params)
}

/**
 * 获取发票详情
 * @param id 发票ID
 * @returns 发票详情
 */
export const getInvoiceById = (id: number): Promise<Invoice> => {
  return get<Invoice>(`/finance/invoices/${id}`)
}

/**
 * 创建发票
 * @param data 发票数据
 * @returns 创建的发票
 */
export const createInvoice = (data: Partial<Invoice>): Promise<Invoice> => {
  return post<Invoice>('/finance/invoices', data)
}

/**
 * 更新发票
 * @param id 发票ID
 * @param data 发票数据
 * @returns 更新的发票
 */
export const updateInvoice = (id: number, data: Partial<Invoice>): Promise<Invoice> => {
  return post<Invoice>(`/finance/invoices/${id}`, data, 'PUT')
}

/**
 * 删除发票
 * @param id 发票ID
 * @returns 删除结果
 */
export const deleteInvoice = (id: number): Promise<void> => {
  return post<void>(`/finance/invoices/${id}`, null, 'DELETE')
}

/**
 * 标记发票为已支付
 * @param id 发票ID
 * @param paymentData 支付数据
 * @returns 更新后的发票
 */
export const markInvoiceAsPaid = (
  id: number,
  paymentData: { paid_date: string },
): Promise<Invoice> => {
  return post<Invoice>(`/finance/invoices/${id}/pay`, paymentData)
}

/**
 * 获取费用列表
 * @param params 查询参数
 * @returns 费用列表
 */
export const getExpenses = (params?: {
  category?: string
  status?: string
  applicant?: string
}): Promise<Expense[]> => {
  return get<Expense[]>('/finance/expenses', params)
}

/**
 * 获取费用详情
 * @param id 费用ID
 * @returns 费用详情
 */
export const getExpenseById = (id: number): Promise<Expense> => {
  return get<Expense>(`/finance/expenses/${id}`)
}

/**
 * 创建费用申请
 * @param data 费用数据
 * @returns 创建的费用
 */
export const createExpense = (data: Partial<Expense>): Promise<Expense> => {
  return post<Expense>('/finance/expenses', data)
}

/**
 * 更新费用
 * @param id 费用ID
 * @param data 费用数据
 * @returns 更新的费用
 */
export const updateExpense = (id: number, data: Partial<Expense>): Promise<Expense> => {
  return post<Expense>(`/finance/expenses/${id}`, data, 'PUT')
}

/**
 * 删除费用
 * @param id 费用ID
 * @returns 删除结果
 */
export const deleteExpense = (id: number): Promise<void> => {
  return post<void>(`/finance/expenses/${id}`, null, 'DELETE')
}

/**
 * 审批费用
 * @param id 费用ID
 * @param action 审批动作
 * @param approvalData 审批数据
 * @returns 更新后的费用
 */
export const handleExpenseApproval = (
  id: number,
  action: 'approve' | 'reject',
  approvalData: { approver: string },
): Promise<Expense> => {
  return post<Expense>(`/finance/expenses/${id}/${action}`, approvalData)
}

/**
 * 获取财务统计
 * @returns 财务统计数据
 */
export const getFinanceStats = (): Promise<FinanceStats> => {
  return get<FinanceStats>('/finance/stats')
}

/**
 * 获取费用类别列表
 * @returns 费用类别列表
 */
export const getExpenseCategories = (): Promise<ExpenseCategory[]> => {
  return get<ExpenseCategory[]>('/finance/expense-categories')
}
