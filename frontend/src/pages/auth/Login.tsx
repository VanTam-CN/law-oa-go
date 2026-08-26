import React from 'react'
import { Card, Form, Input, Button, Checkbox } from 'antd'
import { message } from '@/utils/messageHelper'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
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
  return value.trim().toLowerCase()
}

const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  const { login: appStoreLogin } = useAppStore()
  const [form] = Form.useForm<LoginFormValues>()
  const [loading, setLoading] = React.useState(false)

  const onFinish = async (values: LoginFormValues) => {
    try {
      setLoading(true)
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
          username: userData.email, // 使用email作为username
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
        message.error('账号、邮箱或密码不正确，请重新输入')
        return
      }
      const apiMessage =
        error.response?.data?.error?.details ||
        error.response?.data?.error?.message ||
        error.response?.data?.message
      message.error(apiMessage || '登录失败，请检查账号和密码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className='login-container'>
      <Card className='login-card' title='示例律师事务所OA登录'>
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
            />
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
      <div className='login-footer'>© {new Date().getFullYear()} 示例律师事务所OA - 版权所有</div>
    </div>
  )
}

export default LoginPage
