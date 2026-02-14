/**
 * 代管款交易管理页面
 *
 * 功能：
 * - 交易列表展示与筛选
 * - 交易审批
 * - 交易详情查看
 */

import React, { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  message,
  Modal,
  Descriptions,
  Input,
  Select,
  Row,
  Col,
  Popconfirm,
  Badge,
} from 'antd'
import {
  PlusOutlined,
  EyeOutlined,
  CheckOutlined,
  CloseOutlined,
  ReloadOutlined,
  SearchOutlined,
  FilterOutlined,
  TransactionOutlined,
} from '@ant-design/icons'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import {
  getTrustTransactions,
  getTrustTransaction,
  approveTrustTransaction,
  rejectTrustTransaction,
  type TrustTransaction,
  type TransactionType,
  type TransactionStatus,
  transactionTypeMap,
  transactionStatusMap,
  formatAmount,
  formatDate,
} from '@/services/trust'
import { getTrustAccounts } from '@/services/trust'
import type { TrustAccount, Currency } from '@/services/trust'
import './TrustTransactionManagement.less'

const { Option } = Select

const TrustTransactionManagement: React.FC = () => {
  // 状态管理
  const [loading, setLoading] = useState(false)
  const [transactions, setTransactions] = useState<TrustTransaction[]>([])
  const [accounts, setAccounts] = useState<TrustAccount[]>([])

  // 分页和筛选
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })
  const [filters, setFilters] = useState<{
    accountId?: number
    transactionType?: TransactionType
    status?: TransactionStatus
  }>({})

  // 详情弹窗
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [selectedTransaction, setSelectedTransaction] = useState<TrustTransaction | null>(null)

  // 创建交易弹窗
  const [createModalVisible, setCreateModalVisible] = useState(false)

  // 加载账户列表
  const loadAccounts = async () => {
    try {
      const res = await getTrustAccounts({ page: 1, page_size: 1000 })
      if (res?.data) {
        setAccounts(res.data.accounts || [])
      }
    } catch (error) {
      console.error('加载账户列表失败:', error)
    }
  }

  // 加载交易列表
  const loadTransactions = async (page = 1, pageSize = 10) => {
    setLoading(true)
    try {
      const params: any = {
        page,
        page_size: pageSize,
      }
      if (filters.accountId) params.account_id = filters.accountId
      if (filters.transactionType) params.transaction_type = filters.transactionType
      if (filters.status) params.status = filters.status

      const res = await getTrustTransactions(params)
      if (res?.data) {
        setTransactions(res.data.transactions || [])
        setPagination({
          current: page,
          pageSize,
          total: res.data.pagination?.total || 0,
        })
      }
    } catch (error) {
      message.error('加载交易列表失败')
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  // 查看交易详情
  const handleViewDetail = async (transaction: TrustTransaction) => {
    setSelectedTransaction(transaction)
    setDetailModalVisible(true)
  }

  // 审批通过
  const handleApprove = async (id: number) => {
    try {
      const res = await approveTrustTransaction(id)
      if (res?.data) {
        message.success('交易已审批通过')
        loadTransactions(pagination.current, pagination.pageSize)
        if (selectedTransaction?.id === id) {
          setSelectedTransaction(res.data)
        }
      }
    } catch (error: any) {
      message.error(error?.response?.data?.message || '审批失败')
    }
  }

  // 审批拒绝
  const handleReject = async (id: number) => {
    try {
      const res = await rejectTrustTransaction(id)
      if (res?.data) {
        message.success('交易已拒绝')
        loadTransactions(pagination.current, pagination.pageSize)
        if (selectedTransaction?.id === id) {
          setSelectedTransaction(res.data)
        }
      }
    } catch (error: any) {
      message.error(error?.response?.data?.message || '拒绝失败')
    }
  }

  // 表格变化处理
  const handleTableChange = (newPagination: TablePaginationConfig) => {
    loadTransactions(newPagination.current || 1, newPagination.pageSize || 10)
  }

  // 重置筛选
  const handleResetFilters = () => {
    setFilters({})
    setPagination((prev) => ({ ...prev, current: 1 }))
  }

  // 刷新数据
  const handleRefresh = () => {
    loadTransactions(pagination.current, pagination.pageSize)
  }

  // 初始化
  useEffect(() => {
    loadAccounts()
    loadTransactions()
  }, [])

  // 筛选变化时重新加载
  useEffect(() => {
    loadTransactions(1, pagination.pageSize)
  }, [filters])

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
      title: '账户编号',
      dataIndex: 'account_code',
      key: 'account_code',
      width: 120,
      render: (text) => <code>{text || '-'}</code>,
    },
    {
      title: '类型',
      dataIndex: 'transaction_type',
      key: 'transaction_type',
      width: 100,
      render: (type) => {
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
        const isInflow = ['deposit', 'transfer_in', 'unfreeze'].includes(record.transaction_type)
        const currency = accounts.find((a) => a.id === record.account_id)?.currency || 'CNY'
        const symbol = { CNY: '¥', USD: '$', EUR: '€', HKD: 'HK$' }[currency] || '¥'
        return (
          <span style={{ color: isInflow ? '#3f8600' : '#cf1322', fontWeight: 600 }}>
            {isInflow ? '+' : '-'}
            {symbol}
            {amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </span>
        )
      },
    },
    {
      title: '余额后',
      dataIndex: 'balance_after',
      key: 'balance_after',
      width: 150,
      render: (amount, record) => {
        const currency = accounts.find((a) => a.id === record.account_id)?.currency || 'CNY'
        return formatAmount(amount, currency)
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => {
        const config = transactionStatusMap[status]
        return <Badge status={config.color as any} text={config.text} />
      },
    },
    {
      title: '说明',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '创建人',
      dataIndex: 'created_by_name',
      key: 'created_by_name',
      width: 100,
      render: (name) => name || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
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
          {record.status === 'pending' && (
            <>
              <Popconfirm
                title="确认审批通过此交易？"
                onConfirm={() => handleApprove(record.id)}
                okText="确认"
                cancelText="取消"
              >
                <Button type="link" size="small" icon={<CheckOutlined />}>
                  通过
                </Button>
              </Popconfirm>
              <Popconfirm
                title="确认拒绝此交易？"
                onConfirm={() => handleReject(record.id)}
                okText="确认"
                cancelText="取消"
              >
                <Button type="link" size="small" icon={<CloseOutlined />} danger>
                  拒绝
                </Button>
              </Popconfirm>
            </>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div className="trust-transaction-management">
      {/* 主卡片 */}
      <Card
        title={
          <Space>
            <TransactionOutlined />
            <span>代管款交易</span>
          </Space>
        }
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalVisible(true)}>
              创建交易
            </Button>
          </Space>
        }
      >
        {/* 筛选栏 */}
        <div className="filter-bar" style={{ marginBottom: 16 }}>
          <Row gutter={16}>
            <Col span={6}>
              <Select
                placeholder="选择账户"
                value={filters.accountId}
                onChange={(value) => setFilters((prev) => ({ ...prev, accountId: value }))}
                allowClear
                style={{ width: '100%' }}
                showSearch
                filterOption={(input, option) =>
                  (option?.children as string)?.toLowerCase().includes(input.toLowerCase())
                }
              >
                {accounts.map((account) => (
                  <Option key={account.id} value={account.id}>
                    {account.account_code} - {account.client_name || `客户${account.client_id}`}
                  </Option>
                ))}
              </Select>
            </Col>
            <Col span={4}>
              <Select
                placeholder="交易类型"
                value={filters.transactionType}
                onChange={(value) => setFilters((prev) => ({ ...prev, transactionType: value }))}
                allowClear
                style={{ width: '100%' }}
              >
                <Option value="deposit">存入</Option>
                <Option value="withdraw">支取</Option>
                <Option value="transfer_in">转入</Option>
                <Option value="transfer_out">转出</Option>
                <Option value="freeze">冻结</Option>
                <Option value="unfreeze">解冻</Option>
              </Select>
            </Col>
            <Col span={4}>
              <Select
                placeholder="交易状态"
                value={filters.status}
                onChange={(value) => setFilters((prev) => ({ ...prev, status: value }))}
                allowClear
                style={{ width: '100%' }}
              >
                <Option value="pending">待审批</Option>
                <Option value="approved">已审批</Option>
                <Option value="rejected">已拒绝</Option>
                <Option value="completed">已完成</Option>
              </Select>
            </Col>
            <Col span={10}>
              <Space>
                <Button icon={<FilterOutlined />}>筛选</Button>
                <Button onClick={handleResetFilters}>重置</Button>
              </Space>
            </Col>
          </Row>
        </div>

        {/* 交易表格 */}
        <Table
          columns={transactionColumns}
          dataSource={transactions}
          rowKey="id"
          loading={loading}
          pagination={pagination}
          onChange={handleTableChange}
          scroll={{ x: 1400 }}
        />
      </Card>

      {/* 交易详情弹窗 */}
      <Modal
        title={
          <Space>
            <TransactionOutlined />
            <span>交易详情 - {selectedTransaction?.transaction_code}</span>
          </Space>
        }
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          selectedTransaction?.status === 'pending' ? (
            <>
              <Popconfirm
                key="reject"
                title="确认拒绝此交易？"
                onConfirm={() => handleReject(selectedTransaction.id)}
                okText="确认"
                cancelText="取消"
              >
                <Button icon={<CloseOutlined />} danger>
                  拒绝
                </Button>
              </Popconfirm>
              <Popconfirm
                key="approve"
                title="确认审批通过此交易？"
                onConfirm={() => handleApprove(selectedTransaction.id)}
                okText="确认"
                cancelText="取消"
              >
                <Button type="primary" icon={<CheckOutlined />}>
                  通过
                </Button>
              </Popconfirm>
            </>
          ) : null,
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={700}
      >
        {selectedTransaction && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="交易编号" span={2}>
              <code>{selectedTransaction.transaction_code}</code>
            </Descriptions.Item>
            <Descriptions.Item label="账户编号">
              <code>{selectedTransaction.account_code || '-'}</code>
            </Descriptions.Item>
            <Descriptions.Item label="交易类型">
              {(() => {
                const config = transactionTypeMap[selectedTransaction.transaction_type]
                return (
                  <Tag color={config.color}>
                    {config.icon} {config.text}
                  </Tag>
                )
              })()}
            </Descriptions.Item>
            <Descriptions.Item label="金额">
              <span style={{ fontWeight: 600, fontSize: 16 }}>
                {(() => {
                  const isInflow = ['deposit', 'transfer_in', 'unfreeze'].includes(
                    selectedTransaction.transaction_type
                  )
                  const currency = accounts.find((a) => a.id === selectedTransaction.account_id)?.currency || 'CNY'
                  const symbol = { CNY: '¥', USD: '$', EUR: '€', HKD: 'HK$' }[currency] || '¥'
                  return (
                    <span style={{ color: isInflow ? '#3f8600' : '#cf1322' }}>
                      {isInflow ? '+' : '-'}
                      {symbol}
                      {selectedTransaction.amount.toLocaleString('zh-CN', {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2,
                      })}
                    </span>
                  )
                })()}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="余额后">
              {(() => {
                const currency = accounts.find((a) => a.id === selectedTransaction.account_id)?.currency || 'CNY'
                return formatAmount(selectedTransaction.balance_after, currency)
              })()}
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              {(() => {
                const config = transactionStatusMap[selectedTransaction.status]
                return <Badge status={config.color as any} text={config.text} />
              })()}
            </Descriptions.Item>
            <Descriptions.Item label="创建人">
              {selectedTransaction.created_by_name || `ID: ${selectedTransaction.created_by}`}
            </Descriptions.Item>
            <Descriptions.Item label="审批人">
              {selectedTransaction.approved_by_name || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间" span={2}>
              {formatDate(selectedTransaction.created_at)}
            </Descriptions.Item>
            {selectedTransaction.approved_at && (
              <Descriptions.Item label="审批时间" span={2}>
                {formatDate(selectedTransaction.approved_at)}
              </Descriptions.Item>
            )}
            <Descriptions.Item label="说明" span={2}>
              {selectedTransaction.description}
            </Descriptions.Item>
            {selectedTransaction.reference_no && (
              <Descriptions.Item label="关联单号" span={2}>
                <code>{selectedTransaction.reference_no}</code>
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>

      {/* 创建交易弹窗 */}
      <Modal
        title="创建代管款交易"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setCreateModalVisible(false)}>
            取消
          </Button>,
          <Button key="submit" type="primary" onClick={() => setCreateModalVisible(false)}>
            创建
          </Button>,
        ]}
      >
        <p>创建交易功能待实现，请先选择账户和交易类型</p>
      </Modal>
    </div>
  )
}

export default TrustTransactionManagement
