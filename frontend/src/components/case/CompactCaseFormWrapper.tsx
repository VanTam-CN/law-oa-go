/**
 * CompactCaseFormWrapper组件
 * 简化的案件创建表单包装器，专门为CreateCase.tsx设计
 * 解决复杂的CompactCaseForm与简单数据传递之间的不匹配问题
 */

import React, { useState, useEffect } from 'react'
import {
  Form,
  Button,
  Space,
  Card,
  Typography,
  message,
  Row,
  Col,
  Input,
  Select,
  DatePicker,
  InputNumber,
  Upload,
  Checkbox,
  Radio,
  Alert,
  Progress,
  AutoComplete,
  Divider,
  Tag,
} from 'antd'
import {
  SaveOutlined,
  ReloadOutlined,
  SearchOutlined,
  UploadOutlined,
  UserOutlined,
  FileTextOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
  DollarOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'

const { Option } = Select
const { TextArea } = Input
const { Text, Title } = Typography

interface CompactCaseFormWrapperProps {
  // 客户数据
  clients: any[]
  // 律师数据
  lawyers: any[]
  // 案件类型数据
  caseTypes: any[]
  // 风险类别数据
  riskCategories: any[]
  // 提交处理
  onSubmit: (values: any) => void
  // 取消处理
  onCancel: () => void
  // 冲突检测处理
  onConflictCheck: (clientId: string, otherParties: string[]) => void
  // 加载状态
  loading: boolean
  // 数据加载状态
  dataLoading: boolean
}

/**
 * 案件创建步骤配置
 */
const CASE_STEPS = [
  { key: 'basic', title: '基本信息', icon: <FileTextOutlined /> },
  { key: 'management', title: '内部管理', icon: <TeamOutlined /> },
  { key: 'compliance', title: '合规风控', icon: <SafetyCertificateOutlined /> },
  { key: 'documents', title: '文档管理', icon: <UploadOutlined /> },
]

/**
 * CompactCaseFormWrapper组件
 */
const CompactCaseFormWrapper: React.FC<CompactCaseFormWrapperProps> = ({
  clients,
  lawyers,
  caseTypes,
  riskCategories,
  onSubmit,
  onCancel,
  onConflictCheck,
  loading,
  dataLoading,
}) => {
  const [form] = Form.useForm()
  const [currentStep, setCurrentStep] = useState(0)
  const [conflictCheckResult, setConflictCheckResult] = useState<any>(null)

  // 案件类型fallback数据
  const fallbackCaseTypes = [
    {
      value: '民事案件',
      label: '民事案件',
      causes: ['合同纠纷', '侵权责任', '婚姻家庭', '继承纠纷'],
    },
    {
      value: '商事案件',
      label: '商事案件',
      causes: ['公司纠纷', '金融纠纷', '投资纠纷', '证券纠纷'],
    },
    {
      value: '刑事案件',
      label: '刑事案件',
      causes: ['经济犯罪', '职务犯罪', '暴力犯罪', '网络犯罪'],
    },
    {
      value: '行政案件',
      label: '行政案件',
      causes: ['行政处罚', '行政许可', '信息公开', '征收补偿'],
    },
    {
      value: '知识产权',
      label: '知识产权案件',
      causes: ['商标侵权', '专利侵权', '著作权侵权', '商业秘密'],
    },
  ]

  /**
   * 渲染步骤指示器
   */
  const renderStepIndicator = () => (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        marginBottom: 16,
        padding: '12px 16px',
        background: '#f8f9fa',
        borderRadius: '8px',
        border: '1px solid #e9ecef',
      }}
    >
      {CASE_STEPS.map((step, index) => (
        <div
          key={step.key}
          style={{
            display: 'flex',
            alignItems: 'center',
            flexDirection: 'column',
            gap: '8px',
            flex: 1,
            position: 'relative',
          }}
        >
          <div
            style={{
              width: '32px',
              height: '32px',
              borderRadius: '50%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background:
                currentStep === index ? '#1890ff' : currentStep > index ? '#52c41a' : '#f0f0f0',
              color: currentStep === index ? 'white' : currentStep > index ? 'white' : '#999',
              fontSize: '14px',
              fontWeight: '600',
              cursor: 'pointer',
              transition: 'all 0.3s ease',
            }}
            onClick={() => index < currentStep && setCurrentStep(index)}
          >
            {currentStep > index ? <CheckCircleOutlined /> : step.icon}
          </div>
          <Text
            style={{
              fontSize: '12px',
              fontWeight: currentStep === index ? '600' : '400',
              color: currentStep === index ? '#1890ff' : currentStep > index ? '#52c41a' : '#666',
              textAlign: 'center',
            }}
          >
            {step.title}
          </Text>
          {index < CASE_STEPS.length - 1 && (
            <div
              style={{
                position: 'absolute',
                top: '16px',
                right: '-50%',
                width: '100%',
                height: '2px',
                background: currentStep > index ? '#52c41a' : '#e9ecef',
                zIndex: 0,
              }}
            />
          )}
        </div>
      ))}
    </div>
  )

  /**
   * 渲染基本信息步骤
   */
  const renderBasicInfo = () => (
    <Card title='案件基本信息' size='small' style={{ marginBottom: 16 }}>
      <Row gutter={[12, 12]}>
        <Col span={24}>
          <Form.Item
            label='案件名称'
            name='caseName'
            rules={[
              { required: true, message: '请输入案件名称' },
              { max: 100, message: '案件名称不能超过100个字符' },
            ]}
          >
            <Input placeholder='例如：张三 Vs 李四 - 合同纠纷' prefix={<FileTextOutlined />} />
          </Form.Item>
        </Col>

        <Col span={12}>
          <Form.Item
            label='委托人'
            name='clientId'
            rules={[{ required: true, message: '请选择委托人' }]}
          >
            <Select
              placeholder={dataLoading ? '正在加载...' : '选择委托人'}
              showSearch
              loading={dataLoading}
              filterOption={(input, option) =>
                (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
              }
              notFoundContent={dataLoading ? '正在加载...' : '暂无数据'}
            >
              {clients.map((client) => (
                <Option key={client.id} value={client.id}>
                  <Space>
                    <UserOutlined />
                    {client.name}
                    {client.company && <Tag color='blue'>{client.company}</Tag>}
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
          <Form.Item label='其他当事人' name='otherParties'>
            <Select mode='tags' placeholder='输入对方当事人姓名' style={{ width: '100%' }} />
          </Form.Item>
        </Col>

        <Col span={12}>
          <Form.Item
            label='案件类型'
            name='caseType'
            rules={[{ required: true, message: '请选择案件类型' }]}
          >
            <Select
              placeholder='选择案件类型'
              onChange={() => {
                form.setFieldsValue({ causeOfAction: undefined })
              }}
            >
              {(caseTypes.length > 0 ? caseTypes : fallbackCaseTypes).map((type) => (
                <Option key={type.value} value={type.value}>
                  {type.label}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>

        <Col span={12}>
          <Form.Item
            label='案由'
            name='causeOfAction'
            rules={[
              { required: true, message: '请选择或输入案由' },
              { max: 100, message: '案由不能超过100个字符' },
            ]}
          >
            <AutoComplete
              placeholder='选择或输入具体案由'
              allowClear
              options={
                form.getFieldValue('caseType')
                  ? (caseTypes.length > 0 ? caseTypes : fallbackCaseTypes)
                      .find((type) => type.value === form.getFieldValue('caseType'))
                      ?.causes.map((cause) => ({ value: cause, label: cause })) || []
                  : []
              }
              filterOption={(inputValue, option) =>
                option?.value.toUpperCase().indexOf(inputValue.toUpperCase()) !== -1
              }
            />
          </Form.Item>
        </Col>

        <Col span={24}>
          <Form.Item
            label='案件描述'
            name='caseDescription'
            rules={[
              { required: true, message: '请输入案件描述' },
              { max: 500, message: '描述不能超过500个字符' },
            ]}
          >
            <TextArea
              rows={3}
              placeholder='请详细描述案件背景、争议焦点等关键信息...'
              showCount
              maxLength={500}
            />
          </Form.Item>
        </Col>
      </Row>
    </Card>
  )

  /**
   * 渲染内部管理步骤
   */
  const renderManagementInfo = () => (
    <Card title='内部管理信息' size='small' style={{ marginBottom: 16 }}>
      <Row gutter={[12, 12]}>
        <Col span={12}>
          <Form.Item
            label='主办律师'
            name='leadLawyer'
            rules={[{ required: true, message: '请选择主办律师' }]}
          >
            <Select
              placeholder={dataLoading ? '正在加载...' : '选择主办律师'}
              loading={dataLoading}
              notFoundContent={dataLoading ? '正在加载...' : '暂无数据'}
            >
              {lawyers.map((lawyer) => (
                <Option key={lawyer.id} value={lawyer.id}>
                  <Space>
                    <UserOutlined />
                    {lawyer.name}
                    <Tag color='orange'>律师</Tag>
                  </Space>
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>

        <Col span={12}>
          <Form.Item label='协办律师' name='assistingLawyers'>
            <Select
              mode='multiple'
              placeholder={dataLoading ? '正在加载...' : '选择协办律师（可选）'}
              loading={dataLoading}
              notFoundContent={dataLoading ? '正在加载...' : '暂无数据'}
            >
              {lawyers.map((lawyer) => (
                <Option key={lawyer.id} value={lawyer.id}>
                  {lawyer.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>

        <Col span={12}>
          <Form.Item
            label='收费方式'
            name='billingMethod'
            rules={[{ required: true, message: '请选择收费方式' }]}
          >
            <Select placeholder='选择收费方式'>
              <Option value='FIXED'>定额收费</Option>
              <Option value='HOURLY'>按时收费</Option>
              <Option value='CONTINGENCY'>风险代理</Option>
              <Option value='MIXED'>混合收费</Option>
            </Select>
          </Form.Item>
        </Col>

        <Col span={12}>
          <Form.Item
            label='合同金额'
            name='contractAmount'
            rules={[
              { required: true, message: '请输入合同金额' },
              { type: 'number', min: 0, message: '金额必须大于0' },
            ]}
          >
            <InputNumber
              style={{ width: '100%' }}
              placeholder='请输入金额'
              formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
              parser={(value) => value!.replace(/\$\s?|(,*)/g, '')}
            />
          </Form.Item>
        </Col>

        <Col span={24}>
          <Form.Item label='预估工期' name='estimatedDuration'>
            <InputNumber
              min={1}
              max={60}
              placeholder='预估工期'
              addonAfter='个月'
              style={{ width: '200px' }}
            />
          </Form.Item>
        </Col>
      </Row>
    </Card>
  )

  /**
   * 渲染合规风控步骤
   */
  const renderComplianceInfo = () => (
    <Card title='合规与风险控制' size='small' style={{ marginBottom: 16 }}>
      <Alert
        message='风险控制提醒'
        description='请认真进行利益冲突检查和风险评估，这是律师执业的基本要求。'
        type='warning'
        showIcon
        style={{ marginBottom: 16 }}
        size='small'
      />

      <Row gutter={[12, 12]}>
        <Col span={24}>
          <Form.Item
            label='利益冲突检索'
            name='conflictCheck'
            rules={[{ required: true, message: '请完成利益冲突检查' }]}
          >
            <Space direction='vertical' style={{ width: '100%' }}>
              <Button
                type='primary'
                icon={<SearchOutlined />}
                loading={loading}
                onClick={() => {
                  const clientId = form.getFieldValue('clientId')
                  const otherParties = form.getFieldValue('otherParties') || []
                  if (clientId) {
                    onConflictCheck(clientId, otherParties)
                  } else {
                    message.warning('请先选择委托人')
                  }
                }}
                size='large'
              >
                {loading ? '正在执行冲突检索...' : '执行冲突检索'}
              </Button>

              <Radio.Group>
                <Radio value='NO_CONFLICT'>无冲突（已完成检索）</Radio>
                <Radio value='CONFLICT_RESOLVED'>有冲突但已解决</Radio>
                <Radio value='MANUAL_CHECK'>手动检索确认</Radio>
              </Radio.Group>
            </Space>
          </Form.Item>
        </Col>

        <Col span={24}>
          <Form.Item label='风险标签' name='riskTags'>
            <Checkbox.Group style={{ width: '100%' }}>
              <Row>
                {riskCategories.map((category) => (
                  <Col span={12} key={category.value} style={{ marginBottom: 8 }}>
                    <Checkbox value={category.value}>
                      <Space direction='vertical' size={0}>
                        <Text strong style={{ fontSize: '13px' }}>
                          {category.label}
                        </Text>
                        <Text type='secondary' style={{ fontSize: '11px' }}>
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
          <Form.Item name='isHighRisk' valuePropName='checked'>
            <Checkbox>
              <Space>
                <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
                <Text strong>标记为重大风险项目</Text>
                <Text type='secondary' style={{ fontSize: '12px' }}>
                  （需要合伙人审批）
                </Text>
              </Space>
            </Checkbox>
          </Form.Item>
        </Col>
      </Row>
    </Card>
  )

  /**
   * 渲染文档管理步骤
   */
  const renderDocumentsInfo = () => (
    <Card title='文档管理' size='small' style={{ marginBottom: 16 }}>
      <Alert
        message='文档上传要求'
        description='请上传必要的案件文档，单个文件大小不超过10MB，支持PDF、DOC、DOCX格式。'
        type='info'
        showIcon
        style={{ marginBottom: 16 }}
        size='small'
      />

      <Row gutter={[12, 12]}>
        <Col span={12}>
          <Form.Item label='委托代理合同' name='contractDocument'>
            <Upload
              accept='.pdf,.doc,.docx'
              beforeUpload={(file) => {
                const isValidSize = file.size / 1024 / 1024 < 10
                if (!isValidSize) {
                  message.error('文件大小不能超过10MB!')
                }
                return isValidSize
              }}
            >
              <Button icon={<UploadOutlined />}>上传合同文档</Button>
            </Upload>
          </Form.Item>
        </Col>

        <Col span={12}>
          <Form.Item label='律师所函' name='retainerAgreement'>
            <Upload
              accept='.pdf,.doc,.docx'
              beforeUpload={(file) => {
                const isValidSize = file.size / 1024 / 1024 < 10
                if (!isValidSize) {
                  message.error('文件大小不能超过10MB!')
                }
                return isValidSize
              }}
            >
              <Button icon={<UploadOutlined />}>上传所函</Button>
            </Upload>
          </Form.Item>
        </Col>

        <Col span={24}>
          <Form.Item label='其他文档' name='otherDocuments'>
            <Upload
              multiple
              accept='.pdf,.doc,.docx,.jpg,.png'
              beforeUpload={(file) => {
                const isValidSize = file.size / 1024 / 1024 < 10
                if (!isValidSize) {
                  message.error('文件大小不能超过10MB!')
                }
                return isValidSize
              }}
            >
              <Button icon={<UploadOutlined />}>上传其他文档</Button>
            </Upload>
          </Form.Item>
        </Col>
      </Row>
    </Card>
  )

  /**
   * 渲染当前步骤内容
   */
  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return renderBasicInfo()
      case 1:
        return renderManagementInfo()
      case 2:
        return renderComplianceInfo()
      case 3:
        return renderDocumentsInfo()
      default:
        return null
    }
  }

  /**
   * 验证当前步骤
   */
  const validateCurrentStep = async () => {
    const fieldsToValidate: { [key: number]: string[] } = {
      0: ['caseName', 'clientId', 'caseType', 'causeOfAction', 'caseDescription'],
      1: ['leadLawyer', 'billingMethod', 'contractAmount'],
      2: ['conflictCheck'],
      3: [], // 文档上传为可选
    }

    try {
      if (fieldsToValidate[currentStep].length > 0) {
        await form.validateFields(fieldsToValidate[currentStep])
      }
      return true
    } catch (error) {
      console.error('步骤验证失败:', error)
      message.error('请确保必填字段已正确填写')
      return false
    }
  }

  /**
   * 处理下一步
   */
  const handleNext = async () => {
    const isValid = await validateCurrentStep()
    if (isValid && currentStep < CASE_STEPS.length - 1) {
      setCurrentStep(currentStep + 1)
    }
  }

  /**
   * 处理上一步
   */
  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1)
    }
  }

  /**
   * 处理提交
   */
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      onSubmit(values)
    } catch (error) {
      console.error('表单验证失败:', error)
      message.error('请确保所有必填字段都已正确填写')
    }
  }

  return (
    <div
      style={{
        maxHeight: '75vh',
        overflowY: 'auto',
        padding: '8px',
      }}
    >
      {/* 步骤指示器 */}
      {renderStepIndicator()}

      {/* 表单内容 */}
      <Form
        form={form}
        layout='vertical'
        size='middle'
        initialValues={{
          caseType: '',
          billingMethod: 'FIXED',
          conflictCheck: 'NO_CONFLICT',
          isHighRisk: false,
        }}
      >
        {renderStepContent()}
      </Form>

      {/* 操作按钮 */}
      <Divider style={{ margin: '16px 0' }} />

      <Row justify='space-between' align='middle'>
        <Col>{currentStep > 0 && <Button onClick={handlePrev}>上一步</Button>}</Col>
        <Col>
          <Space>
            <Button onClick={onCancel}>取消</Button>
            {currentStep < CASE_STEPS.length - 1 ? (
              <Button type='primary' onClick={handleNext} loading={loading}>
                下一步
              </Button>
            ) : (
              <Button
                type='primary'
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
  )
}

export default CompactCaseFormWrapper
