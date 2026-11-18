import React, { useEffect, useMemo, useState } from 'react'
import {
  Badge,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
  message,
} from 'antd'
import {
  DeleteOutlined,
  EditOutlined,
  KeyOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import {
  createPermission,
  deletePermission,
  getAllPermissions,
  getPermissionList,
  Permission,
  PermissionQueryParams,
  updatePermission,
} from '@/services/role'

const { Option } = Select

const typeText: Record<string, string> = {
  menu: '菜单',
  button: '按钮',
  api: '接口',
}

const PermissionManagement: React.FC = () => {
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [allPermissions, setAllPermissions] = useState<Permission[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingPermission, setEditingPermission] = useState<Permission | null>(null)
  const [searchParams, setSearchParams] = useState<PermissionQueryParams>({})
  const [form] = Form.useForm()

  const parentOptions = useMemo(
    () => allPermissions.filter((item) => item.id !== editingPermission?.id),
    [allPermissions, editingPermission],
  )

  const loadPermissions = async (params: PermissionQueryParams = searchParams) => {
    setLoading(true)
    try {
      const [tree, flat] = await Promise.all([getPermissionList(params), getAllPermissions()])
      setPermissions(tree || [])
      setAllPermissions(flat || [])
    } catch (error) {
      message.error('加载权限列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadPermissions({})
  }, [])

  const handleSearch = (values: PermissionQueryParams) => {
    setSearchParams(values)
    loadPermissions(values)
  }

  const handleReset = () => {
    setSearchParams({})
    loadPermissions({})
  }

  const handleCreate = () => {
    setEditingPermission(null)
    form.resetFields()
    form.setFieldsValue({ type: 'menu', status: 'active', sort_order: 0 })
    setModalVisible(true)
  }

  const handleEdit = (record: Permission) => {
    setEditingPermission(record)
    form.setFieldsValue({
      ...record,
      parent_id: record.parent_id ?? undefined,
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await deletePermission(id)
      message.success('删除成功')
      loadPermissions()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const payload = {
        ...values,
        parent_id: values.parent_id ?? null,
      }

      if (editingPermission) {
        await updatePermission(editingPermission.id, payload)
        message.success('更新成功')
      } else {
        await createPermission(payload)
        message.success('创建成功')
      }

      setModalVisible(false)
      loadPermissions()
    } catch (error) {
      message.error('保存失败')
    }
  }

  const columns: ColumnsType<Permission> = [
    {
      title: '权限名称',
      dataIndex: 'name',
      key: 'name',
      width: 180,
    },
    {
      title: '权限编码',
      dataIndex: 'code',
      key: 'code',
      width: 180,
      render: (code: string) => <Tag color='blue'>{code}</Tag>,
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 90,
      render: (type: string) => <Tag>{typeText[type] || type}</Tag>,
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      ellipsis: true,
      render: (path: string) => path || '-',
    },
    {
      title: '组件',
      dataIndex: 'component',
      key: 'component',
      width: 160,
      ellipsis: true,
      render: (component: string) => component || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status: string) => (
        <Badge
          status={status === 'active' ? 'success' : 'default'}
          text={status === 'active' ? '启用' : '禁用'}
        />
      ),
    },
    {
      title: '排序',
      dataIndex: 'sort_order',
      key: 'sort_order',
      width: 80,
    },
    {
      title: '操作',
      key: 'action',
      width: 170,
      render: (_, record) => (
        <Space size='small'>
          <Tooltip title='编辑'>
            <Button
              type='link'
              size='small'
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            >
              编辑
            </Button>
          </Tooltip>
          <Popconfirm
            title='确定要删除这个权限吗？'
            onConfirm={() => handleDelete(record.id)}
            okText='确定'
            cancelText='取消'
          >
            <Button type='link' size='small' danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <Card
      title={
        <Space>
          <KeyOutlined />
          权限管理
        </Space>
      }
      extra={
        <Button type='primary' icon={<PlusOutlined />} onClick={handleCreate}>
          新建权限
        </Button>
      }
    >
      <Form layout='inline' onFinish={handleSearch} initialValues={searchParams}>
        <Form.Item name='name' label='权限名称'>
          <Input placeholder='请输入权限名称' allowClear />
        </Form.Item>
        <Form.Item name='code' label='权限编码'>
          <Input placeholder='请输入权限编码' allowClear />
        </Form.Item>
        <Form.Item name='type' label='类型'>
          <Select placeholder='请选择类型' allowClear style={{ width: 120 }}>
            <Option value='menu'>菜单</Option>
            <Option value='button'>按钮</Option>
            <Option value='api'>接口</Option>
          </Select>
        </Form.Item>
        <Form.Item name='status' label='状态'>
          <Select placeholder='请选择状态' allowClear style={{ width: 120 }}>
            <Option value='active'>启用</Option>
            <Option value='inactive'>禁用</Option>
          </Select>
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type='primary' htmlType='submit' icon={<SearchOutlined />}>
              搜索
            </Button>
            <Button onClick={handleReset} icon={<ReloadOutlined />}>
              重置
            </Button>
          </Space>
        </Form.Item>
      </Form>

      <Table
        columns={columns}
        dataSource={permissions}
        loading={loading}
        rowKey='id'
        pagination={false}
        size='middle'
        style={{ marginTop: 16 }}
      />

      <Modal
        title={editingPermission ? '编辑权限' : '新建权限'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={680}
        destroyOnHidden
      >
        <Form
          form={form}
          layout='vertical'
          initialValues={{ type: 'menu', status: 'active', sort_order: 0 }}
        >
          <Form.Item
            name='name'
            label='权限名称'
            rules={[
              { required: true, message: '请输入权限名称' },
              { max: 100, message: '权限名称不能超过100个字符' },
            ]}
          >
            <Input placeholder='请输入权限名称' />
          </Form.Item>
          <Form.Item
            name='code'
            label='权限编码'
            rules={[
              { required: true, message: '请输入权限编码' },
              { pattern: /^[a-zA-Z0-9_:-]+$/, message: '权限编码只能包含字母、数字、下划线、冒号和短横线' },
              { max: 100, message: '权限编码不能超过100个字符' },
            ]}
          >
            <Input placeholder='请输入权限编码（如：case:view）' />
          </Form.Item>
          <Form.Item name='type' label='类型' rules={[{ required: true, message: '请选择类型' }]}>
            <Select placeholder='请选择类型'>
              <Option value='menu'>菜单</Option>
              <Option value='button'>按钮</Option>
              <Option value='api'>接口</Option>
            </Select>
          </Form.Item>
          <Form.Item name='parent_id' label='上级权限'>
            <Select placeholder='无上级权限' allowClear showSearch optionFilterProp='children'>
              {parentOptions.map((item) => (
                <Option key={item.id} value={item.id}>
                  {item.name} ({item.code})
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name='path' label='路径' rules={[{ max: 255, message: '路径不能超过255个字符' }]}>
            <Input placeholder='请输入路由或接口路径' />
          </Form.Item>
          <Form.Item
            name='component'
            label='组件'
            rules={[{ max: 255, message: '组件名不能超过255个字符' }]}
          >
            <Input placeholder='请输入前端组件标识' />
          </Form.Item>
          <Form.Item name='icon' label='图标' rules={[{ max: 100, message: '图标不能超过100个字符' }]}>
            <Input placeholder='请输入图标标识' />
          </Form.Item>
          <Form.Item name='status' label='状态' rules={[{ required: true, message: '请选择状态' }]}>
            <Select placeholder='请选择状态'>
              <Option value='active'>启用</Option>
              <Option value='inactive'>禁用</Option>
            </Select>
          </Form.Item>
          <Form.Item
            name='sort_order'
            label='排序'
            rules={[{ required: true, message: '请输入排序值' }]}
          >
            <InputNumber min={0} max={9999} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}

export default PermissionManagement
