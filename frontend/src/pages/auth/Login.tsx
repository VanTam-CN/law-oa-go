import React from 'react'
import { Card, Form, Input, Button, Checkbox } from 'antd'
import { message } from '@/utils/messageHelper'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router'
import { login, setToken } from '@/services/auth'
import { useAppStore } from '@/stores/useAppStore'
import './Login.less'

interface LoginFormValues {
  email: string
  password: string
  remember: boolean
}

const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  const { login: appStoreLogin } = useAppStore()
  const [form] = Form.useForm<LoginFormValues>()
  const [loading, setLoading] = React.useState(false)

  const onFinish = async (values: LoginFormValues) => {
    try {
      setLoading(true)
      const response = await login(values)

      console.log('Login response:', response)

      // 处理后端返回的token格式
      const token = response.token || response.data?.token
      const userData = response.user || response.data?.user

      if (token && userData) {
        // 构造用户对象，映射后端字段到前端需要的格式
        const user = {
          id: userData.id.toString(),
          username: userData.email, // 使用email作为username
          realName: userData.name,
          email: userData.email,
          roles: [userData.role], // useAppStore期望的roles是数组
          permissions: [], // 可以根据角色设置默认权限
          phone: userData.phone,
          avatar: userData.avatar,
          isActive: userData.status === 'active',
          lastLoginAt: new Date().toISOString(),
          createdAt: userData.created_at,
        }

        console.log('Processed user:', user)

        // 设置token到storage
        setToken(token)
        appStoreLogin(user, token)
        message.success('登录成功')

        // 等待一下让状态更新完成
        setTimeout(() => {
          navigate('/')
        }, 500)
      } else {
        console.error('Token or user missing:', { token, user: userData })
        throw new Error('未获取到有效的登录凭证')
      }
    } catch (error: any) {
      console.error('Login failed:', error)
      message.error(error.response?.data?.message || '登录失败，请检查用户名和密码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className='login-container'>
      <Card className='login-card' title='律所OA系统登录'>
        <Form
          form={form}
          name='login'
          initialValues={{ remember: true }}
          onFinish={onFinish}
          size='large'
        >
          <Form.Item name='email' rules={[{ required: true, message: '请输入邮箱' }]}>
            <Input prefix={<UserOutlined />} placeholder='邮箱' autoComplete='email' />
          </Form.Item>
          <Form.Item name='password' rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password
              prefix={<LockOutlined />}
              placeholder='密码'
              autoComplete='current-password'
            />
          </Form.Item>
          <Form.Item name='remember' valuePropName='checked'>
            <Checkbox>记住我</Checkbox>
          </Form.Item>
          <Form.Item>
            <Button type='primary' htmlType='submit' block loading={loading}>
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
      <div className='login-footer'>© {new Date().getFullYear()} 律所OA系统 - 版权所有</div>
    </div>
  )
}

export default LoginPage
