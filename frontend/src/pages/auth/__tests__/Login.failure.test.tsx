import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { App as AntdApp } from 'antd'
import { MemoryRouter } from 'react-router'

import LoginPage from '../Login'
import { login } from '@/services/auth'

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

describe('LoginPage login failure', () => {
  it('renders a visible non-sensitive alert and preserves account case on 401', async () => {
    loginMock.mockRejectedValue({
      response: {
        status: 401,
        data: { error: { message: '邮箱或密码错误' } },
      },
    })

    render(
      <MemoryRouter>
        <AntdApp>
          <LoginPage />
        </AntdApp>
      </MemoryRouter>,
    )

    await userEvent.type(screen.getByPlaceholderText('账号或邮箱'), '  Lawyer.Wang  ')
    await userEvent.type(screen.getByPlaceholderText('密码'), 'WrongPassword123!')
    await userEvent.click(screen.getByRole('button', { name: /登\s*录/ }))

    await waitFor(() => {
      expect(loginMock).toHaveBeenCalledWith({
        account: 'Lawyer.Wang',
        password: 'WrongPassword123!',
        remember: false,
      })
    })
    expect(await screen.findByRole('alert')).toHaveTextContent(
      '账号、邮箱或密码不正确，请重新输入',
    )
    expect(screen.getByRole('alert')).not.toHaveTextContent('邮箱或密码错误')
  })

  it('sends the account field for a mixed-case username', async () => {
    loginMock.mockResolvedValue({
      token: 'test-access-token',
      user: {
        id: 1,
        username: 'Lawyer.Wang',
        name: 'Wang Lawyer',
        email: 'wang@example.test',
        role: 'lawyer',
        status: 'active',
      },
    })

    render(
      <MemoryRouter>
        <AntdApp>
          <LoginPage />
        </AntdApp>
      </MemoryRouter>,
    )

    await userEvent.type(screen.getByPlaceholderText('账号或邮箱'), ' Lawyer.Wang ')
    await userEvent.type(screen.getByPlaceholderText('密码'), 'Password123!')
    await userEvent.click(screen.getByRole('button', { name: /登\s*录/ }))

    await waitFor(() => {
      expect(loginMock).toHaveBeenCalledTimes(1)
    })
    expect(loginMock).toHaveBeenCalledWith({
      account: 'Lawyer.Wang',
      password: 'Password123!',
      remember: false,
    })
  })

  it('normalizes a mixed-case email account for the API', async () => {
    loginMock.mockResolvedValue({
      token: 'test-access-token',
      user: {
        id: 1,
        username: 'Lawyer.Wang',
        name: 'Wang Lawyer',
        email: 'wang@example.test',
        role: 'lawyer',
        status: 'active',
      },
    })

    render(
      <MemoryRouter>
        <AntdApp>
          <LoginPage />
        </AntdApp>
      </MemoryRouter>,
    )

    await userEvent.type(screen.getByPlaceholderText('账号或邮箱'), ' Wang@Example.TEST ')
    await userEvent.type(screen.getByPlaceholderText('密码'), 'Password123!')
    await userEvent.click(screen.getByRole('button', { name: /登\s*录/ }))

    await waitFor(() => {
      expect(loginMock).toHaveBeenCalledTimes(1)
    })
    expect(loginMock).toHaveBeenCalledWith({
      account: 'wang@example.test',
      password: 'Password123!',
      remember: false,
    })
  })
})
