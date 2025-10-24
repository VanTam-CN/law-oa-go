/**
 * 测试执行引擎类型定义
 */

import { ChromeDevToolsValidator } from '../index';
import { TestResult, TestSuite, TestCase, TestStep } from './test-types';

/**
 * 测试执行配置
 */
export interface TestExecutionConfig {
  // 基础配置
  headless?: boolean;
  slowMo?: number;
  defaultTimeout?: number;
  viewport?: { width: number; height: number };

  // 执行配置
  parallelExecution?: boolean;
  maxConcurrency?: number;
  retryFailedTests?: boolean;
  maxRetries?: number;

  // 报告配置
  generateReport?: boolean;
  reportFormat?: 'json' | 'html' | 'junit';
  screenshotsOnFailure?: boolean;

  // 环境配置
  baseUrl?: string;
  testEnvironment?: 'development' | 'staging' | 'production';
}

/**
 * 测试执行上下文
 */
export interface TestExecutionContext {
  // 执行信息
  executionId: string;
  startTime: Date;
  config: TestExecutionConfig;

  // 验证器实例
  validator: ChromeDevToolsValidator;

  // 当前测试状态
  currentSuite?: TestSuite;
  currentCase?: TestCase;
  currentStep?: TestStep;

  // 共享数据
  sharedData: Map<string, any>;

  // 钩子函数
  hooks: TestHooks;
}

/**
 * 测试钩子函数
 */
export interface TestHooks {
  // 全局钩子
  beforeAll?: (context: TestExecutionContext) => Promise<void>;
  afterAll?: (context: TestExecutionContext) => Promise<void>;

  // 测试套件钩子
  beforeSuite?: (suite: TestSuite, context: TestExecutionContext) => Promise<void>;
  afterSuite?: (suite: TestSuite, context: TestExecutionContext) => Promise<void>;

  // 测试用例钩子
  beforeTest?: (testCase: TestCase, context: TestExecutionContext) => Promise<void>;
  afterTest?: (testCase: TestCase, context: TestExecutionContext) => Promise<void>;

  // 测试步骤钩子
  beforeStep?: (step: TestStep, context: TestExecutionContext) => Promise<void>;
  afterStep?: (step: TestStep, context: TestExecutionContext) => Promise<void>;
}

/**
 * 测试执行器接口
 */
export interface TestExecutor {
  /**
   * 执行测试套件
   */
  executeSuite(suite: TestSuite, context: TestExecutionContext): Promise<TestResult>;

  /**
   * 执行测试用例
   */
  executeTestCase(testCase: TestCase, context: TestExecutionContext): Promise<TestResult>;

  /**
   * 执行测试步骤
   */
  executeTestStep(step: TestStep, context: TestExecutionContext): Promise<TestStepResult>;
}

/**
 * 测试步骤结果
 */
export interface TestStepResult {
  stepId: string;
  name: string;
  status: 'passed' | 'failed' | 'skipped' | 'pending';
  duration: number;
  startTime?: Date;
  endTime?: Date;
  error?: {
    message: string;
    stack?: string;
    screenshot?: string;
  };
  screenshot?: string;
  logs: TestLog[];
  metadata?: Record<string, any>;
}

/**
 * 测试日志条目
 */
export interface TestLog {
  timestamp: Date;
  level: 'debug' | 'info' | 'warn' | 'error';
  message: string;
  context?: Record<string, any>;
  stepId?: string;
}

/**
 * 测试报告生成器接口
 */
export interface TestReportGenerator {
  /**
   * 生成测试报告
   */
  generateReport(results: TestExecutionResult): Promise<string>;

  /**
   * 保存报告到文件
   */
  saveReport(report: string, filePath: string): Promise<void>;
}

/**
 * 测试执行结果
 */
export interface TestExecutionResult {
  // 执行概要
  executionId: string;
  startTime: Date;
  endTime: Date;
  duration: number;
  config: TestExecutionConfig;

  // 测试结果统计
  total: number;
  passed: number;
  failed: number;
  skipped: number;
  pending: number;

  // 详细结果
  suites: TestSuiteResult[];
  errors: TestExecutionError[];

  // 元数据
  metadata: {
    environment: string;
    browser: string;
    version: string;
    executor: string;
  };
}

/**
 * 测试套件结果
 */
export interface TestSuiteResult {
  suiteId: string;
  name: string;
  description?: string;
  status: 'passed' | 'failed' | 'skipped' | 'pending';
  startTime: Date;
  endTime: Date;
  duration: number;
  testCases: TestCaseResult[];
  setupError?: string;
  teardownError?: string;
}

/**
 * 测试用例结果
 */
export interface TestCaseResult {
  caseId: string;
  name: string;
  description?: string;
  status: 'passed' | 'failed' | 'skipped' | 'pending';
  startTime: Date;
  endTime: Date;
  duration: number;
  steps: TestStepResult[];
  setupError?: string;
  teardownError?: string;
  retryCount?: number;
  error?: {
    message: string;
    stack?: string;
    screenshot?: string;
  };
}

/**
 * 测试执行错误
 */
export interface TestExecutionError {
  timestamp: Date;
  type: 'setup' | 'teardown' | 'execution' | 'system';
  message: string;
  stack?: string;
  component: string;
  context?: Record<string, any>;
}

/**
 * 测试数据提供者接口
 */
export interface TestDataProvider {
  /**
   * 获取测试数据
   */
  getTestData(dataKey: string): Promise<any>;

  /**
   * 设置测试数据
   */
  setTestData(dataKey: string, data: any): Promise<void>;

  /**
   * 清理测试数据
   */
  cleanupTestData(): Promise<void>;
}

/**
 * 测试监听器接口
 */
export interface TestListener {
  /**
   * 测试执行开始
   */
  onExecutionStart(context: TestExecutionContext): Promise<void>;

  /**
   * 测试执行结束
   */
  onExecutionEnd(result: TestExecutionResult): Promise<void>;

  /**
   * 测试套件开始
   */
  onSuiteStart(suite: TestSuite, context: TestExecutionContext): Promise<void>;

  /**
   * 测试套件结束
   */
  onSuiteEnd(result: TestSuiteResult): Promise<void>;

  /**
   * 测试用例开始
   */
  onTestCaseStart(testCase: TestCase, context: TestExecutionContext): Promise<void>;

  /**
   * 测试用例结束
   */
  onTestCaseEnd(result: TestCaseResult): Promise<void>;

  /**
   * 测试步骤开始
   */
  onTestStepStart(step: TestStep, context: TestExecutionContext): Promise<void>;

  /**
   * 测试步骤结束
   */
  onTestStepEnd(result: TestStepResult): Promise<void>;
}