import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import LoginPage from '../Login'

jest.mock('@/services/auth', () => ({
  login: jest.fn(),
  setToken: jest.fn(),
}))

jest.mock('@/services/role', () => ({
  getCurrentUserPermissions: jest.fn(),
  getCurrentUserRoles: jest.fn(),
}))

jest.mock('@/stores/useAppStore', () => ({
  useAppStore: () => ({ login: jest.fn() }),
}))

describe('LoginPage keyboard accessibility', () => {
  it('provides an operable password visibility button with a synchronized name and state', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    )

    const passwordInput = screen.getByLabelText('密码')
    const showPasswordButton = screen.getByRole('button', { name: '显示密码' })
    expect(showPasswordButton).toHaveAttribute('aria-pressed', 'false')

    await user.type(passwordInput, 'secret-value')
    await user.click(showPasswordButton)

    expect(passwordInput).toHaveAttribute('type', 'text')
    expect(screen.getByRole('button', { name: '隐藏密码' })).toHaveAttribute('aria-pressed', 'true')

    await user.click(screen.getByRole('button', { name: '隐藏密码' }))
    expect(passwordInput).toHaveAttribute('type', 'password')
    expect(screen.getByRole('button', { name: '显示密码' })).toHaveAttribute('aria-pressed', 'false')
  })
})
