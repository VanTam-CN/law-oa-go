import React, { useState, useEffect } from 'react'
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
} from 'antd'
import {
  UploadOutlined,
  FileTextOutlined,
  UserOutlined,
  TeamOutlined,
  SafetyOutlined,
  FolderOpenOutlined,
} from '@ant-design/icons'
import type { UploadProps } from 'antd'
import { CaseInfo, ClientInfo, LawyerInfo, caseAPI, clientAPI, lawyerAPI } from '@/services/lawfirm'
import dayjs from 'dayjs'

const { TextArea } = Input
const { TabPane } = Tabs
const { Option } = Select

interface CreateCaseModalProps {
  visible: boolean
  onCancel: () => void
  onSuccess: () => void
}

const CreateCaseModal: React.FC<CreateCaseModalProps> = ({ visible, onCancel, onSuccess }) => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [clients, setClients] = useState<ClientInfo[]>([])
  const [lawyers, setLawyers] = useState<LawyerInfo[]>([])
  const [caseTypes, setCaseTypes] = useState<any[]>([])
  const [projectTypes, setProjectTypes] = useState<any[]>([])
  const [billingMethods, setBillingMethods] = useState<any[]>([])
  const [activeTab, setActiveTab] = useState('basic')

  useEffect(() => {
    if (visible) {
      loadInitialData()
      form.resetFields()
    }
  }, [visible])

  const loadInitialData = async () => {
    try {
      const [clientsResponse, lawyersResponse] = await Promise.all([
        clientAPI.getList(),
        lawyerAPI.getList(),
      ])

      setClients(clientsResponse.rows || [])
      setLawyers(lawyersResponse.rows || [])

      // 设置案件类型数据（暂时使用硬编码）
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
      const values = await form.validateFields()
      setLoading(true)

      // 构建提交数据
      const caseData: CaseInfo = {
        ...values,
        startDate: values.startDate ? values.startDate.format('YYYY-MM-DD') : undefined,
        endDate: values.endDate ? values.endDate.format('YYYY-MM-DD') : undefined,
        status: '0', // 默认状态：未开始
        conflictCheckStatus: 'PENDING', // 默认冲突检查状态：待检查
        isMajorRisk: values.isMajorRisk || false,
        isMassCase: values.isMassCase || false,
        isSensitiveCase: values.isSensitiveCase || false,
      }

      await caseAPI.create(caseData)
      message.success('案件创建成功')
      onSuccess()
      onCancel()
    } catch (error) {
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
            name='caseName'
            rules={[{ required: true, message: '请输入案件名称' }]}
          >
            <Input placeholder='请输入案件名称' />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label='案件编号' name='caseNo'>
            <Input placeholder='自动生成' disabled />
          </Form.Item>
        </Col>
      </Row>

      <Row gutter={16}>
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
        <Col span={12}>
          <Form.Item
            label='项目类型'
            name='projectType'
            rules={[{ required: true, message: '请选择项目类型' }]}
          >
            <Select placeholder='请选择项目类型'>
              {projectTypes?.map((type) => (
                <Option key={type.id} value={type.code}>
                  {type.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>
      </Row>

      <Form.Item label='案由' name='causeOfAction'>
        <TextArea rows={2} placeholder='请输入案件的具体原因和背景' />
      </Form.Item>

      <Divider />

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item
            label='委托人信息'
            name='principalInfo'
            rules={[{ required: true, message: '请输入委托人信息' }]}
          >
            <TextArea
              rows={3}
              placeholder='请输入委托人详细信息&#10;姓名：&#10;联系方式：&#10;地址：'
            />
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

      <Form.Item label='案件描述' name='description'>
        <TextArea rows={3} placeholder='请输入案件的详细描述' />
      </Form.Item>
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
          <Form.Item label='合同金额' name='contractAmount'>
            <InputNumber
              style={{ width: '100%' }}
              placeholder='请输入合同金额'
              addonAfter='元'
              precision={2}
            />
          </Form.Item>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={12}>
          <Form.Item label='开始日期' name='startDate'>
            <DatePicker style={{ width: '100%' }} placeholder='请选择开始日期' />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label='预计结束日期' name='endDate'>
            <DatePicker style={{ width: '100%' }} placeholder='请选择预计结束日期' />
          </Form.Item>
        </Col>
      </Row>

      <Form.Item label='团队成员' name='teamMembers'>
        <TextArea rows={2} placeholder='请输入其他团队成员信息' />
      </Form.Item>
    </Card>
  )

  // 合规与风控模块
  const renderComplianceRisk = () => (
    <Card title='合规与风控' size='small'>
      <Row gutter={16}>
        <Col span={12}>
          <Form.Item label='利益冲突检查状态' name='conflictCheckStatus'>
            <Select placeholder='请选择冲突检查状态'>
              <Option value='PENDING'>待检查</Option>
              <Option value='PASSED'>通过</Option>
              <Option value='FAILED'>未通过</Option>
            </Select>
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
                  <Checkbox>敏感性案件</Checkbox>
                </Form.Item>
              </Col>
            </Row>
          </Form.Item>
        </Col>
      </Row>

      <Form.Item label='风险评估说明' name='riskAssessment'>
        <TextArea rows={3} placeholder='请输入风险评估的详细说明' />
      </Form.Item>

      <Card title='利益冲突检查' size='small' style={{ marginTop: 16 }}>
        <Space direction='vertical' style={{ width: '100%' }}>
          <Button type='primary' onClick={() => message.info('正在进行利益冲突检查...')}>
            执行利益冲突检查
          </Button>
          <div style={{ color: '#374151', fontSize: '13px' }}>
            系统将自动检查与现有案件的利益冲突情况，包括客户、律师、案件类型等多个维度
          </div>
        </Space>
      </Card>
    </Card>
  )

  // 文档管理模块
  const renderDocumentManagement = () => (
    <Card title='文档管理' size='small'>
      <Row gutter={16}>
        <Col span={12}>
          <Card title='合同文档' size='small'>
            <Form.Item name='contractDocument'>
              <Upload {...uploadProps}>
                <Button icon={<UploadOutlined />}>上传合同文档</Button>
              </Upload>
            </Form.Item>
            <div style={{ color: '#374151', fontSize: '13px' }}>
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
            <div style={{ color: '#374151', fontSize: '13px' }}>
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
        <div style={{ color: '#374151', fontSize: '13px' }}>
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
      title='新建案件'
      open={visible}
      onCancel={onCancel}
      width={1000}
      footer={[
        <Button key='cancel' onClick={onCancel}>
          取消
        </Button>,
        <Button key='submit' type='primary' loading={loading} onClick={handleSubmit}>
          创建案件
        </Button>,
      ]}
    >
      <Form
        form={form}
        layout='vertical'
        initialValues={{
          status: '0',
          conflictCheckStatus: 'PENDING',
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
              key: 'compliance',
              label: (
                <span>
                  <SafetyOutlined />
                  合规与风控
                </span>
              ),
              children: renderComplianceRisk(),
            },
            {
              key: 'documents',
              label: (
                <span>
                  <FolderOpenOutlined />
                  文档管理
                </span>
              ),
              children: renderDocumentManagement(),
            },
          ]}
        />
      </Form>
    </Modal>
  )
}

export default CreateCaseModal
