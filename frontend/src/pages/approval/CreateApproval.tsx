import React, { useState } from 'react'
import { useNavigate } from 'react-router'
import { Card, Form, Input, Select, Button, message, Radio, DatePicker } from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import { createApproval } from '@/services/approval'
import dayjs from 'dayjs'
import './CreateApproval.less'

const { TextArea } = Input
const { Option } = Select
const { RangePicker } = DatePicker

interface CreateApprovalFormValues {
  type: string
  title: string
  content: string
  urgency: 'normal' | 'urgent' | 'very_urgent'
  expectedDate?: dayjs.Dayjs
}

const CreateApproval: React.FC = () => {
  const navigate = useNavigate()
  const [form] = Form.useForm<CreateApprovalFormValues>()
  const [loading, setLoading] = useState<boolean>(false)

  const handleSubmit = async (values: CreateApprovalFormValues) => {
    try {
      setLoading(true)

      const approvalData = {
        type: values.type,
        title: values.title,
        content: values.content,
        applicant: '管理员',
        applicantId: 1,
        department: '综合管理部',
        urgency: values.urgency,
      }

      await createApproval(approvalData)
      message.success('创建成功')
      navigate('/approval')
    } catch (error) {
      console.error('Failed to create approval:', error)
      message.error('创建失败')
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
