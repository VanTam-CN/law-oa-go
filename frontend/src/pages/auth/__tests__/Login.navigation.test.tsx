import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { App as AntdApp } from 'antd'
import { MemoryRouter, Routes, Route } from 'react-router'

import LoginPage from '../Login'
import { login } from '@/services/auth'

const navigateMock = jest.fn()

jest.mock('react-router', () => ({
  ...jest.requireActual('react-router'),
  useNavigate: () => navigateMock,
}))

jest.mock('@/services/auth', () => ({
  login: jest.fn(),
  setToken: jest.fn(),
}))

jest.mock('@/services/role', () => ({
  getCurrentUserPermissions: jest.fn().mockResolvedValue([]),
  getCurrentUserRoles: jest.fn().mockResolvedValue([]),
}))

jest.mock('@/stores/useAppStore', () => ({
  useAppStore: () => ({ login: jest.fn() }),
}))

const loginMock = login as jest.MockedFunction<typeof login>

describe('LoginPage navigation', () => {
  it('登录成功后立即导航，无 500ms 延迟', async () => {
    jest.useFakeTimers()
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
    loginMock.mockResolvedValue({
      token: 'fake-token',
      user: { id: 7, username: 'lawyer', status: 'active' },
    } as any)
    render(
      <MemoryRouter initialEntries={['/login']}>
        <AntdApp>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<div>首页</div>} />
          </Routes>
        </AntdApp>
      </MemoryRouter>,
    )
    await user.type(screen.getByPlaceholderText('账号或邮箱'), 'lawyer')
    await user.type(screen.getByPlaceholderText('密码'), 'Secret123!')
    await user.click(screen.getByRole('button', { name: /登\s*录/ }))
    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith('/')
    })
    // The old build waited 500ms before redirecting; assert no timer remains.
    jest.runAllTimers()
    expect(navigateMock).toHaveBeenCalledTimes(1)
    jest.useRealTimers()
  })
})
