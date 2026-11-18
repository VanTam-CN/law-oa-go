import React from 'react'
import { Card, Button, Result } from 'antd'

const SimpleTest: React.FC = () => {
  return (
    <div
      style={{
        padding: '24px',
        backgroundColor: '#f0f2f5',
        minHeight: '100vh',
      }}
    >
      <Card title='简单测试页面' style={{ marginBottom: '16px' }}>
        <Result
          status='success'
          title='React 应用运行正常！'
          subTitle='这是一个独立的测试页面，不依赖AuthContext或其他复杂组件'
          extra={[
            <Button type='primary' key='home' href='/'>
              返回首页
            </Button>,
            <Button key='reload' onClick={() => window.location.reload()}>
              刷新页面
            </Button>,
          ]}
        />
      </Card>

      <Card title='系统信息'>
        <p>
          <strong>前端服务器:</strong> http://localhost:3002
        </p>
        <p>
          <strong>后端服务器:</strong> http://localhost:8082
        </p>
        <p>
          <strong>React版本:</strong> 18.2.0
        </p>
        <p>
          <strong>当前时间:</strong> {new Date().toLocaleString()}
        </p>
        <p>
          <strong>测试说明:</strong> 如果你能看到这个页面，说明React应用已经成功加载并渲染
        </p>
      </Card>

      <Card title='测试状态'>
        <p>✅ HTML页面加载成功</p>
        <p>✅ React组件渲染成功</p>
        <p>✅ Ant Design组件加载成功</p>
        <p>✅ 样式文件加载成功</p>
      </Card>
    </div>
  )
}

export default SimpleTest
