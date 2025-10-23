/**
 * 调试助手工具 - 用于捕获和诊断JavaScript运行时错误
 */

// 全局错误捕获器
export class GlobalErrorCatcher {
  private static instance: GlobalErrorCatcher;
  private errors: any[] = [];

  private constructor() {
    this.setupGlobalErrorHandlers();
  }

  static getInstance(): GlobalErrorCatcher {
    if (!GlobalErrorCatcher.instance) {
      GlobalErrorCatcher.instance = new GlobalErrorCatcher();
    }
    return GlobalErrorCatcher.instance;
  }

  private setupGlobalErrorHandlers() {
    // 捕获未处理的Promise rejection
    window.addEventListener('unhandledrejection', (event) => {
      console.error('Unhandled Promise Rejection:', event.reason);
      this.logError('unhandledrejection', event.reason);
      // 防止错误继续传播
      event.preventDefault();
    });

    // 捕获全局错误
    window.addEventListener('error', (event) => {
      console.error('Global Error:', event.error);
      this.logError('global', event.error);
    });

    // 捕获message is not defined错误
    const originalError = window.onerror;
    window.onerror = (message, source, lineno, colno, error) => {
      if (typeof message === 'string' && message.includes('message is not defined')) {
        console.error('Detected message undefined error:', {
          message,
          source,
          lineno,
          colno,
          error,
          stack: error?.stack
        });

        // 尝试提供更具体的错误信息
        this.logError('message-undefined', {
          type: 'message-undefined',
          message,
          source,
          lineno,
          colno,
          stack: error?.stack,
          suggestion: 'Check if message variable is properly imported from antd'
        });
      }

      // 调用原始错误处理器
      if (originalError) {
        return originalError.call(window, message, source, lineno, colno, error);
      }
      return false;
    };
  }

  private logError(type: string, error: any) {
    const errorEntry = {
      type,
      timestamp: new Date().toISOString(),
      error: {
        message: error?.message || String(error),
        stack: error?.stack || 'No stack available',
        name: error?.name || 'UnknownError'
      },
      context: {
        userAgent: navigator.userAgent,
        url: window.location.href,
        timestamp: Date.now()
      }
    };

    this.errors.push(errorEntry);

    // 只保留最近100个错误
    if (this.errors.length > 100) {
      this.errors = this.errors.slice(-100);
    }

    // 在开发环境显示详细错误
    if (import.meta.env.DEV) {
      console.error(`[${type.toUpperCase()}] Error logged:`, errorEntry);
    }
  }

  getRecentErrors(limit: number = 10) {
    return this.errors.slice(-limit);
  }

  clearErrors() {
    this.errors = [];
  }

  hasMessageErrors() {
    return this.errors.some(error =>
      error.type === 'message-undefined' ||
      error.error.message.includes('message is not defined')
    );
  }
}

// 初始化全局错误捕获器
export const errorCatcher = GlobalErrorCatcher.getInstance();

// 安全的消息访问器
export const safeMessage = {
  success: (content: string, duration?: number) => {
    try {
      const { message } = require('antd');
      return message.success(content, duration);
    } catch (error) {
      console.error('Failed to show success message:', error);
      console.log('SUCCESS:', content);
      return null;
    }
  },

  error: (content: string, duration?: number) => {
    try {
      const { message } = require('antd');
      return message.error(content, duration);
    } catch (error) {
      console.error('Failed to show error message:', error);
      console.error('ERROR:', content);
      return null;
    }
  },

  info: (content: string, duration?: number) => {
    try {
      const { message } = require('antd');
      return message.info(content, duration);
    } catch (error) {
      console.error('Failed to show info message:', error);
      console.log('INFO:', content);
      return null;
    }
  },

  warning: (content: string, duration?: number) => {
    try {
      const { message } = require('antd');
      return message.warning(content, duration);
    } catch (error) {
      console.error('Failed to show warning message:', error);
      console.warn('WARNING:', content);
      return null;
    }
  }
};