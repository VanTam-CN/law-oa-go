import React, { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  InputNumber,
  message,
  Popconfirm,
  Statistic,
  Row,
  Col,
  Tooltip,
  Drawer,
  Radio,
  Steps,
  Divider,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SearchOutlined,
  ReloadOutlined,
  SendOutlined,
  CheckOutlined,
  CloseOutlined,
  FileTextOutlined,
  PrinterOutlined,
  DownloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import {
  getInvoices,
  getInvoice,
  createInvoice,
  updateInvoice,
  deleteInvoice,
  submitInvoice,
  approveInvoice,
  rejectInvoice,
  issueInvoice,
  confirmInvoiceReceipt,
  cancelInvoice,
  getInvoiceStats,
  exportInvoices,
  type Invoice,
  type CreateInvoiceRequest,
  type UpdateInvoiceRequest,
  type InvoiceStats,
  invoiceStatusMap,
  formatAmount,
} from '@/services/finance'
import { exportCSV } from '@/utils/export'
import FilterSelect from './FilterSelect'
import './InvoiceList.less'

const { Option } = Select
const { TextArea } = Input
const { RangePicker } = DatePicker

const steps: { title: string; status: 'wait' | 'process' | 'finish' | 'error' }[] = [
  { title: '草稿', status: 'wait' },
  { title: '待审批', status: 'process' },
  { title: '已审批', status: 'process' },
  { title: '已开票', status: 'process' },
  { title: '已签收', status: 'finish' },
]

// 表格列定义
const columns: ColumnsType<Invoice> = [
  {
    title: '发票号码',
    dataIndex: 'invoice_code',
    key: 'invoice_code',
    width: 150,
    fixed: 'left',
  },
  {
    title: '客户名称',
    dataIndex: 'client_name',
    key: 'client_name',
    width: 150,
  },
  {
    title: '发票类型',
    dataIndex: 'invoice_type',
    key: 'invoice_type',
    width: 100,
    render: (type: string) => (
      <Tag color={type === 'credit' ? 'red' : 'blue'}>
        {type === 'credit' ? '红冲发票' : '普通发票'}
      </Tag>
    ),
  },
  {
    title: '金额',
    key: 'amount',
    width: 120,
    render: (_, record) => (
      <div>
        <div>¥{record.amount.toLocaleString()}</div>
        {record.tax_amount > 0 && (
          <div style={{ fontSize: 12, color: '#999' }}>
            税额: ¥{record.tax_amount.toLocaleString()}
          </div>
        )}
      </div>
    ),
  },
  {
    title: '总金额',
    dataIndex: 'total_amount',
    key: 'total_amount',
    width: 120,
    render: (amount: number) => (
      <span style={{ fontWeight: 'bold', color: '#f5222d' }}>
        ¥{amount.toLocaleString()}
      </span>
    ),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    render: (status: string) => {
      const config = invoiceStatusMap[status] || { text: status, color: 'default' }
      return <Tag color={config.color}>{config.text}</Tag>
    },
  },
  {
    title: '已收金额',
    key: 'paid',
    width: 100,
    render: (_, record) => {
      const percentage = record.total_amount > 0
        ? (record.total_paid_amount / record.total_amount) * 100
        : 0
      return (
        <div>
          <div>¥{record.total_paid_amount.toLocaleString()}</div>
          <div style={{ fontSize: 11, color: percentage >= 100 ? '#52c41a' : '#faad14' }}>
            {percentage.toFixed(0)}%
          </div>
        </div>
      )
    },
  },
  {
    title: '剩余金额',
    dataIndex: 'remaining_amount',
    key: 'remaining_amount',
    width: 100,
    render: (amount: number, record) =>
      amount > 0 ? (
        <span style={{ color: '#f5222d' }}>¥{amount.toLocaleString()}</span>
      ) : (
        <Tag color='success'>已结清</Tag>
      ),
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 120,
    render: (date: string) => dayjs(date).format('YYYY-MM-DD'),
  },
  {
    title: '操作',
    key: 'action',
    width: 180,
    fixed: 'right',
    render: (_, record) => (
      <Space size='small'>
        <Tooltip title='查看详情'>
          <Button type='link' size='small' icon={<EyeOutlined />} />
        </Tooltip>
        {record.status === 'draft' && (
          <>
            <Tooltip title='编辑'>
              <Button type='link' size='small' icon={<EditOutlined />} />
            </Tooltip>
            <Tooltip title='提交审批'>
              <Button type='link' size='small' icon={<SendOutlined />} />
            </Tooltip>
            <Popconfirm title='确定要删除这张发票吗？'>
              <Button type='link' size='small' danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </>
        )}
        {record.status === 'submitted' && (
          <>
            <Tooltip title='审批通过'>
              <Button type='link' size='small' icon={<CheckOutlined />} />
            </Tooltip>
            <Tooltip title='审批拒绝'>
              <Button type='link' size='small' danger icon={<CloseOutlined />} />
            </Tooltip>
          </>
        )}
        {record.status === 'approved' && (
          <>
            <Tooltip title='开票'>
              <Button type='link' size='small' icon={<PrinterOutlined />} />
            </Tooltip>
            <Tooltip title='作废'>
              <Button type='link' size='small' danger>
                作废
              </Button>
            </Tooltip>
          </>
        )}
        {record.status === 'issued' && (
          <Tooltip title='确认签收'>
            <Button type='link' size='small' icon={<CheckOutlined />}>
              签收
            </Button>
          </Tooltip>
        )}
      </Space>
    ),
  },
]

const InvoiceList: React.FC = () => {
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [modalVisible, setModalVisible] = useState<boolean>(false)
  const [modalTitle, setModalTitle] = useState<string>('')
  const [editingInvoice, setEditingInvoice] = useState<Invoice | null>(null)
  const [stats, setStats] = useState<InvoiceStats | null>(null)
  const [detailDrawerVisible, setDetailDrawerVisible] = useState<boolean>(false)
  const [selectedInvoice, setSelectedInvoice] = useState<Invoice | null>(null)
  const [issueModalVisible, setIssueModalVisible] = useState<boolean>(false)
  const [form] = Form.useForm()
  const [issueForm] = Form.useForm()

  // 查询参数
  const [queryParams, setQueryParams] = useState<{
    page: number
    page_size: number
    status: string
    invoice_type: string
    client_id?: number
    date_from: string
    date_to: string
  }>({
    page: 1,
    page_size: 10,
    status: '',
    invoice_type: '',
    client_id: undefined,
    date_from: '',
    date_to: '',
  })

  // 搜索表单状态
  const [searchForm, setSearchForm] = useState<{
    status: string
    invoice_type: string
    dateRange: [dayjs.Dayjs, dayjs.Dayjs] | null
  }>({
    status: '',
    invoice_type: '',
    dateRange: null,
  })

  const [total, setTotal] = useState<number>(0)

  // 获取发票列表
  const fetchInvoices = async () => {
    setLoading(true)
    try {
      const params: any = { ...queryParams }
      // 移除空值
      Object.keys(params).forEach((key) => {
        if (params[key] === '') {
          delete params[key]
        }
      })

      const res = await getInvoices(params)
      console.log('发票列表API响应:', res)

      let invoiceData: Invoice[] = []
      let totalCount = 0

      if (res && res.data) {
        if (Array.isArray(res.data)) {
          invoiceData = res.data
          totalCount = res.data.length
        } else if (res.pagination) {
          invoiceData = res.data || []
          totalCount = res.pagination.total || 0
        }
      }

      setInvoices(invoiceData)
      setTotal(totalCount)

      // 获取统计数据
      fetchStats()
    } catch (error) {
      console.error('获取发票列表失败:', error)
      message.error('获取发票列表失败')
    } finally {
      setLoading(false)
    }
  }

  // 获取发票统计
  const fetchStats = async () => {
    try {
      const res = await getInvoiceStats()
      console.log('发票统计API响应:', res)

      if (res && res.data) {
        setStats(res.data)
      }
    } catch (error) {
      console.error('获取统计数据失败:', error)
    }
  }

  useEffect(() => {
    fetchInvoices()
  }, [queryParams])

  // 打开新增发票弹窗
  const handleAdd = () => {
    setModalTitle('新增发票')
    setEditingInvoice(null)
    form.resetFields()
    form.setFieldsValue({
      invoice_type: 'normal',
      tax_rate: 0.06,
    })
    setModalVisible(true)
  }

  // 打开编辑发票弹窗
  const handleEdit = (record: Invoice) => {
    setModalTitle('编辑发票')
    setEditingInvoice(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  // 查看发票详情
  const handleView = (record: Invoice) => {
    setSelectedInvoice(record)
    setDetailDrawerVisible(true)
  }

  // 删除发票
  const handleDelete = async (id: number) => {
    try {
      await deleteInvoice(id)
      message.success('删除成功')
      fetchInvoices()
    } catch (error) {
      message.error('删除失败')
    }
  }

  // 提交审批
  const handleSubmit = async (id: number) => {
    try {
      await submitInvoice(id)
      message.success('已提交审批')
      fetchInvoices()
    } catch (error) {
      message.error('提交失败')
    }
  }

  // 审批通过
  const handleApprove = async (id: number) => {
    try {
      await approveInvoice(id)
      message.success('审批通过')
      fetchInvoices()
    } catch (error) {
      message.error('审批失败')
    }
  }

  // 审批拒绝
  const handleReject = async (id: number) => {
    try {
      await rejectInvoice(id)
      message.success('已拒绝')
      fetchInvoices()
    } catch (error) {
      message.error('操作失败')
    }
  }

  // 打开开票弹窗
  const handleOpenIssueModal = (record: Invoice) => {
    setSelectedInvoice(record)
    issueForm.resetFields()
    setIssueModalVisible(true)
  }

  // 开票
  const handleIssue = async () => {
    try {
      const values = await issueForm.validateFields()
      if (selectedInvoice) {
        await issueInvoice(selectedInvoice.id, values)
        message.success('开票成功')
        setIssueModalVisible(false)
        fetchInvoices()
      }
    } catch (error) {
      message.error('开票失败')
    }
  }

  // 确认签收
  const handleConfirmReceipt = async (id: number) => {
    try {
      await confirmInvoiceReceipt(id)
      message.success('已确认签收')
      fetchInvoices()
    } catch (error) {
      message.error('操作失败')
    }
  }

  // 作废发票
  const handleCancel = async (id: number) => {
    try {
      await cancelInvoice(id)
      message.success('发票已作废')
      fetchInvoices()
    } catch (error) {
      message.error('作废失败')
    }
  }

  // 提交表单
  const handleModalSubmit = async () => {
    try {
      const values = await form.validateFields()

      // 计算税额和总金额
      const taxAmount = values.amount * (values.tax_rate || 0)
      const totalAmount = values.amount + taxAmount

      const formData: CreateInvoiceRequest | UpdateInvoiceRequest = {
        ...values,
        tax_amount: taxAmount,
        total_amount: totalAmount,
      }

      if (editingInvoice) {
        await updateInvoice(editingInvoice.id, formData as UpdateInvoiceRequest)
        message.success('更新成功')
      } else {
        await createInvoice(formData as CreateInvoiceRequest)
        message.success('新增成功')
      }
      setModalVisible(false)
      fetchInvoices()
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  // 搜索
  const handleSearch = () => {
    const [date_from, date_to] = searchForm.dateRange || [null, null]
    setQueryParams({
      ...queryParams,
      status: searchForm.status,
      invoice_type: searchForm.invoice_type,
      date_from: date_from ? date_from.format('YYYY-MM-DD') : '',
      date_to: date_to ? date_to.format('YYYY-MM-DD') : '',
      page: 1,
      client_id: queryParams.client_id,
    })
  }

  // 重置搜索
  const handleReset = () => {
    const resetParams: typeof queryParams = {
      page: 1,
      page_size: 10,
      status: '',
      invoice_type: '',
      date_from: '',
      date_to: '',
      client_id: undefined,
    }
    setSearchForm({
      status: '',
      invoice_type: '',
      dateRange: null,
    })
    setQueryParams(resetParams)
  }

  // 导出报表
  const handleExport = async () => {
    try {
      const params: Record<string, unknown> = {
        status: queryParams.status,
        invoice_type: queryParams.invoice_type,
        date_from: queryParams.date_from,
        date_to: queryParams.date_to,
        client_id: queryParams.client_id,
      }
      // 移除空值
      Object.keys(params).forEach((key) => {
        if (params[key] === '' || params[key] === undefined) {
          delete params[key]
        }
      })

      await exportCSV(
        exportInvoices(params),
        `发票报表_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.csv`
      )
      message.success('导出成功')
    } catch (error) {
      message.error('导出失败')
    }
  }

  // 获取当前步骤索引
  const getCurrentStep = (status: string) => {
    const index = steps.findIndex((s) => s.status === status)
    return index >= 0 ? index : 0
  }

  return (
    <div className='invoice-list'>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className='stats-row'>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='发票总数'
              value={stats?.total_invoices || 0}
              valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='待审批'
              value={stats?.submitted_invoices || 0}
              valueStyle={{ color: '#faad14', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='待收款'
              value={stats?.pending_invoice_amount || 0}
              prefix='¥'
              valueStyle={{ color: '#cf1322', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='本月开票'
              value={stats?.total_invoice_amount || 0}
              prefix='¥'
              valueStyle={{ color: '#3f8600', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索过滤器 */}
      <Card className='search-card'>
        <Space size='middle' wrap>
          <FilterSelect
            placeholder='筛选状态'
            value={searchForm.status || undefined}
            onChange={(value) => setSearchForm({ ...searchForm, status: value })}
            options={[
              { value: 'draft', label: '草稿' },
              { value: 'submitted', label: '待审批' },
              { value: 'approved', label: '已审批' },
              { value: 'issued', label: '已开票' },
              { value: 'received', label: '已签收' },
            ]}
          />

          <Select
            placeholder='发票类型'
            style={{ width: 120 }}
            value={searchForm.invoice_type || undefined}
            onChange={(value) => setSearchForm({ ...searchForm, invoice_type: value || '' })}
            allowClear
            size='large'
          >
            <Option value='normal'>普通发票</Option>
            <Option value='credit'>红冲发票</Option>
          </Select>

          <RangePicker
            placeholder={['开始日期', '结束日期']}
            value={searchForm.dateRange}
            onChange={(dates) => setSearchForm({ ...searchForm, dateRange: dates as [dayjs.Dayjs, dayjs.Dayjs] | null })}
            size='large'
          />

          <Button type='primary' icon={<SearchOutlined />} onClick={handleSearch} size='large'>
            搜索
          </Button>

          <Button icon={<ReloadOutlined />} onClick={handleReset} size='large'>
            重置筛选
          </Button>

          <Button icon={<DownloadOutlined />} onClick={handleExport} size='large'>
            导出
          </Button>
        </Space>
      </Card>

      {/* 发票列表 */}
      <Card
        title='发票列表'
        extra={
          <Button type='primary' icon={<PlusOutlined />} onClick={handleAdd}>
            新增发票
          </Button>
        }
      >
        <Table
          columns={columns.map((col) => ({
            ...col,
            onCell: (record: Invoice) => ({
              onClick: () => handleView(record),
              style: { cursor: 'pointer' },
            }),
          }))}
          dataSource={invoices}
          rowKey='id'
          loading={loading}
          scroll={{ x: 1400 }}
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

      {/* 新增/编辑弹窗 */}
      <Modal
        title={modalTitle}
        open={modalVisible}
        onOk={handleModalSubmit}
        onCancel={() => setModalVisible(false)}
        width={800}
        destroyOnClose
      >
        <Form form={form} layout='vertical'>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label='发票号码'
                name='invoice_code'
                rules={[{ required: true, message: '请输入发票号码' }]}
              >
                <Input placeholder='请输入发票号码' />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label='客户ID'
                name='client_id'
                rules={[{ required: true, message: '请输入客户ID' }]}
              >
                <InputNumber placeholder='请输入客户ID' style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label='合同ID' name='contract_id'>
                <InputNumber placeholder='可选，关联合同' style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label='付款计划ID' name='milestone_id'>
                <InputNumber placeholder='可选，关联付款计划' style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label='发票类型'
            name='invoice_type'
            rules={[{ required: true, message: '请选择发票类型' }]}
          >
            <Radio.Group>
              <Radio value='normal'>普通发票</Radio>
              <Radio value='credit'>红冲发票</Radio>
            </Radio.Group>
          </Form.Item>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label='金额'
                name='amount'
                rules={[{ required: true, message: '请输入金额' }]}
              >
                <InputNumber
                  placeholder='请输入金额'
                  style={{ width: '100%' }}
                  min={0}
                  precision={2}
                  onChange={(value) => {
                    const taxRate = form.getFieldValue('tax_rate') || 0
                    const taxAmount = (value || 0) * taxRate
                    form.setFieldsValue({
                      tax_amount: taxAmount.toFixed(2),
                      total_amount: ((value || 0) + taxAmount).toFixed(2),
                    })
                  }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label='税率' name='tax_rate'>
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  max={1}
                  step={0.01}
                  precision={2}
                  onChange={(value) => {
                    const amount = form.getFieldValue('amount') || 0
                    const taxAmount = amount * (value || 0)
                    form.setFieldsValue({
                      tax_amount: taxAmount.toFixed(2),
                      total_amount: (amount + taxAmount).toFixed(2),
                    })
                  }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label='税额' name='tax_amount'>
                <InputNumber style={{ width: '100%' }} disabled precision={2} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item label='总金额' name='total_amount'>
            <InputNumber style={{ width: '100%' }} disabled precision={2} />
          </Form.Item>

          <Divider orientation='left'>客户开票信息</Divider>

          <Form.Item
            label='客户名称（开票抬头）'
            name='client_name'
            rules={[{ required: true, message: '请输入客户名称' }]}
          >
            <Input placeholder='请输入客户开票名称' />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label='纳税人识别号' name='client_tax_id'>
                <Input placeholder='请输入纳税人识别号' />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label='开户银行' name='client_bank_name'>
                <Input placeholder='请输入开户银行' />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label='银行账号' name='client_bank_account'>
                <Input placeholder='请输入银行账号' />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item label='开票地址' name='client_address'>
            <TextArea placeholder='请输入开票地址' rows={2} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 发票详情抽屉 */}
      <Drawer
        title='发票详情'
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
        width={700}
      >
        {selectedInvoice && (
          <div className='invoice-detail'>
            {/* 发票流程 */}
            <Card title='发票状态' size='small' style={{ marginBottom: 16 }}>
              <Steps
                current={getCurrentStep(selectedInvoice.status)}
                size='small'
                items={steps}
              />
            </Card>

            {/* 基本信息 */}
            <Card title='基本信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>发票号码：</span>
                  {selectedInvoice.invoice_code}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>发票类型：</span>
                  <Tag color={selectedInvoice.invoice_type === 'credit' ? 'red' : 'blue'}>
                    {selectedInvoice.invoice_type === 'credit' ? '红冲发票' : '普通发票'}
                  </Tag>
                </Col>
                <Col span={12}>
                  <span className='detail-label'>客户名称：</span>
                  {selectedInvoice.client_name}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>发票状态：</span>
                  {(() => {
                    const config = invoiceStatusMap[selectedInvoice.status] || {
                      text: selectedInvoice.status,
                      color: 'default',
                    }
                    return <Tag color={config.color}>{config.text}</Tag>
                  })()}
                </Col>
              </Row>
            </Card>

            {/* 金额信息 */}
            <Card title='金额信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={8}>
                  <span className='detail-label'>金额：</span>
                  <span style={{ fontSize: 16 }}>
                    ¥{selectedInvoice.amount.toLocaleString()}
                  </span>
                </Col>
                <Col span={8}>
                  <span className='detail-label'>税额：</span>
                  <span style={{ fontSize: 16 }}>
                    ¥{selectedInvoice.tax_amount.toLocaleString()}
                  </span>
                </Col>
                <Col span={8}>
                  <span className='detail-label'>总金额：</span>
                  <span style={{ fontSize: 18, fontWeight: 'bold', color: '#f5222d' }}>
                    ¥{selectedInvoice.total_amount.toLocaleString()}
                  </span>
                </Col>
                <Col span={12}>
                  <span className='detail-label'>已收金额：</span>
                  <span style={{ fontSize: 16, color: '#52c41a' }}>
                    ¥{selectedInvoice.total_paid_amount.toLocaleString()}
                  </span>
                </Col>
                <Col span={12}>
                  <span className='detail-label'>剩余金额：</span>
                  <span style={{ fontSize: 16, color: selectedInvoice.remaining_amount > 0 ? '#f5222d' : '#52c41a' }}>
                    ¥{selectedInvoice.remaining_amount.toLocaleString()}
                  </span>
                </Col>
              </Row>
            </Card>

            {/* 开票信息 */}
            <Card title='开票信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={24}>
                  <span className='detail-label'>客户名称：</span>
                  {selectedInvoice.client_name}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>纳税人识别号：</span>
                  {selectedInvoice.client_tax_id || '-'}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>开户银行：</span>
                  {selectedInvoice.client_bank_name || '-'}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>银行账号：</span>
                  {selectedInvoice.client_bank_account || '-'}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>开票地址：</span>
                  {selectedInvoice.client_address || '-'}
                </Col>
              </Row>
            </Card>

            {/* 电子发票信息 */}
            {selectedInvoice.status === 'issued' && (
              <Card title='电子发票信息' size='small'>
                <Row gutter={[16, 8]}>
                  <Col span={24}>
                    <span className='detail-label'>电子发票代码：</span>
                    {selectedInvoice.electronic_invoice_code || '-'}
                  </Col>
                  <Col span={24}>
                    <span className='detail-label'>电子发票号码：</span>
                    {selectedInvoice.electronic_invoice_number || '-'}
                  </Col>
                  <Col span={24}>
                    <span className='detail-label'>下载链接：</span>
                    {selectedInvoice.electronic_invoice_url ? (
                      <a href={selectedInvoice.electronic_invoice_url} target='_blank' rel='noopener noreferrer'>
                        下载电子发票
                      </a>
                    ) : (
                      '-'
                    )}
                  </Col>
                </Row>
              </Card>
            )}
          </div>
        )}
      </Drawer>

      {/* 开票弹窗 */}
      <Modal
        title='开票'
        open={issueModalVisible}
        onOk={handleIssue}
        onCancel={() => setIssueModalVisible(false)}
        destroyOnClose
      >
        <Form form={issueForm} layout='vertical'>
          <Form.Item
            label='电子发票URL'
            name='electronic_url'
            rules={[{ required: true, message: '请输入电子发票URL' }]}
          >
            <Input placeholder='请输入电子发票下载链接' />
          </Form.Item>
          <Form.Item
            label='发票代码'
            name='code'
            rules={[{ required: true, message: '请输入发票代码' }]}
          >
            <Input placeholder='请输入发票代码' />
          </Form.Item>
          <Form.Item
            label='发票号码'
            name='number'
            rules={[{ required: true, message: '请输入发票号码' }]}
          >
            <Input placeholder='请输入发票号码' />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default InvoiceList
