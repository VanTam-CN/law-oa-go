import React, { useState, useEffect, useCallback } from 'react'
import {
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  DatePicker,
  Button,
  Tabs,
  Upload,
  message,
  Card,
  Row,
  Col,
  Radio,
  Checkbox,
  Space,
  Divider,
  Switch,
} from 'antd'
import {
  UploadOutlined,
  FileTextOutlined,
  UserOutlined,
  TeamOutlined,
  SafetyOutlined,
  FolderOpenOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import type { UploadProps } from 'antd'
import { enhancedCaseAPI } from '@/services/enhancedCase'
import MultiClientSelector from './case/MultiClientSelector'
import dayjs from 'dayjs'

const { TextArea } = Input
const { Option } = Select

// 已选客户信息接口
interface SelectedClient {
  clientId: string
  clientInfo: {
    id: string
    name: string
    type: 'INDIVIDUAL' | 'COMPANY'
    category: string
    registrationNumber?: string
    contactInfo?: string
  }
  role: string
  relationshipDescription?: string
  contactInfo?: string
  conflictCheckConfig?: {
    enabled: boolean
    checkOnCreate: boolean
    searchYears?: number
    includeCorporateRelations?: boolean
    searchDepth?: 'STANDARD' | 'DEEP' | 'COMPREHENSIVE'
    autoWaiverIfPossible?: boolean
  }
}

interface EnhancedCreateCaseModalProps {
  visible: boolean
  onCancel: () => void
  onSuccess: () => void
}

const EnhancedCreateCaseModal: React.FC<EnhancedCreateCaseModalProps> = ({
  visible,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [lawyers, setLawyers] = useState<any[]>([])
  const [caseTypes, setCaseTypes] = useState<any[]>([])
  const [projectTypes, setProjectTypes] = useState<any[]>([])
  const [billingMethods, setBillingMethods] = useState<any[]>([])
  const [activeTab, setActiveTab] = useState('basic')
  const [selectedClients, setSelectedClients] = useState<SelectedClient[]>([])

  useEffect(() => {
    if (visible) {
      loadInitialData()
      form.resetFields()
      setSelectedClients([])
    }
  }, [visible])

  const loadInitialData = async () => {
    try {
      // 模拟律师数据 - 实际项目中替换为真实API
      const mockLawyers = [
        { lawyerId: 1, lawyerName: '张律师', position: '高级合伙人' },
        { lawyerId: 2, lawyerName: '李律师', position: '合伙人' },
        { lawyerId: 3, lawyerName: '王律师', position: '主办律师' },
      ]
      setLawyers(mockLawyers)

      // 设置案件类型数据
      setCaseTypes([
        { id: 'CIVIL', name: '民事案件', code: 'CIVIL' },
        { id: 'COMMERCIAL', name: '商事案件', code: 'COMMERCIAL' },
        { id: 'CRIMINAL', name: '刑事案件', code: 'CRIMINAL' },
        { id: 'ADMINISTRATIVE', name: '行政案件', code: 'ADMINISTRATIVE' },
      ])

      // 设置项目类型数据
      setProjectTypes([
        { id: 'CASE', name: '诉讼案件', code: 'CASE' },
        { id: 'ADVISORY', name: '法律顾问', code: 'ADVISORY' },
        { id: 'REVIEW', name: '合同审查', code: 'REVIEW' },
        { id: 'CIVIL', name: '民事诉讼', code: 'CIVIL' },
        { id: 'COMMERCIAL', name: '商业诉讼', code: 'COMMERCIAL' },
        { id: 'CRIMINAL', name: '刑事诉讼', code: 'CRIMINAL' },
      ])

      // 设置收费方式数据
      setBillingMethods([
        { id: 'FIXED', name: '定额收费', code: 'FIXED' },
        { id: 'HOURLY', name: '按时收费', code: 'HOURLY' },
        { id: 'RISK', name: '风险代理', code: 'RISK' },
        { id: 'MIXED', name: '混合收费', code: 'MIXED' },
      ])
    } catch (error) {
      message.error('加载初始数据失败')
    }
  }

  const handleSubmit = async () => {
    try {
      // 验证是否选择了客户
      if (selectedClients.length === 0) {
        message.error('请至少选择一个客户')
        return
      }

      // 验证是否有主要委托人
      const hasPrimaryClient = selectedClients.some((client) => client.role === 'PRIMARY')
      if (!hasPrimaryClient) {
        message.error('必须指定一个主要委托人')
        return
      }

      const values = await form.validateFields()
      setLoading(true)

      // 构建客户角色映射
      const clientRoles: Record<string, any> = {}
      selectedClients.forEach((client) => {
        clientRoles[client.clientId] = {
          role: client.role,
          relationshipDescription: client.relationshipDescription,
          contactInfo: client.contactInfo,
        }
      })

      // 构建增强案例创建请求
      const caseData = {
        // 基础字段
        title: values.title,
        description: values.description || '',
        caseType: values.caseType,
        priority: values.priority || 'MEDIUM',
        startDate: values.startDate ? values.startDate.format('YYYY-MM-DD') : undefined,
        practiceArea: values.practiceArea || 'GENERAL',
        estimatedDuration: values.estimatedDuration,
        billingMethod: values.billingMethod,

        // 客户信息
        clientProfileIds: selectedClients.map((client) => client.clientId),
        clientRoles,

        // 团队分配
        lawyerId: values.lawyerId,
        assistingLawyerId: values.assistingLawyerId,

        // 冲突检测配置
        conflictCheckConfig: {
          enabled: true,
          checkOnCreate: true,
          searchYears: values.searchYears || 5,
          includeCorporateRelations: values.includeCorporateRelations || false,
          searchDepth: values.searchDepth || 'STANDARD',
          autoWaiverIfPossible: values.autoWaiverIfPossible || false,
        },

        // 分配信息
        assignedBy: 1, // 当前用户ID，实际项目中从用户上下文获取
        isMajorRisk: values.isMajorRisk || false,
      }

      await enhancedCaseAPI.createEnhancedCase(caseData)
      message.success('增强案例创建成功')
      onSuccess()
      onCancel()
    } catch (error) {
      console.error('创建增强案例失败:', error)
      message.error('创建案件失败')
    } finally {
      setLoading(false)
    }
  }

  const uploadProps: UploadProps = {
    name: 'file',
    action: '/api/lawfirm/file/upload',
    headers: {
      Authorization: `Bearer ${localStorage.getItem('token')}`,
    },
    onChange(info) {
      if (info.file.status === 'done') {
        message.success(`${info.file.name} 文件上传成功`)
      } else if (info.file.status === 'error') {
        message.error(`${info.file.name} 文件上传失败`)
      }
    },
  }

  // 案件基本信息模块
  const renderBasicInfo = () => (
    <Card title='基本信息' size='small'>
      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            label='案件名称'
            name='title'
            rules={[{ required: true, message: '请输入案件名称' }]}
          >
            <Input placeholder='请输入案件名称' />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item
            label='案件类型'
            name='caseType'
            rules={[{ required: true, message: '请选择案件类型' }]}
          >
            <Select placeholder='请选择案件类型'>
              {caseTypes?.map((type) => (
                <Option key={type.id} value={type.code}>
                  {type.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item label='优先级' name='priority' initialValue='MEDIUM'>
            <Select placeholder='请选择优先级'>
              <Option value='HIGH'>高</Option>
              <Option value='MEDIUM'>中</Option>
              <Option value='LOW'>低</Option>
            </Select>
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label='业务领域' name='practiceArea' initialValue='GENERAL'>
            <Select placeholder='请选择业务领域'>
              <Option value='GENERAL'>一般法律业务</Option>
              <Option value='CORPORATE'>公司法务</Option>
              <Option value='LITIGATION'>诉讼业务</Option>
              <Option value='COMPLIANCE'>合规业务</Option>
            </Select>
          </Form.Item>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            label='收费方式'
            name='billingMethod'
            rules={[{ required: true, message: '请选择收费方式' }]}
          >
            <Select placeholder='请选择收费方式'>
              {billingMethods?.map((method) => (
                <Option key={method.id} value={method.code}>
                  {method.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label='预计持续时间' name='estimatedDuration'>
            <Input placeholder='如：3个月' />
          </Form.Item>
        </Col>
      </Row>

      <Form.Item label='案件描述' name='description'>
        <TextArea rows={3} placeholder='请输入案件的详细描述' />
      </Form.Item>

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item label='开始日期' name='startDate'>
            <DatePicker style={{ width: '100%' }} placeholder='请选择开始日期' />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label='对方当事人信息' name='opponentInfo'>
            <TextArea
              rows={3}
              placeholder='请输入对方当事人详细信息&#10;姓名/公司名称：&#10;联系方式：&#10;地址：'
            />
          </Form.Item>
        </Col>
      </Row>
    </Card>
  )

  // 客户选择模块 - 使用新的多客户选择组件
  const renderClientSelection = () => (
    <Card title='客户选择与配置' size='small'>
      <MultiClientSelector
        value={selectedClients}
        onChange={setSelectedClients}
        form={form}
        maxClients={10}
        showConflictConfig
        allowPrimaryOnly={false}
      />

      <Divider />

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item label='冲突检测年限' name='searchYears' initialValue={5}>
            <InputNumber min={1} max={20} placeholder='搜索年限' style={{ width: '100%' }} />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label='搜索深度' name='searchDepth' initialValue='STANDARD'>
            <Select placeholder='选择搜索深度'>
              <Option value='STANDARD'>标准</Option>
              <Option value='DEEP'>深度</Option>
              <Option value='COMPREHENSIVE'>全面</Option>
            </Select>
          </Form.Item>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item name='includeCorporateRelations' valuePropName='checked' initialValue={false}>
            <Checkbox>包含企业关联关系</Checkbox>
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item name='autoWaiverIfPossible' valuePropName='checked' initialValue>
            <Checkbox>自动豁免（如可能）</Checkbox>
          </Form.Item>
        </Col>
      </Row>
    </Card>
  )

  // 内部管理信息模块
  const renderInternalManagement = () => (
    <Card title='内部管理' size='small'>
      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            label='主办律师'
            name='lawyerId'
            rules={[{ required: true, message: '请选择主办律师' }]}
          >
            <Select placeholder='请选择主办律师'>
              {lawyers?.map((lawyer) => (
                <Option key={lawyer.lawyerId} value={lawyer.lawyerId}>
                  {lawyer.lawyerName} - {lawyer.position}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label='协办律师' name='assistingLawyerId'>
            <Select placeholder='请选择协办律师' allowClear>
              {lawyers?.map((lawyer) => (
                <Option key={lawyer.lawyerId} value={lawyer.lawyerId}>
                  {lawyer.lawyerName} - {lawyer.position}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
      </Row>

      <Form.Item label='团队成员' name='teamMembers'>
        <TextArea rows={2} placeholder='请输入其他团队成员信息' />
      </Form.Item>

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item label='合同金额' name='contractAmount'>
            <InputNumber
              style={{ width: '100%' }}
              placeholder='请输入合同金额'
              addonAfter='元'
              precision={2}
            />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label='风险标记'>
            <Row gutter={8}>
              <Col span={8}>
                <Form.Item name='isMajorRisk' valuePropName='checked' noStyle>
                  <Checkbox>重大风险</Checkbox>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name='isMassCase' valuePropName='checked' noStyle>
                  <Checkbox>群体性案件</Checkbox>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name='isSensitiveCase' valuePropName='checked' noStyle>
                  <Checkbox>敏感案件</Checkbox>
                </Form.Item>
              </Col>
            </Row>
          </Form.Item>
        </Col>
      </Row>
    </Card>
  )

  // 文档上传模块
  const renderDocuments = () => (
    <Card title='文档上传' size='small'>
      <Row gutter={16}>
        <Col span={12}>
          <Card title='委托书文档' size='small'>
            <Form.Item name='powerOfAttorneyDocument'>
              <Upload {...uploadProps}>
                <Button icon={<UploadOutlined />}>上传委托书</Button>
              </Upload>
            </Form.Item>
            <div style={{ color: '#666', fontSize: '12px' }}>
              支持格式：PDF、DOC、DOCX，最大50MB
            </div>
          </Card>
        </Col>
        <Col span={12}>
          <Card title='所函文档' size='small'>
            <Form.Item name='legalLetterDocument'>
              <Upload {...uploadProps}>
                <Button icon={<UploadOutlined />}>上传所函文档</Button>
              </Upload>
            </Form.Item>
            <div style={{ color: '#666', fontSize: '12px' }}>
              支持格式：PDF、DOC、DOCX，最大50MB
            </div>
          </Card>
        </Col>
      </Row>

      <Card title='其他文档' size='small' style={{ marginTop: 16 }}>
        <Form.Item name='otherDocuments'>
          <Upload {...uploadProps} multiple>
            <Button icon={<UploadOutlined />}>上传其他文档</Button>
          </Upload>
        </Form.Item>
        <div style={{ color: '#666', fontSize: '12px' }}>
          可上传多个相关文档，包括证据材料、法律意见书等
        </div>
      </Card>

      <Card title='文档清单' size='small' style={{ marginTop: 16 }}>
        <div style={{ color: '#666', fontSize: '12px' }}>
          已上传文档将在此处显示，支持在线预览和下载
        </div>
      </Card>
    </Card>
  )

  return (
    <Modal
      title='新建增强案件'
      open={visible}
      onCancel={onCancel}
      width={1200}
      footer={[
        <Button key='cancel' onClick={onCancel}>
          取消
        </Button>,
        <Button key='submit' type='primary' loading={loading} onClick={handleSubmit}>
          创建增强案件
        </Button>,
      ]}
    >
      <Form
        form={form}
        layout='vertical'
        initialValues={{
          priority: 'MEDIUM',
          practiceArea: 'GENERAL',
          billingMethod: 'FIXED',
          searchYears: 5,
          searchDepth: 'STANDARD',
          includeCorporateRelations: false,
          autoWaiverIfPossible: true,
          isMajorRisk: false,
          isMassCase: false,
          isSensitiveCase: false,
        }}
      >
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'basic',
              label: (
                <span>
                  <FileTextOutlined />
                  基本信息
                </span>
              ),
              children: renderBasicInfo(),
            },
            {
              key: 'clients',
              label: (
                <span>
                  <UserOutlined />
                  客户选择
                </span>
              ),
              children: renderClientSelection(),
            },
            {
              key: 'internal',
              label: (
                <span>
                  <TeamOutlined />
                  内部管理
                </span>
              ),
              children: renderInternalManagement(),
            },
            {
              key: 'documents',
              label: (
                <span>
                  <FolderOpenOutlined />
                  文档上传
                </span>
              ),
              children: renderDocuments(),
            },
          ]}
        />
      </Form>
    </Modal>
  )
}

export default EnhancedCreateCaseModal
