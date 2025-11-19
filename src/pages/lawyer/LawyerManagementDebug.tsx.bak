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
  Spin,
  Empty
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
  CheckCircleOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { lawyerService, Lawyer, LawyerStats } from '@/services/lawyer';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';

const { Option } = Select;
const { TextArea } = Input;

const LawyerManagementDebug: React.FC = () => {
  console.log('LawyerManagementDebug 组件正在渲染');
  const navigate = useNavigate();
  const [lawyers, setLawyers] = useState<Lawyer[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [stats, setStats] = useState<LawyerStats | null>(null);
  const [initialized, setInitialized] = useState<boolean>(false);
  const [form] = Form.useForm();
  
  console.log('当前状态:', { 
    lawyers: lawyers.length, 
    loading, 
    stats,
    initialized
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

  // 获取律师列表 - 简化版本，直接使用fetch
  const fetchLawyers = async () => {
    console.log('fetchLawyers被调用');
    setLoading(true);
    try {
      console.log('开始获取律师列表，参数:', queryParams);
      // 直接使用fetch，不使用lawyerService
      const response = await fetch(`/api/lawfirm/lawyers?name=${queryParams.name}&department=${queryParams.department}&status=${queryParams.status}&specialty=${queryParams.specialty}&pageNum=${queryParams.pageNum}&pageSize=${queryParams.pageSize}`);
      const data = await response.json();
      
      console.log('API响应数据:', data);
      
      if (data.code === 0 && data.data) {
        const convertedLawyers = data.data.list.map((lawyer: any) => ({
          id: lawyer.lawyerId,
          name: lawyer.lawyerName,
          licenseNumber: lawyer.licenseNo,
          specialty: lawyer.specialty ? lawyer.specialty.split(/[,、]/).filter(Boolean) : [],
          experience: 5,
          status: lawyer.delFlag === '0' ? 'active' : 'inactive',
          department: lawyer.department || '',
          position: lawyer.position || '',
          phone: lawyer.phone,
          email: lawyer.email,
          gender: 'male',
          joinDate: '2020-01-01',
          profile: '',
          avatar: '',
          ...lawyer
        }));
        console.log('转换后的律师数据:', convertedLawyers);
        setLawyers(convertedLawyers);
        setTotal(data.data.total || 0);
        setInitialized(true);
      } else {
        throw new Error('API返回数据格式错误');
      }
    } catch (error: any) {
      console.error('获取律师列表失败:', error);
      console.error('错误详情:', {
        message: error.message,
        response: error.response,
        stack: error.stack
      });
      const errorMessage = error.message || '获取律师列表失败';
      message.error(errorMessage);
      setInitialized(true);
    } finally {
      setLoading(false);
    }
  };

  // 获取律师统计 - 简化版本
  const fetchStats = async () => {
    try {
      const response = await fetch('/api/lawfirm/lawyers/stats');
      const data = await response.json();
      
      if (data.code === 0 && data.data) {
        console.log('统计数据:', data.data);
        setStats(data.data);
      }
    } catch (error: any) {
      console.error('获取统计数据失败:', error);
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
  ];

  // 如果正在加载且未初始化，显示加载状态
  if (loading && !initialized) {
    return (
      <div style={{ padding: '24px', textAlign: 'center' }}>
        <Spin size="large" tip="正在加载律师数据..." />
      </div>
    );
  }

  // 如果已初始化但没有数据，显示空状态
  if (initialized && lawyers.length === 0 && !loading) {
    return (
      <div style={{ padding: '24px' }}>
        <Card>
          <Empty 
            description="暂无律师数据"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          >
            <Button type="primary" icon={<PlusOutlined />}>
              新增律师
            </Button>
          </Empty>
        </Card>
      </div>
    );
  }
  
  return (
    <div style={{ padding: '24px' }}>
      <Card title="律师管理 (调试版本)" style={{ marginBottom: '16px' }}>
        <p>调试信息：律师数量 {lawyers.length}, 总数 {total}, 加载状态 {loading ? '是' : '否'}</p>
      </Card>

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
    </div>
  );
};

export default LawyerManagementDebug;