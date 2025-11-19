/**
 * 主应用组件 - React 18+ 现代化架构
 * 集成状态管理、路由、权限控制和错误边界
 */

import React, { useMemo } from 'react'
import { Routes, Route, Navigate } from 'react-router'
import { ConfigProvider, App as AntdApp } from 'antd'
import zhCN from 'antd/locale/zh_CN'

// 状态管理和Hook
import { useAppStore } from './stores/useAppStore'
import { setAppMessage } from './utils/messageHelper'
import './utils/tokenCheck' // 自动进行token统一性检查

// 错误边界和加载状态
import ErrorBoundary from './components/common/ErrorBoundary'
import LoadingFallback, { PageLoading } from './components/common/LoadingFallback'

// 路由守卫组件
// import ProtectedRoute from './components/auth/ProtectedRoute'
// import PublicRoute from './components/auth/PublicRoute'

// 布局
import MainLayout from './layouts/MainLayout'

// 页面
import LoginPage from './pages/auth/Login'
import TestLogin from './pages/auth/TestLogin'
import DashboardPage from './pages/dashboard/Dashboard'

// 审批模块
import ApprovalList from './pages/approval/ApprovalList'
import ApprovalDetail from './pages/approval/ApprovalDetail'
import CreateApproval from './pages/approval/CreateApproval'

// 业务模块
import ConflictCheck from './pages/conflict/ConflictCheck'
import ProjectManagement from './pages/project/ProjectManagement'
import ClientManagement from './pages/client/ClientManagement'
import CaseManagement from './pages/case/CaseManagement'
import CaseDetail from './pages/case/CaseDetail'
import LawyerManagement from './pages/lawyer/LawyerManagement'
import LawyerDetail from './pages/lawyer/LawyerDetail'
import SimpleLawyerManagement from './pages/lawyer/SimpleLawyerManagement'

// 文件管理模块
import FileManagement from './pages/file/FileManagement'

// 用户管理模块
import UserManagement from './pages/user/UserManagement'

// 行政模块
import AdminManagement from './pages/admin/AdminManagement'

// 工具模块
import ToolsPage from './pages/tools/ToolsPage'
import LitigationFeeCalculator from './pages/tools/LitigationFeeCalculator'
import InterestCalculator from './pages/tools/InterestCalculator'
import DeadlineCalculator from './pages/tools/DeadlineCalculator'
import LawSearch from './pages/tools/LawSearch'

// 财务模块
import FinanceManagement from './pages/finance/FinanceManagement'

// 个人中心和设置
import Profile from './pages/profile/Profile'
import Settings from './pages/settings/Settings'
import PermissionTestPage from './pages/PermissionTestPage'

// API测试页面
import ApiTest from './pages/ApiTest'

// 测试页面
import TestPage from './pages/TestPage'
import MinimalTest from './pages/MinimalTest'
import SystemTest from './pages/SystemTest'
import SimpleTest from './pages/SimpleTest'
import AuthTest from './pages/AuthTest'
import DirectTest from './pages/DirectTest'

// 主应用内容组件（简化版本）
interface AppContentProps {
  appMessage: any
}

const AppContent: React.FC<AppContentProps> = ({ appMessage }) => {
  const { isAuthenticated } = useAppStore()

  // 设置message实例
  React.useEffect(() => {
    if (appMessage) {
      setAppMessage(appMessage)
    }
  }, [appMessage])

  return (
    <ErrorBoundary>
      <Routes>
        {/* 独立测试路由 - 不依赖认证 */}
        <Route path='/simple-test' element={<SimpleTest />} />
        <Route path='/auth-test' element={<AuthTest />} />
        <Route path='/minimal' element={<MinimalTest />} />
        <Route path='/test-direct' element={<DirectTest />} />

        {/* 公开路由 - 未登录用户可访问 */}
        <Route
          path='/login'
          element={isAuthenticated ? <Navigate to='/dashboard' replace /> : <LoginPage />}
        />
        <Route
          path='/test-login'
          element={isAuthenticated ? <Navigate to='/dashboard' replace /> : <TestLogin />}
        />

        {/* 受保护的路由 - 需要登录 */}
        <Route
          path='/'
          element={isAuthenticated ? <MainLayout /> : <Navigate to='/login' replace />}
        >
          <Route index element={<Navigate to='/dashboard' replace />} />
          <Route path='dashboard' element={<DashboardPage />} />

          {/* 审批模块 */}
          <Route path='approval' element={<ApprovalList />} />
          <Route path='approval/create' element={<CreateApproval />} />
          <Route path='approval/:id' element={<ApprovalDetail />} />

          {/* 业务模块 */}
          <Route path='project' element={<ProjectManagement />} />
          <Route path='conflict' element={<ConflictCheck />} />
          <Route path='client' element={<ClientManagement />} />
          <Route path='case' element={<CaseManagement />} />
          <Route path='case/:id' element={<CaseDetail />} />
          <Route path='lawyer' element={<LawyerManagement />} />
          <Route path='lawyer/:id' element={<LawyerDetail />} />
          <Route path='lawyer-simple' element={<SimpleLawyerManagement />} />
          <Route path='file' element={<FileManagement />} />

          {/* 用户管理模块 - 需要管理员权限 */}
          <Route path='user' element={<UserManagement />} />

          {/* 行政模块 - 需要管理员权限 */}
          <Route path='admin' element={<AdminManagement />} />

          {/* 工具模块 */}
          <Route path='tools' element={<ToolsPage />} />
          <Route path='tools/litigation-fee' element={<LitigationFeeCalculator />} />
          <Route path='tools/interest-calculator' element={<InterestCalculator />} />
          <Route path='tools/deadline-calculator' element={<DeadlineCalculator />} />
          <Route path='tools/law-search' element={<LawSearch />} />

          {/* 财务模块 - 需要财务权限 */}
          <Route path='finance' element={<FinanceManagement />} />

          {/* 个人中心和设置 */}
          <Route path='profile' element={<Profile />} />
          <Route path='settings' element={<Settings />} />

          {/* API测试页面 - 开发环境 */}
          {import.meta.env.DEV && <Route path='api-test' element={<ApiTest />} />}

          {/* 权限测试页面 */}
          <Route path='permission-test' element={<PermissionTestPage />} />

          {/* 测试页面 - 仅开发环境 */}
          {import.meta.env.DEV && (
            <>
              <Route path='test' element={<TestPage />} />
              <Route path='system-test' element={<SystemTest />} />
            </>
          )}
        </Route>

        {/* 404页面 */}
        <Route
          path='*'
          element={
            isAuthenticated ? (
              <Navigate to='/dashboard' replace />
            ) : (
              <Navigate to='/login' replace />
            )
          }
        />
      </Routes>
    </ErrorBoundary>
  )
}

// 主App组件（修复Message组件问题）
const App: React.FC = React.memo(() => {
  const { isAuthenticated, isLoading } = useAppStore()
  const { message: antdMessage } = AntdApp.useApp()

  // 简化的主题配置
  const themeConfig = useMemo(
    () => ({
      token: {
        colorPrimary: '#1890ff',
        borderRadius: 6,
        fontSize: 14,
        fontFamily:
          '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
      },
      components: {
        Layout: {
          headerBg: '#ffffff',
          siderBg: '#f6f6f6',
          bodyBg: '#f0f2f5',
        },
      },
    }),
    [],
  )

  if (isLoading) {
    return (
      <ConfigProvider locale={zhCN} theme={themeConfig}>
        <PageLoading message='正在初始化应用...' />
      </ConfigProvider>
    )
  }

  return (
    <ConfigProvider locale={zhCN} theme={themeConfig}>
      <AntdApp>
        <ErrorBoundary>
          <React.Suspense fallback={<PageLoading message='正在加载...' />}>
            <AppContent appMessage={antdMessage} />
          </React.Suspense>
        </ErrorBoundary>
      </AntdApp>
    </ConfigProvider>
  )
})

export default App
