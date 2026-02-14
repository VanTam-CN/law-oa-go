import React, { useState, useEffect } from 'react'
import {
  Card,
  Tabs,
  Button,
  Space,
  Statistic,
  Row,
  Col,
  message,
} from 'antd'
import {
  BarChartOutlined,
  FileTextOutlined,
  TransactionOutlined,
  WalletOutlined,
  AccountBookOutlined,
  DollarOutlined,
  FundOutlined,
  ThunderboltOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import {
  getFinanceOverview,
  type FinanceOverview,
} from '@/services/finance'
import ContractList from './ContractList'
import InvoiceList from './InvoiceList'
import PaymentList from './PaymentList'
import CommissionReport from './CommissionReport'
import './FinanceManagement.less'

const FinanceManagement: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview')
  const [overview, setOverview] = useState<FinanceOverview | null>(null)
  const [loading, setLoading] = useState(false)

  // 获取财务概览数据
  const fetchOverview = async () => {
    setLoading(true)
    try {
      const res = await getFinanceOverview()
      if (res && res.data) {
        setOverview(res.data)
      }
    } catch (error) {
      console.error('获取财务概览失败:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (activeTab === 'overview') {
      fetchOverview()
    }
  }, [activeTab])

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
              <Card loading={loading}>
                <Statistic
                  title='合同总数'
                  value={overview?.contract_stats?.total_contracts || 0}
                  valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
                  prefix={<FileTextOutlined />}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='生效合同'
                  value={overview?.contract_stats?.active_contracts || 0}
                  valueStyle={{ color: '#3f8600', fontSize: '24px', fontWeight: 700 }}
                  prefix={<FundOutlined />}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='合同总金额'
                  value={overview?.contract_stats?.total_contract_amount || 0}
                  precision={0}
                  valueStyle={{ color: '#1890ff', fontSize: '24px', fontWeight: 700 }}
                  prefix='¥'
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='本月新增合同'
                  value={overview?.contract_stats?.new_contracts_this_month || 0}
                  valueStyle={{ color: '#722ed1', fontSize: '24px', fontWeight: 700 }}
                  prefix={<ThunderboltOutlined />}
                />
              </Card>
            </Col>
          </Row>

          <Row gutter={16} style={{ marginBottom: 24 }}>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='发票总数'
                  value={overview?.invoice_stats?.total_invoices || 0}
                  valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='待审批发票'
                  value={overview?.invoice_stats?.submitted_invoices || 0}
                  valueStyle={{ color: '#faad14', fontSize: '24px', fontWeight: 700 }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='待收款金额'
                  value={overview?.invoice_stats?.pending_invoice_amount || 0}
                  precision={0}
                  valueStyle={{ color: '#cf1322', fontSize: '24px', fontWeight: 700 }}
                  prefix='¥'
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='本月开票金额'
                  value={overview?.invoice_stats?.total_invoice_amount || 0}
                  precision={0}
                  valueStyle={{ color: '#3f8600', fontSize: '24px', fontWeight: 700 }}
                  prefix='¥'
                />
              </Card>
            </Col>
          </Row>

          <Row gutter={16} style={{ marginBottom: 24 }}>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='回款总数'
                  value={overview?.payment_stats?.total_payments || 0}
                  valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='待确认回款'
                  value={overview?.payment_stats?.pending_payments || 0}
                  valueStyle={{ color: '#faad14', fontSize: '24px', fontWeight: 700 }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='本月回款金额'
                  value={overview?.payment_stats?.month_payment_amount || 0}
                  precision={0}
                  valueStyle={{ color: '#3f8600', fontSize: '24px', fontWeight: 700 }}
                  prefix='¥'
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='待确认金额'
                  value={overview?.payment_stats?.pending_amount || 0}
                  precision={0}
                  valueStyle={{ color: '#cf1322', fontSize: '24px', fontWeight: 700 }}
                  prefix='¥'
                />
              </Card>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='提成总数'
                  value={overview?.commission_stats?.total_commissions || 0}
                  valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='待支付提成'
                  value={overview?.commission_stats?.calculated_commissions || 0}
                  valueStyle={{ color: '#faad14', fontSize: '24px', fontWeight: 700 }}
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='提成总金额'
                  value={overview?.commission_stats?.total_commission_amount || 0}
                  precision={0}
                  valueStyle={{ color: '#1890ff', fontSize: '24px', fontWeight: 700 }}
                  prefix='¥'
                />
              </Card>
            </Col>
            <Col span={6}>
              <Card loading={loading}>
                <Statistic
                  title='本月提成'
                  value={overview?.commission_stats?.month_commission_amount || 0}
                  precision={0}
                  valueStyle={{ color: '#3f8600', fontSize: '24px', fontWeight: 700 }}
                  prefix='¥'
                />
              </Card>
            </Col>
          </Row>

          <Card
            title='财务趋势'
            style={{ marginTop: 24 }}
            extra={
              <Button icon={<ReloadOutlined />} onClick={fetchOverview}>
                刷新
              </Button>
            }
          >
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
      key: 'contracts',
      label: (
        <span>
          <FileTextOutlined />
          合同管理
        </span>
      ),
      children: <ContractList />,
    },
    {
      key: 'invoices',
      label: (
        <span>
          <DollarOutlined />
          发票管理
        </span>
      ),
      children: <InvoiceList />,
    },
    {
      key: 'payments',
      label: (
        <span>
          <TransactionOutlined />
          回款管理
        </span>
      ),
      children: <PaymentList />,
    },
    {
      key: 'bad-debts',
      label: (
        <span>
          <WalletOutlined />
          坏账核销
        </span>
      ),
      children: (
        <Card>
          <div style={{ padding: '40px 0', textAlign: 'center', color: '#999' }}>
            <WalletOutlined style={{ fontSize: 48, marginBottom: 16 }} />
            <p>坏账核销功能开发中...</p>
            <p>将包含坏账申请、审批流程等</p>
          </div>
        </Card>
      ),
    },
    {
      key: 'commissions',
      label: (
        <span>
          <AccountBookOutlined />
          提成管理
        </span>
      ),
      children: <CommissionReport />,
    },
  ]

  return (
    <div className='finance-management'>
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} size='large' />
    </div>
  )
}

export default FinanceManagement
