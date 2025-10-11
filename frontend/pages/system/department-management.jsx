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
  Tree,
  Tooltip,
  Switch,
  Select
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  ReloadOutlined,
  HomeOutlined,
  TeamOutlined,
  ApartmentOutlined
} from '@ant-design/icons';
import { 
  getDepartmentList, 
  createDepartment, 
  updateDepartment, 
  deleteDepartment
} from '../../api/department';

const { Option } = Select;

export default function DepartmentManagement() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [departments, setDepartments] = useState([]);
  const [treeData, setTreeData] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [modalTitle, setModalTitle] = useState('');
  const [editingDept, setEditingDept] = useState(null);
  const [selectedDept, setSelectedDept] = useState(null);
  const [expandedKeys, setExpandedKeys] = useState([]);

  useEffect(() => {
    loadDepartments();
  }, []);

  const loadDepartments = async () => {
    setLoading(true);
    try {
      const data = await getDepartmentList();
      setDepartments(data || []);
      buildTreeData(data || []);
    } catch (error) {
      message.error('加载部门列表失败');
    } finally {
      setLoading(false);
    }
  };

  const buildTreeData = (depts) => {
    const buildTree = (parentId) => {
      return depts
        .filter(dept => dept.parentId === parentId)
        .map(dept => ({
          key: dept.deptId,
          title: dept.deptName,
          children: buildTree(dept.deptId),
          data: dept
        }));
    };
    
    const tree = buildTree(0);
    setTreeData(tree);
    
    // 展开所有节点
    const expandKeys = [];
    const collectKeys = (nodes) => {
      nodes.forEach(node => {
        expandKeys.push(node.key);
        if (node.children && node.children.length > 0) {
          collectKeys(node.children);
        }
      });
    };
    collectKeys(tree);
    setExpandedKeys(expandKeys);
  };

  const handleAdd = (parentId = 0) => {
    setModalTitle('新增部门');
    setEditingDept(null);
    form.resetFields();
    form.setFieldsValue({ parentId });
    setModalVisible(true);
  };

  const handleEdit = (dept) => {
    setModalTitle('编辑部门');
    setEditingDept(dept);
    form.setFieldsValue(dept);
    setModalVisible(true);
  };

  const handleDelete = async (deptId) => {
    try {
      await deleteDepartment(deptId);
      message.success('删除成功');
      loadDepartments();
    } catch (error) {
      message.error(error.message || '删除失败');
    }
  };

  const handleSubmit = async (values) => {
    try {
      if (editingDept) {
        await updateDepartment({ ...values, deptId: editingDept.deptId });
        message.success('更新成功');
      } else {
        await createDepartment(values);
        message.success('创建成功');
      }
      setModalVisible(false);
      loadDepartments();
    } catch (error) {
      message.error(editingRole ? '更新失败' : '创建失败');
    }
  };

  const handleTreeSelect = (selectedKeys, info) => {
    if (selectedKeys.length > 0) {
      const dept = info.node.data;
      setSelectedDept(dept);
    } else {
      setSelectedDept(null);
    }
  };

  const renderDepartmentOptions = (depts, parentId = 0, prefix = '') => {
    return depts
      .filter(dept => dept.parentId === parentId)
      .map(dept => [
        <Option key={dept.deptId} value={dept.deptId}>
          {prefix + dept.deptName}
        </Option>,
        ...renderDepartmentOptions(depts, dept.deptId, prefix + '　　')
      ])
      .flat();
  };

  const tableColumns = [
    {
      title: '部门名称',
      dataIndex: 'deptName',
      key: 'deptName',
      render: (text, record) => (
        <Space>
          <HomeOutlined />
          <span style={{ fontWeight: 'bold' }}>{text}</span>
        </Space>
      )
    },
    {
      title: '负责人',
      dataIndex: 'leader',
      key: 'leader'
    },
    {
      title: '联系电话',
      dataIndex: 'phone',
      key: 'phone'
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email'
    },
    {
      title: '显示顺序',
      dataIndex: 'orderNum',
      key: 'orderNum',
      render: (order) => (
        <span style={{ 
          display: 'inline-block',
          width: '30px',
          height: '30px',
          lineHeight: '30px',
          textAlign: 'center',
          backgroundColor: '#f0f0f0',
          borderRadius: '50%',
          fontWeight: 'bold'
        }}>
          {order}
        </span>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status, record) => (
        <Switch
          checked={status === '1'}
          onChange={(checked) => {
            // TODO: 实现状态切换
            message.info('状态切换功能待实现');
          }}
          checkedChildren="正常"
          unCheckedChildren="禁用"
        />
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
          <Tooltip title="添加子部门">
            <Button
              type="link"
              icon={<PlusOutlined />}
              onClick={() => handleAdd(record.deptId)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Popconfirm
              title="确定删除该部门吗？"
              onConfirm={() => handleDelete(record.deptId)}
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
    <div className="department-management">
      <Card>
        <div className="dept-container">
          <div className="dept-tree">
            <Card 
              title="部门树" 
              size="small"
              extra={
                <Button
                  type="primary"
                  size="small"
                  icon={<PlusOutlined />}
                  onClick={() => handleAdd()}
                >
                  新增
                </Button>
              }
            >
              <Tree
                treeData={treeData}
                onSelect={handleTreeSelect}
                expandedKeys={expandedKeys}
                onExpand={setExpandedKeys}
                showLine
                showIcon
                icon={<HomeOutlined />}
              />
            </Card>
          </div>

          <div className="dept-table">
            <Card>
              <div className="table-toolbar">
                <Space>
                  <Button
                    type="primary"
                    icon={<PlusOutlined />}
                    onClick={() => handleAdd(selectedDept ? selectedDept.deptId : 0)}
                    disabled={!selectedDept}
                  >
                    新增子部门
                  </Button>
                </Space>
                <Space>
                  <Button
                    icon={<ReloadOutlined />}
                    onClick={() => loadDepartments()}
                  >
                    刷新
                  </Button>
                </Space>
              </div>

              <Table
                columns={tableColumns}
                dataSource={departments}
                rowKey="deptId"
                loading={loading}
                pagination={{
                  showSizeChanger: true,
                  showQuickJumper: true,
                  showTotal: (total, range) => `共 ${total} 条记录`,
                }}
              />
            </Card>
          </div>
        </div>
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
            name="parentId"
            label="上级部门"
            rules={[{ required: true, message: '请选择上级部门' }]}
          >
            <Select placeholder="请选择上级部门">
              <Option value={0}>顶级部门</Option>
              {renderDepartmentOptions(departments)}
            </Select>
          </Form.Item>

          <Form.Item
            name="deptName"
            label="部门名称"
            rules={[{ required: true, message: '请输入部门名称' }]}
          >
            <Input placeholder="请输入部门名称" />
          </Form.Item>

          <Form.Item
            name="leader"
            label="负责人"
          >
            <Input placeholder="请输入负责人" />
          </Form.Item>

          <Form.Item
            name="phone"
            label="联系电话"
            rules={[
              { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' }
            ]}
          >
            <Input placeholder="请输入联系电话" />
          </Form.Item>

          <Form.Item
            name="email"
            label="邮箱"
            rules={[
              { type: 'email', message: '请输入有效的邮箱地址' }
            ]}
          >
            <Input placeholder="请输入邮箱" />
          </Form.Item>

          <Form.Item
            name="orderNum"
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
                {editingDept ? '更新' : '创建'}
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