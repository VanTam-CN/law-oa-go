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
  Form,
  Input,
  InputNumber,
  Select,
  Row,
  Col,
  Popconfirm,
  Badge,
  Tooltip,
} from 'antd'
import {
  PlusOutlined,
  EyeOutlined,
  CheckOutlined,
  CloseOutlined,
  ReloadOutlined,
  FilterOutlined,
  TransactionOutlined,
} from '@ant-design/icons'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import {
  getTrustTransactions,
  createTrustTransaction,
  approveTrustTransaction,
  rejectTrustTransaction,
  type TrustTransaction,
  type TransactionType,
  type TransactionStatus,
  transactionTypeMap,
  transactionStatusMap,
  formatAmount,
  formatDate,
  currencySymbolMap,
} from '@/services/trust'
import { getTrustAccounts } from '@/services/trust'
import type { TrustAccount } from '@/services/trust'
import './TrustTransactionManagement.less'

const { Option } = Select

const TrustTransactionManagement: React.FC = () => {
  const [createForm] = Form.useForm()
  const createTransactionType = Form.useWatch('transaction_type', createForm)
  const needsRecipient = ['deposit_refund', 'withdraw', 'transfer'].includes(createTransactionType)

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
  const [createSubmitting, setCreateSubmitting] = useState(false)

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

  const handleCreateTransaction = async () => {
    try {
      const values = await createForm.validateFields()
      setCreateSubmitting(true)
      const res = await createTrustTransaction({
        account_id: values.account_id,
        transaction_type: values.transaction_type,
        amount: Number(values.amount),
        description: values.description,
        purpose_code: values.purpose_code,
        recipient_name: values.recipient_name,
        recipient_bank_account: values.recipient_bank_account,
        recipient_bank_name: values.recipient_bank_name,
      })

      if (res?.data) {
        message.success('交易已创建，等待审批')
        setCreateModalVisible(false)
        createForm.resetFields()
        loadTransactions(1, pagination.pageSize)
        loadAccounts()
      }
    } catch (error: any) {
      if (!error?.errorFields) {
        message.error(error?.response?.data?.message || error?.message || '创建交易失败')
      }
    } finally {
      setCreateSubmitting(false)
    }
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
        const currency = accounts.find((a) => a.id === record.account_id)?.currency || 'CNY'
        const symbol = currencySymbolMap[currency] || '¥'
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
      render: (status: TrustTransaction['status']) => {
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
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setCreateModalVisible(true)}
            >
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
                  String(option?.children ?? '').toLowerCase().includes(input.toLowerCase())
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
                <Option value="deposit_refund">退回存入</Option>
                <Option value="transfer">转账</Option>
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
                <Option value="completed">已完成</Option>
                <Option value="cancelled">已取消</Option>
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
                  const isInflow = selectedTransaction.transaction_type === 'deposit'
                  const currency =
                    accounts.find((a) => a.id === selectedTransaction.account_id)?.currency || 'CNY'
                  const symbol = currencySymbolMap[currency] || '¥'
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
                const currency =
                  accounts.find((a) => a.id === selectedTransaction.account_id)?.currency || 'CNY'
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
            {selectedTransaction.purpose_code && (
              <Descriptions.Item label="用途代码">
                {selectedTransaction.purpose_code}
              </Descriptions.Item>
            )}
            {selectedTransaction.recipient_name && (
              <Descriptions.Item label="收款方">
                {selectedTransaction.recipient_name}
              </Descriptions.Item>
            )}
            {selectedTransaction.recipient_bank_name && (
              <Descriptions.Item label="收款银行">
                {selectedTransaction.recipient_bank_name}
              </Descriptions.Item>
            )}
            {selectedTransaction.recipient_bank_account && (
              <Descriptions.Item label="收款账号">
                <code>{selectedTransaction.recipient_bank_account}</code>
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>

      {/* 创建交易弹窗 */}
      <Modal
        title="创建代管款交易"
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
          <Button key="submit" type="primary" loading={createSubmitting} onClick={handleCreateTransaction}>
            创建
          </Button>,
        ]}
      >
        <Form
          form={createForm}
          layout="vertical"
          initialValues={{ transaction_type: 'deposit', purpose_code: 'case_fee' }}
        >
          <Form.Item
            label="代管款账户"
            name="account_id"
            rules={[{ required: true, message: '请选择账户' }]}
          >
            <Select
              showSearch
              placeholder="选择账户"
              optionFilterProp="label"
              options={accounts
                .filter((account) => account.status === 'active')
                .map((account) => ({
                  label: `${account.account_code} - ${account.client_name || `客户${account.client_id}`}`,
                  value: account.id,
                }))}
            />
          </Form.Item>
          <Row gutter={12}>
            <Col span={12}>
              <Form.Item
                label="交易类型"
                name="transaction_type"
                rules={[{ required: true, message: '请选择交易类型' }]}
              >
                <Select>
                  <Option value="deposit">存入</Option>
                  <Option value="deposit_refund">退回存入</Option>
                  <Option value="withdraw">支取</Option>
                  <Option value="transfer">转账</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="金额"
                name="amount"
                rules={[{ required: true, message: '请输入交易金额' }]}
              >
                <InputNumber min={0.01} precision={2} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item label="用途代码" name="purpose_code">
            <Select>
              <Option value="case_fee">案件费用</Option>
              <Option value="court_fee">诉讼费</Option>
              <Option value="evidence_fee">调查取证费</Option>
              <Option value="settlement">和解款</Option>
              <Option value="other">其他</Option>
            </Select>
          </Form.Item>
          {needsRecipient && (
            <Row gutter={12}>
              <Col span={8}>
                <Form.Item label="收款方" name="recipient_name">
                  <Input maxLength={200} />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="收款银行" name="recipient_bank_name">
                  <Input maxLength={100} />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label="收款账号" name="recipient_bank_account">
                  <Input maxLength={50} />
                </Form.Item>
              </Col>
            </Row>
          )}
          <Form.Item
            label="交易说明"
            name="description"
            rules={[{ required: true, message: '请输入交易说明' }]}
          >
            <Input.TextArea rows={3} maxLength={500} showCount />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default TrustTransactionManagement
