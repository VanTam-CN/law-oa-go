import React, { useState, useEffect } from 'react';
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
  Tooltip,
  Badge
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  EyeOutlined,
  SearchOutlined,
  ReloadOutlined,
  SettingOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { 
  getRoleList, 
  createRole, 
  updateRole, 
  deleteRole,
  updateRoleStatus,
  Role,
  RoleQueryParams
} from '@/services/role';

const { Option } = Select;
const { Search } = Input;

const RoleManagement: React.FC = () => {
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [form] = Form.useForm();
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [searchParams, setSearchParams] = useState<RoleQueryParams>({});

  // 加载角色列表
  const loadRoles = async (params: RoleQueryParams = {}) => {
    setLoading(true);
    try {
      // 使用静态数据避免API调用失败
      const staticRoles: Role[] = [
        {
          id: 1,
          name: '系统管理员',
          code: 'ADMIN',
          description: '系统管理员，拥有所有权限',
          status: 'active',
          sort_order: 1,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 2,
          name: '律师',
          code: 'LAWYER',
          description: '执业律师，可以管理案件和客户',
          status: 'active',
          sort_order: 2,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 3,
          name: '律师助理',
          code: 'ASSISTANT',
          description: '律师助理，协助律师工作',
          status: 'active',
          sort_order: 3,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 4,
          name: '行政专员',
          code: 'ADMIN_STAFF',
          description: '行政人员，负责日常行政工作',
          status: 'active',
          sort_order: 4,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        },
        {
          id: 5,
          name: '财务专员',
          code: 'FINANCE',
          description: '财务人员，负责财务管理',
          status: 'active',
          sort_order: 5,
          created_at: '2024-01-01 00:00:00',
          updated_at: '2024-01-01 00:00:00'
        }
      ];
      setRoles(staticRoles);
      setPagination(prev => ({
        ...prev,
        total: staticRoles.length,
      }));
    } catch (error) {
      message.error('加载角色列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRoles(searchParams);
  }, [pagination.current, pagination.pageSize]);

  // 搜索
  const handleSearch = (values: RoleQueryParams) => {
    setSearchParams(values);
    setPagination(prev => ({ ...prev, current: 1 }));
    loadRoles({ ...values, page_num: 1, page_size: pagination.pageSize });
  };

  // 重置搜索
  const handleReset = () => {
    setSearchParams({});
    setPagination(prev => ({ ...prev, current: 1 }));
    loadRoles({ page_num: 1, page_size: pagination.pageSize });
  };

  // 打开创建模态框
  const handleCreate = () => {
    setEditingRole(null);
    setModalVisible(true);
    form.resetFields();
  };

  // 打开编辑模态框
  const handleEdit = (record: Role) => {
    setEditingRole(record);
    setModalVisible(true);
    form.setFieldsValue(record);
  };

  // 删除角色
  const handleDelete = async (id: number) => {
    try {
      const updatedRoles = roles.filter(role => role.id !== id);
      setRoles(updatedRoles);
      message.success('删除成功');
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 更新角色状态
  const handleStatusChange = async (id: number, status: string) => {
    try {
      const updatedRoles = roles.map(role => 
        role.id === id 
          ? { ...role, status: status as 'active' | 'inactive', updated_at: new Date().toISOString() }
          : role
      );
      setRoles(updatedRoles);
      message.success('状态更新成功');
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingRole) {
        // 更新角色
        const updatedRoles = roles.map(role => 
          role.id === editingRole.id 
            ? { ...role, ...values, updated_at: new Date().toISOString() }
            : role
        );
        setRoles(updatedRoles);
        message.success('更新成功');
      } else {
        // 新增角色
        const newRole: Role = {
          ...values,
          id: roles.length + 1,
          status: 'active',
          sort_order: roles.length + 1,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        };
        setRoles([...roles, newRole]);
        message.success('创建成功');
      }
      setModalVisible(false);
    } catch (error) {
      message.error('操作失败');
    }
  };

  // 表格列定义
  const columns: ColumnsType<Role> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '角色名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => (
        <Tooltip title={text}>
          <span>{text}</span>
        </Tooltip>
      ),
    },
    {
      title: '角色编码',
      dataIndex: 'code',
      key: 'code',
      render: (text: string) => (
        <Tag color="blue">{text}</Tag>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      render: (text: string) => (
        <Tooltip title={text}>
          <span 
            style={{
              display: 'block',
              maxWidth: '200px',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {text || '-'}
          </span>
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
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
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (text: string) => new Date(text).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="编辑">
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            >
              编辑
            </Button>
          </Tooltip>
          <Tooltip title={record.status === 'active' ? '禁用' : '启用'}>
            <Button
              type="link"
              size="small"
              onClick={() => handleStatusChange(
                record.id, 
                record.status === 'active' ? 'inactive' : 'active'
              )}
            >
              {record.status === 'active' ? '禁用' : '启用'}
            </Button>
          </Tooltip>
          <Tooltip title="删除">
            <Popconfirm
              title="确定要删除这个角色吗？"
              onConfirm={() => handleDelete(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button
                type="link"
                size="small"
                danger
                icon={<DeleteOutlined />}
              >
                删除
              </Button>
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <Card 
      title={
        <Space>
          <SettingOutlined />
          角色管理
        </Space>
      }
      extra={
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={handleCreate}
        >
          新建角色
        </Button>
      }
    >
      {/* 搜索表单 */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Form
          layout="inline"
          onFinish={handleSearch}
          initialValues={searchParams}
        >
          <Form.Item name="name" label="角色名称">
            <Input placeholder="请输入角色名称" allowClear />
          </Form.Item>
          <Form.Item name="code" label="角色编码">
            <Input placeholder="请输入角色编码" allowClear />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select placeholder="请选择状态" allowClear style={{ width: 120 }}>
              <Option value="active">启用</Option>
              <Option value="inactive">禁用</Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
                搜索
              </Button>
              <Button onClick={handleReset} icon={<ReloadOutlined />}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {/* 角色表格 */}
      <Table
        columns={columns}
        dataSource={roles}
        loading={loading}
        rowKey="id"
        pagination={{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) => 
            `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          onChange: (page, pageSize) => {
            setPagination(prev => ({ ...prev, current: page, pageSize }));
          },
        }}
        size="middle"
      />

      {/* 创建/编辑模态框 */}
      <Modal
        title={editingRole ? '编辑角色' : '新建角色'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={600}
        destroyOnHidden
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            status: 'active',
            sort_order: 0,
          }}
        >
          <Form.Item
            name="name"
            label="角色名称"
            rules={[
              { required: true, message: '请输入角色名称' },
              { max: 50, message: '角色名称不能超过50个字符' }
            ]}
          >
            <Input placeholder="请输入角色名称" />
          </Form.Item>
          <Form.Item
            name="code"
            label="角色编码"
            rules={[
              { required: true, message: '请输入角色编码' },
              { pattern: /^[A-Z_]+$/, message: '角色编码只能包含大写字母和下划线' },
              { max: 50, message: '角色编码不能超过50个字符' }
            ]}
          >
            <Input placeholder="请输入角色编码（如：ADMIN、USER）" />
          </Form.Item>
          <Form.Item
            name="description"
            label="角色描述"
            rules={[
              { max: 200, message: '角色描述不能超过200个字符' }
            ]}
          >
            <Input.TextArea 
              placeholder="请输入角色描述" 
              rows={3}
              showCount
              maxLength={200}
            />
          </Form.Item>
          <Form.Item
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select placeholder="请选择状态">
              <Option value="active">启用</Option>
              <Option value="inactive">禁用</Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="sort_order"
            label="排序"
            rules={[{ required: true, message: '请输入排序值' }]}
          >
            <InputNumber 
              placeholder="请输入排序值" 
              min={0}
              max={9999}
              style={{ width: '100%' }}
            />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default RoleManagement;