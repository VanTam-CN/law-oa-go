import React, { useState } from 'react'
import { Card, Form, Input, Button, message, Tabs } from 'antd'
import { UserOutlined, LockOutlined, EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons'
import { login } from '@/services/auth'
import { useAuth } from '@/context/AuthContext'
import { setToken, setUserInfo } from '@/utils/storage'

interface LoginFormValues {
  account: string
  password: string
}

const TestLogin: React.FC = () => {
  const { login: authLogin } = useAuth()
  const [loading, setLoading] = useState(false)
  const [loginResponse, setLoginResponse] = useState<any>(null)
  const [activeTab, setActiveTab] = useState<'login' | 'debug'>('login')

  const handleLogin = async (values: LoginFormValues) => {
    try {
      setLoading(true)
      setLoginResponse(null)

      console.log('开始登录，参数:', values)

      const response = await login(values)
      console.log('登录响应:', response)

      setLoginResponse(response)

      // 处理后端返回的token格式
      const token = response.token || response.data?.token
      const userData = response.user || response.data?.user

      if (token && userData) {
        console.log('获取到token和用户信息:', {
          token: `${token.substring(0, 20)}...`,
          user: userData,
        })

        // 构造用户对象，映射后端字段到前端需要的格式
        const user = {
          id: userData.id,
          username: userData.username || userData.email,
          real_name: userData.name,
          email: userData.email,
          role: userData.role,
          phone: userData.phone,
          avatar: userData.avatar,
          status: userData.status,
          department: userData.department || '',
          created_at: userData.created_at,
        }

        console.log('处理后的用户信息:', user)

        authLogin(token, user)
        message.success('登录成功！')
      } else {
        console.error('Token或用户信息缺失:', { token, user: userData })
        throw new Error('未获取到有效的登录凭证')
      }
    } catch (error: any) {
      console.error('登录失败:', error)
      const errorMsg = error.response?.data?.message || error.message || '登录失败'
      message.error(errorMsg)
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user_info')
    localStorage.removeItem('roles')
    localStorage.removeItem('permissions')
    window.location.reload()
  }

  const TestAPI = () => {
    const token = localStorage.getItem('auth_token')
    const user = localStorage.getItem('user_info')

    return (
      <div style={{ marginTop: 20 }}>
        <h3>API调试信息</h3>
        <p>
          <strong>Token:</strong> {token ? `${token.substring(0, 20)}...` : '未设置'}
        </p>
        <p>
          <strong>用户:</strong> {user || '未设置'}
        </p>

        <Button type='primary' onClick={handleLogout} style={{ marginTop: 10 }}>
          清除登录状态
        </Button>

        <Button
          type='default'
          onClick={() => {
            fetch('http://localhost:8080/api/v1/dashboard/statistics', {
              headers: token ? { Authorization: `Bearer ${token}` } : {},
            })
              .then((res) => res.json())
              .then((data) => {
                console.log('API测试响应:', data)
                message.success('API调用成功')
              })
              .catch((error) => {
                console.error('API调用失败:', error)
                message.error('API调用失败')
              })
          }}
          style={{ marginTop: 10 }}
        >
          测试API调用
        </Button>
      </div>
    )
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        padding: 20,
      }}
    >
      <Card
        style={{
          width: 400,
          padding: '30px',
          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.1)',
          borderRadius: '12px',
        }}
      >
        <div style={{ textAlign: 'center', marginBottom: 30 }}>
          <h1 style={{ color: '#2c3e50', marginBottom: 10 }}>示例律师事务所OA</h1>
          <p style={{ color: '#7f8c8d', margin: 0 }}>测试登录页面</p>
        </div>

        <Tabs activeKey={activeTab} onChange={(key) => setActiveTab(key as 'login' | 'debug')}>
          <Tabs.TabPane tab='用户登录' key='login'>
            <Form
              name='login'
              onFinish={handleLogin}
              layout='vertical'
              size='large'
            >
              <Form.Item name='account' rules={[{ required: true, message: '请输入账号或邮箱' }]}>
                <Input
                  prefix={<UserOutlined />}
                  placeholder='账号或邮箱'
                  autoComplete='username'
                  allowClear
                />
              </Form.Item>

              <Form.Item name='password' rules={[{ required: true, message: '请输入密码' }]}>
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder='密码'
                  autoComplete='current-password'
                  allowClear
                />
              </Form.Item>

              <Form.Item>
                <Button
                  type='primary'
                  htmlType='submit'
                  block
                  loading={loading}
                  style={{ height: '48px' }}
                >
                  {loading ? '登录中...' : '登录'}
                </Button>
              </Form.Item>
            </Form>

          </Tabs.TabPane>

          <Tabs.TabPane tab='调试信息' key='debug'>
            <TestAPI />
            {loginResponse && (
              <div style={{ marginTop: 20 }}>
                <h4>登录响应数据:</h4>
                <pre
                  style={{
                    background: '#f5f5f5',
                    padding: '10px',
                    borderRadius: '4px',
                    fontSize: '12px',
                    maxHeight: '300px',
                    overflow: 'auto',
                  }}
                >
                  {JSON.stringify(loginResponse, null, 2)}
                </pre>
              </div>
            )}
          </Tabs.TabPane>
        </Tabs>
      </Card>
    </div>
  )
}

export default TestLogin
