import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router'
import {
  Card,
  Tabs,
  Table,
  Tag,
  Statistic,
  Row,
  Col,
  Button,
  Space,
  Descriptions,
  Spin,
  message,
  Progress,
} from 'antd'
import {
  ArrowLeftOutlined,
  ReloadOutlined,
  DollarOutlined,
  FileTextOutlined,
  TransactionOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getContract,
  getContractMilestones,
  getContractInvoices,
  getContractPayments,
  contractStatusMap,
  invoiceStatusMap,
  paymentStatusMap,
  formatAmount,
  formatDate,
  type Contract,
  type PaymentMilestone,
  type Invoice,
  type Payment,
} from '@/services/finance'
import './ContractDetail.less'

const milestoneColumns: ColumnsType<PaymentMilestone> = [
  { title: '序号', dataIndex: 'sequence', key: 'sequence', width: 60 },
  { title: '名称', dataIndex: 'name', key: 'name' },
  {
    title: '金额',
    dataIndex: 'amount',
    key: 'amount',
    render: (v: number) => formatAmount(v),
  },
  {
    title: '占比',
    dataIndex: 'percentage',
    key: 'percentage',
    render: (v: number) => `${v}%`,
  },
  { title: '付款条件', dataIndex: 'condition', key: 'condition', ellipsis: true },
  {
    title: '计划日期',
    dataIndex: 'due_date',
    key: 'due_date',
    render: (v: string) => formatDate(v),
  },
  {
    title: '已付金额',
    dataIndex: 'paid_amount',
    key: 'paid_amount',
    render: (v: number, record: PaymentMilestone) => (
      <span>
        {formatAmount(v)}
        <Progress
          percent={record.amount > 0 ? Math.round((v / record.amount) * 100) : 0}
          size='small'
          style={{ width: 80, marginLeft: 8 }}
        />
      </span>
    ),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (v: string) => {
      const map: Record<string, { text: string; color: string }> = {
        pending: { text: '待付款', color: 'default' },
        billed: { text: '已开票', color: 'processing' },
        partial_paid: { text: '部分付款', color: 'warning' },
        paid: { text: '已付清', color: 'success' },
        overdue: { text: '已逾期', color: 'error' },
      }
      const info = map[v] || { text: v, color: 'default' }
      return <Tag color={info.color}>{info.text}</Tag>
    },
  },
]

const invoiceColumns: ColumnsType<Invoice> = [
  { title: '发票编号', dataIndex: 'invoice_code', key: 'invoice_code' },
  {
    title: '发票类型',
    dataIndex: 'invoice_type',
    key: 'invoice_type',
    render: (v: string) => (v === 'credit' ? '红字发票' : '蓝字发票'),
  },
  {
    title: '发票金额',
    dataIndex: 'amount',
    key: 'amount',
    render: (v: number) => formatAmount(v),
  },
  { title: '税率', dataIndex: 'tax_rate', key: 'tax_rate', render: (v: number) => `${v}%` },
  {
    title: '价税合计',
    dataIndex: 'total_amount',
    key: 'total_amount',
    render: (v: number) => formatAmount(v),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (v: string) => {
      const info = invoiceStatusMap[v] || { text: v, color: 'default' }
      return <Tag color={info.color}>{info.text}</Tag>
    },
  },
  {
    title: '开票日期',
    dataIndex: 'issued_at',
    key: 'issued_at',
    render: (v: string) => formatDate(v),
  },
  {
    title: '已回款',
    dataIndex: 'total_paid_amount',
    key: 'total_paid_amount',
    render: (v: number) => formatAmount(v),
  },
]

const paymentColumns: ColumnsType<Payment> = [
  { title: '回款编号', dataIndex: 'payment_code', key: 'payment_code' },
  {
    title: '回款金额',
    dataIndex: 'amount',
    key: 'amount',
    render: (v: number) => formatAmount(v),
  },
  {
    title: '回款方式',
    dataIndex: 'payment_method',
    key: 'payment_method',
    render: (v: string) => {
      const map: Record<string, string> = {
        bank_transfer: '银行转账',
        cash: '现金',
        other: '其他',
      }
      return map[v] || v
    },
  },
  {
    title: '回款日期',
    dataIndex: 'payment_date',
    key: 'payment_date',
    render: (v: string) => formatDate(v),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (v: string) => {
      const info = paymentStatusMap[v] || { text: v, color: 'default' }
      return <Tag color={info.color}>{info.text}</Tag>
    },
  },
  { title: '付款人', dataIndex: 'payer_name', key: 'payer_name' },
  { title: '备注', dataIndex: 'remark', key: 'remark', ellipsis: true },
]

const ContractDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [contract, setContract] = useState<Contract | null>(null)
  const [milestones, setMilestones] = useState<PaymentMilestone[]>([])
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [payments, setPayments] = useState<Payment[]>([])
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('overview')

  const fetchAll = async () => {
    if (!id) return
    setLoading(true)
    try {
      const [contractRes, milestoneRes, invoiceRes, paymentRes] = await Promise.all([
        getContract(Number(id)),
        getContractMilestones(Number(id)),
        getContractInvoices(Number(id)),
        getContractPayments(Number(id)),
      ])

      if (contractRes?.data) setContract(contractRes.data)
      if (milestoneRes?.data) setMilestones(Array.isArray(milestoneRes.data) ? milestoneRes.data : [])
      if (invoiceRes?.data) setInvoices(Array.isArray(invoiceRes.data) ? invoiceRes.data : [])
      if (paymentRes?.data) setPayments(Array.isArray(paymentRes.data) ? paymentRes.data : [])
    } catch (error) {
      message.error('加载合同详情失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchAll()
  }, [id])

  // 计算费用统计数据
  const totalInvoiced = invoices.reduce((sum, inv) => sum + (inv.total_amount || 0), 0)
  const totalPaid = payments
    .filter((p) => p.status === 'confirmed')
    .reduce((sum, p) => sum + (p.amount || 0), 0)
  const unpaidAmount = (contract?.contract_amount || 0) - totalPaid
  const milestoneTotalPaid = milestones.reduce((sum, m) => sum + (m.paid_amount || 0), 0)

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 100 }}>
        <Spin size='large' />
      </div>
    )
  }

  if (!contract) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: 60, color: '#999' }}>
          <p>合同不存在或已被删除</p>
          <Button onClick={() => navigate('/finance')}>返回合同列表</Button>
        </div>
      </Card>
    )
  }

  const statusInfo = contractStatusMap[contract.status] || { text: contract.status, color: 'default' }

  const tabItems = [
    {
      key: 'overview',
      label: (
        <span>
          <DollarOutlined />
          费用总览
        </span>
      ),
      children: (
        <div>
          <Descriptions
            bordered
            column={2}
            style={{ marginBottom: 24 }}
            title='合同基本信息'
          >
            <Descriptions.Item label='合同编号'>{contract.contract_code}</Descriptions.Item>
            <Descriptions.Item label='合同状态'>
              <Tag color={statusInfo.color}>{statusInfo.text}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label='客户名称'>{contract.client?.name || '-'}</Descriptions.Item>
            <Descriptions.Item label='案件名称'>{contract.case?.title || '-'}</Descriptions.Item>
            <Descriptions.Item label='合同金额'>
              <span style={{ fontSize: 16, fontWeight: 600, color: '#1890ff' }}>
                {formatAmount(contract.contract_amount)}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label='合同类型'>
              {contract.contract_type === 'original' ? '原始合同' : '补充合同'}
            </Descriptions.Item>
            <Descriptions.Item label='签订日期'>{formatDate(contract.signed_at)}</Descriptions.Item>
            <Descriptions.Item label='合同期限'>
              {formatDate(contract.start_date)} ~ {formatDate(contract.end_date)}
            </Descriptions.Item>
            <Descriptions.Item label='结算周期'>{contract.billing_cycle}</Descriptions.Item>
            <Descriptions.Item label='付款条款'>{contract.payment_terms}</Descriptions.Item>
          </Descriptions>

          <Row gutter={16}>
            <Col span={6}>
              <Card>
                <Statistic
                  title='合同金额'
                  value={contract.contract_amount}
                  precision={2}
                  prefix='¥'
                  valueStyle={{ color: '#1890ff' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title='已开票金额'
                  value={totalInvoiced}
                  precision={2}
                  prefix='¥'
                  valueStyle={{ color: '#722ed1' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title='已回款金额'
                  value={totalPaid}
                  precision={2}
                  prefix='¥'
                  valueStyle={{ color: '#3f8600' }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card>
                <Statistic
                  title='待回款金额'
                  value={unpaidAmount}
                  precision={2}
                  prefix='¥'
                  valueStyle={{ color: unpaidAmount > 0 ? '#cf1322' : '#3f8600' }}
                />
              </Card>
            </Col>
          </Row>

          {contract.contract_amount > 0 && (
            <Card title='回款进度' style={{ marginTop: 24 }}>
              <Progress
                percent={Math.round((totalPaid / contract.contract_amount) * 100)}
                strokeColor={{
                  '0%': '#cf1322',
                  '50%': '#faad14',
                  '100%': '#3f8600',
                }}
                format={(percent) => `${percent}%`}
              />
              <div style={{ marginTop: 16, display: 'flex', justifyContent: 'space-between', color: '#666' }}>
                <span>里程碑完成: {formatAmount(milestoneTotalPaid)} / {formatAmount(contract.contract_amount)}</span>
                <span>开票完成: {formatAmount(totalInvoiced)} / {formatAmount(contract.contract_amount)}</span>
                <span>回款完成: {formatAmount(totalPaid)} / {formatAmount(contract.contract_amount)}</span>
              </div>
            </Card>
          )}
        </div>
      ),
    },
    {
      key: 'milestones',
      label: (
        <span>
          <UnorderedListOutlined />
          付款计划 ({milestones.length})
        </span>
      ),
      children: (
        <Table
          dataSource={milestones}
          columns={milestoneColumns}
          rowKey='id'
          pagination={false}
          size='middle'
        />
      ),
    },
    {
      key: 'invoices',
      label: (
        <span>
          <FileTextOutlined />
          发票记录 ({invoices.length})
        </span>
      ),
      children: (
        <Table
          dataSource={invoices}
          columns={invoiceColumns}
          rowKey='id'
          pagination={false}
          size='middle'
        />
      ),
    },
    {
      key: 'payments',
      label: (
        <span>
          <TransactionOutlined />
          回款记录 ({payments.length})
        </span>
      ),
      children: (
        <Table
          dataSource={payments}
          columns={paymentColumns}
          rowKey='id'
          pagination={false}
          size='middle'
        />
      ),
    },
  ]

  return (
    <div className='contract-detail'>
      <Card
        title={
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/finance')}>
              返回
            </Button>
            <span>合同详情 - {contract.contract_code}</span>
          </Space>
        }
        extra={
          <Button icon={<ReloadOutlined />} onClick={fetchAll} loading={loading}>
            刷新
          </Button>
        }
      >
        <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
      </Card>
    </div>
  )
}

export default ContractDetail
