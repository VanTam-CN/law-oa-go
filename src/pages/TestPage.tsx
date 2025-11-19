import React from 'react'
import { Card, Button, Space } from 'antd'
import { useNavigate } from 'react-router'

const TestPage: React.FC = () => {
  const navigate = useNavigate()

  return (
    <div style={{ padding: '24px' }}>
      <Card title='路由测试页面' style={{ marginBottom: '16px' }}>
        <p>如果你能看到这个页面，说明路由工作正常。</p>
        <Space>
          <Button type='primary' onClick={() => navigate('/dashboard')}>
            去仪表盘
          </Button>
          <Button onClick={() => navigate('/lawyer')}>去律师管理</Button>
          <Button onClick={() => navigate('/user')}>去用户管理</Button>
        </Space>
      </Card>

      <Card title='测试信息'>
        <p>当前时间: {new Date().toLocaleString()}</p>
        <p>测试页面加载成功</p>
      </Card>
    </div>
  )
}

export default TestPage
