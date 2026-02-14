import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  Row,
  Col,
  Space,
  Tag,
  Badge,
  Tooltip,
  Typography,
  Alert,
  Descriptions,
  Divider,
  InputNumber,
  message,
  Popconfirm,
  Tabs,
  Statistic,
} from 'antd'
import {
  AccountBookOutlined,
  PlusOutlined,
  EyeOutlined,
  LockOutlined,
  UnlockOutlined,
  TransactionOutlined,
  DollarOutlined,
  WalletOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import api from '@/utils/api'

const { Option } = Select
const { Title, Text } = Typography

// 账户状态配置
const accountStatusConfig: Record<
  string,
  { label: string; color: string; icon: React.ReactNode }
> = {
  active: {
    label: '正常',
    color: 'success',
    icon: <CheckCircleOutlined />,
  },
  frozen: {
    label: '已冻结',
    color: 'warning',
    icon: <LockOutlined />,
  },
  closed: {
    label: '已关闭',
    color: 'default',
    icon: <CloseCircleOutlined />,
  },
  pending: {
    label: '待审核',
    color: 'processing',
    icon: <ClockCircleOutlined />,
  },
}

// 交易类型配置
const transactionTypeConfig: Record<
  string,
  { label: string; color: string; icon: React.ReactNode; isIncome: boolean }
> = {
  deposit: {
    label: '存入',
    color: 'green',
    icon: <DollarOutlined />,
    isIncome: true,
  },
  deposit_refund: {
    label: '存入退款',
    color: 'cyan',
    icon: <TransactionOutlined />,
    isIncome: true,
  },
  withdraw: {
    label: '取出',
    color: 'orange',
    icon: <TransactionOutlined />,
    isIncome: false,
  },
  transfer: {
    label: '转账',
    color: 'blue',
    icon: <TransactionOutlined />,
    isIncome: false,
  },
}

// 交易状态配置
const transactionStatusConfig: Record<
  string,
  { label: string; color: string }
> = {
  pending: {
    label: '待审批',
    color: 'processing',
  },
  completed: {
    label: '已完成',
    color: 'success',
  },
  cancelled: {
    label: '已取消',
    color: 'default',
  },
}

// 代管款账户数据类型
interface TrustAccount {
  id: number
  client_id: number
  account_code: string
  balance: number
  frozen_amount: number
  available_balance: number
  currency: string
  purpose_restriction?: string
  authorized_uses: string[]
  status: string
  opened_at?: string
  closed_at?: string
  created_at: string
  updated_at: string
  // 关联数据
  client?: {
    id: number
    name: string
  }
  recent_transactions?: TrustTransaction[]
}

// 交易记录数据类型
interface TrustTransaction {
  id: number
  account_id: number
  transaction_code: string
  transaction_type: string
  amount: number
  description: string
  purpose_code: string
  recipient_name?: string
  recipient_bank_account?: string
  recipient_bank_name?: string
  status: string
  completed_at?: string
  created_by: number
  created_at: string
  updated_at: string
  approved_by?: number
  approved_at?: string
  case_id?: number
}

// 账户统计类型
interface AccountStats {
  total_accounts: number
  total_balance: number
  total_frozen: number
  active_accounts: number
}

// 创建账户请求类型
interface CreateAccountRequest {
  client_id: number
  currency: string
  purpose_restriction?: string
  authorized_uses?: string[]
}

// 创建交易请求类型
interface CreateTransactionRequest {
  account_id: number
  transaction_type: 'deposit' | 'deposit_refund' | 'withdraw' | 'transfer'
  amount: number
  description: string
  case_id?: number
  purpose_code?: string
  recipient_name?: string
  recipient_bank_account?: string
  recipient_bank_name?: string
}

interface TrustAccountListProps {
  userRole?: string
  userId?: string
}

const TrustAccountList: React.FC<TrustAccountListProps> = ({
  userRole = 'admin',
  userId,
}) => {
  const [accounts, setAccounts] = useState<TrustAccount[]>([])
  const [transactions, setTransactions] = useState<TrustTransaction[]>([])
  const [loading, setLoading] = useState(false)
  const [transactionLoading, setTransactionLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [currencyFilter, setCurrencyFilter] = useState<string>('')

  // Modal states
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [depositModalVisible, setDepositModalVisible] = useState(false)
  const [withdrawModalVisible, setWithdrawModalVisible] = useState(false)
  const [freezeModalVisible, setFreezeModalVisible] = useState(false)

  const [currentAccount, setCurrentAccount] = useState<TrustAccount | null>(null)
  const [form] = Form.useForm()
  const [depositForm] = Form.useForm()
  const [withdrawForm] = Form.useForm()
  const [freezeForm] = Form.useForm()

  const [stats, setStats] = useState<AccountStats>({
    total_accounts: 0,
    total_balance: 0,
    total_frozen: 0,
    active_accounts: 0,
  })

  // 客户列表（用于选择）
  const [clients, setClients] = useState<any[]>([])

  // 获取账户列表
  const fetchAccounts = useCallback(async () => {
    setLoading(true)
    try {
      const params: any = {
        page: currentPage,
        page_size: pageSize,
      }
      if (statusFilter) params.status = statusFilter
      if (currencyFilter) params.currency = currencyFilter

      const response = await api.get('/trust/accounts', { params })
      setAccounts(response.data.accounts || [])
      setTotal(response.data.pagination?.total || 0)
    } catch (error: any) {
      message.error(error.response?.data?.message || '获取账户列表失败')
    } finally {
      setLoading(false)
    }
  }, [currentPage, pageSize, statusFilter, currencyFilter])

  // 获取统计数据
  const fetchStats = useCallback(async () => {
    try {
      const response = await api.get('/trust/stats')
      setStats(response.data || {})
    } catch (error) {
      console.error('获取统计数据失败', error)
    }
  }, [])

  // 获取客户列表
  const fetchClients = useCallback(async () => {
    try {
      const response = await api.get('/clients', {
        params: { page: 1, page_size: 100 },
      })
      setClients(response.data.clients || [])
    } catch (error) {
      console.error('获取客户列表失败', error)
    }
  }, [])

  // 获取交易记录
  const fetchTransactions = useCallback(async (accountId: number) => {
    setTransactionLoading(true)
    try {
      const response = await api.get(`/trust/accounts/${accountId}/transactions`)
      setTransactions(response.data.transactions || [])
    } catch (error: any) {
      message.error(error.response?.data?.message || '获取交易记录失败')
    } finally {
      setTransactionLoading(false)
    }
  }, [])

  // 创建账户
  const handleCreate = async (values: CreateAccountRequest) => {
    try {
      await api.post('/trust/accounts', values)
      message.success('账户创建成功')
      setCreateModalVisible(false)
      form.resetFields()
      fetchAccounts()
      fetchStats()
    } catch (error: any) {
      message.error(error.response?.data?.message || '创建账户失败')
    }
  }

  // 存款
  const handleDeposit = async (values: any) => {
    if (!currentAccount) return

    try {
      await api.post('/trust/transactions', {
        account_id: currentAccount.id,
        transaction_type: 'deposit',
        amount: values.amount,
        description: values.description || '存款',
        purpose_code: values.purpose_code,
      })
      message.success('存款申请已提交，等待审批')
      setDepositModalVisible(false)
      depositForm.resetFields()
      fetchAccounts()
      if (detailModalVisible) {
        fetchTransactions(currentAccount.id)
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || '存款失败')
    }
  }

  // 取款
  const handleWithdraw = async (values: any) => {
    if (!currentAccount) return

    try {
      await api.post('/trust/transactions', {
        account_id: currentAccount.id,
        transaction_type: 'withdraw',
        amount: values.amount,
        description: values.description || '取款',
        recipient_name: values.recipient_name,
        recipient_bank_account: values.recipient_bank_account,
        recipient_bank_name: values.recipient_bank_name,
      })
      message.success('取款申请已提交，等待审批')
      setWithdrawModalVisible(false)
      withdrawForm.resetFields()
      fetchAccounts()
      if (detailModalVisible) {
        fetchTransactions(currentAccount.id)
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || '取款失败')
    }
  }

  // 冻结账户
  const handleFreeze = async () => {
    if (!currentAccount) return

    try {
      await api.post(`/trust/accounts/${currentAccount.id}/freeze`)
      message.success('账户已冻结')
      setFreezeModalVisible(false)
      freezeForm.resetFields()
      fetchAccounts()
    } catch (error: any) {
      message.error(error.response?.data?.message || '冻结失败')
    }
  }

  // 解冻账户
  const handleUnfreeze = async (accountId: number) => {
    try {
      await api.post(`/trust/accounts/${accountId}/unfreeze`)
      message.success('账户已解冻')
      fetchAccounts()
    } catch (error: any) {
      message.error(error.response?.data?.message || '解冻失败')
    }
  }

  // 关闭账户
  const handleClose = async (accountId: number) => {
    try {
      await api.post(`/trust/accounts/${accountId}/close`)
      message.success('账户已关闭')
      fetchAccounts()
    } catch (error: any) {
      message.error(error.response?.data?.message || '关闭账户失败')
    }
  }

  // 审批交易
  const handleApproveTransaction = async (transactionId: number) => {
    try {
      await api.post(`/trust/transactions/${transactionId}/approve`)
      message.success('交易已审批通过')
      if (currentAccount) {
        fetchTransactions(currentAccount.id)
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || '审批失败')
    }
  }

  // 拒绝交易
  const handleRejectTransaction = async (transactionId: number) => {
    try {
      await api.post(`/trust/transactions/${transactionId}/reject`)
      message.success('交易已拒绝')
      if (currentAccount) {
        fetchTransactions(currentAccount.id)
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || '拒绝失败')
    }
  }

  // 查看详情
  const handleViewDetail = (record: TrustAccount) => {
    setCurrentAccount(record)
    setDetailModalVisible(true)
    fetchTransactions(record.id)
  }

  useEffect(() => {
    fetchAccounts()
    fetchStats()
    fetchClients()
  }, [fetchAccounts, fetchStats, fetchClients])

  // 账户表格列定义
  const accountColumns: ColumnsType<TrustAccount> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '账户编号',
      dataIndex: 'account_code',
      key: 'account_code',
      width: 150,
      render: (code: string) => (
        <Text code copyable>
          {code}
        </Text>
      ),
    },
    {
      title: '客户',
      dataIndex: 'client',
      key: 'client',
      width: 120,
      render: (client?: { name: string }) => client?.name || '-',
    },
    {
      title: '余额',
      dataIndex: 'balance',
      key: 'balance',
      width: 140,
      render: (balance: number, record: TrustAccount) => (
        <Space direction="vertical" size={0}>
          <Text strong>
            {balance.toFixed(2)} {record.currency}
          </Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            可用: {record.available_balance.toFixed(2)}
          </Text>
        </Space>
      ),
    },
    {
      title: '冻结金额',
      dataIndex: 'frozen_amount',
      key: 'frozen_amount',
      width: 110,
      render: (amount: number, record: TrustAccount) =>
        amount > 0 ? (
          <Text type="warning">
            {amount.toFixed(2)} {record.currency}
          </Text>
        ) : (
          <Text type="secondary">-</Text>
        ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status: string) => {
        const config = accountStatusConfig[status] || {
          label: status,
          color: 'default',
          icon: null,
        }
        return (
          <Tag icon={config.icon} color={config.color}>
            {config.label}
          </Tag>
        )
      },
    },
    {
      title: '开户日期',
      dataIndex: 'opened_at',
      key: 'opened_at',
      width: 110,
      render: (date?: string) => date ? dayjs(date).format('YYYY-MM-DD') : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right' as const,
      render: (_: unknown, record: TrustAccount) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>

          {record.status === 'active' && (
            <>
              <Tooltip title="存款">
                <Button
                  type="link"
                  size="small"
                  icon={<DollarOutlined />}
                  onClick={() => {
                    setCurrentAccount(record)
                    setDepositModalVisible(true)
                  }}
                />
              </Tooltip>
              <Tooltip title="取款">
                <Button
                  type="link"
                  size="small"
                  icon={<TransactionOutlined />}
                  onClick={() => {
                    setCurrentAccount(record)
                    setWithdrawModalVisible(true)
                  }}
                />
              </Tooltip>
              <Tooltip title="冻结">
                <Button
                  type="link"
                  size="small"
                  icon={<LockOutlined />}
                  onClick={() => {
                    setCurrentAccount(record)
                    setFreezeModalVisible(true)
                  }}
                />
              </Tooltip>
            </>
          )}

          {record.status === 'frozen' && (
            <Popconfirm
              title="确认解冻该账户？"
              onConfirm={() => handleUnfreeze(record.id)}
              okText="确认"
              cancelText="取消"
            >
              <Button type="link" size="small" icon={<UnlockOutlined />} />
            </Popconfirm>
          )}

          {record.status === 'active' && record.balance === 0 && (
            <Popconfirm
              title="确认关闭该账户？"
              description="关闭后将无法恢复"
              onConfirm={() => handleClose(record.id)}
              okText="确认"
              cancelText="取消"
            >
              <Button type="link" size="small" danger icon={<CloseCircleOutlined />} />
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  // 交易记录表格列定义
  const transactionColumns: ColumnsType<TrustTransaction> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 70,
    },
    {
      title: '交易编号',
      dataIndex: 'transaction_code',
      key: 'transaction_code',
      width: 140,
      render: (code: string) => (
        <Text code copyable ellipsis>
          {code}
        </Text>
      ),
    },
    {
      title: '交易类型',
      dataIndex: 'transaction_type',
      key: 'transaction_type',
      width: 100,
      render: (type: string) => {
        const config = transactionTypeConfig[type] || {
          label: type,
          color: 'default',
          icon: null,
          isIncome: false,
        }
        return (
          <Tag icon={config.icon} color={config.color}>
            {config.label}
          </Tag>
        )
      },
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 110,
      render: (amount: number, record: TrustTransaction) => {
        const config = transactionTypeConfig[record.transaction_type]
        const isIncome = config?.isIncome ?? false
        return (
          <Text type={isIncome ? 'success' : 'danger'} strong>
            {isIncome ? '+' : '-'}
            {amount.toFixed(2)}
          </Text>
        )
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '收款人',
      dataIndex: 'recipient_name',
      key: 'recipient_name',
      width: 100,
      render: (name?: string) => name || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status: string) => {
        const config = transactionStatusConfig[status] || {
          label: status,
          color: 'default',
        }
        return <Tag color={config.color}>{config.label}</Tag>
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: unknown, record: TrustTransaction) =>
        record.status === 'pending' ? (
          <Space size="small">
            <Popconfirm
              title="确认审批通过该交易？"
              onConfirm={() => handleApproveTransaction(record.id)}
              okText="确认"
              cancelText="取消"
            >
              <Button type="link" size="small">
                通过
              </Button>
            </Popconfirm>
            <Popconfirm
              title="确认拒绝该交易？"
              onConfirm={() => handleRejectTransaction(record.id)}
              okText="确认"
              cancelText="取消"
            >
              <Button type="link" size="small" danger>
                拒绝
              </Button>
            </Popconfirm>
          </Space>
        ) : (
          <Text type="secondary">-</Text>
        ),
    },
  ]

  return (
    <div className="trust-account-list">
      <Card>
        {/* 统计信息 */}
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Card size="small">
              <Statistic
                title="总账户数"
                value={stats.total_accounts}
                prefix={<AccountBookOutlined />}
                valueStyle={{ fontSize: 20 }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card size="small">
              <Statistic
                title="总余额"
                value={stats.total_balance}
                prefix={<WalletOutlined />}
                precision={2}
                valueStyle={{ fontSize: 20 }}
                suffix="CNY"
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card size="small">
              <Statistic
                title="冻结金额"
                value={stats.total_frozen}
                prefix={<LockOutlined />}
                precision={2}
                valueStyle={{ fontSize: 20, color: '#faad14' }}
                suffix="CNY"
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card size="small">
              <Statistic
                title="活跃账户"
                value={stats.active_accounts}
                prefix={<CheckCircleOutlined />}
                valueStyle={{ fontSize: 20, color: '#52c41a' }}
              />
            </Card>
          </Col>
        </Row>

        {/* 筛选栏 */}
        <Space style={{ marginBottom: 16 }} wrap>
          <Select
            placeholder="筛选状态"
            allowClear
            style={{ width: 120 }}
            value={statusFilter || undefined}
            onChange={(value) => {
              setStatusFilter(value || '')
              setCurrentPage(1)
            }}
          >
            <Option value="">全部状态</Option>
            {Object.entries(accountStatusConfig).map(([key, config]) => (
              <Option key={key} value={key}>
                {config.icon} {config.label}
              </Option>
            ))}
          </Select>

          <Select
            placeholder="币种"
            allowClear
            style={{ width: 100 }}
            value={currencyFilter || undefined}
            onChange={(value) => {
              setCurrencyFilter(value || '')
              setCurrentPage(1)
            }}
          >
            <Option value="">全部币种</Option>
            <Option value="CNY">人民币</Option>
            <Option value="USD">美元</Option>
            <Option value="EUR">欧元</Option>
          </Select>

          {userRole === 'admin' && (
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setCreateModalVisible(true)}
            >
              新建账户
            </Button>
          )}
        </Space>

        {/* 账户列表 */}
        <Table
          rowKey="id"
          columns={accountColumns}
          dataSource={accounts}
          loading={loading}
          scroll={{ x: 1200 }}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, size) => {
              setCurrentPage(page)
              setPageSize(size || 10)
            },
          }}
        />
      </Card>

      {/* 创建账户弹窗 */}
      <Modal
        title="新建代管款账户"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false)
          form.resetFields()
        }}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item
            name="client_id"
            label="客户"
            rules={[{ required: true, message: '请选择客户' }]}
          >
            <Select placeholder="选择客户" showSearch filterOption={(input, option: any) =>
              (option?.children ?? '').toLowerCase().includes(input.toLowerCase())
            }>
              {clients.map((client) => (
                <Option key={client.id} value={client.id}>
                  {client.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="currency"
            label="币种"
            rules={[{ required: true, message: '请选择币种' }]}
            initialValue="CNY"
          >
            <Select>
              <Option value="CNY">人民币</Option>
              <Option value="USD">美元</Option>
              <Option value="EUR">欧元</Option>
            </Select>
          </Form.Item>

          <Form.Item name="purpose_restriction" label="用途限制">
            <Input placeholder="请输入用途限制（可选）" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 账户详情弹窗 */}
      <Modal
        title={
          <Space>
            <AccountBookOutlined />
            <span>账户详情</span>
            {currentAccount && (
              <Tag color={accountStatusConfig[currentAccount.status]?.color}>
                {accountStatusConfig[currentAccount.status]?.label}
              </Tag>
            )}
          </Space>
        }
        open={detailModalVisible}
        onCancel={() => {
          setDetailModalVisible(false)
          setCurrentAccount(null)
          setTransactions([])
        }}
        footer={null}
        width={1000}
      >
        {currentAccount && (
          <Tabs defaultActiveKey="detail">
            <Tabs.TabPane tab="基本信息" key="detail">
              <Descriptions column={2} bordered size="small">
                <Descriptions.Item label="账户编号" span={2}>
                  <Text code copyable>
                    {currentAccount.account_code}
                  </Text>
                </Descriptions.Item>
                <Descriptions.Item label="客户">
                  {currentAccount.client?.name || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="币种">
                  {currentAccount.currency}
                </Descriptions.Item>
                <Descriptions.Item label="当前余额">
                  <Text strong style={{ fontSize: 16 }}>
                    {currentAccount.balance.toFixed(2)} {currentAccount.currency}
                  </Text>
                </Descriptions.Item>
                <Descriptions.Item label="冻结金额">
                  <Text type={currentAccount.frozen_amount > 0 ? 'warning' : 'secondary'}>
                    {currentAccount.frozen_amount.toFixed(2)} {currentAccount.currency}
                  </Text>
                </Descriptions.Item>
                <Descriptions.Item label="可用余额">
                  <Text
                    type="success"
                    strong
                    style={{ fontSize: 16 }}
                  >
                    {currentAccount.available_balance.toFixed(2)} {currentAccount.currency}
                  </Text>
                </Descriptions.Item>
                <Descriptions.Item label="账户状态">
                  <Tag
                    icon={accountStatusConfig[currentAccount.status]?.icon}
                    color={accountStatusConfig[currentAccount.status]?.color}
                  >
                    {accountStatusConfig[currentAccount.status]?.label}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="开户日期">
                  {currentAccount.opened_at ? dayjs(currentAccount.opened_at).format('YYYY-MM-DD') : '-'}
                </Descriptions.Item>
                {currentAccount.purpose_restriction && (
                  <Descriptions.Item label="用途限制" span={2}>
                    {currentAccount.purpose_restriction}
                  </Descriptions.Item>
                )}
              </Descriptions>

              {currentAccount.frozen_amount > 0 && (
                <Alert
                  message={`账户中有 ${currentAccount.frozen_amount.toFixed(2)} ${currentAccount.currency} 资金被冻结`}
                  type="warning"
                  showIcon
                  style={{ marginTop: 16 }}
                />
              )}

              {currentAccount.status === 'active' && currentAccount.available_balance > 0 && (
                <Space style={{ marginTop: 16 }}>
                  <Button
                    type="primary"
                    icon={<DollarOutlined />}
                    onClick={() => {
                      setDepositModalVisible(true)
                    }}
                  >
                    存款
                  </Button>
                  <Button
                    icon={<TransactionOutlined />}
                    onClick={() => {
                      setWithdrawModalVisible(true)
                    }}
                  >
                    取款
                  </Button>
                </Space>
              )}
            </Tabs.TabPane>

            <Tabs.TabPane
              tab={
                <Space>
                  <span>交易记录</span>
                  <Badge count={transactions.length} />
                </Space>
              }
              key="transactions"
            >
              <Table
                rowKey="id"
                columns={transactionColumns}
                dataSource={transactions}
                loading={transactionLoading}
                size="small"
                pagination={false}
              />
            </Tabs.TabPane>
          </Tabs>
        )}
      </Modal>

      {/* 存款弹窗 */}
      <Modal
        title="存款"
        open={depositModalVisible}
        onCancel={() => {
          setDepositModalVisible(false)
          depositForm.resetFields()
        }}
        onOk={() => depositForm.submit()}
      >
        <Form form={depositForm} layout="vertical" onFinish={handleDeposit}>
          {currentAccount && (
            <Alert
              message={`当前账户可用余额: ${currentAccount.available_balance.toFixed(2)} ${currentAccount.currency}`}
              type="info"
              style={{ marginBottom: 16 }}
            />
          )}

          <Form.Item
            name="amount"
            label="存款金额"
            rules={[{ required: true, message: '请输入存款金额' }]}
          >
            <InputNumber
              style={{ width: '100%' }}
              min={0.01}
              precision={2}
              placeholder="请输入存款金额"
              addonAfter={currentAccount?.currency || 'CNY'}
            />
          </Form.Item>

          <Form.Item name="purpose_code" label="用途代码">
            <Input placeholder="请输入用途代码（可选）" />
          </Form.Item>

          <Form.Item name="description" label="备注">
            <Input.TextArea rows={3} placeholder="请输入备注信息" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 取款弹窗 */}
      <Modal
        title="取款"
        open={withdrawModalVisible}
        onCancel={() => {
          setWithdrawModalVisible(false)
          withdrawForm.resetFields()
        }}
        onOk={() => withdrawForm.submit()}
      >
        <Form form={withdrawForm} layout="vertical" onFinish={handleWithdraw}>
          {currentAccount && (
            <>
              <Alert
                message={`当前账户可用余额: ${currentAccount.available_balance.toFixed(2)} ${currentAccount.currency}`}
                type="info"
                style={{ marginBottom: 16 }}
              />

              <Form.Item
                name="amount"
                label="取款金额"
                rules={[
                  { required: true, message: '请输入取款金额' },
                  {
                    validator: (_, value) => {
                      if (!value) return Promise.resolve()
                      if (value > currentAccount.available_balance) {
                        return Promise.reject(new Error('取款金额不能超过可用余额'))
                      }
                      return Promise.resolve()
                    },
                  },
                ]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0.01}
                  max={currentAccount.available_balance}
                  precision={2}
                  placeholder="请输入取款金额"
                  addonAfter={currentAccount.currency}
                />
              </Form.Item>

              <Form.Item name="recipient_name" label="收款人姓名">
                <Input placeholder="请输入收款人姓名" />
              </Form.Item>

              <Form.Item name="recipient_bank_account" label="收款账号">
                <Input placeholder="请输入收款账号" />
              </Form.Item>

              <Form.Item name="recipient_bank_name" label="收款银行">
                <Input placeholder="请输入收款银行" />
              </Form.Item>
            </>
          )}

          <Form.Item name="description" label="备注">
            <Input.TextArea rows={3} placeholder="请输入备注信息" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 冻结弹窗 */}
      <Modal
        title="冻结账户"
        open={freezeModalVisible}
        onCancel={() => {
          setFreezeModalVisible(false)
          freezeForm.resetFields()
        }}
        onOk={handleFreeze}
      >
        {currentAccount && (
          <>
            <Alert
              message={
                <>
                  <div>账户余额: {currentAccount.balance.toFixed(2)} {currentAccount.currency}</div>
                  <div>已冻结: {currentAccount.frozen_amount.toFixed(2)} {currentAccount.currency}</div>
                  <div>可用: {currentAccount.available_balance.toFixed(2)} {currentAccount.currency}</div>
                </>
              }
              type="warning"
              style={{ marginBottom: 16 }}
            />
            <p>确认要冻结该账户吗？冻结后账户将无法进行交易操作。</p>
          </>
        )}
      </Modal>
    </div>
  )
}

export default TrustAccountList
