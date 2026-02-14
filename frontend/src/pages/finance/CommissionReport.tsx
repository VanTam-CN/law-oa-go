import React, { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Row,
  Col,
  Statistic,
  Select,
  DatePicker,
  Input,
  message,
  Drawer,
  Form,
  InputNumber,
  Modal,
  Progress,
  Divider,
} from 'antd'
import {
  AccountBookOutlined,
  DollarOutlined,
  UserOutlined,
  FileTextOutlined,
  ReloadOutlined,
  SearchOutlined,
  DownloadOutlined,
  CheckOutlined,
  EyeOutlined,
  CalendarOutlined,
  TrendingUpOutlined,
  FundOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import {
  getCommissions,
  getCommission,
  markCommissionAsPaid,
  cancelCommission,
  getCommissionStats,
  calculateCommissions,
  type Commission,
  type CommissionStats,
  commissionStatusMap,
  formatAmount,
} from '@/services/finance'
import './CommissionReport.less'

const { Option } = Select
const { RangePicker } = DatePicker

// 受益人角色映射
const beneficiaryRoleMap: Record<string, { text: string; color: string }> = {
  source: { text: '案源人', color: 'blue' },
  lawyer: { text: '承办律师', color: 'green' },
  assistant: { text: '助理', color: 'orange' },
}

// 表格列定义
const columns: ColumnsType<Commission> = [
  {
    title: '提成编号',
    dataIndex: 'commission_code',
    key: 'commission_code',
    width: 140,
    fixed: 'left',
  },
  {
    title: '受益人',
    key: 'beneficiary',
    width: 120,
    render: (_, record) => (
      <div>
        <div>{record.beneficiary?.name || '-'}</div>
        <Tag color={beneficiaryRoleMap[record.beneficiary_role]?.color || 'default'} style={{ marginTop: 4 }}>
          {beneficiaryRoleMap[record.beneficiary_role]?.text || record.beneficiary_role}
        </Tag>
      </div>
    ),
  },
  {
    title: '合同',
    key: 'contract',
    width: 120,
    render: (_, record) => (
      <div>
        <div style={{ fontSize: 12, color: '#666' }}>{record.contract?.contract_code || '-'}</div>
        <div style={{ fontSize: 12 }}>
          ¥{record.contract?.contract_amount?.toLocaleString() || 0}
        </div>
      </div>
    ),
  },
  {
    title: '回款金额',
    dataIndex: 'payment_amount',
    key: 'payment_amount',
    width: 110,
    render: (amount: number) => (
      <span style={{ color: '#52c41a', fontWeight: 600 }}>
        ¥{amount?.toLocaleString() || 0}
      </span>
    ),
  },
  {
    title: '成本扣除',
    dataIndex: 'cost_deduction',
    key: 'cost_deduction',
    width: 100,
    render: (amount: number) => (
      <span style={{ color: '#ff4d4f' }}>
        -¥{amount?.toLocaleString() || 0}
      </span>
    ),
  },
  {
    title: '提成基数',
    dataIndex: 'commission_base',
    key: 'commission_base',
    width: 110,
    render: (amount: number) => (
      <span style={{ fontWeight: 600 }}>
        ¥{amount?.toLocaleString() || 0}
      </span>
    ),
  },
  {
    title: '提成比例',
    dataIndex: 'commission_rate',
    key: 'commission_rate',
    width: 90,
    render: (rate: number) => (
      <span>{(rate * 100).toFixed(1)}%</span>
    ),
  },
  {
    title: '提成金额',
    dataIndex: 'commission_amount',
    key: 'commission_amount',
    width: 120,
    render: (amount: number) => (
      <span style={{ fontSize: 15, fontWeight: 'bold', color: '#1890ff' }}>
        ¥{amount?.toLocaleString() || 0}
      </span>
    ),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 90,
    render: (status: string) => {
      const config = commissionStatusMap[status] || { text: status, color: 'default' }
      return <Tag color={config.color}>{config.text}</Tag>
    },
  },
  {
    title: '计算时间',
    dataIndex: 'calculated_at',
    key: 'calculated_at',
    width: 110,
    render: (date: string) => date ? dayjs(date).format('YYYY-MM-DD') : '-',
  },
  {
    title: '支付时间',
    dataIndex: 'paid_date',
    key: 'paid_date',
    width: 110,
    render: (date: string) => date ? dayjs(date).format('YYYY-MM-DD') : '-',
  },
  {
    title: '操作',
    key: 'action',
    width: 150,
    fixed: 'right',
    render: (_, record) => (
      <Space size='small'>
        <Button type='link' size='small' icon={<EyeOutlined />} onClick={() => {/* TODO */}}>
          详情
        </Button>
        {record.status === 'calculated' && (
          <Button type='link' size='small' icon={<CheckOutlined />} onClick={() => {/* TODO */}}>
            标记支付
          </Button>
        )}
        {(record.status === 'pending' || record.status === 'calculated') && (
          <Button type='link' size='small' danger onClick={() => {/* TODO */}}>
            取消
          </Button>
        )}
      </Space>
    ),
  },
]

// 提成汇总表格列
const summaryColumns: ColumnsType<Commission> = [
  {
    title: '受益人',
    key: 'beneficiary',
    width: 150,
    render: (_, record) => (
      <div>
        <div style={{ fontWeight: 600 }}>{record.beneficiary?.name || '-'}</div>
        <Tag color={beneficiaryRoleMap[record.beneficiary_role]?.color || 'default'} size='small'>
          {beneficiaryRoleMap[record.beneficiary_role]?.text || record.beneficiary_role}
        </Tag>
      </div>
    ),
  },
  {
    title: '提成笔数',
    dataIndex: 'count',
    key: 'count',
    width: 100,
    align: 'center',
  },
  {
    title: '提成总额',
    dataIndex: 'total_amount',
    key: 'total_amount',
    width: 150,
    render: (amount: number) => (
      <span style={{ fontSize: 15, fontWeight: 'bold', color: '#1890ff' }}>
        ¥{amount?.toLocaleString() || 0}
      </span>
    ),
  },
  {
    title: '已支付',
    dataIndex: 'paid_amount',
    key: 'paid_amount',
    width: 120,
    render: (amount: number) => (
      <span style={{ color: '#52c41a', fontWeight: 600 }}>
        ¥{amount?.toLocaleString() || 0}
      </span>
    ),
  },
  {
    title: '待支付',
    dataIndex: 'pending_amount',
    key: 'pending_amount',
    width: 120,
    render: (amount: number) => (
      <span style={{ color: '#faad14', fontWeight: 600 }}>
        ¥{amount?.toLocaleString() || 0}
      </span>
    ),
  },
  {
    title: '支付进度',
    key: 'progress',
    width: 150,
    render: (_, record) => {
      const percentage = record.total_amount > 0
        ? (record.paid_amount / record.total_amount) * 100
        : 0
      return (
        <Progress
          percent={percentage}
          size='small'
          status={percentage >= 100 ? 'success' : 'active'}
          strokeColor={percentage >= 100 ? '#52c41a' : '#1890ff'}
        />
      )
    },
  },
]

const CommissionReport: React.FC = () => {
  const [commissions, setCommissions] = useState<Commission[]>([])
  const [summaryData, setSummaryData] = useState<any[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [stats, setStats] = useState<CommissionStats | null>(null)
  const [detailDrawerVisible, setDetailDrawerVisible] = useState<boolean>(false)
  const [selectedCommission, setSelectedCommission] = useState<Commission | null>(null)
  const [payModalVisible, setPayModalVisible] = useState<boolean>(false)
  const [calculateModalVisible, setCalculateModalVisible] = useState<boolean>(false)

  const [form] = Form.useForm()
  const [calculateForm] = Form.useForm()

  // 查询参数
  const [queryParams, setQueryParams] = useState<{
    page: number
    page_size: number
    status: string
    beneficiary_role: string
    beneficiary_id?: number
    date_from: string
    date_to: string
  }>({
    page: 1,
    page_size: 10,
    status: '',
    beneficiary_role: '',
    date_from: '',
    date_to: '',
  })

  // 搜索表单状态
  const [searchForm, setSearchForm] = useState<{
    status: string
    beneficiary_role: string
    dateRange: [dayjs.Dayjs, dayjs.Dayjs] | null
  }>({
    status: '',
    beneficiary_role: '',
    dateRange: null,
  })

  const [total, setTotal] = useState<number>(0)
  const [viewMode, setViewMode] = useState<'list' | 'summary'>('list')

  // 获取提成列表
  const fetchCommissions = async () => {
    setLoading(true)
    try {
      const params: any = { ...queryParams }
      // 移除空值
      Object.keys(params).forEach((key) => {
        if (params[key] === '') {
          delete params[key]
        }
      })

      const res = await getCommissions(params)
      console.log('提成列表API响应:', res)

      let commissionData: Commission[] = []
      let totalCount = 0

      if (res && res.data) {
        if (Array.isArray(res.data)) {
          commissionData = res.data
          totalCount = res.data.length
        } else if (res.pagination) {
          commissionData = res.data || []
          totalCount = res.pagination.total || 0
        }
      }

      setCommissions(commissionData)
      setTotal(totalCount)

      // 生成汇总数据
      generateSummaryData(commissionData)

      // 获取统计数据
      fetchStats()
    } catch (error) {
      console.error('获取提成列表失败:', error)
      message.error('获取提成列表失败')
    } finally {
      setLoading(false)
    }
  }

  // 生成汇总数据
  const generateSummaryData = (data: Commission[]) => {
    const summaryMap = new Map<string, any>()

    data.forEach((item) => {
      const key = `${item.beneficiary_id}-${item.beneficiary_role}`
      if (!summaryMap.has(key)) {
        summaryMap.set(key, {
          beneficiary: item.beneficiary,
          beneficiary_role: item.beneficiary_role,
          count: 0,
          total_amount: 0,
          paid_amount: 0,
          pending_amount: 0,
        })
      }
      const summary = summaryMap.get(key)
      if (summary) {
        summary.count++
        summary.total_amount += item.commission_amount || 0
        if (item.status === 'paid') {
          summary.paid_amount += item.commission_amount || 0
        } else if (item.status === 'calculated') {
          summary.pending_amount += item.commission_amount || 0
        }
      }
    })

    setSummaryData(Array.from(summaryMap.values()))
  }

  // 获取提成统计
  const fetchStats = async () => {
    try {
      const res = await getCommissionStats()
      console.log('提成统计API响应:', res)

      if (res && res.data) {
        setStats(res.data)
      }
    } catch (error) {
      console.error('获取统计数据失败:', error)
    }
  }

  useEffect(() => {
    fetchCommissions()
  }, [queryParams])

  // 查看提成详情
  const handleView = async (record: Commission) => {
    try {
      const res = await getCommission(record.id)
      if (res && res.data) {
        setSelectedCommission(res.data)
        setDetailDrawerVisible(true)
      }
    } catch (error) {
      message.error('获取详情失败')
    }
  }

  // 打开标记支付弹窗
  const handleOpenPayModal = (record: Commission) => {
    setSelectedCommission(record)
    form.resetFields()
    form.setFieldsValue({
      paid_date: dayjs(),
    })
    setPayModalVisible(true)
  }

  // 标记已支付
  const handleMarkPaid = async () => {
    try {
      const values = await form.validateFields()
      if (selectedCommission) {
        await markCommissionAsPaid(selectedCommission.id, {
          paid_date: values.paid_date.format('YYYY-MM-DD'),
          voucher: values.voucher,
        })
        message.success('已标记为支付')
        setPayModalVisible(false)
        fetchCommissions()
      }
    } catch (error) {
      message.error('操作失败')
    }
  }

  // 取消提成
  const handleCancel = async (record: Commission) => {
    try {
      await cancelCommission(record.id)
      message.success('已取消提成')
      fetchCommissions()
    } catch (error) {
      message.error('操作失败')
    }
  }

  // 打开计算提成弹窗
  const handleOpenCalculateModal = () => {
    calculateForm.resetFields()
    setCalculateModalVisible(true)
  }

  // 计算提成
  const handleCalculate = async () => {
    try {
      const values = await calculateForm.validateFields()
      await calculateCommissions({
        payment_id: values.payment_id,
        cost_deduction: values.cost_deduction || 0,
      })
      message.success('提成计算完成')
      setCalculateModalVisible(false)
      fetchCommissions()
    } catch (error) {
      message.error('计算失败')
    }
  }

  // 搜索
  const handleSearch = () => {
    const [date_from, date_to] = searchForm.dateRange || [null, null]
    setQueryParams({
      ...queryParams,
      status: searchForm.status,
      beneficiary_role: searchForm.beneficiary_role,
      date_from: date_from ? date_from.format('YYYY-MM-DD') : '',
      date_to: date_to ? date_to.format('YYYY-MM-DD') : '',
      page: 1,
    })
  }

  // 重置搜索
  const handleReset = () => {
    setSearchForm({
      status: '',
      beneficiary_role: '',
      dateRange: null,
    })
    setQueryParams({
      ...queryParams,
      status: '',
      beneficiary_role: '',
      date_from: '',
      date_to: '',
      page: 1,
    })
  }

  // 导出报表
  const handleExport = () => {
    message.info('导出功能开发中...')
  }

  return (
    <div className='commission-report'>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className='stats-row'>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='提成总数'
              value={stats?.total_commissions || 0}
              valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
              prefix={<AccountBookOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='待支付'
              value={stats?.calculated_commissions || 0}
              valueStyle={{ color: '#faad14', fontSize: '24px', fontWeight: 700 }}
              prefix={<CalendarOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='提成总额'
              value={stats?.total_commission_amount || 0}
              precision={0}
              prefix='¥'
              valueStyle={{ color: '#1890ff', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='本月提成'
              value={stats?.month_commission_amount || 0}
              precision={0}
              prefix='¥'
              valueStyle={{ color: '#3f8600', fontSize: '24px', fontWeight: 700 }}
              prefix={<TrendingUpOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 视图切换和搜索 */}
      <Card className='search-card'>
        <Space size='middle' wrap style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space size='middle' wrap>
            <Select
              placeholder='筛选状态'
              style={{ width: 120 }}
              value={searchForm.status || undefined}
              onChange={(value) => setSearchForm({ ...searchForm, status: value || '' })}
              allowClear
              size='large'
            >
              <Option value='pending'>待计算</Option>
              <Option value='calculated'>已计算</Option>
              <Option value='paid'>已支付</Option>
              <Option value='cancelled'>已取消</Option>
            </Select>

            <Select
              placeholder='受益人角色'
              style={{ width: 120 }}
              value={searchForm.beneficiary_role || undefined}
              onChange={(value) => setSearchForm({ ...searchForm, beneficiary_role: value || '' })}
              allowClear
              size='large'
            >
              <Option value='source'>案源人</Option>
              <Option value='lawyer'>承办律师</Option>
              <Option value='assistant'>助理</Option>
            </Select>

            <RangePicker
              placeholder={['开始日期', '结束日期']}
              value={searchForm.dateRange}
              onChange={(dates) => setSearchForm({ ...searchForm, dateRange: dates })}
              size='large'
            />
          </Space>

          <Space size='middle'>
            <Button.Group>
              <Button
                type={viewMode === 'list' ? 'primary' : 'default'}
                icon={<FileTextOutlined />}
                onClick={() => setViewMode('list')}
                size='large'
              >
                明细列表
              </Button>
              <Button
                type={viewMode === 'summary' ? 'primary' : 'default'}
                icon={<FundOutlined />}
                onClick={() => setViewMode('summary')}
                size='large'
              >
                按人汇总
              </Button>
            </Button.Group>
          </Space>

          <Space size='middle'>
            <Button type='primary' icon={<SearchOutlined />} onClick={handleSearch} size='large'>
              搜索
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleReset} size='large'>
              重置
            </Button>
            <Button icon={<DownloadOutlined />} onClick={handleExport} size='large'>
              导出
            </Button>
            <Button type='primary' icon={<DollarOutlined />} onClick={handleOpenCalculateModal} size='large'>
              计算提成
            </Button>
          </Space>
        </Space>
      </Card>

      {/* 提成列表 */}
      {viewMode === 'list' ? (
        <Card title='提成明细' className='table-card'>
          <Table
            columns={columns.map((col) => ({
              ...col,
              onCell: (record) => ({
                onClick: () => handleView(record),
                style: { cursor: 'pointer' },
              }),
            }))}
            dataSource={commissions}
            rowKey='id'
            loading={loading}
            scroll={{ x: 1500 }}
            pagination={{
              current: queryParams.page,
              pageSize: queryParams.page_size,
              total,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
              onChange: (page, size) => {
                setQueryParams({
                  ...queryParams,
                  page,
                  page_size: size || 10,
                })
              },
            }}
          />
        </Card>
      ) : (
        <Card title='提成汇总（按受益人）' className='table-card'>
          <Table
            columns={summaryColumns}
            dataSource={summaryData}
            rowKey={(record) => `${record.beneficiary?.id}-${record.beneficiary_role}`}
            loading={loading}
            pagination={false}
          />
        </Card>
      )}

      {/* 提成详情抽屉 */}
      <Drawer
        title='提成详情'
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
        width={600}
      >
        {selectedCommission && (
          <div className='commission-detail'>
            {/* 基本信息 */}
            <Card title='基本信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>提成编号：</span>
                  {selectedCommission.commission_code}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>状态：</span>
                  {(() => {
                    const config = commissionStatusMap[selectedCommission.status] || {
                      text: selectedCommission.status,
                      color: 'default',
                    }
                    return <Tag color={config.color}>{config.text}</Tag>
                  })()}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>受益人：</span>
                  {selectedCommission.beneficiary?.name || '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>角色：</span>
                  <Tag color={beneficiaryRoleMap[selectedCommission.beneficiary_role]?.color || 'default'}>
                    {beneficiaryRoleMap[selectedCommission.beneficiary_role]?.text || selectedCommission.beneficiary_role}
                  </Tag>
                </Col>
              </Row>
            </Card>

            {/* 关联信息 */}
            <Card title='关联信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={24}>
                  <span className='detail-label'>合同编号：</span>
                  {selectedCommission.contract?.contract_code || '-'}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>合同金额：</span>
                  ¥{selectedCommission.contract?.contract_amount?.toLocaleString() || 0}
                </Col>
                {selectedCommission.case && (
                  <>
                    <Col span={24}>
                      <span className='detail-label'>关联案件：</span>
                      {selectedCommission.case.title}
                    </Col>
                  </>
                )}
                <Col span={24}>
                  <span className='detail-label'>回款编号：</span>
                  {selectedCommission.payment?.payment_code || '-'}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>回款金额：</span>
                  <span style={{ color: '#52c41a', fontWeight: 600 }}>
                    ¥{selectedCommission.payment_amount?.toLocaleString() || 0}
                  </span>
                </Col>
              </Row>
            </Card>

            {/* 提成计算 */}
            <Card title='提成计算' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>回款金额：</span>
                  <span style={{ fontWeight: 600 }}>
                    ¥{selectedCommission.payment_amount?.toLocaleString() || 0}
                  </span>
                </Col>
                <Col span={12}>
                  <span className='detail-label'>成本扣除：</span>
                  <span style={{ color: '#ff4d4f' }}>
                    -¥{selectedCommission.cost_deduction?.toLocaleString() || 0}
                  </span>
                </Col>
                <Col span={12}>
                  <span className='detail-label'>提成基数：</span>
                  <span style={{ fontWeight: 600 }}>
                    ¥{selectedCommission.commission_base?.toLocaleString() || 0}
                  </span>
                </Col>
                <Col span={12}>
                  <span className='detail-label'>提成比例：</span>
                  <span style={{ fontWeight: 600 }}>
                    {(selectedCommission.commission_rate * 100).toFixed(1)}%
                  </span>
                </Col>
                <Col span={24}>
                  <Divider style={{ margin: '8px 0' }} />
                  <span className='detail-label'>提成金额：</span>
                  <span style={{ fontSize: 20, fontWeight: 'bold', color: '#1890ff' }}>
                    ¥{selectedCommission.commission_amount?.toLocaleString() || 0}
                  </span>
                </Col>
              </Row>
            </Card>

            {/* 支付信息 */}
            <Card title='支付信息' size='small'>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>计算时间：</span>
                  {selectedCommission.calculated_at
                    ? dayjs(selectedCommission.calculated_at).format('YYYY-MM-DD HH:mm')
                    : '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>支付时间：</span>
                  {selectedCommission.paid_date
                    ? dayjs(selectedCommission.paid_date).format('YYYY-MM-DD')
                    : '-'}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>支付凭证：</span>
                  {selectedCommission.payment_voucher || '-'}
                </Col>
              </Row>
            </Card>
          </div>
        )}
      </Drawer>

      {/* 标记支付弹窗 */}
      <Modal
        title='标记提成已支付'
        open={payModalVisible}
        onOk={handleMarkPaid}
        onCancel={() => setPayModalVisible(false)}
        destroyOnClose
      >
        <Form form={form} layout='vertical'>
          <Form.Item
            label='支付日期'
            name='paid_date'
            rules={[{ required: true, message: '请选择支付日期' }]}
          >
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            label='支付凭证号'
            name='voucher'
            rules={[{ required: true, message: '请输入支付凭证号' }]}
          >
            <Input placeholder='请输入支付凭证号或流水号' />
          </Form.Item>
        </Form>
      </Modal>

      {/* 计算提成弹窗 */}
      <Modal
        title='计算提成'
        open={calculateModalVisible}
        onOk={handleCalculate}
        onCancel={() => setCalculateModalVisible(false)}
        destroyOnClose
      >
        <Form form={calculateForm} layout='vertical'>
          <Form.Item
            label='回款ID'
            name='payment_id'
            rules={[{ required: true, message: '请输入回款ID' }]}
            extra='根据回款记录计算相关人员的提成'
          >
            <InputNumber placeholder='请输入回款ID' style={{ width: '100%' }} min={1} />
          </Form.Item>
          <Form.Item
            label='成本扣除'
            name='cost_deduction'
            extra='可选，扣除的办案成本等费用'
          >
            <InputNumber
              placeholder='请输入成本扣除金额'
              style={{ width: '100%' }}
              min={0}
              precision={2}
              prefix='¥'
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default CommissionReport
