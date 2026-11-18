import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Table, 
  Button, 
  Space, 
  Modal, 
  Form, 
  Input, 
  message, 
  Popconfirm,
  Tag,
  Tooltip,
  Switch
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  ReloadOutlined,
  SettingOutlined,
  UserOutlined
} from '@ant-design/icons';
import { 
  getRoleList, 
  createRole, 
  updateRole, 
  deleteRole, 
  changeRoleStatus
} from '../../api/role';

const { TextArea } = Input;

export default function RoleManagement() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [roles, setRoles] = useState([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [modalVisible, setModalVisible] = useState(false);
  const [modalTitle, setModalTitle] = useState('');
  const [editingRole, setEditingRole] = useState(null);

  useEffect(() => {
    loadRoles();
  }, [currentPage, pageSize]);

  const loadRoles = async (params = {}) => {
    setLoading(true);
    try {
      const data = await getRoleList({
        pageNum: currentPage,
        pageSize,
        ...params
      });
      setRoles(data.rows || []);
      setTotal(data.total || 0);
    } catch (error) {
      message.error('加载角色列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleAdd = () => {
    setModalTitle('新增角色');
    setEditingRole(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (role) => {
    setModalTitle('编辑角色');
    setEditingRole(role);
    form.setFieldsValue(role);
    setModalVisible(true);
  };

  const handleDelete = async (roleId) => {
    try {
      await deleteRole([roleId]);
      message.success('删除成功');
      loadRoles();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleStatusChange = async (role, status) => {
    try {
      await changeRoleStatus({ roleId: role.roleId, status });
      message.success('状态修改成功');
      loadRoles();
    } catch (error) {
      message.error('状态修改失败');
    }
  };

  const handleSubmit = async (values) => {
    try {
      if (editingRole) {
        await updateRole({ ...values, roleId: editingRole.roleId });
        message.success('更新成功');
      } else {
        await createRole(values);
        message.success('创建成功');
      }
      setModalVisible(false);
      loadRoles();
    } catch (error) {
      message.error(editingRole ? '更新失败' : '创建失败');
    }
  };

  const columns = [
    {
      title: '角色名称',
      dataIndex: 'roleName',
      key: 'roleName',
      render: (text) => (
        <Space>
          <SettingOutlined />
          <span style={{ fontWeight: 'bold' }}>{text}</span>
        </Space>
      )
    },
    {
      title: '角色标识',
      dataIndex: 'roleKey',
      key: 'roleKey',
      render: (text) => (
        <Tag color="blue">{text}</Tag>
      )
    },
    {
      title: '显示顺序',
      dataIndex: 'roleSort',
      key: 'roleSort',
      render: (sort) => (
        <Tag color="orange">{sort}</Tag>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status, record) => (
        <Switch
          checked={status === '1'}
          onChange={(checked) => handleStatusChange(record, checked ? '1' : '0')}
          checkedChildren="正常"
          unCheckedChildren="禁用"
        />
      )
    },
    {
      title: '用户数',
      dataIndex: 'userCount',
      key: 'userCount',
      render: (count) => (
        <Space>
          <UserOutlined />
          <span>{count || 0}</span>
        </Space>
      )
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
          <Tooltip title="删除">
            <Popconfirm
              title="确定删除该角色吗？"
              onConfirm={() => handleDelete(record.roleId)}
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

  return (
    <div className="role-management">
      <Card>
        <div className="table-toolbar">
          <Space>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleAdd}
            >
              新增角色
            </Button>
          </Space>
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => loadRoles()}
            >
              刷新
            </Button>
          </Space>
        </div>

        <Table
          columns={columns}
          dataSource={roles}
          rowKey="roleId"
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
            name="roleName"
            label="角色名称"
            rules={[{ required: true, message: '请输入角色名称' }]}
          >
            <Input placeholder="请输入角色名称" />
          </Form.Item>

          <Form.Item
            name="roleKey"
            label="角色标识"
            rules={[
              { required: true, message: '请输入角色标识' },
              { pattern: /^[A-Z_]+$/, message: '角色标识只能包含大写字母和下划线' }
            ]}
          >
            <Input placeholder="请输入角色标识（如：LAWYER）" />
          </Form.Item>

          <Form.Item
            name="roleSort"
            label="显示顺序"
            rules={[{ required: true, message: '请输入显示顺序' }]}
            initialValue={0}
          >
            <Input 
              type="number" 
              placeholder="请输入显示顺序" 
              min={0}
              max={999}
            />
          </Form.Item>

          <Form.Item
            name="remark"
            label="备注"
          >
            <TextArea 
              rows={4} 
              placeholder="请输入备注信息" 
              maxLength={500}
              showCount
            />
          </Form.Item>

          <Form.Item
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态' }]}
            initialValue="1"
          >
            <Select placeholder="请选择状态">
              <Option value="0">禁用</Option>
              <Option value="1">正常</Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingRole ? '更新' : '创建'}
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