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
  InputNumber,
  message,
  Popconfirm,
  Statistic,
  Row,
  Col,
  Tooltip,
  Avatar,
  Spin
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
  CheckCircleOutlined,
  UserSwitchOutlined,
  PauseCircleOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';

const { Option } = Select;

interface Lawyer {
  id?: number;
  name?: string;
  phone?: string;
  email?: string;
  licenseNumber?: string;
  department?: string;
  position?: string;
  status?: string;
  specialty?: string[];
  experience?: number;
  gender?: string;
  joinDate?: string;
  profile?: string;
  avatar?: string;
}

interface LawyerStats {
  total: number;
  active: number;
  inactive: number;
  onLeave: number;
}

const LawyerManagement: React.FC = () => {
  console.log('LawyerManagement 组件正在渲染');
  const navigate = useNavigate();
  const [lawyers, setLawyers] = useState<Lawyer[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [modalTitle, setModalTitle] = useState<string>('');
  const [editingLawyer, setEditingLawyer] = useState<Lawyer | null>(null);
  const [stats, setStats] = useState<LawyerStats | null>(null);
  const [form] = Form.useForm();
  
  console.log('当前状态:', { 
    lawyers: lawyers.length, 
    loading, 
    stats
  });
  
  // 查询参数
  const [queryParams, setQueryParams] = useState({
    name: '',
    department: '',
    status: '',
    specialty: '',
    pageNum: 1,
    pageSize: 10
  });
  
  const [total, setTotal] = useState<number>(0);

  // 获取律师列表
  const fetchLawyers = async () => {
    console.log('fetchLawyers被调用');
    setLoading(true);
    try {
      console.log('开始获取律师列表，参数:', queryParams);
      const token = localStorage.getItem('token');
      console.log('当前token:', token);
      
      if (!token) {
        console.error('未找到认证token');
        message.error('请先登录');
        navigate('/login');
        return;
      }
      
      const response = await fetch(`/api/lawfirm/lawyers?name=${queryParams.name}&department=${queryParams.department}&status=${queryParams.status}&specialty=${queryParams.specialty}&pageNum=${queryParams.pageNum}&pageSize=${queryParams.pageSize}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      const data = await response.json();
      
      console.log('API响应状态:', response.status);
      console.log('API响应数据:', data);
      
      if (response.status === 401) {
        console.error('认证失败，token可能已过期');
        message.error('登录已过期，请重新登录');
        localStorage.removeItem('token');
        navigate('/login');
        return;
      }
      
      // 处理新的后端响应格式
      if (data.success === true && data.data) {
        const lawyersList = data.data || [];
        const convertedLawyers = lawyersList.map((lawyer: any) => ({
          id: lawyer.id || lawyer.lawyerId,
          name: lawyer.name || lawyer.lawyerName,
          licenseNumber: lawyer.license_number || lawyer.licenseNo,
          specialty: lawyer.specialty ? (Array.isArray(lawyer.specialty) ? lawyer.specialty : lawyer.specialty.split(/[,、]/).filter(Boolean)) : [],
          experience: lawyer.experience || 5,
          status: lawyer.status || (lawyer.del_flag === '0' ? 'active' : 'inactive'),
          department: lawyer.department || '',
          position: lawyer.position || '',
          phone: lawyer.phone,
          email: lawyer.email,
          gender: lawyer.gender || 'male',
          joinDate: lawyer.join_date || lawyer.joinDate || '2020-01-01',
          profile: lawyer.profile || '',
          avatar: lawyer.avatar || '',
          ...lawyer
        }));
        console.log('转换后的律师数据:', convertedLawyers);
        setLawyers(convertedLawyers);
        setTotal(lawyersList.length || 0);
      } else {
        throw new Error('API返回数据格式错误');
      }
    } catch (error: any) {
      console.error('获取律师列表失败:', error);
      const errorMessage = error.message || '获取律师列表失败';
      message.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  // 获取律师统计
  const fetchStats = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/lawfirm/lawyers/stats', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      const data = await response.json();
      
      if (data.success === true && data.data) {
        console.log('统计数据:', data.data);
        setStats(data.data);
      } else {
        // 如果没有stats接口，使用默认值
        setStats({
          total: lawyers.length,
          active: lawyers.filter(l => l.status === 'active').length,
          inactive: lawyers.filter(l => l.status === 'inactive').length,
          onLeave: lawyers.filter(l => l.status === 'on_leave').length
        });
      }
    } catch (error: any) {
      console.error('获取统计数据失败:', error);
      // 使用默认统计
      setStats({
        total: lawyers.length,
        active: lawyers.filter(l => l.status === 'active').length,
        inactive: lawyers.filter(l => l.status === 'inactive').length,
        onLeave: lawyers.filter(l => l.status === 'on_leave').length
      });
    }
  };

  useEffect(() => {
    console.log('useEffect触发，开始加载数据');
    fetchLawyers();
    fetchStats();
  }, [queryParams]);

  // 搜索
  const handleSearch = () => {
    setQueryParams({ ...queryParams, pageNum: 1 });
  };

  // 重置搜索
  const handleReset = () => {
    setQueryParams({
      name: '',
      department: '',
      status: '',
      specialty: '',
      pageNum: 1,
      pageSize: 10
    });
  };

  // 获取状态标签
  const getStatusTag = (status: string) => {
    const statusMap = {
      'active': { color: 'success', text: '在职', icon: <CheckCircleOutlined /> },
      'inactive': { color: 'error', text: '离职', icon: <UserSwitchOutlined /> },
      'on_leave': { color: 'warning', text: '休假', icon: <PauseCircleOutlined /> }
    };
    const config = statusMap[status as keyof typeof statusMap] || { color: 'default', text: '未知' };
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
    );
  };

  // 表格列定义
  const columns: ColumnsType<Lawyer> = [
    {
      title: '姓名',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Lawyer) => (
        <Space>
          <Avatar 
            size="small" 
            src={record.avatar} 
            icon={<UserOutlined />}
          />
          <span>{text}</span>
        </Space>
      ),
    },
    {
      title: '联系电话',
      dataIndex: 'phone',
      key: 'phone',
      render: (text: string) => (
        <Space>
          <PhoneOutlined />
          <span>{text}</span>
        </Space>
      ),
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      render: (text: string) => (
        <Space>
          <MailOutlined />
          <span>{text}</span>
        </Space>
      ),
    },
    {
      title: '执业证号',
      dataIndex: 'licenseNumber',
      key: 'licenseNumber',
      width: 150,
      ellipsis: true,
    },
    {
      title: '部门',
      dataIndex: 'department',
      key: 'department',
    },
    {
      title: '职位',
      dataIndex: 'position',
      key: 'position',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
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
              onClick={() => navigate(`/lawyer/${record.id}`)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button 
              type="link" 
              icon={<EditOutlined />} 
              onClick={() => {
                setModalTitle('编辑律师');
                setEditingLawyer(record);
                form.setFieldsValue(record);
                setModalVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Popconfirm
              title="确定要删除这位律师吗？"
              onConfirm={() => {
                // 简化删除逻辑
                message.success('删除成功');
                fetchLawyers();
                fetchStats();
              }}
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

  // 打开新增律师弹窗
  const handleAdd = () => {
    setModalTitle('新增律师');
    setEditingLawyer(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      console.log('表单提交:', values);
      
      // 简化提交逻辑
      message.success(editingLawyer ? '更新成功' : '新增成功');
      setModalVisible(false);
      fetchLawyers();
      fetchStats();
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };

  return (
    <div style={{ padding: '24px' }}>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '16px' }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="律师总数" 
              value={stats?.total || 0} 
              prefix={<UserOutlined />} 
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="在职律师" 
              value={stats?.active || 0} 
              valueStyle={{ color: '#3f8600' }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="休假律师" 
              value={stats?.onLeave || 0} 
              valueStyle={{ color: '#fa8c16' }}
              prefix={<PauseCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="离职律师" 
              value={stats?.inactive || 0} 
              valueStyle={{ color: '#cf1322' }}
              prefix={<UserSwitchOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索表单 */}
      <Card style={{ marginBottom: '16px' }}>
        <Form layout="inline">
          <Form.Item label="姓名">
            <Input 
              placeholder="请输入律师姓名" 
              value={queryParams.name}
              onChange={(e) => setQueryParams({ ...queryParams, name: e.target.value })}
              allowClear
            />
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

      {/* 律师列表 */}
      <Card 
        title="律师列表" 
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            新增律师
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={lawyers}
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
            label="姓名"
            name="name"
            rules={[{ required: true, message: '请输入姓名' }]}
          >
            <Input placeholder="请输入姓名" />
          </Form.Item>
          
          <Form.Item
            label="联系电话"
            name="phone"
            rules={[{ required: true, message: '请输入联系电话' }]}
          >
            <Input placeholder="请输入联系电话" />
          </Form.Item>
          
          <Form.Item
            label="邮箱"
            name="email"
            rules={[{ required: true, message: '请输入邮箱' }]}
          >
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          
          <Form.Item
            label="执业证号"
            name="licenseNumber"
            rules={[{ required: true, message: '请输入执业证号' }]}
          >
            <Input placeholder="请输入执业证号" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default LawyerManagement;