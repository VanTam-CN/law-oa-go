import React from 'react'
import { Navigate } from 'react-router'
import MainLayout from '../layouts/MainLayout'

// 导入新组件
import TimeTracker from '../components/TimeTracker'
import DocumentManagement from '../components/DocumentManagement'

// 懒加载页面
const CaseEnhanced = React.lazy(() => import('./case/CaseEnhanced'))
const DocumentManagementPage = React.lazy(() => import('./document/DocumentManagementPage'))
const TimeTrackingPage = React.lazy(() => import('./time/TimeTrackingPage'))
const AnalyticsPage = React.lazy(() => import('./analytics/AnalyticsPage'))

const EnhancedRoutes: React.FC = () => {
  return (
    <MainLayout>
      <React.Suspense fallback={<div>加载中...</div>}>
        <Routes>
          {/* 重定向到仪表盘 */}
          <Route index element={<Navigate to='/dashboard' replace />} />

          {/* 增强的案件管理 */}
          <Route path='case-enhanced' element={<CaseEnhanced />} />

          {/* 文档管理 */}
          <Route path='documents' element={<DocumentManagementPage />} />

          {/* 时间记录 */}
          <Route path='time-tracking' element={<TimeTrackingPage />} />

          {/* 数据分析 */}
          <Route path='analytics' element={<AnalyticsPage />} />

          {/* 其他原有路由保持不变 */}
          {/* ... */}
        </Routes>
      </React.Suspense>
    </MainLayout>
  )
}

export default EnhancedRoutes
