import React, { useState, useEffect } from 'react';
import { Layout } from 'antd';
import { Outlet } from 'react-router-dom';
import Header from '@/components/layout/Header';
import Sidebar from '@/components/layout/Sidebar';
import useAuth from '@/hooks/useAuth';
import { Navigate } from 'react-router-dom';

const { Content } = Layout;

const MainLayout: React.FC = () => {
  const { user, loading } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const [sidebarWidth, setSidebarWidth] = useState(220);

  // 处理侧边栏宽度变化
  const handleSidebarWidthChange = (width: number) => {
    setSidebarWidth(width);
  };

  // 动态调整CSS变量来控制侧边栏宽度
  useEffect(() => {
    document.documentElement.style.setProperty('--sidebar-width', `${sidebarWidth}px`);
  }, [sidebarWidth]);

  // 如果正在加载，显示加载状态
  if (loading) {
    return <div>加载中...</div>;
  }

  // 如果未登录，重定向到登录页
  if (!user) {
    return <Navigate to="/login" replace />;
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
        <Content style={{ 
          margin: '80px 16px 24px 16px', 
          marginLeft: 'var(--sidebar-width, 220px)',
          padding: 24, 
          background: '#fff', 
          minHeight: 280,
          position: 'relative',
          zIndex: 1
        }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;