import React, { useState } from 'react'
import { Layout, Dropdown, Space, Badge, Avatar, Menu, Spin, Empty, Modal } from 'antd'
import {
  BellOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined,
  CheckOutlined,
  DeleteOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons'
import { useAppStore } from '@/stores/useAppStore'
import useNotifications from '@/hooks/useNotifications'
import { useNavigate } from 'react-router'
import type { MenuProps } from 'antd'
import './header.less'

const { Header } = Layout

interface Notification {
  id: number
  title: string
  content: string
  type: string
  isRead: boolean
  createdAt: string
}

const AppHeader: React.FC = () => {
  const { user, logout } = useAppStore()
  const { notifications, stats, loading, error, markAsRead, markAllAsRead, deleteNotification } =
    useNotifications()
  const navigate = useNavigate()

  const [notificationVisible, setNotificationVisible] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const unreadCount = stats.unread || notifications.filter((item) => !item.isRead).length

  // 用户菜单项
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人中心',
      onClick: () => navigate('/profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
      onClick: () => navigate('/settings'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'help',
      icon: <QuestionCircleOutlined />,
      label: '帮助中心',
      onClick: () => {
        Modal.info({
          title: '帮助中心',
          content: '当前 MVP 试用版可在工作台发起立案、进入冲突检测、查看审批和维护客户档案。完整帮助中心建设中。',
          okText: '知道了',
        })
      },
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: logout,
    },
  ]

  // 格式化通知时间
  const formatTime = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(diff / 3600000)
    const days = Math.floor(diff / 86400000)

    if (minutes < 1) {
      return '刚刚'
    }
    if (minutes < 60) {
      return `${minutes}分钟前`
    }
    if (hours < 24) {
      return `${hours}小时前`
    }
    if (days < 7) {
      return `${days}天前`
    }
    return date.toLocaleDateString()
  }

  // 获取通知类型配置
  const getNotificationConfig = (type: string) => {
    const configs = {
      approval: {
        icon: <CheckOutlined />,
        color: 'var(--color-success)',
        bgColor: 'var(--color-success-100)',
        label: '审批',
      },
      project: {
        icon: <BellOutlined />,
        color: 'var(--color-primary)',
        bgColor: 'var(--color-primary-100)',
        label: '项目',
      },
      system: {
        icon: <SettingOutlined />,
        color: 'var(--color-warning)',
        bgColor: 'var(--color-warning-100)',
        label: '系统',
      },
      finance: {
        icon: <UserOutlined />,
        color: 'var(--color-error)',
        bgColor: 'var(--color-error-100)',
        label: '财务',
      },
      default: {
        icon: <BellOutlined />,
        color: 'var(--color-text-secondary)',
        bgColor: 'var(--color-gray-100)',
        label: '通知',
      },
    }
    return configs[type as keyof typeof configs] || configs.default
  }

  // 处理通知项点击
  const handleNotificationClick = (notification: Notification) => {
    if (!notification.isRead) {
      markAsRead(notification.id)
    }
  }

  // 处理标记全部已读
  const handleMarkAllAsRead = () => {
    if (unreadCount <= 0) {
      setNotificationVisible(false)
      return
    }
    markAllAsRead()
    setNotificationVisible(false)
  }

  // 处理删除通知
  const handleDeleteNotification = (id: number, e: React.MouseEvent) => {
    e.stopPropagation()
    deleteNotification(id)
  }

  // 处理全屏切换
  const handleFullscreen = () => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen()
      setIsFullscreen(true)
    } else {
      document.exitFullscreen()
      setIsFullscreen(false)
    }
  }

  // 通知菜单项
  const notificationItems: MenuProps['items'] = [
    {
      key: 'header',
      label: (
        <div className='notification-header'>
          <div className='notification-title'>
            <span>通知中心</span>
            {unreadCount > 0 && <span className='unread-count'>{unreadCount} 未读</span>}
          </div>
          {unreadCount > 0 && (
            <div className='notification-actions'>
              <span className='action-link' onClick={handleMarkAllAsRead}>
                全部已读
              </span>
            </div>
          )}
        </div>
      ),
    },
    {
      type: 'divider' as const,
    },
    ...(loading
      ? [
          {
            key: 'loading',
            label: (
              <div className='notification-loading'>
                <Spin size='small' />
                <span>加载中...</span>
              </div>
            ),
          },
        ]
      : error
        ? [
            {
              key: 'error',
              label: (
                <div className='notification-error'>
                  <Empty description={error} />
                </div>
              ),
            },
          ]
        : notifications.length === 0
          ? [
              {
                key: 'empty',
                label: (
                  <div className='notification-empty'>
                    <Empty description='暂无通知' />
                  </div>
                ),
              },
            ]
          : notifications.map((notification: Notification) => {
              const config = getNotificationConfig(notification.type)
              return {
                key: `notification-${notification.id}`,
                label: (
                  <div
                    className={`notification-item ${notification.isRead ? 'read' : 'unread'}`}
                    onClick={() => handleNotificationClick(notification)}
                  >
                    <div className='notification-content'>
                      <div className='notification-meta'>
                        <div
                          className='notification-type'
                          style={{ backgroundColor: config.bgColor, color: config.color }}
                        >
                          {config.icon}
                          <span>{config.label}</span>
                        </div>
                        <span className='notification-time'>
                          {formatTime(notification.createdAt)}
                        </span>
                      </div>
                      <div className='notification-title-text'>{notification.title}</div>
                      <div className='notification-description'>{notification.content}</div>
                    </div>
                    <div className='notification-actions'>
                      <DeleteOutlined
                        className='delete-btn'
                        onClick={(e) => handleDeleteNotification(notification.id, e)}
                      />
                    </div>
                  </div>
                ),
              }
            })),
    ...(notifications.length > 0
      ? [
          {
            type: 'divider' as const,
          },
          {
            key: 'view-all',
            label: <div className='view-all-btn'>查看全部通知</div>,
          },
        ]
      : []),
  ]

  return (
    <Header className='app-header'>
      <div className='header-left'>{/* 面包屑或其他左侧内容可以在这里添加 */}</div>

      <div className='header-right'>
        <Space size='large'>
          {/* 全屏切换 */}
          <div
            className='header-action fullscreen-btn'
            onClick={handleFullscreen}
            title={isFullscreen ? '退出全屏' : '全屏模式'}
          >
            {isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
          </div>

          {/* 通知中心 */}
          <Dropdown
            menu={{
              items: notificationItems,
              onClick: ({ key }) => {
                if (key === 'view-all') {
                  navigate('/notifications')
                }
              },
            }}
            placement='bottomRight'
            onOpenChange={setNotificationVisible}
            open={notificationVisible}
            trigger={['click']}
          >
            <div
              className={`header-action notification-btn ${notificationVisible ? 'active' : ''}`}
              title='通知中心'
            >
              <Badge
                count={unreadCount > 0 ? unreadCount : 0}
                size='small'
                className='notification-badge'
                overflowCount={99}
              >
                <BellOutlined className='action-icon' />
              </Badge>
            </div>
          </Dropdown>

          {/* 用户菜单 */}
          <Dropdown menu={{ items: userMenuItems }} placement='bottomRight' trigger={['click']}>
            <div className='user-menu'>
              <Avatar size='small' icon={<UserOutlined />} className='user-avatar' />
              <span className='user-name'>{user?.realName || user?.username || '用户'}</span>
            </div>
          </Dropdown>
        </Space>
      </div>
    </Header>
  )
}

export default AppHeader
