/**
 * 代管款账户管理页面
 *
 * 功能：
 * - 账户列表展示与筛选
 * - 账户详情查看
 * - 账户操作（冻结/解冻/关闭）
 * - 交易记录查看
 * - 统计数据展示
 */

import React, { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  message,
  Modal,
  Descriptions,
  Tabs,
  Form,
  Input,
  Select,
  Row,
  Col,
  Statistic,
  Tooltip,
  Badge,
  Popconfirm,
} from 'antd'
import {
  PlusOutlined,
  EyeOutlined,
  LockOutlined,
  UnlockOutlined,
  StopOutlined,
  ReloadOutlined,
  SearchOutlined,
  FilterOutlined,
  WalletOutlined,
  TransactionOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import dayjs from 'dayjs'
import {
  getTrustAccounts,
  createTrustAccount,
  freezeTrustAccount,
  unfreezeTrustAccount,
  closeTrustAccount,
  getAccountTransactions,
  getAccountStats,
  type TrustAccount,
  type TrustTransaction,
  type AccountStatus,
  type Currency,
  accountStatusMap,
  transactionTypeMap,
  transactionStatusMap,
  formatAmount,
  formatDate,
} from '@/services/trust'
import { getClients } from '@/services/client'
import type { Client } from '@/services/client'
import './TrustAccountManagement.less'

const { Option } = Select

const TrustAccountManagement: React.FC = () => {
  const [createForm] = Form.useForm()
  // 状态管理
  const [loading, setLoading] = useState(false)
  const [accounts, setAccounts] = useState<TrustAccount[]>([])
  const [stats, setStats] = useState<any>(null)
  const [clients, setClients] = useState<Client[]>([])

  // 分页和筛选
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })
  const [filters, setFilters] = useState<{
    clientId?: number
    status?: AccountStatus
    currency?: Currency
    search?: string
  }>({})

  // 详情弹窗
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [selectedAccount, setSelectedAccount] = useState<TrustAccount | null>(null)
  const [accountTransactions, setAccountTransactions] = useState<TrustTransaction[]>([])
  const [transactionsLoading, setTransactionsLoading] = useState(false)
  const [transactionsPagination, setTransactionsPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })

  // 创建账户弹窗
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [createSubmitting, setCreateSubmitting] = useState(false)

  // 加载客户列表
  const loadClients = async () => {
    try {
      const res = await getClients({ pageNum: 1, pageSize: 100 })
      if (res?.list) {
        setClients(res.list)
      }
    } catch (error) {
      console.error('加载客户列表失败:', error)
    }
  }

  // 加载账户列表
  const loadAccounts = async (page = 1, pageSize = 10) => {
    setLoading(true)
    try {
      const params: any = {
        page,
        page_size: pageSize,
      }
      if (filters.clientId) params.client_id = filters.clientId
      if (filters.status) params.status = filters.status
      if (filters.currency) params.currency = filters.currency
      if (filters.search) params.search = filters.search

      const res = await getTrustAccounts(params)
      if (res?.data) {
        setAccounts(res.data.accounts || [])
        setPagination({
          current: page,
          pageSize,
          total: res.data.pagination?.total || 0,
        })
      }
    } catch (error) {
      message.error('加载账户列表失败')
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  // 加载统计数据
  const loadStats = async () => {
    try {
      const res = await getAccountStats()
      if (res?.data) {
        setStats(res.data)
      }
    } catch (error) {
      console.error('加载统计数据失败:', error)
    }
  }

  // 加载账户交易记录
  const loadAccountTransactions = async (accountId: number, page = 1, pageSize = 10) => {
    setTransactionsLoading(true)
    try {
      const res = await getAccountTransactions(accountId, { page, page_size: pageSize })
      if (res?.data) {
        setAccountTransactions(res.data.transactions || [])
        setTransactionsPagination({
          current: page,
          pageSize,
          total: res.data.pagination?.total || 0,
        })
      }
    } catch (error) {
      message.error('加载交易记录失败')
      console.error(error)
    } finally {
      setTransactionsLoading(false)
    }
  }

  // 查看账户详情
  const handleViewDetail = useCallback(async (account: TrustAccount) => {
    setSelectedAccount(account)
    setDetailModalVisible(true)
    setAccountTransactions([])
    setTransactionsPagination((prev) => ({ ...prev, current: 1, total: 0 }))
    await loadAccountTransactions(account.id)
  }, [])

  // 冻结账户
  const handleFreeze = async (id: number) => {
    try {
      const res = await freezeTrustAccount(id)
      if (res?.data) {
        message.success('账户已冻结')
        loadAccounts(pagination.current, pagination.pageSize)
        if (selectedAccount?.id === id) {
          setSelectedAccount(res.data)
        }
      }
    } catch (error: any) {
      message.error(error?.response?.data?.message || '冻结账户失败')
    }
  }

  // 解冻账户
  const handleUnfreeze = async (id: number) => {
    try {
      const res = await unfreezeTrustAccount(id)
      if (res?.data) {
        message.success('账户已解冻')
        loadAccounts(pagination.current, pagination.pageSize)
        if (selectedAccount?.id === id) {
          setSelectedAccount(res.data)
        }
      }
    } catch (error: any) {
      message.error(error?.response?.data?.message || '解冻账户失败')
    }
  }

  // 关闭账户
  const handleClose = async (id: number) => {
    try {
      const res = await closeTrustAccount(id)
      if (res?.data) {
        message.success('账户已关闭')
        loadAccounts(pagination.current, pagination.pageSize)
        setDetailModalVisible(false)
      }
    } catch (error: any) {
      message.error(error?.response?.data?.message || '关闭账户失败')
    }
  }

  // 表格变化处理
  const handleTableChange = (newPagination: TablePaginationConfig) => {
    loadAccounts(newPagination.current || 1, newPagination.pageSize || 10)
  }

  // 筛选变化处理
  const handleFilterChange = (key: string, value: any) => {
    setFilters((prev) => ({ ...prev, [key]: value }))
    setPagination((prev) => ({ ...prev, current: 1 }))
  }

  // 重置筛选
  const handleResetFilters = () => {
    setFilters({})
    setPagination((prev) => ({ ...prev, current: 1 }))
  }

  // 刷新数据
  const handleRefresh = () => {
    loadAccounts(pagination.current, pagination.pageSize)
    loadStats()
  }

  const handleCreateAccount = async () => {
    try {
      const values = await createForm.validateFields()
      setCreateSubmitting(true)
      const res = await createTrustAccount({
        client_id: values.client_id,
        currency: values.currency,
        purpose_restriction: values.purpose_restriction,
        authorized_uses: values.authorized_uses || [],
      })

      if (res?.data) {
        message.success('代管款账户已创建')
        setCreateModalVisible(false)
        createForm.resetFields()
        loadAccounts(1, pagination.pageSize)
        loadStats()
      }
    } catch (error: any) {
      if (!error?.errorFields) {
        message.error(error?.response?.data?.message || error?.message || '创建账户失败')
      }
    } finally {
      setCreateSubmitting(false)
    }
  }

  // 初始化
  useEffect(() => {
    loadClients()
    loadAccounts()
    loadStats()
  }, [])

  // 筛选变化时重新加载
  useEffect(() => {
    loadAccounts(1, pagination.pageSize)
  }, [filters])

  // 账户表格列定义
  const accountColumns: ColumnsType<TrustAccount> = [
    {
      title: '账户编号',
      dataIndex: 'account_code',
      key: 'account_code',
      width: 150,
      render: (text) => <code>{text}</code>,
    },
    {
      title: '客户',
      dataIndex: 'client_name',
      key: 'client_name',
      width: 150,
      render: (text, record) => text || `客户ID: ${record.client_id}`,
    },
    {
      title: '币种',
      dataIndex: 'currency',
      key: 'currency',
      width: 80,
      render: (currency: Currency) => (
        <Tag color={currency === 'CNY' ? 'red' : currency === 'USD' ? 'green' : 'blue'}>
          {currency}
        </Tag>
      ),
    },
    {
      title: '账户余额',
      dataIndex: 'balance',
      key: 'balance',
      width: 150,
      render: (balance: number, record) => (
        <span style={{ fontWeight: 600 }}>
          {formatAmount(balance, record.currency)}
        </span>
      ),
    },
    {
      title: '冻结金额',
      dataIndex: 'frozen_amount',
      key: 'frozen_amount',
      width: 120,
      render: (amount: number, record) => (
        <span style={{ color: amount > 0 ? '#cf1322' : undefined }}>
          {amount > 0 ? formatAmount(amount, record.currency) : '-'}
        </span>
      ),
    },
    {
      title: '可用余额',
      dataIndex: 'available_amount',
      key: 'available_amount',
      width: 150,
      render: (amount: number, record) => (
        <span style={{ color: '#3f8600', fontWeight: 500 }}>
          {formatAmount(amount, record.currency)}
        </span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: AccountStatus) => {
        const config = accountStatusMap[status]
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '交易笔数',
      dataIndex: 'transaction_count',
      key: 'transaction_count',
      width: 100,
      render: (count) => <span>{count || 0}</span>,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date) => formatDate(date),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right',
      render: (_, record) => (
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
              <Tooltip title="冻结账户">
                <Popconfirm
                  title="确认冻结此账户？"
                  description="冻结后无法进行交易操作"
                  onConfirm={() => handleFreeze(record.id)}
                  okText="确认"
                  cancelText="取消"
                >
                  <Button type="link" size="small" icon={<LockOutlined />} danger />
                </Popconfirm>
              </Tooltip>
              <Tooltip title="关闭账户">
                <Popconfirm
                  title="确认关闭此账户？"
                  description="关闭后将无法恢复，请确认账户余额已清空"
                  onConfirm={() => handleClose(record.id)}
                  okText="确认"
                  cancelText="取消"
                >
                  <Button type="link" size="small" icon={<StopOutlined />} danger />
                </Popconfirm>
              </Tooltip>
            </>
          )}
          {record.status === 'frozen' && (
            <Tooltip title="解冻账户">
              <Popconfirm
                title="确认解冻此账户？"
                onConfirm={() => handleUnfreeze(record.id)}
                okText="确认"
                cancelText="取消"
              >
                <Button type="link" size="small" icon={<UnlockOutlined />} />
              </Popconfirm>
            </Tooltip>
          )}
        </Space>
      ),
    },
  ]

  // 交易表格列定义
  const transactionColumns: ColumnsType<TrustTransaction> = [
    {
      title: '交易编号',
      dataIndex: 'transaction_code',
      key: 'transaction_code',
      width: 150,
      render: (text) => <code>{text}</code>,
    },
    {
      title: '类型',
      dataIndex: 'transaction_type',
      key: 'transaction_type',
      width: 100,
      render: (type: TrustTransaction['transaction_type']) => {
        const config = transactionTypeMap[type]
        return (
          <Tag color={config.color}>
            {config.icon} {config.text}
          </Tag>
        )
      },
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 150,
      render: (amount, record) => {
        const isInflow = record.transaction_type === 'deposit'
        return (
          <span style={{ color: isInflow ? '#3f8600' : '#cf1322', fontWeight: 600 }}>
            {isInflow ? '+' : '-'}
            {formatAmount(amount, selectedAccount?.currency || 'CNY')}
          </span>
        )
      },
    },
    {
      title: '余额后',
      dataIndex: 'balance_after',
      key: 'balance_after',
      width: 150,
      render: (amount) => formatAmount(amount, selectedAccount?.currency || 'CNY'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: TrustTransaction['status']) => {
        const config = transactionStatusMap[status]
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '说明',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (date) => formatDate(date),
    },
  ]

  return (
    <div className="trust-account-management">
      {/* 统计卡片 */}
      {stats && (
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="账户总数"
                value={stats.total_accounts || 0}
                prefix={<WalletOutlined />}
                valueStyle={{ fontSize: 20, fontWeight: 600 }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="活跃账户"
                value={stats.active_accounts || 0}
                prefix={<CheckCircleOutlined />}
                valueStyle={{ color: '#3f8600', fontSize: 20, fontWeight: 600 }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="总余额"
                value={stats.total_balance || 0}
                prefix="¥"
                precision={2}
                valueStyle={{ color: '#1890ff', fontSize: 20, fontWeight: 600 }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="冻结金额"
                value={stats.total_frozen || 0}
                prefix="¥"
                precision={2}
                valueStyle={{ color: '#cf1322', fontSize: 20, fontWeight: 600 }}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* 主卡片 */}
      <Card
        title={
          <Space>
            <WalletOutlined />
            <span>代管款账户</span>
          </Space>
        }
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              刷新
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setCreateModalVisible(true)}
            >
              创建账户
            </Button>
          </Space>
        }
      >
        {/* 筛选栏 */}
        <div className="filter-bar" style={{ marginBottom: 16 }}>
          <Row gutter={16}>
            <Col span={6}>
              <Input
                placeholder="搜索账户编号/客户名"
                prefix={<SearchOutlined />}
                value={filters.search}
                onChange={(e) => handleFilterChange('search', e.target.value)}
                allowClear
              />
            </Col>
            <Col span={4}>
              <Select
                placeholder="选择客户"
                value={filters.clientId}
                onChange={(value) => handleFilterChange('clientId', value)}
                allowClear
                style={{ width: '100%' }}
                showSearch
                filterOption={(input, option) =>
                  String(option?.children ?? '').toLowerCase().includes(input.toLowerCase())
                }
              >
                {clients.map((client) => (
                  <Option key={client.id} value={client.id}>
                    {client.name}
                  </Option>
                ))}
              </Select>
            </Col>
            <Col span={4}>
              <Select
                placeholder="账户状态"
                value={filters.status}
                onChange={(value) => handleFilterChange('status', value)}
                allowClear
                style={{ width: '100%' }}
              >
                <Option value="active">正常</Option>
                <Option value="frozen">已冻结</Option>
                <Option value="closed">已关闭</Option>
              </Select>
            </Col>
            <Col span={4}>
              <Select
                placeholder="币种"
                value={filters.currency}
                onChange={(value) => handleFilterChange('currency', value)}
                allowClear
                style={{ width: '100%' }}
              >
                <Option value="CNY">人民币</Option>
                <Option value="USD">美元</Option>
                <Option value="EUR">欧元</Option>
              </Select>
            </Col>
            <Col span={6}>
              <Space>
                <Button icon={<FilterOutlined />}>筛选</Button>
                <Button onClick={handleResetFilters}>重置</Button>
              </Space>
            </Col>
          </Row>
        </div>

        {/* 账户表格 */}
        <Table
          columns={accountColumns}
          dataSource={accounts}
          rowKey="id"
          loading={loading}
          pagination={pagination}
          onChange={handleTableChange}
          scroll={{ x: 1200 }}
        />
      </Card>

      {/* 账户详情弹窗 */}
      <Modal
        title={
          <Space>
            <WalletOutlined />
            <span>账户详情 - {selectedAccount?.account_code}</span>
          </Space>
        }
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={null}
        width={900}
      >
        {selectedAccount && (
          <Tabs defaultActiveKey="info">
            <Tabs.TabPane tab="基本信息" key="info">
              <Descriptions bordered column={2}>
                <Descriptions.Item label="账户编号" span={2}>
                  <code>{selectedAccount.account_code}</code>
                </Descriptions.Item>
                <Descriptions.Item label="客户">
                  {selectedAccount.client_name || `客户ID: ${selectedAccount.client_id}`}
                </Descriptions.Item>
                <Descriptions.Item label="币种">
                  <Tag color={selectedAccount.currency === 'CNY' ? 'red' : 'blue'}>
                    {selectedAccount.currency}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="账户余额">
                  <span style={{ fontWeight: 600, fontSize: 16 }}>
                    {formatAmount(selectedAccount.balance, selectedAccount.currency)}
                  </span>
                </Descriptions.Item>
                <Descriptions.Item label="可用余额">
                  <span style={{ fontWeight: 600, fontSize: 16, color: '#3f8600' }}>
                    {formatAmount(selectedAccount.available_amount, selectedAccount.currency)}
                  </span>
                </Descriptions.Item>
                <Descriptions.Item label="冻结金额">
                  <span style={{ color: selectedAccount.frozen_amount > 0 ? '#cf1322' : undefined }}>
                    {selectedAccount.frozen_amount > 0
                      ? formatAmount(selectedAccount.frozen_amount, selectedAccount.currency)
                      : '-'}
                  </span>
                </Descriptions.Item>
                <Descriptions.Item label="状态">
                  <Tag color={accountStatusMap[selectedAccount.status].color}>
                    {accountStatusMap[selectedAccount.status].text}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="交易笔数">
                  {selectedAccount.transaction_count || 0} 笔
                </Descriptions.Item>
                <Descriptions.Item label="创建时间" span={2}>
                  {formatDate(selectedAccount.created_at)}
                </Descriptions.Item>
                {selectedAccount.purpose_restriction && (
                  <Descriptions.Item label="备注" span={2}>
                    {selectedAccount.purpose_restriction}
                  </Descriptions.Item>
                )}
              </Descriptions>

              {/* 操作按钮 */}
              {selectedAccount.status === 'active' && (
                <div style={{ marginTop: 16, textAlign: 'center' }}>
                  <Space>
                    <Popconfirm
                      title="确认冻结此账户？"
                      onConfirm={() => handleFreeze(selectedAccount.id)}
                      okText="确认"
                      cancelText="取消"
                    >
                      <Button icon={<LockOutlined />} danger>
                        冻结账户
                      </Button>
                    </Popconfirm>
                    <Popconfirm
                      title="确认关闭此账户？"
                      onConfirm={() => handleClose(selectedAccount.id)}
                      okText="确认"
                      cancelText="取消"
                    >
                      <Button icon={<StopOutlined />} danger>
                        关闭账户
                      </Button>
                    </Popconfirm>
                  </Space>
                </div>
              )}
              {selectedAccount.status === 'frozen' && (
                <div style={{ marginTop: 16, textAlign: 'center' }}>
                  <Popconfirm
                    title="确认解冻此账户？"
                    onConfirm={() => handleUnfreeze(selectedAccount.id)}
                    okText="确认"
                    cancelText="取消"
                  >
                    <Button icon={<UnlockOutlined />} type="primary">
                      解冻账户
                    </Button>
                  </Popconfirm>
                </div>
              )}
            </Tabs.TabPane>

            <Tabs.TabPane
              tab={
                <span>
                  <TransactionOutlined />
                  交易记录
                </span>
              }
              key="transactions"
            >
              <Table
                columns={transactionColumns}
                dataSource={accountTransactions}
                rowKey="id"
                loading={transactionsLoading}
                pagination={{
                  ...transactionsPagination,
                  onChange: (page, pageSize) =>
                    loadAccountTransactions(selectedAccount.id, page, pageSize || 10),
                }}
                scroll={{ x: 800 }}
                size="small"
              />
            </Tabs.TabPane>
          </Tabs>
        )}
      </Modal>

      {/* 创建账户弹窗 */}
      <Modal
        title="创建代管款账户"
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false)
          createForm.resetFields()
        }}
        confirmLoading={createSubmitting}
        footer={[
          <Button
            key="cancel"
            onClick={() => {
              setCreateModalVisible(false)
              createForm.resetFields()
            }}
          >
            取消
          </Button>,
          <Button key="submit" type="primary" loading={createSubmitting} onClick={handleCreateAccount}>
            创建
          </Button>,
        ]}
      >
        <Form
          form={createForm}
          layout="vertical"
          initialValues={{ currency: 'CNY', authorized_uses: ['case_fee'] }}
        >
          <Form.Item
            label="客户"
            name="client_id"
            rules={[{ required: true, message: '请选择客户' }]}
          >
            <Select
              showSearch
              placeholder="选择客户"
              optionFilterProp="label"
              options={clients
                .filter((client) => client.id)
                .map((client) => ({
                  label: client.name,
                  value: client.id,
                }))}
            />
          </Form.Item>
          <Form.Item label="币种" name="currency" rules={[{ required: true, message: '请选择币种' }]}>
            <Select>
              <Option value="CNY">人民币 CNY</Option>
              <Option value="USD">美元 USD</Option>
              <Option value="EUR">欧元 EUR</Option>
            </Select>
          </Form.Item>
          <Form.Item label="授权用途" name="authorized_uses">
            <Select mode="tags" placeholder="输入或选择资金用途">
              <Option value="case_fee">案件费用</Option>
              <Option value="court_fee">诉讼费</Option>
              <Option value="evidence_fee">调查取证费</Option>
              <Option value="settlement">和解款</Option>
            </Select>
          </Form.Item>
          <Form.Item label="用途限制说明" name="purpose_restriction">
            <Input.TextArea rows={3} maxLength={200} showCount placeholder="例如：仅限本案诉讼费和保全费支出" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default TrustAccountManagement
