/**
 * 测试监听器实现
 */

import {
  TestListener,
  TestExecutionContext,
  TestExecutionResult,
  TestSuiteResult,
  TestCaseResult,
  TestStepResult
} from '../types/test-engine-types';
import { Logger } from '../core/logger';

/**
 * 控制台日志监听器
 */
export class ConsoleLoggingListener implements TestListener {
  private logger: Logger;
  private startTime: Date = new Date();
  private suiteStartTime: Map<string, Date> = new Map();
  private caseStartTime: Map<string, Date> = new Map();
  private stepStartTime: Map<string, Date> = new Map();

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('ConsoleLoggingListener');
  }

  override async onExecutionStart(context: TestExecutionContext): Promise<void> {
    this.startTime = context.startTime;
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.logger.info('🚀 开始执行测试', {
      executionId: context.executionId,
      timestamp: context.startTime.toISOString(),
      environment: context.config.testEnvironment
    });
  }

  override async onExecutionEnd(result: TestExecutionResult): Promise<void> {
    const duration = (result.getendTime?.().getTime() || Date.now()) - this.startTime.getTime();
    const successRate = result.total > 0 ? Math.round((result.passed / result.total) * 100) : 0;

    this.logger.info('✅ 测试执行完成', {
      executionId: result.executionId,
      duration: `${Math.round(duration / 1000)}s`,
      total: result.total,
      passed: result.passed,
      failed: result.failed,
      skipped: result.skipped,
      successRate: `${successRate}%`
    });

    if (result.failed > 0) {
      this.logger.warn('⚠️ 存在失败的测试用例', { failedCount: result.failed });
    }

    if (result.errors.length > 0) {
      this.logger.error('❌ 执行过程中发生错误', { errorCount: result.errors.length });
    }
  }

  override async onSuiteStart(suite: any, __context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.suiteStartTime.set(suite.id, new Date());
    this.logger.info('📁 开始执行测试套件', {
      suiteId: suite.id,
      suiteName: suite.name,
      testCaseCount: suite.gettestCases?.().length || 0
    });
  }

  override async onSuiteEnd(result: TestSuiteResult): Promise<void> {
    const startTime = this.suiteStartTime.get(result.suiteId);
    const duration = startTime ? (result.getendTime?.().getTime() || Date.now()) - startTime.getTime() : 0;

    this.logger.info('📁 测试套件执行完成', {
      suiteId: result.suiteId,
      suiteName: result.name,
      status: result.status,
      duration: `${Math.round(duration / 1000)}s`,
      passed: result.testCases.filter(tc => tc.status === 'passed').length,
      failed: result.testCases.filter(tc => tc.status === 'failed').length
    });

    this.suiteStartTime.delete(result.suiteId);
  }

  override async onTestCaseStart(testCase: any, _context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.caseStartTime.set(testCase.id, new Date());
    this.logger.debug('🔸 开始执行测试用例', {
      caseId: testCase.id,
      caseName: testCase.name,
      stepCount: testCase.getsteps?.().length || 0
    });
  }

  override async onTestCaseEnd(result: TestCaseResult): Promise<void> {
    const startTime = this.caseStartTime.get(result.caseId);
    const duration = startTime ? (result.getendTime?.().getTime() || Date.now()) - startTime.getTime() : 0;

    const statusIcon = {
      'passed': '✅',
      'failed': '❌',
      'skipped': '⏭️',
      'pending': '⏳'
    }[result.status] || '❓';

    this.logger.debug(`${statusIcon} 测试用例执行完成`, {
      caseId: result.caseId,
      caseName: result.name,
      status: result.status,
      duration: `${duration}ms`,
      retryCount: result.retryCount || 0
    });

    this.caseStartTime.delete(result.caseId);
  }

  override async onTestStepStart(step: any, _context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.stepStartTime.set(step.id, new Date());
    this.logger.trace('▶️ 开始执行测试步骤', {
      stepId: step.id,
      stepName: step.name,
      action: step.action
    });
  }

  override async onTestStepEnd(result: TestStepResult): Promise<void> {
    const startTime = this.stepStartTime.get(result.stepId);
    const duration = startTime ? (result.getendTime?.().getTime() || Date.now()) - startTime.getTime() : 0;

    const statusIcon = {
      'passed': '✅',
      'failed': '❌',
      'skipped': '⏭️',
      'pending': '⏳'
    }[result.status] || '❓';

    this.logger.trace(`${statusIcon} 测试步骤执行完成`, {
      stepId: result.stepId,
      stepName: result.name,
      status: result.status,
      duration: `${duration}ms`
    });

    if (result.status === 'failed' && result.error) {
      this.logger.warn('⚠️ 测试步骤失败', {
        stepId: result.stepId,
        stepName: result.name,
        error: result.error.message
      });
    }

    this.stepStartTime.delete(result.stepId);
  }
}

/**
 * 文件日志监听器
 */
export class FileLoggingListener implements TestListener {
  private logger: Logger;
  private logFile: string;
  private logs: string[] | undefined = undefined;

  constructor(logFile: string, logger?: Logger) {
    this.logFile = logFile;
    this.logger = logger || new Logger('FileLoggingListener');
  }

  override async onExecutionStart(_context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const logEntry = `[${_context.startTime.toISOString()}] START execution=${_context.executionId} environment=${_context.config.testEnvironment}`;
    this.addLog(logEntry);
  }

  override async onExecutionEnd(result: TestExecutionResult): Promise<void> {
    const duration = (result.getendTime?.().getTime() || Date.now()) - (result.getstartTime?.().getTime() || Date.now());
    const successRate = result.total > 0 ? Math.round((result.passed / result.total) * 100) : 0;

    const logEntry = `[${result.getendTime?.().toISOString() || new Date().toISOString()}] END execution=${result.executionId} duration=${duration}ms total=${result.total} passed=${result.passed} failed=${result.failed} successRate=${successRate}%`;
    this.addLog(logEntry);

    // 写入日志文件
    await this.writeLogs();
  }

  override async onSuiteStart(suite: any, _context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const logEntry = `[${new Date().toISOString()}] SUITE_START suite=${suite.id} name="${suite.name}" cases=${suite.gettestCases?.().length || 0}`;
    this.addLog(logEntry);
  }

  override async onSuiteEnd(result: TestSuiteResult): Promise<void> {
    const logEntry = `[${new Date().toISOString()}] SUITE_END suite=${result.suiteId} name="${result.name}" status=${result.status} passed=${result.testCases.filter(tc => tc.status === 'passed').length} failed=${result.testCases.filter(tc => tc.status === 'failed').length}`;
    this.addLog(logEntry);
  }

  override async onTestCaseStart(testCase: any, _context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const logEntry = `[${new Date().toISOString()}] TEST_START case=${testCase.id} name="${testCase.name}" steps=${testCase.getsteps?.().length || 0}`;
    this.addLog(logEntry);
  }

  override async onTestCaseEnd(result: TestCaseResult): Promise<void> {
    const duration = (result.getendTime?.().getTime() || Date.now()) - (result.getstartTime?.().getTime() || Date.now());
    const logEntry = `[${new Date().toISOString()}] TEST_END case=${result.caseId} name="${result.name}" status=${result.status} duration=${duration}ms retries=${result.retryCount || 0}`;
    this.addLog(logEntry);
  }

  override async onTestStepStart(step: any, _context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const logEntry = `[${new Date().toISOString()}] STEP_START step=${step.id} name="${step.name}" action=${step.action}`;
    this.addLog(logEntry);
  }

  override async onTestStepEnd(result: TestStepResult): Promise<void> {
    const duration = (result.getendTime?.().getTime() || Date.now()) - (result.getstartTime?.().getTime() || Date.now());
    const logEntry = `[${new Date().toISOString()}] STEP_END step=${result.stepId} name="${result.name}" status=${result.status} duration=${duration}ms`;
    this.addLog(logEntry);

    if (result.status === 'failed' && result.error) {
      const errorEntry = `[${new Date().toISOString()}] STEP_ERROR step=${result.stepId} error="${result.error.message}"`;
      this.addLog(errorEntry);
    }
  }

  private addLog(entry: string): void {
    this.logs.push(entry);
  }

  private override async writeLogs(): Promise<void> {
    try {
      // 在实际实现中，这里会写入文件系统
      this.logger.debug('写入测试日志文件', {
        file: this.logFile,
        entryCount: this.logs.length
      });

      // 清空日志
      this.logs = [];
    } catch (error) {
      this.logger.error('写入测试日志文件失败', {
        file: this.logFile,
        error: error instanceof Error ? error.message : error
      });
    }
  }
}

/**
 * 性能监控监听器
 */
export class PerformanceMonitoringListener implements TestListener {
  private logger: Logger;
  private metrics: {
    executionStartTimes: Map<string, number>;
    suiteMetrics: Map<string, { startTime: number; caseCount: number }>;
    caseMetrics: Map<string, { startTime: number; stepCount: number }>;
    stepMetrics: Map<string, { startTime: number; action: string }>;
  };

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('PerformanceMonitoringListener');
    this.metrics = {
      executionStartTimes: new Map(),
      suiteMetrics: new Map(),
      caseMetrics: new Map(),
      stepMetrics: new Map()
    };
  }

  override async onExecutionStart(_context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.metrics.executionStartTimes.set(_context.executionId, _context.startTime.getTime());
    this.logger.debug('性能监控开始', { executionId: _context.executionId });
  }

  override async onExecutionEnd(result: TestExecutionResult): Promise<void> {
    const startTime = this.metrics.executionStartTimes.get(result.executionId);
    if (startTime) {
      const duration = result.endTime.getTime() - startTime;
      const avgTestCaseDuration = result.total > 0 ? duration / result.total : 0;

      this.logger.info('性能统计', {
        executionId: result.executionId,
        totalDuration: `${duration}ms`,
        avgTestCaseDuration: `${Math.round(avgTestCaseDuration)}ms`,
        testCasesPerSecond: result.total > 0 ? Math.round((result.total / duration) * 1000) : 0
      });
    }

    this.metrics.executionStartTimes.delete(result.executionId);
  }

  override async onSuiteStart(suite: any, _context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.metrics.suiteMetrics.set(suite.id, {
      startTime: Date.now(),
      caseCount: suite.gettestCases?.().length || 0
    });
  }

  override async onSuiteEnd(result: TestSuiteResult): Promise<void> {
    const suiteMetrics = this.metrics.suiteMetrics.get(result.suiteId);
    if (suiteMetrics) {
      const duration = result.endTime.getTime() - suiteMetrics.startTime;
      const avgCaseDuration = suiteMetrics.caseCount > 0 ? duration / suiteMetrics.caseCount : 0;

      this.logger.debug('套件性能统计', {
        suiteId: result.suiteId,
        suiteName: result.name,
        totalDuration: `${duration}ms`,
        avgCaseDuration: `${Math.round(avgCaseDuration)}ms`,
        caseCount: suiteMetrics.caseCount
      });
    }

    this.metrics.suiteMetrics.delete(result.suiteId);
  }

  override async onTestCaseStart(testCase: any, _context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.metrics.caseMetrics.set(testCase.id, {
      startTime: Date.now(),
      stepCount: testCase.getsteps?.().length || 0
    });
  }

  override async onTestCaseEnd(result: TestCaseResult): Promise<void> {
    const caseMetrics = this.metrics.caseMetrics.get(result.caseId);
    if (caseMetrics) {
      const duration = (result.getendTime?.().getTime() || Date.now()) - caseMetrics.startTime;
      const avgStepDuration = caseMetrics.stepCount > 0 ? duration / caseMetrics.stepCount : 0;

      if (duration > 5000) { // 超过5秒的测试用例
        this.logger.warn('慢测试用例检测', {
          caseId: result.caseId,
          caseName: result.name,
          duration: `${duration}ms`,
          stepCount: caseMetrics.stepCount,
          avgStepDuration: `${Math.round(avgStepDuration)}ms`
        });
      }
    }

    this.metrics.caseMetrics.delete(result.caseId);
  }

  override async onTestStepStart(step: any, _context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.metrics.stepMetrics.set(step.id, {
      startTime: Date.now(),
      action: step.action
    });
  }

  override async onTestStepEnd(result: TestStepResult): Promise<void> {
    const stepMetrics = this.metrics.stepMetrics.get(result.stepId);
    if (stepMetrics) {
      const duration = (result.getendTime?.().getTime() || Date.now()) - stepMetrics.startTime;

      if (duration > 1000) { // 超过1秒的测试步骤
        this.logger.debug('慢测试步骤检测', {
          stepId: result.stepId,
          stepName: result.name,
          action: stepMetrics.action,
          duration: `${duration}ms`
        });
      }
    }

    this.metrics.stepMetrics.delete(result.stepId);
  }
}

/**
 * 实时进度监听器
 */
export class RealTimeProgressListener implements TestListener {
  private logger: Logger;
  private progress: {
    total: number;
    completed: number;
    passed: number;
    failed: number;
    startTime: Date;
  };

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('RealTimeProgressListener');
    this.progress = {
      total: 0,
      completed: 0,
      passed: 0,
      failed: 0,
      startTime: new Date()
    };
  }

  override async onExecutionStart(_context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.progress = {
      total: 0,
      completed: 0,
      passed: 0,
      failed: 0,
      startTime: _context.startTime
    };
    this.logger.info('实时进度监控开始', { executionId: _context.executionId });
  }

  override async onExecutionEnd(result: TestExecutionResult): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.logger.info('实时进度监控结束', {
      executionId: result.executionId,
      total: this.progress.total,
      completed: this.progress.completed,
      passed: this.progress.passed,
      failed: this.progress.failed
    });
  }

  override async onSuiteStart(suite: any, _context: TestExecutionContext): Promise<void> {
    this.progress.total += suite.gettestCases?.().length || 0;
    this.updateProgress();
  }

  override async onTestCaseEnd(result: TestCaseResult): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.progress.completed++;
    if (result.status === 'passed') {
      this.progress.passed++;
    } else if (result.status === 'failed') {
      this.progress.failed++;
    }

    this.updateProgress();
  }

  override async onSuiteEnd(_result: TestSuiteResult): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.updateProgress();
  }

  override async onTestCaseStart(_testCase: any, _context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    // 测试用例开始时不需要更新进度
  }

  override async onTestStepStart(_step: any, _context: TestExecutionContext): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    // 测试步骤开始时不需要更新进度
  }

  override async onTestStepEnd(_result: TestStepResult): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    // 测试步骤结束时不需要更新进度
  }

  private updateProgress(): void {
    const progress = this.progress.total > 0 ? (this.progress.completed / this.progress.total) * 100 : 0;
    const elapsed = Date.now() - this.progress.startTime.getTime();
    const eta = this.progress.completed > 0 ? (elapsed / this.progress.completed) * (this.progress.total - this.progress.completed) : 0;

    this.logger.debug('实时进度', {
      progress: `${Math.round(progress)}%`,
      completed: this.progress.completed,
      total: this.progress.total,
      passed: this.progress.passed,
      failed: this.progress.failed,
      eta: eta > 0 ? `${Math.round(eta / 1000)}s` : 'N/A'
    });
  }
}