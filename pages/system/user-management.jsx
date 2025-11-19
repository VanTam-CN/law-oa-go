import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Table, 
  Button, 
  Space, 
  Modal, 
  Form, 
  Input, 
  Select, 
  message, 
  Popconfirm,
  Tag,
  Tooltip,
  Upload,
  Drawer
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  ReloadOutlined, 
  SearchOutlined,
  KeyOutlined,
  UserOutlined,
  ImportOutlined,
  ExportOutlined,
  DownloadOutlined
} from '@ant-design/icons';
import { 
  getUserList, 
  createUser, 
  updateUser, 
  deleteUser, 
  resetUserPassword,
  changeUserStatus,
  exportUsers,
  importUsers,
  downloadUserTemplate,
  getDepartmentTree,
  getRoleOptions
} from '../../api/user';

const { Option } = Select;
const { Search } = Input;

export default function UserManagement() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [users, setUsers] = useState([]);
  const [departments, setDepartments] = useState([]);
  const [roles, setRoles] = useState([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [searchForm] = Form.useForm();
  const [modalVisible, setModalVisible] = useState(false);
  const [modalTitle, setModalTitle] = useState('');
  const [editingUser, setEditingUser] = useState(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);

  useEffect(() => {
    loadUsers();
    loadDepartments();
    loadRoles();
  }, [currentPage, pageSize]);

  const loadUsers = async (params = {}) => {
    setLoading(true);
    try {
      const data = await getUserList({
        pageNum: currentPage,
        pageSize,
        ...params
      });
      setUsers(data.rows || []);
      setTotal(data.total || 0);
    } catch (error) {
      message.error('加载用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  const loadDepartments = async () => {
    try {
      const data = await getDepartmentTree();
      setDepartments(data || []);
    } catch (error) {
      console.error('Failed to load departments:', error);
    }
  };

  const loadRoles = async () => {
    try {
      const data = await getRoleOptions();
      setRoles(data || []);
    } catch (error) {
      console.error('Failed to load roles:', error);
    }
  };

  const handleSearch = (values) => {
    setCurrentPage(1);
    loadUsers(values);
  };

  const handleReset = () => {
    searchForm.resetFields();
    setCurrentPage(1);
    loadUsers();
  };

  const handleAdd = () => {
    setModalTitle('新增用户');
    setEditingUser(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (user) => {
    setModalTitle('编辑用户');
    setEditingUser(user);
    form.setFieldsValue({
      ...user,
      roleIds: user.roleIds || []
    });
    setModalVisible(true);
  };

  const handleDelete = async (userId) => {
    try {
      await deleteUser([userId]);
      message.success('删除成功');
      loadUsers();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleResetPassword = async (user) => {
    try {
      await resetUserPassword({ userId: user.userId });
      message.success('密码重置成功，新密码为：123456');
    } catch (error) {
      message.error('密码重置失败');
    }
  };

  const handleStatusChange = async (user, status) => {
    try {
      await changeUserStatus({ userId: user.userId, status });
      message.success('状态修改成功');
      loadUsers();
    } catch (error) {
      message.error('状态修改失败');
    }
  };

  const handleSubmit = async (values) => {
    try {
      if (editingUser) {
        await updateUser({ ...values, userId: editingUser.userId });
        message.success('更新成功');
      } else {
        await createUser({ ...values, password: '123456' });
        message.success('创建成功');
      }
      setModalVisible(false);
      loadUsers();
    } catch (error) {
      message.error(editingUser ? '更新失败' : '创建失败');
    }
  };

  const handleExport = async () => {
    try {
      const blob = await exportUsers(searchForm.getFieldsValue());
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = '用户数据.xlsx';
      link.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      message.error('导出失败');
    }
  };

  const handleImport = async (file) => {
    try {
      const formData = new FormData();
      formData.append('file', file);
      await importUsers(formData, false);
      message.success('导入成功');
      loadUsers();
    } catch (error) {
      message.error('导入失败');
    }
    return false;
  };

  const handleDownloadTemplate = async () => {
    try {
      const blob = await downloadUserTemplate();
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = '用户导入模板.xlsx';
      link.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      message.error('下载模板失败');
    }
  };

  const columns = [
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      render: (text, record) => (
        <Space>
          <UserOutlined />
          {text}
        </Space>
      )
    },
    {
      title: '真实姓名',
      dataIndex: 'realName',
      key: 'realName'
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email'
    },
    {
      title: '手机号',
      dataIndex: 'phone',
      key: 'phone'
    },
    {
      title: '部门',
      dataIndex: 'departmentName',
      key: 'departmentName'
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const statusMap = {
          '0': { color: 'red', text: '禁用' },
          '1': { color: 'green', text: '正常' },
          '2': { color: 'orange', text: '冻结' }
        };
        const { color, text } = statusMap[status] || { color: 'default', text: '未知' };
        return <Tag color={color}>{text}</Tag>;
      }
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      key: 'createTime'
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Tooltip title="编辑">
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Tooltip title="重置密码">
            <Button
              type="link"
              icon={<KeyOutlined />}
              onClick={() => handleResetPassword(record)}
            />
          </Tooltip>
          {record.status === '1' ? (
            <Tooltip title="禁用">
              <Button
                type="link"
                danger
                onClick={() => handleStatusChange(record, '0')}
              >
                禁用
              </Button>
            </Tooltip>
          ) : (
            <Tooltip title="启用">
              <Button
                type="link"
                onClick={() => handleStatusChange(record, '1')}
              >
                启用
              </Button>
            </Tooltip>
          )}
          <Tooltip title="删除">
            <Popconfirm
              title="确定删除该用户吗？"
              onConfirm={() => handleDelete(record.userId)}
              okText="确定"
              cancelText="取消"
            >
              <Button
                type="link"
                danger
                icon={<DeleteOutlined />}
              />
            </Popconfirm>
          </Tooltip>
        </Space>
      )
    }
  ];

  const renderDepartmentTree = (depts) => {
    return depts.map(dept => (
      <Option key={dept.deptId} value={dept.deptId}>
        {dept.deptName}
      </Option>
    ));
  };

  return (
    <div className="user-management">
      <Card>
        <div className="table-toolbar">
          <Space>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleAdd}
            >
              新增用户
            </Button>
            <Button
              icon={<ImportOutlined />}
              onClick={() => {
                const input = document.createElement('input');
                input.type = 'file';
                input.accept = '.xlsx,.xls';
                input.onchange = (e) => {
                  if (e.target.files[0]) {
                    handleImport(e.target.files[0]);
                  }
                };
                input.click();
              }}
            >
              导入
            </Button>
            <Button
              icon={<ExportOutlined />}
              onClick={handleExport}
            >
              导出
            </Button>
            <Button
              icon={<DownloadOutlined />}
              onClick={handleDownloadTemplate}
            >
              下载模板
            </Button>
          </Space>
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => loadUsers()}
            >
              刷新
            </Button>
          </Space>
        </div>

        <Form
          form={searchForm}
          layout="inline"
          onFinish={handleSearch}
          className="search-form"
        >
          <Form.Item name="username">
            <Search
              placeholder="用户名"
              allowClear
              onSearch={handleSearch}
              style={{ width: 200 }}
            />
          </Form.Item>
          <Form.Item name="realName">
            <Input placeholder="真实姓名" allowClear style={{ width: 200 }} />
          </Form.Item>
          <Form.Item name="status">
            <Select placeholder="状态" allowClear style={{ width: 120 }}>
              <Option value="0">禁用</Option>
              <Option value="1">正常</Option>
              <Option value="2">冻结</Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
                搜索
              </Button>
              <Button onClick={handleReset}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>

        <Table
          columns={columns}
          dataSource={users}
          rowKey="userId"
          loading={loading}
          pagination={{
            current: currentPage,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `共 ${total} 条记录`,
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size);
            }
          }}
        />
      </Card>

      <Modal
        title={modalTitle}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input placeholder="请输入用户名" />
          </Form.Item>
          
          {!editingUser && (
            <Form.Item
              name="password"
              label="密码"
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password placeholder="请输入密码" />
            </Form.Item>
          )}

          <Form.Item
            name="realName"
            label="真实姓名"
            rules={[{ required: true, message: '请输入真实姓名' }]}
          >
            <Input placeholder="请输入真实姓名" />
          </Form.Item>

          <Form.Item
            name="email"
            label="邮箱"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' }
            ]}
          >
            <Input placeholder="请输入邮箱" />
          </Form.Item>

          <Form.Item
            name="phone"
            label="手机号"
            rules={[
              { required: true, message: '请输入手机号' },
              { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' }
            ]}
          >
            <Input placeholder="请输入手机号" />
          </Form.Item>

          <Form.Item
            name="deptId"
            label="所属部门"
            rules={[{ required: true, message: '请选择所属部门' }]}
          >
            <Select placeholder="请选择所属部门">
              {renderDepartmentTree(departments)}
            </Select>
          </Form.Item>

          <Form.Item
            name="roleIds"
            label="角色"
            rules={[{ required: true, message: '请选择角色' }]}
          >
            <Select mode="multiple" placeholder="请选择角色">
              {roles.map(role => (
                <Option key={role.roleId} value={role.roleId}>
                  {role.roleName}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingUser ? '更新' : '创建'}
              </Button>
              <Button onClick={() => setModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}