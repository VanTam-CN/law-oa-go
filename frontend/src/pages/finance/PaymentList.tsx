import React, { useState, useEffect, useMemo } from 'react'
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
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SearchOutlined,
  ReloadOutlined,
  CheckOutlined,
  CloseOutlined,
  DownloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import {
  getPayments,
  getInvoices,
  createPayment,
  updatePayment,
  deletePayment,
  confirmPayment,
  rejectPayment,
  getPaymentStats,
  exportPayments,
  type Payment,
  type CreatePaymentRequest,
  type UpdatePaymentRequest,
  type PaymentStats,
  paymentStatusMap,
} from '@/services/finance'
import { exportCSV } from '@/utils/export'
import FilterSelect from './FilterSelect'
import './PaymentList.less'

const { Option } = Select
const { TextArea } = Input
const { RangePicker } = DatePicker

interface InvoiceOption {
  id: number
  invoice_code: string
  client_name: string
  total_amount: number
  remaining_amount: number
}

// 表格列定义 - 使用render props接收handlers
const createColumns = (handlers: {
  onView: (record: Payment) => void
  onEdit: (record: Payment) => void
  onConfirm: (id: number) => void
  onReject: (id: number) => void
  onDelete: (id: number) => void
}): ColumnsType<Payment> => [
  {
    title: '回款编号',
    dataIndex: 'payment_code',
    key: 'payment_code',
    width: 150,
    fixed: 'left',
  },
  {
    title: '关联发票',
    key: 'invoice',
    width: 150,
    render: (_, record) => record.invoice?.invoice_code || '-',
  },
  {
    title: '客户名称',
    key: 'client_name',
    width: 150,
    render: (_, record) => record.invoice?.client_name || '-',
  },
  {
    title: '回款金额',
    dataIndex: 'amount',
    key: 'amount',
    width: 120,
    render: (amount: number) => (
      <span style={{ fontWeight: 'bold', color: '#f5222d' }}>¥{amount.toLocaleString()}</span>
    ),
  },
  {
    title: '付款方式',
    dataIndex: 'payment_method',
    key: 'payment_method',
    width: 100,
    render: (method: string) => {
      const methodMap: Record<string, string> = {
        bank_transfer: '银行转账',
        cash: '现金',
        other: '其他',
      }
      return methodMap[method] || method
    },
  },
  {
    title: '付款方',
    dataIndex: 'payer_name',
    key: 'payer_name',
    width: 120,
  },
  {
    title: '付款方账号',
    dataIndex: 'payer_account',
    key: 'payer_account',
    width: 150,
    ellipsis: true,
  },
  {
    title: '回款日期',
    dataIndex: 'payment_date',
    key: 'payment_date',
    width: 120,
    render: (date: string) => dayjs(date).format('YYYY-MM-DD'),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    render: (status: string) => {
      const config = paymentStatusMap[status] || { text: status, color: 'default' }
      return <Tag color={config.color}>{config.text}</Tag>
    },
  },
  {
    title: '确认时间',
    dataIndex: 'confirmed_at',
    key: 'confirmed_at',
    width: 120,
    render: (date: string) => (date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'),
  },
  {
    title: '备注',
    dataIndex: 'remark',
    key: 'remark',
    width: 150,
    ellipsis: true,
  },
  {
    title: '操作',
    key: 'action',
    width: 180,
    fixed: 'right',
    render: (_, record) => (
      <Space size='small'>
        <Tooltip title='查看详情'>
          <Button type='link' size='small' icon={<EyeOutlined />} onClick={(e) => { e.stopPropagation(); handlers.onView(record) }} />
        </Tooltip>
        {record.status === 'pending' && (
          <>
            <Tooltip title='编辑'>
              <Button type='link' size='small' icon={<EditOutlined />} onClick={(e) => { e.stopPropagation(); handlers.onEdit(record) }} />
            </Tooltip>
            <Tooltip title='确认回款'>
              <Button type='link' size='small' icon={<CheckOutlined />} onClick={(e) => { e.stopPropagation(); handlers.onConfirm(record.id) }} />
            </Tooltip>
            <Tooltip title='拒绝'>
              <Button type='link' size='small' danger icon={<CloseOutlined />} onClick={(e) => { e.stopPropagation(); handlers.onReject(record.id) }} />
            </Tooltip>
            <Popconfirm title='确定要删除这条回款记录吗？' onConfirm={(e) => { e?.stopPropagation(); handlers.onDelete(record.id) }}>
              <Button type='link' size='small' danger icon={<DeleteOutlined />} onClick={(e) => e.stopPropagation()} />
            </Popconfirm>
          </>
        )}
      </Space>
    ),
  },
]

const PaymentList: React.FC = () => {
  const [payments, setPayments] = useState<Payment[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [modalVisible, setModalVisible] = useState<boolean>(false)
  const [modalTitle, setModalTitle] = useState<string>('')
  const [editingPayment, setEditingPayment] = useState<Payment | null>(null)
  const [stats, setStats] = useState<PaymentStats | null>(null)
  const [detailDrawerVisible, setDetailDrawerVisible] = useState<boolean>(false)
  const [selectedPayment, setSelectedPayment] = useState<Payment | null>(null)
  const [rejectModalVisible, setRejectModalVisible] = useState<boolean>(false)
  const [rejectPaymentId, setRejectPaymentId] = useState<number | null>(null)
  const [form] = Form.useForm()
  const [rejectForm] = Form.useForm()
  const [invoiceOptions, setInvoiceOptions] = useState<InvoiceOption[]>([])
  const [invoiceLoading, setInvoiceLoading] = useState(false)

  // 查询参数
  const [queryParams, setQueryParams] = useState<{
    page: number
    page_size: number
    status: string
    invoice_id?: number
    date_from: string
    date_to: string
  }>({
    page: 1,
    page_size: 10,
    status: '',
    invoice_id: undefined,
    date_from: '',
    date_to: '',
  })

  // 搜索表单状态
  const [searchForm, setSearchForm] = useState<{
    status: string
    dateRange: [dayjs.Dayjs, dayjs.Dayjs] | null
  }>({
    status: '',
    dateRange: null,
  })

  const [total, setTotal] = useState<number>(0)

  // 获取回款列表
  const fetchPayments = async () => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { ...queryParams }
      // 移除空值
      Object.keys(params).forEach((key) => {
        if (params[key] === '') {
          delete params[key]
        }
      })

      const res = await getPayments(params)

      let paymentData: Payment[] = []
      let totalCount = 0

      if (res && res.data) {
        if (Array.isArray(res.data)) {
          paymentData = res.data
          totalCount = res.data.length
        } else if (res.pagination) {
          paymentData = res.data || []
          totalCount = res.pagination.total || 0
        }
      }

      setPayments(paymentData)
      setTotal(totalCount)

      fetchStats()
    } catch {
      message.error('获取回款列表失败')
    } finally {
      setLoading(false)
    }
  }

  // 获取回款统计
  const fetchStats = async () => {
    try {
      const res = await getPaymentStats()
      if (res && res.data) {
        setStats(res.data)
      }
    } catch (error) {
      // 静默失败
    }
  }

  // 获取可选发票列表（已开票/已签收状态）
  const fetchInvoiceOptions = async () => {
    setInvoiceLoading(true)
    try {
      const [res, res2] = await Promise.all([
        getInvoices({ page: 1, page_size: 100, status: 'issued' }),
        getInvoices({ page: 1, page_size: 100, status: 'received' }),
      ])
      const invoiceMap = new Map<number, InvoiceOption>()
      const addInvoices = (r: any) => {
        if (r?.data) {
          const list = Array.isArray(r.data) ? r.data : r.pagination ? r.data || [] : []
          for (const inv of list) {
            if (!invoiceMap.has(inv.id)) {
              invoiceMap.set(inv.id, {
                id: inv.id,
                invoice_code: inv.invoice_code,
                client_name: inv.client_name || '-',
                total_amount: inv.total_amount,
                remaining_amount: inv.remaining_amount ?? inv.total_amount,
              })
            }
          }
        }
      }
      addInvoices(res)
      addInvoices(res2)
      setInvoiceOptions(Array.from(invoiceMap.values()))
    } catch {
      // 静默失败
    } finally {
      setInvoiceLoading(false)
    }
  }

  useEffect(() => {
    fetchPayments()
  }, [queryParams])

  // 打开新增回款弹窗
  const handleAdd = () => {
    setModalTitle('新增回款')
    setEditingPayment(null)
    form.resetFields()
    form.setFieldsValue({
      payment_method: 'bank_transfer',
      payment_date: dayjs(),
    })
    setModalVisible(true)
    fetchInvoiceOptions()
  }

  // 打开编辑回款弹窗
  const handleEdit = (record: Payment) => {
    setModalTitle('编辑回款')
    setEditingPayment(record)
    const formValues = {
      ...record,
      payment_date: record.payment_date ? dayjs(record.payment_date) : null,
    }
    form.setFieldsValue(formValues)
    setModalVisible(true)
  }

  // 查看回款详情
  const handleView = (record: Payment) => {
    setSelectedPayment(record)
    setDetailDrawerVisible(true)
  }

  // 删除回款
  const handleDelete = async (id: number) => {
    try {
      await deletePayment(id)
      message.success('删除成功')
      fetchPayments()
    } catch (error) {
      message.error('删除失败')
    }
  }

  // 确认回款
  const handleConfirm = async (id: number) => {
    try {
      await confirmPayment(id)
      message.success('回款已确认')
      fetchPayments()
    } catch (error) {
      message.error('确认失败')
    }
  }

  // 打开拒绝弹窗
  const handleOpenRejectModal = (id: number) => {
    setRejectPaymentId(id)
    rejectForm.resetFields()
    setRejectModalVisible(true)
  }

  // 拒绝回款
  const handleReject = async () => {
    try {
      if (rejectPaymentId) {
        await rejectPayment(rejectPaymentId)
        message.success('已拒绝')
        setRejectModalVisible(false)
        fetchPayments()
      }
    } catch (error) {
      message.error('操作失败')
    }
  }

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      const formData: CreatePaymentRequest | UpdatePaymentRequest = {
        ...values,
        payment_date: values.payment_date ? values.payment_date.format('YYYY-MM-DD') : undefined,
      }

      if (editingPayment) {
        await updatePayment(editingPayment.id, formData as UpdatePaymentRequest)
        message.success('更新成功')
      } else {
        await createPayment(formData as CreatePaymentRequest)
        message.success('新增成功')
      }
      setModalVisible(false)
      fetchPayments()
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
      date_from: date_from ? date_from.format('YYYY-MM-DD') : '',
      date_to: date_to ? date_to.format('YYYY-MM-DD') : '',
      page: 1,
      invoice_id: queryParams.invoice_id,
    })
  }

  // 重置搜索
  const handleReset = () => {
    const resetParams: typeof queryParams = {
      page: 1,
      page_size: 10,
      status: '',
      date_from: '',
      date_to: '',
      invoice_id: undefined,
    }
    setSearchForm({
      status: '',
      dateRange: null,
    })
    setQueryParams(resetParams)
  }

  // 导出报表
  const handleExport = async () => {
    try {
      const params: Record<string, unknown> = {
        status: queryParams.status,
        date_from: queryParams.date_from,
        date_to: queryParams.date_to,
        invoice_id: queryParams.invoice_id,
      }
      // 移除空值
      Object.keys(params).forEach((key) => {
        if (params[key] === '' || params[key] === undefined) {
          delete params[key]
        }
      })

      await exportCSV(
        exportPayments(params),
        `回款报表_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.csv`
      )
      message.success('导出成功')
    } catch (error) {
      message.error('导出失败')
    }
  }

  return (
    <div className='payment-list'>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className='stats-row'>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='回款总数'
              value={stats?.total_payments || 0}
              valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='待确认'
              value={stats?.pending_payments || 0}
              valueStyle={{ color: '#faad14', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='本月回款'
              value={stats?.month_payment_amount || 0}
              prefix='¥'
              valueStyle={{ color: '#3f8600', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='待确认金额'
              value={stats?.pending_amount || 0}
              prefix='¥'
              valueStyle={{ color: '#cf1322', fontSize: '24px', fontWeight: 700 }}
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
              { value: 'pending', label: '待确认' },
              { value: 'confirmed', label: '已确认' },
              { value: 'rejected', label: '已拒绝' },
            ]}
          />

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

      {/* 回款列表 */}
      <Card
        title='回款列表'
        extra={
          <Button type='primary' icon={<PlusOutlined />} onClick={handleAdd}>
            新增回款
          </Button>
        }
      >
        <Table
          columns={useMemo(() => createColumns({
            onView: handleView,
            onEdit: handleEdit,
            onConfirm: handleConfirm,
            onReject: handleOpenRejectModal,
            onDelete: handleDelete,
          }).map((col) => ({
            ...col,
            onCell: (record: Payment) => ({
              onClick: () => handleView(record),
              style: { cursor: 'pointer' },
            }),
          })), [handleView])}
          dataSource={payments}
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

      {/* 新增/编辑弹窗 */}
      <Modal
        title={modalTitle}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={700}
        destroyOnClose
      >
        <Form form={form} layout='vertical'>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label='关联发票'
                name='invoice_id'
                rules={[{ required: true, message: '请选择关联发票' }]}
              >
                <Select
                  showSearch
                  placeholder='请选择关联发票'
                  loading={invoiceLoading}
                  filterOption={(input, option) =>
                    (option?.label as string ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={invoiceOptions.map((inv) => ({
                    value: inv.id,
                    label: `${inv.invoice_code} - ${inv.client_name} (¥${inv.remaining_amount.toLocaleString()})`,
                  }))}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label='回款金额'
                name='amount'
                rules={[{ required: true, message: '请输入回款金额' }]}
              >
                <InputNumber
                  placeholder='请输入回款金额'
                  style={{ width: '100%' }}
                  min={0}
                  precision={2}
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label='回款日期'
                name='payment_date'
                rules={[{ required: true, message: '请选择回款日期' }]}
              >
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label='付款方式'
                name='payment_method'
                rules={[{ required: true, message: '请选择付款方式' }]}
              >
                <Select>
                  <Option value='bank_transfer'>银行转账</Option>
                  <Option value='cash'>现金</Option>
                  <Option value='other'>其他</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label='交易流水号' name='reference_no'>
                <Input placeholder='请输入交易流水号' />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label='付款方名称' name='payer_name'>
                <Input placeholder='请输入付款方名称' />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item label='付款方账号' name='payer_account'>
            <Input placeholder='请输入付款方账号' />
          </Form.Item>

          <Form.Item label='附件ID' name='attachment_id'>
            <InputNumber placeholder='可选，上传回款凭证' style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item label='备注' name='remark'>
            <TextArea placeholder='请输入备注信息' rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 回款详情抽屉 */}
      <Drawer
        title='回款详情'
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
        width={600}
      >
        {selectedPayment && (
          <div className='payment-detail'>
            <Card title='基本信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>回款编号：</span>
                  {selectedPayment.payment_code}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>回款状态：</span>
                  {(() => {
                    const config = paymentStatusMap[selectedPayment.status] || {
                      text: selectedPayment.status,
                      color: 'default',
                    }
                    return <Tag color={config.color}>{config.text}</Tag>
                  })()}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>关联发票：</span>
                  {selectedPayment.invoice?.invoice_code || '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>客户名称：</span>
                  {selectedPayment.invoice?.client_name || '-'}
                </Col>
              </Row>
            </Card>

            <Card title='金额信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>回款金额：</span>
                  <span style={{ fontSize: 20, fontWeight: 'bold', color: '#f5222d' }}>
                    ¥{selectedPayment.amount.toLocaleString()}
                  </span>
                </Col>
                <Col span={12}>
                  <span className='detail-label'>发票总额：</span>
                  <span>
                    ¥{selectedPayment.invoice?.total_amount?.toLocaleString() || 0}
                  </span>
                </Col>
              </Row>
            </Card>

            <Card title='付款信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>付款方式：</span>
                  {selectedPayment.payment_method === 'bank_transfer'
                    ? '银行转账'
                    : selectedPayment.payment_method === 'cash'
                      ? '现金'
                      : '其他'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>回款日期：</span>
                  {dayjs(selectedPayment.payment_date).format('YYYY-MM-DD')}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>付款方名称：</span>
                  {selectedPayment.payer_name || '-'}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>付款方账号：</span>
                  {selectedPayment.payer_account || '-'}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>交易流水号：</span>
                  {selectedPayment.reference_no || '-'}
                </Col>
              </Row>
            </Card>

            <Card title='其他信息' size='small'>
              <Row gutter={[16, 8]}>
                <Col span={24}>
                  <span className='detail-label'>备注：</span>
                  {selectedPayment.remark || '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>创建时间：</span>
                  {dayjs(selectedPayment.created_at).format('YYYY-MM-DD HH:mm:ss')}
                </Col>
                {selectedPayment.confirmed_at && (
                  <Col span={12}>
                    <span className='detail-label'>确认时间：</span>
                    {dayjs(selectedPayment.confirmed_at).format('YYYY-MM-DD HH:mm:ss')}
                  </Col>
                )}
              </Row>
            </Card>
          </div>
        )}
      </Drawer>

      {/* 拒绝回款弹窗 */}
      <Modal
        title='拒绝回款'
        open={rejectModalVisible}
        onOk={handleReject}
        onCancel={() => setRejectModalVisible(false)}
        destroyOnClose
      >
        <Form form={rejectForm} layout='vertical'>
          <Form.Item label='拒绝原因' name='notes' rules={[{ required: true, message: '请输入拒绝原因' }]}>
            <TextArea placeholder='请输入拒绝原因' rows={4} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default PaymentList
