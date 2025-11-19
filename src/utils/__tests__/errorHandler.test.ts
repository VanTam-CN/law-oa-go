/**
 * 错误处理工具的单元测试
 */

import {
  ErrorHandler,
  handleError,
  shouldRetry,
  getUserFriendlyMessage,
  ErrorType,
  ErrorSeverity,
  StandardError
} from '../errorHandler';

describe('ErrorHandler', () => {
  let errorHandler: ErrorHandler;

  beforeEach(() => {
    errorHandler = new ErrorHandler({
      enableLogging: false, // 测试时禁用日志
      enableRetry: true,
      maxRetries: 3,
      retryDelay: 100,
      logLevel: 'ERROR'
    });
  });

  describe('handleError', () => {
    it('应该处理网络错误', () => {
      const networkError = new Error('Network Error');
      networkError.code = 'NETWORK_ERROR';

      const result = errorHandler.handleError(networkError);

      expect(result.type).toBe(ErrorType.NETWORK_ERROR);
      expect(result.severity).toBe(ErrorSeverity.MEDIUM);
      expect(result.userMessage).toBe('网络连接异常，请检查网络连接后重试');
      expect(result.retryable).toBe(true);
    });

    it('应该处理400验证错误', () => {
      const validationError = {
        response: {
          status: 400,
          data: {
            error: '案件类型无效',
            details: {
              caseType: '案件类型必须是有效的类型'
            }
          }
        }
      };

      const result = errorHandler.handleError(validationError);

      expect(result.type).toBe(ErrorType.VALIDATION_ERROR);
      expect(result.severity).toBe(ErrorSeverity.MEDIUM);
      expect(result.userMessage).toContain('输入验证失败');
      expect(result.retryable).toBe(false);
    });

    it('应该处理401认证错误', () => {
      const authError = {
        response: {
          status: 401,
          data: { error: 'Token expired' }
        }
      };

      const result = errorHandler.handleError(authError);

      expect(result.type).toBe(ErrorType.AUTHENTICATION_ERROR);
      expect(result.severity).toBe(ErrorSeverity.HIGH);
      expect(result.userMessage).toBe('登录已过期，请重新登录');
      expect(result.retryable).toBe(false);
    });

    it('应该处理403权限错误', () => {
      const permissionError = {
        response: {
          status: 403,
          data: { error: 'Access denied' }
        }
      };

      const result = errorHandler.handleError(permissionError);

      expect(result.type).toBe(ErrorType.AUTHORIZATION_ERROR);
      expect(result.severity).toBe(ErrorSeverity.HIGH);
      expect(result.userMessage).toBe('您没有执行此操作的权限');
      expect(result.retryable).toBe(false);
    });

    it('应该处理404错误', () => {
      const notFoundError = {
        response: {
          status: 404,
          data: { error: 'Resource not found' }
        }
      };

      const result = errorHandler.handleError(notFoundError);

      expect(result.type).toBe(ErrorType.NOT_FOUND_ERROR);
      expect(result.severity).toBe(ErrorSeverity.MEDIUM);
      expect(result.userMessage).toBe('请求的资源不存在');
      expect(result.retryable).toBe(false);
    });

    it('应该处理500服务器错误', () => {
      const serverError = {
        response: {
          status: 500,
          data: { error: 'Internal server error' }
        }
      };

      const result = errorHandler.handleError(serverError);

      expect(result.type).toBe(ErrorType.SERVER_ERROR);
      expect(result.severity).toBe(ErrorSeverity.HIGH);
      expect(result.userMessage).toBe('服务器暂时无法响应，请稍后重试');
      expect(result.retryable).toBe(true);
    });

    it('应该处理超时错误', () => {
      const timeoutError = new Error('Request timeout');
      timeoutError.code = 'TIMEOUT_ERROR';

      const result = errorHandler.handleError(timeoutError);

      expect(result.type).toBe(ErrorType.TIMEOUT_ERROR);
      expect(result.severity).toBe(ErrorSeverity.MEDIUM);
      expect(result.userMessage).toBe('请求处理超时，请稍后重试');
      expect(result.retryable).toBe(true);
    });

    it('应该处理未知错误', () => {
      const unknownError = new Error('Something went wrong');

      const result = errorHandler.handleError(unknownError);

      expect(result.type).toBe(ErrorType.UNKNOWN_ERROR);
      expect(result.severity).toBe(ErrorSeverity.MEDIUM);
      expect(result.userMessage).toBe('系统异常，请稍后重试');
      expect(result.retryable).toBe(false);
    });
  });

  describe('isRetryableError', () => {
    it('应该正确识别可重试的错误', () => {
      const retryableError = {
        type: ErrorType.NETWORK_ERROR,
        retryable: true
      } as StandardError;

      expect(errorHandler.isRetryableError(retryableError)).toBe(true);
    });

    it('应该正确识别不可重试的错误', () => {
      const nonRetryableError = {
        type: ErrorType.VALIDATION_ERROR,
        retryable: false
      } as StandardError;

      expect(errorHandler.isRetryableError(nonRetryableError)).toBe(false);
    });
  });

  describe('getRetryDelay', () => {
    it('应该计算指数退避延迟', () => {
      expect(errorHandler.getRetryDelay(1)).toBe(100);
      expect(errorHandler.getRetryDelay(2)).toBe(200);
      expect(errorHandler.getRetryDelay(3)).toBe(400);
    });
  });
});

describe('便捷函数', () => {
  describe('handleError', () => {
    it('应该使用默认错误处理器处理错误', () => {
      const error = new Error('Test error');
      const result = handleError(error);

      expect(result).toHaveProperty('type');
      expect(result).toHaveProperty('severity');
      expect(result).toHaveProperty('userMessage');
      expect(result).toHaveProperty('timestamp');
      expect(result).toHaveProperty('requestId');
    });
  });

  describe('shouldRetry', () => {
    it('应该正确判断是否应该重试', () => {
      const retryableError = {
        retryable: true
      } as StandardError;

      const nonRetryableError = {
        retryable: false
      } as StandardError;

      expect(shouldRetry(retryableError, 1)).toBe(true);
      expect(shouldRetry(retryableError, 4)).toBe(false); // 超过最大重试次数
      expect(shouldRetry(nonRetryableError, 1)).toBe(false);
    });
  });

  describe('getUserFriendlyMessage', () => {
    it('应该返回用户友好的错误消息', () => {
      const error = {
        userMessage: '这是一个用户友好的错误消息'
      } as StandardError;

      expect(getUserFriendlyMessage(error)).toBe('这是一个用户友好的错误消息');
    });
  });
});