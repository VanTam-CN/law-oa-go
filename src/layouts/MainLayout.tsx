import React, { useState, useEffect } from 'react'
import { Layout, Alert } from 'antd'
import { Outlet } from 'react-router'
import Header from '@/components/layout/Header'
import Sidebar from '@/components/layout/Sidebar'
import { useAppStore } from '@/stores/useAppStore'
import { Navigate } from 'react-router'

const { Content } = Layout

const MainLayout: React.FC = () => {
  const { user, isAuthenticated, isLoading } = useAppStore()
  const [collapsed, setCollapsed] = useState(false)
  const [sidebarWidth, setSidebarWidth] = useState(220)

  // 处理侧边栏宽度变化
  const handleSidebarWidthChange = (width: number) => {
    setSidebarWidth(width)
  }

  // 动态调整CSS变量来控制侧边栏宽度
  useEffect(() => {
    document.documentElement.style.setProperty('--sidebar-width', `${sidebarWidth}px`)
  }, [sidebarWidth])

  // 如果正在加载，显示加载状态
  if (isLoading) {
    return <div>加载中...</div>
  }

  // 检查开发模式
  const isDevMode = process.env.NODE_ENV === 'development'

  // 如果未登录，重定向到登录页
  if (!isAuthenticated || !user) {
    if (isDevMode) {
      console.log('🛠️ 开发者模式：用户未登录，但显示开发提示')
      return (
        <Layout style={{ minHeight: '100vh' }}>
          <Content style={{ padding: '24px' }}>
            <Alert
              message='开发者模式提示'
              description='当前处于开发模式且用户未登录。请先登录或设置token以访问律师管理页面。'
              type='info'
              showIcon
              style={{ marginBottom: '16px' }}
            />
            <Navigate to='/login' replace />
          </Content>
        </Layout>
      )
    }
    return <Navigate to='/login' replace />
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sidebar
        collapsed={collapsed}
        setCollapsed={setCollapsed}
        onWidthChange={handleSidebarWidthChange}
      />
      <Layout style={{ marginLeft: 0 }}>
        <Header />
        <Content
          style={{
            margin: '80px 16px 24px 16px',
            marginLeft: 'var(--sidebar-width, 220px)',
            padding: 24,
            background: '#fff',
            minHeight: 280,
            position: 'relative',
            zIndex: 1,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

export default MainLayout
