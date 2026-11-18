/**
 * 认证服务Mock - 现代化认证测试支持
 * 支持多种认证场景和权限模拟
 */

import { UserFactory } from './factory'

export interface AuthState {
  user: any
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: Error | null
}

// Mock认证服务类
class MockAuthService {
  private state: AuthState = {
    user: null,
    token: null,
    isAuthenticated: false,
    isLoading: false,
    error: null,
  }

  // 模拟登录
  async login(credentials: { username: string; password: string }) {
    this.state.isLoading = true
    this.state.error = null

    // 模拟网络延迟
    await new Promise((resolve) => setTimeout(resolve, 1000))

    try {
      if (credentials.username === 'admin' && credentials.password === 'admin123') {
        const user = UserFactory.createAdmin()
        const token = 'mock-admin-jwt-token'

        this.state.user = user
        this.state.token = token
        this.state.isAuthenticated = true
        this.state.isLoading = false

        // 模拟存储到localStorage
        localStorage.setItem('auth_token', token)
        localStorage.setItem('user_info', JSON.stringify(user))

        return { user, token }
      } else if (credentials.username === 'lawyer' && credentials.password === 'lawyer123') {
        const user = UserFactory.create({ role: 'lawyer' })
        const token = 'mock-lawyer-jwt-token'

        this.state.user = user
        this.state.token = token
        this.state.isAuthenticated = true
        this.state.isLoading = false

        localStorage.setItem('auth_token', token)
        localStorage.setItem('user_info', JSON.stringify(user))

        return { user, token }
      } else {
        throw new Error('用户名或密码错误')
      }
    } catch (error) {
      this.state.error = error as Error
      this.state.isLoading = false
      throw error
    }
  }

  // 模拟注册
  async register(userData: { username: string; email: string; password: string; name: string }) {
    this.state.isLoading = true
    this.state.error = null

    await new Promise((resolve) => setTimeout(resolve, 800))

    try {
      const user = UserFactory.create({
        username: userData.username,
        email: userData.email,
        name: userData.name,
      })

      const token = 'mock-register-jwt-token'

      this.state.user = user
      this.state.token = token
      this.state.isAuthenticated = true
      this.state.isLoading = false

      localStorage.setItem('auth_token', token)
      localStorage.setItem('user_info', JSON.stringify(user))

      return { user, token }
    } catch (error) {
      this.state.error = error as Error
      this.state.isLoading = false
      throw error
    }
  }

  // 模拟登出
  async logout() {
    this.state.user = null
    this.state.token = null
    this.state.isAuthenticated = false
    this.state.error = null

    localStorage.removeItem('auth_token')
    localStorage.removeItem('user_info')
  }

  // 模拟获取当前用户
  async getCurrentUser() {
    this.state.isLoading = true

    await new Promise((resolve) => setTimeout(resolve, 300))

    const token = localStorage.getItem('auth_token')
    const userData = localStorage.getItem('user_info')

    if (token && userData) {
      try {
        const user = JSON.parse(userData)
        this.state.user = user
        this.state.token = token
        this.state.isAuthenticated = true
        this.state.isLoading = false
        return user
      } catch (error) {
        this.state.error = error as Error
        this.state.isLoading = false
        return null
      }
    } else {
      this.state.isLoading = false
      return null
    }
  }

  // 模拟更新用户资料
  async updateProfile(userData: Partial<any>) {
    this.state.isLoading = true
    this.state.error = null

    await new Promise((resolve) => setTimeout(resolve, 500))

    try {
      if (this.state.user) {
        this.state.user = { ...this.state.user, ...userData }
        localStorage.setItem('user_data', JSON.stringify(this.state.user))
        return this.state.user
      } else {
        throw new Error('用户未登录')
      }
    } catch (error) {
      this.state.error = error as Error
      this.state.isLoading = false
      throw error
    }
  }

  // 模拟修改密码
  async changePassword(passwordData: {
    currentPassword: string
    newPassword: string
    confirmPassword: string
  }) {
    this.state.isLoading = true
    this.state.error = null

    await new Promise((resolve) => setTimeout(resolve, 600))

    try {
      if (passwordData.newPassword !== passwordData.confirmPassword) {
        throw new Error('新密码和确认密码不匹配')
      }

      if (passwordData.currentPassword === passwordData.newPassword) {
        throw new Error('新密码不能与当前密码相同')
      }

      this.state.isLoading = false
      return { success: true, message: '密码修改成功' }
    } catch (error) {
      this.state.error = error as Error
      this.state.isLoading = false
      throw error
    }
  }

  // 权限检查
  hasPermission(permission: string): boolean {
    if (!this.state.user) {
      return false
    }

    if (this.state.user.role === 'admin') {
      return true
    }
    if (this.state.user.permissions && this.state.user.permissions.includes('*')) {
      return true
    }
    if (this.state.user.permissions && this.state.user.permissions.includes(permission)) {
      return true
    }

    return false
  }

  // 角色检查
  hasRole(role: string): boolean {
    return this.state.user?.role === role
  }

  // 获取当前状态
  getState(): AuthState {
    return { ...this.state }
  }

  // 重置状态
  resetState(): void {
    this.state = {
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
    }
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user_info')
  }

  // 设置特定状态（用于测试）
  setState(newState: Partial<AuthState>): void {
    this.state = { ...this.state, ...newState }
  }
}

// 创建单例实例
const mockAuthService = new MockAuthService()

// 导出预设状态
export const authStates = {
  anonymous: {
    user: null,
    token: null,
    isAuthenticated: false,
    isLoading: false,
    error: null,
  },
  authenticated: {
    user: UserFactory.create(),
    token: 'mock-jwt-token',
    isAuthenticated: true,
    isLoading: false,
    error: null,
  },
  admin: {
    user: UserFactory.createAdmin(),
    token: 'mock-admin-jwt-token',
    isAuthenticated: true,
    isLoading: false,
    error: null,
  },
  loading: {
    user: null,
    token: null,
    isAuthenticated: false,
    isLoading: true,
    error: null,
  },
  error: {
    user: null,
    token: null,
    isAuthenticated: false,
    isLoading: false,
    error: new Error('认证失败'),
  },
}

// 测试辅助函数
export const setupAuthState = (state: keyof typeof authStates): void => {
  mockAuthService.setState(authStates[state])
}

export const setupAuthenticatedUser = (user?: any): void => {
  const userToSet = user || UserFactory.create()
  mockAuthService.setState({
    user: userToSet,
    token: 'mock-jwt-token',
    isAuthenticated: true,
    isLoading: false,
    error: null,
  })
}

export const setupAdminUser = (): void => {
  const adminUser = UserFactory.createAdmin()
  mockAuthService.setState({
    user: adminUser,
    token: 'mock-admin-jwt-token',
    isAuthenticated: true,
    isLoading: false,
    error: null,
  })
}

export default mockAuthService
