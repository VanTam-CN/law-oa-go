import dayjs from 'dayjs'

/**
 * 发票状态计算工具
 */
export const invoiceStatusUtils = {
  /**
   * 计算发票当前状态
   * @param dueDate 到期日期
   * @param paidDate 实际支付日期（可选）
   * @returns 当前状态：'pending' | 'paid' | 'overdue'
   */
  calculateStatus: (dueDate: string, paidDate?: string): 'pending' | 'paid' | 'overdue' => {
    const today = dayjs()
    const due = dayjs(dueDate)
    const paid = paidDate ? dayjs(paidDate) : null

    // 如果已经支付，状态为已付款
    if (paid) {
      return 'paid'
    }

    // 如果今天已过到期日期，状态为逾期
    if (today.isAfter(due)) {
      return 'overdue'
    }

    // 否则为待付款
    return 'pending'
  },

  /**
   * 检查发票是否逾期
   * @param dueDate 到期日期
   * @param paidDate 实际支付日期（可选）
   * @returns 是否逾期
   */
  isOverdue: (dueDate: string, paidDate?: string): boolean => {
    return invoiceStatusUtils.calculateStatus(dueDate, paidDate) === 'overdue'
  },

  /**
   * 检查发票是否已支付
   * @param dueDate 到期日期
   * @param paidDate 实际支付日期（可选）
   * @returns 是否已支付
   */
  isPaid: (dueDate: string, paidDate?: string): boolean => {
    return invoiceStatusUtils.calculateStatus(dueDate, paidDate) === 'paid'
  },

  /**
   * 计算逾期天数
   * @param dueDate 到期日期
   * @param paidDate 实际支付日期（可选）
   * @returns 逾期天数，如果未逾期则返回0
   */
  getOverdueDays: (dueDate: string, paidDate?: string): number => {
    if (!invoiceStatusUtils.isOverdue(dueDate, paidDate)) {
      return 0
    }

    const today = dayjs()
    const due = dayjs(dueDate)

    // 如果已经支付，使用支付日期计算逾期天数
    const comparisonDate = paidDate ? dayjs(paidDate) : today
    return comparisonDate.diff(due, 'day')
  },

  /**
   * 获取状态显示文本和颜色
   * @param status 状态
   * @returns 包含文本和颜色的对象
   */
  getStatusDisplay: (status: 'pending' | 'paid' | 'overdue') => {
    const statusMap = {
      pending: { text: '待付款', color: 'orange' },
      paid: { text: '已付款', color: 'green' },
      overdue: { text: '逾期', color: 'red' },
    }
    return statusMap[status]
  },

  /**
   * 获取状态描述（包含额外信息）
   * @param invoice 发票对象
   * @returns 状态描述文本
   */
  getStatusDescription: (invoice: {
    dueDate: string
    paidDate?: string
    status: 'pending' | 'paid' | 'overdue'
  }): string => {
    const { calculateStatus, getOverdueDays } = invoiceStatusUtils
    const currentStatus = calculateStatus(invoice.dueDate, invoice.paidDate)

    if (currentStatus === 'paid') {
      return `已于 ${dayjs(invoice.paidDate).format('YYYY-MM-DD')} 支付`
    } else if (currentStatus === 'overdue') {
      const overdueDays = getOverdueDays(invoice.dueDate, invoice.paidDate)
      return `逾期 ${overdueDays} 天`
    } else {
      const due = dayjs(invoice.dueDate)
      const today = dayjs()
      const daysUntilDue = due.diff(today, 'day')

      if (daysUntilDue === 0) {
        return '今天到期'
      } else if (daysUntilDue === 1) {
        return '明天到期'
      } else if (daysUntilDue < 0) {
        return '已逾期'
      } else {
        return `${daysUntilDue} 天后到期`
      }
    }
  },
}

/**
 * 费用状态工具
 */
export const expenseStatusUtils = {
  /**
   * 获取费用状态显示文本和颜色
   * @param status 状态
   * @returns 包含文本和颜色的对象
   */
  getStatusDisplay: (status: 'pending' | 'approved' | 'rejected') => {
    const statusMap = {
      pending: { text: '待审批', color: 'orange' },
      approved: { text: '已批准', color: 'green' },
      rejected: { text: '已拒绝', color: 'red' },
    }
    return statusMap[status]
  },

  /**
   * 获取费用状态描述
   * @param expense 费用对象
   * @returns 状态描述文本
   */
  getStatusDescription: (expense: {
    status: 'pending' | 'approved' | 'rejected'
    approveDate?: string
    approver?: string
  }): string => {
    if (expense.status === 'approved' && expense.approveDate && expense.approver) {
      return `于 ${dayjs(expense.approveDate).format('YYYY-MM-DD')} 由 ${expense.approver} 批准`
    } else if (expense.status === 'rejected') {
      return '已拒绝'
    } else {
      return '待审批'
    }
  },
}

/**
 * 财务统计计算工具
 */
export const financeStatsUtils = {
  /**
   * 计算发票统计
   * @param invoices 发票列表
   * @returns 发票统计数据
   */
  calculateInvoiceStats: (
    invoices: Array<{
      amount: number
      dueDate: string
      paidDate?: string
      status: 'pending' | 'paid' | 'overdue'
    }>,
  ) => {
    const stats = {
      total: 0,
      paid: 0,
      pending: 0,
      overdue: 0,
    }

    invoices.forEach((invoice) => {
      stats.total += invoice.amount

      const currentStatus = invoiceStatusUtils.calculateStatus(invoice.dueDate, invoice.paidDate)
      switch (currentStatus) {
        case 'paid':
          stats.paid += invoice.amount
          break
        case 'pending':
          stats.pending += invoice.amount
          break
        case 'overdue':
          stats.overdue += invoice.amount
          break
      }
    })

    return stats
  },

  /**
   * 计算费用统计
   * @param expenses 费用列表
   * @returns 费用统计数据
   */
  calculateExpenseStats: (
    expenses: Array<{
      amount: number
      status: 'pending' | 'approved' | 'rejected'
    }>,
  ) => {
    const stats = {
      total: 0,
      approved: 0,
      pending: 0,
      rejected: 0,
    }

    expenses.forEach((expense) => {
      stats.total += expense.amount

      switch (expense.status) {
        case 'approved':
          stats.approved += expense.amount
          break
        case 'pending':
          stats.pending += expense.amount
          break
        case 'rejected':
          stats.rejected += expense.amount
          break
      }
    })

    return stats
  },

  /**
   * 计算净收入和利润率
   * @param invoiceStats 发票统计
   * @param expenseStats 费用统计
   * @returns 净收入和利润率
   */
  calculateNetIncome: (
    invoiceStats: { total: number; paid: number },
    expenseStats: { approved: number },
  ) => {
    const netIncome = invoiceStats.paid - expenseStats.approved
    const profitMargin = invoiceStats.paid > 0 ? (netIncome / invoiceStats.paid) * 100 : 0

    return {
      netIncome,
      profitMargin,
    }
  },
}

export default {
  invoiceStatus: invoiceStatusUtils,
  expenseStatus: expenseStatusUtils,
  financeStats: financeStatsUtils,
}
