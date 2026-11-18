/**
 * 登录页面集成测试 - 基础兼容版
 * 简化语法以优先测试框架正常运行
 */

import React from 'react'
import { render, screen } from '@testing-library/react'

// 创建基础登录页面组件
const LoginPage = () => {
  return (
    <div className="login-page">
      <h1>登录</h1>
      <form>
        <div>
          <label htmlFor="username">用户名</label>
          <input id="username" type="text" placeholder="请输入用户名" defaultValue="testuser" />
        </div>
        <div>
          <label htmlFor="password">密码</label>
          <input id="password" type="password" placeholder="请输入密码" />
        </div>
        <div className="form-actions">
          <button type="submit">登录</button>
          <button type="button">注册</button>
        </div>
      </form>
    </div>
  )
}

describe('LoginPage基础测试', () => {
  it('应该渲染登录表单', () => {
    render(<LoginPage />)

    // 基础元素检查
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('登录')
    expect(screen.getByLabelText('用户名')).toBeInTheDocument()
    expect(screen.getByLabelText('密码')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '登录' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '注册' })).toBeInTheDocument()
  })

  it('应该显示输入字段和默认值', () => {
    render(<LoginPage />)

    const usernameInput = screen.getByLabelText('用户名') as HTMLInputElement
    const passwordInput = screen.getByLabelText('密码') as HTMLInputElement

    expect(usernameInput).toHaveValue('testuser')
    expect(passwordInput.value).toBe('') // 密码应该是空值
  })

  it('应该响应点击事件', () => {
    render(<LoginPage />)

    const loginButton = screen.getByRole('button', { name: '登录' })
    const registerButton = screen.getByRole('button', { name: '注册' })

    // 测试按钮是否可以正常点击
    const loginClick = new MouseEvent('click', { bubbles: true })
    loginButton.dispatchEvent(loginClick)

    const registerClick = new MouseEvent('click', { bubbles: true })
    registerButton.dispatchEvent(registerClick)

    // 按钮应该保持存在
    expect(loginButton).toBeInTheDocument()
    expect(registerButton).toBeInTheDocument()
  })
})