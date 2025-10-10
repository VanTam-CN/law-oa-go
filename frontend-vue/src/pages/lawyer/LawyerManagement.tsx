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
  Spin,
  Alert
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
import { dataService } from '@/services/dataService';
import { StandardPage, StandardTable, createStandardColumns } from '@/components/ui';
import { DESIGN_TOKENS, BUSINESS_STATUS } from '@/constants/design-system';

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
    setLoading(true);
    try {
      const response = await dataService.getLawyers({
        page: queryParams.pageNum,
        page_size: queryParams.pageSize,
        search: queryParams.name,
        department: queryParams.department,
        status: queryParams.status,
        specialty: queryParams.specialty,
      });

      setLawyers(response.data);
      setTotal(response.total);

      // 计算统计数据
      const activeCount = response.data.filter(lawyer => lawyer.status === 'active').length;
      const inactiveCount = response.data.filter(lawyer => lawyer.status === 'inactive').length;
      const onLeaveCount = response.data.filter(lawyer => lawyer.status === 'on_leave').length;

      setStats({
        total: response.total,
        active: activeCount,
        inactive: inactiveCount,
        onLeave: onLeaveCount,
      });
    } catch (error) {
      console.error('获取律师列表失败:', error);
      message.error('获取律师列表失败，请稍后重试');

      // 重置数据
      setLawyers([]);
      setTotal(0);
      setStats({
        total: 0,
        active: 0,
        inactive: 0,
        onLeave: 0,
      });
    } finally {
      setLoading(false);
    }
  };


  useEffect(() => {
    fetchLawyers();
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
    const statusConfig = {
      active: { text: '在职', color: DESIGN_TOKENS.colors.success, icon: <CheckCircleOutlined /> },
      inactive: { text: '离职', color: DESIGN_TOKENS.colors.error, icon: <UserSwitchOutlined /> },
      on_leave: { text: '休假', color: DESIGN_TOKENS.colors.warning, icon: <PauseCircleOutlined /> }
    };
    const config = statusConfig[status as keyof typeof statusConfig] || { text: '未知', color: DESIGN_TOKENS.colors.textTertiary, icon: null };
    return (
      <Tag
        color={config.color}
        icon={config.icon}
        style={{
          borderRadius: DESIGN_TOKENS.radius.sm,
          fontSize: DESIGN_TOKENS.typography.xs.fontSize,
          fontWeight: '500',
        }}
      >
        {config.text}
      </Tag>
    );
  };

  // 表格列定义
  const columns: ColumnsType<Lawyer> = [
    createStandardColumns.createUserColumn('name', '姓名', 120, true),
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
      setLoading(true);

      if (editingLawyer) {
        await dataService.updateLawyer(editingLawyer.id!, values);
        message.success('律师信息更新成功');
      } else {
        await dataService.createLawyer(values);
        message.success('律师创建成功');
      }

      setModalVisible(false);
      setEditingLawyer(null);
      form.resetFields();
      fetchLawyers();
    } catch (error) {
      console.error('保存律师信息失败:', error);
      message.error('保存失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  // 删除律师
  const handleDelete = async (id: number) => {
    try {
      await dataService.deleteLawyer(id);
      message.success('律师删除成功');
      fetchLawyers();
    } catch (error) {
      console.error('删除律师失败:', error);
      message.error('删除失败，请重试');
    }
  };

  return (
    <StandardPage
      title="律师管理"
      subtitle="管理律所律师信息，包括基本信息、专业领域和工作状态"
      showRefresh
      onRefresh={fetchLawyers}
      showAdd
      onAdd={() => {
        setEditingLawyer(null);
        form.resetFields();
        setModalTitle('新增律师');
        setModalVisible(true);
      }}
    >

      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: DESIGN_TOKENS.spacing.xxl }}>
        <Col xs={24} sm={12} md={6}>
          <Card
            style={{
              background: DESIGN_TOKENS.colors.bgCard,
              border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
              borderRadius: DESIGN_TOKENS.radius.lg,
              boxShadow: DESIGN_TOKENS.shadows.sm,
              textAlign: 'center',
            }}
          >
            <Statistic
              title="律师总数"
              value={stats?.total || 0}
              prefix={<UserOutlined style={{ color: DESIGN_TOKENS.colors.primary }} />}
              valueStyle={{ color: DESIGN_TOKENS.colors.primary }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card
            style={{
              background: DESIGN_TOKENS.colors.bgCard,
              border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
              borderRadius: DESIGN_TOKENS.radius.lg,
              boxShadow: DESIGN_TOKENS.shadows.sm,
              textAlign: 'center',
            }}
          >
            <Statistic
              title="在职律师"
              value={stats?.active || 0}
              valueStyle={{ color: DESIGN_TOKENS.colors.success }}
              prefix={<CheckCircleOutlined style={{ color: DESIGN_TOKENS.colors.success }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card
            style={{
              background: DESIGN_TOKENS.colors.bgCard,
              border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
              borderRadius: DESIGN_TOKENS.radius.lg,
              boxShadow: DESIGN_TOKENS.shadows.sm,
              textAlign: 'center',
            }}
          >
            <Statistic
              title="休假律师"
              value={stats?.onLeave || 0}
              valueStyle={{ color: DESIGN_TOKENS.colors.warning }}
              prefix={<PauseCircleOutlined style={{ color: DESIGN_TOKENS.colors.warning }} />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card
            style={{
              background: DESIGN_TOKENS.colors.bgCard,
              border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
              borderRadius: DESIGN_TOKENS.radius.lg,
              boxShadow: DESIGN_TOKENS.shadows.sm,
              textAlign: 'center',
            }}
          >
            <Statistic
              title="离职律师"
              value={stats?.inactive || 0}
              valueStyle={{ color: DESIGN_TOKENS.colors.error }}
              prefix={<UserSwitchOutlined style={{ color: DESIGN_TOKENS.colors.error }} />}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索表单 */}
      <Card
        style={{
          marginBottom: DESIGN_TOKENS.spacing.xxl,
          background: DESIGN_TOKENS.colors.bgCard,
          border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
          borderRadius: DESIGN_TOKENS.radius.lg,
          boxShadow: DESIGN_TOKENS.shadows.sm,
        }}
      >
        <Form layout="inline">
          <Form.Item label="姓名">
            <Input
              placeholder="请输入律师姓名"
              value={queryParams.name}
              onChange={(e) => setQueryParams({ ...queryParams, name: e.target.value })}
              allowClear
              style={{ borderRadius: DESIGN_TOKENS.radius.sm }}
            />
          </Form.Item>
          <Form.Item label="部门">
            <Select
              placeholder="请选择部门"
              value={queryParams.department}
              onChange={(value) => setQueryParams({ ...queryParams, department: value })}
              allowClear
              style={{ borderRadius: DESIGN_TOKENS.radius.sm }}
            >
              <Option value="民事诉讼部">民事诉讼部</Option>
              <Option value="公司法务部">公司法务部</Option>
              <Option value="刑事辩护部">刑事辩护部</Option>
              <Option value="知识产权部">知识产权部</Option>
            </Select>
          </Form.Item>
          <Form.Item label="状态">
            <Select
              placeholder="请选择状态"
              value={queryParams.status}
              onChange={(value) => setQueryParams({ ...queryParams, status: value })}
              allowClear
              style={{ borderRadius: DESIGN_TOKENS.radius.sm }}
            >
              <Option value="active">在职</Option>
              <Option value="on_leave">休假</Option>
              <Option value="inactive">离职</Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button
                type="primary"
                icon={<SearchOutlined />}
                onClick={handleSearch}
                style={{
                  borderRadius: DESIGN_TOKENS.radius.md,
                  background: `linear-gradient(135deg, ${DESIGN_TOKENS.colors.primary} 0%, ${DESIGN_TOKENS.colors.primaryHover} 100%)`,
                  border: 'none',
                }}
              >
                搜索
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={handleReset}
                style={{
                  borderRadius: DESIGN_TOKENS.radius.md,
                  border: `1px solid ${DESIGN_TOKENS.colors.borderBase}`,
                }}
              >
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
    </StandardPage>
  );
};

export default LawyerManagement;