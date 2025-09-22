import React, { useState } from 'react';
import { Outlet } from 'react-router-dom';
import Navbar from './Navbar';
import Sidebar from './Sidebar';
import ThemeToggle from '../ui/ThemeToggle';
import useTranslation from '../../hooks/useTranslation';
import './AppLayout.css';

interface AppLayoutProps {}

const AppLayout: React.FC<AppLayoutProps> = () => {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const { t } = useTranslation();

  const toggleSidebar = () => {
    setSidebarCollapsed(!sidebarCollapsed);
  };

  return (
    <div className="d-flex min-vh-100 bg-light">
      {/* 侧边栏 */}
      <div
        className={`sidebar bg-dark text-white ${
          sidebarCollapsed ? 'collapsed' : ''
        } d-none d-md-block`}
      >
        <Sidebar collapsed={sidebarCollapsed} />
      </div>

      {/* 主要内容区域 */}
      <div className="flex-1 d-flex flex-column overflow-hidden main-content">
        {/* 顶部导航栏 */}
        <header className="navbar navbar-expand-lg navbar-light bg-white border-bottom sticky-top">
          <div className="container-fluid">
            <div className="d-flex align-items-center justify-content-between">
              <div className="d-flex align-items-center">
                {/* 移动端侧边栏切换按钮 */}
                <button
                  onClick={toggleSidebar}
                  className="btn btn-link d-md-none me-3"
                  aria-label={sidebarCollapsed ? t('common.expandMenu') : t('common.collapseMenu')}
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-6 w-6"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 6h16M4 12h16M4 18h16"
                    />
                  </svg>
                </button>

                {/* Logo */}
                <div className="d-flex align-items-center">
                  <div className="bg-primary w-8 h-8 rounded-lg d-flex align-items-center justify-center">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-5 w-5 text-white"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fillRule="evenodd"
                        d="M10 2a1 1 0 011 1v1.323l3.954 1.582 1.599-.8a1 1 0 01.894 1.79l-1.233.616 1.738 5.42a1 1 0 01-.285 1.05A3.989 3.989 0 0115 15a3.989 3.989 0 01-2.667-1.019 1 1 0 01-.285-1.049l1.715-5.349L11 6.477V5h2a1 1 0 110 2H9a1 1 0 01-1-1V3a1 1 0 011-1h1zm-6 8a1 1 0 011 1v1.323l3.954 1.582 1.599-.8a1 1 0 11.894 1.79l-1.233.616 1.738 5.42a1 1 0 01-.285 1.05A3.989 3.989 0 019 21a3.989 3.989 0 01-2.667-1.019 1 1 0 01-.285-1.049l1.715-5.349L5 12.477V11a1 1 0 011-1z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </div>
                  {!sidebarCollapsed && (
                    <span className="ms-2 h4 mb-0 text-dark">
                      Law OA
                    </span>
                  )}
                </div>
              </div>

              {/* 右侧操作区域 */}
              <div className="d-flex align-items-center">
                <ThemeToggle />
                <button className="btn btn-link position-relative">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
                    />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </header>

        {/* 页面内容 */}
        <main className="flex-1 overflow-y-auto p-4">
          <div className="fade-in">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
};

export default AppLayout;