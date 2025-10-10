/**
 * 全局错误处理工具
 * 提供统一的错误处理、用户提示和日志记录功能
 */

import { showSuccessToast, showErrorToast, showWarningToast, showInfoToast } from '../components/Toast';

export interface ErrorInfo {
  message: string;
  type: 'network' | 'validation' | 'permission' | 'server' | 'unknown';
  context?: string;
  details?: any;
  shouldNotify?: boolean;
  shouldLog?: boolean;
}

export class ErrorHandler {
  private static instance: ErrorHandler;
  private toastEnabled: boolean = true;

  private constructor() {
    // 私有构造函数
  }

  public static getInstance(): ErrorHandler {
    if (!ErrorHandler.instance) {
      ErrorHandler.instance = new ErrorHandler();
    }
    return ErrorHandler.instance;
  }

  // 设置Toast功能启用状态
  public setToastEnabled(enabled: boolean): void {
    this.toastEnabled = enabled;
  }

  // 主要错误处理方法
  public handleError(error: Error | string, context?: string): void {
    const errorInfo = this.parseError(error, context);

    // 记录错误
    if (errorInfo.shouldLog !== false) {
      this.logError(errorInfo);
    }

    // 显示用户提示
    if (errorInfo.shouldNotify !== false && this.toastEnabled) {
      this.notifyError(errorInfo);
    }
  }

  // 网络错误处理
  public handleNetworkError(error: Error | string, context?: string): void {
    const errorInfo: ErrorInfo = {
      message: this.extractMessage(error),
      type: 'network',
      context,
      shouldNotify: true,
      shouldLog: true
    };

    this.handleError(error as Error, context);
  }

  // API错误处理
  public handleApiError(status: number, error: Error | string, context?: string): void {
    const errorInfo: ErrorInfo = {
      message: this.getApiErrorMessage(status, error),
      type: this.getApiErrorType(status),
      context,
      shouldNotify: true,
      shouldLog: true,
      details: { status }
    };

    this.handleError(error as Error, context);
  }

  // 表单验证错误处理
  public handleValidationError(errors: Record<string, string>, context?: string): void {
    const messages = Object.values(errors);
    const errorInfo: ErrorInfo = {
      message: messages.join('; '),
      type: 'validation',
      context,
      shouldNotify: true,
      shouldLog: false,
      details: errors
    };

    this.handleError(new Error(errorInfo.message), context);
  }

  // 权限错误处理
  public handlePermissionError(message?: string, context?: string): void {
    const errorInfo: ErrorInfo = {
      message: message || '权限不足',
      type: 'permission',
      context,
      shouldNotify: true,
      shouldLog: true
    };

    this.handleError(new Error(errorInfo.message), context);
  }

  // 成功操作提示
  public showSuccess(message: string, context?: string): void {
    if (this.toastEnabled) {
      showSuccessToast(message, context);
    }
  }

  // 警告提示
  public showWarning(message: string, context?: string): void {
    if (this.toastEnabled) {
      showWarningToast(message, context);
    }
  }

  // 信息提示
  public showInfo(message: string, context?: string): void {
    if (this.toastEnabled) {
      showInfoToast(message, context);
    }
  }

  // 解析错误
  private parseError(error: Error | string, context?: string): ErrorInfo {
    const message = this.extractMessage(error);
    const type = this.getErrorType(message);

    return {
      message,
      type,
      context,
      shouldNotify: type !== 'unknown',
      shouldLog: type !== 'validation'
    };
  }

  // 提取错误消息
  private extractMessage(error: Error | string): string {
    if (typeof error === 'string') {
      return error;
    }
    return error.message || '未知错误';
  }

  // 获取错误类型
  private getErrorType(message: string): ErrorInfo['type'] {
    if (message.includes('fetch') || message.includes('Network') || message.includes('timeout')) {
      return 'network';
    }
    if (message.includes('401') || message.includes('403') || message.includes('permission') || message.includes('unauthorized')) {
      return 'permission';
    }
    if (message.includes('400') || message.includes('validation') || message.includes('required')) {
      return 'validation';
    }
    if (message.includes('500') || message.includes('server error') || message.includes('internal')) {
      return 'server';
    }
    return 'unknown';
  }

  // 获取API错误类型
  private getApiErrorType(status: number): ErrorInfo['type'] {
    if (status === 401 || status === 403) return 'permission';
    if (status >= 400 && status < 500) return 'validation';
    if (status >= 500) return 'server';
    return 'unknown';
  }

  // 获取API错误消息
  private getApiErrorMessage(status: number, error: Error | string): string {
    const message = this.extractMessage(error);

    // 网络错误优先使用状态码
    if (status === 401) return '登录已过期，请重新登录';
    if (status === 403) return '权限不足，请联系管理员';
    if (status === 404) return '请求的资源不存在';
    if (status === 422) return '数据验证失败，请检查输入';
    if (status === 429) return '请求过于频繁，请稍后重试';
    if (status >= 500) return '服务器内部错误，请稍后重试';

    return message || `请求失败 (${status})`;
  }

  // 记录错误日志
  private logError(errorInfo: ErrorInfo): void {
    const logData = {
      timestamp: new Date().toISOString(),
      type: errorInfo.type,
      message: errorInfo.message,
      context: errorInfo.context,
      details: errorInfo.details,
      userAgent: navigator.userAgent,
      url: window.location.href
    };

    // 在开发环境中输出详细错误信息
    if (process.env.NODE_ENV === 'development') {
      console.group(`🚨 ${errorInfo.type.toUpperCase()} ERROR`);
      console.error('Message:', errorInfo.message);
      console.error('Context:', errorInfo.context);
      console.error('Details:', errorInfo.details);
      console.groupEnd();
    } else {
      // 在生产环境中记录基本信息
      console.error(`[${errorInfo.type.toUpperCase()}] ${errorInfo.message}`, errorInfo.context);
    }

    // 这里可以添加错误上报服务
    // this.reportErrorToService(logData);
  }

  // 通知用户
  private notifyError(errorInfo: ErrorInfo): void {
    switch (errorInfo.type) {
      case 'network':
        showErrorToast(
          errorInfo.message,
          '网络错误',
          {
            title: '网络连接失败',
            persistent: true
          }
        );
        break;
      case 'permission':
        showErrorToast(
          errorInfo.message,
          '权限错误',
          {
            persistent: true
          }
        );
        break;
      case 'server':
        showErrorToast(
          errorInfo.message,
          '服务器错误',
          {
            title: '服务器繁忙',
            persistent: true
          }
        );
        break;
      case 'validation':
        showWarningToast(
          errorInfo.message,
          '验证错误'
        );
        break;
      default:
        showErrorToast(
          errorInfo.message,
          '系统错误',
          {
            persistent: true
          }
        );
    }
  }

  // 错误上报到服务（可选）
  private reportErrorToService(logData: any): void {
    // 这里可以集成Sentry、LogRocket等错误监控服务
    // 例如：
    // Sentry.captureException(logData.details?.error);
    console.log('Error reported to service:', logData);
  }

  // 重试机制
  public async retry<T>(
    operation: () => Promise<T>,
    maxRetries: number = 3,
    delay: number = 1000,
    context?: string
  ): Promise<T> {
    let lastError: Error;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        return await operation();
      } catch (error) {
        lastError = error as Error;

        if (attempt === maxRetries) {
          this.handleNetworkError(lastError, `${context} (Attempt ${attempt}/${maxRetries})`);
          throw lastError;
        }

        // 指数退避延迟
        const retryDelay = delay * Math.pow(2, attempt - 1);
        await new Promise(resolve => setTimeout(resolve, retryDelay));
      }
    }

    throw lastError;
  }
}

// 导出单例实例
export const errorHandler = ErrorHandler.getInstance();

// 导出便捷方法
export const handleError = (error: Error | string, context?: string) =>
  errorHandler.handleError(error, context);

export const handleNetworkError = (error: Error | string, context?: string) =>
  errorHandler.handleNetworkError(error, context);

export const handleApiError = (status: number, error: Error | string, context?: string) =>
  errorHandler.handleApiError(status, error, context);

export const handleValidationError = (errors: Record<string, string>, context?: string) =>
  errorHandler.handleValidationError(errors, context);

export const handlePermissionError = (message?: string, context?: string) =>
  errorHandler.handlePermissionError(message, context);

export const showSuccess = (message: string, context?: string) =>
  errorHandler.showSuccess(message, context);

export const showWarning = (message: string, context?: string) =>
  errorHandler.showWarning(message, context);

export const showInfo = (message: string, context?: string) =>
  errorHandler.showInfo(message, context);

export default errorHandler;