import React, { useState, useEffect } from 'react';
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
  DatePicker,
  InputNumber,
  message,
  Popconfirm,
  Statistic,
  Row,
  Col,
  Tooltip,
  Tabs,
  Avatar,
  Descriptions,
  Badge,
  Switch
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  EyeOutlined,
  UserOutlined,
  PhoneOutlined,
  MailOutlined,
  SearchOutlined,
  ReloadOutlined,
  CalendarOutlined,
  TrophyOutlined,
  BankOutlined,
  MedicineBoxOutlined,
  BankOutlined as GavelOutlined,
  UserSwitchOutlined,
  PauseCircleOutlined,
  CheckCircleOutlined,
  TeamOutlined,
  SettingOutlined,
  KeyOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { userService } from '@/api/user';
import { roleService } from '@/api/role';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';

const { Option } = Select;
const { Search } = Input;

interface User {
  userId?: number;
  username: string;
  realName: string;
  email: string;
  phone: string;
  status: '0' | '1';
  userType: '1' | '2' | '3' | '4';
  departmentId?: number;
  position: string;
  employeeId: string;
  avatar?: string;
  roleIds?: number[];
  departmentName?: string;
  roleNames?: string[];
  admin?: boolean;
}

interface Role {
  roleId: number;
  roleName: string;
  roleKey: string;
  status: '0' | '1';
}

interface UserStats {
  total: number;
  active: number;
  inactive: number;
  byType: Record<string, number>;
  byDepartment: Record<string, number>;
}

const UserManagement: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [viewingUser, setViewingUser] = useState<User | null>(null);
  const [form] = Form.useForm();
  const [searchText, setSearchText] = useState('');
  const [activeTab, setActiveTab] = useState('users');

  // 加载用户列表
  const loadUsers = async () => {
    setLoading(true);
    try {
      const response = await userService.getUserList();
      if (response && response.list) {
        setUsers(response.list);
      }
    } catch (error) {
      console.error('加载用户列表失败:', error);
      message.error('获取用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 加载角色列表
  const loadRoles = async () => {
    try {
      const response = await roleService.getRoleList();
      if (response && response.list) {
        setRoles(response.list);
      }
    } catch (error) {
      console.error('加载角色列表失败:', error);
      message.error('获取角色列表失败');
    }
  };

  // 加载统计数据
  const loadStats = async () => {
    try {
      const response = await userService.getUserStats();
      if (response) {
        setStats(response);
      }
    } catch (error) {
      console.error('加载统计数据失败:', error);
      // 如果没有统计数据API，则根据当前用户数据计算
      const calculatedStats: UserStats = {
        total: users.length,
        active: users.filter(u => u.status === '0').length,
        inactive: users.filter(u => u.status === '1').length,
        byType: {
          '1': users.filter(u => u.userType === '1').length,
          '2': users.filter(u => u.userType === '2').length,
          '3': users.filter(u => u.userType === '3').length,
          '4': users.filter(u => u.userType === '4').length,
        },
        byDepartment: {
          '诉讼部': users.filter(u => u.departmentId === 2).length,
          '非诉部': users.filter(u => u.departmentId === 3).length,
          '行政部': users.filter(u => u.departmentId === 4).length,
        }
      };
      setStats(calculatedStats);
    }
  };

  useEffect(() => {
    loadUsers();
    loadRoles();
  }, []);

  useEffect(() => {
    if (users.length > 0) {
      loadStats();
    }
  }, [users]);

  // 打开新增模态框
  const handleAdd = () => {
    setEditingUser(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 打开编辑模态框
  const handleEdit = (user: User) => {
    setEditingUser(user);
    form.setFieldsValue(user);
    setModalVisible(true);
  };

  // 查看用户详情
  const handleView = (user: User) => {
    setViewingUser(user);
    setDetailVisible(true);
  };

  // 删除用户
  const handleDelete = async (userId: number) => {
    try {
      await userService.deleteUser(userId);
      message.success('删除成功');
      loadUsers();
      loadStats();
    } catch (error) {
      console.error('删除用户失败:', error);
      message.error('删除失败');
    }
  };

  // 保存用户
  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);
      
      if (editingUser) {
        // 更新用户
        await userService.updateUser(editingUser.userId!, values);
        message.success('更新成功');
      } else {
        // 新增用户
        await userService.createUser(values);
        message.success('创建成功');
      }
      
      setModalVisible(false);
      loadUsers();
      loadStats();
    } catch (error) {
      console.error('保存用户失败:', error);
      message.error('保存失败');
    } finally {
      setLoading(false);
    }
  };

  // 切换用户状态
  const toggleUserStatus = async (user: User) => {
    try {
      const newStatus = user.status === '0' ? '1' : '0';
      await userService.changeUserStatus(user.userId!, newStatus);
      message.success('状态更新成功');
      loadUsers();
      loadStats();
    } catch (error) {
      console.error('更新用户状态失败:', error);
      message.error('状态更新失败');
    }
  };

  // 获取用户类型标签
  const getUserTypeTag = (userType: string) => {
    const typeMap = {
      '1': { text: '系统用户', color: 'red' },
      '2': { text: '律师', color: 'blue' },
      '3': { text: '助理', color: 'green' },
      '4': { text: '行政', color: 'orange' }
    };
    const config = typeMap[userType as keyof typeof typeMap] || { text: '未知', color: 'default' };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  // 获取状态标签
  const getStatusBadge = (status: string) => {
    return status === '0' ? 
      <Badge status="success" text="正常" /> : 
      <Badge status="error" text="禁用" />;
  };

  // 表格列定义
  const columns: ColumnsType<User> = [
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      render: (text, record) => (
        <Space>
          <Avatar icon={<UserOutlined />} src={record.avatar} />
          <span>{text}</span>
        </Space>
      ),
    },
    {
      title: '姓名',
      dataIndex: 'realName',
      key: 'realName',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '电话',
      dataIndex: 'phone',
      key: 'phone',
    },
    {
      title: '职位',
      dataIndex: 'position',
      key: 'position',
    },
    {
      title: '类型',
      dataIndex: 'userType',
      key: 'userType',
      render: getUserTypeTag,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: getStatusBadge,
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Tooltip title="查看详情">
            <Button 
              type="text" 
              icon={<EyeOutlined />} 
              onClick={() => handleView(record)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button 
              type="text" 
              icon={<EditOutlined />} 
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Tooltip title={record.status === '0' ? '禁用' : '启用'}>
            <Switch 
              checked={record.status === '0'} 
              onChange={() => toggleUserStatus(record)}
              size="small"
            />
          </Tooltip>
          <Popconfirm
            title="确定删除这个用户吗？"
            onConfirm={() => handleDelete(record.userId!)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button 
                type="text" 
                icon={<DeleteOutlined />} 
                danger
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 过滤用户数据
  const filteredUsers = users.filter(user =>
    Object.values(user).some(value =>
      String(value).toLowerCase().includes(searchText.toLowerCase())
    )
  );

  // 定义tabs items
  const tabItems = [
    {
      key: 'users',
      label: '用户管理',
      children: (
        <>
            {/* 统计卡片 */}
            {stats && (
              <Row gutter={16} style={{ marginBottom: 24 }}>
                <Col span={6}>
                  <Card>
                    <Statistic title="总用户数" value={stats.total} prefix={<TeamOutlined />} />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic title="正常用户" value={stats.active} prefix={<CheckCircleOutlined />} valueStyle={{ color: '#3f8600' }} />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic title="禁用用户" value={stats.inactive} prefix={<PauseCircleOutlined />} valueStyle={{ color: '#cf1322' }} />
                  </Card>
                </Col>
                <Col span={6}>
                  <Card>
                    <Statistic title="律师用户" value={stats.byType['2'] || 0} prefix={<BankOutlined />} />
                  </Card>
                </Col>
              </Row>
            )}

            {/* 操作栏 */}
            <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
              <Space>
                <Search
                  placeholder="搜索用户..."
                  allowClear
                  enterButton={<SearchOutlined />}
                  size="middle"
                  style={{ width: 300 }}
                  onSearch={setSearchText}
                />
                <Button 
                  icon={<ReloadOutlined />} 
                  onClick={loadUsers}
                >
                  刷新
                </Button>
              </Space>
              <Button 
                type="primary" 
                icon={<PlusOutlined />} 
                onClick={handleAdd}
              >
                新增用户
              </Button>
            </div>

            {/* 用户表格 */}
            <Table
              columns={columns}
              dataSource={filteredUsers}
              rowKey="userId"
              loading={loading}
              pagination={{
                total: filteredUsers.length,
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
              }}
            />
        </>
      ),
    },
    {
      key: 'roles',
      label: '角色管理',
      children: <RoleManagementContent />
    }
  ];

  return (
    <div className="user-management">
      <Card>
        <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
      </Card>

      {/* 用户编辑模态框 */}
      <Modal
        title={editingUser ? '编辑用户' : '新增用户'}
        open={modalVisible}
        onOk={handleSave}
        onCancel={() => setModalVisible(false)}
        width={600}
        confirmLoading={loading}
      >
        <Form
          form={form}
          layout="vertical"
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="用户名"
                name="username"
                rules={[{ required: true, message: '请输入用户名' }]}
              >
                <Input prefix={<UserOutlined />} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="姓名"
                name="realName"
                rules={[{ required: true, message: '请输入姓名' }]}
              >
                <Input />
              </Form.Item>
            </Col>
          </Row>
          
          {!editingUser && (
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  label="密码"
                  name="password"
                  rules={[{ required: true, message: '请输入密码' }]}
                >
                  <Input.Password prefix={<KeyOutlined />} />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  label="确认密码"
                  name="confirmPassword"
                  dependencies={['password']}
                  rules={[
                    { required: true, message: '请确认密码' },
                    ({ getFieldValue }) => ({
                      validator(_, value) {
                        if (!value || getFieldValue('password') === value) {
                          return Promise.resolve();
                        }
                        return Promise.reject(new Error('两次输入的密码不一致'));
                      },
                    }),
                  ]}
                >
                  <Input.Password prefix={<KeyOutlined />} />
                </Form.Item>
              </Col>
            </Row>
          )}
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="邮箱"
                name="email"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效的邮箱地址' }
                ]}
              >
                <Input prefix={<MailOutlined />} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="电话"
                name="phone"
                rules={[{ required: true, message: '请输入电话号码' }]}
              >
                <Input prefix={<PhoneOutlined />} />
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="用户类型"
                name="userType"
                rules={[{ required: true, message: '请选择用户类型' }]}
              >
                <Select>
                  <Option value="1">系统用户</Option>
                  <Option value="2">律师</Option>
                  <Option value="3">助理</Option>
                  <Option value="4">行政</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="职位"
                name="position"
                rules={[{ required: true, message: '请选择职位' }]}
              >
                <Select>
                  <Option value="系统管理员">系统管理员</Option>
                  <Option value="高级律师">高级律师</Option>
                  <Option value="律师">律师</Option>
                  <Option value="律师助理">律师助理</Option>
                  <Option value="行政专员">行政专员</Option>
                  <Option value="财务专员">财务专员</Option>
                  <Option value="实习生">实习生</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="员工编号"
                name="employeeId"
                rules={[{ required: true, message: '请输入员工编号' }]}
              >
                <Input 
                  placeholder="请输入员工编号"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="状态"
                name="status"
                rules={[{ required: true, message: '请选择状态' }]}
              >
                <Select>
                  <Option value="0">正常</Option>
                  <Option value="1">禁用</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>

      {/* 用户详情模态框 */}
      <Modal
        title="用户详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={800}
      >
        {viewingUser && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="用户名">{viewingUser.username}</Descriptions.Item>
            <Descriptions.Item label="姓名">{viewingUser.realName}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{viewingUser.email}</Descriptions.Item>
            <Descriptions.Item label="电话">{viewingUser.phone}</Descriptions.Item>
            <Descriptions.Item label="职位">{viewingUser.position}</Descriptions.Item>
            <Descriptions.Item label="员工编号">{viewingUser.employeeId}</Descriptions.Item>
            <Descriptions.Item label="用户类型">{getUserTypeTag(viewingUser.userType)}</Descriptions.Item>
            <Descriptions.Item label="状态">{getStatusBadge(viewingUser.status)}</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
};

// 角色管理内容组件
const RoleManagementContent: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [roles, setRoles] = useState<Role[]>([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [form] = Form.useForm();

  // 加载角色列表
  const loadRoles = async () => {
    setLoading(true);
    try {
      const response = await roleService.getRoleList();
      if (response && response.list) {
        setRoles(response.list);
      }
    } catch (error) {
      console.error('加载角色列表失败:', error);
      message.error('获取角色列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRoles();
  }, []);

  const handleSaveRole = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);
      
      if (editingRole) {
        // 更新角色
        await roleService.updateRole(editingRole.roleId, values);
        message.success('更新成功');
      } else {
        // 新增角色
        await roleService.createRole(values);
        message.success('创建成功');
      }
      
      setModalVisible(false);
      loadRoles();
    } catch (error) {
      console.error('保存角色失败:', error);
      message.error('保存失败');
    } finally {
      setLoading(false);
    }
  };

  const roleColumns: ColumnsType<Role> = [
    {
      title: '角色名称',
      dataIndex: 'roleName',
      key: 'roleName',
    },
    {
      title: '角色标识',
      dataIndex: 'roleKey',
      key: 'roleKey',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        status === '0' ? 
        <Badge status="error" text="禁用" /> : 
        <Badge status="success" text="正常" />
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button 
            type="text" 
            icon={<EditOutlined />} 
            onClick={() => {
              setEditingRole(record);
              form.setFieldsValue(record);
              setModalVisible(true);
            }}
          />
          <Button 
            type="text" 
            icon={<DeleteOutlined />} 
            danger
          />
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Button 
          type="primary" 
          icon={<PlusOutlined />} 
          onClick={() => {
            setEditingRole(null);
            form.resetFields();
            setModalVisible(true);
          }}
        >
          新增角色
        </Button>
      </div>

      <Table
        columns={roleColumns}
        dataSource={roles}
        rowKey="roleId"
        loading={loading}
        pagination={{
          total: roles.length,
          pageSize: 10,
          showSizeChanger: true,
          showQuickJumper: true,
        }}
      />

      <Modal
        title={editingRole ? '编辑角色' : '新增角色'}
        open={modalVisible}
        onOk={handleSaveRole}
        onCancel={() => setModalVisible(false)}
        confirmLoading={loading}
      >
        <Form
          form={form}
          layout="vertical"
        >
          <Form.Item
            label="角色名称"
            name="roleName"
            rules={[{ required: true, message: '请输入角色名称' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="角色标识"
            name="roleKey"
            rules={[{ required: true, message: '请输入角色标识' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="状态"
            name="status"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select>
              <Option value="1">正常</Option>
              <Option value="0">禁用</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default UserManagement;