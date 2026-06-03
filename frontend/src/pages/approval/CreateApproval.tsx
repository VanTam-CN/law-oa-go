import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router'
import { Card, Form, Input, Select, Button, Radio, DatePicker, Space, Row, Col, Divider } from 'antd'
import { ArrowLeftOutlined, FileTextOutlined } from '@ant-design/icons'
import { createApproval, submitApproval, listApprovalTemplates, type ApprovalTemplate } from '@/services/approval'
import { getUserInfo } from '@/utils/storage'
import { message } from '@/utils/messageHelper'
import dayjs from 'dayjs'
import './CreateApproval.less'

const { TextArea } = Input
const { Option } = Select
const { RangePicker } = DatePicker

interface CreateApprovalFormValues {
  type: string
  templateName?: string
  title: string
  content: string
  urgency: 'normal' | 'urgent' | 'very_urgent'
  expectedDate?: dayjs.Dayjs
}

const CreateApproval: React.FC = () => {
  const navigate = useNavigate()
  const [form] = Form.useForm<CreateApprovalFormValues>()
  const [loading, setLoading] = useState<boolean>(false)
  const [templatesLoading, setTemplatesLoading] = useState<boolean>(false)
  const [currentUserId, setCurrentUserId] = useState<string>('')
  const [currentUserName, setCurrentUserName] = useState<string>('')
  const [currentUserDepartment, setCurrentUserDepartment] = useState<string>('')
  const [templates, setTemplates] = useState<ApprovalTemplate[]>([])
  const [selectedTemplate, setSelectedTemplate] = useState<ApprovalTemplate | null>(null)
  const [useTemplate, setUseTemplate] = useState<boolean>(false)

  // 获取当前用户信息
  useEffect(() => {
    const userInfo = getUserInfo()
    if (userInfo) {
      setCurrentUserId(userInfo.id?.toString() || '1')
      setCurrentUserName(userInfo.name || userInfo.username || '用户')
      setCurrentUserDepartment(userInfo.department || '综合管理部')
    } else {
      // 开发环境默认值
      setCurrentUserId('1')
      setCurrentUserName('系统管理员')
      setCurrentUserDepartment('综合管理部')
    }

    // 加载审批模板
    fetchTemplates()
  }, [])

  const fetchTemplates = async () => {
    try {
      setTemplatesLoading(true)
      const data = await listApprovalTemplates()
      setTemplates(data)
    } catch (error) {
      console.error('Failed to fetch templates:', error)
      // 静默失败，不影响主流程
    } finally {
      setTemplatesLoading(false)
    }
  }

  const handleTemplateChange = (templateName: string) => {
    const template = templates.find((t) => t.name === templateName)
    if (template) {
      setSelectedTemplate(template)
      // 自动填充表单
      form.setFieldsValue({
        title: template.display_name,
        content: template.description || '',
        type: template.category,
      })
    }
  }

  const handleSubmit = async (values: CreateApprovalFormValues) => {
    try {
      setLoading(true)

      if (useTemplate && selectedTemplate) {
        // 使用模板创建审批
        const formData = {
          ...values,
          templateName: selectedTemplate.name,
        }
        const createdApproval = await createApproval({
          type: selectedTemplate.category,
          title: values.title,
          content: values.content,
          applicant: currentUserName,
          applicantId: currentUserId,
          department: currentUserDepartment,
          urgency: values.urgency,
          metadata: {
            template_name: selectedTemplate.name,
            form_data: formData,
          },
        })

        const submittedApproval = await submitApproval(createdApproval.id)
        message.success('审批申请提交成功')
        navigate('/approval')
      } else {
        // 普通创建审批
        const approvalData = {
          type: values.type,
          title: values.title,
          content: values.content,
          applicant: currentUserName,
          applicantId: currentUserId,
          department: currentUserDepartment,
          urgency: values.urgency,
        }

        const createdApproval = await createApproval(approvalData)
        const submittedApproval = await submitApproval(createdApproval.id)
        message.success('审批申请提交成功')
        navigate('/approval')
      }
    } catch (error) {
      console.error('Failed to create/submit approval:', error)
      message.error('提交失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  const approvalTypes = [
    { value: '立项申请', label: '立项申请' },
    { value: '用章申请', label: '用章申请' },
    { value: '开票申请', label: '开票申请' },
    { value: '合同变更', label: '合同变更' },
    { value: '投标申请', label: '投标申请' },
    { value: '请假申请', label: '请假申请' },
    { value: '报销申请', label: '报销申请' },
    { value: '采购申请', label: '采购申请' },
  ]

  return (
    <div className='create-approval-container'>
      <Card>
        <div className='page-header'>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/approval')}
            className='back-button'
          >
            返回
          </Button>
          <h2>新建审批</h2>
        </div>

        <Form
          form={form}
          layout='vertical'
          onFinish={handleSubmit}
          initialValues={{
            urgency: 'normal',
          }}
        >
          {/* 模板选择 */}
          {templates.length > 0 && (
            <>
              <Divider orientation='left'>快捷方式</Divider>
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item label='使用模板'>
                    <Select
                      placeholder='选择审批模板（可选）'
                      allowClear
                      loading={templatesLoading}
                      onChange={(value) => {
                        if (value) {
                          setUseTemplate(true)
                          handleTemplateChange(value)
                        } else {
                          setUseTemplate(false)
                          setSelectedTemplate(null)
                        }
                      }}
                    >
                      {templates.map((template) => (
                        <Option key={template.name} value={template.name}>
                          <Space>
                            <FileTextOutlined />
                            {template.display_name}
                          </Space>
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>
                </Col>
                <Col span={12}>
                  {selectedTemplate && (
                    <div style={{ marginTop: 30, color: '#666', fontSize: '12px' }}>
                      {selectedTemplate.description}
                    </div>
                  )}
                </Col>
              </Row>
              <Divider />
            </>
          )}

          <Divider orientation='left'>基本信息</Divider>

          <Form.Item
            name='type'
            label='申请类型'
            rules={[{ required: true, message: '请选择申请类型' }]}
          >
            <Select placeholder='请选择申请类型'>
              {approvalTypes.map((type) => (
                <Option key={type.value} value={type.value}>
                  {type.label}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name='title'
            label='申请标题'
            rules={[
              { required: true, message: '请输入申请标题' },
              { max: 100, message: '标题最多100个字符' },
            ]}
          >
            <Input placeholder='请输入申请标题' />
          </Form.Item>

          <Form.Item
            name='content'
            label='申请内容'
            rules={[
              { required: true, message: '请输入申请内容' },
              { min: 10, message: '内容至少10个字符' },
            ]}
          >
            <TextArea rows={6} placeholder='请详细描述申请内容、原因等信息' />
          </Form.Item>

          <Form.Item
            name='urgency'
            label='紧急程度'
            rules={[{ required: true, message: '请选择紧急程度' }]}
          >
            <Radio.Group>
              <Radio value='normal'>普通</Radio>
              <Radio value='urgent'>紧急</Radio>
              <Radio value='very_urgent'>特急</Radio>
            </Radio.Group>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type='primary' htmlType='submit' loading={loading}>
                提交申请
              </Button>
              <Button onClick={() => navigate('/approval')}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export default CreateApproval
