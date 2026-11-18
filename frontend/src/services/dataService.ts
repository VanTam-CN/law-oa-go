/**
 * 统一数据服务层
 * 统一使用后端API调用
 * 提供一致的数据获取和处理逻辑
 */

import { message } from '@/utils/messageHelper'
import { get, post, put, del } from './http'

// =============================================================================
// 1. API错误处理和响应处理
// =============================================================================

interface ApiResponse<T = any> {
  data: T
  message?: string
  success: boolean
  pagination?: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

class DataServiceError extends Error {
  constructor(
    message: string,
    public code: string,
    public statusCode: number,
    public details?: any,
  ) {
    super(message)
    this.name = 'DataServiceError'
  }
}

const handleApiError = (error: any): never => {
  console.error('API调用失败:', error)

  if (error.response) {
    const status = error.response.status
    const data = error.response.data

    switch (status) {
      case 401:
        throw new DataServiceError('认证失败，请重新登录', 'AUTHENTICATION_ERROR', status, data)
      case 403:
        throw new DataServiceError('权限不足', 'AUTHORIZATION_ERROR', status, data)
      case 404:
        throw new DataServiceError('请求的资源不存在', 'NOT_FOUND', status, data)
      case 422:
        throw new DataServiceError(
          data?.message || '请求参数验证失败',
          'VALIDATION_ERROR',
          status,
          data,
        )
      case 500:
        throw new DataServiceError('服务器内部错误', 'INTERNAL_ERROR', status, data)
      default:
        throw new DataServiceError(
          data?.message || `请求失败 (${status})`,
          'UNKNOWN_ERROR',
          status,
          data,
        )
    }
  } else if (error.request) {
    throw new DataServiceError('网络连接失败，请检查网络设置', 'NETWORK_ERROR', 0, error)
  } else {
    throw new DataServiceError(error.message || '未知错误', 'UNKNOWN_ERROR', 0, error)
  }
}

// =============================================================================
// 2. 通用数据获取方法
// =============================================================================

class DataService {
  // 通用GET请求
  private async get<T>(url: string, params?: any): Promise<T> {
    try {
      return await get<T>(url, params)
    } catch (error) {
      throw handleApiError(error)
    }
  }

  // 通用POST请求
  private async post<T>(url: string, data?: any): Promise<T> {
    try {
      return await post<T>(url, data)
    } catch (error) {
      throw handleApiError(error)
    }
  }

  // 通用PUT请求
  private async put<T>(url: string, data?: any): Promise<T> {
    try {
      return await put<T>(url, data)
    } catch (error) {
      throw handleApiError(error)
    }
  }

  // 通用DELETE请求
  private async delete<T>(url: string): Promise<T> {
    try {
      return await del<T>(url)
    } catch (error) {
      throw handleApiError(error)
    }
  }

  // =============================================================================
  // 3. 仪表盘数据服务
  // =============================================================================

  async getDashboardStatistics(): Promise<any> {
    try {
      const response = await this.get<any>('/dashboard/statistics')
      return (
        response || {
          totalProjects: 0,
          completedProjects: 0,
          pendingApprovals: 0,
          activeClients: 0,
          projectStatus: {},
          approvalStatus: {},
          financeStats: {
            totalRevenue: 0,
            totalExpenses: 0,
          },
        }
      )
    } catch (error) {
      console.warn('获取仪表盘统计数据失败，返回空统计:', error)
      return {
        totalProjects: 0,
        completedProjects: 0,
        pendingApprovals: 0,
        activeClients: 0,
        projectStatus: {},
        approvalStatus: {},
        financeStats: {
          totalRevenue: 0,
          totalExpenses: 0,
        },
      }
    }
  }

  async getDashboardTodos(): Promise<any[]> {
    try {
      const response = await this.get<any[] | { todos?: any[] }>('/dashboard/todos')
      if (Array.isArray(response)) {
        return response
      }
      return Array.isArray(response?.todos) ? response.todos : []
    } catch (error) {
      console.warn('获取待办事项失败，返回空数组:', error)
      return []
    }
  }

  async getDashboardActivities(): Promise<any[]> {
    try {
      const response = await this.get<any[] | { activities?: any[] }>('/dashboard/activities')
      if (Array.isArray(response)) {
        return response
      }
      return Array.isArray(response?.activities) ? response.activities : []
    } catch (error) {
      console.warn('获取活动记录失败，返回空数组:', error)
      return []
    }
  }

  // =============================================================================
  // 4. 律师数据服务
  // =============================================================================

  async getLawyers(
    params: {
      page?: number
      page_size?: number
      search?: string
      status?: string
      department?: string
      specialty?: string
    } = {},
  ): Promise<{
    data: any[]
    pagination: any
    total: number
  }> {
    try {
      const response = await this.get<ApiResponse<any[]>>('/lawfirm/lawyers', params)

      return {
        data: response.data || [],
        pagination: response.pagination || {
          page: 1,
          page_size: 10,
          total: 0,
          total_pages: 0,
        },
        total: response.pagination?.total || 0,
      }
    } catch (error) {
      console.error('获取律师列表失败:', error)
      return {
        data: [],
        pagination: {
          page: 1,
          page_size: 10,
          total: 0,
          total_pages: 0,
        },
        total: 0,
      }
    }
  }

  async getLawyerById(id: number): Promise<any> {
    try {
      const response = await this.get<any>(`/lawfirm/lawyers/${id}`)
      return response
    } catch (error) {
      console.error('获取律师详情失败:', error)
      throw handleApiError(error)
    }
  }

  async createLawyer(data: any): Promise<any> {
    try {
      const response = await this.post<any>('/lawfirm/lawyers', data)
      return response
    } catch (error) {
      console.error('创建律师失败:', error)
      throw handleApiError(error)
    }
  }

  async updateLawyer(id: number, data: any): Promise<any> {
    try {
      const response = await this.put<any>(`/lawfirm/lawyers/${id}`, data)
      return response
    } catch (error) {
      console.error('更新律师失败:', error)
      throw handleApiError(error)
    }
  }

  async deleteLawyer(id: number): Promise<void> {
    try {
      await this.delete(`/lawfirm/lawyers/${id}`)
    } catch (error) {
      console.error('删除律师失败:', error)
      throw handleApiError(error)
    }
  }

  // =============================================================================
  // 5. 案件数据服务
  // =============================================================================

  async getCases(
    params: {
      page?: number
      page_size?: number
      search?: string
      status?: string
      case_type?: string
      lawyer_id?: number
      client_id?: number
    } = {},
  ): Promise<{
    data: any[]
    pagination: any
    total: number
  }> {
    try {
      const response = await this.get<{ cases: any[]; pagination: any }>('/cases', params)

      return {
        data: response.cases || [],
        pagination: response.pagination || {
          page: 1,
          page_size: 10,
          total: 0,
          total_page: 0,
        },
        total: response.pagination?.total || 0,
      }
    } catch (error) {
      console.error('获取案件列表失败:', error)
      return {
        data: [],
        pagination: {
          page: 1,
          page_size: 10,
          total: 0,
          total_page: 0,
        },
        total: 0,
      }
    }
  }

  async getCaseById(id: number): Promise<any> {
    try {
      const response = await this.get<any>(`/cases/${id}`)
      return response
    } catch (error) {
      console.error('获取案件详情失败:', error)
      throw handleApiError(error)
    }
  }

  async createCase(data: any): Promise<any> {
    try {
      const response = await this.post<any>('/cases', data)
      return response
    } catch (error) {
      console.error('创建案件失败:', error)
      throw handleApiError(error)
    }
  }

  async updateCase(id: number, data: any): Promise<any> {
    try {
      const response = await this.put<any>(`/cases/${id}`, data)
      return response
    } catch (error) {
      console.error('更新案件失败:', error)
      throw handleApiError(error)
    }
  }

  async deleteCase(id: number): Promise<void> {
    try {
      await this.delete(`/cases/${id}`)
    } catch (error) {
      console.error('删除案件失败:', error)
      throw handleApiError(error)
    }
  }

  // =============================================================================
  // 6. 用户数据服务
  // =============================================================================

  async getUsers(
    params: {
      page?: number
      page_size?: number
      search?: string
      status?: string
      user_type?: string
      department_id?: number
    } = {},
  ): Promise<{
    data: any[]
    pagination: any
    total: number
  }> {
    try {
      const response = await this.get<ApiResponse<any[]>>('/users', params)

      return {
        data: response.data || [],
        pagination: response.pagination || {
          page: 1,
          page_size: 10,
          total: 0,
          total_pages: 0,
        },
        total: response.pagination?.total || 0,
      }
    } catch (error) {
      console.error('获取用户列表失败:', error)
      return {
        data: [],
        pagination: {
          page: 1,
          page_size: 10,
          total: 0,
          total_pages: 0,
        },
        total: 0,
      }
    }
  }

  async getUserById(id: number): Promise<any> {
    try {
      const response = await this.get<any>(`/users/${id}`)
      return response
    } catch (error) {
      console.error('获取用户详情失败:', error)
      throw handleApiError(error)
    }
  }

  async createUser(data: any): Promise<any> {
    try {
      const response = await this.post<any>('/users', data)
      return response
    } catch (error) {
      console.error('创建用户失败:', error)
      throw handleApiError(error)
    }
  }

  async updateUser(id: number, data: any): Promise<any> {
    try {
      const response = await this.put<any>(`/users/${id}`, data)
      return response
    } catch (error) {
      console.error('更新用户失败:', error)
      throw handleApiError(error)
    }
  }

  async deleteUser(id: number): Promise<void> {
    try {
      await this.delete(`/users/${id}`)
    } catch (error) {
      console.error('删除用户失败:', error)
      throw handleApiError(error)
    }
  }

  async getUserStats(): Promise<any> {
    try {
      const response = await this.get<any>('/users/stats')
      return (
        response || {
          total: 0,
          active: 0,
          inactive: 0,
          byType: {},
          byDepartment: {},
        }
      )
    } catch (error) {
      console.warn('获取用户统计失败，返回空统计:', error)
      return {
        total: 0,
        active: 0,
        inactive: 0,
        byType: {},
        byDepartment: {},
      }
    }
  }

  // =============================================================================
  // 7. 客户数据服务
  // =============================================================================

  async getClients(
    params: {
      page?: number
      page_size?: number
      search?: string
      type?: string
    } = {},
  ): Promise<{
    data: any[]
    pagination: any
    total: number
  }> {
    try {
      const response = await this.get<{ clients: any[]; pagination: any }>('/clients', params)

      return {
        data: response.clients || [],
        pagination: response.pagination || {
          page: 1,
          page_size: 10,
          total: 0,
          total_page: 0,
        },
        total: response.pagination?.total || 0,
      }
    } catch (error) {
      console.error('获取客户列表失败:', error)
      return {
        data: [],
        pagination: {
          page: 1,
          page_size: 10,
          total: 0,
          total_page: 0,
        },
        total: 0,
      }
    }
  }

  async getClientById(id: number): Promise<any> {
    try {
      const response = await this.get<any>(`/clients/${id}`)
      return response
    } catch (error) {
      console.error('获取客户详情失败:', error)
      throw handleApiError(error)
    }
  }

  async createClient(data: any): Promise<any> {
    try {
      const response = await this.post<any>('/clients', data)
      return response
    } catch (error) {
      console.error('创建客户失败:', error)
      throw handleApiError(error)
    }
  }

  async updateClient(id: number, data: any): Promise<any> {
    try {
      const response = await this.put<any>(`/clients/${id}`, data)
      return response
    } catch (error) {
      console.error('更新客户失败:', error)
      throw handleApiError(error)
    }
  }

  async deleteClient(id: number): Promise<void> {
    try {
      await this.delete(`/clients/${id}`)
    } catch (error) {
      console.error('删除客户失败:', error)
      throw handleApiError(error)
    }
  }

  // =============================================================================
  // 8. 工具方法
  // =============================================================================

  // 格式化分页参数
  private formatPaginationParams(params: any) {
    return {
      page: params.page || 1,
      page_size: params.page_size || 10,
      ...params,
    }
  }

  // 处理搜索参数
  private formatSearchParams(searchParams: any) {
    const formatted: any = {}

    if (searchParams.search && searchParams.search.trim()) {
      formatted.search = searchParams.search.trim()
    }

    Object.keys(searchParams).forEach((key) => {
      if (key !== 'search' && searchParams[key] !== undefined && searchParams[key] !== '') {
        formatted[key] = searchParams[key]
      }
    })

    return formatted
  }

  // 通用的列表数据获取方法
  async getListData<T>(
    endpoint: string,
    params: any = {},
    searchParams: any = {},
  ): Promise<{
    data: T[]
    pagination: any
    total: number
  }> {
    try {
      const formattedParams = {
        ...this.formatPaginationParams(params),
        ...this.formatSearchParams(searchParams),
      }

      const response = await this.get<ApiResponse<T[]>>(endpoint, formattedParams)

      return {
        data: response.data || [],
        pagination: response.pagination || {
          page: 1,
          page_size: 10,
          total: 0,
          total_pages: 0,
        },
        total: response.pagination?.total || 0,
      }
    } catch (error) {
      console.error(`获取${endpoint}列表失败:`, error)
      return {
        data: [],
        pagination: {
          page: 1,
          page_size: 10,
          total: 0,
          total_pages: 0,
        },
        total: 0,
      }
    }
  }

  // 批量操作
  async batchOperation<T>(
    endpoint: string,
    operation: 'delete' | 'update',
    items: any[],
    updateData?: any,
  ): Promise<{
    success: number
    failed: number
    errors?: string[]
  }> {
    try {
      const response = await this.post<{
        success: number
        failed: number
        errors?: string[]
      }>(`${endpoint}/batch/${operation}`, {
        items,
        update_data: updateData,
      })

      return response
    } catch (error) {
      console.error(`批量${operation}失败:`, error)
      throw handleApiError(error)
    }
  }
}

// =============================================================================
// 9. 导出单例实例
// =============================================================================

export const dataService = new DataService()

// 为了向后兼容，也导出独立的方法
export const getDashboardStatistics = dataService.getDashboardStatistics.bind(dataService)
export const getDashboardTodos = dataService.getDashboardTodos.bind(dataService)
export const getDashboardActivities = dataService.getDashboardActivities.bind(dataService)
export const getLawyers = dataService.getLawyers.bind(dataService)
export const getLawyerById = dataService.getLawyerById.bind(dataService)
export const createLawyer = dataService.createLawyer.bind(dataService)
export const updateLawyer = dataService.updateLawyer.bind(dataService)
export const deleteLawyer = dataService.deleteLawyer.bind(dataService)
export const getCases = dataService.getCases.bind(dataService)
export const getCaseById = dataService.getCaseById.bind(dataService)
export const createCase = dataService.createCase.bind(dataService)
export const updateCase = dataService.updateCase.bind(dataService)
export const deleteCase = dataService.deleteCase.bind(dataService)
export const getUsers = dataService.getUsers.bind(dataService)
export const getUserById = dataService.getUserById.bind(dataService)
export const createUser = dataService.createUser.bind(dataService)
export const updateUser = dataService.updateUser.bind(dataService)
export const deleteUser = dataService.deleteUser.bind(dataService)
export const getUserStats = dataService.getUserStats.bind(dataService)
export const getClients = dataService.getClients.bind(dataService)
export const getClientById = dataService.getClientById.bind(dataService)
export const createClient = dataService.createClient.bind(dataService)
export const updateClient = dataService.updateClient.bind(dataService)
export const deleteClient = dataService.deleteClient.bind(dataService)

export default dataService
