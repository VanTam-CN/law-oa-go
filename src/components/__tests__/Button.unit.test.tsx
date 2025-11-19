/**
 * Button组件单元测试 - 现代化React Testing Library示例
 * 基于Jest 30.2和React Testing Library v16.3.0
 */

import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, jest, beforeEach, afterEach } from '@jest/globals'
import Button from '../Button'

// Mock样式模块
jest.mock('../Button.module.less', () => ({
  button: 'mock-button-class',
  primary: 'mock-primary-class',
  loading: 'mock-loading-class',
  disabled: 'mock-disabled-class'
}))

describe('Button组件', () => {
  const defaultProps = {
    children: '测试按钮',
    onClick: jest.fn()
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  afterEach(() => {
    jest.restoreAllMocks()
  })

  describe('基础渲染', () => {
    it('应该正确渲染按钮文本', () => {
      render(<Button {...defaultProps} />)

      const button = screen.getByRole('button', { name: '测试按钮' })
      expect(button).toBeInTheDocument()
      expect(button).toHaveTextContent('测试按钮')
    })

    it('应该应用正确的CSS类', () => {
      render(<Button {...defaultProps} />)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('mock-button-class')
    })

    it('应该支持自定义className', () => {
      render(<Button {...defaultProps} className="custom-class" />)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('mock-button-class')
      expect(button).toHaveClass('custom-class')
    })
  })

  describe('点击事件处理', () => {
    it('应该在点击时调用onClick处理器', async () => {
      const user = userEvent.setup()
      render(<Button {...defaultProps} />)

      const button = screen.getByRole('button')
      await user.click(button)

      expect(defaultProps.onClick).toHaveBeenCalledTimes(1)
    })

    it('应该在键盘回车时触发点击', async () => {
      const user = userEvent.setup()
      render(<Button {...defaultProps} />)

      const button = screen.getByRole('button')
      button.focus()
      await user.keyboard('{Enter}')

      expect(defaultProps.onClick).toHaveBeenCalledTimes(1)
    })

    it('应该在空格键按下时触发点击', async () => {
      const user = userEvent.setup()
      render(<Button {...defaultProps} />)

      const button = screen.getByRole('button')
      button.focus()
      await user.keyboard(' ')

      expect(defaultProps.onClick).toHaveBeenCalledTimes(1)
    })
  })

  describe('变体样式', () => {
    it('应该在primary类型时应用primary样式', () => {
      render(<Button {...defaultProps} variant="primary" />)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('mock-primary-class')
    })

    it('应该在danger类型时应用danger样式', () => {
      render(<Button {...defaultProps} variant="danger" />)

      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-variant', 'danger')
    })

    it('应该在ghost类型时应用ghost样式', () => {
      render(<Button {...defaultProps} variant="ghost" />)

      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-variant', 'ghost')
    })
  })

  describe('禁用状态', () => {
    it('应该在disabled属性为true时禁用按钮', () => {
      render(<Button {...defaultProps} disabled />)

      const button = screen.getByRole('button')
      expect(button).toBeDisabled()
      expect(button).toHaveClass('mock-disabled-class')
    })

    it('应该在禁用状态下不触发点击事件', async () => {
      const user = userEvent.setup()
      render(<Button {...defaultProps} disabled />)

      const button = screen.getByRole('button')
      await user.click(button)

      expect(defaultProps.onClick).not.toHaveBeenCalled()
    })

    it('应该在loading状态下禁用按钮', () => {
      render(<Button {...defaultProps} loading />)

      const button = screen.getByRole('button')
      expect(button).toBeDisabled()
      expect(button).toHaveClass('mock-loading-class')
    })
  })

  describe('加载状态', () => {
    it('应该在loading时显示加载指示器', () => {
      render(<Button {...defaultProps} loading />)

      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('aria-busy', 'true')

      // 检查是否有加载指示器
      const loadingIcon = button.querySelector('.anticon-loading')
      expect(loadingIcon).toBeInTheDocument()
    })

    it('应该在loading时不触发点击事件', async () => {
      const user = userEvent.setup()
      render(<Button {...defaultProps} loading />)

      const button = screen.getByRole('button')
      await user.click(button)

      expect(defaultProps.onClick).not.toHaveBeenCalled()
    })
  })

  describe('尺寸变体', () => {
    it('应该在small尺寸时应用small样式', () => {
      render(<Button {...defaultProps} size="small" />)

      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-size', 'small')
    })

    it('应该在large尺寸时应用large样式', () => {
      render(<Button {...defaultProps} size="large" />)

      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-size', 'large')
    })
  })

  describe('可访问性', () => {
    it('应该支持自定义aria-label', () => {
      render(
        <Button {...defaultProps} aria-label="自定义标签">
          测试按钮
        </Button>
      )

      const button = screen.getByRole('button', { name: '自定义标签' })
      expect(button).toBeInTheDocument()
    })

    it('应该支持aria-describedby', () => {
      const descriptionId = 'button-description'
      render(
        <div>
          <div id={descriptionId}>按钮描述</div>
          <Button {...defaultProps} aria-describedby={descriptionId} />
        </div>
      )

      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('aria-describedby', descriptionId)
    })

    it('应该在loading时设置aria-busy', () => {
      render(<Button {...defaultProps} loading />)

      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('aria-busy', 'true')
    })
  })

  describe('表单集成', () => {
    it('应该在表单中正常工作', () => {
      const handleSubmit = jest.fn(e => e.preventDefault())

      render(
        <form onSubmit={handleSubmit}>
          <Button {...defaultProps} type="submit" />
        </form>
      )

      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('type', 'submit')
    })

    it('应该支持自定义type属性', () => {
      render(<Button {...defaultProps} type="reset" />)

      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('type', 'reset')
    })
  })

  describe('图标支持', () => {
    it('应该支持前置图标', () => {
      const IconComponent = () => <span data-testid="icon">Icon</span>

      render(<Button {...defaultProps} icon={<IconComponent />} />)

      const icon = screen.getByTestId('icon')
      expect(icon).toBeInTheDocument()

      const button = screen.getByRole('button')
      expect(button).toContainElement(icon)
    })

    it('应该支持后置图标', () => {
      const IconComponent = () => <span data-testid="icon">Icon</span>

      render(<Button {...defaultProps} suffix={<IconComponent />} />)

      const icon = screen.getByTestId('icon')
      expect(icon).toBeInTheDocument()

      const button = screen.getByRole('button')
      expect(button).toContainElement(icon)
    })
  })

  describe('异步操作', () => {
    it('应该支持异步onClick处理器', async () => {
      const asyncOnClick = jest.fn(async () => {
        await new Promise(resolve => setTimeout(resolve, 100))
      })

      const user = userEvent.setup()
      render(<Button {...defaultProps} onClick={asyncOnClick} />)

      const button = screen.getByRole('button')
      await user.click(button)

      expect(asyncOnClick).toHaveBeenCalled()
    })

    it('应该在异步操作期间保持按钮可用性', async () => {
      let resolveAsync: () => void
      const asyncOnClick = jest.fn(() => {
        return new Promise<void>(resolve => {
          resolveAsync = resolve
        })
      })

      const user = userEvent.setup()
      render(<Button {...defaultProps} onClick={asyncOnClick} />)

      const button = screen.getByRole('button')

      // 第一次点击
      await user.click(button)
      expect(asyncOnClick).toHaveBeenCalledTimes(1)

      // 异步操作未完成，按钮仍应可点击
      await user.click(button)
      expect(asyncOnClick).toHaveBeenCalledTimes(2)

      // 完成异步操作
      if (resolveAsync) {resolveAsync()}
    })
  })

  describe('错误处理', () => {
    it('应该处理onClick中的错误', async () => {
      const consoleError = jest.spyOn(console, 'error').mockImplementation(() => {})
      const errorCallback = jest.fn(() => {
        throw new Error('Click error')
      })

      const user = userEvent.setup()
      render(<Button {...defaultProps} onClick={errorCallback} />)

      const button = screen.getByRole('button')

      expect(() => user.click(button)).not.toThrow()

      await waitFor(() => {
        expect(errorCallback).toHaveBeenCalled()
      })

      consoleError.mockRestore()
    })
  })

  describe('ref支持', () => {
    it('应该支持forwardRef', () => {
      const ref = { current: null }

      render(<Button {...defaultProps} ref={ref} />)

      expect(ref.current).toBeInstanceOf(HTMLButtonElement)
    })
  })

  describe('children类型', () => {
    it('应该支持字符串children', () => {
      render(<Button {...defaultProps}>字符串按钮</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveTextContent('字符串按钮')
    })

    it('应该支持React节点children', () => {
      render(
        <Button {...defaultProps}>
          <span>React节点</span>
        </Button>
      )

      const button = screen.getByRole('button')
      const span = button.querySelector('span')
      expect(span).toBeInTheDocument()
      expect(span).toHaveTextContent('React节点')
    })

    it('应该支持数字children', () => {
      render(<Button {...defaultProps}>{123}</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveTextContent('123')
    })
  })
})