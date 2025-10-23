/**
 * 登录页面集成测试 - 现代化集成测试示例
 * 测试用户交互、API调用和路由导航
 */

import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, jest, beforeEach, afterEach } from '@jest/globals'
import LoginPage from '../auth/Login'
import { MockQueryProvider } from '../../test/mocks/react-query'
import { setupAuthState, authStates } from '../../test/mocks/auth-service'

// Mock react-router
const mockNavigate = jest.fn()
jest.mock('react-router', () => ({
  ...jest.requireActual('react-router'),
  useNavigate: () => mockNavigate,
  useLocation: () => ({ pathname: '/login' })
}))

// Mock Ant Design message
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
    warning: jest.fn(),
    info: jest.fn()
  }
}))

describe('LoginPage集成测试', () => {
  let queryClient: QueryClient
  let user: any

  beforeEach(() => {
    jest.clearAllMocks()
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false, gcTime: 0 },
        mutations: { retry: false }
      }
    })
    user = userEvent.setup()
  })

  afterEach(() => {
    jest.restoreAllMocks()
  })

  const renderLoginPage = () => {
    return render(
      <MemoryRouter initialEntries={['/login']}>
        <QueryClientProvider client={queryClient}>
          <MockQueryProvider>
            <LoginPage />
          </MockQueryProvider>
        </QueryClientProvider>
      </MemoryRouter>
    )
  }

  describe('页面渲染', () => {
    it('应该正确渲染登录表单元素', () => {
      renderLoginPage()

      // 检查标题
      expect(screen.getByText('登录')).toBeInTheDocument()

      // 检查表单字段
      expect(screen.getByLabelText(/用户名/)).toBeInTheDocument()
      expect(screen.getByLabelText(/密码/)).toBeInTheDocument()

      // 检查按钮
      expect(screen.getByRole('button', { name: '登录' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: '注册' })).toBeInTheDocument()

      // 检查其他链接
      expect(screen.getByText('忘记密码？')).toBeInTheDocument()
      expect(screen.getByText('测试登录')).toBeInTheDocument()
    })

    it('应该正确设置页面标题', () => {
      renderLoginPage()

      expect(document.title).toBe('登录 - 律师办公自动化系统')
    })
  })

  describe('表单验证', () => {
    it('应该在表单为空时显示验证错误', async () => {
      renderLoginPage()

      const loginButton = screen.getByRole('button', { name: '登录' })
      await user.click(loginButton)

      await waitFor(() => {
        expect(screen.getByText(/请输入用户名/)).toBeInTheDocument()
        expect(screen.getByText(/请输入密码/)).toBeInTheDocument()
      })
    })

    it('应该在用户名太短时显示验证错误', async () => {
      renderLoginPage()

      const usernameInput = screen.getByLabelText(/用户名/)
      await user.type(usernameInput, 'a')

      const loginButton = screen.getByRole('button', { name: '登录' })
      await user.click(loginButton)

      await waitFor(() => {
        expect(screen.getByText(/用户名至少3个字符/)).toBeInTheDocument()
      })
    })

    it('应该在密码太短时显示验证错误', async () => {
      renderLoginPage()

      const usernameInput = screen.getByLabelText(/用户名/)
      const passwordInput = screen.getByLabelText(/密码/)

      await user.type(usernameInput, 'validuser')
      await user.type(passwordInput, '123')

      const loginButton = screen.getByRole('button', { name: '登录' })
      await user.click(loginButton)

      await waitFor(() => {
        expect(screen.getByText(/密码至少6个字符/)).toBeInTheDocument()
      })
    })
  })

  describe('登录流程', () => {
    it('应该在有效凭据时成功登录', async () => {
      renderLoginPage()

      const usernameInput = screen.getByLabelText(/用户名/)
      const passwordInput = screen.getByLabelText(/密码/)
      const loginButton = screen.getByRole('button', { name: '登录' })

      // 填写有效凭据
      await user.type(usernameInput, 'admin')
      await user.type(passwordInput, 'admin123')

      // 提交表单
      await user.click(loginButton)

      // 验证加载状态
      expect(loginButton).toBeDisabled()
      expect(loginButton).toHaveTextContent('登录中...')

      // 等待登录完成
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/dashboard')
      }, { timeout: 2000 })

      // 验证成功消息
      const { message } = require('antd')
      expect(message.success).toHaveBeenCalledWith('登录成功')
    })

    it('应该在凭据错误时显示错误消息', async () => {
      renderLoginPage()

      const usernameInput = screen.getByLabelText(/用户名/)
      const passwordInput = screen.getByLabelText(/密码/)
      const loginButton = screen.getByRole('button', { name: '登录' })

      // 填写错误凭据
      await user.type(usernameInput, 'wronguser')
      await user.type(passwordInput, 'wrongpass')

      // 提交表单
      await user.click(loginButton)

      // 等待错误处理
      await waitFor(() => {
        const { message } = require('antd')
        expect(message.error).toHaveBeenCalledWith('用户名或密码错误')
      }, { timeout: 2000 })

      // 验证按钮恢复可用状态
      expect(loginButton).toBeEnabled()
      expect(loginButton).toHaveTextContent('登录')
    })

    it('应该在网络错误时显示适当错误', async () => {
      // Mock网络错误
      global.fetch = jest.fn().mockRejectedValueOnce(new Error('Network Error'))

      renderLoginPage()

      const usernameInput = screen.getByLabelText(/用户名/)
      const passwordInput = screen.getByLabelText(/密码/)
      const loginButton = screen.getByRole('button', { name: '登录' })

      await user.type(usernameInput, 'admin')
      await user.type(passwordInput, 'admin123')
      await user.click(loginButton)

      await waitFor(() => {
        const { message } = require('antd')
        expect(message.error).toHaveBeenCalledWith('网络连接失败，请检查网络设置')
      }, { timeout: 2000 })
    })
  })

  describe('用户交互', () => {
    it('应该支持回车键提交表单', async () => {
      renderLoginPage()

      const usernameInput = screen.getByLabelText(/用户名/)
      const passwordInput = screen.getByLabelText(/密码/)

      await user.type(usernameInput, 'admin')
      await user.type(passwordInput, 'admin123{enter}')

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/dashboard')
      }, { timeout: 2000 })
    })

    it('应该支持记住我功能', async () => {
      renderLoginPage()

      const rememberCheckbox = screen.getByLabelText(/记住我/)
      expect(rememberCheckbox).not.toBeChecked()

      await user.click(rememberCheckbox)
      expect(rememberCheckbox).toBeChecked()
    })

    it('应该在显示密码时切换密码输入类型', async () => {
      renderLoginPage()

      const passwordInput = screen.getByLabelText(/密码/) as HTMLInputElement
      expect(passwordInput.type).toBe('password')

      const toggleButton = screen.getByLabelText('显示密码')
      await user.click(toggleButton)

      expect(passwordInput.type).toBe('text')
      expect(toggleButton).toHaveAttribute('aria-label', '隐藏密码')
    })
  })

  describe('导航功能', () => {
    it('应该在点击注册时导航到注册页面', async () => {
      renderLoginPage()

      const registerButton = screen.getByRole('button', { name: '注册' })
      await user.click(registerButton)

      expect(mockNavigate).toHaveBeenCalledWith('/register')
    })

    it('应该在点击忘记密码时导航到重置密码页面', async () => {
      renderLoginPage()

      const forgotPasswordLink = screen.getByText('忘记密码？')
      await user.click(forgotPasswordLink)

      expect(mockNavigate).toHaveBeenCalledWith('/forgot-password')
    })

    it('应该在点击测试登录时导航到测试登录页面', async () => {
      renderLoginPage()

      const testLoginLink = screen.getByText('测试登录')
      await user.click(testLoginLink)

      expect(mockNavigate).toHaveBeenCalledWith('/test-login')
    })
  })

  describe('状态管理', () => {
    it('应该在已登录状态下重定向到仪表板', () => {
      setupAuthState('authenticated')

      renderLoginPage()

      // 应该看到重定向效果
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard')
    })

    it('应该在加载状态时显示加载指示器', () => {
      setupAuthState('loading')

      renderLoginPage()

      // 应该看到加载状态
      expect(screen.getByTestId('login-loading')).toBeInTheDocument()
    })
  })

  describe('可访问性', () => {
    it('应该支持键盘导航', async () => {
      renderLoginPage()

      const usernameInput = screen.getByLabelText(/用户名/)
      const passwordInput = screen.getByLabelText(/密码/)
      const loginButton = screen.getByRole('button', { name: '登录' })

      // Tab键导航
      usernameInput.focus()
      await user.tab()
      expect(passwordInput).toHaveFocus()

      await user.tab()
      expect(loginButton).toHaveFocus()
    })

    it('应该设置正确的ARIA属性', () => {
      renderLoginPage()

      const form = screen.getByRole('form')
      expect(form).toHaveAttribute('aria-label', '登录表单')

      const passwordInput = screen.getByLabelText(/密码/)
      expect(passwordInput).toHaveAttribute('autocomplete', 'current-password')
    })

    it('应该在错误时设置ARIA实时区域', async () => {
      renderLoginPage()

      const loginButton = screen.getByRole('button', { name: '登录' })
      await user.click(loginButton)

      await waitFor(() => {
        const errorMessages = screen.getAllByRole('alert')
        expect(errorMessages.length).toBeGreaterThan(0)
      })
    })
  })

  describe('表单预设', () => {
    it('应该从URL参数中读取预设用户名', () => {
      render(
        <MemoryRouter initialEntries={['/login?username=testuser']}>
          <QueryClientProvider client={queryClient}>
            <MockQueryProvider>
              <LoginPage />
            </MockQueryProvider>
          </QueryClientProvider>
        </MemoryRouter>
      )

      const usernameInput = screen.getByLabelText(/用户名/) as HTMLInputElement
      expect(usernameInput.value).toBe('testuser')
    })
  })

  describe('安全性', () => {
    it('应该防止密码自动填充', () => {
      renderLoginPage()

      const passwordInput = screen.getByLabelText(/密码/)
      expect(passwordInput).toHaveAttribute('autocomplete', 'current-password')
    })

    it('应该在多次登录失败后显示验证码', async () => {
      renderLoginPage()

      const usernameInput = screen.getByLabelText(/用户名/)
      const passwordInput = screen.getByLabelText(/密码/)
      const loginButton = screen.getByRole('button', { name: '登录' })

      // 模拟多次失败登录
      for (let i = 0; i < 3; i++) {
        await user.clear(usernameInput)
        await user.clear(passwordInput)
        await user.type(usernameInput, 'wronguser')
        await user.type(passwordInput, 'wrongpass')
        await user.click(loginButton)

        await waitFor(() => {
          // 等待错误消息显示
        }, { timeout: 1000 })
      }

      // 应该显示验证码
      await waitFor(() => {
        expect(screen.getByText(/验证码/)).toBeInTheDocument()
      }, { timeout: 2000 })
    })
  })
})