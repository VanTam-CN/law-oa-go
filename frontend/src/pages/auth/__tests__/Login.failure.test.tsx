import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { App as AntdApp } from 'antd'
import { MemoryRouter } from 'react-router'

import LoginPage from '../Login'
import { login } from '@/services/auth'
import { message } from '@/utils/messageHelper'

jest.mock('@/services/auth', () => ({
  login: jest.fn(),
  setToken: jest.fn(),
}))

jest.mock('@/services/role', () => ({
  getCurrentUserPermissions: jest.fn().mockResolvedValue([]),
  getCurrentUserRoles: jest.fn().mockResolvedValue([]),
}))

jest.mock('@/utils/messageHelper', () => ({
  message: {
    error: jest.fn(),
    success: jest.fn(),
  },
}))

const loginMock = login as jest.MockedFunction<typeof login>
const messageErrorMock = message.error as jest.MockedFunction<typeof message.error>

describe('LoginPage login failure', () => {
  it('shows a non-sensitive message and sends account when login returns 401', async () => {
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

    await userEvent.type(screen.getByPlaceholderText('账号或邮箱'), 'Lawyer.Wang')
    await userEvent.type(screen.getByPlaceholderText('密码'), 'WrongPassword123!')
    await userEvent.click(screen.getByRole('button', { name: /登\s*录/ }))

    await waitFor(() => {
      expect(loginMock).toHaveBeenCalledWith({
        account: 'lawyer.wang',
        password: 'WrongPassword123!',
        remember: false,
      })
    })
    expect(messageErrorMock).toHaveBeenCalledWith('账号、邮箱或密码不正确，请重新输入')
    expect(messageErrorMock).not.toHaveBeenCalledWith('邮箱或密码错误')
  })
})
