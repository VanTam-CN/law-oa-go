/**
 * API客户端Mock - 现代化API请求模拟
 * 支持多种响应场景和错误处理
 */

import { ApiResponse, PaginatedResponse } from '../../services/apiClient'

// Mock响应数据类型
export interface MockApiResponse<T> extends ApiResponse<T> {}
export interface MockPaginatedResponse<T> extends PaginatedResponse<T> {}

// API客户端Mock类
class ApiClientMock {
  private responses = new Map<string, any>()
  private delays = new Map<string, number>()
  private errors = new Map<string, Error>()

  // 设置成功响应
  setSuccessResponse<T>(url: string, data: T, delay = 0): void {
    const response: MockApiResponse<T> = {
      data,
      error: null,
      meta: {
        timestamp: Date.now(),
        requestId: `mock-${Date.now()}`,
        version: '1.0.0',
      },
    }
    this.responses.set(url, response)
    if (delay > 0) {
      this.delays.set(url, delay)
    }
  }

  // 设置分页响应
  setPaginatedResponse<T>(
    url: string,
    data: T[],
    pagination: Partial<PaginatedResponse<T>['pagination']> = {},
    delay = 0,
  ): void {
    const defaultPagination = {
      page: 1,
      pageSize: 20,
      total: data.length,
      totalPages: Math.ceil(data.length / 20),
      hasNext: false,
      hasPrev: false,
    }

    const response: MockPaginatedResponse<T> = {
      data,
      pagination: { ...defaultPagination, ...pagination },
    }
    this.responses.set(url, { data: response, error: null })
    if (delay > 0) {
      this.delays.set(url, delay)
    }
  }

  // 设置错误响应
  setErrorResponse(url: string, error: Error | string, delay = 0): void {
    const errorObj = typeof error === 'string' ? new Error(error) : error
    this.errors.set(url, errorObj)
    if (delay > 0) {
      this.delays.set(url, delay)
    }
  }

  // 清除特定URL的Mock
  clearMock(url?: string): void {
    if (url) {
      this.responses.delete(url)
      this.delays.delete(url)
      this.errors.delete(url)
    } else {
      this.responses.clear()
      this.delays.clear()
      this.errors.clear()
    }
  }

  // 模拟GET请求
  async get<T>(url: string): Promise<MockApiResponse<T>> {
    const delay = this.delays.get(url) || 0

    if (delay > 0) {
      await new Promise((resolve) => setTimeout(resolve, delay))
    }

    const error = this.errors.get(url)
    if (error) {
      throw error
    }

    const response = this.responses.get(url)
    if (!response) {
      throw new Error(`No mock response configured for GET ${url}`)
    }

    return response
  }

  // 模拟POST请求
  async post<T>(url: string, data?: any): Promise<MockApiResponse<T>> {
    const delay = this.delays.get(url) || 0

    if (delay > 0) {
      await new Promise((resolve) => setTimeout(resolve, delay))
    }

    const error = this.errors.get(url)
    if (error) {
      throw error
    }

    const response = this.responses.get(url)
    if (!response) {
      throw new Error(`No mock response configured for POST ${url}`)
    }

    return response
  }

  // 模拟PUT请求
  async put<T>(url: string, data?: any): Promise<MockApiResponse<T>> {
    const delay = this.delays.get(url) || 0

    if (delay > 0) {
      await new Promise((resolve) => setTimeout(resolve, delay))
    }

    const error = this.errors.get(url)
    if (error) {
      throw error
    }

    const response = this.responses.get(url)
    if (!response) {
      throw new Error(`No mock response configured for PUT ${url}`)
    }

    return response
  }

  // 模拟DELETE请求
  async delete<T>(url: string): Promise<MockApiResponse<T>> {
    const delay = this.delays.get(url) || 0

    if (delay > 0) {
      await new Promise((resolve) => setTimeout(resolve, delay))
    }

    const error = this.errors.get(url)
    if (error) {
      throw error
    }

    const response = this.responses.get(url)
    if (!response) {
      throw new Error(`No mock response configured for DELETE ${url}`)
    }

    return response
  }

  // 模拟分页请求
  async getPaginated<T>(url: string): Promise<MockPaginatedResponse<T>> {
    const delay = this.delays.get(url) || 0

    if (delay > 0) {
      await new Promise((resolve) => setTimeout(resolve, delay))
    }

    const error = this.errors.get(url)
    if (error) {
      throw error
    }

    const response = this.responses.get(url)
    if (!response || !response.data.pagination) {
      throw new Error(`No paginated mock response configured for ${url}`)
    }

    return response.data
  }
}

// 创建单例实例
const apiClientMock = new ApiClientMock()

// 预设的常用Mock数据
export const mockUser = {
  id: 1,
  username: 'testuser',
  email: 'test@example.com',
  name: '测试用户',
  role: 'lawyer',
  permissions: ['case.view', 'case.create', 'client.view'],
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

export const mockCase = {
  id: 1,
  title: '测试案件',
  description: '这是一个测试案件',
  status: 'active',
  clientId: 1,
  lawyerId: 1,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

export const mockClient = {
  id: 1,
  name: '测试客户',
  email: 'client@example.com',
  phone: '13800138000',
  address: '测试地址',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

// 设置常用API响应
export const setupCommonMocks = (): void => {
  // 用户认证
  apiClientMock.setSuccessResponse('/auth/login', {
    user: mockUser,
    token: 'mock-jwt-token',
  })

  apiClientMock.setSuccessResponse('/auth/me', mockUser)

  // 案件管理
  apiClientMock.setPaginatedResponse('/cases', [mockCase], {
    page: 1,
    pageSize: 20,
    total: 1,
  })

  apiClientMock.setSuccessResponse('/cases/1', mockCase)

  // 客户管理
  apiClientMock.setPaginatedResponse('/clients', [mockClient], {
    page: 1,
    pageSize: 20,
    total: 1,
  })

  apiClientMock.setSuccessResponse('/clients/1', mockClient)
}

// 清理常用Mock
export const clearCommonMocks = (): void => {
  apiClientMock.clearMock('/auth/login')
  apiClientMock.clearMock('/auth/me')
  apiClientMock.clearMock('/cases')
  apiClientMock.clearMock('/cases/1')
  apiClientMock.clearMock('/clients')
  apiClientMock.clearMock('/clients/1')
}

export default apiClientMock
