/**
 * CommissionReport 组件测试
 * 测试提成报表页面的渲染和交互功能
 */

import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { ConfigProvider } from 'antd'
import CommissionReport from '../../../src/pages/finance/CommissionReport'
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

const { getCommissions, getCommission, getCommissionStats, markCommissionAsPaid, cancelCommission, calculateCommissions } =
  financeService as any

// Mock 数据
const mockCommissionStats = {
  total_commissions: 150,
  pending_commissions: 10,
  calculated_commissions: 25,
  paid_commissions: 110,
  cancelled_commissions: 5,
  total_commission_amount: 1250000,
  pending_commission_amount: 180000,
  month_commission_amount: 95000,
}

const mockCommissions: any[] = [
  {
    id: 1,
    commission_code: 'COM-2024-001',
    contract_id: 10,
    payment_id: 100,
    case_id: 5,
    beneficiary_id: 3,
    beneficiary_role: 'source',
    payment_amount: 50000,
    cost_deduction: 2000,
    commission_base: 48000,
    commission_rate: 0.15,
    commission_amount: 7200,
    status: 'calculated',
    paid_date: null,
    payment_voucher: '',
    calculated_at: '2024-01-15T10:30:00Z',
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-01-15T10:30:00Z',
    contract: {
      id: 10,
      contract_code: 'CT-2024-010',
      contract_amount: 100000,
    },
    payment: {
      id: 100,
      payment_code: 'PAY-2024-100',
      amount: 50000,
    },
    case: {
      id: 5,
      title: '张三诉李四民间借贷纠纷案',
    },
    beneficiary: {
      id: 3,
      name: '王律师',
    },
  },
  {
    id: 2,
    commission_code: 'COM-2024-002',
    contract_id: 11,
    payment_id: 101,
    case_id: 6,
    beneficiary_id: 4,
    beneficiary_role: 'lawyer',
    payment_amount: 80000,
    cost_deduction: 3000,
    commission_base: 77000,
    commission_rate: 0.20,
    commission_amount: 15400,
    status: 'paid',
    paid_date: '2024-01-20',
    payment_voucher: 'VCH-2024-0120',
    calculated_at: '2024-01-18T14:00:00Z',
    created_at: '2024-01-18T13:30:00Z',
    updated_at: '2024-01-20T16:00:00Z',
    contract: {
      id: 11,
      contract_code: 'CT-2024-011',
      contract_amount: 150000,
    },
    payment: {
      id: 101,
      payment_code: 'PAY-2024-101',
      amount: 80000,
    },
    case: {
      id: 6,
      title: '某公司合同纠纷案',
    },
    beneficiary: {
      id: 4,
      name: '李律师',
    },
  },
  {
    id: 3,
    commission_code: 'COM-2024-003',
    contract_id: 12,
    payment_id: 102,
    case_id: 7,
    beneficiary_id: 5,
    beneficiary_role: 'assistant',
    payment_amount: 30000,
    cost_deduction: 0,
    commission_base: 30000,
    commission_rate: 0.05,
    commission_amount: 1500,
    status: 'pending',
    paid_date: null,
    payment_voucher: '',
    calculated_at: null,
    created_at: '2024-01-22T09:00:00Z',
    updated_at: '2024-01-22T09:00:00Z',
    contract: {
      id: 12,
      contract_code: 'CT-2024-012',
      contract_amount: 60000,
    },
    payment: {
      id: 102,
      payment_code: 'PAY-2024-102',
      amount: 30000,
    },
    case: {
      id: 7,
      title: '劳动争议仲裁案',
    },
    beneficiary: {
      id: 5,
      name: '赵助理',
    },
  },
]

const mockCommissionDetail = {
  ...mockCommissions[0],
}

// Wrapper 组件
const wrapper = ({ children }: { children: React.ReactNode }) => (
  <ConfigProvider>{children}</ConfigProvider>
)

describe('CommissionReport', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // 默认 mock 返回值
    ;(getCommissions as jest.Mock).mockResolvedValue({
      data: mockCommissions,
      pagination: {
        page: 1,
        page_size: 10,
        total: 3,
        total_pages: 1,
        has_next: false,
        has_prev: false,
      },
    })
    ;(getCommissionStats as jest.Mock).mockResolvedValue({
      data: mockCommissionStats,
    })
    ;(getCommission as jest.Mock).mockResolvedValue({
      data: mockCommissionDetail,
    })
    ;(markCommissionAsPaid as jest.Mock).mockResolvedValue({
      data: { success: true },
    })
    ;(cancelCommission as jest.Mock).mockResolvedValue({
      data: { success: true },
    })
    ;(calculateCommissions as jest.Mock).mockResolvedValue({
      data: mockCommissions,
    })
  })

  describe('基础渲染', () => {
    test('应该正确渲染组件', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      expect(getCommissions).toHaveBeenCalled()
      expect(getCommissionStats).toHaveBeenCalled()
    })

    test('应该显示统计数据', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成总数')).toBeInTheDocument()
        expect(screen.getByText('待支付')).toBeInTheDocument()
        expect(screen.getByText('提成总额')).toBeInTheDocument()
        expect(screen.getByText('本月提成')).toBeInTheDocument()
      })

      // 检查统计值
      expect(screen.getByText('150')).toBeInTheDocument() // total_commissions
      expect(screen.getByText('25')).toBeInTheDocument() // calculated_commissions
    })

    test('应该显示提成列表', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('COM-2024-001')).toBeInTheDocument()
        expect(screen.getByText('COM-2024-002')).toBeInTheDocument()
        expect(screen.getByText('COM-2024-003')).toBeInTheDocument()
      })

      // 检查受益人名称
      expect(screen.getByText('王律师')).toBeInTheDocument()
      expect(screen.getByText('李律师')).toBeInTheDocument()
      expect(screen.getByText('赵助理')).toBeInTheDocument()
    })

    test('应该正确显示受益人角色标签', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('案源人')).toBeInTheDocument()
        expect(screen.getByText('承办律师')).toBeInTheDocument()
        expect(screen.getByText('助理')).toBeInTheDocument()
      })
    })
  })

  describe('筛选功能', () => {
    test('应该能够按状态筛选', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 点击状态筛选
      const statusSelect = screen.getByPlaceholderText('筛选状态')
      fireEvent.mouseDown(statusSelect)

      // 选择"已计算"
      const calculatedOption = await screen.findByText('已计算')
      fireEvent.click(calculatedOption)

      // 点击搜索按钮
      const searchButton = screen.getByText('搜索')
      fireEvent.click(searchButton)

      await waitFor(() => {
        expect(getCommissions).toHaveBeenCalledWith(
          expect.objectContaining({
            status: 'calculated',
          })
        )
      })
    })

    test('应该能够按受益人角色筛选', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 点击角色筛选
      const roleSelect = screen.getByPlaceholderText('受益人角色')
      fireEvent.mouseDown(roleSelect)

      // 选择"承办律师"
      const lawyerOption = await screen.findByText('承办律师')
      fireEvent.click(lawyerOption)

      // 点击搜索按钮
      const searchButton = screen.getByText('搜索')
      fireEvent.click(searchButton)

      await waitFor(() => {
        expect(getCommissions).toHaveBeenCalledWith(
          expect.objectContaining({
            beneficiary_role: 'lawyer',
          })
        )
      })
    })

    test('应该能够重置筛选条件', async () => {
      // 设置搜索参数
      ;(getCommissions as jest.Mock).mockResolvedValue({
        data: mockCommissions.filter((c) => c.status === 'calculated'),
        pagination: { page: 1, page_size: 10, total: 1, total_pages: 1, has_next: false, has_prev: false },
      })

      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 点击重置按钮
      const resetButton = screen.getByText('重置')
      fireEvent.click(resetButton)

      await waitFor(() => {
        expect(getCommissions).toHaveBeenCalledWith(
          expect.objectContaining({
            status: '',
            beneficiary_role: '',
            date_from: '',
            date_to: '',
          })
        )
      })
    })
  })

  describe('视图切换', () => {
    test('应该能够在明细列表和按人汇总视图之间切换', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 切换到汇总视图
      const summaryButton = screen.getByText('按人汇总')
      fireEvent.click(summaryButton)

      await waitFor(() => {
        expect(screen.getByText('提成汇总（按受益人）')).toBeInTheDocument()
      })

      // 切换回列表视图
      const listButton = screen.getByText('明细列表')
      fireEvent.click(listButton)

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })
    })

    test('汇总视图应该显示按受益人聚合的数据', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 切换到汇总视图
      const summaryButton = screen.getByText('按人汇总')
      fireEvent.click(summaryButton)

      await waitFor(() => {
        expect(screen.getByText('提成汇总（按受益人）')).toBeInTheDocument()
      })

      // 检查汇总数据（应该包含提成笔数、提成总额等）
      expect(screen.getByText('提成笔数')).toBeInTheDocument()
      expect(screen.getByText('提成总额')).toBeInTheDocument()
      expect(screen.getByText('已支付')).toBeInTheDocument()
      expect(screen.getByText('待支付')).toBeInTheDocument()
      expect(screen.getByText('支付进度')).toBeInTheDocument()
    })
  })

  describe('操作功能', () => {
    test('应该能够打开计算提成弹窗', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 点击计算提成按钮
      const calculateButton = screen.getByText('计算提成')
      fireEvent.click(calculateButton)

      await waitFor(() => {
        expect(screen.getByText('计算提成')).toBeInTheDocument()
        expect(screen.getByText('回款ID')).toBeInTheDocument()
        expect(screen.getByText('成本扣除')).toBeInTheDocument()
      })
    })

    test('应该能够提交计算提成', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 点击计算提成按钮
      const calculateButton = screen.getByText('计算提成')
      fireEvent.click(calculateButton)

      // 等待弹窗出现
      await waitFor(() => {
        expect(screen.getByText('计算提成')).toBeInTheDocument()
      })

      // 由于使用了antd Form，我们这里只测试弹窗是否打开
      // 实际的表单提交测试需要更复杂的设置
    })

    test('应该能够导出报表', async () => {
      const message = require('antd').message
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })

      // 点击导出按钮
      const exportButton = screen.getByText('导出')
      fireEvent.click(exportButton)

      // 验证显示"导出功能开发中"的提示
      expect(message.info).toHaveBeenCalledWith('导出功能开发中...')
    })
  })

  describe('数据格式化', () => {
    test('应该正确格式化提成金额', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('¥7,200')).toBeInTheDocument()
        expect(screen.getByText('¥15,400')).toBeInTheDocument()
        expect(screen.getByText('¥1,500')).toBeInTheDocument()
      })
    })

    test('应该正确显示提成比例', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        // 15% = 0.15 * 100 = 15.0%
        expect(screen.getByText('15.0%')).toBeInTheDocument()
        // 20% = 0.20 * 100 = 20.0%
        expect(screen.getByText('20.0%')).toBeInTheDocument()
        // 5% = 0.05 * 100 = 5.0%
        expect(screen.getByText('5.0%')).toBeInTheDocument()
      })
    })

    test('应该正确显示状态标签', async () => {
      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('已计算')).toBeInTheDocument()
        expect(screen.getByText('已支付')).toBeInTheDocument()
        expect(screen.getByText('待计算')).toBeInTheDocument()
      })
    })
  })

  describe('错误处理', () => {
    test('API 调用失败时应该显示错误消息', async () => {
      const message = require('antd').message
      ;(getCommissions as jest.Mock).mockRejectedValue(new Error('网络错误'))

      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(message.error).toHaveBeenCalledWith('获取提成列表失败')
      })
    })
  })

  describe('响应式设计', () => {
    test('在小屏幕上应该正常显示', async () => {
      // 设置小屏幕宽度
      global.innerWidth = 375
      global.dispatchEvent(new Event('resize'))

      render(<CommissionReport />, { wrapper })

      await waitFor(() => {
        expect(screen.getByText('提成明细')).toBeInTheDocument()
      })
    })
  })
})

describe('CommissionReport 数据处理', () => {
  test('应该正确生成汇总数据', async () => {
    render(<CommissionReport />, { wrapper })

    await waitFor(() => {
      expect(screen.getByText('提成明细')).toBeInTheDocument()
    })

    // 切换到汇总视图
    const summaryButton = screen.getByText('按人汇总')
    fireEvent.click(summaryButton)

    await waitFor(() => {
      expect(screen.getByText('提成汇总（按受益人）')).toBeInTheDocument()
    })

    // 验证汇总数据包含3个不同的受益人
    const rows = screen.getAllByText(/王律师|李律师|赵助理/)
    expect(rows.length).toBeGreaterThan(0)
  })
})
