import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import './assets/styles/design-tokens.css';

// 布局
import MainLayout from './layouts/MainLayout';

// 页面
import LoginPage from './pages/auth/Login';
import DashboardPage from './pages/dashboard/Dashboard';

// 审批模块
import ApprovalList from './pages/approval/ApprovalList';
import ApprovalDetail from './pages/approval/ApprovalDetail';
import CreateApproval from './pages/approval/CreateApproval';

// 业务模块
import ConflictCheck from './pages/conflict/ConflictCheck';
import ProjectManagement from './pages/project/ProjectManagement';
import ClientManagement from './pages/client/ClientManagement';
import CaseManagement from './pages/case/CaseManagement';
import CaseDetail from './pages/case/CaseDetail';
import LawyerManagement from './pages/lawyer/LawyerManagement';
import LawyerDetail from './pages/lawyer/LawyerDetail';
import SimpleLawyerManagement from './pages/lawyer/SimpleLawyerManagement';

// 文件管理模块
import FileManagement from './pages/file/FileManagement';

// 用户管理模块
import UserManagement from './pages/user/UserManagement';

// 行政模块
import AdminManagement from './pages/admin/AdminManagement';

// 工具模块
import ToolsPage from './pages/tools/ToolsPage';
import LitigationFeeCalculator from './pages/tools/LitigationFeeCalculator';
import InterestCalculator from './pages/tools/InterestCalculator';
import DeadlineCalculator from './pages/tools/DeadlineCalculator';
import LawSearch from './pages/tools/LawSearch';

// 财务模块
import FinanceManagement from './pages/finance/FinanceManagement';

// 个人中心和设置
import Profile from './pages/profile/Profile';
import Settings from './pages/settings/Settings';

// API测试页面
import ApiTest from './pages/ApiTest';

// 测试页面
import TestPage from './pages/TestPage';
import MinimalTest from './pages/MinimalTest';
import SystemTest from './pages/SystemTest';
import SimpleTest from './pages/SimpleTest';
import AuthTest from './pages/AuthTest';
import DirectTest from './pages/DirectTest';

// 上下文
import { AuthProvider } from './context/AuthContext';

const App: React.FC = () => {
  return (
    <ConfigProvider locale={zhCN}>
      <Routes>
        {/* 独立测试路由 - 不依赖AuthProvider */}
        <Route path="/simple-test" element={<SimpleTest />} />
        <Route path="/auth-test" element={<AuthTest />} />
        
        <Route path="/login" element={<LoginPage />} />
        
        {/* 临时测试路由 */}
        <Route path="/test-direct" element={<DirectTest />} />
        
        {/* 主布局下的路由 */}
        <Route path="/" element={
          <AuthProvider>
            <MainLayout />
          </AuthProvider>
        }>
          <Route index element={<DashboardPage />} />
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
          <Route path="lawyer-simple" element={<SimpleLawyerManagement />} />
          <Route path="file" element={<FileManagement />} />
          
          {/* 用户管理模块 */}
          <Route path="user" element={<UserManagement />} />
          
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
          
          {/* 个人中心和设置 */}
          <Route path="profile" element={<Profile />} />
          <Route path="settings" element={<Settings />} />
          
          {/* API测试页面 */}
          <Route path="api-test" element={<ApiTest />} />
          
          {/* 测试页面 */}
          <Route path="test" element={<TestPage />} />
          <Route path="minimal" element={<MinimalTest />} />
          <Route path="system-test" element={<SystemTest />} />
        </Route>
        
        {/* 404页面 */}
        <Route path="*" element={<Navigate to="/simple-test" replace />} />
      </Routes>
    </ConfigProvider>
  );
  /*
  return (
    <ConfigProvider locale={zhCN}>
      <Routes>
        {/* 独立测试路由 - 不依赖AuthProvider * /}
        <Route path="/simple-test" element={<SimpleTest />} />
        <Route path="/auth-test" element={<AuthTest />} />
        
        <Route path="/login" element={<LoginPage />} />
        
        {/* 临时测试路由 * /}
        <Route path="/test-direct" element={<MinimalTest />} />
        
        {/* 主布局下的路由 * /}
        <Route path="/" element={
          <AuthProvider>
            <MainLayout />
          </AuthProvider>
        }>
          <Route index element={<DashboardPage />} />
          <Route path="dashboard" element={<DashboardPage />} />
          
          {/* 审批模块 * /}
          <Route path="approval" element={<ApprovalList />} />
          <Route path="approval/create" element={<CreateApproval />} />
          <Route path="approval/:id" element={<ApprovalDetail />} />
          
          {/* 业务模块 * /}
          <Route path="project" element={<ProjectManagement />} />
          <Route path="conflict" element={<ConflictCheck />} />
          <Route path="client" element={<ClientManagement />} />
          <Route path="case" element={<CaseManagement />} />
          <Route path="case/:id" element={<CaseDetail />} />
          <Route path="lawyer" element={<LawyerManagement />} />
          <Route path="lawyer/:id" element={<LawyerDetail />} />
          <Route path="lawyer-simple" element={<SimpleLawyerManagement />} />
          
          {/* 用户管理模块 * /}
          <Route path="user" element={<UserManagement />} />
          
          {/* 行政模块 * /}
          <Route path="admin" element={<AdminManagement />} />
          
          {/* 工具模块 * /}
          <Route path="tools" element={<ToolsPage />} />
          <Route path="tools/litigation-fee" element={<LitigationFeeCalculator />} />
          <Route path="tools/interest-calculator" element={<InterestCalculator />} />
          <Route path="tools/deadline-calculator" element={<DeadlineCalculator />} />
          <Route path="tools/law-search" element={<LawSearch />} />
          
          {/* 财务模块 * /}
          <Route path="finance" element={<FinanceManagement />} />
          
          {/* 个人中心和设置 * /}
          <Route path="profile" element={<Profile />} />
          <Route path="settings" element={<Settings />} />
          
          {/* API测试页面 * /}
          <Route path="api-test" element={<ApiTest />} />
          
          {/* 测试页面 * /}
          <Route path="test" element={<TestPage />} />
          <Route path="minimal" element={<MinimalTest />} />
          <Route path="system-test" element={<SystemTest />} />
        </Route>
        
        {/* 404页面 * /}
        <Route path="*" element={<Navigate to="/simple-test" replace />} />
      </Routes>
    </ConfigProvider>
  );
  */
};

export default App;