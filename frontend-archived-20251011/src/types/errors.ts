/**
 * 自定义应用错误类
 */
export class AppError extends Error {
  public readonly code: string;
  public readonly statusCode?: number;
  public readonly details?: any;
  public readonly suggestions?: string[];
  public readonly isRetryable: boolean;

  constructor(
    message: string,
    code: string = 'UNKNOWN_ERROR',
    statusCode?: number,
    details?: any,
    suggestions?: string[],
    isRetryable: boolean = false
  ) {
    super(message);
    this.name = 'AppError';
    this.code = code;
    this.statusCode = statusCode;
    this.details = details;
    this.suggestions = suggestions;
    this.isRetryable = isRetryable;

    // 保持正确的原型链
    Object.setPrototypeOf(this, AppError.prototype);
  }

  /**
   * 创建网络错误
   */
  static networkError(message: string = '网络连接失败，请检查网络设置'): AppError {
    return new AppError(
      message,
      'NETWORK_ERROR',
      undefined,
      undefined,
      ['检查网络连接', '确认服务器正常运行'],
      true
    );
  }

  /**
   * 创建认证错误
   */
  static authError(message: string = '认证失败'): AppError {
    return new AppError(
      message,
      'AUTHENTICATION_ERROR',
      401,
      undefined,
      ['重新登录', '检查用户名和密码'],
      false
    );
  }

  /**
   * 创建权限错误
   */
  static permissionError(message: string = '权限不足'): AppError {
    return new AppError(
      message,
      'AUTHORIZATION_ERROR',
      403,
      undefined,
      ['联系管理员获取权限', '确认账号状态正常'],
      false
    );
  }

  /**
   * 创建验证错误
   */
  static validationError(message: string, details?: any): AppError {
    return new AppError(
      message,
      'VALIDATION_ERROR',
      400,
      details,
      ['检查请求参数格式', '确认必填字段已填写'],
      false
    );
  }

  /**
   * 创建资源未找到错误
   */
  static notFoundError(message: string = '请求的资源不存在'): AppError {
    return new AppError(
      message,
      'NOT_FOUND',
      404,
      undefined,
      ['检查资源ID是否正确', '确认资源存在'],
      false
    );
  }

  /**
   * 创建服务器错误
   */
  static serverError(message: string = '服务器内部错误'): AppError {
    return new AppError(
      message,
      'INTERNAL_ERROR',
      500,
      undefined,
      ['联系技术支持', '稍后重试'],
      true
    );
  }

  /**
   * 创建请求超时错误
   */
  static timeoutError(message: string = '请求超时'): AppError {
    return new AppError(
      message,
      'TIMEOUT_ERROR',
      undefined,
      undefined,
      ['检查网络连接', '稍后重试'],
      true
    );
  }

  /**
   * 创建请求频率限制错误
   */
  static rateLimitError(message: string = '请求频率超限'): AppError {
    return new AppError(
      message,
      'RATE_LIMIT_ERROR',
      429,
      undefined,
      ['稍后重试', '降低请求频率'],
      true
    );
  }

  /**
   * 从API响应错误创建AppError
   */
  static fromApiError(apiError: any): AppError {
    return new AppError(
      apiError.message || 'API请求失败',
      apiError.code || 'API_ERROR',
      apiError.statusCode,
      apiError.details,
      apiError.suggestions,
      this.isErrorRetryable(apiError.code)
    );
  }

  /**
   * 判断错误是否可重试
   */
  private static isErrorRetryable(code: string): boolean {
    const nonRetryableCodes = [
      'AUTHENTICATION_ERROR',
      'AUTHORIZATION_ERROR',
      'VALIDATION_ERROR',
      'NOT_FOUND',
      'CONFLICT'
    ];
    return !nonRetryableCodes.includes(code);
  }

  /**
   * 判断是否为网络相关错误
   */
  isNetworkError(): boolean {
    return this.code === 'NETWORK_ERROR' || this.code === 'TIMEOUT_ERROR';
  }

  /**
   * 判断是否为服务器错误
   */
  isServerError(): boolean {
    return this.code === 'INTERNAL_ERROR' || (this.statusCode !== undefined && this.statusCode >= 500);
  }

  /**
   * 获取用户友好的错误信息
   */
  getUserMessage(): string {
    return this.message;
  }

  /**
   * 获取错误处理建议
   */
  getSuggestions(): string[] {
    return this.suggestions || ['稍后重试'];
  }

  /**
   * 转换为JSON对象
   */
  toJSON() {
    return {
      name: this.name,
      message: this.message,
      code: this.code,
      statusCode: this.statusCode,
      details: this.details,
      suggestions: this.suggestions,
      isRetryable: this.isRetryable,
      stack: this.stack
    };
  }
}

/**
 * 重试配置接口
 */
export interface RetryConfig {
  maxAttempts: number;
  baseDelay: number;
  maxDelay: number;
  retryableErrors: string[];
}

/**
 * 缓存配置接口
 */
export interface CacheConfig {
  enabled: boolean;
  ttl: number;
  key?: string;
}

/**
 * 重试默认配置
 */
export const DEFAULT_RETRY_CONFIG: RetryConfig = {
  maxAttempts: 3,
  baseDelay: 1000,
  maxDelay: 30000,
  retryableErrors: [
    'NETWORK_ERROR',
    'TIMEOUT_ERROR',
    'INTERNAL_ERROR',
    'RATE_LIMIT_ERROR'
  ]
};

/**
 * 缓存默认配置
 */
export const DEFAULT_CACHE_CONFIG: CacheConfig = {
  enabled: true,
  ttl: 5 * 60 * 1000 // 5分钟
};
