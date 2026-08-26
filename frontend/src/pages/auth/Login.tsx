import React from 'react'
import { Alert, Card, Form, Input, Button, Checkbox } from 'antd'
import { message } from '@/utils/messageHelper'
import {
  EyeInvisibleOutlined,
  EyeOutlined,
  LockOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router'
import { login, setToken } from '@/services/auth'
import { getCurrentUserPermissions, getCurrentUserRoles } from '@/services/role'
import { useAppStore } from '@/stores/useAppStore'
import {
  setPermissions as cachePermissions,
  setRoles as cacheRoles,
  setStoragePersistence,
} from '@/utils/storage'
import './Login.less'

interface LoginFormValues {
  account: string
  password: string
  remember: boolean
}

const normalizeLoginIdentifier = (value: string) => {
  return value.trim()
}

const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  const { login: appStoreLogin } = useAppStore()
  const [form] = Form.useForm<LoginFormValues>()
  const [loading, setLoading] = React.useState(false)
  const [loginError, setLoginError] = React.useState('')
  const [passwordVisible, setPasswordVisible] = React.useState(false)

  const onFinish = async (values: LoginFormValues) => {
    try {
      setLoading(true)
      setLoginError('')
      const response = await login({
        ...values,
        account: normalizeLoginIdentifier(values.account),
      })

      // 处理后端返回的token格式
      const token = response.token || response.data?.token
      const userData = response.user || response.data?.user

      if (token && userData) {
        setStoragePersistence(Boolean(values.remember))
        setToken(token)

        let roleCodes = [userData.role || 'user']
        let permissionCodes: string[] = []

        try {
          const [roles, permissions] = await Promise.all([
            getCurrentUserRoles(),
            getCurrentUserPermissions(),
          ])

          if (roles?.length) {
            cacheRoles(roles)
            roleCodes = Array.from(new Set([...roleCodes, ...roles.map((role) => role.code)]))
          }

          if (permissions?.length) {
            cachePermissions(permissions)
            permissionCodes = Array.from(new Set(permissions.map((permission) => permission.code)))
          }
        } catch (rbacError) {
          console.warn('加载当前用户RBAC权限失败，将使用登录角色作为降级权限:', rbacError)
        }

        // 构造用户对象，映射后端字段到前端需要的格式
        const user = {
          id: userData.id.toString(),
          username: userData.username || userData.email,
          realName: userData.name || userData.real_name || userData.username || userData.email,
          email: userData.email,
          roles: roleCodes,
          permissions: permissionCodes,
          phone: userData.phone,
          avatar: userData.avatar,
          isActive: userData.status === 'active',
          lastLoginAt: new Date().toISOString(),
          createdAt: userData.created_at,
        }

        appStoreLogin(user, token)
        message.success('登录成功')

        // 等待一下让状态更新完成
        setTimeout(() => {
          navigate('/')
        }, 500)
      } else {
        throw new Error('未获取到有效的登录凭证')
      }
    } catch (error: any) {
      console.error('Login failed:', error)
      if (error?.response?.status === 401) {
        message.error('账号或密码错误')
        setLoginError('账号、邮箱或密码不正确，请重新输入')
        return
      }
      const apiMessage =
        error.response?.data?.error?.details ||
        error.response?.data?.error?.message ||
        error.response?.data?.message
      setLoginError(apiMessage || '登录失败，请检查账号和密码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className='login-container'>
      <Card className='login-card' title='示例律师事务所OA登录'>
        {loginError && (
          <Alert
            type='error'
            showIcon
            role='alert'
            message='登录失败'
            description={loginError}
          />
        )}
        <Form
          form={form}
          name='login'
          initialValues={{ remember: false }}
          onFinish={onFinish}
          size='large'
        >
          <Form.Item name='account' rules={[{ required: true, message: '请输入账号或邮箱' }]}>
            <Input
              prefix={<UserOutlined />}
              placeholder='账号或邮箱'
              autoComplete='username'
              aria-label='账号或邮箱'
            />
          </Form.Item>
          <Form.Item name='password' rules={[{ required: true, message: '请输入密码' }]}>
            <Input
              type={passwordVisible ? 'text' : 'password'}
              prefix={<LockOutlined />}
              suffix={
                <button
                  type='button'
                  className='login-password-toggle'
                  aria-label={passwordVisible ? '隐藏密码' : '显示密码'}
                  aria-pressed={passwordVisible}
                  onClick={() => setPasswordVisible((visible) => !visible)}
                >
                  {passwordVisible ? <EyeOutlined /> : <EyeInvisibleOutlined />}
                </button>
              }
              placeholder='密码'
              autoComplete='current-password'
              aria-label='密码'
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
      <div className='login-footer'>© {new Date().getFullYear()} 示例律师事务所OA - 版权所有</div>
    </div>
  )
}

export default LoginPage
