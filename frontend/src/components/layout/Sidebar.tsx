import React, { useEffect, useMemo } from 'react'
import { Layout, Menu } from 'antd'
import { message } from '@/utils/messageHelper'
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
  AuditOutlined,
  BellOutlined,
  WalletOutlined,
} from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router'
import { useAppStore, hasRole } from '@/stores/useAppStore'
import './sidebar.less'

const { Sider } = Layout

interface MenuItem {
  key: string
  label: string
  icon?: React.ReactNode
  children?: MenuItem[]
  onClick?: () => void
  color?: string
  permission?: string
}

interface SidebarProps {
  collapsed: boolean
  setCollapsed: (collapsed: boolean) => void
  onWidthChange?: (width: number) => void
}

const Sidebar: React.FC<SidebarProps> = ({ collapsed, setCollapsed, onWidthChange }) => {
  const navigate = useNavigate()
  const location = useLocation()
  const { user } = useAppStore()

  // 简单的权限检查函数
  const hasPermission = (permission: string): boolean => {
    if (!user) {
      return false
    }
    // 管理员拥有所有权限
    if (user.roles.includes('admin')) {
      return true
    }
    // 简单的权限映射
    const permissionMap: Record<string, string[]> = {
      'dashboard.view': ['admin', 'lawyer', 'user'],
      'user.view': ['admin'],
      'user.manage': ['admin'],
      'case.view': ['admin', 'lawyer'],
      'case.manage': ['admin', 'lawyer'],
      'client.view': ['admin', 'lawyer'],
      'client.manage': ['admin', 'lawyer'],
      'project.view': ['admin', 'lawyer'],
      'project.manage': ['admin'],
      'conflict.check': ['admin', 'lawyer'],
      'file.view': ['admin', 'lawyer'],
      'file.manage': ['admin', 'lawyer'],
      'finance.view': ['admin'],
      'finance.manage': ['admin'],
    }

    const requiredRoles = permissionMap[permission]
    return requiredRoles ? requiredRoles.some((role) => user.roles.includes(role)) : false
  }

  // 监听折叠状态变化，通知父组件
  useEffect(() => {
    if (onWidthChange) {
      onWidthChange(collapsed ? 80 : 240)
    }
  }, [collapsed, onWidthChange])

  // 根据当前路径获取选中的菜单项
  const getSelectedKeys = () => {
    const path = location.pathname
    const keyMap: Record<string, string> = {
      '/dashboard': 'dashboard',
      '/inbox': 'inbox',
      '/approval': 'approval',
      // 项目管理功能已禁用
      // '/project': 'project',
      '/case': 'case',
      '/client': 'client',
      '/lawyer': 'lawyer',
      '/conflict': 'conflict',
      '/file': 'file',
      '/user': 'user',
      '/tools': 'tools',
      '/tools/law-search': 'law-search',
      '/finance': 'finance',
      '/trust': 'trust',
      '/reports': 'reports',
      '/calendar': 'calendar',
      '/documents': 'documents',
      '/settings': 'settings',
    }

    // 精确匹配
    if (keyMap[path]) {
      return [keyMap[path]]
    }

    // 模糊匹配
    for (const [route, key] of Object.entries(keyMap)) {
      if (path.startsWith(route) && path !== route) {
        return [key]
      }
    }

    return []
  }

  // 获取展开的子菜单
  const getOpenKeys = () => {
    const path = location.pathname
    const openKeys: string[] = []

    if (
      // 项目管理功能已禁用
      // path.startsWith('/project') ||
      path.startsWith('/case') ||
      path.startsWith('/client') ||
      path.startsWith('/lawyer') ||
      path.startsWith('/conflict') ||
      path.startsWith('/file')
    ) {
      openKeys.push('business')
    }

    if (
      path.startsWith('/tools') ||
      path.startsWith('/reports') ||
      path.startsWith('/calendar') ||
      path.startsWith('/documents')
    ) {
      openKeys.push('tools')
    }

    return openKeys
  }

  // 基础菜单项配置
  const baseMenuItems: MenuItem[] = [
    {
      key: 'dashboard',
      label: '工作台',
      icon: <DashboardOutlined />,
      onClick: () => navigate('/dashboard'),
      color: 'var(--color-primary)',
      permission: 'dashboard:view',
    },
    {
      key: 'inbox',
      label: '收件箱',
      icon: <BellOutlined />,
      onClick: () => navigate('/inbox'),
      color: 'var(--color-warning)',
      permission: 'inbox:view',
    },
    {
      key: 'business',
      label: '业务管理',
      icon: <ProjectOutlined />,
      color: 'var(--color-success)',
      children: [
        // 项目管理功能已屏蔽，与案件管理重复
        // {
        //   key: 'project',
        //   label: '项目管理',
        //   icon: <AppstoreOutlined />,
        //   onClick: () => navigate('/project'),
        //   permission: 'project:manage',
        // },
        {
          key: 'case',
          label: '案件管理',
          icon: <SolutionOutlined />,
          onClick: () => navigate('/case'),
          permission: 'case:manage',
        },
        {
          key: 'client',
          label: '客户管理',
          icon: <TeamOutlined />,
          onClick: () => navigate('/client'),
          permission: 'client:manage',
        },
        {
          key: 'lawyer',
          label: '律师管理',
          icon: <UserOutlined />,
          onClick: () => navigate('/lawyer'),
          permission: 'lawyer:manage',
        },
        {
          key: 'conflict',
          label: '利益冲突检查',
          icon: <FileSearchOutlined />,
          onClick: () => navigate('/conflict'),
          permission: 'conflict:check',
        },
        {
          key: 'file',
          label: '文件管理',
          icon: <CloudUploadOutlined />,
          onClick: () => navigate('/file'),
          permission: 'file:manage',
        },
      ],
    },
    {
      key: 'approval',
      label: '审批中心',
      icon: <FileDoneOutlined />,
      onClick: () => navigate('/approval'),
      color: 'var(--color-warning)',
      permission: 'approval:manage',
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
          onClick: () => navigate('/tools/law-search'),
          permission: 'law:search',
        },
        {
          key: 'case-search',
          label: '案例检索',
          icon: <DatabaseOutlined />,
          onClick: () => message.info('案例检索功能开发中...'),
          permission: 'case:search',
        },
        {
          key: 'company-search',
          label: '企业信息查询',
          icon: <BankOutlined />,
          onClick: () => message.info('企业信息查询功能开发中...'),
          permission: 'company:search',
        },
        {
          key: 'calendar',
          label: '日程安排',
          icon: <ScheduleOutlined />,
          onClick: () => navigate('/calendar'),
          permission: 'calendar:manage',
        },
        {
          key: 'documents',
          label: '文档模板',
          icon: <FileTextOutlined />,
          onClick: () => navigate('/documents'),
          permission: 'document:template',
        },
      ],
    },
    {
      key: 'finance',
      label: '财务管理',
      icon: <CalculatorOutlined />,
      onClick: () => navigate('/finance'),
      color: 'var(--color-error)',
      permission: 'finance:manage',
    },
    {
      key: 'trust',
      label: '代管款',
      icon: <WalletOutlined />,
      onClick: () => navigate('/trust'),
      color: 'var(--color-warning)',
      permission: 'finance:manage',
    },
    {
      key: 'reports',
      label: '统计报表',
      icon: <BarChartOutlined />,
      onClick: () => navigate('/reports'),
      color: 'var(--color-accent)',
      permission: 'report:view',
    },
    {
      key: 'user',
      label: '用户管理',
      icon: <UserOutlined />,
      onClick: () => navigate('/user'),
      color: 'var(--color-text-secondary)',
      permission: 'user:manage',
    },
    {
      key: 'settings',
      label: '系统设置',
      icon: <SettingOutlined />,
      onClick: () => navigate('/settings'),
      color: 'var(--color-text-secondary)',
      permission: 'system:manage',
    },
  ]

  // 根据权限过滤菜单项
  const menuItems = useMemo(() => {
    const filterMenuItems = (items: MenuItem[]): MenuItem[] => {
      return items
        .filter((item) => {
          // 如果是父级菜单，检查是否有可访问的子菜单
          if (item.children) {
            const filteredChildren = filterMenuItems(item.children)
            return filteredChildren.length > 0
          }

          // 如果有权限要求，检查权限
          if (item.permission) {
            return hasPermission(item.permission)
          }

          // 默认显示
          return true
        })
        .map((item) => {
          // 递归处理子菜单
          if (item.children) {
            return {
              ...item,
              children: filterMenuItems(item.children),
            }
          }
          return item
        })
    }

    return filterMenuItems(baseMenuItems)
  }, [hasPermission])

  // 渲染菜单项
  const renderMenuItems = (items: MenuItem[]): any[] => {
    return items.map((item) => {
      const itemStyle: React.CSSProperties = {
        transition: 'all var(--duration-fast) var(--ease-out)',
        borderRadius: 'var(--radius-md)',
        margin: 'var(--space-0-5) var(--space-1)',
      }

      if (item.children) {
        return {
          key: item.key,
          icon: item.icon,
          label: item.label,
          children: renderMenuItems(item.children),
          style: itemStyle,
          className: 'sidebar-submenu',
        }
      }

      return {
        key: item.key,
        icon: item.icon,
        label: item.label,
        onClick: item.onClick,
        style: itemStyle,
        className: 'sidebar-menu-item',
      }
    })
  }

  return (
    <Sider
      collapsible
      collapsed={collapsed}
      onCollapse={setCollapsed}
      width={240}
      collapsedWidth={80}
      className='app-sidebar'
      trigger={null}
    >
      {/* Logo区域 */}
      <div className='sidebar-logo'>
        <div className='logo-container'>
          {collapsed ? (
            <div className='logo-collapsed'>
              <span className='logo-text'>LF</span>
            </div>
          ) : (
            <div className='logo-expanded'>
              <div className='logo-icon'>⚖️</div>
              <div className='logo-text'>
                <div className='logo-title'>律所OA系统</div>
                <div className='logo-subtitle'>专业 · 高效 · 智能</div>
              </div>
            </div>
          )}
        </div>

        {/* 折叠按钮 */}
        <div className='sidebar-trigger' onClick={() => setCollapsed(!collapsed)}>
          {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        </div>
      </div>

      {/* 菜单区域 */}
      <div className='sidebar-menu-container'>
        <Menu
          mode='inline'
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={renderMenuItems(menuItems)}
          className='sidebar-menu'
          inlineCollapsed={collapsed}
          expandIcon={<span className='menu-expand-icon'>▼</span>}
        />
      </div>

      {/* 底部信息 */}
      {!collapsed && (
        <div className='sidebar-footer'>
          <div className='footer-info'>
            <div className='version-info'>
              <span className='version-label'>版本</span>
              <span className='version-number'>v2.0.0</span>
            </div>
            <div className='user-info'>
              <span className='user-role'>{user?.role_name || '用户'}</span>
            </div>
          </div>
        </div>
      )}
    </Sider>
  )
}

export default Sidebar
