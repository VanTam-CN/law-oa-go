import React, { useState } from 'react'
import { Card, Form, Switch, Button, message, Divider, Input, Select, Radio, Space } from 'antd'
import {
  SettingOutlined,
  BellOutlined,
  LockOutlined,
  BgColorsOutlined,
  GlobalOutlined,
} from '@ant-design/icons'

const { Item: FormItem } = Form
const { Option } = Select
const { TextArea } = Input

const Settings: React.FC = () => {
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)

  // 模拟设置数据
  const settingsData = {
    notifications: {
      email: true,
      browser: true,
      approval: true,
      system: true,
      finance: false,
    },
    appearance: {
      theme: 'light',
      language: 'zh-CN',
      fontSize: 'medium',
      compact: false,
    },
    privacy: {
      profileVisible: true,
      activityTracking: true,
      dataSharing: false,
    },
    account: {
      twoFactorAuth: false,
      loginNotifications: true,
      sessionTimeout: 30,
    },
  }

  const handleSave = async () => {
    try {
      setLoading(true)
      const values = await form.validateFields()
      console.log('保存设置:', values)
      message.success('设置保存成功！')
    } catch (error) {
      console.error('保存失败:', error)
      message.error('保存失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  const handleReset = () => {
    form.resetFields()
    message.info('设置已重置')
  }

  return (
    <div style={{ padding: '24px' }}>
      <Card title='系统设置' style={{ maxWidth: 1000, margin: '0 auto' }}>
        <Form form={form} layout='vertical' initialValues={settingsData}>
          {/* 通知设置 */}
          <Card
            title={
              <span>
                <BellOutlined style={{ marginRight: 8 }} />
                通知设置
              </span>
            }
            style={{ marginBottom: 16 }}
          >
            <FormItem label='邮件通知' name={['notifications', 'email']} valuePropName='checked'>
              <Switch />
            </FormItem>

            <FormItem
              label='浏览器通知'
              name={['notifications', 'browser']}
              valuePropName='checked'
            >
              <Switch />
            </FormItem>

            <FormItem label='审批通知' name={['notifications', 'approval']} valuePropName='checked'>
              <Switch />
            </FormItem>

            <FormItem label='系统通知' name={['notifications', 'system']} valuePropName='checked'>
              <Switch />
            </FormItem>

            <FormItem label='财务通知' name={['notifications', 'finance']} valuePropName='checked'>
              <Switch />
            </FormItem>
          </Card>

          {/* 外观设置 */}
          <Card
            title={
              <span>
                <BgColorsOutlined style={{ marginRight: 8 }} />
                外观设置
              </span>
            }
            style={{ marginBottom: 16 }}
          >
            <FormItem label='主题模式' name={['appearance', 'theme']}>
              <Radio.Group>
                <Radio value='light'>浅色主题</Radio>
                <Radio value='dark'>深色主题</Radio>
                <Radio value='auto'>跟随系统</Radio>
              </Radio.Group>
            </FormItem>

            <FormItem label='语言' name={['appearance', 'language']}>
              <Select style={{ width: 200 }}>
                <Option value='zh-CN'>简体中文</Option>
                <Option value='zh-TW'>繁体中文</Option>
                <Option value='en-US'>English</Option>
              </Select>
            </FormItem>

            <FormItem label='字体大小' name={['appearance', 'fontSize']}>
              <Select style={{ width: 200 }}>
                <Option value='small'>小</Option>
                <Option value='medium'>中</Option>
                <Option value='large'>大</Option>
              </Select>
            </FormItem>

            <FormItem label='紧凑模式' name={['appearance', 'compact']} valuePropName='checked'>
              <Switch />
            </FormItem>
          </Card>

          {/* 隐私设置 */}
          <Card
            title={
              <span>
                <LockOutlined style={{ marginRight: 8 }} />
                隐私设置
              </span>
            }
            style={{ marginBottom: 16 }}
          >
            <FormItem
              label='个人资料可见'
              name={['privacy', 'profileVisible']}
              valuePropName='checked'
            >
              <Switch />
            </FormItem>

            <FormItem
              label='活动追踪'
              name={['privacy', 'activityTracking']}
              valuePropName='checked'
            >
              <Switch />
            </FormItem>

            <FormItem label='数据共享' name={['privacy', 'dataSharing']} valuePropName='checked'>
              <Switch />
            </FormItem>
          </Card>

          {/* 账户设置 */}
          <Card
            title={
              <span>
                <GlobalOutlined style={{ marginRight: 8 }} />
                账户设置
              </span>
            }
          >
            <FormItem
              label='双因子认证'
              name={['account', 'twoFactorAuth']}
              valuePropName='checked'
            >
              <Switch />
            </FormItem>

            <FormItem
              label='登录通知'
              name={['account', 'loginNotifications']}
              valuePropName='checked'
            >
              <Switch />
            </FormItem>

            <FormItem label='会话超时（分钟）' name={['account', 'sessionTimeout']}>
              <Input type='number' min={5} max={480} style={{ width: 200 }} />
            </FormItem>
          </Card>

          <Divider />

          <FormItem>
            <Space>
              <Button
                type='primary'
                icon={<SettingOutlined />}
                onClick={handleSave}
                loading={loading}
              >
                保存设置
              </Button>
              <Button onClick={handleReset}>重置</Button>
            </Space>
          </FormItem>
        </Form>
      </Card>
    </div>
  )
}

export default Settings
