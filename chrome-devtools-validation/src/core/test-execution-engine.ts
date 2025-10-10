/**
 * 测试执行引擎 - 核心实现
 *
 * 提供完整的测试执行引擎，包括：
 * - 测试用例执行器
 * - 测试套件执行器
 * - 测试步骤执行器
 * - 执行上下文管理
 * - 错误处理和重试机制
 * - 并行执行支持
 * - 钩子函数支持
 * - 监听器支持
 * - 报告生成支持
 */

import {
  TestExecutor,
  TestExecutionConfig,
  TestExecutionContext,
  TestExecutionResult,
  TestSuiteResult,
  TestCaseResult,
  TestStepResult,
  TestExecutionError,
  TestHooks,
  TestListener
} from '../types/test-engine-types';
import { TestSuite, TestCase, TestStep } from '../types/test-types';
import { TestExecutionResult } from '../types/test-types';
import { TestCase } from '../types/test-types';
import { ChromeDevToolsValidator } from '../index';
import { Logger } from '../core/logger';

/**
 * 默认测试执行配置
 */
export const DEFAULT_EXECUTION_CONFIG: TestExecutionConfig = {
  headless: false,
  slowMo: 0,
  defaultTimeout: 30000,
  viewport: { width: 1920, height: 1080 },
  parallelExecution: false,
  maxConcurrency: 3,
  retryFailedTests: true,
  maxRetries: 2,
  generateReport: true,
  reportFormat: 'json',
  screenshotsOnFailure: true,
  baseUrl: '',
  testEnvironment: 'development'
};

/**
 * 测试执行引擎实现
 */
export class ChromeDevToolsTestExecutionEngine implements TestExecutor {
  private logger: Logger;
  private config: TestExecutionConfig;
  private listeners: TestListener[] | undefined = undefined;
  private reportGenerator?: any;
  private dataProvider?: any;
  private context?: TestExecutionContext;

  constructor(config: Partial<TestExecutionConfig> = {}, logger?: Logger) {
    this.config = { ...DEFAULT_EXECUTION_CONFIG, ...config };
    this.logger = logger || new Logger('TestExecutionEngine');
  }

  /**
   * 添加测试监听器
   */
  addListener(listener: TestListener): void {
    this.listeners.push(listener);
  }

  /**
   * 移除测试监听器
   */
  removeListener(listener: TestListener): void {
    const index = this.listeners.indexOf(listener);
    if (index > -1) {
      this.listeners.splice(index, 1);
    }
  }

  /**
   * 设置报告生成器
   */
  setReportGenerator(generator: any): void {
    this.reportGenerator = generator;
  }

  /**
   * 设置数据提供者
   */
  setDataProvider(_provider: any): void {
    // this.dataProvider = provider;
  }

  /**
   * 设置钩子函数
   */
  setHooks(hooks: TestHooks): void {
    if (this.context) {
      this.context.hooks = { ...this.context.hooks, ...hooks };
    }
  }

  /**
   * 获取配置
   */
  getConfig(): TestExecutionConfig {
    return { ...this.config };
  }

  /**
   * 更新配置
   */
  updateConfig(config: Partial<TestExecutionConfig>): void {
    this.config = { ...this.config, ...config };
  }

  /**
   * 执行测试套件（实现TestExecutor接口）
   */
  async (suite: TestSuite, context: TestExecutionContext): Promise<any> {
    this.logger.info('开始执行测试套件', { suiteId: suite.id, suiteName: suite.name });

    const startTime = new Date();
    const executionId = this.generateExecutionId();

    // 创建执行上下文
    if (!context) {
      const validator = new ChromeDevToolsValidator();

      await validator.initialize();

      context = {
        executionId,
        startTime,
        config: this.config,
        validator,
        sharedData: new Map(),
        hooks: {}
      };
    }

    this.context = context;

    try {
      // 执行全局前置钩子
      if (context.hooks.beforeAll) {
        await this.executeHook('beforeAll', () => context.hooks.beforeAll!(context));
      }

      // 通知监听器执行开始
      await this.notifyListeners('onExecutionStart', context);

      // 执行测试套件
      const suiteResult = await this.executeSuiteInternal(suite, context);

      // 执行全局后置钩子
      if (context.hooks.afterAll) {
        await this.executeHook('afterAll', () => context.hooks.afterAll!(context));
      }

      // 创建执行结果
      const endTime = new Date();
      const result: TestExecutionResult = {
        executionId,
        startTime,
        endTime,
        duration: endTime.getTime() - startTime.getTime(),
        config: this.config,
        total: this.countTotalTests(suite),
        passed: this.countTestsByStatus(suiteResult, 'passed'),
        failed: this.countTestsByStatus(suiteResult, 'failed'),
        skipped: this.countTestsByStatus(suiteResult, 'skipped'),
        pending: this.countTestsByStatus(suiteResult, 'pending'),
        suites: [suiteResult],
        errors: [],
        metadata: {
          environment: this.config.testEnvironment || 'development',
          browser: 'Chrome',
          version: '120.0.0',
          executor: 'ChromeDevToolsValidationEngine'
        }
      };

      // 通知监听器执行结束
      await this.notifyListeners('onExecutionEnd', result);

      // 生成报告
      if (this.config.generateReport && this.reportGenerator) {
        await this.generateReport(result);
      }

      this.logger.info('测试套件执行完成', {
        executionId,
        total: result.total,
        passed: result.passed,
        failed: result.failed,
        duration: result.duration
      });

      return result;
    } catch (error) {
      const endTime = new Date();
      const executionError: TestExecutionError = {
        timestamp: new Date(),
        type: 'system',
        message: error instanceof Error ? error.message : '未知错误',
        stack: error instanceof Error ? error.stack || '' : '',
        component: 'TestExecutionEngine'
      };

      const result: TestExecutionResult = {
        executionId,
        startTime,
        endTime,
        duration: endTime.getTime() - startTime.getTime(),
        config: this.config,
        total: 0,
        passed: 0,
        failed: 0,
        skipped: 0,
        pending: 0,
        suites: [],
        errors: [executionError],
        metadata: {
          environment: this.config.testEnvironment || 'development',
          browser: 'Chrome',
          version: '120.0.0',
          executor: 'ChromeDevToolsValidationEngine'
        }
      };

      await this.notifyListeners('onExecutionEnd', result);
      return result;
    }
  }

  /**
   * 创建测试执行错误对象
   */
  private createTestExecutionError(message: string, stack?: string, screenshot?: string): { message: string; stack?: string; screenshot?: string } {
    const error: { message: string; stack?: string; screenshot?: string } = {
      message
    };

    if (stack !== undefined) {
      error.stack = stack;
    }

    if (screenshot !== undefined) {
      error.screenshot = screenshot;
    }

    return error;
  }

  /**
   * 执行测试用例（实现TestExecutor接口）
   */
  async (testCase: TestCase, context: TestExecutionContext): Promise<any> {
    this.logger.info('开始执行测试用例', { caseId: testCase.id, caseName: testCase.name });

    const startTime = new Date();
    const executionId = this.generateExecutionId();

    try {
      // 创建测试用例的执行上下文
      const caseContext: TestExecutionContext = {
        ...context,
        executionId,
        startTime,
        currentCase: testCase
      };

      // 通知监听器测试用例开始
      await this.notifyListeners('onTestCaseStart', testCase, caseContext);

      // 执行测试用例
      const caseResult = await this.executeTestCaseInternal(testCase, caseContext);

      // 通知监听器测试用例结束
      await this.notifyListeners('onTestCaseEnd', caseResult);

      const endTime = new Date();
      const result: TestExecutionResult = {
        executionId,
        startTime,
        endTime,
        duration: endTime.getTime() - startTime.getTime(),
        config: this.config,
        total: 1,
        passed: caseResult.status === 'passed' ? 1 : 0,
        failed: caseResult.status === 'failed' ? 1 : 0,
        skipped: caseResult.status === 'skipped' ? 1 : 0,
        pending: caseResult.status === 'pending' ? 1 : 0,
        suites: [{
          suiteId: 'single-case',
          name: 'Single Test Case',
          status: caseResult.status,
          startTime,
          endTime,
          duration: endTime.getTime() - startTime.getTime(),
          testCases: [caseResult]
        }],
        errors: [],
        metadata: {
          environment: this.config.testEnvironment || 'development',
          browser: 'Chrome',
          version: '120.0.0',
          executor: 'ChromeDevToolsValidationEngine'
        }
      };

      return result;
    } catch (error) {
      const endTime = new Date();
      const executionError: TestExecutionError = {
        timestamp: new Date(),
        type: 'execution',
        message: error instanceof Error ? error.message : '未知错误',
        stack: error instanceof Error ? error.stack || '' : '',
        component: 'TestCaseExecutor'
      };

      const result: TestExecutionResult = {
        executionId,
        startTime,
        endTime,
        duration: endTime.getTime() - startTime.getTime(),
        config: this.config,
        total: 1,
        passed: 0,
        failed: 1,
        skipped: 0,
        pending: 0,
        suites: [{
          suiteId: 'single-case',
          name: 'Single Test Case',
          status: 'failed',
          startTime,
          endTime,
          duration: endTime.getTime() - startTime.getTime(),
          testCases: [{
            caseId: testCase.id,
            name: testCase.name,
            status: 'failed',
            startTime,
            endTime,
            duration: endTime.getTime() - startTime.getTime(),
            steps: [],
            error: this.createTestExecutionError(executionError.message, executionError.stack)
          }]
        }],
        errors: [executionError],
        metadata: {
          environment: this.config.testEnvironment || 'development',
          browser: 'Chrome',
          version: '120.0.0',
          executor: 'ChromeDevToolsValidationEngine'
        }
      };

      return result;
    }
  }

  /**
   * 执行测试步骤（实现TestExecutor接口）
   */
  async (step: TestStep, context: TestExecutionContext): Promise<TestStepResult> {
    this.logger.debug('开始执行测试步骤', { stepId: step.id, stepName: step.name });

    const startTime = new Date();
    const logs: any[] | undefined = undefined;

    try {
      // 更新上下文
      context.currentStep = step;

      // 执行步骤前置钩子
      if (context.hooks.beforeStep) {
        await this.executeHook('beforeStep', () => context.hooks.beforeStep!(step, context));
      }

      // 通知监听器测试步骤开始
      await this.notifyListeners('onTestStepStart', step, context);

      // 记录开始日志
      logs.push({
        timestamp: new Date(),
        level: 'info',
        message: `开始执行步骤: ${step.name}`,
        context: { stepId: step.id, type: step.type },
        stepId: step.id
      });

      // 执行步骤
      const stepResult = await this.executeSingleStep(step, context);

      // 记录成功日志
      logs.push({
        timestamp: new Date(),
        level: 'info',
        message: `步骤执行成功: ${step.name}`,
        context: { stepId: step.id, duration: stepResult.duration },
        stepId: step.id
      });

      // 执行步骤后置钩子
      if (context.hooks.afterStep) {
        await this.executeHook('afterStep', () => context.hooks.afterStep!(step, context));
      }

      // 通知监听器测试步骤结束
      await this.notifyListeners('onTestStepEnd', stepResult);

      const endTime = new Date();
      return {
        ...stepResult,
        startTime,
        endTime,
        duration: endTime.getTime() - startTime.getTime(),
        logs
      };
    } catch (error) {
      const endTime = new Date();
      const errorMessage = error instanceof Error ? error.message : '未知错误';

      // 记录错误日志
      logs.push({
        timestamp: new Date(),
        level: 'error',
        message: `步骤执行失败: ${step.name}`,
        context: {
          stepId: step.id,
          error: errorMessage,
          stack: error instanceof Error ? error.stack : undefined
        },
        stepId: step.id
      });

      const result: TestStepResult = {
        stepId: step.id,
        name: step.name,
        status: 'failed',
        duration: endTime.getTime() - startTime.getTime(),
        startTime,
        endTime,
        error: this.createTestExecutionError(
        errorMessage,
        error instanceof Error ? error.stack : undefined,
        this.config.screenshotsOnFailure ? await this.takeScreenshot(context) : undefined
      ),
        logs
      };

      // 通知监听器测试步骤结束
      await this.notifyListeners('onTestStepEnd', result);

      return result;
    }
  }

  /**
   * 执行单个测试套件
   */
  private async (suite: TestSuite, context: TestExecutionContext): Promise<TestSuiteResult> {
    const startTime = new Date();
    const testCases: TestCaseResult[] | undefined = undefined;

    try {
      // 执行套件前置钩子
      if (context.hooks.beforeSuite) {
        await this.executeHook('beforeSuite', () => context.hooks.beforeSuite!(suite, context));
      }

      // 通知监听器套件开始
      await this.notifyListeners('onSuiteStart', suite, context);

      // 更新上下文
      context.currentSuite = suite;

      // 执行测试用例
      if (this.config.parallelExecution && suite.testCases.length > 1) {
        // 并行执行
        const concurrency = Math.min(this.config.maxConcurrency || 3, suite.testCases.length);
        testCases.push(...await this.executeTestCasesParallel(suite.testCases, context, concurrency));
      } else {
        // 串行执行
        for (const testCase of suite.testCases) {
          const caseResult = await this.executeTestCaseInternal(testCase, context);
          testCases.push(caseResult);
        }
      }

      // 执行套件后置钩子
      if (context.hooks.afterSuite) {
        await this.executeHook('afterSuite', () => context.hooks.afterSuite!(suite, context));
      }

      const endTime = new Date();
      const status = this.determineSuiteStatus(testCases);

      const result: TestSuiteResult = {
        suiteId: suite.id,
        name: suite.name,
        description: suite.description,
        status,
        startTime,
        endTime,
        duration: endTime.getTime() - startTime.getTime(),
        testCases
      };

      // 通知监听器套件结束
      await this.notifyListeners('onSuiteEnd', result);

      return result;
    } catch (error) {
      const endTime = new Date();
      const errorMessage = error instanceof Error ? error.message : '未知错误';

      const result: TestSuiteResult = {
        suiteId: suite.id,
        name: suite.name,
        description: suite.description,
        status: 'failed',
        startTime,
        endTime,
        duration: endTime.getTime() - startTime.getTime(),
        testCases,
        setupError: errorMessage
      };

      // 通知监听器套件结束
      await this.notifyListeners('onSuiteEnd', result);

      return result;
    }
  }

  /**
   * 执行单个测试用例
   */
  private async (testCase: TestCase, context: TestExecutionContext): Promise<TestCaseResult> {
    const startTime = new Date();
    const steps: TestStepResult[] | undefined = undefined;
    let retryCount = 0;

    while (retryCount <= (this.config.maxRetries || 0)) {
      try {
        // 执行用例前置钩子
        if (context.hooks.beforeTest) {
          await this.executeHook('beforeTest', () => context.hooks.beforeTest!(testCase, context));
        }

        // 执行测试步骤
        for (const step of testCase.steps) {
          const stepResult = await this.executeTestStep(step, context);
          steps.push(stepResult);

          // 如果步骤失败，停止执行后续步骤
          if (stepResult.status === 'failed') {
            break;
          }
        }

        // 执行用例后置钩子
        if (context.hooks.afterTest) {
          await this.executeHook('afterTest', () => context.hooks.afterTest!(testCase, context));
        }

        // 执行成功，跳出重试循环
        break;
      } catch (error) {
        retryCount++;

        if (retryCount <= (this.config.maxRetries || 0)) {
          this.logger.warn(`测试用例重试 ${retryCount}/${this.config.maxRetries}`, {
            caseId: testCase.id,
            error: error instanceof Error ? error.message : error
          });

          // 等待一段时间后重试
          await this.delay(1000 * retryCount);
        } else {
          // 重试次数用完，记录错误
          throw error;
        }
      }
    }

    const endTime = new Date();
    const status = this.determineCaseStatus(steps);

    const result: TestCaseResult = {
      caseId: testCase.id,
      name: testCase.name,
      description: testCase.description,
      status,
      startTime,
      endTime,
      duration: endTime.getTime() - startTime.getTime(),
      steps,
      retryCount
    };

    return result;
  }

  /**
   * 并行执行测试用例
   */
  private async (
    testCases: TestCase[],
    context: TestExecutionContext,
    concurrency: number
  ): Promise<TestCaseResult[]> {
    const results: TestCaseResult[] | undefined = undefined;
    const batches = this.createBatches(testCases, concurrency);

    for (const batch of batches) {
      const batchPromises = batch.map(async (testCase) => {
        try {
          return await this.executeTestCaseInternal(testCase, context);
        } catch (error) {
          this.logger.error('并行执行测试用例失败', {
            caseId: testCase.id,
            error: error instanceof Error ? error.message : error
          });

          const endTime = new Date();
          return {
            caseId: testCase.id,
            name: testCase.name,
            description: testCase.description,
            status: 'failed' as const,
            startTime: new Date(),
            endTime,
            duration: 0,
            steps: [],
            error: this.createTestExecutionError(
              error instanceof Error ? error.message : '未知错误',
              error instanceof Error ? error.stack : undefined
            )
          };
        }
      });

      const batchResults = await Promise.all(batchPromises);
      results.push(...batchResults);
    }

    return results;
  }

  /**
   * 执行单个测试步骤
   */
  private async (step: TestStep, context: TestExecutionContext): Promise<TestStepResult> {
    const startTime = Date.now();

    try {
      // 根据步骤类型执行相应的操作
      switch (step.type) {
        case 'navigate':
          if (!step.url) throw new Error('导航步骤缺少URL');
          await (context.validator as any).navigate(step.url);
          break;
        case 'click':
          if (!step.selector) throw new Error('点击步骤缺少选择器');
          await (context.validator as any).click(step.selector);
          break;
        case 'fill':
          if (!step.selector || step.value === undefined) throw new Error('填写步骤缺少选择器或值');
          await (context.validator as any).fill(step.selector, String(step.value));
          break;
        case 'wait':
          const timeout = step.timeout || this.config.defaultTimeout || 5000;
          if (step.selector) {
            await (context.validator as any).wait(timeout, step.selector);
          } else {
            await new Promise(resolve => setTimeout(resolve, timeout));
          }
          break;
        case 'screenshot':
          await (context.validator as any).screenshot('fullpage', `screenshot-${step.id}`, undefined, false);
          break;
        case 'verify':
          if (!step.selector) throw new Error('验证步骤缺少选择器');
          await this.verifyElementState(step, context);
          break;
        case 'executeScript':
          if (!step.script) throw new Error('执行脚本步骤缺少脚本内容');
          await (context.validator as any).executeScript(step.script);
          break;
        default:
          throw new Error(`未知的步骤类型: ${step.type}`);
      }

      return {
        stepId: step.id,
        name: step.name,
        status: 'passed',
        duration: Date.now() - startTime,
        logs: []
      };
    } catch (error) {
      return {
        stepId: step.id,
        name: step.name,
        status: 'failed',
        duration: Date.now() - startTime,
        startTime: new Date(startTime),
        endTime: new Date(),
        error: this.createTestExecutionError(
          error instanceof Error ? error.message : '未知错误',
          error instanceof Error ? error.stack : undefined
        ),
        logs: []
      };
    }
  }

  /**
   * 执行钩子函数
   */
  private async executeHook(name: string, hookFn: () => Promise<void>): Promise<void> {
    try {
      this.logger.debug(`执行钩子函数: ${name}`);
      await hookFn();
    } catch (error) {
      this.logger.error(`钩子函数执行失败: ${name}`, {
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  /**
   * 通知监听器
   */
  private async (event: string, ...args: any[]): Promise<void> {
    const promises = this.listeners.map(async (listener) => {
      try {
        switch (event) {
          case 'onExecutionStart':
            await listener.onExecutionStart(args[0]);
            break;
          case 'onExecutionEnd':
            await listener.onExecutionEnd(args[0]);
            break;
          case 'onSuiteStart':
            await listener.onSuiteStart(args[0], args[1]);
            break;
          case 'onSuiteEnd':
            await listener.onSuiteEnd(args[0]);
            break;
          case 'onTestCaseStart':
            await listener.onTestCaseStart(args[0], args[1]);
            break;
          case 'onTestCaseEnd':
            await listener.onTestCaseEnd(args[0]);
            break;
          case 'onTestStepStart':
            await listener.onTestStepStart(args[0], args[1]);
            break;
          case 'onTestStepEnd':
            await listener.onTestStepEnd(args[0]);
            break;
        }
      } catch (error) {
        this.logger.error('监听器执行失败', {
          event,
          listener: listener.constructor.name,
          error: error instanceof Error ? error.message : error
        });
      }
    });

    await Promise.all(promises);
  }

  /**
   * 生成测试报告
   */
  private async (result: TestExecutionResult): Promise<void> {
    if (!this.reportGenerator) return;

    try {
      const report = await this.reportGenerator.generateReport(result);
      const filename = `test-report-${new Date().toISOString().replace(/[:.]/g, '-')}.${this.config.reportFormat}`;
      await this.reportGenerator.saveReport(report, filename);

      this.logger.info('测试报告生成完成', { filename });
    } catch (error) {
      this.logger.error('生成测试报告失败', {
        error: error instanceof Error ? error.message : error
      });
    }
  }

  /**
   * 截图
   */
  private async (context: TestExecutionContext): Promise<string> {
    try {
      return await (context.validator as any).screenshot({
        filename: `error-${Date.now()}.png`
      });
    } catch (error) {
      this.logger.error('截图失败', {
        error: error instanceof Error ? error.message : error
      });
      return '';
    }
  }

  /**
   * 辅助方法：确定套件状态
   */
  private determineSuiteStatus(testCases: TestCaseResult[]): 'passed' | 'failed' | 'skipped' | 'pending' {
    if (testCases.length === 0) return 'skipped';

    const hasFailures = testCases.some(tc => tc.status === 'failed');
    const allSkipped = testCases.every(tc => tc.status === 'skipped');
    const allPending = testCases.every(tc => tc.status === 'pending');

    if (hasFailures) return 'failed';
    if (allSkipped) return 'skipped';
    if (allPending) return 'pending';
    return 'passed';
  }

  /**
   * 辅助方法：确定用例状态
   */
  private determineCaseStatus(steps: TestStepResult[]): 'passed' | 'failed' | 'skipped' | 'pending' {
    if (steps.length === 0) return 'skipped';

    const hasFailures = steps.some(s => s.status === 'failed');
    const allSkipped = steps.every(s => s.status === 'skipped');
    const allPending = steps.every(s => s.status === 'pending');

    if (hasFailures) return 'failed';
    if (allSkipped) return 'skipped';
    if (allPending) return 'pending';
    return 'passed';
  }

  /**
   * 辅助方法：计算总测试数
   */
  private countTotalTests(suite: TestSuite): number {
    return suite.testCases.length;
  }

  /**
   * 辅助方法：按状态统计测试数
   */
  private countTestsByStatus(suiteResult: TestSuiteResult, status: string): number {
    return suiteResult.testCases.filter(tc => tc.status === status).length;
  }

  /**
   * 辅助方法：创建批次
   */
  private createBatches<T>(items: T[], batchSize: number): T[][] {
    const batches: T[][] = [];
    for (let i = 0; i < items.length; i += batchSize) {
      batches.push(items.slice(i, i + batchSize));
    }
    return batches;
  }

  /**
   * 辅助方法：生成执行ID
   */
  private generateExecutionId(): string {
    return `exec_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * 验证元素状态
   */
  private async (step: TestStep, context: TestExecutionContext): Promise<void> {
    const validator = context.validator;

    if (!step.selector) {
      throw new Error('验证步骤缺少选择器');
    }

    // 使用快照验证元素状态
    const snapshot = await (validator as any).takeSnapshot();
    const element = snapshot.elements.find((el: any) => el.uid === step.selector);

    if (!element) {
      throw new Error(`未找到元素: ${step.selector}`);
    }

    // 验证期望状态
    if (step.expectedState) {
      if (step.expectedState.visible !== undefined && element.visible !== step.expectedState.visible) {
        throw new Error(`元素可见性不匹配: 期望 ${step.expectedState.visible}, 实际 ${element.visible}`);
      }

      if (step.expectedState.enabled !== undefined && element.enabled !== step.expectedState.enabled) {
        throw new Error(`元素启用状态不匹配: 期望 ${step.expectedState.enabled}, 实际 ${element.enabled}`);
      }

      if (step.expectedState.text !== undefined && element.textContent !== step.expectedState.text) {
        throw new Error(`元素文本不匹配: 期望 "${step.expectedState.text}", 实际 "${element.textContent}"`);
      }
    }
  }

  /**
   * 辅助方法：延迟
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * 清理资源
   */
  async (): Promise<void> {
    try {
      // 清理验证器资源
      if (this.getcontext?.().validator) {
        await (this.context.validator as any).close();
      }

      // 清理数据提供者资源
      if (this.dataProvider) {
        await this.dataProvider.cleanupTestData();
      }

      this.logger.info('资源清理完成');
    } catch (error) {
      this.logger.error('资源清理失败', {
        error: error instanceof Error ? error.message : error
      });
    }
  }
}