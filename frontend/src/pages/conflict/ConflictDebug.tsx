import React, { useState } from 'react'
import { Card, Button, Form, Input, Select, message, Spin, Alert, Descriptions, Space } from 'antd'
import { performConflictCheck } from '@/services/conflict'

const { Option } = Select
const { TextArea } = Input

const ConflictDebug: React.FC = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<any>(null)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (values: any) => {
    setLoading(true)
    setError(null)
    setResult(null)

    try {
      console.log('发送冲突检查请求:', values)

      const response = await performConflictCheck(values)
      console.log('收到冲突检查响应:', response)

      setResult(response)
      message.success('冲突检测完成')
    } catch (err: any) {
      console.error('冲突检测失败:', err)
      setError(err.message || '未知错误')
      message.error('冲突检测失败')
    } finally {
      setLoading(false)
    }
  }

  const testHealthCheck = async () => {
    try {
      const response = await fetch('/conflict/health')
      const data = await response.json()
      console.log('健康检查响应:', data)
      message.success(`健康检查完成: ${data.data?.status || 'unknown'}`)
    } catch (err: any) {
      console.error('健康检查失败:', err)
      message.error('健康检查失败')
    }
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px' }}>
      <Card title='利益冲突检测调试工具' style={{ marginBottom: '24px' }}>
        <Space style={{ marginBottom: '16px' }}>
          <Button onClick={testHealthCheck}>测试健康检查</Button>
          <Button
            onClick={() => {
              console.log('当前环境:', {
                origin: window.location.origin,
                pathname: window.location.pathname,
                href: window.location.href,
              })
              message.info('环境信息已输出到控制台')
            }}
          >
            输出环境信息
          </Button>
        </Space>
      </Card>

      <Card title='冲突检测测试' style={{ marginBottom: '24px' }}>
        <Form
          form={form}
          layout='vertical'
          onFinish={handleSubmit}
          initialValues={{
            client_name: '测试客户',
            project_name: '测试案件',
            project_type: 'civil',
            opposite_parties: '对方当事人',
            team_members: ['律师1', '律师2'],
            searchYears: 5,
            searchDepth: 'deep',
            includeCorporateRelations: true,
          }}
        >
          <Form.Item
            label='客户名称'
            name='client_name'
            rules={[{ required: true, message: '请输入客户名称' }]}
          >
            <Input placeholder='输入客户名称' />
          </Form.Item>

          <Form.Item
            label='案件名称'
            name='project_name'
            rules={[{ required: true, message: '请输入案件名称' }]}
          >
            <Input placeholder='输入案件名称' />
          </Form.Item>

          <Form.Item
            label='案件类型'
            name='project_type'
            rules={[{ required: true, message: '请选择案件类型' }]}
          >
            <Select placeholder='选择案件类型'>
              <Option value='civil'>民事案件</Option>
              <Option value='criminal'>刑事案件</Option>
              <Option value='commercial'>商事案件</Option>
              <Option value='administrative'>行政案件</Option>
            </Select>
          </Form.Item>

          <Form.Item label='对方当事人' name='opposite_parties'>
            <Input placeholder='输入对方当事人信息' />
          </Form.Item>

          <Form.Item label='团队成员' name='team_members'>
            <Select mode='tags' placeholder='输入团队成员'>
              <Option value='律师1'>律师1</Option>
              <Option value='律师2'>律师2</Option>
              <Option value='律师3'>律师3</Option>
            </Select>
          </Form.Item>

          <Form.Item label='案件描述' name='description'>
            <TextArea rows={4} placeholder='输入案件描述（可选）' />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type='primary' htmlType='submit' loading={loading}>
                执行冲突检测
              </Button>
              <Button
                onClick={() => {
                  form.resetFields()
                  setResult(null)
                  setError(null)
                }}
              >
                重置表单
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {loading && (
        <Card>
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <Spin size='large' />
            <div style={{ marginTop: '16px' }}>正在执行冲突检测...</div>
          </div>
        </Card>
      )}

      {error && (
        <Alert
          message='检测失败'
          description={error}
          type='error'
          showIcon
          style={{ marginBottom: '24px' }}
        />
      )}

      {result && (
        <Card title='检测结果'>
          <Descriptions bordered column={1}>
            <Descriptions.Item label='检测状态'>
              {result.has_conflict ? '发现冲突' : '无冲突'}
            </Descriptions.Item>
            <Descriptions.Item label='冲突等级'>{result.conflict_level}</Descriptions.Item>
            <Descriptions.Item label='冲突数量'>{result.conflicts?.length || 0}</Descriptions.Item>
          </Descriptions>

          {result.conflicts && result.conflicts.length > 0 && (
            <div style={{ marginTop: '16px' }}>
              <h4>冲突详情:</h4>
              <pre
                style={{
                  background: '#f5f5f5',
                  padding: '12px',
                  borderRadius: '4px',
                  overflow: 'auto',
                  maxHeight: '300px',
                }}
              >
                {JSON.stringify(result.conflicts, null, 2)}
              </pre>
            </div>
          )}

          <div style={{ marginTop: '16px' }}>
            <h4>完整响应:</h4>
            <pre
              style={{
                background: '#f5f5f5',
                padding: '12px',
                borderRadius: '4px',
                overflow: 'auto',
                maxHeight: '400px',
              }}
            >
              {JSON.stringify(result, null, 2)}
            </pre>
          </div>
        </Card>
      )}
    </div>
  )
}

export default ConflictDebug
