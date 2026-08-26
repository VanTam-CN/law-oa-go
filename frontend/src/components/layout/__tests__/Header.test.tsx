import React from 'react'
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import AppHeader from '../Header'
import { getToken, logout } from '@/services/auth'

const navigateMock = jest.fn()
const clearLocalAuthMock = jest.fn()
const markAllAsReadMock = jest.fn()
const deleteNotificationMock = jest.fn()
const markAsReadMock = jest.fn()

let notificationState = {
  notifications: [] as Array<{
    id: number
    title: string
    content: string
    type: string
    isRead: boolean
    createdAt: string
  }>,
  unread: 0,
}

jest.mock('react-router', () => {
  const actual = jest.requireActual<typeof import('react-router')>('react-router')
  return {
    ...actual,
    useNavigate: () => navigateMock,
  }
})

jest.mock('@/stores/useAppStore', () => ({
  useAppStore: () => ({
    user: { realName: '测试主任' },
    logout: clearLocalAuthMock,
  }),
}))

jest.mock('@/hooks/useNotifications', () => ({
  __esModule: true,
  default: () => ({
    notifications: notificationState.notifications,
    stats: { unread: notificationState.unread },
    loading: false,
    error: null,
    markAsRead: markAsReadMock,
    markAllAsRead: markAllAsReadMock,
    deleteNotification: deleteNotificationMock,
  }),
}))

jest.mock('@/services/auth', () => ({
  getToken: jest.fn(),
  logout: jest.fn(),
}))

const getTokenMock = getToken as jest.MockedFunction<typeof getToken>
const logoutMockFn = logout as jest.MockedFunction<typeof logout>

async function openUserMenuAndLogout() {
  render(
    <MemoryRouter>
      <AppHeader />
    </MemoryRouter>,
  )

  await userEvent.click(screen.getByText('测试主任'))
  await userEvent.click(screen.getByText('退出登录'))
}

function renderHeader() {
  return render(
    <MemoryRouter>
      <AppHeader />
    </MemoryRouter>,
  )
}

describe('AppHeader logout', () => {
  beforeEach(() => {
    getTokenMock.mockReturnValue('current-access-token')
    navigateMock.mockClear()
    clearLocalAuthMock.mockClear()
    logoutMockFn.mockReset()
  })

  it('revokes the token and clears local auth', async () => {
    logoutMockFn.mockResolvedValue(undefined)

    await openUserMenuAndLogout()

    await waitFor(() => {
      expect(logoutMockFn).toHaveBeenCalledWith('current-access-token')
      expect(clearLocalAuthMock).toHaveBeenCalledTimes(1)
      expect(navigateMock).toHaveBeenCalledWith('/login')
    })
  })

  it('still clears local auth when server revocation fails', async () => {
    logoutMockFn.mockRejectedValue(new Error('network unavailable'))

    await openUserMenuAndLogout()

    await waitFor(() => {
      expect(clearLocalAuthMock).toHaveBeenCalledTimes(1)
      expect(navigateMock).toHaveBeenCalledWith('/login')
    })
  })
})

describe('AppHeader keyboard accessibility', () => {
  const requestFullscreenMock = jest.fn(() => Promise.resolve())

  beforeEach(() => {
    notificationState = { notifications: [], unread: 0 }
    navigateMock.mockClear()
    markAllAsReadMock.mockClear()
    deleteNotificationMock.mockClear()
    markAsReadMock.mockClear()
    Object.defineProperty(document, 'fullscreenElement', {
      configurable: true,
      value: null,
    })
    document.documentElement.requestFullscreen = requestFullscreenMock
  })

  it('exposes the three header controls as named buttons in tab order', () => {
    renderHeader()

    const fullscreenButton = screen.getByRole('button', { name: '进入全屏' })
    const notificationButton = screen.getByRole('button', { name: '通知中心' })
    const userButton = screen.getByRole('button', { name: '用户菜单：测试主任' })

    expect(fullscreenButton).toHaveAttribute('type', 'button')
    expect(fullscreenButton).toHaveAttribute('aria-pressed', 'false')
    expect(notificationButton).toHaveAttribute('aria-haspopup', 'menu')
    expect(notificationButton).toHaveAttribute('aria-expanded', 'false')
    expect(userButton).toHaveAttribute('aria-haspopup', 'menu')
  })

  it('activates fullscreen from the keyboard', async () => {
    const user = userEvent.setup()
    renderHeader()

    const fullscreenButton = screen.getByRole('button', { name: '进入全屏' })
    fullscreenButton.focus()
    expect(screen.getByRole('button', { name: '进入全屏' })).toHaveFocus()

    await user.keyboard('{Enter}')
    expect(requestFullscreenMock).toHaveBeenCalledTimes(1)
  })

  it('opens the notification menu from the keyboard and marks all notifications read', async () => {
    notificationState = {
      notifications: [
        {
          id: 1,
          title: '审批待处理',
          content: '请复核新案件',
          type: 'approval',
          isRead: false,
          createdAt: new Date().toISOString(),
        },
      ],
      unread: 1,
    }
    const user = userEvent.setup()
    renderHeader()
    const notificationButton = screen.getByRole('button', {
      name: '通知中心，1 条未读',
    })

    await user.click(notificationButton)
    expect(notificationButton).toHaveAttribute('aria-expanded', 'true')

    const markAllButton = await screen.findByRole('button', { name: '全部已读' })
    act(() => {
      markAllButton.focus()
    })
    await waitFor(() => expect(markAllButton).toHaveFocus())
    await user.keyboard('{Enter}')
    expect(markAllAsReadMock).toHaveBeenCalledTimes(1)
  })

  it('provides stable actions for reading, deleting, and viewing notifications', async () => {
    notificationState = {
      notifications: [
        {
          id: 2,
          title: '系统维护通知',
          content: '今晚维护',
          type: 'system',
          isRead: false,
          createdAt: new Date().toISOString(),
        },
      ],
      unread: 1,
    }
    const user = userEvent.setup()
    renderHeader()

    await user.click(screen.getByRole('button', { name: '通知中心，1 条未读' }))
    const notificationButton = await screen.findByRole('button', { name: '未读通知：系统维护通知' })
    await user.click(notificationButton)
    expect(markAsReadMock).toHaveBeenCalledWith(2)

    const deleteButton = screen.getByRole('button', { name: '删除通知：系统维护通知' })
    await user.click(deleteButton)
    expect(deleteNotificationMock).toHaveBeenCalledWith(2)

    const viewAllButton = screen.getByRole('button', { name: '查看全部通知' })
    await user.click(viewAllButton)
    expect(navigateMock).toHaveBeenCalledWith('/notifications')
  })
})
