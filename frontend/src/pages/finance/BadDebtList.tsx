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
  EyeOutlined,
  SearchOutlined,
  ReloadOutlined,
  CheckOutlined,
  CloseOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import {
  getBadDebts,
  createBadDebt,
  approveBadDebt,
  rejectBadDebt,
  getBadDebtStats,
  getContracts,
  getInvoices,
  type BadDebt,
  type CreateBadDebtRequest,
  type BadDebtStats,
  badDebtStatusMap,
  badDebtReasonTypeMap,
} from '@/services/finance'

const { Option } = Select
const { TextArea } = Input

interface ContractOption {
  id: number
  contract_code: string
  client_name?: string
  contract_amount: number
}

interface InvoiceOption {
  id: number
  invoice_code: string
  client_name: string
  total_amount: number
  remaining_amount: number
}

// 表格列定义
const createColumns = (handlers: {
  onView: (record: BadDebt) => void
  onApprove: (id: number) => void
  onReject: (id: number) => void
}): ColumnsType<BadDebt> => [
  {
    title: '合同编号',
    key: 'contract',
    width: 150,
    fixed: 'left',
    render: (_, record) => record.contract?.contract_code || '-',
  },
  {
    title: '发票编号',
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
    title: '原始金额',
    dataIndex: 'original_amount',
    key: 'original_amount',
    width: 120,
    render: (amount: number) => (
      <span style={{ fontWeight: 'bold' }}>¥{amount.toLocaleString()}</span>
    ),
  },
  {
    title: '核销金额',
    dataIndex: 'write_off_amount',
    key: 'write_off_amount',
    width: 120,
    render: (amount: number) => (
      <span style={{ fontWeight: 'bold', color: '#f5222d' }}>¥{amount.toLocaleString()}</span>
    ),
  },
  {
    title: '剩余金额',
    dataIndex: 'remaining_amount',
    key: 'remaining_amount',
    width: 120,
    render: (amount: number) => (
      <span style={{ fontWeight: 'bold', color: amount > 0 ? '#52c41a' : '#999' }}>
        ¥{amount.toLocaleString()}
      </span>
    ),
  },
  {
    title: '原因类型',
    dataIndex: 'reason_type',
    key: 'reason_type',
    width: 100,
    render: (type: string) => {
      const config = badDebtReasonTypeMap[type] || { text: type, color: 'default' }
      return <Tag color={config.color}>{config.text}</Tag>
    },
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    render: (status: string) => {
      const config = badDebtStatusMap[status] || { text: status, color: 'default' }
      return <Tag color={config.color}>{config.text}</Tag>
    },
  },
  {
    title: '创建时间',
    dataIndex: 'created_at',
    key: 'created_at',
    width: 150,
    render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm'),
  },
  {
    title: '操作',
    key: 'action',
    width: 180,
    fixed: 'right',
    render: (_, record) => (
      <Space size='small'>
        <Tooltip title='查看详情'>
          <Button type='link' size='small' icon={<EyeOutlined />} onClick={() => handlers.onView(record)} />
        </Tooltip>
        {record.status === 'pending' && (
          <>
            <Tooltip title='审批通过'>
              <Button type='link' size='small' icon={<CheckOutlined />} onClick={() => handlers.onApprove(record.id)} />
            </Tooltip>
            <Popconfirm title='确定要拒绝这条坏账核销申请吗？' onConfirm={() => handlers.onReject(record.id)}>
              <Button type='link' size='small' danger icon={<CloseOutlined />} />
            </Popconfirm>
          </>
        )}
      </Space>
    ),
  },
]

const BadDebtList: React.FC = () => {
  const [badDebts, setBadDebts] = useState<BadDebt[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [modalVisible, setModalVisible] = useState<boolean>(false)
  const [detailDrawerVisible, setDetailDrawerVisible] = useState<boolean>(false)
  const [selectedBadDebt, setSelectedBadDebt] = useState<BadDebt | null>(null)
  const [rejectModalVisible, setRejectModalVisible] = useState<boolean>(false)
  const [rejectBadDebtId, setRejectBadDebtId] = useState<number | null>(null)
  const [stats, setStats] = useState<BadDebtStats | null>(null)
  const [form] = Form.useForm()
  const [rejectForm] = Form.useForm()
  const [contractOptions, setContractOptions] = useState<ContractOption[]>([])
  const [invoiceOptions, setInvoiceOptions] = useState<InvoiceOption[]>([])

  // 查询参数
  const [queryParams, setQueryParams] = useState<{
    page: number
    page_size: number
    status: string
    contract_id?: number
    reason_type: string
  }>({
    page: 1,
    page_size: 10,
    status: '',
    reason_type: '',
  })

  // 搜索表单状态
  const [searchForm, setSearchForm] = useState<{
    status: string
    reason_type: string
    contract_id?: number
  }>({
    status: '',
    reason_type: '',
  })

  const [total, setTotal] = useState<number>(0)

  // 获取坏账核销列表
  const fetchBadDebts = async () => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = { ...queryParams }
      // 移除空值
      Object.keys(params).forEach((key) => {
        if (params[key] === '' || params[key] === undefined) {
          delete params[key]
        }
      })

      const res = await getBadDebts(params)

      let badDebtData: BadDebt[] = []
      let totalCount = 0

      if (res && res.data) {
        if (Array.isArray(res.data)) {
          badDebtData = res.data
          totalCount = res.data.length
        } else if (res.pagination) {
          badDebtData = res.data || []
          totalCount = res.pagination.total || 0
        }
      }

      setBadDebts(badDebtData)
      setTotal(totalCount)

      fetchStats()
    } catch {
      message.error('获取坏账核销列表失败')
    } finally {
      setLoading(false)
    }
  }

  // 获取坏账核销统计
  const fetchStats = async () => {
    try {
      const res = await getBadDebtStats()
      if (res && res.data) {
        setStats(res.data)
      }
    } catch {
      // 静默失败
    }
  }

  // 获取可选合同列表
  const fetchContractOptions = async () => {
    try {
      const res = await getContracts({ page: 1, page_size: 100 })
      if (res?.data) {
        const list = Array.isArray(res.data) ? res.data : []
        setContractOptions(
          list.map((c) => ({
            id: c.id,
            contract_code: c.contract_code,
            client_name: c.client?.name,
            contract_amount: c.contract_amount,
          }))
        )
      }
    } catch {
      // 静默失败
    }
  }

  // 获取可选发票列表（根据合同ID）
  const fetchInvoiceOptions = async (contractId: number) => {
    try {
      const res = await getInvoices({ page: 1, page_size: 100, contract_id: contractId })
      if (res?.data) {
        const list = Array.isArray(res.data) ? res.data : []
        setInvoiceOptions(
          list.map((inv) => ({
            id: inv.id,
            invoice_code: inv.invoice_code,
            client_name: inv.client_name,
            total_amount: inv.total_amount,
            remaining_amount: inv.remaining_amount ?? inv.total_amount,
          }))
        )
      }
    } catch {
      // 静默失败
    }
  }

  useEffect(() => {
    fetchBadDebts()
  }, [queryParams])

  // 打开新增申请弹窗
  const handleAdd = () => {
    form.resetFields()
    setInvoiceOptions([])
    setModalVisible(true)
    fetchContractOptions()
  }

  // 查看详情
  const handleView = (record: BadDebt) => {
    setSelectedBadDebt(record)
    setDetailDrawerVisible(true)
  }

  // 审批通过
  const handleApprove = async (id: number) => {
    try {
      await approveBadDebt(id, '')
      message.success('已审批通过')
      fetchBadDebts()
    } catch {
      message.error('审批失败')
    }
  }

  // 打开拒绝弹窗
  const handleOpenRejectModal = (id: number) => {
    setRejectBadDebtId(id)
    rejectForm.resetFields()
    setRejectModalVisible(true)
  }

  // 拒绝申请
  const handleReject = async () => {
    try {
      const values = await rejectForm.validateFields()
      if (rejectBadDebtId) {
        await rejectBadDebt(rejectBadDebtId, values.notes)
        message.success('已拒绝')
        setRejectModalVisible(false)
        fetchBadDebts()
      }
    } catch {
      message.error('操作失败')
    }
  }

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      const formData: CreateBadDebtRequest = {
        contract_id: values.contract_id,
        invoice_id: values.invoice_id,
        original_amount: values.original_amount || 0,
        write_off_amount: values.write_off_amount,
        reason: values.reason,
        reason_type: values.reason_type,
        attachment_ids: values.attachment_ids,
      }

      await createBadDebt(formData)
      message.success('申请已提交')
      setModalVisible(false)
      fetchBadDebts()
    } catch {
      message.error('操作失败')
    }
  }

  // 合同选择变化
  const handleContractChange = (contractId: number) => {
    form.setFieldValue('invoice_id', undefined)
    setInvoiceOptions([])
    if (contractId) {
      fetchInvoiceOptions(contractId)
    }
  }

  // 搜索
  const handleSearch = () => {
    setQueryParams({
      ...queryParams,
      status: searchForm.status,
      reason_type: searchForm.reason_type,
      contract_id: searchForm.contract_id,
      page: 1,
    })
  }

  // 重置搜索
  const handleReset = () => {
    const resetParams: typeof queryParams = {
      page: 1,
      page_size: 10,
      status: '',
      reason_type: '',
    }
    setSearchForm({
      status: '',
      reason_type: '',
    })
    setQueryParams(resetParams)
  }

  return (
    <div className='bad-debt-list'>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className='stats-row'>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='总申请数'
              value={stats?.total_bad_debts || 0}
              valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='待审批'
              value={stats?.pending_bad_debts || 0}
              valueStyle={{ color: '#faad14', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='已通过'
              value={stats?.approved_bad_debts || 0}
              valueStyle={{ color: '#52c41a', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='核销总金额'
              value={stats?.total_write_off_amount || 0}
              prefix='¥'
              valueStyle={{ color: '#f5222d', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索过滤器 */}
      <Card className='search-card'>
        <Space size='middle' wrap>
          <Select
            placeholder='筛选状态'
            style={{ width: 120 }}
            value={searchForm.status || undefined}
            onChange={(value) => setSearchForm({ ...searchForm, status: value || '' })}
            allowClear
            size='large'
          >
            <Option value='pending'>待审批</Option>
            <Option value='approved'>已通过</Option>
            <Option value='rejected'>已拒绝</Option>
          </Select>

          <Select
            placeholder='原因类型'
            style={{ width: 120 }}
            value={searchForm.reason_type || undefined}
            onChange={(value) => setSearchForm({ ...searchForm, reason_type: value || '' })}
            allowClear
            size='large'
          >
            <Option value='bankruptcy'>破产</Option>
            <Option value='dispute'>纠纷</Option>
            <Option value='uncollectible'>无法收回</Option>
            <Option value='other'>其他</Option>
          </Select>

          <Button type='primary' icon={<SearchOutlined />} onClick={handleSearch} size='large'>
            搜索
          </Button>

          <Button icon={<ReloadOutlined />} onClick={handleReset} size='large'>
            重置筛选
          </Button>
        </Space>
      </Card>

      {/* 坏账核销列表 */}
      <Card
        title='坏账核销列表'
        extra={
          <Button type='primary' icon={<PlusOutlined />} onClick={handleAdd}>
            新增申请
          </Button>
        }
      >
        <Table
          columns={createColumns({
            onView: handleView,
            onApprove: handleApprove,
            onReject: handleOpenRejectModal,
          })}
          dataSource={badDebts}
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

      {/* 新增申请弹窗 */}
      <Modal
        title='新增坏账核销申请'
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={700}
      >
        <Form form={form} layout='vertical'>
          <Form.Item
            label='关联合同'
            name='contract_id'
            rules={[{ required: true, message: '请选择关联合同' }]}
          >
            <Select
              showSearch
              placeholder='请选择关联合同'
              filterOption={(input, option) =>
                (option?.label as string ?? '').toLowerCase().includes(input.toLowerCase())
              }
              onChange={handleContractChange}
              options={contractOptions.map((c) => ({
                value: c.id,
                label: `${c.contract_code}${c.client_name ? ` - ${c.client_name}` : ''}`,
              }))}
            />
          </Form.Item>

          <Form.Item label='关联发票（可选）' name='invoice_id'>
            <Select
              showSearch
              placeholder='请选择关联发票（不选则核销整个合同）'
              filterOption={(input, option) =>
                (option?.label as string ?? '').toLowerCase().includes(input.toLowerCase())
              }
              options={invoiceOptions.map((inv) => ({
                value: inv.id,
                label: `${inv.invoice_code} - ${inv.client_name} (¥${inv.remaining_amount.toLocaleString()})`,
              }))}
            />
          </Form.Item>

          <Form.Item
            label='原始应收金额'
            name='original_amount'
            rules={[{ required: true, message: '请输入原始应收金额' }]}
          >
            <InputNumber
              placeholder='请输入原始应收金额'
              style={{ width: '100%' }}
              min={0}
              precision={2}
            />
          </Form.Item>

          <Form.Item
            label='核销金额'
            name='write_off_amount'
            rules={[
              { required: true, message: '请输入核销金额' },
              { type: 'number', min: 0.01, message: '核销金额必须大于0' },
            ]}
          >
            <InputNumber placeholder='请输入核销金额' style={{ width: '100%' }} min={0} precision={2} />
          </Form.Item>

          <Form.Item
            label='原因类型'
            name='reason_type'
            rules={[{ required: true, message: '请选择原因类型' }]}
          >
            <Select placeholder='请选择原因类型'>
              <Option value='bankruptcy'>破产</Option>
              <Option value='dispute'>纠纷</Option>
              <Option value='uncollectible'>无法收回</Option>
              <Option value='other'>其他</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label='详细原因'
            name='reason'
            rules={[
              { required: true, message: '请输入详细原因' },
              { max: 1000, message: '原因不能超过1000字符' },
            ]}
          >
            <TextArea placeholder='请输入详细原因' rows={4} />
          </Form.Item>

          <Form.Item label='附件ID（可选）' name='attachment_ids'>
            <Select mode='multiple' placeholder='请输入附件ID' tokenSeparators={[',']} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 详情抽屉 */}
      <Drawer
        title='坏账核销详情'
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
        width={600}
      >
        {selectedBadDebt && (
          <div className='bad-debt-detail'>
            <Card title='基本信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>合同编号：</span>
                  {selectedBadDebt.contract?.contract_code || '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>状态：</span>
                  {(() => {
                    const config = badDebtStatusMap[selectedBadDebt.status] || {
                      text: selectedBadDebt.status,
                      color: 'default',
                    }
                    return <Tag color={config.color}>{config.text}</Tag>
                  })()}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>发票编号：</span>
                  {selectedBadDebt.invoice?.invoice_code || '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>客户名称：</span>
                  {selectedBadDebt.invoice?.client_name || '-'}
                </Col>
              </Row>
            </Card>

            <Card title='金额信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={24}>
                  <span className='detail-label'>原始金额：</span>
                  <span style={{ fontSize: 18, fontWeight: 'bold' }}>
                    ¥{selectedBadDebt.original_amount.toLocaleString()}
                  </span>
                </Col>
                <Col span={24}>
                  <span className='detail-label'>核销金额：</span>
                  <span style={{ fontSize: 20, fontWeight: 'bold', color: '#f5222d' }}>
                    ¥{selectedBadDebt.write_off_amount.toLocaleString()}
                  </span>
                </Col>
                <Col span={24}>
                  <span className='detail-label'>剩余金额：</span>
                  <span
                    style={{
                      fontSize: 18,
                      fontWeight: 'bold',
                      color: selectedBadDebt.remaining_amount > 0 ? '#52c41a' : '#999',
                    }}
                  >
                    ¥{selectedBadDebt.remaining_amount.toLocaleString()}
                  </span>
                </Col>
              </Row>
            </Card>

            <Card title='核销原因' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>原因类型：</span>
                  {(() => {
                    const config = badDebtReasonTypeMap[selectedBadDebt.reason_type] || {
                      text: selectedBadDebt.reason_type,
                      color: 'default',
                    }
                    return <Tag color={config.color}>{config.text}</Tag>
                  })()}
                </Col>
                <Col span={24}>
                  <span className='detail-label'>详细原因：</span>
                  {selectedBadDebt.reason || '-'}
                </Col>
              </Row>
            </Card>

            <Card title='其他信息' size='small'>
              <Row gutter={[16, 8]}>
                <Col span={24}>
                  <span className='detail-label'>审批备注：</span>
                  {selectedBadDebt.approval_notes || '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>创建时间：</span>
                  {dayjs(selectedBadDebt.created_at).format('YYYY-MM-DD HH:mm:ss')}
                </Col>
                {selectedBadDebt.approved_at && (
                  <Col span={12}>
                    <span className='detail-label'>审批时间：</span>
                    {dayjs(selectedBadDebt.approved_at).format('YYYY-MM-DD HH:mm:ss')}
                  </Col>
                )}
              </Row>
            </Card>
          </div>
        )}
      </Drawer>

      {/* 拒绝弹窗 */}
      <Modal
        title='拒绝坏账核销申请'
        open={rejectModalVisible}
        onOk={handleReject}
        onCancel={() => setRejectModalVisible(false)}
      >
        <Form form={rejectForm} layout='vertical'>
          <Form.Item
            label='拒绝原因'
            name='notes'
            rules={[{ required: true, message: '请输入拒绝原因' }]}
          >
            <TextArea placeholder='请输入拒绝原因' rows={4} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default BadDebtList
