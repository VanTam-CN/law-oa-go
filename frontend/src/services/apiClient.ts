/**
 * 现代化API客户端 - 基于TanStack Query和标准化API响应格式
 * 实现统一的错误处理、缓存策略和类型安全
 */

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { getToken } from '@/utils/storage'

// 统一API响应格式
export interface ApiResponse<T = any> {
  data: T
  error: null | {
    message: string
    code: string
    details?: any
  }
  meta?: {
    timestamp: number
    requestId: string
    version: string
  }
}

// 分页响应格式
export interface PaginatedResponse<T> {
  data: T[]
  pagination: {
    page: number
    pageSize: number
    total: number
    totalPages: number
    hasNext: boolean
    hasPrev: boolean
  }
}

// API错误类
export class ApiError extends Error {
  constructor(
    public message: string,
    public code: string,
    public status?: number,
    public details?: any,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

// API请求配置接口
export interface ApiRequestConfig extends AxiosRequestConfig {
  skipErrorHandler?: boolean
  retryCount?: number
}

class ModernApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: '/api/v1',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.setupInterceptors()
  }

  private setupInterceptors() {
    // 请求拦截器
    this.client.interceptors.request.use(
      (config) => {
        // 添加请求ID
        config.headers = config.headers || {}
        config.headers['X-Request-ID'] = this.generateRequestId()

        // 添加认证token
        const token = getToken()
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }

        // 添加时间戳防止缓存
        if (config.method?.toLowerCase() === 'get') {
          config.params = {
            ...config.params,
            _t: Date.now(),
          }
        }

        console.log(`[API Request] ${config.method?.toUpperCase()} ${config.url}`)
        return config
      },
      (error) => {
        console.error('[API Request Error]', error)
        return Promise.reject(error)
      },
    )

    // 响应拦截器
    this.client.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        console.log(`[API Response] ${response.status} ${response.config.url}`)

        // 检查业务状态码
        if (response.data.error) {
          throw new ApiError(
            response.data.error.message,
            response.data.error.code,
            response.status,
            response.data.error.details,
          )
        }

        return response
      },
      (error) => {
        console.error('[API Response Error]', error)

        // 统一错误处理
        if (error.response) {
          const { status, data } = error.response
          const errorMessage = data?.error?.message || data?.message || '请求失败'
          const errorCode = data?.error?.code || 'UNKNOWN_ERROR'

          throw new ApiError(errorMessage, errorCode, status, data)
        } else if (error.request) {
          throw new ApiError('网络请求失败', 'NETWORK_ERROR')
        } else {
          throw new ApiError(error.message || '未知错误', 'UNKNOWN_ERROR')
        }
      },
    )
  }

  private generateRequestId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
  }

  // 通用GET请求
  async get<T>(url: string, config?: ApiRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.client.get<ApiResponse<T>>(url, config)
    return response.data
  }

  // 通用POST请求
  async post<T>(url: string, data?: any, config?: ApiRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.client.post<ApiResponse<T>>(url, data, config)
    return response.data
  }

  // 通用PUT请求
  async put<T>(url: string, data?: any, config?: ApiRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.client.put<ApiResponse<T>>(url, data, config)
    return response.data
  }

  // 通用DELETE请求
  async delete<T>(url: string, config?: ApiRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.client.delete<ApiResponse<T>>(url, config)
    return response.data
  }

  // 分页请求
  async getPaginated<T>(
    url: string,
    params: {
      page?: number
      pageSize?: number
      [key: string]: any
    } = {},
    config?: ApiRequestConfig,
  ): Promise<PaginatedResponse<T>> {
    const response = await this.client.get<ApiResponse<PaginatedResponse<T>>>(url, {
      ...config,
      params: {
        page: 1,
        pageSize: 20,
        ...params,
      },
    })
    return response.data.data
  }

  // 文件上传
  async uploadFile(
    url: string,
    file: File,
    onProgress?: (progress: number) => void,
    config?: ApiRequestConfig,
  ): Promise<ApiResponse<{ url: string; size: number; name: string }>> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await this.client.post<
      ApiResponse<{ url: string; size: number; name: string }>
    >(url, formData, {
      ...config,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(progress)
        }
      },
    })
    return response.data
  }

  // 批量请求
  async batch<T>(
    requests: Array<{ url: string; method?: string; data?: any }>,
  ): Promise<ApiResponse<T>[]> {
    const responses = await Promise.allSettled(
      requests.map(async (req) => {
        switch (req.method?.toLowerCase()) {
          case 'post':
            return this.post<T>(req.url, req.data)
          case 'put':
            return this.put<T>(req.url, req.data)
          case 'delete':
            return this.delete<T>(req.url)
          default:
            return this.get<T>(req.url)
        }
      }),
    )

    return responses.map((response, index) => {
      if (response.status === 'fulfilled') {
        return response.value
      } else {
        throw new ApiError(
          `批量请求失败 (${index + 1}/${requests.length})`,
          'BATCH_REQUEST_ERROR',
          undefined,
          response.reason,
        )
      }
    })
  }
}

// 创建单例实例
export const modernApiClient = new ModernApiClient()

// 为了向后兼容，导出为 apiClient
export const apiClient = modernApiClient
