import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Select,
  DatePicker,
  InputNumber,
  Upload,
  Button,
  Steps,
  Row,
  Col,
  Checkbox,
  Radio,
  message,
  Space,
  Typography,
  Divider,
  Alert,
  Modal,
  List,
  Tag,
  Progress,
  AutoComplete
} from 'antd';
import {
  PlusOutlined,
  UploadOutlined,
  FileTextOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
  DollarOutlined,
  UserOutlined,
  SearchOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import { CaseCreationService, CaseValidationService } from '@/services/caseCreation';
import { ConflictCheckService, ConflictCheckResultProcessor } from '@/services/conflictCheck';
import { get } from '@/services/api';
import './CreateCase.module.css';

const { Option } = Select;
const { TextArea } = Input;
const { Text, Title } = Typography;
const { Step } = Steps;

interface CreateCaseProps {
  visible?: boolean;
  onCancel?: () => void;
  onSuccess?: () => void;
}

// 案件数据接口
interface CaseFormData {
  // 基本信息
  caseName: string;
  clientName: string;
  clientId: string;
  otherParties: string[];
  caseType: string;
  causeOfAction: string;
  caseDescription: string;
  
  // 内部管理
  leadLawyer: string;
  assistingLawyers: string[];
  billingMethod: string;
  contractAmount: number;
  estimatedDuration: number;
  
  // 合规风控
  conflictCheck: string;
  riskTags: string[];
  isHighRisk: boolean;
  approvalRequired: boolean;
  
  // 文档管理
  contractDocument: any[];
  retainerAgreement: any[];
  otherDocuments: any[];
}

const CreateCase: React.FC<CreateCaseProps> = ({
  visible = false,
  onCancel,
  onSuccess
}) => {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [conflictCheckResult, setConflictCheckResult] = useState<any>(null);
  const [conflictCheckProgress, setConflictCheckProgress] = useState<{
    step: number;
    stepName: string;
    progress: number;
    details: string[];
    isCompleted: boolean;
  } | null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);

  // 动态数据
  const [clients, setClients] = useState<any[]>([]);
  const [lawyers, setLawyers] = useState<any[]>([]);
  const [dataLoading, setDataLoading] = useState(false);

  // 获取客户和律师数据
  const fetchDropdownData = async () => {
    console.log('开始获取下拉框数据...');
    setDataLoading(true);
    try {
      const [clientsResponse, lawyersResponse] = await Promise.all([
        get('/clients', { pageNum: 1, pageSize: 9999 }),
        get('/lawfirm/lawyers', { pageNum: 1, pageSize: 9999 })
      ]);

      console.log('API响应:', { clientsResponse, lawyersResponse });

      if (clientsResponse?.data) {
        const formattedClients = clientsResponse.data.map((client: any) => ({
          id: client.id.toString(),
          name: client.name || client.company || `客户${client.id}`,
          type: client.company ? 'COMPANY' : 'PERSON',
          phone: client.phone || '未提供',
          email: client.email || '',
          company: client.company || ''
        }));
        console.log('格式化客户数据:', formattedClients);
        setClients(formattedClients);
      }

      if (lawyersResponse?.data) {
        const formattedLawyers = lawyersResponse.data.map((lawyer: any) => ({
          id: lawyer.id.toString(),
          name: lawyer.name,
          level: 'SENIOR', // 默认级别，可以根据需要调整
          specialties: ['法律咨询'], // 默认专业，可以根据需要调整
          email: lawyer.email || '',
          phone: lawyer.phone || ''
        }));
        console.log('格式化律师数据:', formattedLawyers);
        setLawyers(formattedLawyers);
      }
    } catch (error) {
      console.error('获取下拉框数据失败:', error);
      message.error('获取数据失败，请稍后重试');
    } finally {
      setDataLoading(false);
      console.log('数据加载完成，状态:', { clients: clients.length, lawyers: lawyers.length });
    }
  };

  // 组件挂载时获取数据
  useEffect(() => {
    console.log('CreateCase组件useEffect触发，visible:', visible);
    if (visible) {
      fetchDropdownData();
    }
  }, [visible]);

  const caseTypes = [
    { 
      value: 'CIVIL', 
      label: '民事案件', 
      causes: [
        '合同纠纷', '侵权责任', '婚姻家庭', '继承纠纷', 
        '物权纠纷', '人格权纠纷', '劳动争议', '负欠纠纷',
        '不当得利纠纷', '无因管理纠纷', '医疗纠纷', '交通事故赔偿纠纷'
      ] 
    },
    { 
      value: 'COMMERCIAL', 
      label: '商事案件', 
      causes: [
        '公司纠纷', '金融纠纷', '知识产权', '投资纠纷',
        '证券纠纷', '保险纠纷', '票据纠纷', '信托纠纷',
        '担保物权纠纷', '不正当竞争纠纷', '破产纠纷', '租赁合同纠纷'
      ] 
    },
    { 
      value: 'CRIMINAL', 
      label: '刑事案件', 
      causes: [
        '经济犯罪', '职务犯罪', '暴力犯罪', '网络犯罪',
        '毒品犯罪', '财产犯罪', '政务犯罪', '破坏社会主义市场经济秩序罪',
        '侵犯公民人身权利、民主权利罪', '危害公共安全罪'
      ] 
    },
    { 
      value: 'ADMINISTRATIVE', 
      label: '行政案件', 
      causes: [
        '行政处罚', '行政许可', '信息公开', '征收补偿',
        '政府采购', '土地使用权', '环保执法', '税务争议',
        '城市管理', '教育行政', '卫生行政', '公安行政'
      ] 
    }
  ];

  const riskCategories = [
    { value: 'HIGH_VALUE', label: '重大标的项目', description: '标的额超过1000万元' },
    { value: 'GROUP_CASE', label: '群体性项目', description: '涉及多方当事人' },
    { value: 'SENSITIVE', label: '敏感性项目', description: '涉及政府机关或公共利益' },
    { value: 'MEDIA_ATTENTION', label: '媒体关注项目', description: '可能引起媒体关注' },
    { value: 'CROSS_BORDER', label: '涉外项目', description: '涉及境外当事人或法律' }
  ];

  const steps = [
    { title: '基本信息', icon: <FileTextOutlined /> },
    { title: '内部管理', icon: <TeamOutlined /> },
    { title: '合规风控', icon: <SafetyCertificateOutlined /> },
    { title: '文档上传', icon: <UploadOutlined /> }
  ];

  // 利益冲突检查
  const performConflictCheck = async (clientId: string, otherParties: string[]) => {
    setLoading(true);
    setConflictCheckResult(null);
    setConflictCheckProgress(null);
    
    try {
      const selectedClient = clients.find(c => c.id === clientId);
      
      if (!selectedClient) {
        throw new Error('无法找到指定的委托人信息');
      }
      
      // 检查必要的参数
      const caseName = form.getFieldValue('caseName');
      const caseType = form.getFieldValue('caseType');
      
      if (!caseName || !caseType) {
        throw new Error('请先填写完整的案件名称和类型');
      }
      
      // 构建请求参数
      const request = {
        clientId: clientId,
        clientName: selectedClient?.name || '',
        clientType: (selectedClient?.type || 'PERSON') as 'PERSON' | 'COMPANY',
        otherParties: otherParties || [],
        caseName: caseName,
        caseType: caseType,
        searchYears: 5,
        includeCorporateRelations: selectedClient?.type === 'COMPANY',
        searchDepth: 'STANDARD' as const
      };
      
      // 模拟分步显示检索进度
      const checkSteps = [
        {
          step: 1,
          stepName: '检索委托人信息',
          progress: 10,
          details: [
            `正在检索委托人: ${selectedClient?.name || '未知'}`,
            `委托人类型: ${selectedClient?.type === 'COMPANY' ? '企业客户' : '个人客户'}`,
            `联系方式: ${selectedClient?.phone || '未提供'}`
          ]
        },
        {
          step: 2,
          stepName: '检索历史案件',
          progress: 30,
          details: [
            '正在检索该委托人的历史案件...',
            '查询时间范围: 2020年1月 - 至今',
            '检索字段: 委托人姓名、证件号码、企业统一社会信用代码'
          ]
        },
        {
          step: 3,
          stepName: '检索对方当事人',
          progress: 50,
          details: [
            `正在检索对方当事人: ${otherParties.length > 0 ? otherParties.join('、') : '未填写'}`,
            '检索范围: 所有历史案件的当事人信息',
            '匹配方式: 姓名完全匹配 + 模糊匹配'
          ]
        },
        {
          step: 4,
          stepName: '检索关联企业',
          progress: 70,
          details: [
            '正在检索委托人关联企业...',
            '检索股东关系、董事关系、投资关系',
            '检索时间: 过去5年内的工商变更记录'
          ]
        },
        {
          step: 5,
          stepName: '分析潜在冲突',
          progress: 90,
          details: [
            '正在分析潜在利益冲突...',
            '检查同一案件中的对立当事人',
            '检查关联案件中的利益关系',
            '应用冲突检索规则引擎'
          ]
        },
        {
          step: 6,
          stepName: '生成检索报告',
          progress: 100,
          details: [
            '正在生成详细检索报告...',
            '汇总检索结果',
            '标记风险等级'
          ]
        }
      ];

      // 模拟逐步检索过程显示
      for (let i = 0; i < checkSteps.length - 1; i++) {
        const step = checkSteps[i];
        setConflictCheckProgress({
          ...step,
          isCompleted: false
        });
        
        // 模拟每个步骤的处理时间
        await new Promise(resolve => setTimeout(resolve, 600 + Math.random() * 800));
      }
      
      // 调用后端服务执行冲突检索
      try {
        const backendResult = await ConflictCheckService.performConflictCheck(request);
        
        // 验证返回数据
        if (!backendResult) {
          throw new Error('后端服务返回空数据');
        }
        
        // 处理后端返回的数据
        const result = ConflictCheckResultProcessor.processResult(backendResult);
        
        setConflictCheckResult(result);
        setConflictCheckProgress({
          step: 6,
          stepName: '检索完成',
          progress: 100,
          details: [
            '冲突检索已完成', 
            `检索耗时: ${((backendResult.duration || 0) / 1000).toFixed(1)}秒`,
            `检索ID: ${backendResult.checkId || 'N/A'}`,
            ConflictCheckResultProcessor.generateSummary(result.checkDetails),
            backendResult.checkId?.startsWith('MOCK_') ? 'ℹ️ 开发模式：使用模拟数据' : '✅ 生产模式：真实后端数据'
          ],
          isCompleted: true
        });
        
        // 自动设置冲突检索结果
        form.setFieldsValue({
          conflictCheck: result.hasConflict ? 'CONFLICT_RESOLVED' : 'NO_CONFLICT'
        });
      } catch (serviceError) {
        // 重新抛出服务层错误，让外层catch处理
        throw serviceError;
      }
      
    } catch (error) {
      console.error('冲突检查失败:', error);
      
      let errorMessage = '冲突检查失败';
      let errorDetails: string[] = [];
      
      if (error instanceof Error) {
        errorMessage = error.message;
        
        // 根据不同错误类型提供不同的错误详情
        if (error.message.includes('Failed to fetch') || error.message.includes('网络连接失败')) {
          errorDetails = [
            '网络连接失败，请检查：',
            '1. 网络连接是否正常',
            '2. 后端服务是否启动 (http://localhost:8082)',
            '3. 代理配置是否正确',
            '如问题持续，请联系技术支持'
          ];
        } else if (error.message.includes('超时')) {
          errorDetails = [
            '请求超时，请检查：',
            '1. 网络连接速度',
            '2. 后端服务响应速度',
            '3. 尝试重新执行检索'
          ];
        } else if (error.message.includes('参数')) {
          errorDetails = [
            '参数验证失败，请检查：',
            '1. 委托人信息是否完整',
            '2. 案件名称和类型是否填写',
            '3. 表单数据是否有效'
          ];
        } else {
          errorDetails = [
            errorMessage,
            '请检查网络连接和后端服务状态',
            '如问题持续，请联系技术支持'
          ];
        }
      } else {
        errorDetails = [
          '未知错误类型',
          '请联系技术支持并提供错误信息'
        ];
      }
      
      message.error(errorMessage);
      setConflictCheckProgress({
        step: 0,
        stepName: '检索失败',
        progress: 0,
        details: errorDetails,
        isCompleted: true
      });
    } finally {
      setLoading(false);
    }
  };

  // 表单提交
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      // 使用数据验证服务
      const validation = CaseValidationService.validateAll(values);
      if (!validation.isValid) {
        message.error(`数据验证失败: ${validation.errors.join(', ')}`);
        return;
      }

      // 使用案件创建服务
      const result = await CaseCreationService.createCase(values);
      
      if (result.success) {
        message.success(result.message);
        onSuccess?.();
        form.resetFields();
        setCurrentStep(0);
      } else {
        message.error(result.message);
      }
    } catch (error) {
      console.error('创建失败:', error);
      message.error('案件创建失败');
    } finally {
      setLoading(false);
    }
  };

  // 文件上传处理
  const handleUpload = (info: any, fieldName: string) => {
    if (info.file.status === 'uploading') {
      setUploadProgress(info.file.percent || 0);
    }
    if (info.file.status === 'done') {
      message.success(`${info.file.name} 上传成功`);
      setUploadProgress(0);
    } else if (info.file.status === 'error') {
      message.error(`${info.file.name} 上传失败`);
      setUploadProgress(0);
    }
  };

  // 步骤验证
  const validateStep = async (step: number) => {
    const fieldsToValidate: { [key: number]: string[] } = {
      0: ['caseName', 'clientId', 'caseType', 'causeOfAction', 'caseDescription'],
      1: ['leadLawyer', 'billingMethod', 'contractAmount'],
      2: ['conflictCheck'],
      3: [] // 文档上传为可选，不强制验证
    };

    try {
      if (fieldsToValidate[step].length > 0) {
        await form.validateFields(fieldsToValidate[step]);
      }
      
      // 特殊验证逻辑
      if (step === 2) {
        // 验证是否已完成冲突检索
        if (!conflictCheckResult) {
          message.warning('请先执行冲突检索');
          return false;
        }
        
        // 验证冲突检索选择
        const conflictCheckValue = form.getFieldValue('conflictCheck');
        if (!conflictCheckValue) {
          message.error('请选择冲突检索结果');
          return false;
        }
      }
      
      return true;
    } catch (error) {
      console.error('步骤验证失败:', error);
      message.error('请确保必填字段已正确填写');
      return false;
    }
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <Card title="案件基本信息" className="step-card">
            <Row gutter={16}>
              <Col span={24}>
                <Form.Item
                  label="案件名称"
                  name="caseName"
                  rules={[
                    { required: true, message: '请输入案件名称' },
                    { max: 100, message: '案件名称不能超过100个字符' }
                  ]}
                  tooltip="格式建议：委托人 Vs 对方当事人"
                >
                  <Input 
                    placeholder="例如：武大郎 Vs 潘金莲"
                    prefix={<FileTextOutlined />}
                  />
                </Form.Item>
              </Col>
              
              <Col span={12}>
                <Form.Item
                  label="委托人"
                  name="clientId"
                  rules={[{ required: true, message: '请选择委托人' }]}
                >
                  <Select
                    placeholder={dataLoading ? "正在加载..." : "选择委托人"}
                    showSearch
                    loading={dataLoading}
                    filterOption={(input, option) =>
                      (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
                    }
                    notFoundContent={dataLoading ? "正在加载..." : "暂无数据"}
                  >
                    {clients.map(client => (
                      <Option key={client.id} value={client.id}>
                        <Space>
                          <UserOutlined />
                          {client.name}
                          {client.company && (
                            <Tag color="blue">{client.company}</Tag>
                          )}
                          <Tag color={client.type === 'COMPANY' ? 'blue' : 'green'}>
                            {client.type === 'COMPANY' ? '企业' : '个人'}
                          </Tag>
                        </Space>
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              </Col>

              <Col span={12}>
                <Form.Item
                  label="其他当事人"
                  name="otherParties"
                  tooltip="可添加多个对方当事人"
                >
                  <Select
                    mode="tags"
                    placeholder="输入对方当事人姓名"
                    style={{ width: '100%' }}
                  />
                </Form.Item>
              </Col>

              <Col span={12}>
                <Form.Item
                  label="案件类型"
                  name="caseType"
                  rules={[{ required: true, message: '请选择案件类型' }]}
                >
                  <Select
                    placeholder="选择案件类型"
                    onChange={(value) => {
                      form.setFieldsValue({ causeOfAction: undefined });
                    }}
                  >
                    {caseTypes.map(type => (
                      <Option key={type.value} value={type.value}>
                        {type.label}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              </Col>

              <Col span={12}>
                <Form.Item
                  label="案由"
                  name="causeOfAction"
                  rules={[
                    { required: true, message: '请选择或输入案由' },
                    { max: 100, message: '案由不能超过100个字符' }
                  ]}
                  tooltip="可从下拉列表选择常见案由，或直接输入自定义案由"
                  extra={
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      提示：可选择预设案由或直接输入自定义内容
                    </Text>
                  }
                >
                  <AutoComplete
                    placeholder="选择或输入具体案由（如：合同纠纷）"
                    allowClear
                    options={
                      form.getFieldValue('caseType') ? 
                        caseTypes
                          .find(type => type.value === form.getFieldValue('caseType'))
                          ?.causes.map(cause => ({ value: cause, label: cause })) || []
                        : []
                    }
                    filterOption={(inputValue, option) =>
                      option?.value.toUpperCase().indexOf(inputValue.toUpperCase()) !== -1
                    }
                    notFoundContent={<div style={{ padding: '8px', color: '#999' }}>没有匹配项，可直接输入自定义案由</div>}
                  />
                </Form.Item>
              </Col>

              <Col span={24}>
                <Form.Item
                  label="案件描述"
                  name="caseDescription"
                  rules={[
                    { required: true, message: '请输入案件描述' },
                    { max: 500, message: '描述不能超过500个字符' }
                  ]}
                >
                  <TextArea 
                    rows={4} 
                    placeholder="请详细描述案件背景、争议焦点等关键信息..."
                    showCount
                    maxLength={500}
                  />
                </Form.Item>
              </Col>
            </Row>
          </Card>
        );

      case 1:
        return (
          <Card title="内部管理信息" className="step-card">
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  label="主办律师"
                  name="leadLawyer"
                  rules={[{ required: true, message: '请选择主办律师' }]}
                >
                  <Select
                    placeholder={dataLoading ? "正在加载..." : "选择主办律师"}
                    loading={dataLoading}
                    notFoundContent={dataLoading ? "正在加载..." : "暂无数据"}
                  >
                    {lawyers.map(lawyer => (
                      <Option key={lawyer.id} value={lawyer.id}>
                        <Space>
                          <UserOutlined />
                          {lawyer.name}
                          <Tag color="orange">律师</Tag>
                        </Space>
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              </Col>

              <Col span={12}>
                <Form.Item
                  label="协办律师"
                  name="assistingLawyers"
                  tooltip="可选择多名协办律师"
                >
                  <Select
                    mode="multiple"
                    placeholder={dataLoading ? "正在加载..." : "选择协办律师（可选）"}
                    loading={dataLoading}
                    notFoundContent={dataLoading ? "正在加载..." : "暂无数据"}
                  >
                    {lawyers.map(lawyer => (
                      <Option key={lawyer.id} value={lawyer.id}>
                        {lawyer.name}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              </Col>

              <Col span={12}>
                <Form.Item
                  label="收费方式"
                  name="billingMethod"
                  rules={[{ required: true, message: '请选择收费方式' }]}
                >
                  <Select placeholder="选择收费方式">
                    <Option value="FIXED">定额收费</Option>
                    <Option value="HOURLY">按时收费</Option>
                    <Option value="CONTINGENCY">风险代理</Option>
                    <Option value="MIXED">混合收费</Option>
                  </Select>
                </Form.Item>
              </Col>

              <Col span={12}>
                <Form.Item
                  label="合同金额"
                  name="contractAmount"
                  rules={[
                    { required: true, message: '请输入合同金额' },
                    { type: 'number', min: 0, message: '金额必须大于0' }
                  ]}
                >
                  <InputNumber
                    style={{ width: '100%' }}
                    placeholder="请输入金额"
                    prefix="¥"
                    formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                    parser={(value) => value!.replace(/\$\s?|(,*)/g, '')}
                  />
                </Form.Item>
              </Col>

              <Col span={24}>
                <Form.Item
                  label="预估工期"
                  name="estimatedDuration"
                  tooltip="预估案件处理时长（月）"
                >
                  <InputNumber
                    min={1}
                    max={60}
                    placeholder="预估工期"
                    addonAfter="个月"
                  />
                </Form.Item>
              </Col>
            </Row>
          </Card>
        );

      case 2:
        return (
          <Card title="合规与风险控制" className="step-card">
            <Alert
              message="风险控制提醒"
              description="请认真进行利益冲突检查和风险评估，这是律师执业的基本要求。"
              type="warning"
              showIcon
              style={{ marginBottom: 16 }}
            />

            <Row gutter={16}>
              <Col span={24}>
                <Form.Item
                  label="利益冲突检索"
                  name="conflictCheck"
                  rules={[{ required: true, message: '请完成利益冲突检查' }]}
                >
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Button
                      type="primary"
                      icon={<SearchOutlined />}
                      loading={loading}
                      onClick={() => {
                        const clientId = form.getFieldValue('clientId');
                        const otherParties = form.getFieldValue('otherParties') || [];
                        if (clientId) {
                          performConflictCheck(clientId, otherParties);
                        } else {
                          message.warning('请先选择委托人');
                        }
                      }}
                      size="large"
                    >
                      {loading ? '正在执行冲突检索...' : '执行冲突检索'}
                    </Button>
                    
                    {/* 检索进度显示 */}
                    {conflictCheckProgress && (
                      <Card 
                        size="small" 
                        style={{ 
                          backgroundColor: '#f8f9fa',
                          border: '1px solid #e9ecef'
                        }}
                      >
                        <div style={{ marginBottom: 12 }}>
                          <Text strong style={{ color: '#1890ff' }}>
                            步骤 {conflictCheckProgress.step}/6: {conflictCheckProgress.stepName}
                          </Text>
                          <Progress 
                            percent={conflictCheckProgress.progress} 
                            size="small" 
                            style={{ marginTop: 8 }}
                            status={conflictCheckProgress.isCompleted ? 'success' : 'active'}
                          />
                        </div>
                        <div>
                          {conflictCheckProgress.details.map((detail, index) => (
                            <div key={index} style={{ 
                              fontSize: '12px', 
                              color: '#666',
                              marginBottom: '4px',
                              paddingLeft: '16px',
                              position: 'relative'
                            }}>
                              <span style={{
                                position: 'absolute',
                                left: '0',
                                top: '2px',
                                width: '6px',
                                height: '6px',
                                backgroundColor: '#1890ff',
                                borderRadius: '50%'
                              }}></span>
                              {detail}
                            </div>
                          ))}
                        </div>
                      </Card>
                    )}
                    
                    {/* 检索结果显示 */}
                    {conflictCheckResult && (
                      <Card size="small">
                        <Alert
                          message={conflictCheckResult.hasConflict ? '发现潜在冲突' : '无冲突'}
                          description={
                            <div>
                              <div style={{ marginBottom: 12 }}>
                                <Text strong>检索统计：</Text>
                                <ul style={{ margin: '8px 0', paddingLeft: '20px' }}>
                                  <li>已检索案件数量：{conflictCheckResult.checkDetails.totalCasesChecked} 件</li>
                                  <li>委托人历史案件：{conflictCheckResult.checkDetails.clientHistoryCases} 件</li>
                                  <li>已检索当事人：{conflictCheckResult.checkDetails.relatedPartiesChecked} 人</li>
                                  {conflictCheckResult.checkDetails.corporateRelationsChecked > 0 && (
                                    <li>关联企业检索：{conflictCheckResult.checkDetails.corporateRelationsChecked} 家</li>
                                  )}
                                  <li>检索时间范围：{conflictCheckResult.checkDetails.timeRange}</li>
                                  <li>风险评估：{conflictCheckResult.checkDetails.riskAssessment}</li>
                                </ul>
                              </div>
                              
                              {conflictCheckResult.hasConflict && (
                                <div style={{ marginBottom: 12 }}>
                                  <Text strong style={{ color: '#ff4d4f' }}>发现冲突案件：</Text>
                                  {conflictCheckResult.conflictCases.map((conflict: any, index: number) => (
                                    <Card key={index} size="small" style={{ marginTop: 8 }}>
                                      <div style={{ fontSize: '13px' }}>
                                        <div><Text strong>案件名称：</Text>{conflict.caseName}</div>
                                        <div><Text strong>案件编号：</Text>{conflict.caseNo}</div>
                                        <div><Text strong>冲突类型：</Text>
                                          <Tag color={conflict.riskLevel === 'HIGH' ? 'red' : 'orange'}>
                                            {conflict.conflictType}
                                          </Tag>
                                        </div>
                                        <div><Text strong>冲突说明：</Text>{conflict.description}</div>
                                        <div><Text strong>案件状态：</Text>{conflict.caseStatus}</div>
                                      </div>
                                    </Card>
                                  ))}
                                </div>
                              )}
                              
                              <div>
                                <Text strong>建议措施：</Text>
                                <ul style={{ margin: '8px 0', paddingLeft: '20px' }}>
                                  {conflictCheckResult.recommendations.map((rec: string, index: number) => (
                                    <li key={index} style={{ marginBottom: '4px' }}>{rec}</li>
                                  ))}
                                </ul>
                              </div>
                            </div>
                          }
                          type={conflictCheckResult.hasConflict ? 'error' : 'success'}
                          showIcon
                        />
                      </Card>
                    )}
                    
                    <Radio.Group>
                      <Radio value="NO_CONFLICT">无冲突（已完成检索）</Radio>
                      <Radio value="CONFLICT_RESOLVED">有冲突但已解决</Radio>
                      <Radio value="MANUAL_CHECK">手动检索确认</Radio>
                    </Radio.Group>
                  </Space>
                </Form.Item>
              </Col>

              <Col span={24}>
                <Form.Item
                  label="风险标签"
                  name="riskTags"
                  tooltip="请勾选适用的风险类别"
                >
                  <Checkbox.Group style={{ width: '100%' }}>
                    <Row>
                      {riskCategories.map(category => (
                        <Col span={24} key={category.value} style={{ marginBottom: 8 }}>
                          <Checkbox value={category.value}>
                            <Space direction="vertical" size={0}>
                              <Text strong>{category.label}</Text>
                              <Text type="secondary" style={{ fontSize: '12px' }}>
                                {category.description}
                              </Text>
                            </Space>
                          </Checkbox>
                        </Col>
                      ))}
                    </Row>
                  </Checkbox.Group>
                </Form.Item>
              </Col>

              <Col span={24}>
                <Form.Item name="isHighRisk" valuePropName="checked">
                  <Checkbox>
                    <Space>
                      <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
                      <Text strong>标记为重大风险项目</Text>
                      <Text type="secondary">（需要合伙人审批）</Text>
                    </Space>
                  </Checkbox>
                </Form.Item>
              </Col>
            </Row>
          </Card>
        );

      case 3:
        return (
          <Card title="文档管理" className="step-card">
            <Alert
              message="文档上传要求"
              description="请上传必要的案件文档，单个文件大小不超过10MB，支持PDF、DOC、DOCX格式。"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />

            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  label="委托代理合同"
                  name="contractDocument"
                  tooltip="需要上传签署的委托代理合同"
                >
                  <Upload
                    accept=".pdf,.doc,.docx"
                    beforeUpload={(file) => {
                      const isValidSize = file.size / 1024 / 1024 < 10;
                      if (!isValidSize) {
                        message.error('文件大小不能超过10MB!');
                      }
                      return isValidSize;
                    }}
                    onChange={(info) => handleUpload(info, 'contractDocument')}
                  >
                    <Button icon={<UploadOutlined />}>上传合同文档</Button>
                  </Upload>
                  {uploadProgress > 0 && (
                    <Progress percent={uploadProgress} size="small" />
                  )}
                </Form.Item>
              </Col>

              <Col span={12}>
                <Form.Item
                  label="律师所函"
                  name="retainerAgreement"
                  tooltip="律师事务所接受委托的正式文件"
                >
                  <Upload
                    accept=".pdf,.doc,.docx"
                    beforeUpload={(file) => {
                      const isValidSize = file.size / 1024 / 1024 < 10;
                      if (!isValidSize) {
                        message.error('文件大小不能超过10MB!');
                      }
                      return isValidSize;
                    }}
                    onChange={(info) => handleUpload(info, 'retainerAgreement')}
                  >
                    <Button icon={<UploadOutlined />}>上传所函</Button>
                  </Upload>
                </Form.Item>
              </Col>

              <Col span={24}>
                <Form.Item
                  label="其他文档"
                  name="otherDocuments"
                  tooltip="其他相关文件，如授权委托书、身份证明等"
                >
                  <Upload
                    multiple
                    accept=".pdf,.doc,.docx,.jpg,.png"
                    beforeUpload={(file) => {
                      const isValidSize = file.size / 1024 / 1024 < 10;
                      if (!isValidSize) {
                        message.error('文件大小不能超过10MB!');
                      }
                      return isValidSize;
                    }}
                    onChange={(info) => handleUpload(info, 'otherDocuments')}
                  >
                    <Button icon={<UploadOutlined />}>上传其他文档</Button>
                  </Upload>
                </Form.Item>
              </Col>
            </Row>
          </Card>
        );

      default:
        return null;
    }
  };

  return (
    <Modal
      title="新建立案"
      open={visible}
      onCancel={onCancel}
      footer={null}
      width={1000}
      destroyOnClose
    >
      <div className="create-case-container">
        <Steps current={currentStep} style={{ marginBottom: 24 }}>
          {steps.map((step, index) => (
            <Step key={index} title={step.title} icon={step.icon} />
          ))}
        </Steps>

        <Form
          form={form}
          layout="vertical"
          requiredMark="optional"
        >
          {renderStepContent()}
        </Form>

        <Divider />

        <Row justify="space-between">
          <Col>
            {currentStep > 0 && (
              <Button onClick={() => setCurrentStep(currentStep - 1)}>
                上一步
              </Button>
            )}
          </Col>
          <Col>
            <Space>
              <Button onClick={onCancel}>
                取消
              </Button>
              {currentStep < steps.length - 1 ? (
                <Button
                  type="primary"
                  onClick={async () => {
                    const isValid = await validateStep(currentStep);
                    if (isValid) {
                      setCurrentStep(currentStep + 1);
                    }
                  }}
                >
                  下一步
                </Button>
              ) : (
                <Button
                  type="primary"
                  loading={loading}
                  onClick={handleSubmit}
                  icon={<CheckCircleOutlined />}
                >
                  创建案件
                </Button>
              )}
            </Space>
          </Col>
        </Row>
      </div>
    </Modal>
  );
};

export default CreateCase;