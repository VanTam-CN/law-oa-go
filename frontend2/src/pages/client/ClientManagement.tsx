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
  message,
  Popconfirm,
  Statistic,
  Row,
  Col,
  Tooltip
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  EyeOutlined,
  UserOutlined,
  BankOutlined,
  SearchOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { clientService } from '@/api/client';
import './ClientManagement.less';

interface Client {
  id?: number;
  name: string;
  type: string;
  contact: string;
  phone: string;
  email: string;
  address: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
}

interface ClientStats {
  total: number;
  active: number;
  inactive: number;
  byType: Record<string, number>;
}

const { Option } = Select;
const { TextArea } = Input;

const ClientManagement: React.FC = () => {
  const [clients, setClients] = useState<Client[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [modalTitle, setModalTitle] = useState<string>('');
  const [editingClient, setEditingClient] = useState<Client | null>(null);
  const [stats, setStats] = useState<ClientStats | null>(null);
  const [form] = Form.useForm();
  
  // 查询参数
  const [queryParams, setQueryParams] = useState({
    name: '',
    type: '',
    status: '',
    pageNum: 1,
    pageSize: 10
  });
  
  const [total, setTotal] = useState<number>(0);

  // 获取客户列表
  const fetchClients = async () => {
    setLoading(true);
    try {
      const res = await clientService.getClientList(queryParams);
      setClients(res.data || res.list || []);
      setTotal(res.pagination?.total || res.total || 0);
    } catch (error) {
      message.error('获取客户列表失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取客户统计
  const fetchStats = async () => {
    try {
      const res = await clientService.getClientStats();
      setStats(res);
    } catch (error) {
      message.error('获取统计数据失败');
    }
  };

  useEffect(() => {
    fetchClients();
    fetchStats();
  }, [queryParams]);

  // 打开新增客户弹窗
  const handleAdd = () => {
    setModalTitle('新增客户');
    setEditingClient(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 打开编辑客户弹窗
  const handleEdit = (record: Client) => {
    setModalTitle('编辑客户');
    setEditingClient(record);
    form.setFieldsValue(record);
    setModalVisible(true);
  };

  // 查看客户详情
  const handleView = (record: Client) => {
    Modal.info({
      title: '客户详情',
      width: 600,
      content: (
        <div className="client-detail">
          <p><strong>客户名称：</strong>{record.name}</p>
          <p><strong>客户类型：</strong>{record.type === '个人' ? '个人' : '企业'}</p>
          <p><strong>联系电话：</strong>{record.phone}</p>
          <p><strong>电子邮箱：</strong>{record.email}</p>
          {record.type === '个人' && record.idCard && (
            <p><strong>身份证号：</strong>{record.idCard}</p>
          )}
          <p><strong>地址：</strong>{record.address}</p>
          {record.type === '企业' && (
            <>
              <p><strong>公司名称：</strong>{record.company}</p>
              <p><strong>所属行业：</strong>{record.industry}</p>
              <p><strong>联系人：</strong>{record.contactPerson}</p>
              <p><strong>联系电话：</strong>{record.contactPhone}</p>
            </>
          )}
          <p><strong>客户来源：</strong>{record.source || '-'}</p>
          <p><strong>客户状态：</strong>
            <Tag color={record.status === 'active' ? 'green' : 'red'}>
              {record.status === 'active' ? '活跃' : '非活跃'}
            </Tag>
          </p>
          <p><strong>备注：</strong>{record.remark || '-'}</p>
        </div>
      )
    });
  };

  // 删除客户
  const handleDelete = async (id: number) => {
    try {
      await clientService.deleteClient(id);
      message.success('删除成功');
      fetchClients();
      fetchStats();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingClient) {
        await clientService.updateClient(editingClient.id!, values);
        message.success('更新成功');
      } else {
        await clientService.createClient(values);
        message.success('新增成功');
      }
      setModalVisible(false);
      fetchClients();
      fetchStats();
    } catch (error: any) {
      message.error(error.message || '操作失败');
    }
  };

  // 搜索
  const handleSearch = () => {
    setQueryParams({ ...queryParams, pageNum: 1 });
  };

  // 重置搜索
  const handleReset = () => {
    setQueryParams({
      name: '',
      type: '',
      status: '',
      pageNum: 1,
      pageSize: 10
    });
  };

  // 表格列定义
  const columns: ColumnsType<Client> = [
    {
      title: '客户名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Client) => (
        <Space>
          {record.type === '个人' ? <UserOutlined /> : <BankOutlined />}
          {text}
        </Space>
      ),
    },
    {
      title: '客户类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color={type === '个人' ? 'blue' : 'purple'}>
          {type === '个人' ? '个人' : '企业'}
        </Tag>
      ),
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
        <Space size="middle">
          <Tooltip title="查看详情">
            <Button 
              type="link" 
              icon={<EyeOutlined />} 
              onClick={() => handleView(record)}
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
              title="确定要删除这个客户吗？"
              onConfirm={() => handleDelete(record.id!)}
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
      ),
    },
  ];

  return (
    <div className="client-management">
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className="stats-row">
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="客户总数" 
              value={stats?.total || 0} 
              prefix={<UserOutlined />} 
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="活跃客户" 
              value={stats?.statusStats?.active || 0} 
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="个人客户" 
              value={stats?.typeStats?.['个人'] || 0} 
              prefix={<UserOutlined />} 
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="企业客户" 
              value={stats?.typeStats?.['企业'] || 0} 
              prefix={<BankOutlined />} 
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索表单 */}
      <Card className="search-card">
        <Form layout="inline">
          <Form.Item label="客户名称">
            <Input 
              placeholder="请输入客户名称" 
              value={queryParams.name}
              onChange={(e) => setQueryParams({ ...queryParams, name: e.target.value })}
              allowClear
            />
          </Form.Item>
          <Form.Item label="客户类型">
            <Select 
              style={{ width: 120 }}
              value={queryParams.type}
              onChange={(value) => setQueryParams({ ...queryParams, type: value })}
              allowClear
              placeholder="全部"
            >
              <Option value="个人">个人</Option>
              <Option value="企业">企业</Option>
            </Select>
          </Form.Item>
          <Form.Item label="客户状态">
            <Select 
              style={{ width: 120 }}
              value={queryParams.status}
              onChange={(value) => setQueryParams({ ...queryParams, status: value })}
              allowClear
              placeholder="全部"
            >
              <Option value="active">活跃</Option>
              <Option value="inactive">非活跃</Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
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
        title="客户列表" 
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            新增客户
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={clients}
          rowKey="id"
          loading={loading}
          pagination={{
            current: queryParams.pageNum,
            pageSize: queryParams.pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            onChange: (page, size) => {
              setQueryParams({
                ...queryParams,
                pageNum: page,
                pageSize: size || 10
              });
            }
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
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
        >
          <Form.Item
            label="客户类型"
            name="type"
            rules={[{ required: true, message: '请选择客户类型' }]}
          >
            <Radio.Group>
              <Radio value="个人">个人</Radio>
              <Radio value="企业">企业</Radio>
            </Radio.Group>
          </Form.Item>

          <Form.Item
            label={form.getFieldValue('type') === '企业' ? '公司名称' : '客户姓名'}
            name="name"
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input placeholder="请输入名称" />
          </Form.Item>

          <Form.Item
            label="联系电话"
            name="phone"
            rules={[{ required: true, message: '请输入联系电话' }]}
          >
            <Input placeholder="请输入联系电话" />
          </Form.Item>

          <Form.Item
            label="电子邮箱"
            name="email"
            rules={[
              { required: true, message: '请输入电子邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' }
            ]}
          >
            <Input placeholder="请输入电子邮箱" />
          </Form.Item>

          {form.getFieldValue('type') === '个人' && (
            <Form.Item
              label="身份证号"
              name="idCard"
              rules={[
                { pattern: /^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$/, message: '请输入有效的身份证号' }
              ]}
            >
              <Input placeholder="请输入身份证号" />
            </Form.Item>
          )}

          {form.getFieldValue('type') === '企业' && (
            <>
              <Form.Item
                label="所属行业"
                name="industry"
              >
                <Input placeholder="请输入所属行业" />
              </Form.Item>
              <Form.Item
                label="联系人"
                name="contactPerson"
              >
                <Input placeholder="请输入联系人" />
              </Form.Item>
              <Form.Item
                label="联系电话"
                name="contactPhone"
              >
                <Input placeholder="请输入联系电话" />
              </Form.Item>
            </>
          )}

          <Form.Item
            label="地址"
            name="address"
            rules={[{ required: true, message: '请输入地址' }]}
          >
            <TextArea placeholder="请输入详细地址" rows={2} />
          </Form.Item>

          <Form.Item
            label="客户来源"
            name="source"
          >
            <Select placeholder="请选择客户来源">
              <Option value="推荐">推荐</Option>
              <Option value="自主开发">自主开发</Option>
              <Option value="网络推广">网络推广</Option>
              <Option value="合作机构">合作机构</Option>
              <Option value="其他">其他</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="客户状态"
            name="status"
            rules={[{ required: true, message: '请选择客户状态' }]}
          >
            <Radio.Group>
              <Radio value="active">活跃</Radio>
              <Radio value="inactive">非活跃</Radio>
            </Radio.Group>
          </Form.Item>

          <Form.Item
            label="备注"
            name="remark"
          >
            <TextArea placeholder="请输入备注信息" rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ClientManagement;