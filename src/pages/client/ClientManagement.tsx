import React, { useState, useEffect } from 'react'
import {
  Table,
  Card,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  Radio,
  message,
  Popconfirm,
  Statistic,
  Row,
  Col,
  Tooltip,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  UserOutlined,
  BankOutlined,
  SearchOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { clientService } from '@/services/client'
import './ClientManagement.less'

interface Client {
  id?: number
  name: string
  type: string
  phone: string
  email: string
  address: string
  idCard?: string
  company?: string
  industry?: string
  contactPerson?: string
  contactPhone?: string
  source?: string
  notes?: string
  status: string
  createdAt?: string
  updatedAt?: string
}

interface ClientStats {
  total: number
  active: number
  inactive: number
  byType: Record<string, number>
}

const { Option } = Select
const { TextArea } = Input

const ClientManagement: React.FC = () => {
  const [clients, setClients] = useState<Client[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [modalVisible, setModalVisible] = useState<boolean>(false)
  const [modalTitle, setModalTitle] = useState<string>('')
  const [editingClient, setEditingClient] = useState<Client | null>(null)
  const [stats, setStats] = useState<ClientStats | null>(null)
  const [form] = Form.useForm()

  // 监听客户类型字段变化
  const clientType = Form.useWatch('type', form)

  // 查询参数
  const [queryParams, setQueryParams] = useState({
    name: '',
    type: '',
    status: '',
    pageNum: 1,
    pageSize: 10,
  })

  // 搜索表单状态
  const [searchForm, setSearchForm] = useState({
    name: '',
    type: '',
    status: '',
  })

  const [total, setTotal] = useState<number>(0)

  // 获取客户列表
  const fetchClients = async () => {
    setLoading(true)
    try {
      const res = await clientService.getClientList(queryParams)
      console.log('客户列表API响应:', res) // 调试日志

      // 🔧 修复：处理后端实际API响应格式
      // 后端返回格式：{success: true, data: {clients: [...], pagination: {...}}}
      let clientData = []
      let totalCount = 0

      if (res.success && res.data) {
        // 新API格式：数据在res.data.clients中
        clientData = res.data.clients || []
        totalCount = res.data.pagination?.total || 0
        console.log('使用新API格式解析，clients数量:', clientData.length)
      } else if (res.data && res.data.clients) {
        // 兼容格式：直接从data.clients获取
        clientData = res.data.clients
        totalCount = res.data.pagination?.total || 0
        console.log('使用兼容格式解析，clients数量:', clientData.length)
      } else if (res.data && Array.isArray(res.data)) {
        // 直接是数组数据
        clientData = res.data
        totalCount = res.data.length
        console.log('使用数组格式解析，clients数量:', clientData.length)
      } else if (res.clients && Array.isArray(res.clients)) {
        // 旧格式：clients字段直接包含数据
        clientData = res.clients
        totalCount = res.pagination?.total || res.total || 0
        console.log('使用旧格式解析，clients数量:', clientData.length)
      } else {
        console.warn('未识别的响应格式:', res)
        clientData = []
        totalCount = 0
      }

      // 🔧 修复：字段映射 - 将后端的下划线字段名映射为前端的驼峰命名
      const mappedClientData = clientData.map((client: any) => ({
        ...client,
        idCard: client.id_card, // 后端 id_card -> 前端 idCard
        contactPerson: client.contact_person, // 后端 contact_person -> 前端 contactPerson
        contactPhone: client.contact_phone, // 后端 contact_phone -> 前端 contactPhone
        createdAt: client.created_at, // 后端 created_at -> 前端 createdAt
        updatedAt: client.updated_at, // 后端 updated_at -> 前端 updatedAt
      }))

      console.log('映射后的客户数据:', mappedClientData)
      console.log('总数:', totalCount)

      setClients(mappedClientData)
      setTotal(totalCount)

      // 获取统计数据
      fetchStats()
    } catch (error) {
      console.error('获取客户列表失败:', error)
      message.error('获取客户列表失败')
    } finally {
      setLoading(false)
    }
  }

  // 获取客户统计
  const fetchStats = async () => {
    try {
      const res = await clientService.getClientStats()
      console.log('客户统计API响应:', res) // 调试日志

      // 🔧 修复：处理后端统计数据的实际格式
      let statsData = null

      console.log('统计数据原始响应:', res) // 调试日志

      if (res.success && res.data) {
        // 新API格式：{success: true, data: {total_clients, active_clients, ...}}
        statsData = {
          total: res.data.total_clients || 0,
          statusStats: {
            active: res.data.active_clients || 0,
            inactive: res.data.inactive_clients || 0,
          },
          typeStats: {
            个人: res.data.personal_clients || 0,
            企业: res.data.enterprise_clients || 0,
          },
          monthlyNew: res.data.new_clients_this_month || 0,
        }
        console.log('使用新API格式解析统计数据:', statsData)
      } else if (res.data && res.data.total_clients !== undefined) {
        // 兼容格式：直接在data中包含统计字段
        const data = res.data
        statsData = {
          total: data.total_clients || 0,
          statusStats: {
            active: data.active_clients || 0,
            inactive: data.inactive_clients || 0,
          },
          typeStats: {
            个人: data.personal_clients || 0,
            企业: data.enterprise_clients || 0,
          },
          monthlyNew: data.new_clients_this_month || 0,
        }
        console.log('使用兼容格式解析统计数据:', statsData)
      } else if (res.total_clients !== undefined) {
        // 后端直接返回格式：{total_clients, active_clients, ...}
        // 基于现有客户数据计算类型统计（如果后端没有提供）
        const personalCount = clients.filter((client) => client.type === '个人').length
        const enterpriseCount = clients.filter((client) => client.type === '企业').length

        statsData = {
          total: res.total_clients,
          statusStats: {
            active: res.active_clients || 0,
            inactive: res.inactive_clients || 0,
          },
          typeStats: {
            个人: res.personal_clients || personalCount,
            企业: res.enterprise_clients || enterpriseCount,
          },
          monthlyNew: res.new_clients_this_month || 0,
        }
        console.log('使用后端直接格式解析统计数据:', statsData)
      } else if (res.total !== undefined) {
        // 直接是统计数据对象
        statsData = res
        console.log('使用直接统计数据对象:', statsData)
      } else {
        console.warn('未识别的统计数据格式:', res)
        // 提供默认值防止页面报错
        statsData = {
          total: 0,
          statusStats: { active: 0, inactive: 0 },
          typeStats: { 个人: 0, 企业: 0 },
          monthlyNew: 0,
        }
      }

      console.log('处理后的统计数据:', statsData)
      setStats(statsData)
    } catch (error) {
      console.error('获取统计数据失败:', error)
      message.error('获取统计数据失败')
    }
  }

  useEffect(() => {
    fetchClients()
  }, [queryParams])

  // 打开新增客户弹窗
  const handleAdd = () => {
    setModalTitle('新增客户')
    setEditingClient(null)
    form.resetFields()
    // 设置默认客户类型为个人
    form.setFieldsValue({ type: '个人' })
    setModalVisible(true)
  }

  // 打开编辑客户弹窗
  const handleEdit = (record: Client) => {
    setModalTitle('编辑客户')
    setEditingClient(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  // 🔧 修复：智能判断客户类型的辅助函数
  const getClientType = (record: Client): '个人' | '企业' => {
    if (record.type === '企业') {
      return '企业'
    } else if (!record.type || record.type === '') {
      // 如果类型为空，根据公司名称字段判断
      return record.company && record.company.trim() !== '' ? '企业' : '个人'
    }
    return '个人'
  }

  // 查看客户详情
  const handleView = (record: Client) => {
    const clientType = getClientType(record)

    Modal.info({
      title: '客户详情',
      width: 600,
      content: (
        <div className='client-detail'>
          <p>
            <strong>客户名称：</strong>
            {record.name || '未命名客户'}
          </p>
          <p>
            <strong>客户类型：</strong>
            <Tag color={clientType === '个人' ? 'blue' : 'purple'}>{clientType}</Tag>
          </p>
          <p>
            <strong>联系电话：</strong>
            {record.phone || '-'}
          </p>
          <p>
            <strong>电子邮箱：</strong>
            {record.email || '-'}
          </p>
          {clientType === '个人' && record.idCard && (
            <p>
              <strong>身份证号：</strong>
              {record.idCard}
            </p>
          )}
          <p>
            <strong>地址：</strong>
            {record.address || '-'}
          </p>
          {clientType === '企业' && (
            <>
              <p>
                <strong>公司名称：</strong>
                {record.company || '-'}
              </p>
              <p>
                <strong>所属行业：</strong>
                {record.industry || '-'}
              </p>
              <p>
                <strong>联系人：</strong>
                {record.contactPerson || '-'}
              </p>
              <p>
                <strong>联系电话：</strong>
                {record.contactPhone || '-'}
              </p>
            </>
          )}
          <p>
            <strong>客户来源：</strong>
            {record.source || '-'}
          </p>
          <p>
            <strong>客户状态：</strong>
            <Tag color={record.status === 'active' ? 'green' : 'red'}>
              {record.status === 'active' ? '活跃' : '非活跃'}
            </Tag>
          </p>
          {record.createdAt && (
            <p>
              <strong>创建时间：</strong>
              {new Date(record.createdAt).toLocaleString()}
            </p>
          )}
          <p>
            <strong>备注：</strong>
            {record.notes || '-'}
          </p>
        </div>
      ),
    })
  }

  // 删除客户
  const handleDelete = async (id: number) => {
    try {
      await clientService.deleteClient(id)
      message.success('删除成功')
      fetchClients()
    } catch (error) {
      message.error('删除失败')
    }
  }

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingClient) {
        await clientService.updateClient(editingClient.id!, values)
        message.success('更新成功')
      } else {
        await clientService.createClient(values)
        message.success('新增成功')
      }
      setModalVisible(false)
      fetchClients()
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  // 搜索
  const handleSearch = () => {
    setQueryParams({
      ...queryParams,
      name: searchForm.name,
      type: searchForm.type,
      status: searchForm.status,
      pageNum: 1,
    })
  }

  // 重置搜索
  const handleReset = () => {
    const resetParams = {
      name: '',
      type: '',
      status: '',
      pageNum: 1,
      pageSize: 10,
    }
    setSearchForm({
      name: '',
      type: '',
      status: '',
    })
    setQueryParams(resetParams)
  }

  // 表格列定义
  const columns: ColumnsType<Client> = [
    {
      title: '客户名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Client) => {
        // 🔧 修复：更智能的图标选择逻辑
        let isEnterprise = false

        if (record.type === '企业') {
          isEnterprise = true
        } else if (!record.type || record.type === '') {
          // 如果类型为空，根据公司名称字段判断
          isEnterprise = record.company && record.company.trim() !== ''
        }

        return (
          <Space>
            {isEnterprise ? <BankOutlined /> : <UserOutlined />}
            <span style={{ fontWeight: isEnterprise ? 'bold' : 'normal' }}>
              {text || '未命名客户'}
            </span>
          </Space>
        )
      },
    },
    {
      title: '客户类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string, record: Client) => {
        // 🔧 修复：更智能的类型判断和显示
        let displayType = type
        let color = 'blue' // 默认个人客户颜色

        // 如果数据库中类型为空，根据名称和其他字段智能判断
        if (!type || type === '') {
          if (record.company && record.company.trim() !== '') {
            displayType = '企业'
            color = 'purple'
          } else {
            displayType = '个人'
            color = 'blue'
          }
        } else if (type === '个人') {
          color = 'blue'
        } else if (type === '企业') {
          color = 'purple'
        } else {
          // 处理其他可能的类型值
          displayType = type
          color = 'green'
        }

        return <Tag color={color}>{displayType}</Tag>
      },
    },
    {
      title: '联系电话',
      dataIndex: 'phone',
      key: 'phone',
    },
    {
      title: '电子邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '客户状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'red'}>
          {status === 'active' ? '活跃' : '非活跃'}
        </Tag>
      ),
    },
    {
      title: '客户来源',
      dataIndex: 'source',
      key: 'source',
      render: (source: string) => source || '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space size='middle'>
          <Tooltip title='查看详情'>
            <Button type='link' icon={<EyeOutlined />} onClick={() => handleView(record)} />
          </Tooltip>
          <Tooltip title='编辑'>
            <Button type='link' icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          </Tooltip>
          <Tooltip title='删除'>
            <Popconfirm
              title='确定要删除这个客户吗？'
              onConfirm={() => handleDelete(record.id!)}
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

  return (
    <div className='client-management'>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className='stats-row'>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic title='客户总数' value={stats?.total || 0} prefix={<UserOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='活跃客户'
              value={stats?.statusStats?.active || 0}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='个人客户'
              value={stats?.typeStats?.['个人'] || 0}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='企业客户'
              value={stats?.typeStats?.['企业'] || 0}
              prefix={<BankOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索表单 */}
      <Card className='search-card'>
        <Form layout='inline'>
          <Form.Item label='客户名称'>
            <Input
              placeholder='请输入客户名称'
              value={searchForm.name}
              onChange={(e) => setSearchForm({ ...searchForm, name: e.target.value })}
              allowClear
            />
          </Form.Item>
          <Form.Item label='客户类型'>
            <Select
              style={{ width: 120 }}
              value={searchForm.type}
              onChange={(value) => setSearchForm({ ...searchForm, type: value })}
              allowClear
              placeholder='全部'
            >
              <Option value='个人'>个人</Option>
              <Option value='企业'>企业</Option>
            </Select>
          </Form.Item>
          <Form.Item label='客户状态'>
            <Select
              style={{ width: 120 }}
              value={searchForm.status}
              onChange={(value) => setSearchForm({ ...searchForm, status: value })}
              allowClear
              placeholder='全部'
            >
              <Option value='active'>活跃</Option>
              <Option value='inactive'>非活跃</Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type='primary' icon={<SearchOutlined />} onClick={handleSearch}>
                搜索
              </Button>
              <Button icon={<ReloadOutlined />} onClick={handleReset}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {/* 客户列表 */}
      <Card
        title='客户列表'
        extra={
          <Button type='primary' icon={<PlusOutlined />} onClick={handleAdd}>
            新增客户
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={clients}
          rowKey='id'
          loading={loading}
          pagination={{
            current: queryParams.pageNum,
            pageSize: queryParams.pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            onChange: (page, size) => {
              setQueryParams({
                ...queryParams,
                pageNum: page,
                pageSize: size || 10,
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
        width={600}
        destroyOnHidden
      >
        <Form form={form} layout='vertical'>
          <Form.Item
            label='客户类型'
            name='type'
            rules={[{ required: true, message: '请选择客户类型' }]}
          >
            <Radio.Group>
              <Radio value='个人'>个人</Radio>
              <Radio value='企业'>企业</Radio>
            </Radio.Group>
          </Form.Item>

          <Form.Item
            label={clientType === '企业' ? '公司名称' : '客户姓名'}
            name='name'
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input placeholder='请输入名称' />
          </Form.Item>

          <Form.Item
            label='联系电话'
            name='phone'
            rules={[{ required: true, message: '请输入联系电话' }]}
          >
            <Input placeholder='请输入联系电话' />
          </Form.Item>

          <Form.Item
            label='电子邮箱'
            name='email'
            rules={[
              { required: true, message: '请输入电子邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input placeholder='请输入电子邮箱' />
          </Form.Item>

          {clientType === '个人' && (
            <Form.Item
              label='身份证号'
              name='idCard'
              rules={[
                {
                  pattern:
                    /^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$/,
                  message: '请输入有效的身份证号',
                },
              ]}
            >
              <Input placeholder='请输入身份证号' />
            </Form.Item>
          )}

          {clientType === '企业' && (
            <>
              <Form.Item label='所属行业' name='industry'>
                <Input placeholder='请输入所属行业' />
              </Form.Item>
              <Form.Item label='联系人' name='contactPerson'>
                <Input placeholder='请输入联系人' />
              </Form.Item>
              <Form.Item label='联系电话' name='contactPhone'>
                <Input placeholder='请输入联系电话' />
              </Form.Item>
            </>
          )}

          <Form.Item
            label='地址'
            name='address'
            rules={[{ required: true, message: '请输入地址' }]}
          >
            <TextArea placeholder='请输入详细地址' rows={2} />
          </Form.Item>

          <Form.Item label='客户来源' name='source'>
            <Select placeholder='请选择客户来源'>
              <Option value='推荐'>推荐</Option>
              <Option value='自主开发'>自主开发</Option>
              <Option value='网络推广'>网络推广</Option>
              <Option value='合作机构'>合作机构</Option>
              <Option value='其他'>其他</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label='客户状态'
            name='status'
            rules={[{ required: true, message: '请选择客户状态' }]}
          >
            <Radio.Group>
              <Radio value='active'>活跃</Radio>
              <Radio value='inactive'>非活跃</Radio>
            </Radio.Group>
          </Form.Item>

          <Form.Item label='备注' name='notes'>
            <TextArea placeholder='请输入备注信息' rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ClientManagement
