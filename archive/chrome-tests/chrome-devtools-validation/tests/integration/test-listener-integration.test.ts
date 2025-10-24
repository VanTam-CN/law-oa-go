import {
  ConsoleLoggingListener,
  FileLoggingListener,
  PerformanceMonitoringListener,
  RealTimeProgressListener
} from '../../src/core/test-listeners';
import { ChromeDevToolsTestExecutionEngine } from '../../src/core/test-execution-engine';
import { MemoryDataProvider } from '../../src/core/test-data-provider';
import { JsonReportGenerator } from '../../src/core/test-report-generator';
import { ChromeDevToolsValidator } from '../../src/index';
import { TestSuite, TestCase, TestStep } from '../../src/types/test-types';
import { TestExecutionResult, TestExecutionContext } from '../../src/types/test-engine-types';
import { writeFileSync, readFileSync, existsSync, unlinkSync, mkdirSync } from 'fs';
import { join } from 'path';
import { jest } from '@jest/globals';

// Mock ChromeDevToolsValidator
jest.mock('../../src/index');
const MockChromeDevToolsValidator = require('../../src/index').ChromeDevToolsValidator;

// Mock console methods
const mockConsoleLog = jest.fn();
const mockConsoleError = jest.fn();
const mockConsoleWarn = jest.fn();
const mockConsoleInfo = jest.fn();

describe('Listener Integration Tests', () => {
  let testOutputDir: string;
  let mockValidator: any;
  let engine: TestExecutionEngine;
  let dataProvider: MemoryDataProvider;
  let reportGenerator: JsonReportGenerator;

  beforeEach(() => {
    testOutputDir = join(__dirname, '../test-logs');

    // Create test output directory
    if (!existsSync(testOutputDir)) {
      mkdirSync(testOutputDir, { recursive: true });
    }

    // Mock console
    global.console = {
      log: mockConsoleLog,
      error: mockConsoleError,
      warn: mockConsoleWarn,
      info: mockConsoleInfo,
      debug: jest.fn(),
      trace: jest.fn()
    } as any;

    // Mock validator
    mockValidator = {
      navigate: jest.fn(),
      click: jest.fn(),
      fill: jest.fn(),
      wait: jest.fn(),
      screenshot: jest.fn(),
      executeScript: jest.fn(),
      close: jest.fn()
    };

    MockChromeDevToolsValidator.mockImplementation(() => mockValidator);

    // Setup real components
    dataProvider = new MemoryDataProvider();
    reportGenerator = new JsonReportGenerator();
    engine = new TestExecutionEngine();
    engine.setDataProvider(dataProvider);
    engine.setReportGenerator(reportGenerator);

    // Setup validator responses
    mockValidator.navigate.mockResolvedValue(undefined);
    mockValidator.click.mockResolvedValue(undefined);
    mockValidator.fill.mockResolvedValue(undefined);
    mockValidator.wait.mockResolvedValue(undefined);
    mockValidator.executeScript.mockResolvedValue(true);
    mockValidator.screenshot.mockResolvedValue('screenshot.png');
  });

  afterEach(() => {
    jest.clearAllMocks();

    // Clean up test files
    const cleanup = (dir: string) => {
      if (existsSync(dir)) {
        const files = require('fs').readdirSync(dir);
        files.forEach((file: string) => {
          const filePath = join(dir, file);
          if (require('fs').statSync(filePath).isDirectory()) {
            cleanup(filePath);
            require('fs').rmdirSync(filePath);
          } else {
            unlinkSync(filePath);
          }
        });
      }
    };
    cleanup(testOutputDir);
  });

  describe('ConsoleLoggingListener Integration', () => {
    let listener: ConsoleLoggingListener;

    beforeEach(() => {
      listener = new ConsoleLoggingListener();
      engine.addListener(listener);
    });

    it('should log execution start and end', async () => {
      const testSuite: TestSuite = {
        id: 'console-log-test',
        name: 'Console Log Test',
        testCases: [
          {
            id: 'simple-test',
            name: 'Simple Test',
            steps: [
              {
                id: 'simple-step',
                name: 'Simple Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      // Verify execution start was logged
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('测试执行开始')
      );

      // Verify execution end was logged
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('测试执行完成')
      );

      // Verify summary was logged
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('总测试数: 1')
      );
    });

    it('should log test suite events', async () => {
      const testSuite: TestSuite = {
        id: 'suite-events-test',
        name: 'Suite Events Test',
        testCases: [
          {
            id: 'suite-test-case',
            name: 'Suite Test Case',
            steps: [
              {
                id: 'suite-step',
                name: 'Suite Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      // Verify suite start was logged
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('测试套件开始: Suite Events Test')
      );

      // Verify suite end was logged
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('测试套件结束: Suite Events Test')
      );
    });

    it('should log test case events', async () => {
      const testSuite: TestSuite = {
        id: 'case-events-test',
        name: 'Case Events Test',
        testCases: [
          {
            id: 'case-events-test-case',
            name: 'Case Events Test Case',
            steps: [
              {
                id: 'case-step',
                name: 'Case Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      // Verify test case start was logged
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('测试用例开始: Case Events Test Case')
      );

      // Verify test case end was logged
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('测试用例结束: Case Events Test Case')
      );
    });

    it('should log test step events', async () => {
      const testSuite: TestSuite = {
        id: 'step-events-test',
        name: 'Step Events Test',
        testCases: [
          {
            id: 'step-events-test-case',
            name: 'Step Events Test Case',
            steps: [
              {
                id: 'step-events-step',
                name: 'Step Events Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      // Verify test step start was logged
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('测试步骤开始: Step Events Step')
      );

      // Verify test step end was logged
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('测试步骤结束: Step Events Step')
      );
    });

    it('should log errors and failures', async () => {
      // Setup validator to fail
      mockValidator.navigate.mockRejectedValue(new Error('Navigation failed'));

      const testSuite: TestSuite = {
        id: 'error-test',
        name: 'Error Test',
        testCases: [
          {
            id: 'error-test-case',
            name: 'Error Test Case',
            steps: [
              {
                id: 'error-step',
                name: 'Error Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      // Verify error was logged
      expect(mockConsoleError).toHaveBeenCalledWith(
        expect.stringContaining('测试步骤失败: Error Step')
      );
      expect(mockConsoleError).toHaveBeenCalledWith(
        expect.stringContaining('Navigation failed')
      );
    });
  });

  describe('FileLoggingListener Integration', () => {
    let listener: FileLoggingListener;
    let logFilePath: string;

    beforeEach(() => {
      logFilePath = join(testOutputDir, 'test-execution.log');
      listener = new FileLoggingListener(testOutputDir);
      engine.addListener(listener);
    });

    it('should write logs to file', async () => {
      const testSuite: TestSuite = {
        id: 'file-log-test',
        name: 'File Log Test',
        testCases: [
          {
            id: 'file-test-case',
            name: 'File Test Case',
            steps: [
              {
                id: 'file-step',
                name: 'File Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      // Verify log file was created
      expect(existsSync(logFilePath)).toBe(true);

      // Verify log content
      const logContent = readFileSync(logFilePath, 'utf-8');
      expect(logContent).toContain('测试执行开始');
      expect(logContent).toContain('测试执行完成');
      expect(logContent).toContain('File Log Test');
      expect(logContent).toContain('File Test Case');
      expect(logContent).toContain('File Step');
    });

    it('should handle log rotation', async () => {
      const largeTestSuite: TestSuite = {
        id: 'large-log-test',
        name: 'Large Log Test',
        testCases: Array.from({ length: 100 }, (_, i) => ({
          id: `test-case-${i}`,
          name: `Test Case ${i}`,
          steps: [
            {
              id: `step-${i}`,
              name: `Step ${i}`,
              type: 'navigate',
              url: 'https://example.com', url: `https://example.com/page${i}` }
            }
          ]
        }))
      };

      await engine.executeSuite(largeTestSuite);

      // Verify log file was created and contains content
      expect(existsSync(logFilePath)).toBe(true);
      const logContent = readFileSync(logFilePath, 'utf-8');
      expect(logContent.length).toBeGreaterThan(0);
      expect(logContent).toContain('Large Log Test');
    });

    it('should handle concurrent logging', async () => {
      const concurrentTestSuite: TestSuite = {
        id: 'concurrent-log-test',
        name: 'Concurrent Log Test',
        testCases: Array.from({ length: 10 }, (_, i) => ({
          id: `concurrent-test-${i}`,
          name: `Concurrent Test ${i}`,
          steps: Array.from({ length: 5 }, (_, j) => ({
            id: `concurrent-step-${i}-${j}`,
            name: `Concurrent Step ${i}-${j}`,
            type: 'navigate',
            url: `https://example.com/page${i}/${j}`
          }))
        }))
      };

      await engine.executeSuite(concurrentTestSuite);

      // Verify log file contains all concurrent operations
      expect(existsSync(logFilePath)).toBe(true);
      const logContent = readFileSync(logFilePath, 'utf-8');
      expect(logContent).toContain('Concurrent Log Test');

      // Verify multiple test cases were logged
      for (let i = 0; i < 10; i++) {
        expect(logContent).toContain(`Concurrent Test ${i}`);
      }
    });
  });

  describe('PerformanceMonitoringListener Integration', () => {
    let listener: PerformanceMonitoringListener;

    beforeEach(() => {
      listener = new PerformanceMonitoringListener();
      engine.addListener(listener);
    });

    it('should track execution performance metrics', async () => {
      const testSuite: TestSuite = {
        id: 'perf-test',
        name: 'Performance Test',
        testCases: [
          {
            id: 'perf-test-case',
            name: 'Performance Test Case',
            steps: [
              {
                id: 'perf-step',
                name: 'Performance Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      // Add delay to simulate processing time
      const originalNavigate = mockValidator.navigate;
      mockValidator.navigate.mockImplementation(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
        return originalNavigate();
      });

      const startTime = Date.now();
      await engine.executeSuite(testSuite);
      const endTime = Date.now();

      // Verify performance metrics were collected
      const metrics = listener.getPerformanceMetrics();
      expect(metrics).toBeDefined();
      expect(metrics.totalExecutionTime).toBeGreaterThan(0);
      expect(metrics.totalExecutionTime).toBeLessThan(endTime - startTime + 100);
      expect(metrics.averageTestTime).toBeGreaterThan(0);
      expect(metrics.slowestTest).toBeDefined();
      expect(metrics.fastestTest).toBeDefined();
    });

    it('should track memory usage', async () => {
      const testSuite: TestSuite = {
        id: 'memory-test',
        name: 'Memory Test',
        testCases: [
          {
            id: 'memory-test-case',
            name: 'Memory Test Case',
            steps: [
              {
                id: 'memory-step',
                name: 'Memory Step',
                action: 'executeScript',
                url: 'https://example.com',
                  script: 'return Array(1000).fill(0).map(() => Math.random());'
                }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      const metrics = listener.getPerformanceMetrics();
      expect(metrics).toBeDefined();
      expect(metrics.memoryUsage).toBeDefined();
      expect(metrics.memoryUsage.heapUsed).toBeGreaterThan(0);
      expect(metrics.memoryUsage.heapTotal).toBeGreaterThan(0);
      expect(metrics.memoryUsage.external).toBeGreaterThan(0);
    });

    it('should generate performance report', async () => {
      const testSuite: TestSuite = {
        id: 'perf-report-test',
        name: 'Performance Report Test',
        testCases: [
          {
            id: 'perf-report-case',
            name: 'Performance Report Case',
            steps: [
              {
                id: 'perf-report-step',
                name: 'Performance Report Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      const report = listener.generatePerformanceReport();
      expect(report).toBeDefined();
      expect(report).toContain('Performance Report');
      expect(report).toContain('Total Execution Time');
      expect(report).toContain('Average Test Time');
      expect(report).toContain('Memory Usage');
    });

    it('should handle multiple test cases performance tracking', async () => {
      const testSuite: TestSuite = {
        id: 'multi-perf-test',
        name: 'Multi Performance Test',
        testCases: Array.from({ length: 5 }, (_, i) => ({
          id: `perf-test-case-${i}`,
          name: `Performance Test Case ${i}`,
          steps: [
            {
              id: `perf-step-${i}`,
              name: `Performance Step ${i}`,
              type: 'navigate',
              url: 'https://example.com', url: `https://example.com/page${i}` }
            }
          ]
        }))
      };

      await engine.executeSuite(testSuite);

      const metrics = listener.getPerformanceMetrics();
      expect(metrics).toBeDefined();
      expect(metrics.totalTests).toBe(5);
      expect(metrics.testTimes).toHaveLength(5);
      expect(metrics.averageTestTime).toBeGreaterThan(0);
    });
  });

  describe('RealTimeProgressListener Integration', () => {
    let listener: RealTimeProgressListener;

    beforeEach(() => {
      listener = new RealTimeProgressListener();
      engine.addListener(listener);
    });

    it('should track real-time progress', async () => {
      const testSuite: TestSuite = {
        id: 'progress-test',
        name: 'Progress Test',
        testCases: [
          {
            id: 'progress-test-case',
            name: 'Progress Test Case',
            steps: [
              {
                id: 'progress-step-1',
                name: 'Progress Step 1',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              },
              {
                id: 'progress-step-2',
                name: 'Progress Step 2',
                type: 'click',
                url: 'https://example.com', selector: '#button' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      const progress = listener.getCurrentProgress();
      expect(progress).toBeDefined();
      expect(progress.total).toBe(2); // 2 steps
      expect(progress.completed).toBe(2);
      expect(progress.failed).toBe(0);
      expect(progress.passed).toBe(2);
      expect(progress.progressPercentage).toBe(100);
    });

    it('should update progress during execution', async () => {
      const testSuite: TestSuite = {
        id: 'dynamic-progress-test',
        name: 'Dynamic Progress Test',
        testCases: [
          {
            id: 'dynamic-test-case',
            name: 'Dynamic Test Case',
            steps: [
              {
                id: 'dynamic-step-1',
                name: 'Dynamic Step 1',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              },
              {
                id: 'dynamic-step-2',
                name: 'Dynamic Step 2',
                type: 'click',
                url: 'https://example.com', selector: '#button' }
              }
            ]
          }
        ]
      };

      // Check initial progress
      const initialProgress = listener.getCurrentProgress();
      expect(initialProgress.total).toBe(0);
      expect(initialProgress.completed).toBe(0);

      await engine.executeSuite(testSuite);

      // Check final progress
      const finalProgress = listener.getCurrentProgress();
      expect(finalProgress.total).toBe(2);
      expect(finalProgress.completed).toBe(2);
      expect(finalProgress.progressPercentage).toBe(100);
    });

    it('should handle failed tests in progress tracking', async () => {
      // Setup validator to fail
      mockValidator.navigate.mockRejectedValue(new Error('Test failure'));

      const testSuite: TestSuite = {
        id: 'fail-progress-test',
        name: 'Fail Progress Test',
        testCases: [
          {
            id: 'fail-test-case',
            name: 'Fail Test Case',
            steps: [
              {
                id: 'fail-step',
                name: 'Fail Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      const progress = listener.getCurrentProgress();
      expect(progress).toBeDefined();
      expect(progress.total).toBe(1);
      expect(progress.completed).toBe(1);
      expect(progress.failed).toBe(1);
      expect(progress.passed).toBe(0);
      expect(progress.progressPercentage).toBe(100);
    });

    it('should generate progress report', async () => {
      const testSuite: TestSuite = {
        id: 'progress-report-test',
        name: 'Progress Report Test',
        testCases: [
          {
            id: 'progress-report-case',
            name: 'Progress Report Case',
            steps: [
              {
                id: 'progress-report-step',
                name: 'Progress Report Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      const report = listener.generateProgressReport();
      expect(report).toBeDefined();
      expect(report).toContain('Progress Report');
      expect(report).toContain('Total Steps');
      expect(report).toContain('Completed');
      expect(report).toContain('Failed');
      expect(report).toContain('Success Rate');
    });
  });

  describe('Multi-Listener Integration', () => {
    let listeners: any[];

    beforeEach(() => {
      listeners = [
        new ConsoleLoggingListener(),
        new FileLoggingListener(testOutputDir),
        new PerformanceMonitoringListener(),
        new RealTimeProgressListener()
      ];

      listeners.forEach(listener => engine.addListener(listener));
    });

    it('should coordinate multiple listeners', async () => {
      const testSuite: TestSuite = {
        id: 'multi-listener-test',
        name: 'Multi Listener Test',
        testCases: [
          {
            id: 'multi-test-case',
            name: 'Multi Test Case',
            steps: [
              {
                id: 'multi-step',
                name: 'Multi Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      await engine.executeSuite(testSuite);

      // Verify all listeners received events
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('测试执行开始')
      );

      // Check file logging
      const logFilePath = join(testOutputDir, 'test-execution.log');
      expect(existsSync(logFilePath)).toBe(true);

      // Check performance metrics
      const perfListener = listeners[2] as PerformanceMonitoringListener;
      const metrics = perfListener.getPerformanceMetrics();
      expect(metrics).toBeDefined();

      // Check progress tracking
      const progressListener = listeners[3] as RealTimeProgressListener;
      const progress = progressListener.getCurrentProgress();
      expect(progress).toBeDefined();
      expect(progress.completed).toBe(1);
    });

    it('should handle listener errors gracefully', async () => {
      // Create a faulty listener
      const faultyListener = {
        onExecutionStart: jest.fn().mockRejectedValue(new Error('Listener error')),
        onExecutionEnd: jest.fn(),
        onSuiteStart: jest.fn(),
        onSuiteEnd: jest.fn(),
        onTestCaseStart: jest.fn(),
        onTestCaseEnd: jest.fn(),
        onTestStepStart: jest.fn(),
        onTestStepEnd: jest.fn()
      };

      engine.addListener(faultyListener);

      const testSuite: TestSuite = {
        id: 'faulty-listener-test',
        name: 'Faulty Listener Test',
        testCases: [
          {
            id: 'faulty-test-case',
            name: 'Faulty Test Case',
            steps: [
              {
                id: 'faulty-step',
                name: 'Faulty Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      // Should not throw even if listener fails
      await expect(engine.executeSuite(testSuite)).resolves.toBeDefined();
    });

    it('should support adding and removing listeners dynamically', async () => {
      const testSuite: TestSuite = {
        id: 'dynamic-listener-test',
        name: 'Dynamic Listener Test',
        testCases: [
          {
            id: 'dynamic-test-case',
            name: 'Dynamic Test Case',
            steps: [
              {
                id: 'dynamic-step',
                name: 'Dynamic Step',
                type: 'navigate',
                url: 'https://example.com', url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      // Add a new listener
      const newListener = new ConsoleLoggingListener();
      engine.addListener(newListener);

      await engine.executeSuite(testSuite);

      // Verify new listener received events
      expect(mockConsoleLog).toHaveBeenCalledWith(
        expect.stringContaining('测试执行开始')
      );

      // Remove listener
      engine.removeListener(newListener);

      // Clear console logs
      mockConsoleLog.mockClear();

      // Execute again
      await engine.executeSuite(testSuite);

      // Should not have additional logs from removed listener
      const startLogCalls = mockConsoleLog.mock.calls.filter(call =>
        call[0].includes('测试执行开始')
      );
      expect(startLogCalls).toHaveLength(0);
    });
  });

  describe('Listener Performance Impact', () => {
    it('should minimize performance overhead', async () => {
      const testSuite: TestSuite = {
        id: 'perf-overhead-test',
        name: 'Performance Overhead Test',
        testCases: Array.from({ length: 100 }, (_, i) => ({
          id: `perf-test-case-${i}`,
          name: `Performance Test Case ${i}`,
          steps: [
            {
              id: `perf-step-${i}`,
              name: `Performance Step ${i}`,
              type: 'navigate',
              url: 'https://example.com', url: `https://example.com/page${i}` }
            }
          ]
        }))
      };

      // Add multiple listeners
      const listeners = [
        new ConsoleLoggingListener(),
        new FileLoggingListener(testOutputDir),
        new PerformanceMonitoringListener(),
        new RealTimeProgressListener()
      ];

      listeners.forEach(listener => engine.addListener(listener));

      const startTime = Date.now();
      await engine.executeSuite(testSuite);
      const endTime = Date.now();

      const executionTime = endTime - startTime;

      // Should complete within reasonable time (adjust threshold as needed)
      expect(executionTime).toBeLessThan(10000); // 10 seconds for 100 tests
    });
  });
});