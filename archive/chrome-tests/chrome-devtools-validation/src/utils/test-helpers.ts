/**
 * 测试工具类 - 提供通用的测试辅助功能
 */

import { Logger } from '../core/logger';
import { TestExecutionResult } from '../types/test-types';
import { TestCase } from '../types/test-types';
export interface TestTimeoutConfig {
  element: number;
  page: number;
  assertion: number;
  global: number;
}

export interface RetryConfig {
  attempts: number;
  delay: number;
  backoffFactor: number;
}

export interface TestMetrics {
  startTime: number;
  endTime?: number;
  stepCount: number;
  assertionCount: number;
  errorCount: number;
  retryCount: number;
}

export class TestHelpers {
  private logger: Logger;
  private timeoutConfig: TestTimeoutConfig;
  private retryConfig: RetryConfig;

  constructor(
    timeoutConfig: TestTimeoutConfig,
    retryConfig: RetryConfig,
    logger?: Logger
  ) {
    this.timeoutConfig = timeoutConfig;
    this.retryConfig = retryConfig;
    this.logger = logger || new Logger('TestHelpers');
  }

  /**
   * 生成唯一的测试ID
   */
  generateTestId(prefix: string = 'test'): string {
    const timestamp = Date.now();
    const random = Math.floor(Math.random() * 1000);
    return `${prefix}-${timestamp}-${random}`;
  }

  /**
   * 生成唯一的步骤ID
   */
  generateStepId(testId: string, stepIndex: number): string {
    return `${testId}-step-${stepIndex}`;
  }

  /**
   * 生成唯一的断言ID
   */
  generateAssertionId(testId: string, assertionIndex: number): string {
    return `${testId}-assert-${assertionIndex}`;
  }

  /**
   * 格式化持续时间
   */
  formatDuration(milliseconds: number): string {
    if (milliseconds < 1000) {
      return `${milliseconds}ms`;
    } else if (milliseconds < 60000) {
      return `${(milliseconds / 1000).toFixed(2)}s`;
    } else {
      const minutes = Math.floor(milliseconds / 60000);
      const seconds = ((milliseconds % 60000) / 1000).toFixed(2);
      return `${minutes}m ${seconds}s`;
    }
  }

  /**
   * 计算成功率
   */
  calculateSuccessRate(passed: number, total: number): number {
    if (total === 0) return 0;
    return Math.round((passed / total) * 100);
  }

  /**
   * 验证测试用例结构
   */
  validateTestCase(testCase: TestCase): { valid: boolean; errors: string[] } {
    let errors: string[] = [];

    if (!testCase.id || testCase.id.trim() === '') {
      errors.push('测试用例ID不能为空');
    }

    if (!testCase.name || testCase.name.trim() === '') {
      errors.push('测试用例名称不能为空');
    }

    if (!testCase.steps || testCase.steps.length === 0) {
      errors.push('测试用例必须包含至少一个步骤');
    }

    // 验证步骤
    if (testCase.steps) {
      testCase.steps.forEach((step: any, index: any) => {
        if (!step.id || step.id.trim() === '') {
          errors.push(`步骤 ${index + 1} 的ID不能为空`);
        }

        if (!step.name || step.name.trim() === '') {
          errors.push(`步骤 ${index + 1} 的名称不能为空`);
        }

        if (!step.type || !this.isValidStepType(step.type)) {
          errors.push(`步骤 ${index + 1} 的类型无效: ${step.type}`);
        }

        // 根据步骤类型验证必需属性
        if (step.type === 'navigate' && !step.url) {
          errors.push(`导航步骤 ${step.name} 必须包含URL`);
        }

        if (step.type === 'fill' && !step.selector) {
          errors.push(`输入步骤 ${step.name} 必须包含选择器`);
        }

        if (step.type === 'click' && !step.selector) {
          errors.push(`点击步骤 ${step.name} 必须包含选择器`);
        }

        if (step.type === 'executeScript' && !step.script) {
          errors.push(`脚本执行步骤 ${step.name} 必须包含脚本内容`);
        }
      });
    }

    // 验证断言
    if (testCase.assertions) {
      testCase.assertions.forEach((assertion: any, index: any) => {
        if (!assertion.id || assertion.id.trim() === '') {
          errors.push(`断言 ${index + 1} 的ID不能为空`);
        }

        if (!assertion.type || !this.isValidAssertionType(assertion.type)) {
          errors.push(`断言 ${index + 1} 的类型无效: ${assertion.type}`);
        }

        if (!assertion.selector) {
          errors.push(`断言 ${index + 1} 必须包含选择器`);
        }
      });
    }

    return { valid: (errors || []).length === 0, errors };
  }

  /**
   * 验证步骤类型
   */
  private isValidStepType(type: string): boolean {
    const validTypes = ['navigate', 'click', 'fill', 'select', 'wait', 'verify', 'screenshot', 'executeScript'];
    return validTypes.includes(type);
  }

  /**
   * 验证断言类型
   */
  private isValidAssertionType(type: string): boolean {
    const validTypes = ['element-exists', 'element-visible', 'element-enabled', 'text-contains', 'value-equals', 'url-contains'];
    return validTypes.includes(type);
  }

  /**
   * 延迟执行
   */
  async delay(milliseconds: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, milliseconds));
  }

  /**
   * 重试机制
   */
  async retry<T>(
    operation: () => Promise<T>,
    operationName: string,
    config?: Partial<RetryConfig>
  ): Promise<{ success: boolean; result?: T; error?: Error; attempts: number }> {
    const retryConf = { ...this.retryConfig, ...config };
    let lastError: Error | undefined;

    for (let attempt = 1; attempt <= retryConf.attempts; attempt++) {
      try {
        this.logger.debug(`重试操作第${attempt}次`, { operationName });
        const result = await operation();
        this.logger.debug(`重试操作成功`, { operationName, attempts: attempt });
        return { success: true, result, attempts: attempt };
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));
        this.logger.warn(`重试操作失败`, { operationName, attempt, error: lastError.message });

        if (attempt < retryConf.attempts) {
          const delay = retryConf.delay * Math.pow(retryConf.backoffFactor, attempt - 1);
          await this.delay(delay);
        }
      }
    }

    return { success: false, error: lastError, attempts: retryConf.attempts };
  }

  /**
   * 创建测试指标
   */
  createMetrics(): TestMetrics {
    return {
      startTime: Date.now(),
      stepCount: 0,
      assertionCount: 0,
      errorCount: 0,
      retryCount: 0
    };
  }

  /**
   * 更新测试指标
   */
  updateMetrics(metrics: TestMetrics, updates: Partial<TestMetrics>): TestMetrics {
    return { ...metrics, ...updates };
  }

  /**
   * 完成测试指标
   */
  completeMetrics(metrics: TestMetrics): TestMetrics {
    return {
      ...metrics,
      endTime: Date.now()
    };
  }

  /**
   * 格式化测试结果摘要
   */
  formatResultSummary(result: TestExecutionResult): string {
    const duration = result.duration || 0;
    const successRate = this.calculateSuccessRate(result.passedTests, result.totalTests);

    return `
测试执行摘要:
================
总测试数: ${result.totalTests}
通过: ${result.passedTests}
失败: ${result.failedTests}
跳过: ${result.skippedTests}
成功率: ${successRate}%
执行时间: ${this.formatDuration(duration)}
错误数: ${result.error.length}
    `.trim();
  }

  /**
   * 截断长文本
   */
  truncateText(text: string, maxLength: number = 100): string {
    if (text.length <= maxLength) {
      return text;
    }
    return text.substring(0, maxLength - 3) + '...';
  }

  /**
   * 安全的JSON解析
   */
  safeJsonParse<T>(jsonString: string, defaultValue: T): T {
    try {
      return JSON.parse(jsonString);
    } catch (error) {
      this.logger.warn('JSON解析失败', { jsonString: this.truncateText(jsonString), error: error instanceof Error ? error.message : error });
      return defaultValue;
    }
  }

  /**
   * 安全的JSON字符串化
   */
  safeJsonStringify(obj: any, indent?: number): string {
    try {
      return JSON.stringify(obj, null, indent);
    } catch (error) {
      this.logger.warn('JSON字符串化失败', { error: error instanceof Error ? error.message : error });
      return '{}';
    }
  }

  /**
   * 获取超时配置
   */
  getTimeoutConfig(): TestTimeoutConfig {
    return { ...this.timeoutConfig };
  }

  /**
   * 获取重试配置
   */
  getRetryConfig(): RetryConfig {
    return { ...this.retryConfig };
  }

  /**
   * 验证URL格式
   */
  isValidUrl(url: string): boolean {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * 规范化URL
   */
  normalizeUrl(url: string, baseUrl?: string): string {
    if (this.isValidUrl(url)) {
      return url;
    }

    if (baseUrl && url.startsWith('/')) {
      return new URL(url, baseUrl).href;
    }

    return url;
  }

  /**
   * 生成随机字符串
   */
  generateRandomString(length: number = 8): string {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  }

  /**
   * 生成随机邮箱
   */
  generateRandomEmail(domain: string = 'example.com'): string {
    const username = this.generateRandomString(10);
    return `${username}@${domain}`;
  }

  /**
   * 获取当前时间戳
   */
  getTimestamp(): string {
    return new Date().toISOString();
  }

  /**
   * 格式化文件大小
   */
  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
}