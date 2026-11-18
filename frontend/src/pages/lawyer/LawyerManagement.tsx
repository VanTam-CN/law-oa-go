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
  message,
  Popconfirm,
  Tooltip,
  Row,
  Col,
  Statistic,
} from 'antd'

const { TextArea } = Input
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SearchOutlined,
  UserOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import { lawyerService } from '@/services/lawyer'
import './LawyerManagement.less'

// 🔧 修复：统一律师接口定义，与后端LawyerResponse保持一致
interface Lawyer {
  id?: number
  name: string
  phone: string
  email: string
  avatar?: string
  status?: string
  licenseNumber?: string // 执业证号
  specialty?: string // 专业领域
  department?: string // 部门
  position?: string // 职位
  gender?: string // 性别
  experience?: number // 从业年限
  joinDate?: string // 入职日期
  profile?: string // 个人简介
  created_at?: string // 创建时间
}

const LawyerManagement: React.FC = () => {
  // 统一的状态管理
  const [loading, setLoading] = useState<boolean>(false)
  const [data, setData] = useState<Lawyer[]>([])
  const [visible, setVisible] = useState<boolean>(false)
  const [editingItem, setEditingItem] = useState<Lawyer | null>(null)
  const [viewMode, setViewMode] = useState<boolean>(false) // 查看模式
  const [form] = Form.useForm()

  // 统一的搜索状态
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')

  // 统一的分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })

  // 表格列定义 - 移到组件内部以访问状态
  const columns: any[] = [
    {
      title: '姓名',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Lawyer) => (
        <Space>
          {record.avatar ? (
            <img
              src={record.avatar}
              alt={text}
              style={{ width: 24, height: 24, borderRadius: '50%' }}
            />
          ) : (
            <UserOutlined />
          )}
          {text}
        </Space>
      ),
    },
    {
      title: '手机号',
      dataIndex: 'phone',
      key: 'phone',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '执业证号',
      dataIndex: 'licenseNumber',
      key: 'licenseNumber',
      render: (text: string) => text || '-',
    },
    {
      title: '专业领域',
      dataIndex: 'specialty',
      key: 'specialty',
    },
    {
      title: '部门',
      dataIndex: 'department',
      key: 'department',
    },
    {
      title: '职位',
      dataIndex: 'position',
      key: 'position',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusConfig = {
          active: { color: 'green', text: '活跃' },
          inactive: { color: 'red', text: '非活跃' },
          on_leave: { color: 'orange', text: '休假' },
        }
        const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.inactive
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '从业年限',
      dataIndex: 'experience',
      key: 'experience',
      render: (experience: number) => (experience ? `${experience}年` : '-'),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: any, record: Lawyer) => (
        <Space size='middle'>
          <Tooltip title='查看详情'>
            <Button type='link' icon={<EyeOutlined />} onClick={() => handleViewDetails(record)} />
          </Tooltip>
          <Tooltip title='编辑'>
            <Button
              type='link'
              icon={<EditOutlined />}
              onClick={() => {
                setEditingItem(record)
                form.setFieldsValue(record)
                setVisible(true)
              }}
            />
          </Tooltip>
          <Tooltip title='删除'>
            <Popconfirm
              title='确定要删除这条记录吗？'
              onConfirm={() => {
                if (record.id == null) {
                  message.error('律师ID缺失，无法删除')
                  return
                }
                handleDelete(record.id)
              }}
              okText='确定'
              cancelText='取消'
            >
              <Button type='link' danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ]

  // 🔧 修复：优化律师数据获取逻辑
  const fetchData = async () => {
    setLoading(true)
    try {
      // 构建查询参数 - 使用正确的参数名
      const params: any = {
        page: pagination.current,
        page_size: pagination.pageSize,
      }

      if (searchText) {
        params.search = searchText // 修复：使用search而不是keyword
      }

      if (statusFilter) {
        params.status = statusFilter
      }

      console.log('获取律师数据...', params)

      // 调用API
      const response = await lawyerService.getLawyerList(params)
      console.log('律师数据响应:', response)

      // 🔧 修复：处理多种可能的API响应格式
      let lawyerData = []
      let totalCount = 0

      if (response && response.list) {
        // 标准格式：{list: [...], total: number}
        lawyerData = response.list
        totalCount = response.total || response.list.length
      } else if (response && Array.isArray(response)) {
        // 直接返回数组格式
        lawyerData = response
        totalCount = response.length
      } else if (response && response.data) {
        // 新API格式：{data: [...], pagination: {...}}
        lawyerData = response.data
        totalCount = response.pagination?.total || response.data.length
      } else {
        console.warn('未知的数据格式:', response)
        lawyerData = []
        totalCount = 0
      }

      console.log('处理后的律师数据:', lawyerData)
      console.log('总数:', totalCount)

      setData(lawyerData)
      setPagination({
        ...pagination,
        total: totalCount,
      })

      if (lawyerData.length === 0) {
        message.info('暂无律师数据')
      }
    } catch (error: any) {
      console.error('获取律师数据失败:', error)
      message.error(`获取律师数据失败: ${error.message || '未知错误'}`)
      setData([])
    } finally {
      setLoading(false)
    }
  }

  // 删除律师
  const handleDelete = async (id: number) => {
    try {
      await lawyerService.deleteLawyer(id)
      message.success('删除成功')
      fetchData()
    } catch (error: any) {
      console.error('删除律师失败:', error)
      message.error(`删除失败: ${error.message || '未知错误'}`)
    }
  }

  // 新增律师
  const handleAdd = () => {
    setEditingItem(null)
    setViewMode(false) // 设置为编辑模式
    form.resetFields()
    setVisible(true)
  }

  // 查看律师详情
  const handleViewDetails = (record: Lawyer) => {
    setEditingItem(record)
    setViewMode(true) // 设置为查看模式
    form.setFieldsValue(record)
    setVisible(true)
  }

  // 编辑律师
  const handleEdit = (record: Lawyer) => {
    setEditingItem(record)
    setViewMode(false) // 设置为编辑模式
    form.setFieldsValue(record)
    setVisible(true)
  }

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingItem) {
        if (editingItem.id == null) {
          message.error('律师ID缺失，无法更新')
          return
        }
        await lawyerService.updateLawyer(editingItem.id, values)
        message.success('更新成功')
      } else {
        await lawyerService.createLawyer(values)
        message.success('新增成功')
      }
      setVisible(false)
      fetchData()
    } catch (error: any) {
      console.error('提交失败:', error)
      message.error(error.message || '操作失败')
    }
  }

  useEffect(() => {
    fetchData()
  }, [searchText, statusFilter, pagination.current, pagination.pageSize])

  return (
    <div className='lawyer-management'>
      {/* 统一的统计卡片 */}
      <Row gutter={[16, 16]} className='stats-row'>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='律师总数'
              value={data.length}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='活跃数量'
              value={data.filter((item) => item.status === 'active').length}
              valueStyle={{ color: '#3f8600', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='本月新增'
              value={0}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='待处理'
              value={0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#1a1a1a', fontSize: '24px', fontWeight: 700 }}
            />
          </Card>
        </Col>
      </Row>

      {/* 统一的搜索表单 */}
      <Card className='search-card'>
        <Form layout='inline'>
          <Form.Item label='搜索'>
            <Input
              placeholder='搜索律师信息'
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              allowClear
              style={{ width: 250 }}
            />
          </Form.Item>
          <Form.Item label='状态'>
            <Select
              style={{ width: 120 }}
              value={statusFilter}
              onChange={setStatusFilter}
              allowClear
              placeholder='全部'
            >
              <Select.Option value='active'>活跃</Select.Option>
              <Select.Option value='inactive'>非活跃</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type='primary' icon={<SearchOutlined />} onClick={fetchData}>
                搜索
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => {
                  setSearchText('')
                  setStatusFilter('')
                  fetchData()
                }}
              >
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {/* 统一的数据表格 */}
      <Card
        title='律师列表'
        extra={
          <Button type='primary' icon={<PlusOutlined />} onClick={handleAdd}>
            新建律师
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={data}
          rowKey='id'
          loading={loading}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            onChange: (page, size) => {
              setPagination({
                ...pagination,
                current: page,
                pageSize: size || 10,
              })
            },
          }}
        />
      </Card>

      {/* 新增/编辑弹窗 */}
      <Modal
        title={editingItem ? (viewMode ? '律师详情' : '编辑律师') : '新增律师'}
        open={visible}
        onOk={viewMode ? undefined : handleSubmit}
        onCancel={() => {
          setVisible(false)
          setViewMode(false)
        }}
        width={600}
        destroyOnHidden
        footer={
          viewMode
            ? [
                <Button key='close' onClick={() => setVisible(false)}>
                  关闭
                </Button>,
              ]
            : undefined
        }
      >
        <Form form={form} layout='vertical' disabled={viewMode}>
          <Form.Item label='姓名' name='name' rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder='请输入姓名' />
          </Form.Item>

          <Form.Item
            label='手机号'
            name='phone'
            rules={[{ required: true, message: '请输入手机号' }]}
          >
            <Input placeholder='请输入手机号' />
          </Form.Item>

          <Form.Item
            label='邮箱'
            name='email'
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input placeholder='请输入邮箱' />
          </Form.Item>

          <Form.Item label='执业证号' name='licenseNumber'>
            <Input placeholder='请输入执业证号' />
          </Form.Item>

          <Form.Item
            label='专业领域'
            name='specialty'
            rules={[{ required: true, message: '请输入专业领域' }]}
          >
            <Input placeholder='请输入专业领域' />
          </Form.Item>

          <Form.Item
            label='部门'
            name='department'
            rules={[{ required: true, message: '请输入部门' }]}
          >
            <Input placeholder='请输入部门' />
          </Form.Item>

          <Form.Item
            label='职位'
            name='position'
            rules={[{ required: true, message: '请输入职位' }]}
          >
            <Input placeholder='请输入职位' />
          </Form.Item>

          <Form.Item label='性别' name='gender'>
            <Select placeholder='请选择性别'>
              <Select.Option value='male'>男</Select.Option>
              <Select.Option value='female'>女</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item label='从业年限' name='experience'>
            <Input type='number' placeholder='请输入从业年限' />
          </Form.Item>

          <Form.Item label='状态' name='status' rules={[{ required: true, message: '请选择状态' }]}>
            <Select placeholder='请选择状态'>
              <Select.Option value='active'>活跃</Select.Option>
              <Select.Option value='inactive'>非活跃</Select.Option>
              <Select.Option value='on_leave'>休假</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item label='个人简介' name='profile'>
            <TextArea rows={3} placeholder='请输入个人简介' />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default LawyerManagement
