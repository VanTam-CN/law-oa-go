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
  const [compactViewport, setCompactViewport] = useState(false)
  const [mobileViewport, setMobileViewport] = useState(false)

  // 处理侧边栏宽度变化
  const handleSidebarWidthChange = (width: number) => {
    setSidebarWidth(width)
  }

  // 动态调整CSS变量来控制侧边栏宽度
  useEffect(() => {
    document.documentElement.style.setProperty('--sidebar-width', `${sidebarWidth}px`)
  }, [sidebarWidth])

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return
    const media = window.matchMedia('(max-width: 1024px)')
    const mobileMedia = window.matchMedia('(max-width: 768px)')
    const syncViewport = () => {
      setCompactViewport(media.matches)
      setMobileViewport(mobileMedia.matches)
      if (media.matches) setCollapsed(true)
    }
    syncViewport()
    media.addEventListener?.('change', syncViewport)
    mobileMedia.addEventListener?.('change', syncViewport)
    return () => {
      media.removeEventListener?.('change', syncViewport)
      mobileMedia.removeEventListener?.('change', syncViewport)
    }
  }, [])

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
    <Layout style={{ height: '100vh', overflow: 'hidden' }}>
      <Sidebar
        collapsed={collapsed}
        setCollapsed={setCollapsed}
        onWidthChange={handleSidebarWidthChange}
      />
      <Layout
        style={{
          marginLeft: 0,
          display: 'flex',
          flexDirection: 'column',
          flex: 1,
          overflow: 'hidden',
        }}
      >
        <Header />
        <Content
          style={{
            margin: compactViewport ? '56px 8px 8px 8px' : '56px 16px 16px 16px',
            marginLeft: mobileViewport ? 0 : 'var(--sidebar-width, 220px)',
            padding: compactViewport ? 12 : 16,
            background: '#f5f7fa',
            flex: 1,
            overflowY: 'auto',
            overflowX: 'hidden',
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
