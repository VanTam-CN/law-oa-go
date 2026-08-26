/**
 * 主应用组件 - React 18+ 现代化架构
 * 集成状态管理、路由、权限控制和错误边界
 */

import React from 'react'
import { Routes, Route, Navigate } from 'react-router'
import { ConfigProvider, App as AntdApp, Button, Result } from 'antd'
import zhCN from 'antd/locale/zh_CN'

// 主题配置 - 律所专业配色
import { antdTheme } from './config/theme'

// 状态管理和Hook
import { useAppStore } from './stores/useAppStore'
import { canAccess } from './utils/accessControl'
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
import {
  ApprovalDecisionFlow,
  ApprovalWorkbench,
  CaseDetailCenter,
  CaseManagementCenter,
  CaseIntakeWorkbench,
  ClientMasterProfile,
  ConflictCheckResults,
  ConflictGovernanceCenter,
  DashboardCommandCenter,
  LawyerProfileCenter,
  LawyerResourceCenter,
  SystemSettingsCenter,
  UserAccessCenter,
} from './pages/batch01/Batch01Prototype'

// 审批模块
import CreateApproval from './pages/approval/CreateApproval'

// 业务模块
// 项目管理功能已禁用，与案件管理重复
// import ProjectManagement from './pages/project/ProjectManagement'

// 文件管理模块
import FileManagement from './pages/file/FileManagement'

// 行政模块
import RoleManagement from './pages/admin/RoleManagement'
import PermissionManagement from './pages/admin/PermissionManagement'

// 工具模块
import ToolsPage from './pages/tools/ToolsPage'
import LitigationFeeCalculator from './pages/tools/LitigationFeeCalculator'
import InterestCalculator from './pages/tools/InterestCalculator'
import DeadlineCalculator from './pages/tools/DeadlineCalculator'
import LawSearch from './pages/tools/LawSearch'

// 财务模块
import FinanceManagement from './pages/finance/FinanceManagement'
import ContractDetail from './pages/finance/ContractDetail'
import BadDebtList from './pages/finance/BadDebtList'
import CommissionRuleList from './pages/finance/CommissionRuleList'

// 代管款模块
import TrustManagement from './pages/trust/TrustManagement'

// 个人中心和设置
import Profile from './pages/profile/Profile'

// 收件箱模块
import InboxList from './pages/inbox/InboxList'
import OperationsReadiness from './pages/operations/OperationsReadiness'

// MVP 组件
import MvpUnavailable from './components/mvp/MvpUnavailable'
import ConflictWorkbench from './pages/conflict/ConflictWorkbench'

const enableDevRoutes = import.meta.env.DEV && import.meta.env.VITE_ENABLE_DEV_ROUTES === 'true'

const DevTestLogin = enableDevRoutes
  ? React.lazy(() => import('./pages/auth/TestLogin'))
  : undefined
const DevApiTest = enableDevRoutes ? React.lazy(() => import('./pages/ApiTest')) : undefined
const DevPermissionTestPage = enableDevRoutes
  ? React.lazy(() => import('./pages/PermissionTestPage'))
  : undefined
const DevTestPage = enableDevRoutes ? React.lazy(() => import('./pages/TestPage')) : undefined
const DevMinimalTest = enableDevRoutes ? React.lazy(() => import('./pages/MinimalTest')) : undefined
const DevSystemTest = enableDevRoutes ? React.lazy(() => import('./pages/SystemTest')) : undefined
const DevSimpleTest = enableDevRoutes ? React.lazy(() => import('./pages/SimpleTest')) : undefined
const DevAuthTest = enableDevRoutes ? React.lazy(() => import('./pages/AuthTest')) : undefined
const DevDirectTest = enableDevRoutes ? React.lazy(() => import('./pages/DirectTest')) : undefined
const DevSimpleLawyerManagement = enableDevRoutes
  ? React.lazy(() => import('./pages/lawyer/SimpleLawyerManagement'))
  : undefined

interface RequireAccessProps {
  children: React.ReactNode
  permissions?: string | string[]
  roles?: string | string[]
}

const missingAccessText = (permissions?: string | string[]) => {
  const required = Array.isArray(permissions) ? permissions : permissions ? [permissions] : []
  if (required.some((permission) => permission.startsWith('finance:'))) {
    return '当前账号没有访问财务中心的权限。该模块需要财务角色或管理员授权。'
  }
  return '当前账号没有访问该功能的权限。'
}

const ForbiddenPage: React.FC<{ permissions?: string | string[] }> = ({ permissions }) => (
  <Result
    status='403'
    title='无权访问'
    subTitle={missingAccessText(permissions)}
    extra={
      <Button
        type='primary'
        onClick={() => {
          window.location.href = '/dashboard'
        }}
      >
        返回工作台
      </Button>
    }
  />
)

const RequireAccess: React.FC<RequireAccessProps> = ({ children, permissions, roles }) => {
  const { user } = useAppStore()
  if (!canAccess(user, permissions, roles)) {
    return <ForbiddenPage permissions={permissions} />
  }
  return <>{children}</>
}

const withAccess = (
  element: React.ReactElement,
  permissions?: string | string[],
  roles?: string | string[],
) => (
  <RequireAccess permissions={permissions} roles={roles}>
    {element}
  </RequireAccess>
)

const AppContent: React.FC = () => {
  const { isAuthenticated } = useAppStore()
  const { message: appMessage } = AntdApp.useApp()

  // 设置message实例
  React.useEffect(() => {
    if (appMessage) {
      setAppMessage(appMessage)
    }
  }, [appMessage])

  return (
    <ErrorBoundary>
      <Routes>
        {enableDevRoutes && DevSimpleTest && DevAuthTest && DevMinimalTest && DevDirectTest && (
          <>
            <Route path='/simple-test' element={<DevSimpleTest />} />
            <Route path='/auth-test' element={<DevAuthTest />} />
            <Route path='/minimal' element={<DevMinimalTest />} />
            <Route path='/test-direct' element={<DevDirectTest />} />
          </>
        )}

        {/* 公开路由 - 未登录用户可访问 */}
        <Route
          path='/login'
          element={isAuthenticated ? <Navigate to='/dashboard' replace /> : <LoginPage />}
        />
        {enableDevRoutes && DevTestLogin && (
          <Route
            path='/test-login'
            element={isAuthenticated ? <Navigate to='/dashboard' replace /> : <DevTestLogin />}
          />
        )}

        {/* 受保护的路由 - 需要登录 */}
        <Route
          path='/'
          element={isAuthenticated ? <MainLayout /> : <Navigate to='/login' replace />}
        >
          <Route index element={<Navigate to='/dashboard' replace />} />
          <Route
            path='dashboard'
            element={withAccess(<DashboardCommandCenter />, 'dashboard:view')}
          />

          {/* 审批模块 */}
          <Route path='approval' element={withAccess(<ApprovalWorkbench />, 'approval:view')} />
          <Route
            path='approval/create'
            element={withAccess(<CreateApproval />, 'approval:manage')}
          />
          <Route
            path='approval/:id'
            element={withAccess(<ApprovalDecisionFlow />, 'approval:view')}
          />

          {/* 业务模块 */}
          {/* 项目管理功能已禁用，与案件管理重复 */}
          {/* <Route path='project' element={<ProjectManagement />} /> */}
          <Route path='conflict' element={withAccess(<ConflictWorkbench />, 'conflict:check')} />
          <Route
            path='conflict-governance'
            element={withAccess(
              <ConflictGovernanceCenter />,
              'conflict:governance',
              ['director', 'partner', 'management', 'compliance', 'risk', 'risk_control', 'conflict_officer'],
            )}
          />
          {/* Legacy form is intentionally redirected so users enter the audited list-first workflow. */}
          <Route path='conflict/v2' element={<Navigate to='/conflict' replace />} />
          <Route path='client' element={withAccess(<ClientMasterProfile />, 'client:view')} />
          <Route path='case' element={withAccess(<CaseManagementCenter />, 'case:view')} />
          <Route path='cases' element={<Navigate to='/case' replace />} />
          <Route
            path='case/create'
            element={withAccess(<CaseIntakeWorkbench />, 'case:manage', 'assistant')}
          />
          <Route path='case/:id' element={withAccess(<CaseDetailCenter />, 'case:view')} />
          <Route path='lawyer' element={withAccess(<LawyerResourceCenter />, 'lawyer:manage')} />
          <Route path='lawyer/:id' element={withAccess(<LawyerProfileCenter />, 'lawyer:manage')} />
          <Route
            path='file'
            element={withAccess(<MvpUnavailable moduleName='文档中心' />, 'file:view')}
          />

          {/* 用户管理模块 - 需要管理员权限 */}
          <Route path='user' element={withAccess(<UserAccessCenter />, 'user:manage')} />

          {/* 行政模块 - 需要管理员权限 */}
          <Route path='admin' element={<Navigate to='/user' replace />} />
          <Route path='admin/roles' element={withAccess(<RoleManagement />, 'role:manage')} />
          <Route
            path='admin/permissions'
            element={withAccess(<PermissionManagement />, 'permission:manage')}
          />

          {/* 工具模块 */}
          <Route path='tools' element={withAccess(<ToolsPage />, 'tools:view')} />
          <Route
            path='tools/litigation-fee'
            element={withAccess(<LitigationFeeCalculator />, 'tools:view')}
          />
          <Route
            path='tools/interest-calculator'
            element={withAccess(<InterestCalculator />, 'tools:view')}
          />
          <Route
            path='tools/deadline-calculator'
            element={withAccess(<DeadlineCalculator />, 'tools:view')}
          />
          <Route path='tools/law-search' element={withAccess(<LawSearch />, 'tools:view')} />

          {/* 财务模块 - MVP 期间显示不可用页面 */}
          <Route
            path='finance'
            element={withAccess(<MvpUnavailable moduleName='财务中心' />, 'finance:view')}
          />
          <Route
            path='finance/contracts/:id'
            element={withAccess(<MvpUnavailable moduleName='财务中心' />, 'finance:view')}
          />
          <Route
            path='finance/bad-debts'
            element={withAccess(<MvpUnavailable moduleName='财务中心' />, 'finance:manage')}
          />
          <Route
            path='finance/commission-rules'
            element={withAccess(<MvpUnavailable moduleName='财务中心' />, 'finance:manage')}
          />

          {/* 代管款模块 - 需要财务权限 */}
          <Route path='trust' element={withAccess(<TrustManagement />, 'trust:manage')} />

          {/* 个人中心和设置 */}
          <Route path='profile' element={<Profile />} />
          <Route
            path='settings'
            element={withAccess(<MvpUnavailable moduleName='系统设置' />, 'system:manage')}
          />

          {/* 收件箱模块 */}
          <Route path='inbox' element={<InboxList />} />
          <Route
            path='operations/readiness'
            element={withAccess(<OperationsReadiness />, 'system:manage')}
          />

          {enableDevRoutes &&
            DevApiTest &&
            DevPermissionTestPage &&
            DevTestPage &&
            DevSystemTest &&
            DevSimpleLawyerManagement && (
              <>
                <Route path='api-test' element={<DevApiTest />} />
                <Route path='permission-test' element={<DevPermissionTestPage />} />
                <Route path='test' element={<DevTestPage />} />
                <Route path='system-test' element={<DevSystemTest />} />
                <Route path='lawyer-simple' element={<DevSimpleLawyerManagement />} />
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
  const { isLoading } = useAppStore()

  if (isLoading) {
    return (
      <ConfigProvider locale={zhCN} theme={antdTheme}>
        <PageLoading message='正在初始化应用...' />
      </ConfigProvider>
    )
  }

  return (
    <ConfigProvider locale={zhCN} theme={antdTheme}>
      <AntdApp>
        <ErrorBoundary>
          <React.Suspense fallback={<PageLoading message='正在加载...' />}>
            <AppContent />
          </React.Suspense>
        </ErrorBoundary>
      </AntdApp>
    </ConfigProvider>
  )
})

export default App
