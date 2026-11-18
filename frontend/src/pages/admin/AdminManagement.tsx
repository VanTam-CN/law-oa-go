import React, { useState, useEffect } from 'react'
import {
  Card,
  Tabs,
  Table,
  Button,
  Space,
  Tag,
  Input,
  Modal,
  Form,
  message,
  Select,
  DatePicker,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  UserOutlined,
  FileTextOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  getEmployees,
  createEmployee,
  updateEmployee,
  deleteEmployee,
  getDocuments,
  createDocument,
  updateDocument,
  deleteDocument,
  getDepartments,
  Department,
  Employee,
  Document,
} from '@/services/admin'

const { Search } = Input
const { Option } = Select

interface Employee {
  id: string
  name: string
  position: string
  department: string
  email: string
  phone: string
  status: 'active' | 'inactive'
  joinDate: string
}

interface Document {
  id: string
  title: string
  type: string
  category: string
  createDate: string
  creator: string
  status: 'draft' | 'published' | 'archived'
}

const AdminManagement: React.FC = () => {
  const [activeTab, setActiveTab] = useState('employees')
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingItem, setEditingItem] = useState<Employee | Document | null>(null)
  const [form] = Form.useForm()

  // 数据状态
  const [employees, setEmployees] = useState<Employee[]>([])
  const [documents, setDocuments] = useState<Document[]>([])
  const [departments, setDepartments] = useState<Department[]>([])

  // 获取数据
  useEffect(() => {
    fetchData()
  }, [])

  const fetchData = async () => {
    try {
      setLoading(true)
      const [employeesData, documentsData, departmentsData] = await Promise.all([
        getEmployees(),
        getDocuments(),
        getDepartments(),
      ])
      setEmployees(employeesData)
      setDocuments(documentsData)
      setDepartments(departmentsData)
    } catch (error) {
      message.error('获取数据失败')
    } finally {
      setLoading(false)
    }
  }

  const employeeColumns: ColumnsType<Employee> = [
    {
      title: '姓名',
      dataIndex: 'name',
      key: 'name',
      width: 120,
    },
    {
      title: '职位',
      dataIndex: 'position',
      key: 'position',
      width: 120,
    },
    {
      title: '部门',
      dataIndex: 'department',
      key: 'department',
      width: 120,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      width: 180,
    },
    {
      title: '电话',
      dataIndex: 'phone',
      key: 'phone',
      width: 130,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'red'}>
          {status === 'active' ? '在职' : '离职'}
        </Tag>
      ),
    },
    {
      title: '入职日期',
      dataIndex: 'join_date',
      key: 'join_date',
      width: 120,
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space size='small'>
          <Button
            type='link'
            size='small'
            icon={<EditOutlined />}
            onClick={() => handleEditEmployee(record)}
          >
            编辑
          </Button>
          <Button
            type='link'
            size='small'
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteEmployee(record.id)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  const documentColumns: ColumnsType<Document> = [
    {
      title: '文档标题',
      dataIndex: 'title',
      key: 'title',
      width: 200,
      ellipsis: true,
    },
    {
      title: '文档类型',
      dataIndex: 'type',
      key: 'type',
      width: 120,
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      width: 120,
    },
    {
      title: '创建者',
      dataIndex: 'creator',
      key: 'creator',
      width: 100,
    },
    {
      title: '创建日期',
      dataIndex: 'create_date',
      key: 'create_date',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const statusMap = {
          draft: { text: '草稿', color: 'orange' },
          published: { text: '已发布', color: 'green' },
          archived: { text: '已归档', color: 'default' },
        }
        return (
          <Tag color={statusMap[status as keyof typeof statusMap].color}>
            {statusMap[status as keyof typeof statusMap].text}
          </Tag>
        )
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space size='small'>
          <Button
            type='link'
            size='small'
            icon={<EditOutlined />}
            onClick={() => handleEditDocument(record)}
          >
            编辑
          </Button>
          <Button
            type='link'
            size='small'
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteDocument(record.id)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  // 员工操作函数
  const handleAddEmployee = () => {
    setEditingItem(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEditEmployee = (employee: Employee) => {
    setEditingItem(employee)
    form.setFieldsValue(employee)
    setModalVisible(true)
  }

  const handleDeleteEmployee = async (id: number) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个员工吗？',
      onOk: async () => {
        try {
          await deleteEmployee(id)
          message.success('删除成功')
          fetchData()
        } catch (error) {
          message.error('删除失败')
        }
      },
    })
  }

  // 文档操作函数
  const handleAddDocument = () => {
    setEditingItem(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEditDocument = (document: Document) => {
    setEditingItem(document)
    form.setFieldsValue(document)
    setModalVisible(true)
  }

  const handleDeleteDocument = async (id: number) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个文档吗？',
      onOk: async () => {
        try {
          await deleteDocument(id)
          message.success('删除成功')
          fetchData()
        } catch (error) {
          message.error('删除失败')
        }
      },
    })
  }

  // 提交表单
  const handleSubmit = async (values: any) => {
    try {
      setLoading(true)

      if (activeTab === 'employees') {
        if (editingItem) {
          await updateEmployee((editingItem as Employee).id, values)
          message.success('员工更新成功')
        } else {
          await createEmployee(values)
          message.success('员工创建成功')
        }
      } else if (activeTab === 'documents') {
        if (editingItem) {
          await updateDocument((editingItem as Document).id, values)
          message.success('文档更新成功')
        } else {
          await createDocument(values)
          message.success('文档创建成功')
        }
      }

      setModalVisible(false)
      form.resetFields()
      fetchData()
    } catch (error) {
      message.error('操作失败')
    } finally {
      setLoading(false)
    }
  }

  const tabItems = [
    {
      key: 'employees',
      label: (
        <span>
          <UserOutlined />
          员工管理
        </span>
      ),
      children: (
        <Card>
          <div
            style={{
              marginBottom: 16,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <Space>
              <Search placeholder='搜索员工姓名' allowClear style={{ width: 200 }} />
              <Select placeholder='部门' style={{ width: 120 }} allowClear>
                {departments.map((dept) => (
                  <Option key={dept.id} value={dept.name}>
                    {dept.name}
                  </Option>
                ))}
              </Select>
              <Select placeholder='状态' style={{ width: 100 }} allowClear>
                <Option value='active'>在职</Option>
                <Option value='inactive'>离职</Option>
              </Select>
            </Space>
            <Button type='primary' icon={<PlusOutlined />} onClick={handleAddEmployee}>
              添加员工
            </Button>
          </div>
          <Table
            columns={employeeColumns}
            dataSource={employees}
            rowKey='id'
            pagination={{
              total: employees.length,
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条记录`,
            }}
          />
        </Card>
      ),
    },
    {
      key: 'documents',
      label: (
        <span>
          <FileTextOutlined />
          文档管理
        </span>
      ),
      children: (
        <Card>
          <div
            style={{
              marginBottom: 16,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <Space>
              <Search placeholder='搜索文档标题' allowClear style={{ width: 200 }} />
              <Select placeholder='文档类型' style={{ width: 120 }} allowClear>
                <Option value='制度文件'>制度文件</Option>
                <Option value='流程文件'>流程文件</Option>
                <Option value='标准文件'>标准文件</Option>
                <Option value='模板文件'>模板文件</Option>
              </Select>
              <Select placeholder='状态' style={{ width: 100 }} allowClear>
                <Option value='draft'>草稿</Option>
                <Option value='published'>已发布</Option>
                <Option value='archived'>已归档</Option>
              </Select>
            </Space>
            <Button type='primary' icon={<PlusOutlined />} onClick={handleAddDocument}>
              添加文档
            </Button>
          </div>
          <Table
            columns={documentColumns}
            dataSource={documents}
            rowKey='id'
            pagination={{
              total: documents.length,
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条记录`,
            }}
          />
        </Card>
      ),
    },
    {
      key: 'settings',
      label: (
        <span>
          <SettingOutlined />
          系统设置
        </span>
      ),
      children: (
        <Card>
          <div style={{ padding: '40px 0', textAlign: 'center', color: '#999' }}>
            <SettingOutlined style={{ fontSize: 48, marginBottom: 16 }} />
            <p>系统设置功能开发中...</p>
            <p>将包含：权限管理、系统配置、数据备份等功能</p>
          </div>
        </Card>
      ),
    },
  ]

  return (
    <div>
      <div className='page-header'>
        <h1 className='page-title'>行政管理</h1>
        <p>管理律所的人员、文档和系统设置</p>
      </div>

      <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} size='large' />

      {/* 员工/文档编辑模态框 */}
      <Modal
        title={
          activeTab === 'employees'
            ? editingItem
              ? '编辑员工'
              : '添加员工'
            : editingItem
              ? '编辑文档'
              : '添加文档'
        }
        open={modalVisible}
        onOk={() => form.submit()}
        onCancel={() => setModalVisible(false)}
        width={600}
        destroyOnHidden
      >
        <Form form={form} layout='vertical' onFinish={handleSubmit}>
          {activeTab === 'employees' ? (
            <>
              <Form.Item
                name='name'
                label='姓名'
                rules={[{ required: true, message: '请输入姓名' }]}
              >
                <Input placeholder='请输入姓名' />
              </Form.Item>
              <Form.Item
                name='position'
                label='职位'
                rules={[{ required: true, message: '请输入职位' }]}
              >
                <Input placeholder='请输入职位' />
              </Form.Item>
              <Form.Item
                name='department'
                label='部门'
                rules={[{ required: true, message: '请选择部门' }]}
              >
                <Select placeholder='请选择部门'>
                  {departments.map((dept) => (
                    <Select.Option key={dept.id} value={dept.name}>
                      {dept.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
              <Form.Item
                name='email'
                label='邮箱'
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效的邮箱地址' },
                ]}
              >
                <Input placeholder='请输入邮箱' />
              </Form.Item>
              <Form.Item
                name='phone'
                label='电话'
                rules={[{ required: true, message: '请输入电话' }]}
              >
                <Input placeholder='请输入电话' />
              </Form.Item>
              <Form.Item
                name='status'
                label='状态'
                rules={[{ required: true, message: '请选择状态' }]}
                initialValue='active'
              >
                <Select placeholder='请选择状态'>
                  <Select.Option value='active'>在职</Select.Option>
                  <Select.Option value='inactive'>离职</Select.Option>
                </Select>
              </Form.Item>
              <Form.Item
                name='join_date'
                label='入职日期'
                rules={[{ required: true, message: '请选择入职日期' }]}
              >
                <DatePicker placeholder='请选择入职日期' style={{ width: '100%' }} />
              </Form.Item>
            </>
          ) : (
            <>
              <Form.Item
                name='title'
                label='文档标题'
                rules={[{ required: true, message: '请输入文档标题' }]}
              >
                <Input placeholder='请输入文档标题' />
              </Form.Item>
              <Form.Item
                name='type'
                label='文档类型'
                rules={[{ required: true, message: '请选择文档类型' }]}
              >
                <Select placeholder='请选择文档类型'>
                  <Select.Option value='制度文件'>制度文件</Select.Option>
                  <Select.Option value='流程文件'>流程文件</Select.Option>
                  <Select.Option value='标准文件'>标准文件</Select.Option>
                  <Select.Option value='模板文件'>模板文件</Select.Option>
                </Select>
              </Form.Item>
              <Form.Item
                name='category'
                label='分类'
                rules={[{ required: true, message: '请输入分类' }]}
              >
                <Input placeholder='请输入分类' />
              </Form.Item>
              <Form.Item
                name='content'
                label='内容'
                rules={[{ required: true, message: '请输入内容' }]}
              >
                <Input.TextArea placeholder='请输入内容' rows={4} />
              </Form.Item>
              <Form.Item
                name='creator'
                label='创建者'
                rules={[{ required: true, message: '请输入创建者' }]}
              >
                <Input placeholder='请输入创建者' />
              </Form.Item>
              <Form.Item
                name='status'
                label='状态'
                rules={[{ required: true, message: '请选择状态' }]}
                initialValue='draft'
              >
                <Select placeholder='请选择状态'>
                  <Select.Option value='draft'>草稿</Select.Option>
                  <Select.Option value='published'>已发布</Select.Option>
                  <Select.Option value='archived'>已归档</Select.Option>
                </Select>
              </Form.Item>
            </>
          )}
        </Form>
      </Modal>
    </div>
  )
}

export default AdminManagement
