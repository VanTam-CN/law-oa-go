/**
 * 完整的路由配置 - 使用React Router v7
 * 基于最新最佳实践配置的现代化路由系统
 */

import React from 'react'
import { createBrowserRouter } from 'react-router'

// 页面组件
import LoginPage from '../pages/auth/Login'
import TestLogin from '../pages/auth/TestLogin'
import DashboardPage from '../pages/dashboard/Dashboard'

// 审批模块
import ApprovalList from '../pages/approval/ApprovalList'
import ApprovalDetail from '../pages/approval/ApprovalDetail'
import CreateApproval from '../pages/approval/CreateApproval'

// 业务模块
import ConflictCheck from '../pages/conflict/ConflictCheck'
import ProjectManagement from '../pages/project/ProjectManagement'
import ClientManagement from '../pages/client/ClientManagement'
import CaseManagement from '../pages/case/CaseManagement'
import CaseDetail from '../pages/case/CaseDetail'
import LawyerManagement from '../pages/lawyer/LawyerManagement'
import LawyerDetail from '../pages/lawyer/LawyerDetail'
import SimpleLawyerManagement from '../pages/lawyer/SimpleLawyerManagement'

// 文件管理模块
import FileManagement from '../pages/file/FileManagement'

// 用户管理模块
import UserManagement from '../pages/user/UserManagement'

// 行政模块
import AdminManagement from '../pages/admin/AdminManagement'

// 工具模块
import ToolsPage from '../pages/tools/ToolsPage'
import LitigationFeeCalculator from '../pages/tools/LitigationFeeCalculator'
import InterestCalculator from '../pages/tools/InterestCalculator'
import DeadlineCalculator from '../pages/tools/DeadlineCalculator'
import LawSearch from '../pages/tools/LawSearch'

// 财务模块
import FinanceManagement from '../pages/finance/FinanceManagement'

// 个人中心和设置
import Profile from '../pages/profile/Profile'
import Settings from '../pages/settings/Settings'
import PermissionTestPage from '../pages/PermissionTestPage'

// API测试页面
import ApiTest from '../pages/ApiTest'

// 测试页面
import TestPage from '../pages/TestPage'
import MinimalTest from '../pages/MinimalTest'
import SystemTest from '../pages/SystemTest'
import SimpleTest from '../pages/SimpleTest'
import AuthTest from '../pages/AuthTest'
import DirectTest from '../pages/DirectTest'

// 布局组件
import MainLayout from '../layouts/MainLayout'

// 导航组件（React Router v7需要）
import { Navigate, Outlet } from 'react-router'

// 公开路由布局组件
const PublicLayout: React.FC = () => <Outlet />

// 受保护的路由布局组件
const ProtectedLayout: React.FC = () => <Outlet />

// 简单的权限检查组件
const PermissionGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  // 这里可以根据实际需求添加权限检查逻辑
  // 暂时允许所有用户通过
  return <>{children}</>
}

const RoleGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  // 这里可以根据实际需求添加角色检查逻辑
  // 暂时允许所有用户通过
  return <>{children}</>
}

export const router = createBrowserRouter([
  // 独立测试路由 - 不依赖认证
  {
    path: '/simple-test',
    element: <SimpleTest />,
  },
  {
    path: '/auth-test',
    element: <AuthTest />,
  },
  {
    path: '/minimal',
    element: <MinimalTest />,
  },
  {
    path: '/test-direct',
    element: <DirectTest />,
  },
  // 公开路由 - 未登录用户可访问
  {
    path: '/login',
    element: <PublicLayout />,
    children: [
      {
        index: true,
        element: <LoginPage />,
      },
    ],
  },
  {
    path: '/test-login',
    element: <PublicLayout />,
    children: [
      {
        index: true,
        element: <TestLogin />,
      },
    ],
  },

  // 受保护的路由 - 需要登录
  {
    path: '/',
    element: <ProtectedLayout />,
    children: [
      {
        element: <MainLayout />,
        children: [
          {
            index: true,
            element: <Navigate to="/dashboard" replace />,
          },
          {
            path: 'dashboard',
            element: <DashboardPage />,
          },

          // 审批模块
          {
            path: 'approval',
            children: [
              {
                index: true,
                element: <ApprovalList />,
              },
              {
                path: 'create',
                element: <CreateApproval />,
              },
              {
                path: ':id',
                element: <ApprovalDetail />,
              },
            ],
          },

          // 业务模块
          {
            path: 'project',
            element: (
              <PermissionGuard>
                <ProjectManagement />
              </PermissionGuard>
            ),
          },
          {
            path: 'conflict',
            element: (
              <PermissionGuard>
                <ConflictCheck />
              </PermissionGuard>
            ),
          },
          {
            path: 'client',
            element: (
              <PermissionGuard>
                <ClientManagement />
              </PermissionGuard>
            ),
          },
          {
            path: 'case',
            children: [
              {
                index: true,
                element: (
                  <PermissionGuard>
                    <CaseManagement />
                  </PermissionGuard>
                ),
              },
              {
                path: ':id',
                element: (
                  <PermissionGuard>
                    <CaseDetail />
                  </PermissionGuard>
                ),
              },
            ],
          },
          {
            path: 'lawyer',
            children: [
              {
                index: true,
                element: (
                  <PermissionGuard>
                    <LawyerManagement />
                  </PermissionGuard>
                ),
              },
              {
                path: ':id',
                element: (
                  <PermissionGuard>
                    <LawyerDetail />
                  </PermissionGuard>
                ),
              },
              {
                path: 'simple',
                element: (
                  <PermissionGuard>
                    <SimpleLawyerManagement />
                  </PermissionGuard>
                ),
              },
            ],
          },
          {
            path: 'file',
            element: (
              <PermissionGuard>
                <FileManagement />
              </PermissionGuard>
            ),
          },

          // 用户管理模块 - 需要管理员权限
          {
            path: 'user',
            element: (
              <RoleGuard>
                <UserManagement />
              </RoleGuard>
            ),
          },

          // 行政模块 - 需要管理员权限
          {
            path: 'admin',
            element: (
              <RoleGuard>
                <AdminManagement />
              </RoleGuard>
            ),
          },

          // 工具模块
          {
            path: 'tools',
            element: <ToolsPage />,
          },
          {
            path: 'tools/litigation-fee',
            element: <LitigationFeeCalculator />,
          },
          {
            path: 'tools/interest-calculator',
            element: <InterestCalculator />,
          },
          {
            path: 'tools/deadline-calculator',
            element: <DeadlineCalculator />,
          },
          {
            path: 'tools/law-search',
            element: <LawSearch />,
          },

          // 财务模块 - 需要财务权限
          {
            path: 'finance',
            element: (
              <PermissionGuard>
                <FinanceManagement />
              </PermissionGuard>
            ),
          },

          // 个人中心和设置
          {
            path: 'profile',
            element: <Profile />,
          },
          {
            path: 'settings',
            element: <Settings />,
          },

          // 权限测试页面
          {
            path: 'permission-test',
            element: <PermissionTestPage />,
          },

          // API测试页面 - 开发环境
          ...(import.meta.env.DEV ? [
            {
              path: 'api-test',
              element: <ApiTest />,
            },
            {
              path: 'test',
              element: <TestPage />,
            },
            {
              path: 'system-test',
              element: <SystemTest />,
            },
          ] : []),
        ],
      },
    ],
  },

  // 404页面 - 捕获所有未匹配的路由
  {
    path: '*',
    element: <Navigate to="/login" replace />,
  },
])