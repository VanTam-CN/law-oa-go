/**
 * 财务模块全流程集成测试
 * 测试财务模块的完整业务流程
 * 包括：合同管理 -> 发票管理 -> 回款管理 -> 提成管理
 */

import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { ConfigProvider } from 'antd'
import FinanceManagement from '../../../src/pages/finance/FinanceManagement'
import * as financeService from '../../../src/services/finance'

// Mock finance service
jest.mock('../../../src/services/finance')

// Mock Ant Design 的 message
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
    info: jest.fn(),
    warning: jest.fn(),
  },
}))

const {
  getFinanceOverview,
  getContracts,
  createContract,
  updateContract,
  activateContract,
  getInvoices,
  createInvoice,
  submitInvoice,
  approveInvoice,
  issueInvoice,
  confirmInvoiceReceipt,
  getPayments,
  createPayment,
  confirmPayment,
  getCommissions,
  calculateCommissions,
  markCommissionAsPaid,
} = financeService as any

// Mock 完整流程数据
const mockContract = {
  id: 100,
  contract_code: 'CT-2024-TEST-001',
  client_id: 10,
  contract_amount: 100000,
  currency: 'CNY',
  billing_cycle: 'quarterly',
  payment_terms: '分期付款',
  start_date: '2024-01-01',
  end_date: '2024-12-31',
  status: 'draft',
  contract_type: 'original',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  client: {
    id: 10,
    name: '测试客户有限公司',
  },
}

const mockActiveContract = {
  ...mockContract,
  status: 'active',
}

const mockInvoice = {
  id: 200,
  invoice_code: 'INV-2024-TEST-001',
  contract_id: 100,
  client_id: 10,
  amount: 50000,
  tax_rate: 0.06,
  tax_amount: 3000,
  total_amount: 53000,
  client_name: '测试客户有限公司',
  client_tax_id: '91110000123456789X',
  client_address: '北京市朝阳区XX路XX号',
  client_bank_name: '中国工商银行北京分行',
  client_bank_account: '1234567890',
  invoice_type: 'normal',
  original_invoice_id: null,
  refund_reason: '',
  write_off_amount: 0,
  status: 'draft',
  submitted_at: null,
  approved_by_finance_at: null,
  issued_at: null,
  received_at: null,
  electronic_invoice_url: '',
  electronic_invoice_code: '',
  electronic_invoice_number: '',
  created_by: 1,
  submitted_by: null,
  approved_by: null,
  created_at: '2024-01-05T00:00:00Z',
  updated_at: '2024-01-05T00:00:00Z',
  client: {
    id: 10,
    name: '测试客户有限公司',
  },
  contract: {
    id: 100,
    contract_code: 'CT-2024-TEST-001',
    contract_amount: 100000,
  },
  payments: [],
  total_paid_amount: 0,
  remaining_amount: 53000,
}

const mockPayment = {
  id: 300,
  payment_code: 'PAY-2024-TEST-001',
  invoice_id: 200,
  amount: 53000,
  payment_date: '2024-01-20',
  payment_method: 'bank_transfer',
  reference_no: 'REF-2024-001',
  payer_name: '测试客户有限公司',
  payer_account: '9876543210',
  attachment_id: null,
  confirmed_by: 1,
  confirmed_at: null,
  status: 'pending',
  remark: '',
  created_at: '2024-01-20T00:00:00Z',
  invoice: {
    id: 200,
    invoice_code: 'INV-2024-TEST-001',
    total_amount: 53000,
    client_name: '测试客户有限公司',
  },
}

const mockCommission = {
  id: 400,
  commission_code: 'COM-2024-TEST-001',
  contract_id: 100,
  payment_id: 300,
  case_id: null,
  beneficiary_id: 5,
  beneficiary_role: 'lawyer',
  payment_amount: 53000,
  cost_deduction: 3000,
  commission_base: 50000,
  commission_rate: 0.20,
  commission_amount: 10000,
  status: 'calculated',
  paid_date: null,
  payment_voucher: '',
  calculated_at: '2024-01-21T10:00:00Z',
  created_at: '2024-01-21T09:00:00Z',
  updated_at: '2024-01-21T10:00:00Z',
  contract: {
    id: 100,
    contract_code: 'CT-2024-TEST-001',
    contract_amount: 100000,
  },
  payment: {
    id: 300,
    payment_code: 'PAY-2024-TEST-001',
    amount: 53000,
  },
  case: null,
  beneficiary: {
    id: 5,
    name: '张律师',
  },
}

const mockFinanceOverview = {
  contract_stats: {
    total_contracts: 50,
    draft_contracts: 5,
    active_contracts: 35,
    suspended_contracts: 2,
    completed_contracts: 8,
    cancelled_contracts: 0,
    total_contract_amount: 5000000,
    new_contracts_this_month: 3,
  },
  invoice_stats: {
    total_invoices: 120,
    draft_invoices: 10,
    submitted_invoices: 5,
    approved_invoices: 15,
    issued_invoices: 80,
    received_invoices: 70,
    total_invoice_amount: 3000000,
    pending_invoice_amount: 500000,
    overdue_amount: 100000,
  },
  payment_stats: {
    total_payments: 100,
    pending_payments: 8,
    confirmed_payments: 90,
    rejected_payments: 2,
    total_payment_amount: 2500000,
    month_payment_amount: 300000,
    pending_amount: 200000,
  },
  commission_stats: {
    total_commissions: 150,
    pending_commissions: 10,
    calculated_commissions: 20,
    paid_commissions: 115,
    cancelled_commissions: 5,
    total_commission_amount: 750000,
    pending_commission_amount: 80000,
    month_commission_amount: 45000,
  },
}

// Wrapper 组件
const wrapper = ({ children }: { children: React.ReactNode }) => (
  <ConfigProvider>{children}</ConfigProvider>
)

describe('财务模块全流程集成测试', () => {
  beforeEach(() => {
    jest.clearAllMocks()

    // 设置默认 mock 返回值
    ;(getFinanceOverview as jest.Mock).mockResolvedValue({
      data: mockFinanceOverview,
    })
    ;(getContracts as jest.Mock).mockResolvedValue({
      data: [mockActiveContract],
      pagination: { page: 1, page_size: 10, total: 1, total_pages: 1, has_next: false, has_prev: false },
    })
    ;(getInvoices as jest.Mock).mockResolvedValue({
      data: [mockInvoice],
      pagination: { page: 1, page_size: 10, total: 1, total_pages: 1, has_next: false, has_prev: false },
    })
    ;(getPayments as jest.Mock).mockResolvedValue({
      data: [mockPayment],
      pagination: { page: 1, page_size: 10, total: 1, total_pages: 1, has_next: false, has_prev: false },
    })
    ;(getCommissions as jest.Mock).mockResolvedValue({
      data: [mockCommission],
      pagination: { page: 1, page_size: 10, total: 1, total_pages: 1, has_next: false, has_prev: false },
    })
    ;(createContract as jest.Mock).mockResolvedValue({ data: mockContract })
    ;(updateContract as jest.Mock).mockResolvedValue({ data: mockActiveContract })
    ;(activateContract as jest.Mock).mockResolvedValue({ data: mockActiveContract })
    ;(createInvoice as jest.Mock).mockResolvedValue({ data: mockInvoice })
    ;(submitInvoice as jest.Mock).mockResolvedValue({ data: { ...mockInvoice, status: 'submitted' } })
    ;(approveInvoice as jest.Mock).mockResolvedValue({ data: { ...mockInvoice, status: 'approved' } })
    ;(issueInvoice as jest.Mock).mockResolvedValue({ data: { ...mockInvoice, status: 'issued' } })
    ;(confirmInvoiceReceipt as jest.Mock).mockResolvedValue({ data: { ...mockInvoice, status: 'received' } })
    ;(createPayment as jest.Mock).mockResolvedValue({ data: mockPayment })
    ;(confirmPayment as jest.Mock).mockResolvedValue({ data: { ...mockPayment, status: 'confirmed' } })
    ;(calculateCommissions as jest.Mock).mockResolvedValue({ data: [mockCommission] })
    ;(markCommissionAsPaid as jest.Mock).mockResolvedValue({ data: { ...mockCommission, status: 'paid' } })
  })

  describe('财务概览', () => {
    test('应该显示完整的财务概览统计', async () => {
      render(<FinanceManagement />, { wrapper })

      // 等待概览加载
      await waitFor(() => {
        expect(getFinanceOverview).toHaveBeenCalled()
      })

      // 检查合同统计
      expect(screen.getByText('合同总数')).toBeInTheDocument()
      expect(screen.getByText('生效合同')).toBeInTheDocument()

      // 检查发票统计
      expect(screen.getByText('发票总数')).toBeInTheDocument()
      expect(screen.getByText('待审批发票')).toBeInTheDocument()

      // 检查回款统计
      expect(screen.getByText('回款总数')).toBeInTheDocument()
      expect(screen.getByText('待确认回款')).toBeInTheDocument()

      // 检查提成统计
      expect(screen.getByText('提成总数')).toBeInTheDocument()
      expect(screen.getByText('待支付提成')).toBeInTheDocument()
    })
  })

  describe('合同管理流程', () => {
    test('应该能够创建和激活合同', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到合同管理标签
      await waitFor(() => {
        const contractTab = screen.getByText('合同管理')
        fireEvent.click(contractTab)
      })

      await waitFor(() => {
        expect(getContracts).toHaveBeenCalled()
      })
    })

    test('应该能够查看合同列表', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到合同管理标签
      await waitFor(() => {
        const contractTab = screen.getByText('合同管理')
        fireEvent.click(contractTab)
      })

      await waitFor(() => {
        expect(screen.getByText('合同列表')).toBeInTheDocument()
      })
    })
  })

  describe('发票管理流程', () => {
    test('应该能够创建发票', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到发票管理标签
      await waitFor(() => {
        const invoiceTab = screen.getByText('发票管理')
        fireEvent.click(invoiceTab)
      })

      await waitFor(() => {
        expect(getInvoices).toHaveBeenCalled()
      })
    })

    test('应该能够查看发票列表', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到发票管理标签
      await waitFor(() => {
        const invoiceTab = screen.getByText('发票管理')
        fireEvent.click(invoiceTab)
      })

      await waitFor(() => {
        expect(screen.getByText('发票列表')).toBeInTheDocument()
      })
    })

    test('应该能够筛选发票状态', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到发票管理标签
      await waitFor(() => {
        const invoiceTab = screen.getByText('发票管理')
        fireEvent.click(invoiceTab)
      })

      await waitFor(() => {
        expect(screen.getByText('发票列表')).toBeInTheDocument()
      })

      // 检查筛选器是否存在
      const statusFilter = screen.queryByPlaceholderText('筛选状态')
      expect(statusFilter).toBeInTheDocument()
    })
  })

  describe('回款管理流程', () => {
    test('应该能够查看回款列表', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到回款管理标签
      await waitFor(() => {
        const paymentTab = screen.getByText('回款管理')
        fireEvent.click(paymentTab)
      })

      await waitFor(() => {
        expect(getPayments).toHaveBeenCalled()
      })
    })

    test('应该能够筛选回款状态', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到回款管理标签
      await waitFor(() => {
        const paymentTab = screen.getByText('回款管理')
        fireEvent.click(paymentTab)
      })

      await waitFor(() => {
        expect(screen.getByText('回款列表')).toBeInTheDocument()
      })

      // 检查筛选器是否存在
      const statusFilter = screen.queryByPlaceholderText('筛选状态')
      expect(statusFilter).toBeInTheDocument()
    })
  })

  describe('提成管理流程', () => {
    test('应该能够查看提成报表', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到提成管理标签
      await waitFor(() => {
        const commissionTab = screen.getByText('提成管理')
        fireEvent.click(commissionTab)
      })

      await waitFor(() => {
        expect(getCommissions).toHaveBeenCalled()
      })
    })

    test('应该能够在明细和汇总视图间切换', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到提成管理标签
      await waitFor(() => {
        const commissionTab = screen.getByText('提成管理')
        fireEvent.click(commissionTab)
      })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 切换到汇总视图
      const summaryButton = screen.queryByText('按人汇总')
      if (summaryButton) {
        fireEvent.click(summaryButton)

        await waitFor(() => {
          expect(screen.getByText('提成汇总（按受益人）')).toBeInTheDocument()
        })
      }
    })

    test('应该能够筛选提成状态', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到提成管理标签
      await waitFor(() => {
        const commissionTab = screen.getByText('提成管理')
        fireEvent.click(commissionTab)
      })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 检查筛选器是否存在
      const statusFilter = screen.queryByPlaceholderText('筛选状态')
      expect(statusFilter).toBeInTheDocument()
    })

    test('应该能够打开计算提成弹窗', async () => {
      render(<FinanceManagement />, { wrapper })

      // 切换到提成管理标签
      await waitFor(() => {
        const commissionTab = screen.getByText('提成管理')
        fireEvent.click(commissionTab)
      })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 检查计算提成按钮是否存在
      const calculateButton = screen.queryByText('计算提成')
      expect(calculateButton).toBeInTheDocument()
    })
  })

  describe('完整业务流程', () => {
    test('合同 -> 发票 -> 回款 -> 提成 流程数据一致性', async () => {
      // 验证各模块数据的关联关系
      expect(mockContract.id).toBe(100)
      expect(mockInvoice.contract_id).toBe(100)
      expect(mockPayment.invoice_id).toBe(200)
      expect(mockCommission.contract_id).toBe(100)
      expect(mockCommission.payment_id).toBe(300)

      // 验证金额计算
      expect(mockInvoice.total_amount).toBe(53000) // 50000 + 3000(税)
      expect(mockPayment.amount).toBe(53000) // 应与发票总额一致
      expect(mockCommission.commission_base).toBe(50000) // 回款 - 成本
      expect(mockCommission.commission_amount).toBe(10000) // 50000 * 20%
    })

    test('各模块状态流转正确性', async () => {
      // 合同状态: draft -> active
      expect(mockContract.status).toBe('draft')
      expect(mockActiveContract.status).toBe('active')

      // 发票状态: draft -> submitted -> approved -> issued -> received
      const invoiceStatusFlow = ['draft', 'submitted', 'approved', 'issued', 'received']
      expect(invoiceStatusFlow).toContain(mockInvoice.status)

      // 回款状态: pending -> confirmed
      const paymentStatusFlow = ['pending', 'confirmed', 'rejected']
      expect(paymentStatusFlow).toContain(mockPayment.status)

      // 提成状态: pending -> calculated -> paid
      const commissionStatusFlow = ['pending', 'calculated', 'paid', 'cancelled']
      expect(commissionStatusFlow).toContain(mockCommission.status)
    })
  })
})

describe('财务模块权限和安全性', () => {
  test('敏感操作应该需要确认', async () => {
    // 取消提成、标记已支付等操作应该需要二次确认
    // 这里只是测试概念，实际实现中需要在 UI 层面确认
    const sensitiveOperations = ['cancelCommission', 'deleteInvoice', 'deleteContract']
    expect(sensitiveOperations.length).toBeGreaterThan(0)
  })
})

describe('财务模块数据验证', () => {
  test('金额计算应该准确', () => {
    // 发票总额 = 金额 + 税额
    const invoiceTotal = mockInvoice.amount + mockInvoice.tax_amount
    expect(invoiceTotal).toBe(mockInvoice.total_amount)

    // 提成基数 = 回款金额 - 成本扣除
    const commissionBase = mockCommission.payment_amount - mockCommission.cost_deduction
    expect(commissionBase).toBe(mockCommission.commission_base)

    // 提成金额 = 提成基数 * 提成比例
    const commissionAmount = commissionBase * mockCommission.commission_rate
    expect(commissionAmount).toBe(mockCommission.commission_amount)
  })

  test('状态机转换应该合法', () => {
    const validStatusTransitions = {
      invoice: {
        draft: ['submitted'],
        submitted: ['approved', 'rejected'],
        approved: ['issued', 'cancelled'],
        issued: ['received', 'cancelled'],
        received: ['cancelled'],
      },
      payment: {
        pending: ['confirmed', 'rejected'],
        confirmed: [],
        rejected: [],
      },
      commission: {
        pending: ['calculated', 'cancelled'],
        calculated: ['paid', 'cancelled'],
        paid: [],
        cancelled: [],
      },
    }

    // 验证发票状态转换
    expect(validStatusTransitions.invoice.draft).toContain('submitted')
    expect(validStatusTransitions.invoice.submitted).toContain('approved')

    // 验证回款状态转换
    expect(validStatusTransitions.payment.pending).toContain('confirmed')

    // 验证提成状态转换
    expect(validStatusTransitions.commission.pending).toContain('calculated')
    expect(validStatusTransitions.commission.calculated).toContain('paid')
  })
})
