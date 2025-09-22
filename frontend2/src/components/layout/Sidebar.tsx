import React, { useEffect, useState } from 'react';
import { Layout, Menu, message, Badge } from 'antd';
import {
  FileDoneOutlined,
  ProjectOutlined,
  SettingOutlined,
  ToolOutlined,
  MoneyCollectOutlined,
  DashboardOutlined,
  UserOutlined,
  BankOutlined,
  TeamOutlined,
  FileTextOutlined,
  SearchOutlined,
  DatabaseOutlined,
  CloudUploadOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  AppstoreOutlined,
  CalculatorOutlined,
  FileSearchOutlined,
  ScheduleOutlined,
  BarChartOutlined,
  SolutionOutlined,
  FileProtectOutlined,
  AuditOutlined
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import useAuth from '@/hooks/useAuth';
import './sidebar.less';

const { Sider } = Layout;

interface MenuItem {
  key: string;
  label: string;
  icon?: React.ReactNode;
  children?: MenuItem[];
  onClick?: () => void;
  badge?: number;
  color?: string;
}

interface SidebarProps {
  collapsed: boolean;
  setCollapsed: (collapsed: boolean) => void;
  onWidthChange?: (width: number) => void;
}

const Sidebar: React.FC<SidebarProps> = ({ collapsed, setCollapsed, onWidthChange }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();
  const [hoveredItem, setHoveredItem] = useState<string | null>(null);

  // 监听折叠状态变化，通知父组件
  useEffect(() => {
    if (onWidthChange) {
      onWidthChange(collapsed ? 80 : 240);
    }
  }, [collapsed, onWidthChange]);

  // 根据当前路径获取选中的菜单项
  const getSelectedKeys = () => {
    const path = location.pathname;
    const keyMap: Record<string, string> = {
      '/dashboard': 'dashboard',
      '/approval': 'approval',
      '/project': 'project',
      '/case': 'case',
      '/client': 'client',
      '/lawyer': 'lawyer',
      '/conflict': 'conflict',
      '/file': 'file',
      '/user': 'user',
      '/tools': 'tools',
      '/tools/law-search': 'law-search',
      '/finance': 'finance',
      '/reports': 'reports',
      '/calendar': 'calendar',
      '/documents': 'documents',
      '/settings': 'settings'
    };

    // 精确匹配
    if (keyMap[path]) return [keyMap[path]];
    
    // 模糊匹配
    for (const [route, key] of Object.entries(keyMap)) {
      if (path.startsWith(route) && path !== route) {
        return [key];
      }
    }
    
    return [];
  };

  // 获取展开的子菜单
  const getOpenKeys = () => {
    const path = location.pathname;
    const openKeys: string[] = [];
    
    if (path.startsWith('/project') || path.startsWith('/case') || 
        path.startsWith('/client') || path.startsWith('/lawyer') || 
        path.startsWith('/conflict') || path.startsWith('/file')) {
      openKeys.push('business');
    }
    
    if (path.startsWith('/tools') || path.startsWith('/reports') || 
        path.startsWith('/calendar') || path.startsWith('/documents')) {
      openKeys.push('tools');
    }
    
    return openKeys;
  };

  // 菜单项配置
  const menuItems: MenuItem[] = [
    {
      key: 'dashboard',
      label: '工作台',
      icon: <DashboardOutlined />,
      onClick: () => navigate('/dashboard'),
      color: 'var(--color-primary)'
    },
    {
      key: 'business',
      label: '业务管理',
      icon: <ProjectOutlined />,
      color: 'var(--color-success)',
      children: [
        {
          key: 'project',
          label: '项目管理',
          icon: <AppstoreOutlined />,
          onClick: () => navigate('/project')
        },
        {
          key: 'case',
          label: '案件管理',
          icon: <SolutionOutlined />,
          onClick: () => navigate('/case'),
          badge: 3
        },
        {
          key: 'client',
          label: '客户管理',
          icon: <TeamOutlined />,
          onClick: () => navigate('/client')
        },
        {
          key: 'lawyer',
          label: '律师管理',
          icon: <UserOutlined />,
          onClick: () => navigate('/lawyer')
        },
        {
          key: 'conflict',
          label: '利益冲突检查',
          icon: <FileSearchOutlined />,
          onClick: () => navigate('/conflict')
        },
        {
          key: 'file',
          label: '文件管理',
          icon: <CloudUploadOutlined />,
          onClick: () => navigate('/file')
        }
      ]
    },
    {
      key: 'approval',
      label: '审批中心',
      icon: <FileDoneOutlined />,
      onClick: () => navigate('/approval'),
      badge: 5,
      color: 'var(--color-warning)'
    },
    {
      key: 'tools',
      label: '办公工具',
      icon: <ToolOutlined />,
      color: 'var(--color-info)',
      children: [
        {
          key: 'law-search',
          label: '法条查询',
          icon: <SearchOutlined />,
          onClick: () => navigate('/tools/law-search')
        },
        {
          key: 'case-search',
          label: '案例检索',
          icon: <DatabaseOutlined />,
          onClick: () => message.info('案例检索功能开发中...')
        },
        {
          key: 'company-search',
          label: '企业信息查询',
          icon: <BankOutlined />,
          onClick: () => message.info('企业信息查询功能开发中...')
        },
        {
          key: 'calendar',
          label: '日程安排',
          icon: <ScheduleOutlined />,
          onClick: () => navigate('/calendar')
        },
        {
          key: 'documents',
          label: '文档模板',
          icon: <FileTextOutlined />,
          onClick: () => navigate('/documents')
        }
      ]
    },
    {
      key: 'finance',
      label: '财务管理',
      icon: <CalculatorOutlined />,
      onClick: () => navigate('/finance'),
      color: 'var(--color-error)'
    },
    {
      key: 'reports',
      label: '统计报表',
      icon: <BarChartOutlined />,
      onClick: () => navigate('/reports'),
      color: 'var(--color-accent)'
    },
    {
      key: 'user',
      label: '用户管理',
      icon: <UserOutlined />,
      onClick: () => navigate('/user'),
      color: 'var(--color-text-secondary)'
    },
    {
      key: 'settings',
      label: '系统设置',
      icon: <SettingOutlined />,
      onClick: () => navigate('/settings'),
      color: 'var(--color-text-secondary)'
    }
  ];

  // 渲染菜单项
  const renderMenuItems = (items: MenuItem[]): any[] => {
    return items.map(item => {
      const itemStyle: React.CSSProperties = {
        transition: 'all var(--duration-fast) var(--ease-out)',
        borderRadius: 'var(--radius-md)',
        margin: 'var(--space-0-5) var(--space-1)'
      };

      if (item.children) {
        return {
          key: item.key,
          icon: item.icon,
          label: (
            <div className="menu-item-wrapper">
              <span className="menu-item-label">{item.label}</span>
              {item.badge && (
                <Badge 
                  count={item.badge} 
                  size="small"
                  className="menu-item-badge"
                />
              )}
            </div>
          ),
          children: renderMenuItems(item.children),
          style: itemStyle,
          className: 'sidebar-submenu'
        };
      }

      return {
        key: item.key,
        icon: item.icon,
        label: (
          <div className="menu-item-wrapper">
            <span className="menu-item-label">{item.label}</span>
            {item.badge && (
              <Badge 
                count={item.badge} 
                size="small"
                className="menu-item-badge"
              />
            )}
          </div>
        ),
        onClick: item.onClick,
        style: itemStyle,
        className: 'sidebar-menu-item'
      };
    });
  };

  return (
    <Sider
      collapsible
      collapsed={collapsed}
      onCollapse={setCollapsed}
      width={240}
      collapsedWidth={80}
      className="app-sidebar"
      trigger={null}
    >
      {/* Logo区域 */}
      <div className="sidebar-logo">
        <div className="logo-container">
          {collapsed ? (
            <div className="logo-collapsed">
              <span className="logo-text">LF</span>
            </div>
          ) : (
            <div className="logo-expanded">
              <div className="logo-icon">⚖️</div>
              <div className="logo-text">
                <div className="logo-title">律所OA系统</div>
                <div className="logo-subtitle">专业 · 高效 · 智能</div>
              </div>
            </div>
          )}
        </div>
        
        {/* 折叠按钮 */}
        <div 
          className="sidebar-trigger"
          onClick={() => setCollapsed(!collapsed)}
        >
          {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        </div>
      </div>

      {/* 菜单区域 */}
      <div className="sidebar-menu-container">
        <Menu
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={renderMenuItems(menuItems)}
          className="sidebar-menu"
          inlineCollapsed={collapsed}
          expandIcon={<span className="menu-expand-icon">▼</span>}
        />
      </div>

      {/* 底部信息 */}
      {!collapsed && (
        <div className="sidebar-footer">
          <div className="footer-info">
            <div className="version-info">
              <span className="version-label">版本</span>
              <span className="version-number">v2.0.0</span>
            </div>
            <div className="user-info">
              <span className="user-role">{user?.role_name || '用户'}</span>
            </div>
          </div>
        </div>
      )}
    </Sider>
  );
};

export default Sidebar;