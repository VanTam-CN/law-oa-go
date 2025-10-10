import React, { useEffect } from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import DevToolsReminder from "./components/common/DevToolsReminder";
import { Provider } from "react-redux";
import { store } from "./store";
import { useAppDispatch, useAppSelector } from "./store/hooks";
import { getCurrentUser } from "./store/slices/authSlice";
import LoginPage from "./pages/LoginPage";
import RegisterPage from "./pages/RegisterPage";
import ForgotPasswordPage from "./pages/ForgotPasswordPage";
import ResetPasswordPage from "./pages/ResetPasswordPage";
import DashboardPage from "./pages/DashboardPage";
import ClientsPage from "./pages/ClientsPage";
import CasesPage from "./pages/CasesPage";
import UsersPage from "./pages/UsersPage";
import ProfilePage from "./pages/ProfilePage";
import AdminUsersPage from "./pages/AdminUsersPage";
import DocumentsPage from "./pages/DocumentsPage";
import CalendarPage from "./pages/CalendarPage";
import ReportsPage from "./pages/ReportsPage";
import TasksPage from "./pages/TasksPage";
import SettingsPage from "./pages/SettingsPage";
import HelpPage from "./pages/HelpPage";
import NotFoundPage from "./pages/NotFoundPage";
import AccessDeniedPage from "./pages/AccessDeniedPage";
import AppLayout from "./components/layout/AppLayout";
import CaseManagementPage from "./pages/CaseManagementPage";
import CaseDetailPage from "./pages/CaseDetailPage";
import CreateCasePage from "./pages/CreateCasePage";
import LawyerManagementPage from "./pages/LawyerManagementPage";
import LawyerDetailPage from "./pages/LawyerDetailPage";
import ClientManagementPage from "./pages/ClientManagementPage";
import ApprovalManagementPage from "./pages/ApprovalManagementPage";
import ApprovalDetailPage from "./pages/ApprovalDetailPage";
import CreateApprovalPage from "./pages/CreateApprovalPage";
import ToolsPage from "./pages/ToolsPage";
import LitigationFeeCalculatorPage from "./pages/LitigationFeeCalculatorPage";
import LawSearchPage from "./pages/LawSearchPage";
import ProjectManagementPage from "./pages/ProjectManagementPage";
import FinanceManagementPage from "./pages/FinanceManagementPage";
import FileManagementPage from "./pages/FileManagementPage";
import AnalyticsPage from "./pages/AnalyticsPage";
import ErrorBoundary from "./components/ErrorBoundary";
import ToastProvider, { useToast } from "./components/Toast";
import errorHandler from "./utils/errorHandler";
import "./App.css";

// Import Font Awesome CSS
import "font-awesome/css/font-awesome.min.css";

// 设置错误处理器
errorHandler.setToastEnabled(true);

// Auth wrapper component
const AuthWrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const dispatch = useAppDispatch();
  const { token, loading, isAuthenticated, error } = useAppSelector(
    (state) => state.auth,
  );

  // 开发者模式：临时绕过认证
  const isDevMode = process.env.NODE_ENV === 'development';

  useEffect(() => {
    console.log('AuthWrapper state:', { token, loading, isAuthenticated, error, isDevMode });
    // 在开发模式下跳过getCurrentUser调用，避免不必要的API请求和重定向
    if (token && !isAuthenticated && !loading && !isDevMode) {
      console.log('Getting current user...');
      dispatch(getCurrentUser());
    } else if (isDevMode && token) {
      console.log('🛠️ 开发者模式：跳过getCurrentUser调用');
    }
  }, [token, isAuthenticated, loading, dispatch, isDevMode]);

  // 开发者模式下直接渲染内容
  if (isDevMode) {
    console.log('🛠️ 开发者模式：绕过认证');
    return <>{children}</>;
  }

  // 正常认证逻辑
  if (token && loading && !isAuthenticated) {
    return (
      <div className="d-flex min-vh-100 align-items-center justify-content-center">
        <div className="spinner-border text-primary" role="status">
          <span className="visually-hidden">加载中...</span>
        </div>
      </div>
    );
  }

  return <>{children}</>;
};

function App() {
  return (
    <ErrorBoundary>
      <Provider store={store}>
        <ToastProvider>
          <DevToolsReminder />
          <AuthWrapper>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route path="/forgot-password" element={<ForgotPasswordPage />} />
              <Route path="/reset-password" element={<ResetPasswordPage />} />
              <Route path="/" element={<AppLayout />}>
                <Route index element={<Navigate to="/dashboard" replace />} />
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/clients" element={<ClientsPage />} />
                <Route path="/cases" element={<CasesPage />} />
                <Route path="/case-management" element={<CaseManagementPage />} />
                <Route path="/case-detail/:id" element={<CaseDetailPage />} />
                <Route path="/create-case" element={<CreateCasePage />} />
                <Route path="/users" element={<UsersPage />} />
                <Route path="/admin/users" element={<AdminUsersPage />} />
                <Route path="/lawyer-management" element={<LawyerManagementPage />} />
                <Route path="/lawyer/:id" element={<LawyerDetailPage />} />
                <Route path="/client-management" element={<ClientManagementPage />} />
                <Route path="/profile" element={<ProfilePage />} />
                <Route path="/documents" element={<DocumentsPage />} />
                <Route path="/calendar" element={<CalendarPage />} />
                <Route path="/reports" element={<ReportsPage />} />
                <Route path="/tasks" element={<TasksPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/help" element={<HelpPage />} />
                <Route path="/access-denied" element={<AccessDeniedPage />} />
                <Route path="/approval" element={<ApprovalManagementPage />} />
                <Route path="/approval-detail/:id" element={<ApprovalDetailPage />} />
                <Route path="/create-approval" element={<CreateApprovalPage />} />
                <Route path="/tools" element={<ToolsPage />} />
                <Route path="/fee-calculator" element={<LitigationFeeCalculatorPage />} />
                <Route path="/law-search" element={<LawSearchPage />} />
                <Route path="/project" element={<ProjectManagementPage />} />
                <Route path="/finance" element={<FinanceManagementPage />} />
                <Route path="/file" element={<FileManagementPage />} />
                <Route path="/analytics" element={<AnalyticsPage />} />
              </Route>
              <Route path="*" element={<NotFoundPage />} />
            </Routes>
          </AuthWrapper>
        </ToastProvider>
      </Provider>
    </ErrorBoundary>
  );
}

export default App;
