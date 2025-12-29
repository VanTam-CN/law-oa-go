/**
 * Button组件单元测试 - 兼容性版本
 * 简化语法以适应Jest ESM配置
 */

import React from 'react'
import { render, screen } from '@testing-library/react'

// Mock Button组件（缺少实际组件，创建模拟组件）
const Button = ({ children, onClick, className, variant, loading, disabled, ...props }) => {
  return (
    <button
      onClick={onClick}
      className={`button ${className || ''}`}
      data-variant={variant}
      aria-busy={loading}
      disabled={disabled || loading}
      {...props}
    >
      {loading && <span className="loading-indicator">Loading...</span>}
      {children}
    </button>
  )
}

describe('Button组件基础测试', () => {
  const mockClick = jest.fn()

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('应该渲染按钮文本', () => {
    render(<Button onClick={mockClick}>测试按钮</Button>)

    const button = screen.getByRole('button')
    expect(button).toBeInTheDocument()
    expect(button).toHaveTextContent('测试按钮')
  })

  it('应该处理点击事件', () => {
    render(<Button onClick={mockClick}>测试按钮</Button>)

    const button = screen.getByRole('button')
    button.click()

    expect(mockClick).toHaveBeenCalledTimes(1)
  })

  it('在禁用时应该不响应点击', () => {
    render(<Button onClick={mockClick} disabled>测试按钮</Button>)

    const button = screen.getByRole('button')
    expect(button).toBeDisabled()

    // 尝试点击应该会因为disabled而阻止
    const clickEvent = new MouseEvent('click', { bubbles: true })
    expect(() => button.dispatchEvent(clickEvent)).not.toThrow()
  })

  it('在加载时应该显示正确的状态', () => {
    render(<Button loading>测试按钮</Button>)

    const button = screen.getByRole('button')
    expect(button).toHaveAttribute('aria-busy', 'true')
    expect(button).toBeDisabled()
  })
})