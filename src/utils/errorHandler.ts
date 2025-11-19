/**
 * API错误处理工具
 * 提供统一的错误处理和用户友好的错误消息
 */

// 错误类型枚举
export enum ErrorType {
  NETWORK_ERROR = 'NETWORK_ERROR',
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  AUTHENTICATION_ERROR = 'AUTHENTICATION_ERROR',
  AUTHORIZATION_ERROR = 'AUTHORIZATION_ERROR',
  NOT_FOUND_ERROR = 'NOT_FOUND_ERROR',
  SERVER_ERROR = 'SERVER_ERROR',
  TIMEOUT_ERROR = 'TIMEOUT_ERROR',
  UNKNOWN_ERROR = 'UNKNOWN_ERROR',
}

// 错误严重级别
export enum ErrorSeverity {
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  CRITICAL = 'CRITICAL',
}

// 标准化的错误信息
export interface StandardError {
  type: ErrorType
  severity: ErrorSeverity
  code: string
  message: string
  userMessage: string
  details?: any
  timestamp: string
  requestId?: string
  retryable: boolean
}

// 错误处理配置
export interface ErrorHandlerConfig {
  enableLogging: boolean
  enableRetry: boolean
  maxRetries: number
  retryDelay: number
  logLevel: 'ERROR' | 'WARN' | 'INFO' | 'DEBUG'
}

// 默认配置
const DEFAULT_CONFIG: ErrorHandlerConfig = {
  enableLogging: true,
  enableRetry: true,
  maxRetries: 3,
  retryDelay: 1000,
  logLevel: 'ERROR',
}

/**
 * 错误处理器类
 */
export class ErrorHandler {
  private config: ErrorHandlerConfig

  constructor(config: Partial<ErrorHandlerConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config }
  }

  /**
   * 处理API错误
   */
  handleError(error: any, context?: string): StandardError {
    const standardError = this.standardizeError(error, context)

    if (this.config.enableLogging) {
      this.logError(standardError)
    }

    return standardError
  }

  /**
   * 将各种类型的错误标准化
   */
  private standardizeError(error: any, context?: string): StandardError {
    const timestamp = new Date().toISOString()
    const requestId = this.generateRequestId()

    // 网络错误
    if (this.isNetworkError(error)) {
      return {
        type: ErrorType.NETWORK_ERROR,
        severity: ErrorSeverity.MEDIUM,
        code: 'NETWORK_ERROR',
        message: error.message || '网络连接失败',
        userMessage: '网络连接异常，请检查网络连接后重试',
        details: error,
        timestamp,
        requestId,
        retryable: true,
      }
    }

    // HTTP状态码错误
    if (error.response) {
      const status = error.response.status
      const data = error.response.data

      switch (status) {
        case 400:
          return {
            type: ErrorType.VALIDATION_ERROR,
            severity: ErrorSeverity.MEDIUM,
            code: 'VALIDATION_ERROR',
            message: data?.error || '请求参数无效',
            userMessage: this.getUserFriendlyValidationMessage(data),
            details: data,
            timestamp,
            requestId,
            retryable: false,
          }

        case 401:
          return {
            type: ErrorType.AUTHENTICATION_ERROR,
            severity: ErrorSeverity.HIGH,
            code: 'AUTHENTICATION_ERROR',
            message: '身份验证失败',
            userMessage: '登录已过期，请重新登录',
            details: data,
            timestamp,
            requestId,
            retryable: false,
          }

        case 403:
          return {
            type: ErrorType.AUTHORIZATION_ERROR,
            severity: ErrorSeverity.HIGH,
            code: 'AUTHORIZATION_ERROR',
            message: '权限不足',
            userMessage: '您没有执行此操作的权限',
            details: data,
            timestamp,
            requestId,
            retryable: false,
          }

        case 404:
          return {
            type: ErrorType.NOT_FOUND_ERROR,
            severity: ErrorSeverity.MEDIUM,
            code: 'NOT_FOUND_ERROR',
            message: '资源未找到',
            userMessage: '请求的资源不存在',
            details: data,
            timestamp,
            requestId,
            retryable: false,
          }

        case 408:
          return {
            type: ErrorType.TIMEOUT_ERROR,
            severity: ErrorSeverity.MEDIUM,
            code: 'TIMEOUT_ERROR',
            message: '请求超时',
            userMessage: '请求处理超时，请稍后重试',
            details: data,
            timestamp,
            requestId,
            retryable: true,
          }

        case 500:
        case 502:
        case 503:
        case 504:
          return {
            type: ErrorType.SERVER_ERROR,
            severity: ErrorSeverity.HIGH,
            code: 'SERVER_ERROR',
            message: '服务器错误',
            userMessage: '服务器暂时无法响应，请稍后重试',
            details: data,
            timestamp,
            requestId,
            retryable: true,
          }

        default:
          return {
            type: ErrorType.UNKNOWN_ERROR,
            severity: ErrorSeverity.MEDIUM,
            code: `HTTP_${status}`,
            message: data?.error || `HTTP错误 ${status}`,
            userMessage: '系统异常，请稍后重试',
            details: data,
            timestamp,
            requestId,
            retryable: status >= 500,
          }
      }
    }

    // 超时错误
    if (this.isTimeoutError(error)) {
      return {
        type: ErrorType.TIMEOUT_ERROR,
        severity: ErrorSeverity.MEDIUM,
        code: 'TIMEOUT_ERROR',
        message: '请求超时',
        userMessage: '请求处理超时，请稍后重试',
        details: error,
        timestamp,
        requestId,
        retryable: true,
      }
    }

    // 其他未知错误
    return {
      type: ErrorType.UNKNOWN_ERROR,
      severity: ErrorSeverity.MEDIUM,
      code: 'UNKNOWN_ERROR',
      message: error?.message || '未知错误',
      userMessage: '系统异常，请稍后重试',
      details: error,
      timestamp,
      requestId,
      retryable: false,
    }
  }

  /**
   * 判断是否为网络错误
   */
  private isNetworkError(error: any): boolean {
    return (
      error.code === 'NETWORK_ERROR' ||
      error.code === 'ECONNREFUSED' ||
      error.code === 'ECONNRESET' ||
      error.code === 'ENOTFOUND' ||
      error.message?.includes('Network Error') ||
      !navigator.onLine
    )
  }

  /**
   * 判断是否为超时错误
   */
  private isTimeoutError(error: any): boolean {
    return (
      error.code === 'TIMEOUT_ERROR' ||
      error.code === 'ETIMEDOUT' ||
      error.message?.includes('timeout')
    )
  }

  /**
   * 获取用户友好的验证错误消息
   */
  private getUserFriendlyValidationMessage(data: any): string {
    if (!data || typeof data !== 'object') {
      return '输入数据格式不正确，请检查后重试'
    }

    // 如果有字段级别的错误信息
    if (data.details && typeof data.details === 'object') {
      const fieldErrors = Object.entries(data.details)
        .map(([field, message]) => `${field}: ${message}`)
        .join('; ')
      return `输入验证失败: ${fieldErrors}`
    }

    // 如果有具体的错误消息
    if (data.error && typeof data.error === 'string') {
      return data.error
    }

    return '输入数据不正确，请检查后重试'
  }

  /**
   * 记录错误日志
   */
  private logError(error: StandardError): void {
    const logData = {
      type: error.type,
      code: error.code,
      message: error.message,
      severity: error.severity,
      requestId: error.requestId,
      timestamp: error.timestamp,
      details: error.details,
    }

    switch (this.config.logLevel) {
      case 'ERROR':
        console.error('🚨 API Error:', logData)
        break
      case 'WARN':
        console.warn('⚠️ API Warning:', logData)
        break
      case 'INFO':
        console.info('ℹ️ API Info:', logData)
        break
      case 'DEBUG':
        console.debug('🐛 API Debug:', logData)
        break
    }
  }

  /**
   * 生成请求ID
   */
  private generateRequestId(): string {
    return `REQ_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  /**
   * 检查错误是否可重试
   */
  isRetryableError(error: StandardError): boolean {
    return error.retryable
  }

  /**
   * 获取重试延迟时间
   */
  getRetryDelay(attempt: number): number {
    return this.config.retryDelay * Math.pow(2, attempt - 1) // 指数退避
  }
}

// 默认错误处理器实例
export const defaultErrorHandler = new ErrorHandler()

/**
 * 便捷的错误处理函数
 */
export const handleError = (error: any, context?: string): StandardError => {
  return defaultErrorHandler.handleError(error, context)
}

/**
 * 判断是否应该重试
 */
export const shouldRetry = (error: StandardError, attempt: number): boolean => {
  return error.retryable && attempt <= defaultErrorHandler['config'].maxRetries
}

/**
 * 获取用户友好的错误消息
 */
export const getUserFriendlyMessage = (error: StandardError): string => {
  return error.userMessage
}
