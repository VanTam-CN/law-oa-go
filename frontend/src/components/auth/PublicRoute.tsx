/**
 * 公开路由组件 - 未认证用户可访问，已认证用户重定向
 * 适配React Router v7
 */

import React from 'react'
import { Navigate, Outlet } from 'react-router'
import { useAppStore } from '../../stores/useAppStore'

const PublicRoute: React.FC = () => {
  const { isAuthenticated } = useAppStore()

  if (isAuthenticated) {
    // 如果用户已登录，重定向到仪表盘
    return <Navigate to="/dashboard" replace />
  }

  return <Outlet />
}

export default PublicRoute