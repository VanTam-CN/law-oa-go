import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Header from '../../../components/layout/Header.jsx'

// Mock Ant Design components
vi.mock('antd', async () => {
  const actual = await vi.importActual('antd')
  return {
    ...actual,
    Layout: {
      Header: ({ children, className, ...props }) => (
        <header className={className} {...props}>
          {children}
        </header>
      )
    },
    Dropdown: ({ children, menu, ...props }) => (
      <div {...props}>
        {children}
        <div className="dropdown-menu">
          {menu?.items?.map(item => (
            <div key={item.key} className="dropdown-item">
              {item.label}
            </div>
          ))}
        </div>
      </div>
    ),
    Space: ({ children, size, ...props }) => (
      <div className={`space space-${size}`} {...props}>
        {children}
      </div>
    ),
    Badge: ({ children, count, ...props }) => (
      <div className="badge" {...props}>
        {children}
        {count > 0 && <span className="badge-count">{count}</span>}
      </div>
    ),
    Avatar: ({ icon, ...props }) => (
      <div className="avatar" {...props}>
        {icon}
      </div>
    )
  }
})

// Mock Ant Design icons
vi.mock('@ant-design/icons', () => ({
  BellOutlined: () => <div className="bell-icon">Bell</div>,
  UserOutlined: () => <div className="user-icon">User</div>
}))

// Mock CSS
vi.mock('./header.less', () => ({}))

describe('Header Component', () => {
  const mockOnMenuClick = vi.fn()
  const mockOnNotificationClick = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render header component', () => {
    render(<Header />)
    
    const header = screen.getByRole('banner')
    expect(header).toBeInTheDocument()
    expect(header).toHaveClass('app-header')
  })

  it('should display user avatar and name', () => {
    render(<Header />)
    
    const avatar = screen.getByText('User')
    const userName = screen.getByText('管理员')
    
    expect(avatar).toBeInTheDocument()
    expect(userName).toBeInTheDocument()
  })

  it('should display notification bell with badge', () => {
    render(<Header />)
    
    const bellIcon = screen.getByText('Bell')
    const badgeCount = screen.getByText('5')
    
    expect(bellIcon).toBeInTheDocument()
    expect(badgeCount).toBeInTheDocument()
  })

  it('should show dropdown menu when clicking on user area', async () => {
    const user = userEvent.setup()
    render(<Header />)
    
    const userArea = screen.getByText('管理员')
    await user.click(userArea)
    
    const dropdownMenu = screen.getByRole('menu')
    expect(dropdownMenu).toBeInTheDocument()
    
    const personalCenter = screen.getByText('个人中心')
    const logout = screen.getByText('退出登录')
    
    expect(personalCenter).toBeInTheDocument()
    expect(logout).toBeInTheDocument()
  })

  it('should have correct dropdown menu items', () => {
    render(<Header />)
    
    // The dropdown items should be present in the DOM
    expect(screen.getByText('个人中心')).toBeInTheDocument()
    expect(screen.getByText('退出登录')).toBeInTheDocument()
  })

  it('should have correct spacing between elements', () => {
    render(<Header />)
    
    const space = screen.getByRole('banner').querySelector('.space')
    expect(space).toBeInTheDocument()
    expect(space).toHaveClass('space-large')
  })

  it('should have badge with count 5', () => {
    render(<Header />)
    
    const badge = screen.getByRole('banner').querySelector('.badge')
    expect(badge).toBeInTheDocument()
    
    const badgeCount = badge.querySelector('.badge-count')
    expect(badgeCount).toBeInTheDocument()
    expect(badgeCount).toHaveTextContent('5')
  })

  it('should have correct header layout structure', () => {
    render(<Header />)
    
    const header = screen.getByRole('banner')
    const headerRight = header.querySelector('.header-right')
    const space = headerRight.querySelector('.space')
    
    expect(header).toContainElement(headerRight)
    expect(headerRight).toContainElement(space)
  })

  it('should render all required elements', () => {
    render(<Header />)
    
    expect(screen.getByRole('banner')).toBeInTheDocument()
    expect(screen.getByText('Bell')).toBeInTheDocument()
    expect(screen.getByText('User')).toBeInTheDocument()
    expect(screen.getByText('管理员')).toBeInTheDocument()
    expect(screen.getByText('5')).toBeInTheDocument()
    expect(screen.getByText('个人中心')).toBeInTheDocument()
    expect(screen.getByText('退出登录')).toBeInTheDocument()
  })

  it('should have correct CSS classes', () => {
    render(<Header />)
    
    const header = screen.getByRole('banner')
    const headerRight = header.querySelector('.header-right')
    const space = header.querySelector('.space')
    const badge = header.querySelector('.badge')
    const avatar = header.querySelector('.avatar')
    
    expect(header).toHaveClass('app-header')
    expect(headerRight).toBeInTheDocument()
    expect(space).toHaveClass('space-large')
    expect(badge).toBeInTheDocument()
    expect(avatar).toBeInTheDocument()
  })

  it('should be accessible', () => {
    render(<Header />)
    
    const header = screen.getByRole('banner')
    expect(header).toBeInTheDocument()
    
    // Check for interactive elements
    const bellIcon = screen.getByText('Bell')
    const userArea = screen.getByText('管理员')
    
    expect(bellIcon).toBeInTheDocument()
    expect(userArea).toBeInTheDocument()
  })

  it('should handle user interaction', async () => {
    const user = userEvent.setup()
    const handleMenuClick = vi.fn()
    
    render(<Header onMenuClick={handleMenuClick} />)
    
    const userArea = screen.getByText('管理员')
    await user.click(userArea)
    
    const personalCenter = screen.getByText('个人中心')
    await user.click(personalCenter)
    
    expect(handleMenuClick).toHaveBeenCalled()
  })

  it('should render consistently', () => {
    const { container } = render(<Header />)
    const { container: container2 } = render(<Header />)
    
    expect(container.innerHTML).toBe(container2.innerHTML)
  })

  it('should have proper component structure', () => {
    render(<Header />)
    
    const header = screen.getByRole('banner')
    const headerRight = header.querySelector('.header-right')
    const space = headerRight.querySelector('.space')
    const badge = space.querySelector('.badge')
    const dropdown = space.querySelector('div')
    
    expect(header).toBeInTheDocument()
    expect(headerRight).toBeInTheDocument()
    expect(space).toBeInTheDocument()
    expect(badge).toBeInTheDocument()
    expect(dropdown).toBeInTheDocument()
  })
})