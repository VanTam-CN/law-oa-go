import React from 'react'
import { Card, Form, Input, Button } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router'

const SimpleLogin: React.FC = () => {
  const navigate = useNavigate()

  const onFinish = (values: any) => {
    console.log('Login attempt:', values)
    // 简单重定向到测试页面
    navigate('/simple-test')
  }

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        backgroundColor: '#f0f2f5',
      }}
    >
      <Card title='简化登录页面' style={{ width: 400 }}>
        <Form onFinish={onFinish} layout='vertical'>
          <Form.Item name='email' label='邮箱' rules={[{ required: true, message: '请输入邮箱!' }]}>
            <Input
              prefix={<UserOutlined />}
              placeholder='请输入邮箱'
            />
          </Form.Item>

          <Form.Item
            name='password'
            label='密码'
            rules={[{ required: true, message: '请输入密码!' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder='请输入密码'
            />
          </Form.Item>

          <Form.Item>
            <Button type='primary' htmlType='submit' style={{ width: '100%' }}>
              登录
            </Button>
          </Form.Item>
        </Form>

        <div style={{ marginTop: 16, textAlign: 'center' }}>
          <Button type='link' onClick={() => navigate('/simple-test')}>
            跳转到测试页面
          </Button>
        </div>
      </Card>
    </div>
  )
}

export default SimpleLogin
