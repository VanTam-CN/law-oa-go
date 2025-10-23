import React, { Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router';
import { ConfigProvider, Spin } from 'antd';
import zhCN from 'antd/locale/zh_CN';

// 布局
import MainLayout from './layouts/MainLayout';

// 懒加载页面组件
const LoginPage = lazy(() => import('./pages/auth/Login'));
const DashboardPage = lazy(() => import('./pages/dashboard/Dashboard'));

// 审批模块
const ApprovalList = lazy(() => import('./pages/approval/ApprovalList'));
const ApprovalDetail = lazy(() => import('./pages/approval/ApprovalDetail'));
const CreateApproval = lazy(() => import('./pages/approval/CreateApproval'));

// 业务模块
const ConflictCheck = lazy(() => import('./pages/conflict/ConflictCheck'));
const ProjectManagement = lazy(() => import('./pages/project/ProjectManagement'));
const ClientManagement = lazy(() => import('./pages/client/ClientManagement'));
const CaseManagement = lazy(() => import('./pages/case/CaseManagement'));
const CaseDetail = lazy(() => import('./pages/case/CaseDetail'));
const LawyerManagement = lazy(() => import('./pages/lawyer/LawyerManagement'));
const LawyerDetail = lazy(() => import('./pages/lawyer/LawyerDetail'));

// 行政模块
const AdminManagement = lazy(() => import('./pages/admin/AdminManagement'));

// 工具模块
const ToolsPage = lazy(() => import('./pages/tools/ToolsPage'));
const LitigationFeeCalculator = lazy(() => import('./pages/tools/LitigationFeeCalculator'));
const InterestCalculator = lazy(() => import('./pages/tools/InterestCalculator'));
const DeadlineCalculator = lazy(() => import('./pages/tools/DeadlineCalculator'));
const LawSearch = lazy(() => import('./pages/tools/LawSearch'));

// 财务模块
const FinanceManagement = lazy(() => import('./pages/finance/FinanceManagement'));

// API测试页面
const ApiTest = lazy(() => import('./pages/ApiTest'));

// 上下文
import { AuthProvider } from './context/AuthContext';

// 加载状态组件
const PageLoading: React.FC = () => (
  <div style={{ 
    display: 'flex', 
    justifyContent: 'center', 
    alignItems: 'center', 
    height: '100vh' 
  }}>
    <Spin size="large" tip="加载中..." />
  </div>
);

// 路由配置
const App: React.FC = () => {
  return (
    <ConfigProvider locale={zhCN}>
      <AuthProvider>
        <Suspense fallback={<PageLoading />}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            
            {/* 主布局下的路由 */}
            <Route path="/" element={<MainLayout />}>
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<DashboardPage />} />
              
              {/* 审批模块 */}
              <Route path="approval" element={<ApprovalList />} />
              <Route path="approval/create" element={<CreateApproval />} />
              <Route path="approval/:id" element={<ApprovalDetail />} />
              
              {/* 业务模块 */}
              <Route path="project" element={<ProjectManagement />} />
              <Route path="conflict" element={<ConflictCheck />} />
              <Route path="client" element={<ClientManagement />} />
              <Route path="case" element={<CaseManagement />} />
              <Route path="case/:id" element={<CaseDetail />} />
              <Route path="lawyer" element={<LawyerManagement />} />
              <Route path="lawyer/:id" element={<LawyerDetail />} />
              
              {/* 行政模块 */}
              <Route path="admin" element={<AdminManagement />} />
              
              {/* 工具模块 */}
              <Route path="tools" element={<ToolsPage />} />
              <Route path="tools/litigation-fee" element={<LitigationFeeCalculator />} />
              <Route path="tools/interest-calculator" element={<InterestCalculator />} />
              <Route path="tools/deadline-calculator" element={<DeadlineCalculator />} />
              <Route path="tools/law-search" element={<LawSearch />} />
              
              {/* 财务模块 */}
              <Route path="finance" element={<FinanceManagement />} />
              
              {/* API测试页面 */}
              <Route path="api-test" element={<ApiTest />} />
            </Route>
            
            {/* 404页面 */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Suspense>
      </AuthProvider>
    </ConfigProvider>
  );
};

export default App;