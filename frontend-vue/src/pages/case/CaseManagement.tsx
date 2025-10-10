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
  message,
  Popconfirm,
  Tooltip,
  Row,
  Col,
  Statistic
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  EyeOutlined,
  SearchOutlined,
  FileTextOutlined,
  UserOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import CreateCaseWizard from '@/components/CreateCaseWizard';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import { getCaseList, createCase, updateCase, deleteCase } from '@/api/case';
import { get } from '@/services/http';
import { setTestToken } from '@/utils/auth';
import './CaseManagement.less';

const { Option } = Select;
const { TextArea } = Input;

interface Case {
  caseId: number;
  caseNo: string;
  caseName: string;
  caseType: string;
  clientName: string;
  lawyerName: string;
  status: string;
  description: string;
  createTime: string;
  updateTime: string;
}

interface CaseFormData {
  caseNo: string;
  caseName: string;
  caseType: string;
  clientId: number | null;
  lawyerId: number | null;
  status: string;
  description: string;
}

const CaseManagement: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [cases, setCases] = useState<Case[]>([]);
  const [visible, setVisible] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editingCase, setEditingCase] = useState<Case | null>(null);
  const [form] = Form.useForm();
  
  // 搜索状态
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [lawyerFilter, setLawyerFilter] = useState<string>('');
  const [clientFilter, setClientFilter] = useState<string>('');
  
  // 下拉选项
  const [lawyerOptions, setLawyerOptions] = useState<{ label: string; value: string | number }[]>([]);
  const [clientOptions, setClientOptions] = useState<{ label: string; value: string | number }[]>([]);
  
  // 分页状态
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  useEffect(() => {
    fetchCases();
    loadOptions();
  }, []);

  useEffect(() => {
    fetchCases();
  }, [pagination.current, pagination.pageSize, searchText, statusFilter, typeFilter, lawyerFilter, clientFilter]);

  const loadOptions = async () => {
    try {
      const [clientRes, lawyerRes] = await Promise.all([
        get<any>('/clients', { pageNum: 1, pageSize: 9999 }).catch(() => ({ data: { list: [] } })),
        get<any>('/lawfirm/lawyers', { pageNum: 1, pageSize: 9999 }).catch(() => ({ data: { list: [] } })),
      ]);
      
      const cOpts = (clientRes?.data?.list ?? []).map((c: any) => ({
        label: c.name ?? c.clientName ?? c.company ?? '',
        value: c.id ?? c.clientId,
      })).filter((o: any) => o.label && o.value !== undefined);
      
      const lOpts = (lawyerRes?.data?.list ?? []).map((l: any) => ({
        label: l.name ?? l.lawyerName ?? '',
        value: l.id ?? l.lawyerId,
      })).filter((o: any) => o.label && o.value !== undefined);
      
      setClientOptions(cOpts);
      setLawyerOptions(lOpts);
    } catch (error) {
      console.error('加载选项失败:', error);
    }
  };

  const fetchCases = async () => {
    setLoading(true);
    try {
      const params: any = {
        page: pagination.current,
        page_size: pagination.pageSize,
      };
      
      if (searchText) params.search = searchText;
      if (statusFilter) params.status = statusFilter;
      if (typeFilter) params.case_type = typeFilter;
      if (lawyerFilter) {
        const lawyerOption = lawyerOptions.find(opt => opt.label === lawyerFilter);
        if (lawyerOption) params.lawyer_id = lawyerOption.value;
      }
      if (clientFilter) {
        const clientOption = clientOptions.find(opt => opt.label === clientFilter);
        if (clientOption) params.client_id = clientOption.value;
      }
      
      const [caseRes, clientRes, lawyerRes] = await Promise.all([
        getCaseList(params),
        get<any>('/clients', { pageNum: 1, pageSize: 9999 }).catch(() => ({ data: { list: [] } })),
        get<any>('/lawfirm/lawyers', { pageNum: 1, pageSize: 9999 }).catch(() => ({ data: { list: [] } })),
      ]);
      
      // 构建映射
      const clientMap = new Map<string | number, string>();
      const clientList = clientRes?.data?.list || clientRes?.data || [];
      for (const c of clientList) {
        const id = c.id ?? c.clientId;
        const name = c.name ?? c.clientName ?? c.company ?? '';
        if (id != null && name) clientMap.set(id, name);
      }
      
      const lawyerMap = new Map<string | number, string>();
      const lawyerList = lawyerRes?.data?.list || lawyerRes?.data || [];
      for (const l of lawyerList) {
        const id = l.id ?? l.lawyerId;
        const name = l.name ?? l.lawyerName ?? '';
        if (id != null && name) lawyerMap.set(id, name);
      }
      
      // 处理案件数据 - 修复数据解析
      console.log('API响应:', caseRes);
      const caseList = caseRes || [];
      const mappedRows = caseList.map((item: any) => {
        const clientId = item.client_id ?? item.clientId;
        const lawyerId = item.lawyer_id ?? item.lawyerId;
        
        return {
          caseId: item.id ?? 0,
          caseNo: item.case_number ?? item.caseNo ?? `CASE-${item.id}`,
          caseName: item.title ?? item.caseName ?? '未命名案件',
          caseType: item.case_type ?? item.caseType ?? '',
          clientName: item.client_name ?? clientMap.get(clientId) ?? '未知客户',
          lawyerName: item.lawyer_name ?? lawyerMap.get(lawyerId) ?? '未分配',
          status: item.status ?? 'pending',
          description: item.description ?? '',
          createTime: item.created_at ?? item.createTime ?? '',
          updateTime: item.updated_at ?? item.updateTime ?? '',
        } as Case;
      });
      
      console.log('处理后的案件数据:', mappedRows);
      
      setCases(mappedRows);
      setPagination({
        ...pagination,
        total: caseRes.pagination?.total || caseRes.total || mappedRows.length
      });
    } catch (error) {
      console.error('获取案件列表失败:', error);
      
      // 检查是否是认证错误
      if (error instanceof Error && error.message.includes('401')) {
        message.error('请先设置测试Token');
      } else {
        message.error('获取案件列表失败: ' + (error instanceof Error ? error.message : '未知错误'));
      }
      
      setCases([]);
      setPagination({ ...pagination, total: 0 });
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setCreateModalVisible(true);
  };

  const handleEdit = (record: Case) => {
    setEditingCase(record);
    setVisible(true);
    
    setTimeout(() => {
      const clientOpt = clientOptions.find(opt => opt.label === record.clientName);
      const lawyerOpt = lawyerOptions.find(opt => opt.label === record.lawyerName);
      
      form.setFieldsValue({
        caseNo: record.caseNo,
        caseName: record.caseName,
        caseType: record.caseType,
        clientId: clientOpt?.value || null,
        lawyerId: lawyerOpt?.value || null,
        status: record.status,
        description: record.description,
      });
    }, 0);
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteCase(id);
      message.success('删除成功');
      fetchCases();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleSubmit = async (values: CaseFormData) => {
    try {
      setLoading(true);
      
      const submitData = {
        title: values.caseName,
        description: values.description || '',
        client_id: values.clientId,
        lawyer_id: values.lawyerId,
        case_type: values.caseType,
        priority: 'medium',
        status: values.status || 'pending'
      };
      
      if (editingCase) {
        await updateCase(editingCase.caseId, submitData);
        message.success('案件更新成功');
      } else {
        await createCase(submitData);
        message.success('案件创建成功');
      }
      
      setVisible(false);
      setEditingCase(null);
      form.resetFields();
      fetchCases();
    } catch (error) {
      console.error('保存案件失败:', error);
      message.error('保存失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const statusMap: Record<string, { text: string; color: string }> = {
      'pending': { text: '待处理', color: 'orange' },
      'active': { text: '进行中', color: 'green' },
      'closed': { text: '已结案', color: 'blue' },
      'suspended': { text: '已暂停', color: 'red' }
    };
    const config = statusMap[status] || { text: '未知', color: 'default' };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const getCaseTypeTag = (type: string) => {
    const typeMap: Record<string, { text: string; color: string }> = {
      'civil': { text: '民事案件', color: 'blue' },
      'commercial': { text: '商事案件', color: 'orange' },
      'criminal': { text: '刑事案件', color: 'red' },
      'administrative': { text: '行政案件', color: 'purple' }
    };
    const config = typeMap[type] || { text: type || '其他', color: 'default' };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const columns: ColumnsType<Case> = [
    {
      title: '案件编号',
      dataIndex: 'caseNo',
      key: 'caseNo',
      width: 150,
      render: (text: string) => (
        <Space>
          <FileTextOutlined />
          {text || '未设置'}
        </Space>
      )
    },
    {
      title: '案件名称',
      dataIndex: 'caseName',
      key: 'caseName',
      ellipsis: true,
      render: (text: string) => text || '未命名案件'
    },
    {
      title: '案件类型',
      dataIndex: 'caseType',
      key: 'caseType',
      width: 120,
      render: (text: string) => getCaseTypeTag(text)
    },
    {
      title: '客户',
      dataIndex: 'clientName',
      key: 'clientName',
      width: 150,
      render: (text: string) => (
        <Space>
          <UserOutlined />
          {text || '未知客户'}
        </Space>
      )
    },
    {
      title: '负责律师',
      dataIndex: 'lawyerName',
      key: 'lawyerName',
      width: 120,
      render: (text: string) => text || '未分配'
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (text: string) => getStatusBadge(text)
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      key: 'createTime',
      width: 180,
      render: (text: string) => text ? dayjs(text).format('YYYY-MM-DD HH:mm') : '-'
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space size="middle">
          <Tooltip title="查看详情">
            <Button 
              type="link" 
              icon={<EyeOutlined />}
              onClick={() => navigate(`/case/${record.caseId}`)}
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
              title="确定要删除这个案件吗？"
              onConfirm={() => handleDelete(record.caseId)}
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
    <div className="case-management">
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} className="stats-row">
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="案件总数" 
              value={cases.length} 
              prefix={<FileTextOutlined />} 
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="进行中" 
              value={cases.filter(c => c.status === 'active').length} 
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="已结案" 
              value={cases.filter(c => c.status === 'closed').length} 
              prefix={<CheckCircleOutlined />} 
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic 
              title="本月新增" 
              value={cases.filter(c => dayjs(c.createTime).isAfter(dayjs().startOf('month'))).length} 
              prefix={<ClockCircleOutlined />} 
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索表单 */}
      <Card className="search-card">
        <Form layout="inline">
          <Form.Item label="案件搜索">
            <Input 
              placeholder="搜索案件名称、编号或客户" 
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              allowClear
              style={{ width: 250 }}
            />
          </Form.Item>
          <Form.Item label="案件状态">
            <Select 
              style={{ width: 120 }}
              value={statusFilter}
              onChange={setStatusFilter}
              allowClear
              placeholder="全部"
            >
              <Option value="pending">待处理</Option>
              <Option value="active">进行中</Option>
              <Option value="closed">已结案</Option>
              <Option value="suspended">已暂停</Option>
            </Select>
          </Form.Item>
          <Form.Item label="案件类型">
            <Select 
              style={{ width: 120 }}
              value={typeFilter}
              onChange={setTypeFilter}
              allowClear
              placeholder="全部"
            >
              <Option value="civil">民事案件</Option>
              <Option value="commercial">商事案件</Option>
              <Option value="criminal">刑事案件</Option>
              <Option value="administrative">行政案件</Option>
            </Select>
          </Form.Item>
          <Form.Item label="负责律师">
            <Select 
              style={{ width: 120 }}
              value={lawyerFilter}
              onChange={setLawyerFilter}
              allowClear
              placeholder="全部"
              showSearch
              optionFilterProp="label"
              options={lawyerOptions}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={() => fetchCases()}>
                搜索
              </Button>
              <Button icon={<ReloadOutlined />} onClick={() => {
                setSearchText('');
                setStatusFilter('');
                setTypeFilter('');
                setLawyerFilter('');
                setClientFilter('');
                fetchCases();
              }}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {/* 案件列表 */}
      <Card 
        title="案件列表" 
        extra={
          <Space>
            <Button onClick={() => {
              setTestToken();
              message.success('测试token已设置，请刷新页面');
            }}>
              设置测试Token
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              新建案件
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={cases}
          rowKey="caseId"
          loading={loading}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
            onChange: (page, size) => {
              setPagination({
                ...pagination,
                current: page,
                pageSize: size || 10
              });
            }
          }}
        />
      </Card>

      {/* 新增案件向导 */}
      <CreateCaseWizard
        visible={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onSuccess={() => {
          setCreateModalVisible(false);
          fetchCases();
        }}
      />

      {/* 编辑案件弹窗 */}
      <Modal
        title={editingCase ? '编辑案件' : '新建案件'}
        open={visible}
        onOk={() => form.submit()}
        onCancel={() => {
          setVisible(false);
          setEditingCase(null);
          form.resetFields();
        }}
        width={600}
        destroyOnClose
        confirmLoading={loading}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item
            label="案件名称"
            name="caseName"
            rules={[{ required: true, message: '请输入案件名称' }]}
          >
            <Input placeholder="请输入案件名称" />
          </Form.Item>

          <Form.Item
            label="案件编号"
            name="caseNo"
            rules={[{ required: true, message: '请输入案件编号' }]}
          >
            <Input placeholder="请输入案件编号" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="案件类型"
                name="caseType"
                rules={[{ required: true, message: '请选择案件类型' }]}
              >
                <Select placeholder="请选择案件类型">
                  <Option value="civil">民事案件</Option>
                  <Option value="commercial">商事案件</Option>
                  <Option value="criminal">刑事案件</Option>
                  <Option value="administrative">行政案件</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="案件状态"
                name="status"
                rules={[{ required: true, message: '请选择案件状态' }]}
              >
                <Select placeholder="请选择案件状态">
                  <Option value="pending">待处理</Option>
                  <Option value="active">进行中</Option>
                  <Option value="closed">已结案</Option>
                  <Option value="suspended">已暂停</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="委托客户"
                name="clientId"
                rules={[{ required: true, message: '请选择委托客户' }]}
              >
                <Select 
                  placeholder="请选择委托客户"
                  showSearch
                  optionFilterProp="label"
                  options={clientOptions}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="负责律师"
                name="lawyerId"
                rules={[{ required: true, message: '请选择负责律师' }]}
              >
                <Select 
                  placeholder="请选择负责律师"
                  showSearch
                  optionFilterProp="label"
                  options={lawyerOptions}
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="案件描述"
            name="description"
          >
            <TextArea 
              rows={4} 
              placeholder="请输入案件描述" 
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default CaseManagement;