import {
  ConsoleLoggingListener,
  FileLoggingListener,
  PerformanceMonitoringListener,
  RealTimeProgressListener
} from '../../src/core/test-listeners';
import { TestExecutionResult, TestSuiteResult, TestCaseResult, TestStepResult } from '../../src/types/test-engine-types';
import { jest } from '@jest/globals';

describe('Test Listeners', () => {
  describe('ConsoleLoggingListener', () => {
    let listener: ConsoleLoggingListener;

    beforeEach(() => {
      listener = new ConsoleLoggingListener();
    });

    afterEach(() => {
      jest.clearAllMocks();
    });

    describe('onExecutionStart', () => {
      it('should log execution start', async () => {
        const mockContext = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          config: { testEnvironment: 'development' }
        };

        await listener.onExecutionStart(mockContext as any);
        // Should not throw, logs execution start
      });

      it('should handle missing environment config', async () => {
        const mockContext = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          config: {}
        };

        await expect(listener.onExecutionStart(mockContext as any)).resolves.not.toThrow();
      });
    });

    describe('onExecutionEnd', () => {
      it('should log execution completion with success', async () => {
        const mockResult: TestExecutionResult = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          config: {},
          total: 10,
          passed: 10,
          failed: 0,
          skipped: 0,
          pending: 0,
          suites: [],
          errors: [],
          metadata: {
            environment: 'development',
            browser: 'Chrome',
            version: '120.0.0',
            executor: 'TestEngine'
          }
        };

        await expect(listener.onExecutionEnd(mockResult)).resolves.not.toThrow();
      });

      it('should log execution completion with failures', async () => {
        const mockResult: TestExecutionResult = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          config: {},
          total: 10,
          passed: 8,
          failed: 2,
          skipped: 0,
          pending: 0,
          suites: [],
          errors: [],
          metadata: {
            environment: 'development',
            browser: 'Chrome',
            version: '120.0.0',
            executor: 'TestEngine'
          }
        };

        await expect(listener.onExecutionEnd(mockResult)).resolves.not.toThrow();
      });

      it('should log execution with errors', async () => {
        const mockResult: TestExecutionResult = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          config: {},
          total: 5,
          passed: 5,
          failed: 0,
          skipped: 0,
          pending: 0,
          suites: [],
          errors: [
            {
              timestamp: new Date(),
              type: 'system',
              message: 'System error occurred',
              component: 'TestRunner'
            }
          ],
          metadata: {
            environment: 'development',
            browser: 'Chrome',
            version: '120.0.0',
            executor: 'TestEngine'
          }
        };

        await expect(listener.onExecutionEnd(mockResult)).resolves.not.toThrow();
      });

      it('should calculate success rate correctly', async () => {
        const mockResult: TestExecutionResult = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          config: {},
          total: 3,
          passed: 2,
          failed: 1,
          skipped: 0,
          pending: 0,
          suites: [],
          errors: [],
          metadata: {
            environment: 'development',
            browser: 'Chrome',
            version: '120.0.0',
            executor: 'TestEngine'
          }
        };

        await expect(listener.onExecutionEnd(mockResult)).resolves.not.toThrow();
      });

      it('should handle zero total tests', async () => {
        const mockResult: TestExecutionResult = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          config: {},
          total: 0,
          passed: 0,
          failed: 0,
          skipped: 0,
          pending: 0,
          suites: [],
          errors: [],
          metadata: {
            environment: 'development',
            browser: 'Chrome',
            version: '120.0.0',
            executor: 'TestEngine'
          }
        };

        await expect(listener.onExecutionEnd(mockResult)).resolves.not.toThrow();
      });
    });

    describe('onSuiteStart', () => {
      it('should log suite start', async () => {
        const mockSuite = {
          id: 'suite-123',
          name: 'Test Suite',
          testCases: [{}, {}, {}] // 3 test cases
        };

        await expect(listener.onSuiteStart(mockSuite, {} as any)).resolves.not.toThrow();
      });

      it('should handle suite with no test cases', async () => {
        const mockSuite = {
          id: 'suite-123',
          name: 'Empty Suite',
          testCases: []
        };

        await expect(listener.onSuiteStart(mockSuite, {} as any)).resolves.not.toThrow();
      });
    });

    describe('onSuiteEnd', () => {
      it('should log suite completion', async () => {
        const mockResult: TestSuiteResult = {
          suiteId: 'suite-123',
          name: 'Test Suite',
          status: 'passed',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          testCases: [
            { status: 'passed' } as TestCaseResult,
            { status: 'passed' } as TestCaseResult,
            { status: 'failed' } as TestCaseResult
          ]
        };

        await expect(listener.onSuiteEnd(mockResult)).resolves.not.toThrow();
      });
    });

    describe('onTestCaseStart', () => {
      it('should log test case start', async () => {
        const mockTestCase = {
          id: 'case-123',
          name: 'Test Case',
          steps: [{}, {}] // 2 steps
        };

        await expect(listener.onTestCaseStart(mockTestCase, {} as any)).resolves.not.toThrow();
      });

      it('should handle test case with no steps', async () => {
        const mockTestCase = {
          id: 'case-123',
          name: 'Test Case',
          steps: []
        };

        await expect(listener.onTestCaseStart(mockTestCase, {} as any)).resolves.not.toThrow();
      });
    });

    describe('onTestCaseEnd', () => {
      it('should log test case completion with different statuses', async () => {
        const statuses = ['passed', 'failed', 'skipped', 'pending'] as const;

        for (const status of statuses) {
          const mockResult: TestCaseResult = {
            caseId: 'case-123',
            name: 'Test Case',
            status,
            startTime: new Date('2025-01-01T10:00:00Z'),
            endTime: new Date('2025-01-01T10:00:30Z'),
            duration: 30000,
            steps: [],
            retryCount: 0
          };

          await expect(listener.onTestCaseEnd(mockResult)).resolves.not.toThrow();
        }
      });

      it('should log retry count', async () => {
        const mockResult: TestCaseResult = {
          caseId: 'case-123',
          name: 'Test Case',
          status: 'passed',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:00:30Z'),
          duration: 30000,
          steps: [],
          retryCount: 3
        };

        await expect(listener.onTestCaseEnd(mockResult)).resolves.not.toThrow();
      });
    });

    describe('onTestStepStart', () => {
      it('should log test step start', async () => {
        const mockStep = {
          id: 'step-123',
          name: 'Test Step',
          action: 'click'
        };

        await expect(listener.onTestStepStart(mockStep, {} as any)).resolves.not.toThrow();
      });
    });

    describe('onTestStepEnd', () => {
      it('should log test step completion with different statuses', async () => {
        const statuses = ['passed', 'failed', 'skipped', 'pending'] as const;

        for (const status of statuses) {
          const mockResult: TestStepResult = {
            stepId: 'step-123',
            name: 'Test Step',
            status,
            duration: 1000,
            startTime: new Date('2025-01-01T10:00:00Z'),
            endTime: new Date('2025-01-01T10:00:01Z'),
            logs: []
          };

          await expect(listener.onTestStepEnd(mockResult)).resolves.not.toThrow();
        }
      });

      it('should log step errors', async () => {
        const mockResult: TestStepResult = {
          stepId: 'step-123',
          name: 'Test Step',
          status: 'failed',
          duration: 1000,
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:00:01Z'),
          error: {
            message: 'Step failed: Element not found'
          },
          logs: []
        };

        await expect(listener.onTestStepEnd(mockResult)).resolves.not.toThrow();
      });
    });
  });

  describe('FileLoggingListener', () => {
    let listener: FileLoggingListener;

    beforeEach(() => {
      listener = new FileLoggingListener('/tmp/test-logs.log');
    });

    afterEach(() => {
      jest.clearAllMocks();
    });

    describe('log formatting', () => {
      it('should format execution start log', async () => {
        const mockContext = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          config: { testEnvironment: 'development' }
        };

        await listener.onExecutionStart(mockContext as any);
      });

      it('should format execution end log', async () => {
        const mockResult: TestExecutionResult = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          config: {},
          total: 10,
          passed: 8,
          failed: 2,
          skipped: 0,
          pending: 0,
          suites: [],
          errors: [],
          metadata: {
            environment: 'development',
            browser: 'Chrome',
            version: '120.0.0',
            executor: 'TestEngine'
          }
        };

        await listener.onExecutionEnd(mockResult);
      });

      it('should format suite logs', async () => {
        const mockSuite = {
          id: 'suite-123',
          name: 'Test Suite',
          testCases: [{}, {}]
        };

        const mockResult: TestSuiteResult = {
          suiteId: 'suite-123',
          name: 'Test Suite',
          status: 'passed',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          testCases: [
            { status: 'passed' } as TestCaseResult,
            { status: 'passed' } as TestCaseResult
          ]
        };

        await listener.onSuiteStart(mockSuite, {} as any);
        await listener.onSuiteEnd(mockResult);
      });

      it('should format test case logs', async () => {
        const mockTestCase = {
          id: 'case-123',
          name: 'Test Case',
          steps: [{}]
        };

        const mockResult: TestCaseResult = {
          caseId: 'case-123',
          name: 'Test Case',
          status: 'passed',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:00:30Z'),
          duration: 30000,
          steps: [],
          retryCount: 0
        };

        await listener.onTestCaseStart(mockTestCase, {} as any);
        await listener.onTestCaseEnd(mockResult);
      });

      it('should format test step logs', async () => {
        const mockStep = {
          id: 'step-123',
          name: 'Test Step',
          action: 'click'
        };

        const mockResult: TestStepResult = {
          stepId: 'step-123',
          name: 'Test Step',
          status: 'passed',
          duration: 1000,
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:00:01Z'),
          logs: []
        };

        await listener.onTestStepStart(mockStep, {} as any);
        await listener.onTestStepEnd(mockResult);
      });

      it('should format step error logs', async () => {
        const mockResult: TestStepResult = {
          stepId: 'step-123',
          name: 'Test Step',
          status: 'failed',
          duration: 1000,
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:00:01Z'),
          error: {
            message: 'Step execution failed'
          },
          logs: []
        };

        await listener.onTestStepEnd(mockResult);
      });
    });
  });

  describe('PerformanceMonitoringListener', () => {
    let listener: PerformanceMonitoringListener;

    beforeEach(() => {
      listener = new PerformanceMonitoringListener();
    });

    afterEach(() => {
      jest.clearAllMocks();
    });

    describe('performance tracking', () => {
      it('should track execution performance', async () => {
        const mockContext = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          config: {}
        };

        const mockResult: TestExecutionResult = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          config: {},
          total: 10,
          passed: 10,
          failed: 0,
          skipped: 0,
          pending: 0,
          suites: [],
          errors: [],
          metadata: {
            environment: 'development',
            browser: 'Chrome',
            version: '120.0.0',
            executor: 'TestEngine'
          }
        };

        await listener.onExecutionStart(mockContext as any);
        await listener.onExecutionEnd(mockResult);
      });

      it('should track suite performance', async () => {
        const mockSuite = {
          id: 'suite-123',
          name: 'Test Suite',
          testCases: [{}, {}, {}]
        };

        const mockResult: TestSuiteResult = {
          suiteId: 'suite-123',
          name: 'Test Suite',
          status: 'passed',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          testCases: [
            { status: 'passed' } as TestCaseResult,
            { status: 'passed' } as TestCaseResult,
            { status: 'passed' } as TestCaseResult
          ]
        };

        await listener.onSuiteStart(mockSuite, {} as any);
        await listener.onSuiteEnd(mockResult);
      });

      it('should detect slow test cases', async () => {
        const mockTestCase = {
          id: 'case-123',
          name: 'Slow Test Case',
          steps: [{}]
        };

        const mockResult: TestCaseResult = {
          caseId: 'case-123',
          name: 'Slow Test Case',
          status: 'passed',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:00:06Z'), // 6 seconds - should be detected as slow
          duration: 6000,
          steps: [],
          retryCount: 0
        };

        await listener.onTestCaseStart(mockTestCase, {} as any);
        await listener.onTestCaseEnd(mockResult);
      });

      it('should detect slow test steps', async () => {
        const mockStep = {
          id: 'step-123',
          name: 'Slow Test Step',
          action: 'wait'
        };

        const mockResult: TestStepResult = {
          stepId: 'step-123',
          name: 'Slow Test Step',
          status: 'passed',
          duration: 1500, // 1.5 seconds - should be detected as slow
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:00:01.5Z'),
          logs: []
        };

        await listener.onTestStepStart(mockStep, {} as any);
        await listener.onTestStepEnd(mockResult);
      });

      it('should handle normal performance tests', async () => {
        const mockStep = {
          id: 'step-123',
          name: 'Normal Test Step',
          action: 'click'
        };

        const mockResult: TestStepResult = {
          stepId: 'step-123',
          name: 'Normal Test Step',
          status: 'passed',
          duration: 100, // 100ms - should not be detected as slow
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:00:00.1Z'),
          logs: []
        };

        await listener.onTestStepStart(mockStep, {} as any);
        await listener.onTestStepEnd(mockResult);
      });
    });
  });

  describe('RealTimeProgressListener', () => {
    let listener: RealTimeProgressListener;

    beforeEach(() => {
      listener = new RealTimeProgressListener();
    });

    afterEach(() => {
      jest.clearAllMocks();
    });

    describe('progress tracking', () => {
      it('should track execution progress', async () => {
        const mockContext = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          config: {}
        };

        const mockResult: TestExecutionResult = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          config: {},
          total: 5,
          passed: 3,
          failed: 2,
          skipped: 0,
          pending: 0,
          suites: [],
          errors: [],
          metadata: {
            environment: 'development',
            browser: 'Chrome',
            version: '120.0.0',
            executor: 'TestEngine'
          }
        };

        const mockSuite = {
          id: 'suite-123',
          name: 'Test Suite',
          testCases: [{}, {}, {}, {}, {}]
        };

        await listener.onExecutionStart(mockContext as any);
        await listener.onSuiteStart(mockSuite, {} as any);

        // Simulate test case completions
        for (let i = 0; i < 5; i++) {
          const mockCaseResult: TestCaseResult = {
            caseId: `case-${i}`,
            name: `Test Case ${i}`,
            status: i < 3 ? 'passed' : 'failed',
            startTime: new Date('2025-01-01T10:00:00Z'),
            endTime: new Date('2025-01-01T10:00:10Z'),
            duration: 10000,
            steps: []
          };

          await listener.onTestCaseEnd(mockCaseResult);
        }

        await listener.onExecutionEnd(mockResult);
      });

      it('should handle empty test execution', async () => {
        const mockContext = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          config: {}
        };

        const mockResult: TestExecutionResult = {
          executionId: 'test-execution-123',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:00:01Z'),
          duration: 1000,
          config: {},
          total: 0,
          passed: 0,
          failed: 0,
          skipped: 0,
          pending: 0,
          suites: [],
          errors: [],
          metadata: {
            environment: 'development',
            browser: 'Chrome',
            version: '120.0.0',
            executor: 'TestEngine'
          }
        };

        await listener.onExecutionStart(mockContext as any);
        await listener.onExecutionEnd(mockResult);
      });
    });
  });

  describe('Listener Error Handling', () => {
    describe('all listeners', () => {
      it('should handle all listener methods without throwing', async () => {
        const listeners = [
          new ConsoleLoggingListener(),
          new FileLoggingListener('/tmp/test.log'),
          new PerformanceMonitoringListener(),
          new RealTimeProgressListener()
        ];

        const mockContext = { executionId: 'test', startTime: new Date(), config: {} };
        const mockResult: TestExecutionResult = {
          executionId: 'test',
          startTime: new Date(),
          endTime: new Date(),
          duration: 1000,
          config: {},
          total: 1,
          passed: 1,
          failed: 0,
          skipped: 0,
          pending: 0,
          suites: [],
          errors: [],
          metadata: {
            environment: 'test',
            browser: 'Chrome',
            version: '1.0.0',
            executor: 'TestEngine'
          }
        };

        for (const listener of listeners) {
          await expect(listener.onExecutionStart(mockContext as any)).resolves.not.toThrow();
          await expect(listener.onExecutionEnd(mockResult)).resolves.not.toThrow();
          await expect(listener.onSuiteStart({ id: 'suite', name: 'Suite', testCases: [] }, {} as any)).resolves.not.toThrow();
          await expect(listener.onSuiteEnd({
            suiteId: 'suite',
            name: 'Suite',
            status: 'passed',
            startTime: new Date(),
            endTime: new Date(),
            duration: 1000,
            testCases: []
          })).resolves.not.toThrow();
          await expect(listener.onTestCaseStart({ id: 'case', name: 'Case', steps: [] }, {} as any)).resolves.not.toThrow();
          await expect(listener.onTestCaseEnd({
            caseId: 'case',
            name: 'Case',
            status: 'passed',
            startTime: new Date(),
            endTime: new Date(),
            duration: 1000,
            steps: []
          })).resolves.not.toThrow();
          await expect(listener.onTestStepStart({ id: 'step', name: 'Step', action: 'click' }, {} as any)).resolves.not.toThrow();
          await expect(listener.onTestStepEnd({
            stepId: 'step',
            name: 'Step',
            status: 'passed',
            duration: 100,
            startTime: new Date(),
            endTime: new Date(),
            logs: []
          })).resolves.not.toThrow();
        }
      });
    });
  });
});