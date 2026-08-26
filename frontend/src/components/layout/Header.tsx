import React, { useEffect, useRef, useState } from 'react'
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
import { getToken, logout as logoutRequest } from '@/services/auth'
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
  const { user, logout: clearLocalAuth } = useAppStore()
  const { notifications, stats, loading, error, markAsRead, markAllAsRead, deleteNotification } =
    useNotifications()
  const navigate = useNavigate()

  const [notificationVisible, setNotificationVisible] = useState(false)
  const [userMenuVisible, setUserMenuVisible] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const keyboardMenuActivationRef = useRef(false)
  const notificationButtonRef = useRef<HTMLButtonElement>(null)
  const userButtonRef = useRef<HTMLButtonElement>(null)
  const unreadCount = stats.unread || notifications.filter((item) => !item.isRead).length

  const handleLogout = async () => {
    const token = getToken()

    try {
      if (token) {
        await logoutRequest(token)
      }
    } catch {
      // 服务端撤销失败不阻塞本地退出
    } finally {
      clearLocalAuth()
      navigate('/login')
    }
  }

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
          content:
            '当前 MVP 试用版可在工作台发起立案、进入冲突检测、查看审批和维护客户档案。完整帮助中心建设中。',
          okText: '知道了',
        })
      },
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
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
  const handleDeleteNotification = (id: number, e: React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation()
    deleteNotification(id)
  }

  useEffect(() => {
    const syncFullscreenState = () => setIsFullscreen(Boolean(document.fullscreenElement))

    document.addEventListener('fullscreenchange', syncFullscreenState)
    syncFullscreenState()

    return () => document.removeEventListener('fullscreenchange', syncFullscreenState)
  }, [])

  useEffect(() => {
    if (!notificationVisible) {
      return
    }
    document.getElementById('header-notification-menu')?.focus()
  }, [notificationVisible])

  useEffect(() => {
    if (!userMenuVisible) {
      return
    }
    document.getElementById('header-user-menu')?.focus()
  }, [userMenuVisible])

  // 处理全屏切换。状态由 fullscreenchange 事件统一同步，避免请求失败时 UI 提前变更。
  const handleFullscreen = async () => {
    try {
      if (document.fullscreenElement) {
        await document.exitFullscreen()
      } else {
        await document.documentElement.requestFullscreen()
      }
    } catch {
      // 浏览器可能禁止全屏（例如 iframe 权限）。保持真实状态，不向用户伪造成功。
    }
  }

  // 键盘激活直接驱动受控开合，并兼容旧浏览器的按键命名。
  const handleMenuKeyDown = (
    event: React.KeyboardEvent<HTMLElement>,
    setVisible: React.Dispatch<React.SetStateAction<boolean>>,
    triggerRef: React.RefObject<HTMLButtonElement | null>,
  ) => {
    if (event.key === 'Escape') {
      event.preventDefault()
      setVisible(false)
      // rc-menu may schedule a menu-item focus after Escape. Queue the trigger
      // restoration after that library callback so keyboard users stay on the
      // toggle when the popup closes.
      requestAnimationFrame(() => triggerRef.current?.focus())
      return
    }

    const isMenuActivationKey =
      event.key === 'Enter' ||
      event.key === ' ' ||
      event.key === 'Spacebar' ||
      event.code === 'Space' ||
      event.code === 'NumpadEnter'

    if (!isMenuActivationKey) {
      return
    }

    keyboardMenuActivationRef.current = true
    event.preventDefault()
    setVisible((visible) => !visible)
    event.stopPropagation()
  }

  const handleMenuContainerKeyDown = (
    event: React.KeyboardEvent<HTMLElement>,
    setVisible: React.Dispatch<React.SetStateAction<boolean>>,
    triggerRef: React.RefObject<HTMLButtonElement | null>,
  ) => {
    if (event.key !== 'Escape') {
      return
    }

    handleMenuKeyDown(event, setVisible, triggerRef)
  }

  const handleMenuOpenChange = (
    open: boolean,
    setVisible: React.Dispatch<React.SetStateAction<boolean>>,
    triggerRef: React.RefObject<HTMLButtonElement | null>,
  ) => {
    setVisible(open)

    if (open) {
      return
    }

    // Dropdown unmounts its menu after outside-click dismissal. Restore focus on
    // the next frame only when the browser has dropped it to the document body,
    // so keyboard users do not lose their place and direct focus elsewhere is
    // respected.
    requestAnimationFrame(() => {
      if (document.activeElement === document.body) {
        triggerRef.current?.focus()
      }
    })
  }

  const handleMenuClickCapture = (event: React.MouseEvent<HTMLButtonElement>) => {
    if (!keyboardMenuActivationRef.current) {
      return
    }

    // Space 在 keyup 后才生成 click；避免 Dropdown 对同一次键盘激活做第二次开合。
    keyboardMenuActivationRef.current = false
    event.preventDefault()
    event.stopPropagation()
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
              <button type='button' className='action-link' onClick={handleMarkAllAsRead}>
                全部已读
              </button>
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
                  <div className={`notification-item ${notification.isRead ? 'read' : 'unread'}`}>
                    <button
                      type='button'
                      className='notification-content'
                      aria-label={`${notification.isRead ? '已读' : '未读'}通知：${notification.title}`}
                      onClick={() => handleNotificationClick(notification)}
                    >
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
                    </button>
                    <div className='notification-actions'>
                      <button
                        type='button'
                        className='delete-btn'
                        aria-label={`删除通知：${notification.title}`}
                        onClick={(e) => handleDeleteNotification(notification.id, e)}
                      >
                        <DeleteOutlined />
                      </button>
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
            label: (
              <button type='button' className='view-all-btn'>
                查看全部通知
              </button>
            ),
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
          <button
            type='button'
            className='header-action fullscreen-btn'
            onClick={handleFullscreen}
            aria-pressed={isFullscreen}
            aria-label={isFullscreen ? '退出全屏' : '进入全屏'}
            title={isFullscreen ? '退出全屏' : '全屏模式'}
          >
            {isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
          </button>

          {/* 通知中心 */}
          <Dropdown
            menu={{
              items: notificationItems,
              id: 'header-notification-menu',
              'aria-label': '通知中心',
              onKeyDown: (event) =>
                handleMenuContainerKeyDown(event, setNotificationVisible, notificationButtonRef),
              onClick: ({ key }) => {
                if (key === 'view-all') {
                  navigate('/notifications')
                }
              },
            }}
            placement='bottomRight'
            transitionName=''
            onOpenChange={(open) =>
              handleMenuOpenChange(open, setNotificationVisible, notificationButtonRef)
            }
            open={notificationVisible}
            trigger={['click']}
            destroyOnHidden
          >
            <button
              ref={notificationButtonRef}
              type='button'
              className={`header-action notification-btn ${notificationVisible ? 'active' : ''}`}
              aria-expanded={notificationVisible}
              aria-haspopup='menu'
              aria-controls='header-notification-menu'
              aria-label={`通知中心${unreadCount > 0 ? `，${unreadCount} 条未读` : ''}`}
              title='通知中心'
              onKeyDown={(event) =>
                handleMenuKeyDown(event, setNotificationVisible, notificationButtonRef)
              }
              onClickCapture={handleMenuClickCapture}
            >
              <Badge
                count={unreadCount > 0 ? unreadCount : 0}
                size='small'
                className='notification-badge'
                overflowCount={99}
              >
                <BellOutlined className='action-icon' />
              </Badge>
            </button>
          </Dropdown>

          {/* 用户菜单 */}
          <Dropdown
            menu={{
              items: userMenuItems,
              id: 'header-user-menu',
              'aria-label': '用户菜单',
              onKeyDown: (event) =>
                handleMenuContainerKeyDown(event, setUserMenuVisible, userButtonRef),
            }}
            placement='bottomRight'
            transitionName=''
            onOpenChange={(open) => handleMenuOpenChange(open, setUserMenuVisible, userButtonRef)}
            open={userMenuVisible}
            trigger={['click']}
            destroyOnHidden
          >
            <button
              ref={userButtonRef}
              type='button'
              className='user-menu'
              aria-expanded={userMenuVisible}
              aria-haspopup='menu'
              aria-controls='header-user-menu'
              aria-label={`用户菜单：${user?.realName || user?.username || '用户'}`}
              onKeyDown={(event) => handleMenuKeyDown(event, setUserMenuVisible, userButtonRef)}
              onClickCapture={handleMenuClickCapture}
            >
              <Avatar size='small' icon={<UserOutlined />} className='user-avatar' />
              <span className='user-name'>{user?.realName || user?.username || '用户'}</span>
            </button>
          </Dropdown>
        </Space>
      </div>
    </Header>
  )
}

export default AppHeader
