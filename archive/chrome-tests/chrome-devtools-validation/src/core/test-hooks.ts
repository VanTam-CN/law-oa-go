/**
 * 测试钩子系统实现
 *
 * 提供测试生命周期的钩子函数支持，包括：
 * - 全局钩子（beforeAll, afterAll）
 * - 套件钩子（beforeSuite, afterSuite）
 * - 用例钩子（beforeTest, afterTest）
 * - 步骤钩子（beforeStep, afterStep）
 */

import {
  TestHooks,
  TestExecutionContext
} from '../types/test-engine-types';
import { TestSuite, TestCase, TestStep } from '../types/test-types';
import { Logger } from '../core/logger';

/**
 * 钩子函数执行结果
 */
interface HookExecutionResult {
  success: boolean;
  error?: Error;
  duration: number;
}

/**
 * 测试钩子管理器
 */
export class TestHookManager {
  private hooks: TestHooks = {};
  private logger: Logger;
  private hookResults: Map<string, HookExecutionResult[]> = new Map();

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('TestHookManager');
  }

  /**
   * 设置钩子函数
   */
  setHook(name: keyof TestHooks, hookFn: any): void {
    this.hooks[name] = hookFn;
    this.logger.debug('设置钩子函数', { hookName: name });
  }

  /**
   * 移除钩子函数
   */
  removeHook(name: keyof TestHooks): void {
    delete this.hooks[name];
    this.logger.debug('移除钩子函数', { hookName: name });
  }

  /**
   * 设置多个钩子函数
   */
  setHooks(hooks: Partial<TestHooks>): void {
    this.hooks = { ...this.hooks, ...hooks };
    this.logger.debug('设置多个钩子函数', { hookNames: Object.keys(hooks) });
  }

  /**
   * 获取所有钩子函数
   */
  getHooks(): TestHooks {
    return { ...this.hooks };
  }

  /**
   * 执行 beforeAll 钩子
   */
  override async executeBeforeAll(context: TestExecutionContext): Promise<void> {
    await this.executeHook('beforeAll', this.hooks.beforeAll, context);
  }

  /**
   * 执行 afterAll 钩子
   */
  override async executeAfterAll(context: TestExecutionContext): Promise<void> {
    await this.executeHook('afterAll', this.hooks.afterAll, context);
  }

  /**
   * 执行 beforeSuite 钩子
   */
  override async executeBeforeSuite(suite: TestSuite, context: TestExecutionContext): Promise<void> {
    await this.executeHook('beforeSuite', this.hooks.beforeSuite, suite, context);
  }

  /**
   * 执行 afterSuite 钩子
   */
  override async executeAfterSuite(suite: TestSuite, context: TestExecutionContext): Promise<void> {
    await this.executeHook('afterSuite', this.hooks.afterSuite, suite, context);
  }

  /**
   * 执行 beforeTest 钩子
   */
  override async executeBeforeTest(testCase: TestCase, context: TestExecutionContext): Promise<void> {
    await this.executeHook('beforeTest', this.hooks.beforeTest, testCase, context);
  }

  /**
   * 执行 afterTest 钩子
   */
  override async executeAfterTest(testCase: TestCase, context: TestExecutionContext): Promise<void> {
    await this.executeHook('afterTest', this.hooks.afterTest, testCase, context);
  }

  /**
   * 执行 beforeStep 钩子
   */
  override async executeBeforeStep(step: TestStep, context: TestExecutionContext): Promise<void> {
    await this.executeHook('beforeStep', this.hooks.beforeStep, step, context);
  }

  /**
   * 执行 afterStep 钩子
   */
  override async executeAfterStep(step: TestStep, context: TestExecutionContext): Promise<void> {
    await this.executeHook('afterStep', this.hooks.afterStep, step, context);
  }

  /**
   * 获取钩子执行结果
   */
  getHookResults(hookName: string): HookExecutionResult[] {
    return this.hookResults.get(hookName) || [];
  }

  /**
   * 清空钩子执行结果
   */
  clearHookResults(): void {
    this.hookResults.clear();
  }

  /**
   * 执行钩子函数
   */
  private override async executeHook(
    name: string,
    hookFn?: Function,
    ...args: any[]
  ): Promise<void> {
    if (!hookFn) {
      this.logger.trace('钩子函数未定义，跳过执行', { hookName: name });
      return;
    }

    const startTime = Date.now();
    this.logger.trace('开始执行钩子函数', { hookName: name, argsCount: args.length });

    try {
      await hookFn(...args);

      const duration = Date.now() - startTime;
      this.recordHookResult(name, { success: true, duration });

      this.logger.trace('钩子函数执行成功', {
        hookName: name,
        duration: `${duration}ms`
      });

    } catch (error) {
      const duration = Date.now() - startTime;
      const hookError = error instanceof Error ? error : new Error(String(error));

      this.recordHookResult(name, { success: false, error: hookError, duration });

      this.logger.error('钩子函数执行失败', {
        hookName: name,
        duration: `${duration}ms`,
        error: hookError.message
      });

      throw hookError;
    }
  }

  /**
   * 记录钩子执行结果
   */
  private recordHookResult(name: string, result: HookExecutionResult): void {
    if (!this.hookResults.has(name)) {
      this.hookResults.set(name, []);
    }
    this.hookResults.get(name)!.push(result);
  }
}

/**
 * 预定义钩子函数集合
 */
export class PredefinedHooks {
  /**
   * 创建数据初始化钩子
   */
  static createDataInitializationHook(dataProvider: any) {
    return async (context: TestExecutionContext) => {
      if (dataProvider && typeof dataProvider.initialize === 'function') {
        await dataProvider.initialize();
        context.sharedData.set('dataProviderInitialized', true);
      }
    };
  }

  /**
   * 创建数据清理钩子
   */
  static createDataCleanupHook(dataProvider: any) {
    return async (context: TestExecutionContext) => {
      if (dataProvider && typeof dataProvider.cleanup === 'function') {
        await dataProvider.cleanup();
        context.sharedData.set('dataProviderInitialized', false);
      }
    };
  }

  /**
   * 创建会话开始钩子
   */
  static createSessionStartHook() {
    return async (context: TestExecutionContext) => {
      const sessionId = `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
      context.sharedData.set('sessionId', sessionId);
      context.sharedData.set('sessionStartTime', new Date());
    };
  }

  /**
   * 创建会话结束钩子
   */
  static createSessionEndHook() {
    return async (context: TestExecutionContext) => {
      const sessionId = context.sharedData.get('sessionId');
      const startTime = context.sharedData.get('sessionStartTime') as Date;

      if (sessionId && startTime) {
        const duration = Date.now() - startTime.getTime();
        context.sharedData.set('sessionDuration', duration);
        context.sharedData.delete('sessionId');
        context.sharedData.delete('sessionStartTime');
      }
    };
  }

  /**
   * 创建截图钩子
   */
  static createScreenshotHook(filename?: string) {
    return async (_step: TestStep, context: TestExecutionContext) => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      try {
        const screenshotName = filename || `hook-screenshot-${Date.now()}.png`;
        await (context.validator as any).screenshot({ filename: screenshotName });
        context.sharedData.set('lastScreenshot', screenshotName);
      } catch (error) {
        // 截图失败不影响测试执行
        context.sharedData.set('screenshotError', error instanceof Error ? error.message : String(error));
      }
    };
  }

  /**
   * 创建性能监控钩子
   */
  static createPerformanceMonitorHook() {
    return async (context: TestExecutionContext) => {
      const performanceData = {
        memoryUsage: process.memoryUsage(),
        cpuUsage: process.cpuUsage(),
        timestamp: new Date()
      };
      context.sharedData.set('performanceData', performanceData);
    };
  }

  /**
   * 创建超时保护钩子
   */
  static createTimeoutProtectionHook(timeout: number = 30000) {
    return async (_testCase: TestCase, context: TestExecutionContext) => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const timeoutPromise = new Promise((_, reject) => {
        setTimeout(() => reject(new Error(`测试用例超时: ${_testCase.name}`)), timeout);
      });

      // 将超时Promise存储在上下文中，供测试引擎使用
      context.sharedData.set('timeoutPromise', timeoutPromise);
    };
  }

  /**
   * 创建重试逻辑钩子
   */
  static createRetryLogicHook(maxRetries: number = 3) {
    return async (_testCase: TestCase, context: TestExecutionContext) => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const retryCount = context.sharedData.get('retryCount') as number || 0;

      if (retryCount > 0) {
        context.sharedData.set('lastRetryTime', new Date());
        context.sharedData.set('totalRetries', retryCount);
      }

      context.sharedData.set('maxRetries', maxRetries);
    };
  }

  /**
   * 创建环境检查钩子
   */
  static createEnvironmentCheckHook() {
    return async (context: TestExecutionContext) => {
      const envInfo = {
        nodeVersion: process.version,
        platform: process.platform,
        arch: process.arch,
        testEnvironment: context.config.testEnvironment,
        executionId: context.executionId
      };
      context.sharedData.set('environmentInfo', envInfo);
    };
  }

  /**
   * 创建日志记录钩子
   */
  static createLoggingHook(customLogger?: Logger) {
    const logger = customLogger || new Logger('HookLogger');

    return async (...args: any[]) => {
      const hookName = args[0]?.getconstructor?.().name || 'Unknown';
      const context = args.find((arg): arg is TestExecutionContext =>
        arg && arg.executionId && arg.validator
      );

      if (context) {
        logger.debug('钩子函数执行', {
          hookName,
          executionId: context.executionId,
          timestamp: new Date().toISOString()
        });
      }
    };
  }

  /**
   * 创建数据验证钩子
   */
  static createDataValidationHook(validationRules: Record<string, any>) {
    return async (_testCase: TestCase, context: TestExecutionContext) => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const testData = context.sharedData.get('testData');

      if (testData && validationRules) {
        const errors: string[] | undefined = undefined;

        for (const [key, rule] of Object.entries(validationRules)) {
          const value = testData[key];

          if (rule.required && (value === undefined || value === null)) {
            errors.push(`缺少必需字段: ${key}`);
          } else if (rule.type && typeof value !== rule.type) {
            errors.push(`字段类型错误: ${key} 期望 ${rule.type}, 实际 ${typeof value}`);
          } else if (rule.validator && typeof rule.validator === 'function') {
            try {
              const isValid = rule.validator(value);
              if (!isValid) {
                errors.push(`字段验证失败: ${key}`);
              }
            } catch (error) {
              errors.push(`字段验证异常: ${key} - ${error instanceof Error ? error.message : error}`);
            }
          }
        }

        if (errors.length > 0) {
          context.sharedData.set('validationErrors', errors);
          throw new Error(`数据验证失败: ${errors.join(', ')}`);
        }

        context.sharedData.set('validationPassed', true);
      }
    };
  }

  /**
   * 创建报告生成钩子
   */
  static createReportGenerationHook(reportGenerator?: any) {
    return async (context: TestExecutionContext) => {
      if (reportGenerator && typeof reportGenerator.generateIntermediateReport === 'function') {
        try {
          const intermediateData = {
            executionId: context.executionId,
            sharedData: Object.fromEntries(context.sharedData),
            timestamp: new Date()
          };

          const report = await reportGenerator.generateIntermediateReport(intermediateData);
          context.sharedData.set('intermediateReport', report);
        } catch (error) {
          context.sharedData.set('reportGenerationError', error instanceof Error ? error.message : String(error));
        }
      }
    };
  }
}

/**
 * 钩子函数构建器
 */
export class HookBuilder {
  private hooks: Partial<TestHooks> = {};

  /**
   * 添加 beforeAll 钩子
   */
  beforeAll(hookFn: (context: TestExecutionContext) => Promise<void>): this {
    this.hooks.beforeAll = hookFn;
    return this;
  }

  /**
   * 添加 afterAll 钩子
   */
  afterAll(hookFn: (context: TestExecutionContext) => Promise<void>): this {
    this.hooks.afterAll = hookFn;
    return this;
  }

  /**
   * 添加 beforeSuite 钩子
   */
  beforeSuite(hookFn: (suite: TestSuite, context: TestExecutionContext) => Promise<void>): this {
    this.hooks.beforeSuite = hookFn;
    return this;
  }

  /**
   * 添加 afterSuite 钩子
   */
  afterSuite(hookFn: (suite: TestSuite, context: TestExecutionContext) => Promise<void>): this {
    this.hooks.afterSuite = hookFn;
    return this;
  }

  /**
   * 添加 beforeTest 钩子
   */
  beforeTest(hookFn: (testCase: TestCase, context: TestExecutionContext) => Promise<void>): this {
    this.hooks.beforeTest = hookFn;
    return this;
  }

  /**
   * 添加 afterTest 钩子
   */
  afterTest(hookFn: (testCase: TestCase, context: TestExecutionContext) => Promise<void>): this {
    this.hooks.afterTest = hookFn;
    return this;
  }

  /**
   * 添加 beforeStep 钩子
   */
  beforeStep(hookFn: (step: TestStep, context: TestExecutionContext) => Promise<void>): this {
    this.hooks.beforeStep = hookFn;
    return this;
  }

  /**
   * 添加 afterStep 钩子
   */
  afterStep(hookFn: (step: TestStep, context: TestExecutionContext) => Promise<void>): this {
    this.hooks.afterStep = hookFn;
    return this;
  }

  /**
   * 添加预定义数据初始化钩子
   */
  withDataInitialization(dataProvider: any): this {
    this.hooks.beforeAll = PredefinedHooks.createDataInitializationHook(dataProvider);
    this.hooks.afterAll = PredefinedHooks.createDataCleanupHook(dataProvider);
    return this;
  }

  /**
   * 添加预定义会话管理钩子
   */
  withSessionManagement(): this {
    this.hooks.beforeAll = PredefinedHooks.createSessionStartHook();
    this.hooks.afterAll = PredefinedHooks.createSessionEndHook();
    return this;
  }

  /**
   * 添加预定义性能监控钩子
   */
  withPerformanceMonitoring(): this {
    this.hooks.beforeAll = PredefinedHooks.createPerformanceMonitorHook();
    return this;
  }

  /**
   * 添加预定义超时保护钩子
   */
  withTimeoutProtection(timeout: number): this {
    this.hooks.beforeTest = PredefinedHooks.createTimeoutProtectionHook(timeout);
    return this;
  }

  /**
   * 添加预定义重试逻辑钩子
   */
  withRetryLogic(maxRetries: number): this {
    this.hooks.beforeTest = PredefinedHooks.createRetryLogicHook(maxRetries);
    return this;
  }

  /**
   * 添加预定义环境检查钩子
   */
  withEnvironmentCheck(): this {
    this.hooks.beforeAll = PredefinedHooks.createEnvironmentCheckHook();
    return this;
  }

  /**
   * 添加预定义日志记录钩子
   */
  withLogging(logger?: Logger): this {
    const loggingHook = PredefinedHooks.createLoggingHook(logger);
    this.hooks.beforeAll = loggingHook;
    this.hooks.afterAll = loggingHook;
    this.hooks.beforeSuite = loggingHook;
    this.hooks.afterSuite = loggingHook;
    this.hooks.beforeTest = loggingHook;
    this.hooks.afterTest = loggingHook;
    this.hooks.beforeStep = loggingHook;
    this.hooks.afterStep = loggingHook;
    return this;
  }

  /**
   * 构建钩子集合
   */
  build(): Partial<TestHooks> {
    return { ...this.hooks };
  }

  /**
   * 重置构建器
   */
  reset(): this {
    this.hooks = {};
    return this;
  }
}