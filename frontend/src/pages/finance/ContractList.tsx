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
  Divider,
  Radio,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SearchOutlined,
  ReloadOutlined,
  CheckOutlined,
  StopOutlined,
  TransactionOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import {
  getContracts,
  getContract,
  createContract,
  updateContract,
  deleteContract,
  activateContract,
  suspendContract,
  completeContract,
  getContractStats,
  getContractMilestones,
  createMilestone,
  updateMilestone,
  deleteMilestone,
  type Contract,
  type CreateContractRequest,
  type UpdateContractRequest,
  type ContractStats,
  type PaymentMilestone,
  type CreateMilestoneRequest,
  contractStatusMap,
  formatAmount,
} from '@/services/finance'
import './ContractList.less'

const { Option } = Select
const { TextArea } = Input
const { RangePicker } = DatePicker

// 表格列定义
const columns: ColumnsType<Contract> = [
  {
    title: '合同编号',
    dataIndex: 'contract_code',
    key: 'contract_code',
    width: 150,
    fixed: 'left',
  },
  {
    title: '客户名称',
    key: 'client_name',
    width: 150,
    render: (_, record) => record.client?.name || '-',
  },
  {
    title: '案件名称',
    key: 'case_title',
    width: 200,
    ellipsis: true,
    render: (_, record) => record.case?.title || '-',
  },
  {
    title: '合同金额',
    dataIndex: 'contract_amount',
    key: 'contract_amount',
    width: 130,
    render: (amount: number, record) => (
      <span>
        {record.currency === 'USD' ? '$' : '¥'}
        {amount.toLocaleString()}
      </span>
    ),
  },
  {
    title: '合同类型',
    dataIndex: 'contract_type',
    key: 'contract_type',
    width: 100,
    render: (type: string) => (
      <Tag color={type === 'supplementary' ? 'blue' : 'green'}>
        {type === 'supplementary' ? '补充合同' : '原始合同'}
      </Tag>
    ),
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    render: (status: string) => {
      const config = contractStatusMap[status] || { text: status, color: 'default' }
      return <Tag color={config.color}>{config.text}</Tag>
    },
  },
  {
    title: '签订日期',
    dataIndex: 'signed_at',
    key: 'signed_at',
    width: 120,
    render: (date: string) => (date ? dayjs(date).format('YYYY-MM-DD') : '-'),
  },
  {
    title: '合同期限',
    key: 'period',
    width: 180,
    render: (_, record) => {
      if (!record.start_date && !record.end_date) return '-'
      const start = record.start_date ? dayjs(record.start_date).format('YYYY-MM-DD') : '?'
      const end = record.end_date ? dayjs(record.end_date).format('YYYY-MM-DD') : '?'
      return `${start} ~ ${end}`
    },
  },
  {
    title: '付款方式',
    dataIndex: 'payment_terms',
    key: 'payment_terms',
    width: 120,
    ellipsis: true,
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    fixed: 'right',
    render: (_, record) => (
      <Space size='small'>
        <Tooltip title='查看详情'>
          <Button type='link' size='small' icon={<EyeOutlined />} onClick={() => {}} />
        </Tooltip>
        {record.status === 'draft' && (
          <>
            <Tooltip title='编辑'>
              <Button type='link' size='small' icon={<EditOutlined />} onClick={() => {}} />
            </Tooltip>
            <Popconfirm title='确定要删除这个合同吗？' onConfirm={() => {}}>
              <Button type='link' size='small' danger icon={<DeleteOutlined />} />
            </Popconfirm>
            <Tooltip title='激活合同'>
              <Button
                type='link'
                size='small'
                icon={<CheckOutlined />}
                onClick={() => {}}
              />
            </Tooltip>
          </>
        )}
        {record.status === 'active' && (
          <>
            <Tooltip title='暂停合同'>
              <Button
                type='link'
                size='small'
                icon={<StopOutlined />}
                onClick={() => {}}
              />
            </Tooltip>
            <Tooltip title='完成合同'>
              <Button
                type='link'
                size='small'
                icon={<TransactionOutlined />}
                onClick={() => {}}
              />
            </Tooltip>
          </>
        )}
        {record.status === 'suspended' && (
          <Tooltip title='恢复合同'>
            <Button type='link' size='small' icon={<CheckOutlined />} onClick={() => {}} />
          </Tooltip>
        )}
      </Space>
    ),
  },
]

// 付款计划表格列
const milestoneColumns: ColumnsType<PaymentMilestone> = [
  {
    title: '序号',
    dataIndex: 'sequence',
    key: 'sequence',
    width: 60,
  },
  {
    title: '付款阶段',
    dataIndex: 'name',
    key: 'name',
    width: 150,
  },
  {
    title: '金额',
    dataIndex: 'amount',
    key: 'amount',
    width: 120,
    render: (amount: number) => `¥${amount.toLocaleString()}`,
  },
  {
    title: '比例',
    dataIndex: 'percentage',
    key: 'percentage',
    width: 80,
    render: (percentage: number) => `${percentage}%`,
  },
  {
    title: '计划付款日期',
    dataIndex: 'due_date',
    key: 'due_date',
    width: 120,
    render: (date: string) => (date ? dayjs(date).format('YYYY-MM-DD') : '-'),
  },
  {
    title: '付款条件',
    dataIndex: 'condition',
    key: 'condition',
    width: 200,
    ellipsis: true,
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    render: (status: string) => {
      const statusMap: Record<string, { text: string; color: string }> = {
        pending: { text: '待开票', color: 'default' },
        billed: { text: '已开票', color: 'processing' },
        partial_paid: { text: '部分付款', color: 'warning' },
        paid: { text: '已付清', color: 'success' },
        overdue: { text: '逾期', color: 'error' },
      }
      const config = statusMap[status] || { text: status, color: 'default' }
      return <Tag color={config.color}>{config.text}</Tag>
    },
  },
  {
    title: '已付金额',
    dataIndex: 'paid_amount',
    key: 'paid_amount',
    width: 100,
    render: (amount: number) => `¥${amount.toLocaleString()}`,
  },
  {
    title: '操作',
    key: 'action',
    width: 120,
    render: (_, record) => (
      <Space size='small'>
        <Button type='link' size='small'>
          编辑
        </Button>
        <Popconfirm title='确定要删除这个付款计划吗？' onConfirm={() => {}}>
          <Button type='link' size='small' danger>
            删除
          </Button>
        </Popconfirm>
      </Space>
    ),
  },
]

const ContractList: React.FC = () => {
  const [contracts, setContracts] = useState<Contract[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [modalVisible, setModalVisible] = useState<boolean>(false)
  const [modalTitle, setModalTitle] = useState<string>('')
  const [editingContract, setEditingContract] = useState<Contract | null>(null)
  const [stats, setStats] = useState<ContractStats | null>(null)
  const [detailDrawerVisible, setDetailDrawerVisible] = useState<boolean>(false)
  const [selectedContract, setSelectedContract] = useState<Contract | null>(null)
  const [milestones, setMilestones] = useState<PaymentMilestone[]>([])
  const [milestoneModalVisible, setMilestoneModalVisible] = useState<boolean>(false)
  const [editingMilestone, setEditingMilestone] = useState<PaymentMilestone | null>(null)
  const [form] = Form.useForm()
  const [milestoneForm] = Form.useForm()

  // 查询参数
  const [queryParams, setQueryParams] = useState<{
    page: number
    page_size: number
    status: string
    contract_type: string
    client_id?: number
    case_id?: number
    search: string
  }>({
    page: 1,
    page_size: 10,
    status: '',
    contract_type: '',
    client_id: undefined,
    search: '',
  })

  // 搜索表单状态
  const [searchForm, setSearchForm] = useState({
    search: '',
    status: '',
    contract_type: '',
  })

  const [total, setTotal] = useState<number>(0)

  // 获取合同列表
  const fetchContracts = async () => {
    setLoading(true)
    try {
      const params = { ...queryParams }
      // 移除空值
      Object.keys(params).forEach((key) => {
        if (params[key as keyof typeof params] === '') {
          delete params[key as keyof typeof params]
        }
      })

      const res = await getContracts(params)
      console.log('合同列表API响应:', res)

      let contractData: Contract[] = []
      let totalCount = 0

      if (res && res.data) {
        if (Array.isArray(res.data)) {
          contractData = res.data
          totalCount = res.data.length
        } else if (res.pagination) {
          contractData = res.data || []
          totalCount = res.pagination.total || 0
        }
      }

      setContracts(contractData)
      setTotal(totalCount)

      // 获取统计数据
      fetchStats()
    } catch (error) {
      console.error('获取合同列表失败:', error)
      message.error('获取合同列表失败')
    } finally {
      setLoading(false)
    }
  }

  // 获取合同统计
  const fetchStats = async () => {
    try {
      const res = await getContractStats()
      console.log('合同统计API响应:', res)

      if (res && res.data) {
        setStats(res.data)
      }
    } catch (error) {
      console.error('获取统计数据失败:', error)
    }
  }

  // 获取合同的付款计划
  const fetchMilestones = async (contractId: number) => {
    try {
      const res = await getContractMilestones(contractId)
      if (res && res.data) {
        setMilestones(Array.isArray(res.data) ? res.data : [])
      }
    } catch (error) {
      console.error('获取付款计划失败:', error)
    }
  }

  useEffect(() => {
    fetchContracts()
  }, [queryParams])

  // 打开新增合同弹窗
  const handleAdd = () => {
    setModalTitle('新增合同')
    setEditingContract(null)
    form.resetFields()
    form.setFieldsValue({
      currency: 'CNY',
      billing_cycle: '一次性',
      payment_terms: '一次性付款',
      contract_type: 'original',
    })
    setModalVisible(true)
  }

  // 打开编辑合同弹窗
  const handleEdit = (record: Contract) => {
    setModalTitle('编辑合同')
    setEditingContract(record)
    const formValues = {
      ...record,
      start_date: record.start_date ? dayjs(record.start_date) : null,
      end_date: record.end_date ? dayjs(record.end_date) : null,
      signed_at: record.signed_at ? dayjs(record.signed_at) : null,
    }
    form.setFieldsValue(formValues)
    setModalVisible(true)
  }

  // 查看合同详情
  const handleView = async (record: Contract) => {
    setSelectedContract(record)
    setDetailDrawerVisible(true)
    await fetchMilestones(record.id)
  }

  // 删除合同
  const handleDelete = async (id: number) => {
    try {
      await deleteContract(id)
      message.success('删除成功')
      fetchContracts()
    } catch (error) {
      message.error('删除失败')
    }
  }

  // 激活合同
  const handleActivate = async (id: number) => {
    try {
      await activateContract(id)
      message.success('合同已激活')
      fetchContracts()
    } catch (error) {
      message.error('激活失败')
    }
  }

  // 暂停合同
  const handleSuspend = async (id: number) => {
    try {
      await suspendContract(id)
      message.success('合同已暂停')
      fetchContracts()
    } catch (error) {
      message.error('暂停失败')
    }
  }

  // 完成合同
  const handleComplete = async (id: number) => {
    try {
      await completeContract(id)
      message.success('合同已完成')
      fetchContracts()
    } catch (error) {
      message.error('操作失败')
    }
  }

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()

      // 处理日期格式
      const formData: CreateContractRequest | UpdateContractRequest = {
        ...values,
        start_date: values.start_date ? values.start_date.format('YYYY-MM-DD') : undefined,
        end_date: values.end_date ? values.end_date.format('YYYY-MM-DD') : undefined,
      }

      if (editingContract) {
        await updateContract(editingContract.id, formData as UpdateContractRequest)
        message.success('更新成功')
      } else {
        await createContract(formData as CreateContractRequest)
        message.success('新增成功')
      }
      setModalVisible(false)
      fetchContracts()
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  // 打开新增付款计划弹窗
  const handleAddMilestone = () => {
    setEditingMilestone(null)
    milestoneForm.resetFields()
    // 自动设置序号
    const nextSequence = milestones.length + 1
    milestoneForm.setFieldsValue({
      contract_id: selectedContract?.id,
      sequence: nextSequence,
    })
    setMilestoneModalVisible(true)
  }

  // 编辑付款计划
  const handleEditMilestone = (record: PaymentMilestone) => {
    setEditingMilestone(record)
    const formValues = {
      ...record,
      due_date: record.due_date ? dayjs(record.due_date) : null,
    }
    milestoneForm.setFieldsValue(formValues)
    setMilestoneModalVisible(true)
  }

  // 删除付款计划
  const handleDeleteMilestone = async (id: number) => {
    try {
      await deleteMilestone(id)
      message.success('删除成功')
      if (selectedContract) {
        await fetchMilestones(selectedContract.id)
      }
    } catch (error) {
      message.error('删除失败')
    }
  }

  // 提交付款计划表单
  const handleMilestoneSubmit = async () => {
    try {
      const values = await milestoneForm.validateFields()

      const formData: CreateMilestoneRequest = {
        ...values,
        due_date: values.due_date ? values.due_date.format('YYYY-MM-DD') : undefined,
      }

      if (editingMilestone) {
        await updateMilestone(editingMilestone.id, formData)
        message.success('更新成功')
      } else {
        await createMilestone(formData)
        message.success('新增成功')
      }
      setMilestoneModalVisible(false)
      if (selectedContract) {
        await fetchMilestones(selectedContract.id)
      }
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  // 搜索
  const handleSearch = () => {
    setQueryParams({
      ...queryParams,
      search: searchForm.search,
      status: searchForm.status,
      contract_type: searchForm.contract_type,
      page: 1,
    })
  }

  // 重置搜索
  const handleReset = () => {
    const resetParams = {
      page: 1,
      page_size: 10,
      status: '',
      contract_type: '',
      search: '',
    }
    setSearchForm({
      search: '',
      status: '',
      contract_type: '',
    })
    setQueryParams(resetParams)
  }

  // 计算付款计划总额
  const totalMilestoneAmount = milestones.reduce((sum, m) => sum + m.amount, 0)

  return (
    <div className='contract-list'>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className='stats-row'>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='合同总数'
              value={stats?.total_contracts || 0}
              valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='生效中合同'
              value={stats?.active_contracts || 0}
              valueStyle={{ color: '#3f8600', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='合同总金额'
              value={stats?.total_contract_amount || 0}
              prefix='¥'
              valueStyle={{ color: '#1890ff', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='本月新增'
              value={stats?.new_contracts_this_month || 0}
              valueStyle={{ color: '#722ed1', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索过滤器 */}
      <Card className='search-card'>
        <Space size='middle' wrap>
          <Input
            placeholder='搜索合同编号'
            value={searchForm.search}
            onChange={(e) => setSearchForm({ ...searchForm, search: e.target.value })}
            allowClear
            style={{ width: 200 }}
            size='large'
          />

          <Select
            placeholder='筛选状态'
            style={{ width: 120 }}
            value={searchForm.status || undefined}
            onChange={(value) => setSearchForm({ ...searchForm, status: value || '' })}
            allowClear
            size='large'
          >
            <Option value='draft'>草稿</Option>
            <Option value='active'>生效中</Option>
            <Option value='suspended'>已暂停</Option>
            <Option value='completed'>已完成</Option>
            <Option value='cancelled'>已取消</Option>
          </Select>

          <Select
            placeholder='合同类型'
            style={{ width: 120 }}
            value={searchForm.contract_type || undefined}
            onChange={(value) => setSearchForm({ ...searchForm, contract_type: value || '' })}
            allowClear
            size='large'
          >
            <Option value='original'>原始合同</Option>
            <Option value='supplementary'>补充合同</Option>
          </Select>

          <Button type='primary' icon={<SearchOutlined />} onClick={handleSearch} size='large'>
            搜索
          </Button>

          <Button icon={<ReloadOutlined />} onClick={handleReset} size='large'>
            重置筛选
          </Button>
        </Space>
      </Card>

      {/* 合同列表 */}
      <Card
        title='合同列表'
        extra={
          <Button type='primary' icon={<PlusOutlined />} onClick={handleAdd}>
            新增合同
          </Button>
        }
      >
        <Table
          columns={columns.map((col) => ({
            ...col,
            onCell: (record) => ({
              onClick: () => handleView(record),
              style: { cursor: 'pointer' },
            }),
          }))}
          dataSource={contracts}
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
        width={800}
        destroyOnClose
      >
        <Form form={form} layout='vertical'>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label='合同编号'
                name='contract_code'
                rules={[{ required: true, message: '请输入合同编号' }]}
              >
                <Input placeholder='请输入合同编号' />
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
              <Form.Item label='案件ID' name='case_id'>
                <InputNumber placeholder='可选，关联案件' style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label='合同类型'
                name='contract_type'
                rules={[{ required: true, message: '请选择合同类型' }]}
              >
                <Radio.Group>
                  <Radio value='original'>原始合同</Radio>
                  <Radio value='supplementary'>补充合同</Radio>
                </Radio.Group>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label='合同金额'
                name='contract_amount'
                rules={[{ required: true, message: '请输入合同金额' }]}
              >
                <InputNumber
                  placeholder='请输入合同金额'
                  style={{ width: '100%' }}
                  min={0}
                  precision={2}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label='币种'
                name='currency'
                rules={[{ required: true, message: '请选择币种' }]}
              >
                <Select>
                  <Option value='CNY'>人民币</Option>
                  <Option value='USD'>美元</Option>
                  <Option value='EUR'>欧元</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label='结算周期'
                name='billing_cycle'
                rules={[{ required: true, message: '请选择结算周期' }]}
              >
                <Select>
                  <Option value='一次性'>一次性</Option>
                  <Option value='按月'>按月</Option>
                  <Option value='按季'>按季</Option>
                  <Option value='按年'>按年</Option>
                  <Option value='按阶段'>按阶段</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item label='付款条款' name='payment_terms'>
            <Input placeholder='请输入付款条款' />
          </Form.Item>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item label='签订日期' name='signed_at'>
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label='开始日期' name='start_date'>
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label='结束日期' name='end_date'>
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item label='关联文档ID' name='document_id'>
            <InputNumber placeholder='可选，关联文档' style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 合同详情抽屉 */}
      <Drawer
        title='合同详情'
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
        width={800}
      >
        {selectedContract && (
          <div className='contract-detail'>
            <Card title='基本信息' size='small' style={{ marginBottom: 16 }}>
              <Row gutter={[16, 8]}>
                <Col span={12}>
                  <span className='detail-label'>合同编号：</span>
                  {selectedContract.contract_code}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>合同类型：</span>
                  <Tag color={selectedContract.contract_type === 'supplementary' ? 'blue' : 'green'}>
                    {selectedContract.contract_type === 'supplementary' ? '补充合同' : '原始合同'}
                  </Tag>
                </Col>
                <Col span={12}>
                  <span className='detail-label'>客户名称：</span>
                  {selectedContract.client?.name || '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>案件名称：</span>
                  {selectedContract.case?.title || '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>合同金额：</span>
                  <span style={{ fontSize: 16, fontWeight: 'bold', color: '#f5222d' }}>
                    {selectedContract.currency === 'USD' ? '$' : '¥'}
                    {selectedContract.contract_amount.toLocaleString()}
                  </span>
                </Col>
                <Col span={12}>
                  <span className='detail-label'>合同状态：</span>
                  {(() => {
                    const config = contractStatusMap[selectedContract.status] || {
                      text: selectedContract.status,
                      color: 'default',
                    }
                    return <Tag color={config.color}>{config.text}</Tag>
                  })()}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>结算周期：</span>
                  {selectedContract.billing_cycle}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>付款条款：</span>
                  {selectedContract.payment_terms || '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>签订日期：</span>
                  {selectedContract.signed_at
                    ? dayjs(selectedContract.signed_at).format('YYYY-MM-DD')
                    : '-'}
                </Col>
                <Col span={12}>
                  <span className='detail-label'>合同期限：</span>
                  {selectedContract.start_date || selectedContract.end_date
                    ? `${selectedContract.start_date ? dayjs(selectedContract.start_date).format('YYYY-MM-DD') : '?'} ~ ${selectedContract.end_date ? dayjs(selectedContract.end_date).format('YYYY-MM-DD') : '?'}`
                    : '-'}
                </Col>
              </Row>
            </Card>

            <Card
              title='付款计划'
              size='small'
              extra={
                <Space>
                  <span>
                    计划总额: <strong>¥{totalMilestoneAmount.toLocaleString()}</strong>
                  </span>
                  <Button
                    type='primary'
                    size='small'
                    icon={<PlusOutlined />}
                    onClick={handleAddMilestone}
                    disabled={selectedContract.status !== 'draft'}
                  >
                    添加付款计划
                  </Button>
                </Space>
              }
            >
              <Table
                columns={[
                  ...milestoneColumns.slice(0, -1),
                  {
                    title: '操作',
                    key: 'action',
                    width: 120,
                    render: (_, record) => (
                      <Space size='small'>
                        <Button
                          type='link'
                          size='small'
                          onClick={() => handleEditMilestone(record)}
                          disabled={selectedContract.status !== 'draft'}
                        >
                          编辑
                        </Button>
                        <Popconfirm
                          title='确定要删除这个付款计划吗？'
                          onConfirm={() => handleDeleteMilestone(record.id)}
                          disabled={selectedContract.status !== 'draft'}
                        >
                          <Button type='link' size='small' danger disabled={selectedContract.status !== 'draft'}>
                            删除
                          </Button>
                        </Popconfirm>
                      </Space>
                    ),
                  },
                ]}
                dataSource={milestones}
                rowKey='id'
                pagination={false}
                size='small'
              />
            </Card>
          </div>
        )}
      </Drawer>

      {/* 新增/编辑付款计划弹窗 */}
      <Modal
        title={editingMilestone ? '编辑付款计划' : '新增付款计划'}
        open={milestoneModalVisible}
        onOk={handleMilestoneSubmit}
        onCancel={() => setMilestoneModalVisible(false)}
        destroyOnClose
      >
        <Form form={milestoneForm} layout='vertical'>
          <Form.Item name='contract_id' hidden>
            <Input />
          </Form.Item>
          <Form.Item name='id' hidden>
            <Input />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label='序号'
                name='sequence'
                rules={[{ required: true, message: '请输入序号' }]}
              >
                <InputNumber style={{ width: '100%' }} min={1} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label='付款阶段'
                name='name'
                rules={[{ required: true, message: '请输入付款阶段名称' }]}
              >
                <Input placeholder='如：首期款、进度款、尾款' />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label='金额'
                name='amount'
                rules={[{ required: true, message: '请输入金额' }]}
              >
                <InputNumber style={{ width: '100%' }} min={0} precision={2} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label='比例(%)'
                name='percentage'
                rules={[{ required: true, message: '请输入比例' }]}
              >
                <InputNumber style={{ width: '100%' }} min={0} max={100} precision={2} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label='计划付款日期' name='due_date'>
                <DatePicker style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item label='付款条件' name='condition'>
            <TextArea
              placeholder='如：合同签订后X日内、案件立案后X日内等'
              rows={3}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ContractList
