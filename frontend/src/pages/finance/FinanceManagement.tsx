import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Tabs,
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  DatePicker,
  Statistic,
  Row,
  Col,
  Modal,
  message,
  Form,
  Tooltip,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  MoneyCollectOutlined,
  AccountBookOutlined,
  CreditCardOutlined,
  BarChartOutlined,
  CheckOutlined,
  CloseOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  SearchOutlined,
  ReloadOutlined,
  WalletOutlined,
  TransactionOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getInvoices,
  createInvoice,
  updateInvoice,
  deleteInvoice,
  markInvoiceAsPaid,
  getExpenses,
  createExpense,
  updateExpense,
  deleteExpense,
  handleExpenseApproval,
  getFinanceStats,
  getExpenseCategories,
  Invoice,
  Expense,
  FinanceStats,
  ExpenseCategory,
} from '@/services/finance'
import { invoiceStatusUtils, expenseStatusUtils, financeStatsUtils } from '@/utils/financeStatus'
import dayjs from 'dayjs'
import './FinanceManagement.less'

const { Search } = Input
const { Option } = Select
const { RangePicker } = DatePicker

interface Invoice {
  id: string
  clientName: string
  projectName: string
  amount: number
  status: 'pending' | 'paid' | 'overdue'
  issueDate: string
  dueDate: string
  paidDate?: string
}

interface Expense {
  id: string
  category: string
  description: string
  amount: number
  date: string
  status: 'pending' | 'approved' | 'rejected'
  applicant: string
}

const FinanceManagement: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview')
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingItem, setEditingItem] = useState<Invoice | Expense | null>(null)

  // 数据状态
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [expenses, setExpenses] = useState<Expense[]>([])
  const [financeStats, setFinanceStats] = useState<FinanceStats | null>(null)
  const [expenseCategories, setExpenseCategories] = useState<ExpenseCategory[]>([])

  // 计算发票状态和统计
  const calculateInvoiceData = useCallback(() => {
    const calculatedInvoices = invoices.map((invoice) => ({
      ...invoice,
      calculatedStatus: invoiceStatusUtils.calculateStatus(invoice.due_date, invoice.paid_date),
      statusDescription: invoiceStatusUtils.getStatusDescription({
        dueDate: invoice.due_date,
        paidDate: invoice.paid_date,
        status: invoice.status,
      }),
      overdueDays: invoiceStatusUtils.getOverdueDays(invoice.due_date, invoice.paid_date),
    }))

    const invoiceStats = financeStatsUtils.calculateInvoiceStats(calculatedInvoices)

    return { calculatedInvoices, invoiceStats }
  }, [invoices])

  // 计算费用统计
  const calculateExpenseData = useCallback(() => {
    const calculatedExpenses = expenses.map((expense) => ({
      ...expense,
      statusDescription: expenseStatusUtils.getStatusDescription({
        status: expense.status,
        approveDate: expense.approve_date,
        approver: expense.approver,
      }),
    }))

    const expenseStats = financeStatsUtils.calculateExpenseStats(calculatedExpenses)

    return { calculatedExpenses, expenseStats }
  }, [expenses])

  // 获取数据
  useEffect(() => {
    fetchData()
  }, [])

  // 设置定时器自动更新状态
  useEffect(() => {
    const interval = setInterval(() => {
      // 重新计算状态
      setInvoices((prev) => [...prev])
      setExpenses((prev) => [...prev])
    }, 60000) // 每分钟更新一次

    return () => clearInterval(interval)
  }, [])

  const fetchData = async () => {
    try {
      setLoading(true)
      const [invoicesData, expensesData, statsData, categoriesData] = await Promise.all([
        getInvoices(),
        getExpenses(),
        getFinanceStats(),
        getExpenseCategories(),
      ])
      setInvoices(invoicesData)
      setExpenses(expensesData)
      setFinanceStats(statsData)
      setExpenseCategories(categoriesData)
    } catch (error) {
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  const invoiceColumns: ColumnsType<Invoice> = [
    {
      title: '发票号',
      dataIndex: 'invoice_number',
      key: 'invoice_number',
      width: 120,
    },
    {
      title: '客户名称',
      dataIndex: 'client_name',
      key: 'client_name',
      width: 120,
    },
    {
      title: '项目名称',
      dataIndex: 'project_name',
      key: 'project_name',
      width: 150,
      ellipsis: true,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number) => `¥${amount.toLocaleString()}`,
    },
    {
      title: '状态',
      key: 'status',
      width: 140,
      render: (_, record: Invoice) => {
        const status =
          record.calculatedStatus ||
          invoiceStatusUtils.calculateStatus(record.due_date, record.paid_date)
        const statusDisplay = invoiceStatusUtils.getStatusDisplay(status)
        const overdueDays = invoiceStatusUtils.getOverdueDays(record.due_date, record.paid_date)

        return (
          <Space direction='vertical' size='small' style={{ width: '100%' }}>
            <Tag color={statusDisplay.color}>{statusDisplay.text}</Tag>
            {status === 'overdue' && overdueDays > 0 && (
              <Tooltip title={`已逾期 ${overdueDays} 天`}>
                <Tag color='red' icon={<ExclamationCircleOutlined />}>
                  逾期 {overdueDays} 天
                </Tag>
              </Tooltip>
            )}
            {status === 'pending' && (
              <Tooltip title={record.statusDescription}>
                <Tag color='orange' icon={<ClockCircleOutlined />}>
                  {record.statusDescription}
                </Tag>
              </Tooltip>
            )}
            {status === 'paid' && record.paid_date && (
              <Tooltip title={`支付日期: ${dayjs(record.paid_date).format('YYYY-MM-DD')}`}>
                <Tag color='green' icon={<CheckOutlined />}>
                  已支付
                </Tag>
              </Tooltip>
            )}
          </Space>
        )
      },
    },
    {
      title: '开票日期',
      dataIndex: 'issue_date',
      key: 'issue_date',
      width: 120,
    },
    {
      title: '到期日期',
      dataIndex: 'due_date',
      key: 'due_date',
      width: 120,
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space size='small'>
          <Button type='link' size='small' icon={<EditOutlined />}>
            编辑
          </Button>
          <Button type='link' size='small' danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Space>
      ),
    },
  ]

  const expenseColumns: ColumnsType<Expense> = [
    {
      title: '费用编号',
      dataIndex: 'expense_number',
      key: 'expense_number',
      width: 120,
    },
    {
      title: '类别',
      dataIndex: 'category',
      key: 'category',
      width: 100,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      width: 200,
      ellipsis: true,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number) => `¥${amount.toLocaleString()}`,
    },
    {
      title: '申请人',
      dataIndex: 'applicant',
      key: 'applicant',
      width: 100,
    },
    {
      title: '申请日期',
      dataIndex: 'date',
      key: 'date',
      width: 120,
    },
    {
      title: '状态',
      key: 'status',
      width: 140,
      render: (_, record: Expense) => {
        const statusDisplay = expenseStatusUtils.getStatusDisplay(record.status)

        return (
          <Space direction='vertical' size='small' style={{ width: '100%' }}>
            <Tag color={statusDisplay.color}>{statusDisplay.text}</Tag>
            {record.statusDescription && (
              <Tooltip title={record.statusDescription}>
                <Tag color='blue' style={{ fontSize: '11px' }}>
                  {record.statusDescription.length > 15
                    ? `${record.statusDescription.substring(0, 15)}...`
                    : record.statusDescription}
                </Tag>
              </Tooltip>
            )}
          </Space>
        )
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space size='small'>
          <Button type='link' size='small' icon={<EditOutlined />}>
            编辑
          </Button>
          <Button type='link' size='small' danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Space>
      ),
    },
  ]

  // 计算统计数据
  const { calculatedInvoices, invoiceStats } = calculateInvoiceData()
  const { expenseStats } = calculateExpenseData()

  // 使用计算后的统计数据
  const totalRevenue = invoiceStats.paid || 0
  const pendingRevenue = invoiceStats.pending || 0
  const overdueRevenue = invoiceStats.overdue || 0
  const totalExpenses = expenseStats.approved || 0

  const tabItems = [
    {
      key: 'overview',
      label: (
        <span>
          <BarChartOutlined />
          财务概览
        </span>
      ),
      children: (
        <div>
          <Row gutter={16} style={{ marginBottom: 24 }}>
            <Col span={6}>
              <Card>
                <Statistic
                  title='已收入金额'
                  value={totalRevenue}
                  precision={0}
                  valueStyle={{ color: '#3f8600' }}
                  prefix='¥'
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title='待收入金额'
                  value={pendingRevenue}
                  precision={0}
                  valueStyle={{ color: '#cf1322' }}
                  prefix='¥'
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title='逾期金额'
                  value={overdueRevenue}
                  precision={0}
                  valueStyle={{ color: '#ff4d4f' }}
                  prefix='¥'
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title='总支出'
                  value={totalExpenses}
                  precision={0}
                  valueStyle={{ color: '#1890ff' }}
                  prefix='¥'
                />
              </Card>
            </Col>
          </Row>

          <Card title='财务趋势图' style={{ marginBottom: 16 }}>
            <div style={{ padding: '60px 0', textAlign: 'center', color: '#999' }}>
              <BarChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
              <p>财务图表功能开发中...</p>
              <p>将展示收入支出趋势、项目盈利分析等</p>
            </div>
          </Card>
        </div>
      ),
    },
    {
      key: 'invoices',
      label: (
        <span>
          <MoneyCollectOutlined />
          发票管理
        </span>
      ),
      children: (
        <Card>
          <div
            style={{
              marginBottom: 16,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <Space>
              <Search placeholder='搜索客户或项目' allowClear style={{ width: 200 }} />
              <Select placeholder='状态' style={{ width: 120 }} allowClear>
                <Option value='pending'>待付款</Option>
                <Option value='paid'>已付款</Option>
                <Option value='overdue'>逾期</Option>
              </Select>
              <RangePicker placeholder={['开始日期', '结束日期']} />
            </Space>
            <Button type='primary' icon={<PlusOutlined />}>
              新建发票
            </Button>
          </div>
          <Table
            columns={invoiceColumns}
            dataSource={calculatedInvoices}
            rowKey='id'
            pagination={{
              total: calculatedInvoices.length,
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条记录`,
            }}
          />
        </Card>
      ),
    },
    {
      key: 'expenses',
      label: (
        <span>
          <CreditCardOutlined />
          费用管理
        </span>
      ),
      children: (
        <Card>
          <div
            style={{
              marginBottom: 16,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <Space>
              <Search placeholder='搜索费用描述' allowClear style={{ width: 200 }} />
              <Select placeholder='类别' style={{ width: 120 }} allowClear>
                {expenseCategories.map((cat) => (
                  <Option key={cat.id} value={cat.name}>
                    {cat.name}
                  </Option>
                ))}
              </Select>
              <Select placeholder='状态' style={{ width: 120 }} allowClear>
                <Option value='pending'>待审批</Option>
                <Option value='approved'>已批准</Option>
                <Option value='rejected'>已拒绝</Option>
              </Select>
            </Space>
            <Button type='primary' icon={<PlusOutlined />}>
              申请费用
            </Button>
          </div>
          <Table
            columns={expenseColumns}
            dataSource={calculateExpenseData().calculatedExpenses}
            rowKey='id'
            pagination={{
              total: calculateExpenseData().calculatedExpenses.length,
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条记录`,
            }}
          />
        </Card>
      ),
    },
    {
      key: 'reports',
      label: (
        <span>
          <AccountBookOutlined />
          财务报表
        </span>
      ),
      children: (
        <Card>
          <div style={{ padding: '40px 0', textAlign: 'center', color: '#999' }}>
            <AccountBookOutlined style={{ fontSize: 48, marginBottom: 16 }} />
            <p>财务报表功能开发中...</p>
            <p>将包含：月度报表、年度报表、项目盈利分析等</p>
          </div>
        </Card>
      ),
    },
  ]

  return (
    <div className='finance-management'>
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} size='large' />
    </div>
  )
}

export default FinanceManagement
