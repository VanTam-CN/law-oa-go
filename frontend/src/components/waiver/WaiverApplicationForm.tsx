import React, { useState, useEffect, useCallback } from 'react';
import {
  Form,
  Input,
  Select,
  Button,
  Card,
  Row,
  Col,
  Upload,
  message,
  Steps,
  Divider,
  Space,
  Alert,
  Tag,
  Table,
  Modal,
  Checkbox,
  Radio,
  InputNumber,
  DatePicker,
  Tooltip,
  Typography,
  Popconfirm,
  Progress,
  Badge,
  Tabs,
} from 'antd';
import {
  UploadOutlined,
  PlusOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  FileTextOutlined,
  UserOutlined,
  WarningOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import type { UploadProps, UploadFile } from 'antd';
import dayjs from 'dayjs';
import { waiverApprovalService } from '@/services/waiverApproval';
import type {
  WaiverType,
  RiskLevel,
  CreateWaiverApplicationRequest,
  Stakeholder,
  StakeholderType,
  WaiverApplication,
  WaiverTemplate,
} from '@/types/waiverApproval';

const { TextArea } = Input;
const { Option } = Select;
const { Step } = Steps;
const { Title, Text } = Typography;
const { TabPane } = Tabs;

interface WaiverApplicationFormProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: (application: WaiverApplication) => void;
  initialData?: Partial<CreateWaiverApplicationRequest>;
  caseId?: string;
  caseTitle?: string;
}

// 豁免类型配置
const waiverTypeConfig: Record<WaiverType, {
  label: string;
  description: string;
  defaultRiskLevel: RiskLevel;
  requiredFields: string[];
  stakeholderTypes: StakeholderType[];
}> = {
  CONFLICT_OF_INTEREST: {
    label: '利益冲突',
    description: '律师或律师事务所与客户存在潜在或实际的利益冲突',
    defaultRiskLevel: 'HIGH',
    requiredFields: ['description', 'justification', 'affectedParties'],
    stakeholderTypes: ['CLIENT', 'LAWYER', 'PARTY', 'ORGANIZATION'],
  },
  BUSINESS_RELATIONSHIP: {
    label: '业务关系',
    description: '与客户存在除法律服务外的其他业务关系',
    defaultRiskLevel: 'MEDIUM',
    requiredFields: ['description', 'justification', 'mitigationMeasures'],
    stakeholderTypes: ['CLIENT', 'ORGANIZATION'],
  },
  REPRESENTATION_CONFLICT: {
    label: '代理冲突',
    description: '同时代理利益冲突的多方当事人',
    defaultRiskLevel: 'HIGH',
    requiredFields: ['description', 'justification', 'affectedParties'],
    stakeholderTypes: ['CLIENT', 'PARTY'],
  },
  FINANCIAL_INTEREST: {
    label: '财务利益',
    description: '在客户业务中拥有直接或间接财务利益',
    defaultRiskLevel: 'CRITICAL',
    requiredFields: ['description', 'justification', 'mitigationMeasures'],
    stakeholderTypes: ['CLIENT', 'ORGANIZATION'],
  },
  PERSONAL_RELATIONSHIP: {
    label: '个人关系',
    description: '与客户存在密切个人关系可能影响独立判断',
    defaultRiskLevel: 'MEDIUM',
    requiredFields: ['description', 'justification'],
    stakeholderTypes: ['CLIENT', 'LAWYER'],
  },
  ORGANIZATIONAL: {
    label: '组织冲突',
    description: '组织结构导致的潜在冲突',
    defaultRiskLevel: 'LOW',
    requiredFields: ['description', 'justification'],
    stakeholderTypes: ['LAWYER', 'ORGANIZATION'],
  },
};

// 风险等级配置
const riskLevelConfig: Record<RiskLevel, {
  label: string;
  color: string;
  description: string;
  requiredApproval: string[];
}> = {
  LOW: {
    label: '低风险',
    color: 'green',
    description: '轻微冲突，可由主管律师直接批准',
    requiredApproval: ['SENIOR_LAWYER'],
  },
  MEDIUM: {
    label: '中风险',
    color: 'orange',
    description: '中等风险，需要合伙人级别批准',
    requiredApproval: ['PARTNER'],
  },
  HIGH: {
    label: '高风险',
    color: 'red',
    description: '高风险，需要管理委员会批准',
    requiredApproval: ['MANAGEMENT_COMMITTEE'],
  },
  CRITICAL: {
    label: '关键风险',
    color: 'purple',
    description: '关键风险，需要全体合伙人批准',
    requiredApproval: ['ALL_PARTNERS'],
  },
};

const WaiverApplicationForm: React.FC<WaiverApplicationFormProps> = ({
  visible,
  onCancel,
  onSuccess,
  initialData,
  caseId,
  caseTitle,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [currentStep, setCurrentStep] = useState(0);
  const [attachments, setAttachments] = useState<UploadFile[]>([]);
  const [stakeholders, setStakeholders] = useState<Stakeholder[]>([]);
  const [selectedWaiverType, setSelectedWaiverType] = useState<WaiverType>();
  const [selectedRiskLevel, setSelectedRiskLevel] = useState<RiskLevel>();
  const [availableTemplates, setAvailableTemplates] = useState<WaiverTemplate[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<WaiverTemplate>();
  const [showPreview, setShowPreview] = useState(false);
  const [formData, setFormData] = useState<Partial<CreateWaiverApplicationRequest>>({});

  // 初始化表单
  useEffect(() => {
    if (visible) {
      form.resetFields();
      setAttachments([]);
      setStakeholders([]);
      setCurrentStep(0);
      setSelectedWaiverType(undefined);
      setSelectedRiskLevel(undefined);
      setSelectedTemplate(undefined);
      setFormData({});

      if (caseId) {
        form.setFieldsValue({ caseId });
      }
      if (initialData) {
        form.setFieldsValue(initialData);
        setFormData(initialData);
        if (initialData.waiverType) {
          setSelectedWaiverType(initialData.waiverType);
          const config = waiverTypeConfig[initialData.waiverType];
          setSelectedRiskLevel(config.defaultRiskLevel);
          form.setFieldsValue({ riskLevel: config.defaultRiskLevel });
        }
      }
    }
  }, [visible, form, caseId, initialData]);

  // 获取可用模板
  const loadTemplates = useCallback(async (waiverType?: WaiverType, riskLevel?: RiskLevel) => {
    if (!waiverType || !riskLevel) return;

    try {
      const response = await waiverApprovalService.getWaiverTemplates({
        waiverType,
        riskLevel,
        isActive: true,
      });
      if (response.success) {
        setAvailableTemplates(response.data);
      }
    } catch (error) {
      console.error('加载模板失败:', error);
    }
  }, []);

  // 豁免类型变化处理
  const handleWaiverTypeChange = (value: WaiverType) => {
    setSelectedWaiverType(value);
    const config = waiverTypeConfig[value];
    setSelectedRiskLevel(config.defaultRiskLevel);
    form.setFieldsValue({
      riskLevel: config.defaultRiskLevel,
      description: '',
      justification: '',
      mitigationMeasures: '',
    });
    loadTemplates(value, config.defaultRiskLevel);
  };

  // 风险等级变化处理
  const handleRiskLevelChange = (value: RiskLevel) => {
    setSelectedRiskLevel(value);
    if (selectedWaiverType) {
      loadTemplates(selectedWaiverType, value);
    }
  };

  // 模板选择处理
  const handleTemplateSelect = async (templateId: string) => {
    try {
      const response = await waiverApprovalService.getWaiverTemplate(templateId);
      if (response.success) {
        const template = response.data;
        setSelectedTemplate(template);

        // 根据模板填充表单
        form.setFieldsValue({
          description: template.template,
          justification: '',
          mitigationMeasures: '',
        });
      }
    } catch (error) {
      message.error('加载模板失败');
    }
  };

  // 添加利益相关方
  const addStakeholder = () => {
    const newStakeholder: Stakeholder = {
      id: `stakeholder_${Date.now()}`,
      name: '',
      type: 'CLIENT',
      relationshipDescription: '',
      conflictDetails: '',
    };
    setStakeholders([...stakeholders, newStakeholder]);
  };

  // 更新利益相关方
  const updateStakeholder = (index: number, field: keyof Stakeholder, value: any) => {
    const updatedStakeholders = [...stakeholders];
    updatedStakeholders[index] = {
      ...updatedStakeholders[index],
      [field]: value,
    };
    setStakeholders(updatedStakeholders);
  };

  // 删除利益相关方
  const removeStakeholder = (index: number) => {
    setStakeholders(stakeholders.filter((_, i) => i !== index));
  };

  // 文件上传配置
  const uploadProps: UploadProps = {
    multiple: true,
    fileList: attachments,
    onChange: ({ fileList }) => {
      setAttachments(fileList);
    },
    beforeUpload: (file) => {
      const isValidType = ['application/pdf', 'application/msword',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'image/jpeg', 'image/png'].includes(file.type);

      if (!isValidType) {
        message.error('只支持 PDF、DOC、DOCX、JPG、PNG 格式的文件');
        return false;
      }

      const isLt10M = file.size / 1024 / 1024 < 10;
      if (!isLt10M) {
        message.error('文件大小不能超过 10MB');
        return false;
      }

      return false; // 阻止自动上传
    },
  };

  // 下一步
  const handleNext = async () => {
    try {
      if (currentStep === 0) {
        // 验证基础信息
        await form.validateFields(['waiverType', 'riskLevel', 'description']);
      } else if (currentStep === 1) {
        // 验证详细信息
        const config = selectedWaiverType ? waiverTypeConfig[selectedWaiverType] : null;
        const requiredFields = config?.requiredFields || [];
        await form.validateFields(requiredFields);

        // 验证利益相关方
        if (stakeholders.length === 0) {
          message.error('请至少添加一个利益相关方');
          return;
        }
      }

      setCurrentStep(currentStep + 1);
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };

  // 上一步
  const handlePrev = () => {
    setCurrentStep(currentStep - 1);
  };

  // 预览申请
  const handlePreview = () => {
    form.validateFields().then(values => {
      setFormData({
        ...values,
        stakeholders,
        attachments: attachments.map(file => file.originFileObj).filter(Boolean) as File[],
      });
      setShowPreview(true);
    });
  };

  // 提交申请
  const handleSubmit = async (submitImmediately = true) => {
    try {
      setLoading(true);
      const values = await form.validateFields();

      const request: CreateWaiverApplicationRequest = {
        caseId: values.caseId || caseId!,
        waiverType: values.waiverType,
        description: values.description,
        justification: values.justification || '',
        mitigationMeasures: values.mitigationMeasures || '',
        affectedParties: values.affectedParties || [],
        stakeholders,
        attachments: attachments.map(file => file.originFileObj).filter(Boolean) as File[],
        submitImmediately,
      };

      const response = await waiverApprovalService.createWaiverApplication(request);

      if (response.success) {
        message.success(submitImmediately ? '豁免申请提交成功' : '豁免申请保存为草稿');
        onSuccess(response.data);
        onCancel();
      }
    } catch (error: any) {
      message.error(error.message || '提交失败');
    } finally {
      setLoading(false);
    }
  };

  // 渲染基础信息步骤
  const renderBasicInfo = () => (
    <Card title="基础信息" size="small">
      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            label="案件编号"
            name="caseId"
            rules={[{ required: true, message: '请输入案件编号' }]}
          >
            <Input placeholder="请输入案件编号" disabled={!!caseId} />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item
            label="豁免类型"
            name="waiverType"
            rules={[{ required: true, message: '请选择豁免类型' }]}
          >
            <Select
              placeholder="请选择豁免类型"
              onChange={handleWaiverTypeChange}
              showSearch
              filterOption={(input, option) =>
                option?.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
              }
            >
              {Object.entries(waiverTypeConfig).map(([key, config]) => (
                <Option key={key} value={key}>
                  <div>
                    <div style={{ fontWeight: 'bold' }}>{config.label}</div>
                    <div style={{ fontSize: '12px', color: '#666' }}>
                      {config.description}
                    </div>
                  </div>
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
      </Row>

      {selectedWaiverType && (
        <Alert
          message="豁免类型说明"
          description={waiverTypeConfig[selectedWaiverType].description}
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            label="风险等级"
            name="riskLevel"
            rules={[{ required: true, message: '请选择风险等级' }]}
          >
            <Select
              placeholder="请选择风险等级"
              onChange={handleRiskLevelChange}
            >
              {Object.entries(riskLevelConfig).map(([key, config]) => (
                <Option key={key} value={key}>
                  <Badge color={config.color} text={config.label} />
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
        <Col span={12}>
          {selectedRiskLevel && (
            <Alert
              message="审批要求"
              description={`需要${riskLevelConfig[selectedRiskLevel].requiredApproval.join('、')}批准`}
              type="warning"
              showIcon
              style={{ marginTop: 24 }}
            />
          )}
        </Col>
      </Row>

      <Form.Item
        label="豁免描述"
        name="description"
        rules={[{ required: true, message: '请输入豁免描述' }]}
      >
        <TextArea
          rows={4}
          placeholder="请详细描述豁免的具体情况和背景"
          showCount
          maxLength={2000}
        />
      </Form.Item>

      {availableTemplates.length > 0 && (
        <Form.Item label="使用模板">
          <Select
            placeholder="选择模板快速填写"
            allowClear
            onChange={handleTemplateSelect}
          >
            {availableTemplates.map(template => (
              <Option key={template.id} value={template.id}>
                {template.name} - {template.description}
              </Option>
            ))}
          </Select>
        </Form.Item>
      )}
    </Card>
  );

  // 渲染详细信息步骤
  const renderDetailedInfo = () => {
    const config = selectedWaiverType ? waiverTypeConfig[selectedWaiverType] : null;
    const requiredFields = config?.requiredFields || [];

    return (
      <Card title="详细信息" size="small">
        {requiredFields.includes('justification') && (
          <Form.Item
            label="申请理由"
            name="justification"
            rules={[{ required: true, message: '请输入申请理由' }]}
          >
            <TextArea
              rows={4}
              placeholder="请详细说明申请豁免的理由和依据"
              showCount
              maxLength={2000}
            />
          </Form.Item>
        )}

        {requiredFields.includes('mitigationMeasures') && (
          <Form.Item
            label="风险控制措施"
            name="mitigationMeasures"
            rules={[{ required: true, message: '请输入风险控制措施' }]}
          >
            <TextArea
              rows={3}
              placeholder="请说明将采取哪些措施来控制和管理相关风险"
              showCount
              maxLength={1500}
            />
          </Form.Item>
        )}

        {requiredFields.includes('affectedParties') && (
          <Form.Item
            label="受影响方"
            name="affectedParties"
          >
            <Select
              mode="tags"
              placeholder="请输入受影响的当事人或组织"
              style={{ width: '100%' }}
            />
          </Form.Item>
        )}

        <Divider>利益相关方</Divider>

        <div style={{ marginBottom: 16 }}>
          <Button
            type="dashed"
            onClick={addStakeholder}
            icon={<PlusOutlined />}
            block
          >
            添加利益相关方
          </Button>
        </div>

        {stakeholders.map((stakeholder, index) => (
          <Card
            key={stakeholder.id}
            size="small"
            title={`利益相关方 ${index + 1}`}
            extra={
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
                onClick={() => removeStakeholder(index)}
              />
            }
            style={{ marginBottom: 16 }}
          >
            <Row gutter={16}>
              <Col span={8}>
                <Input
                  placeholder="姓名/组织名称"
                  value={stakeholder.name}
                  onChange={(e) => updateStakeholder(index, 'name', e.target.value)}
                />
              </Col>
              <Col span={8}>
                <Select
                  placeholder="类型"
                  value={stakeholder.type}
                  onChange={(value) => updateStakeholder(index, 'type', value)}
                  style={{ width: '100%' }}
                >
                  <Option value="CLIENT">客户</Option>
                  <Option value="LAWYER">律师</Option>
                  <Option value="PARTY">当事人</Option>
                  <Option value="WITNESS">证人</Option>
                  <Option value="EXPERT">专家</Option>
                  <Option value="ORGANIZATION">组织</Option>
                </Select>
              </Col>
              <Col span={8}>
                <Input
                  placeholder="组织/机构"
                  value={stakeholder.organization}
                  onChange={(e) => updateStakeholder(index, 'organization', e.target.value)}
                />
              </Col>
            </Row>
            <Row gutter={16} style={{ marginTop: 8 }}>
              <Col span={24}>
                <Input
                  placeholder="关系描述"
                  value={stakeholder.relationshipDescription}
                  onChange={(e) => updateStakeholder(index, 'relationshipDescription', e.target.value)}
                />
              </Col>
            </Row>
            <Row gutter={16} style={{ marginTop: 8 }}>
              <Col span={24}>
                <TextArea
                  rows={2}
                  placeholder="冲突详情"
                  value={stakeholder.conflictDetails}
                  onChange={(e) => updateStakeholder(index, 'conflictDetails', e.target.value)}
                />
              </Col>
            </Row>
          </Card>
        ))}

        <Divider>相关附件</Divider>

        <Upload {...uploadProps}>
          <Button icon={<UploadOutlined />}>上传附件</Button>
        </Upload>

        <div style={{ color: '#666', fontSize: '12px', marginTop: 8 }}>
          支持格式：PDF、DOC、DOCX、JPG、PNG，单个文件不超过10MB
        </div>
      </Card>
    );
  };

  // 渲染确认步骤
  const renderConfirmation = () => (
    <Card title="申请确认" size="small">
      <Alert
        message="请确认申请信息"
        description="提交后申请将进入审批流程，请确保所有信息准确无误"
        type="warning"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <div style={{ background: '#f5f5f5', padding: 16, borderRadius: 6 }}>
        <Row gutter={16}>
          <Col span={12}>
            <Text strong>案件编号：</Text>
            <Text>{form.getFieldValue('caseId') || caseId}</Text>
          </Col>
          <Col span={12}>
            <Text strong>豁免类型：</Text>
            <Tag color="blue">
              {selectedWaiverType ? waiverTypeConfig[selectedWaiverType].label : '-'}
            </Tag>
          </Col>
        </Row>
        <Row gutter={16} style={{ marginTop: 8 }}>
          <Col span={12}>
            <Text strong>风险等级：</Text>
            {selectedRiskLevel && (
              <Badge color={riskLevelConfig[selectedRiskLevel].color} text={riskLevelConfig[selectedRiskLevel].label} />
            )}
          </Col>
          <Col span={12}>
            <Text strong>利益相关方：</Text>
            <Text>{stakeholders.length} 个</Text>
          </Col>
        </Row>
        <Row gutter={16} style={{ marginTop: 8 }}>
          <Col span={12}>
            <Text strong>附件数量：</Text>
            <Text>{attachments.length} 个</Text>
          </Col>
          <Col span={12}>
            <Text strong>提交时间：</Text>
            <Text>{dayjs().format('YYYY-MM-DD HH:mm')}</Text>
          </Col>
        </Row>
      </div>

      <div style={{ marginTop: 16 }}>
        <Text strong>描述摘要：</Text>
        <div style={{
          background: '#f9f9f9',
          padding: 12,
          borderRadius: 4,
          marginTop: 8,
          maxHeight: 100,
          overflow: 'auto'
        }}>
          {form.getFieldValue('description') || '-'}
        </div>
      </div>

      {selectedRiskLevel && (
        <Alert
          message="审批流程"
          description={`此申请需要${riskLevelConfig[selectedRiskLevel].requiredApproval.join('、')}批准，预计审批时间${selectedRiskLevel === 'LOW' ? '1-2' : selectedRiskLevel === 'MEDIUM' ? '3-5' : selectedRiskLevel === 'HIGH' ? '5-7' : '7-10'}个工作日`}
          type="info"
          showIcon
          style={{ marginTop: 16 }}
        />
      )}
    </Card>
  );

  // 步骤配置
  const steps = [
    {
      title: '基础信息',
      description: '选择豁免类型和风险等级',
      content: renderBasicInfo(),
    },
    {
      title: '详细信息',
      description: '填写详细信息和利益相关方',
      content: renderDetailedInfo(),
    },
    {
      title: '确认提交',
      description: '确认信息并提交申请',
      content: renderConfirmation(),
    },
  ];

  return (
    <Modal
      title="创建豁免申请"
      open={visible}
      onCancel={onCancel}
      width={1000}
      footer={null}
      destroyOnClose
    >
      <Form form={form} layout="vertical">
        <Steps current={currentStep} style={{ marginBottom: 24 }}>
          {steps.map((step, index) => (
            <Step
              key={index}
              title={step.title}
              description={step.description}
              icon={index < currentStep ? <CheckCircleOutlined /> :
                    index === currentStep ? <ClockCircleOutlined /> :
                    <FileTextOutlined />}
            />
          ))}
        </Steps>

        <div style={{ minHeight: 400 }}>
          {steps[currentStep].content}
        </div>

        <Divider />

        <div style={{ textAlign: 'right' }}>
          <Space>
            <Button onClick={onCancel}>
              取消
            </Button>

            {currentStep > 0 && (
              <Button onClick={handlePrev}>
                上一步
              </Button>
            )}

            {currentStep < steps.length - 1 && (
              <Button type="primary" onClick={handleNext}>
                下一步
              </Button>
            )}

            {currentStep === steps.length - 1 && (
              <>
                <Button onClick={handlePreview}>
                  预览
                </Button>
                <Button onClick={() => handleSubmit(false)}>
                  保存草稿
                </Button>
                <Popconfirm
                  title="确认提交申请？"
                  description="提交后将无法修改，请确认所有信息准确无误"
                  onConfirm={() => handleSubmit(true)}
                  okText="确认提交"
                  cancelText="取消"
                >
                  <Button type="primary" loading={loading}>
                    提交申请
                  </Button>
                </Popconfirm>
              </>
            )}
          </Space>
        </div>
      </Form>

      {/* 预览模态框 */}
      <Modal
        title="申请预览"
        open={showPreview}
        onCancel={() => setShowPreview(false)}
        width={800}
        footer={[
          <Button key="close" onClick={() => setShowPreview(false)}>
            关闭
          </Button>,
          <Button key="submit" type="primary" onClick={() => handleSubmit(true)} loading={loading}>
            确认提交
          </Button>,
        ]}
      >
        <div style={{ maxHeight: 600, overflow: 'auto' }}>
          {/* 预览内容可以根据需要展开 */}
          <pre>{JSON.stringify(formData, null, 2)}</pre>
        </div>
      </Modal>
    </Modal>
  );
};

export default WaiverApplicationForm;