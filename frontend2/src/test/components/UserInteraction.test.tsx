import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'

// Mock登录组件
const LoginComponent = ({ onLogin }) => {
  const [username, setUsername] = React.useState('')
  const [password, setPassword] = React.useState('')
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState('')

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      await onLogin(username, password)
    } catch (err) {
      setError('登录失败，请检查用户名和密码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} data-testid="login-form">
      <div>
        <label htmlFor="username">用户名</label>
        <input
          id="username"
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          data-testid="username-input"
          required
        />
      </div>
      <div>
        <label htmlFor="password">密码</label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          data-testid="password-input"
          required
        />
      </div>
      {error && <div data-testid="error-message">{error}</div>}
      <button type="submit" disabled={loading} data-testid="submit-button">
        {loading ? '登录中...' : '登录'}
      </button>
    </form>
  )
}

// Mock表格组件
const TableComponent = ({ data, onRowClick, onDelete }) => {
  return (
    <table data-testid="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>名称</th>
          <th>状态</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        {data.map((item) => (
          <tr key={item.id} onClick={() => onRowClick(item)} data-testid={`row-${item.id}`}>
            <td>{item.id}</td>
            <td>{item.name}</td>
            <td>{item.status}</td>
            <td>
              <button onClick={(e) => {
                e.stopPropagation()
                onDelete(item.id)
              }} data-testid={`delete-${item.id}`}>
                删除
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

// Mock模态框组件
const ModalComponent = ({ isOpen, onClose, onConfirm, title, children }) => {
  if (!isOpen) return null

  const handleKeyDown = (event) => {
    if (event.key === 'Escape') {
      onClose()
    }
  }

  return (
    <div 
      data-testid="modal-overlay" 
      tabIndex={0}
      onKeyDown={handleKeyDown}
    >
      <div data-testid="modal-content">
        <h2 data-testid="modal-title">{title}</h2>
        <div data-testid="modal-body">
          {children}
        </div>
        <div data-testid="modal-actions">
          <button onClick={onClose} data-testid="cancel-button">取消</button>
          <button onClick={onConfirm} data-testid="confirm-button">确认</button>
        </div>
      </div>
    </div>
  )
}

describe('User Interaction Tests', () => {
  let user

  beforeEach(() => {
    user = userEvent.setup()
    vi.clearAllMocks()
  })

  describe('Login Form Interactions', () => {
    it('should handle form submission with valid data', async () => {
      const mockLogin = vi.fn().mockResolvedValue({ success: true })
      
      render(<LoginComponent onLogin={mockLogin} />)
      
      const usernameInput = screen.getByTestId('username-input')
      const passwordInput = screen.getByTestId('password-input')
      const submitButton = screen.getByTestId('submit-button')
      
      await user.type(usernameInput, 'admin')
      await user.type(passwordInput, 'password123')
      await user.click(submitButton)
      
      expect(mockLogin).toHaveBeenCalledWith('admin', 'password123')
      expect(mockLogin).toHaveBeenCalledTimes(1)
    })

    it('should show loading state during submission', async () => {
      const mockLogin = vi.fn().mockImplementation(() => 
        new Promise(resolve => setTimeout(resolve, 1000))
      )
      
      render(<LoginComponent onLogin={mockLogin} />)
      
      const usernameInput = screen.getByTestId('username-input')
      const passwordInput = screen.getByTestId('password-input')
      const submitButton = screen.getByTestId('submit-button')
      
      await user.type(usernameInput, 'admin')
      await user.type(passwordInput, 'password123')
      await user.click(submitButton)
      
      expect(screen.getByTestId('submit-button')).toHaveTextContent('登录中...')
      expect(screen.getByTestId('submit-button')).toBeDisabled()
      
      await waitFor(() => {
        expect(screen.getByTestId('submit-button')).toHaveTextContent('登录')
      }, { timeout: 1500 })
    })

    it('should show error message on login failure', async () => {
      const mockLogin = vi.fn().mockRejectedValue(new Error('Invalid credentials'))
      
      render(<LoginComponent onLogin={mockLogin} />)
      
      const usernameInput = screen.getByTestId('username-input')
      const passwordInput = screen.getByTestId('password-input')
      const submitButton = screen.getByTestId('submit-button')
      
      await user.type(usernameInput, 'admin')
      await user.type(passwordInput, 'wrongpassword')
      await user.click(submitButton)
      
      await waitFor(() => {
        expect(screen.getByTestId('error-message')).toBeInTheDocument()
        expect(screen.getByTestId('error-message')).toHaveTextContent('登录失败，请检查用户名和密码')
      })
    })

    it('should validate required fields', async () => {
      const mockLogin = vi.fn()
      
      render(<LoginComponent onLogin={mockLogin} />)
      
      const submitButton = screen.getByTestId('submit-button')
      
      await user.click(submitButton)
      
      expect(mockLogin).not.toHaveBeenCalled()
      
      const usernameInput = screen.getByTestId('username-input')
      const passwordInput = screen.getByTestId('password-input')
      
      expect(usernameInput).toBeInvalid()
      expect(passwordInput).toBeInvalid()
    })

    it('should handle keyboard navigation', async () => {
      const mockLogin = vi.fn().mockResolvedValue({ success: true })
      
      render(<LoginComponent onLogin={mockLogin} />)
      
      const usernameInput = screen.getByTestId('username-input')
      const passwordInput = screen.getByTestId('password-input')
      
      await user.type(usernameInput, 'admin')
      await user.tab()
      await user.type(passwordInput, 'password123')
      await user.tab()
      await user.keyboard('{Enter}')
      
      expect(mockLogin).toHaveBeenCalledWith('admin', 'password123')
    })
  })

  describe('Table Interactions', () => {
    const mockData = [
      { id: 1, name: '案件1', status: '进行中' },
      { id: 2, name: '案件2', status: '已完成' },
      { id: 3, name: '案件3', status: '待处理' }
    ]

    it('should handle row click events', async () => {
      const mockRowClick = vi.fn()
      
      render(
        <TableComponent 
          data={mockData} 
          onRowClick={mockRowClick}
          onDelete={() => {}}
        />
      )
      
      const firstRow = screen.getByTestId('row-1')
      await user.click(firstRow)
      
      expect(mockRowClick).toHaveBeenCalledWith(mockData[0])
      expect(mockRowClick).toHaveBeenCalledTimes(1)
    })

    it('should handle delete button click', async () => {
      const mockDelete = vi.fn()
      
      render(
        <TableComponent 
          data={mockData} 
          onRowClick={() => {}}
          onDelete={mockDelete}
        />
      )
      
      const deleteButton = screen.getByTestId('delete-1')
      await user.click(deleteButton)
      
      expect(mockDelete).toHaveBeenCalledWith(1)
      expect(mockDelete).toHaveBeenCalledTimes(1)
    })

    it('should prevent event bubbling when clicking delete button', async () => {
      const mockRowClick = vi.fn()
      const mockDelete = vi.fn()
      
      render(
        <TableComponent 
          data={mockData} 
          onRowClick={mockRowClick}
          onDelete={mockDelete}
        />
      )
      
      const deleteButton = screen.getByTestId('delete-1')
      await user.click(deleteButton)
      
      expect(mockDelete).toHaveBeenCalledWith(1)
      expect(mockRowClick).not.toHaveBeenCalled()
    })

    it('should render table with correct data', () => {
      render(
        <TableComponent 
          data={mockData} 
          onRowClick={() => {}}
          onDelete={() => {}}
        />
      )
      
      expect(screen.getByText('案件1')).toBeInTheDocument()
      expect(screen.getByText('案件2')).toBeInTheDocument()
      expect(screen.getByText('案件3')).toBeInTheDocument()
      expect(screen.getByText('进行中')).toBeInTheDocument()
      expect(screen.getByText('已完成')).toBeInTheDocument()
      expect(screen.getByText('待处理')).toBeInTheDocument()
    })
  })

  describe('Modal Interactions', () => {
    it('should open and close modal', async () => {
      const mockOnClose = vi.fn()
      const mockOnConfirm = vi.fn()
      
      const { rerender } = render(
        <ModalComponent 
          isOpen={false}
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
          title="测试模态框"
        >
          <div>模态框内容</div>
        </ModalComponent>
      )
      
      expect(screen.queryByTestId('modal-overlay')).not.toBeInTheDocument()
      
      rerender(
        <ModalComponent 
          isOpen={true}
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
          title="测试模态框"
        >
          <div>模态框内容</div>
        </ModalComponent>
      )
      
      expect(screen.getByTestId('modal-overlay')).toBeInTheDocument()
      expect(screen.getByTestId('modal-title')).toHaveTextContent('测试模态框')
      expect(screen.getByTestId('modal-body')).toHaveTextContent('模态框内容')
    })

    it('should handle modal close action', async () => {
      const mockOnClose = vi.fn()
      const mockOnConfirm = vi.fn()
      
      render(
        <ModalComponent 
          isOpen={true}
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
          title="测试模态框"
        >
          <div>模态框内容</div>
        </ModalComponent>
      )
      
      const cancelButton = screen.getByTestId('cancel-button')
      await user.click(cancelButton)
      
      expect(mockOnClose).toHaveBeenCalledTimes(1)
      expect(mockOnConfirm).not.toHaveBeenCalled()
    })

    it('should handle modal confirm action', async () => {
      const mockOnClose = vi.fn()
      const mockOnConfirm = vi.fn()
      
      render(
        <ModalComponent 
          isOpen={true}
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
          title="测试模态框"
        >
          <div>模态框内容</div>
        </ModalComponent>
      )
      
      const confirmButton = screen.getByTestId('confirm-button')
      await user.click(confirmButton)
      
      expect(mockOnConfirm).toHaveBeenCalledTimes(1)
      expect(mockOnClose).not.toHaveBeenCalled()
    })

    it('should handle keyboard events in modal', async () => {
      const mockOnClose = vi.fn()
      const mockOnConfirm = vi.fn()
      
      render(
        <ModalComponent 
          isOpen={true}
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
          title="测试模态框"
        >
          <div>模态框内容</div>
        </ModalComponent>
      )
      
      // 让模态框获得焦点
      const modalOverlay = screen.getByTestId('modal-overlay')
      modalOverlay.focus()
      
      await user.keyboard('{Escape}')
      expect(mockOnClose).toHaveBeenCalledTimes(1)
    })
  })

  describe('Complex User Flows', () => {
    it('should handle complete user workflow', async () => {
      const mockLogin = vi.fn().mockResolvedValue({ success: true })
      const mockRowClick = vi.fn()
      const mockDelete = vi.fn()
      const mockOnClose = vi.fn()
      const mockOnConfirm = vi.fn()
      
      // Step 1: Login
      render(<LoginComponent onLogin={mockLogin} />)
      
      const usernameInput = screen.getByTestId('username-input')
      const passwordInput = screen.getByTestId('password-input')
      const submitButton = screen.getByTestId('submit-button')
      
      await user.type(usernameInput, 'admin')
      await user.type(passwordInput, 'password123')
      await user.click(submitButton)
      
      expect(mockLogin).toHaveBeenCalledWith('admin', 'password123')
      
      // Step 2: View table data
      const mockData = [
        { id: 1, name: '案件1', status: '进行中' }
      ]
      
      render(
        <TableComponent 
          data={mockData} 
          onRowClick={mockRowClick}
          onDelete={mockDelete}
        />
      )
      
      expect(screen.getByText('案件1')).toBeInTheDocument()
      
      // Step 3: Click on row
      const firstRow = screen.getByTestId('row-1')
      await user.click(firstRow)
      
      expect(mockRowClick).toHaveBeenCalledWith(mockData[0])
      
      // Step 4: Show confirmation modal
      render(
        <ModalComponent 
          isOpen={true}
          onClose={mockOnClose}
          onConfirm={mockOnConfirm}
          title="确认删除"
        >
          <div>确定要删除这个案件吗？</div>
        </ModalComponent>
      )
      
      expect(screen.getByTestId('modal-title')).toHaveTextContent('确认删除')
      expect(screen.getByTestId('modal-body')).toHaveTextContent('确定要删除这个案件吗？')
      
      // Step 5: Confirm deletion
      const confirmButton = screen.getByTestId('confirm-button')
      await user.click(confirmButton)
      
      expect(mockOnConfirm).toHaveBeenCalledTimes(1)
    })

    it('should handle error scenarios gracefully', async () => {
      const mockLogin = vi.fn().mockRejectedValue(new Error('Network error'))
      
      render(<LoginComponent onLogin={mockLogin} />)
      
      const usernameInput = screen.getByTestId('username-input')
      const passwordInput = screen.getByTestId('password-input')
      const submitButton = screen.getByTestId('submit-button')
      
      await user.type(usernameInput, 'admin')
      await user.type(passwordInput, 'password123')
      await user.click(submitButton)
      
      await waitFor(() => {
        expect(screen.getByTestId('error-message')).toBeInTheDocument()
        expect(screen.getByTestId('error-message')).toHaveTextContent('登录失败，请检查用户名和密码')
      })
      
      expect(submitButton).not.toBeDisabled()
      expect(submitButton).toHaveTextContent('登录')
    })
  })
})