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
  SafetyCertificateOutlined,
} from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router'
import { useAppStore } from '@/stores/useAppStore'
import { hasPermission as userHasPermission } from '@/utils/accessControl'
import { isMvpMenuKey } from '@/config/mvp'
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

  const hasPermission = React.useCallback(
    (permission: string): boolean => userHasPermission(user, permission),
    [user],
  )

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
      '/case': 'case',
      '/client': 'client',
      '/lawyer': 'lawyer',
      '/conflict': 'conflict',
      '/file': 'file',
      '/user': 'user',
      '/admin/roles': 'role',
      '/admin/permissions': 'permission',
      '/tools': 'tools',
      '/tools/law-search': 'law-search',
      '/finance': 'finance',
      '/trust': 'trust',
      '/settings': 'settings',
      '/operations/readiness': 'operations-readiness',
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
      path.startsWith('/tools') ||
      path.startsWith('/reports') ||
      path.startsWith('/calendar') ||
      path.startsWith('/documents')
    ) {
      openKeys.push('tools')
    }

    return openKeys
  }

  // 基础菜单项配置（MVP 项目按正确顺序排列在前）
  const baseMenuItems: MenuItem[] = [
    // MVP 核心路径
    {
      key: 'dashboard',
      label: '工作台',
      icon: <DashboardOutlined />,
      onClick: () => navigate('/dashboard'),
      color: 'var(--color-primary)',
      permission: 'dashboard:view',
    },
    {
      key: 'case',
      label: '案件管理',
      icon: <SolutionOutlined />,
      onClick: () => navigate('/case'),
      color: 'var(--color-success)',
      permission: 'case:manage',
    },
    {
      key: 'conflict',
      label: '利益冲突检查',
      icon: <FileSearchOutlined />,
      onClick: () => navigate('/conflict'),
      color: 'var(--color-success)',
      permission: 'conflict:check',
    },
    {
      key: 'conflict-governance',
      label: '冲突治理',
      icon: <SafetyCertificateOutlined />,
      onClick: () => navigate('/conflict-governance'),
      color: 'var(--color-text-secondary)',
      permission: 'conflict:governance',
    },
    {
      key: 'client',
      label: '客户管理',
      icon: <TeamOutlined />,
      onClick: () => navigate('/client'),
      color: 'var(--color-success)',
      permission: 'client:manage',
    },
    {
      key: 'approval',
      label: '审批中心',
      icon: <FileDoneOutlined />,
      onClick: () => navigate('/approval'),
      color: 'var(--color-warning)',
      permission: 'approval:view',
    },
    {
      key: 'trust',
      label: '代管款管理',
      icon: <WalletOutlined />,
      onClick: () => navigate('/trust'),
      color: 'var(--color-warning)',
      permission: 'trust:manage',
    },
    {
      key: 'operations-readiness',
      label: '运维准备度',
      icon: <SafetyCertificateOutlined />,
      onClick: () => navigate('/operations/readiness'),
      color: 'var(--color-warning)',
      permission: 'system:manage',
    },
    // 非 MVP 项目（MVP 模式下隐藏）
    {
      key: 'lawyer',
      label: '律师管理',
      icon: <UserOutlined />,
      onClick: () => navigate('/lawyer'),
      color: 'var(--color-success)',
      permission: 'lawyer:manage',
    },
    {
      key: 'file',
      label: '文件管理',
      icon: <FileTextOutlined />,
      onClick: () => navigate('/file'),
      color: 'var(--color-success)',
      permission: 'document:manage',
    },
    {
      key: 'tools',
      label: '工具箱',
      icon: <ToolOutlined />,
      onClick: () => navigate('/tools'),
      color: 'var(--color-text-secondary)',
      permission: 'tools:view',
    },
    {
      key: 'finance',
      label: '财务管理',
      icon: <MoneyCollectOutlined />,
      onClick: () => navigate('/finance'),
      color: 'var(--color-warning)',
      permission: 'finance:view',
    },
    {
      key: 'inbox',
      label: '待办中心',
      icon: <BellOutlined />,
      onClick: () => navigate('/inbox'),
      color: 'var(--color-text-secondary)',
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
      key: 'role',
      label: '角色管理',
      icon: <AuditOutlined />,
      onClick: () => navigate('/admin/roles'),
      color: 'var(--color-text-secondary)',
      permission: 'role:manage',
    },
    {
      key: 'permission',
      label: '权限管理',
      icon: <FileProtectOutlined />,
      onClick: () => navigate('/admin/permissions'),
      color: 'var(--color-text-secondary)',
      permission: 'permission:manage',
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

  // 根据权限和 MVP 配置过滤菜单项
  const menuItems = useMemo(() => {
    const filterMenuItems = (items: MenuItem[]): MenuItem[] => {
      return items
        .filter((item) => {
          // MVP 过滤：只保留 MVP 菜单项
          if (!isMvpMenuKey(item.key)) {
            return false
          }

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
              <span className='logo-text'>海</span>
            </div>
          ) : (
            <div className='logo-expanded'>
              <div className='logo-icon'>⚖</div>
              <div className='logo-text'>
                <div className='logo-title'>示例律师事务所OA</div>
                <div className='logo-subtitle'>DEMO LAW · OA</div>
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
              <span className='user-role'>{user?.roles?.[0] || '用户'}</span>
            </div>
          </div>
        </div>
      )}
    </Sider>
  )
}

export default Sidebar
