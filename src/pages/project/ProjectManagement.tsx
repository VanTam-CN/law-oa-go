import React, { useState } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Input,
  Select,
  DatePicker,
  Modal,
  Form,
  message,
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

const { Search } = Input
const { Option } = Select
const { RangePicker } = DatePicker

interface Project {
  id: string
  name: string
  client: string
  type: string
  status: 'planning' | 'active' | 'completed' | 'suspended'
  startDate: string
  endDate?: string
  lawyer: string
  description: string
}

const ProjectManagement: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([
    {
      id: '1',
      name: '张三诉李四借款纠纷案',
      client: '张三',
      type: '民事诉讼',
      status: 'active',
      startDate: '2024-01-15',
      lawyer: '王律师',
      description: '民间借贷纠纷，涉及金额50万元',
    },
    {
      id: '2',
      name: 'ABC公司合同审查',
      client: 'ABC公司',
      type: '合同审查',
      status: 'planning',
      startDate: '2024-02-01',
      lawyer: '李律师',
      description: '企业合同条款审查和风险评估',
    },
    {
      id: '3',
      name: '知识产权保护咨询',
      client: '科技公司',
      type: '法律咨询',
      status: 'completed',
      startDate: '2023-12-01',
      endDate: '2024-01-10',
      lawyer: '赵律师',
      description: '商标注册和专利申请咨询',
    },
  ])

  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingProject, setEditingProject] = useState<Project | null>(null)
  const [form] = Form.useForm()

  const statusMap = {
    planning: { text: '计划中', color: 'blue' },
    active: { text: '进行中', color: 'green' },
    completed: { text: '已完成', color: 'default' },
    suspended: { text: '暂停', color: 'orange' },
  }

  const columns: ColumnsType<Project> = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      ellipsis: true,
    },
    {
      title: '客户',
      dataIndex: 'client',
      key: 'client',
      width: 120,
    },
    {
      title: '项目类型',
      dataIndex: 'type',
      key: 'type',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: keyof typeof statusMap) => (
        <Tag color={statusMap[status].color}>{statusMap[status].text}</Tag>
      ),
    },
    {
      title: '负责律师',
      dataIndex: 'lawyer',
      key: 'lawyer',
      width: 120,
    },
    {
      title: '开始日期',
      dataIndex: 'startDate',
      key: 'startDate',
      width: 120,
    },
    {
      title: '结束日期',
      dataIndex: 'endDate',
      key: 'endDate',
      width: 120,
      render: (date) => date || '-',
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
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Button
            type='link'
            size='small'
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record.id)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  const handleAdd = () => {
    setEditingProject(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (project: Project) => {
    setEditingProject(project)
    form.setFieldsValue(project)
    setModalVisible(true)
  }

  const handleDelete = (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个项目吗？',
      onOk: () => {
        setProjects(projects.filter((p) => p.id !== id))
        message.success('删除成功')
      },
    })
  }

  const handleSubmit = async (values: any) => {
    try {
      setLoading(true)

      if (editingProject) {
        // 编辑项目
        setProjects(projects.map((p) => (p.id === editingProject.id ? { ...p, ...values } : p)))
        message.success('项目更新成功')
      } else {
        // 新增项目
        const newProject: Project = {
          ...values,
          id: Date.now().toString(),
        }
        setProjects([...projects, newProject])
        message.success('项目创建成功')
      }

      setModalVisible(false)
      form.resetFields()
    } catch (error) {
      message.error('操作失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <div className='page-header'>
        <h1 className='page-title'>项目管理</h1>
        <p>管理律所的各类法律服务项目</p>
      </div>

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
            <Search
              placeholder='搜索项目名称或客户'
              allowClear
              style={{ width: 250 }}
              onSearch={(value) => console.log('搜索:', value)}
            />
            <Select placeholder='项目类型' style={{ width: 120 }} allowClear>
              <Option value='民事诉讼'>民事诉讼</Option>
              <Option value='刑事辩护'>刑事辩护</Option>
              <Option value='合同审查'>合同审查</Option>
              <Option value='法律咨询'>法律咨询</Option>
            </Select>
            <Select placeholder='状态' style={{ width: 100 }} allowClear>
              <Option value='planning'>计划中</Option>
              <Option value='active'>进行中</Option>
              <Option value='completed'>已完成</Option>
              <Option value='suspended'>暂停</Option>
            </Select>
          </Space>
          <Button type='primary' icon={<PlusOutlined />} onClick={handleAdd}>
            新建项目
          </Button>
        </div>

        <Table
          columns={columns}
          dataSource={projects}
          rowKey='id'
          pagination={{
            total: projects.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      <Modal
        title={editingProject ? '编辑项目' : '新建项目'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} layout='vertical' onFinish={handleSubmit}>
          <Form.Item
            name='name'
            label='项目名称'
            rules={[{ required: true, message: '请输入项目名称' }]}
          >
            <Input placeholder='请输入项目名称' />
          </Form.Item>

          <Form.Item
            name='client'
            label='客户'
            rules={[{ required: true, message: '请输入客户名称' }]}
          >
            <Input placeholder='请输入客户名称' />
          </Form.Item>

          <Form.Item
            name='type'
            label='项目类型'
            rules={[{ required: true, message: '请选择项目类型' }]}
          >
            <Select placeholder='请选择项目类型'>
              <Option value='民事诉讼'>民事诉讼</Option>
              <Option value='刑事辩护'>刑事辩护</Option>
              <Option value='合同审查'>合同审查</Option>
              <Option value='法律咨询'>法律咨询</Option>
              <Option value='知识产权'>知识产权</Option>
              <Option value='公司法务'>公司法务</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name='status'
            label='项目状态'
            rules={[{ required: true, message: '请选择项目状态' }]}
          >
            <Select placeholder='请选择项目状态'>
              <Option value='planning'>计划中</Option>
              <Option value='active'>进行中</Option>
              <Option value='completed'>已完成</Option>
              <Option value='suspended'>暂停</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name='lawyer'
            label='负责律师'
            rules={[{ required: true, message: '请输入负责律师' }]}
          >
            <Input placeholder='请输入负责律师' />
          </Form.Item>

          <Form.Item
            name='startDate'
            label='开始日期'
            rules={[{ required: true, message: '请选择开始日期' }]}
          >
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item name='endDate' label='结束日期'>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item name='description' label='项目描述'>
            <Input.TextArea rows={3} placeholder='请输入项目描述' />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalVisible(false)}>取消</Button>
              <Button type='primary' htmlType='submit' loading={loading}>
                {editingProject ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ProjectManagement
