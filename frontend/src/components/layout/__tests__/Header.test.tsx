import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import AppHeader from '../Header'
import { getToken, logout } from '@/services/auth'

const navigateMock = jest.fn()
const clearLocalAuthMock = jest.fn()

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
    notifications: [],
    stats: { unread: 0 },
    loading: false,
    error: null,
    markAsRead: jest.fn(),
    markAllAsRead: jest.fn(),
    deleteNotification: jest.fn(),
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
