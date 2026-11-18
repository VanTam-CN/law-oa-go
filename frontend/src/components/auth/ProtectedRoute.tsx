/**
 * 受保护路由组件 - 需要认证的用户才能访问
 * 适配React Router v7
 */

import React from 'react'
import { Navigate, Outlet } from 'react-router'
import { useAppStore } from '../../stores/useAppStore'

const ProtectedRoute: React.FC = () => {
  const { isAuthenticated } = useAppStore()

  if (!isAuthenticated) {
    return <Navigate to='/login' replace />
  }

  return <Outlet />
}

export default ProtectedRoute
