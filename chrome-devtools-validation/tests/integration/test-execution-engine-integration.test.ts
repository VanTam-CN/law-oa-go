import { ChromeDevToolsTestExecutionEngine } from '../../src/core/test-execution-engine';
import { JsonReportGenerator, HtmlReportGenerator, JunitReportGenerator } from '../../src/core/test-report-generator';
import { MemoryDataProvider } from '../../src/core/test-data-provider';
import { ConsoleLoggingListener, FileLoggingListener, PerformanceMonitoringListener } from '../../src/core/test-listeners';
import { TestSuite } from '../../src/types/test-types';
import { TestListener } from '../../src/types/test-engine-types';
import { jest } from '@jest/globals';

// Mock ChromeDevToolsValidator
jest.mock('../../src/index');
const MockChromeDevToolsValidator = require('../../src/index').ChromeDevToolsValidator;

describe('Test Execution Engine Integration Tests', () => {
  let engine: ChromeDevToolsTestExecutionEngine;
  let mockValidator: any;
  let mockDataProvider: MemoryDataProvider;
  let mockListeners: TestListener[];
  let mockReportGenerators: any[];

  // Helper function to create test context
  const createTestContext = (executionId: string = 'test-execution'): any => ({
    executionId,
    startTime: new Date(),
    config: engine.getConfig(),
    validator: mockValidator,
    sharedData: {},
    hooks: {
      beforeSuite: jest.fn(),
      afterSuite: jest.fn(),
      beforeTestCase: jest.fn(),
      afterTestCase: jest.fn(),
      beforeTestStep: jest.fn(),
      afterTestStep: jest.fn()
    }
  });

  beforeEach(() => {
    // Mock validator
    mockValidator = {
      navigate: jest.fn(),
      click: jest.fn(),
      fill: jest.fn(),
      wait: jest.fn(),
      screenshot: jest.fn(),
      executeScript: jest.fn(),
      takeSnapshot: jest.fn(),
      evaluate: jest.fn(),
      close: jest.fn()
    };

    MockChromeDevToolsValidator.mockImplementation(() => mockValidator);

    // Create real data provider
    mockDataProvider = new MemoryDataProvider();

    // Create real listeners
    mockListeners = [
      new ConsoleLoggingListener(),
      new FileLoggingListener('/tmp/test-logs'),
      new PerformanceMonitoringListener()
    ];

    // Create real report generators
    mockReportGenerators = [
      new JsonReportGenerator(),
      new HtmlReportGenerator(),
      new JunitReportGenerator()
    ];

    // Create engine with real components
    engine = new ChromeDevToolsTestExecutionEngine({
      headless: true,
      parallelExecution: false,
      maxConcurrency: 1,
      retryFailedTests: false, // 禁用重试以简化测试
      screenshotsOnFailure: true,
      generateReport: true,
      reportFormat: 'json'
    });

    // Add all listeners
    mockListeners.forEach(listener => engine.addListener(listener));

    // Set data provider
    engine.setDataProvider(mockDataProvider);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('End-to-End Test Execution', () => {
    const testSuite: TestSuite = {
      id: 'integration-test-suite',
      name: 'Integration Test Suite',
      description: 'Comprehensive integration test suite',
      tags: ['integration'],
      timeout: 60000,
      testCases: [
        {
          id: 'login-test',
          name: 'Login Functionality Test',
          description: 'Test user login workflow',
          assertions: [],
          tags: ['auth', 'login'],
          timeout: 30000,
          priority: 'P0',
          steps: [
            {
              id: 'navigate-to-login',
              name: 'Navigate to Login Page',
              type: 'navigate',
              url: 'https://example.com/login'
            },
            {
              id: 'enter-credentials',
              name: 'Enter User Credentials',
              type: 'fill',
              selector: '#username',
              value: 'testuser@example.com'
            },
            {
              id: 'enter-password',
              name: 'Enter Password',
              type: 'fill',
              selector: '#password',
              value: 'testpass123'
            },
            {
              id: 'submit-login',
              name: 'Submit Login Form',
              type: 'click',
              selector: '#login-button'
            },
            {
              id: 'verify-login',
              name: 'Verify Successful Login',
              type: 'wait',
              selector: '#welcome-message',
              timeout: 5000,
              expectedState: { visible: true }
            },
            {
              id: 'execute-post-login-script',
              name: 'Execute Post Login Script',
              type: 'executeScript',
              script: 'return document.readyState'
            }
          ]
        },
        {
          id: 'dashboard-test',
          name: 'Dashboard Functionality Test',
          description: 'Test dashboard features',
          assertions: [],
          tags: ['dashboard'],
          timeout: 30000,
          priority: 'P1',
          steps: [
            {
              id: 'navigate-to-dashboard',
              name: 'Navigate to Dashboard',
              type: 'navigate',
              url: 'https://example.com/dashboard'
            },
            {
              id: 'check-dashboard-elements',
              name: 'Verify Dashboard Elements',
              type: 'verify',
              selector: '.dashboard-widget',
              expectedState: { visible: true }
            }
          ]
        }
      ]
    };

    it('should execute complete test suite with all components', async () => {
      // Setup validator responses
      mockValidator.navigate.mockResolvedValue(undefined);
      mockValidator.fill.mockResolvedValue(undefined);
      mockValidator.click.mockResolvedValue(undefined);
      mockValidator.wait.mockResolvedValue(undefined);
      mockValidator.executeScript.mockResolvedValue(true);
      mockValidator.screenshot.mockResolvedValue('screenshot.png');
      mockValidator.takeSnapshot.mockResolvedValue({
        elements: [
          {
            uid: '.dashboard-widget',
            visible: true,
            enabled: true,
            textContent: 'Dashboard Widget'
          }
        ]
      });

      // Set report generator
      engine.setReportGenerator(mockReportGenerators[0]); // JSON generator

      // Create execution context
      const context: any = {
        executionId: 'test-execution-123',
        startTime: new Date(),
        config: engine.getConfig(),
        validator: mockValidator,
        sharedData: {},
        hooks: {
          beforeSuite: jest.fn(),
          afterSuite: jest.fn(),
          beforeTestCase: jest.fn(),
          afterTestCase: jest.fn(),
          beforeTestStep: jest.fn(),
          afterTestStep: jest.fn()
        }
      };

      // Execute test suite
      const result = await engine.executeSuite(testSuite, context);

      // Verify execution result
      expect(result).toBeDefined();
      expect(result.executionId).toBeDefined();
      expect(result.total).toBe(2);
      expect(result.suites).toHaveLength(1);
      expect(result.suites[0].testCases).toHaveLength(2);
      expect(result.startTime).toBeInstanceOf(Date);
      expect(result.endTime).toBeInstanceOf(Date);
      expect(result.duration).toBeGreaterThan(0);

      // Verify all validator calls
      expect(mockValidator.navigate).toHaveBeenCalledTimes(2);
      expect(mockValidator.fill).toHaveBeenCalledTimes(2);
      expect(mockValidator.click).toHaveBeenCalledTimes(1);
      expect(mockValidator.wait).toHaveBeenCalledTimes(1);
      expect(mockValidator.executeScript).toHaveBeenCalledTimes(1);

      // Verify test results
      const loginTest = result.suites[0].testCases[0];
      expect(loginTest.status).toBe('passed');
      expect(loginTest.steps).toHaveLength(6);
      expect(loginTest.steps.every((step: any) => step.status === 'passed')).toBe(true);

      const dashboardTest = result.suites[0].testCases[1];
      expect(dashboardTest.status).toBe('passed');
      expect(dashboardTest.steps).toHaveLength(2);
      expect(dashboardTest.steps.every((step: any) => step.status === 'passed')).toBe(true);
    });

    it('should handle test failures gracefully', async () => {
      // Setup validator to simulate failure
      mockValidator.navigate.mockResolvedValue(undefined);
      mockValidator.fill.mockRejectedValue(new Error('Element not found: #username'));
      mockValidator.click.mockResolvedValue(undefined);
      mockValidator.wait.mockResolvedValue(undefined);
      mockValidator.executeScript.mockResolvedValue(true);
      mockValidator.screenshot.mockResolvedValue('error-screenshot.png');

      engine.setReportGenerator(mockReportGenerators[0]);

      const result = await engine.executeSuite(testSuite, createTestContext('failure-test'));

      expect(result.failed).toBe(1);
      expect(result.passed).toBe(1);

      const loginTest = result.suites[0].testCases[0];
      expect(loginTest.status).toBe('failed');
      expect(loginTest.error).toBeDefined();
      expect(loginTest.error?.message).toContain('Element not found');

      // Verify screenshot was taken on failure
      expect(mockValidator.screenshot).toHaveBeenCalled();
    });

    it('should support multiple report formats', async () => {
      mockValidator.navigate.mockResolvedValue(undefined);
      mockValidator.fill.mockResolvedValue(undefined);
      mockValidator.click.mockResolvedValue(undefined);
      mockValidator.wait.mockResolvedValue(undefined);
      mockValidator.executeScript.mockResolvedValue(true);

      const results: any[] = [];

      // Test each report generator
      for (const generator of mockReportGenerators) {
        engine.setReportGenerator(generator);
        const result = await engine.executeSuite(testSuite, createTestContext(`report-test-${generator.constructor.name}`));
        results.push(result);

        expect(result).toBeDefined();
        expect(result.total).toBe(2);
      }

      // Verify all generators produced results
      expect(results).toHaveLength(3);
      results.forEach(result => {
        expect(result.total).toBe(2);
        expect(result.suites).toHaveLength(1);
      });
    });
  });

  describe('Data Provider Integration', () => {
    it('should integrate with data provider for test data', async () => {
      // Setup test data
      const testData = {
        users: [
          { username: 'admin', password: 'admin123' },
          { username: 'user', password: 'user123' }
        ],
        urls: {
          login: 'https://example.com/login',
          dashboard: 'https://example.com/dashboard'
        }
      };

      await mockDataProvider.setTestData('test-data', testData);

      const testSuiteWithData: TestSuite = {
        id: 'data-driven-test',
        name: 'Data Driven Test',
        description: 'Test with external data',
        tags: ['data'],
        timeout: 30000,
        testCases: [
          {
            id: 'test-with-data',
            name: 'Test with External Data',
            description: 'Test using external data provider',
            assertions: [],
            tags: ['data'],
            timeout: 30000,
            priority: 'P1',
            steps: [
              {
                id: 'use-data',
                name: 'Use Test Data',
                type: 'navigate',
                url: 'https://example.com'
              },
              {
                id: 'execute-data-script',
                name: 'Execute Data Script',
                type: 'executeScript',
                script: `console.log('Test data:', ${JSON.stringify(testData)})`
              }
            ]
          }
        ]
      };

      mockValidator.executeScript.mockResolvedValue(true);
      engine.setReportGenerator(mockReportGenerators[0]);

      const result = await engine.executeSuite(testSuiteWithData, createTestContext('data-test'));

      expect(result).toBeDefined();
      expect(result.total).toBe(1);
      expect(result.passed).toBe(1);

      // Verify data was accessible during test execution
      expect(mockValidator.executeScript).toHaveBeenCalledWith(
        expect.stringContaining(JSON.stringify(testData))
      );
    });
  });

  describe('Performance Monitoring', () => {
    it('should track execution performance metrics', async () => {
      const testSuite: TestSuite = {
        id: 'performance-test',
        name: 'Performance Test',
        description: 'Performance monitoring test',
        tags: ['performance'],
        timeout: 30000,
        testCases: [
          {
            id: 'quick-test',
            name: 'Quick Test',
            description: 'Quick performance test',
            assertions: [],
            tags: ['performance'],
            timeout: 30000,
            priority: 'P1',
            steps: [
              {
                id: 'fast-action',
                name: 'Fast Action',
                type: 'navigate',
                url: 'https://example.com'
              }
            ]
          }
        ]
      };

      mockValidator.navigate.mockResolvedValue(undefined);
      engine.setReportGenerator(mockReportGenerators[0]);

      const startTime = Date.now();
      const result = await engine.executeSuite(testSuite, createTestContext());
      const endTime = Date.now();

      expect(result.duration).toBeGreaterThan(0);
      expect(result.duration).toBeLessThan(endTime - startTime + 100); // Allow some overhead

      // Verify performance tracking in test cases
      const testCase = result.suites[0].testCases[0];
      expect(testCase.duration).toBeGreaterThan(0);
      expect(testCase.startTime).toBeInstanceOf(Date);
      expect(testCase.endTime).toBeInstanceOf(Date);
    });
  });

  describe('Listener Integration', () => {
    it('should notify all listeners during execution', async () => {
      const testSuite: TestSuite = {
        id: 'listener-test',
        name: 'Listener Test',
        description: 'Listener system test',
        tags: ['listener'],
        timeout: 30000,
        testCases: [
          {
            id: 'simple-test',
            name: 'Simple Test',
            description: 'Simple listener test',
            assertions: [],
            tags: ['listener'],
            timeout: 30000,
            priority: 'P1',
            steps: [
              {
                id: 'simple-step',
                name: 'Simple Step',
                type: 'navigate',
                url: 'https://example.com'
              }
            ]
          }
        ]
      };

      mockValidator.navigate.mockResolvedValue(undefined);
      engine.setReportGenerator(mockReportGenerators[0]);

      // Spy on listener methods
      const consoleSpy = jest.spyOn(mockListeners[0] as any, 'onExecutionStart');
      const fileSpy = jest.spyOn(mockListeners[1] as any, 'onTestCaseStart');
      const perfSpy = jest.spyOn(mockListeners[2] as any, 'onTestStepEnd');

      await engine.executeSuite(testSuite, createTestContext());

      // Verify listeners were called
      expect(consoleSpy).toHaveBeenCalled();
      expect(fileSpy).toHaveBeenCalled();
      expect(perfSpy).toHaveBeenCalled();
    });
  });

  describe('Error Recovery and Retry', () => {
    it('should retry failed tests when configured', async () => {
      const testSuite: TestSuite = {
        id: 'retry-test',
        name: 'Retry Test',
        description: 'Retry mechanism test',
        tags: ['retry'],
        timeout: 30000,
        testCases: [
          {
            id: 'flaky-test',
            name: 'Flaky Test',
            description: 'Flaky test for retry mechanism',
            assertions: [],
            tags: ['retry'],
            timeout: 30000,
            priority: 'P1',
            steps: [
              {
                id: 'flaky-step',
                name: 'Flaky Step',
                type: 'navigate',
                url: 'https://example.com'
              }
            ]
          }
        ]
      };

      // Setup validator to fail first time, succeed second time
      let callCount = 0;
      mockValidator.navigate.mockImplementation(() => {
        callCount++;
        if (callCount === 1) {
          return Promise.reject(new Error('Network error'));
        }
        return Promise.resolve(undefined);
      });

      // Create a new engine with retry enabled for this test
      const retryEngine = new ChromeDevToolsTestExecutionEngine({
        retryFailedTests: true,
        maxRetries: 3,
        generateReport: false
      });
      retryEngine.setReportGenerator(mockReportGenerators[0]);

      const result = await retryEngine.executeSuite(testSuite, createTestContext());

      // Should eventually succeed due to retry
      expect(result.passed).toBe(1);
      expect(result.failed).toBe(0);
      expect(callCount).toBeGreaterThan(1); // Should have been called multiple times
    });
  });

  describe('Concurrent Execution', () => {
    it('should handle parallel test execution', async () => {
      const parallelEngine = new ChromeDevToolsTestExecutionEngine({
        parallelExecution: true,
        maxConcurrency: 2
      });

      parallelEngine.setReportGenerator(mockReportGenerators[0]);

      const testSuite: TestSuite = {
        id: 'parallel-test',
        name: 'Parallel Test',
        description: 'Parallel execution test',
        tags: ['parallel'],
        timeout: 30000,
        testCases: [
          {
            id: 'test-1',
            name: 'Test 1',
            description: 'Parallel test 1',
            assertions: [],
            tags: ['parallel'],
            timeout: 30000,
            priority: 'P1',
            steps: [
              {
                id: 'step-1',
                name: 'Step 1',
                type: 'navigate',
                url: 'https://example.com/page1'
              }
            ]
          },
          {
            id: 'test-2',
            name: 'Test 2',
            description: 'Parallel test 2',
            assertions: [],
            tags: ['parallel'],
            timeout: 30000,
            priority: 'P1',
            steps: [
              {
                id: 'step-2',
                name: 'Step 2',
                type: 'navigate',
                url: 'https://example.com/page2'
              }
            ]
          }
        ]
      };

      mockValidator.navigate.mockResolvedValue(undefined);

      const startTime = Date.now();
      const result = await parallelEngine.executeSuite(testSuite, createTestContext('parallel-test'));
      const endTime = Date.now();

      expect(result.total).toBe(2);
      expect(result.passed).toBe(2);

      // Parallel execution should be faster than sequential
      // (This is a rough estimate, actual performance may vary)
      expect(result.duration).toBeLessThan(endTime - startTime);
    });
  });

  describe('Resource Cleanup', () => {
    it('should clean up resources after execution', async () => {
      const testSuite: TestSuite = {
        id: 'cleanup-test',
        name: 'Cleanup Test',
        description: 'Resource cleanup test',
        tags: ['cleanup'],
        timeout: 30000,
        testCases: [
          {
            id: 'cleanup-test-case',
            name: 'Cleanup Test Case',
            description: 'Resource cleanup test case',
            assertions: [],
            tags: ['cleanup'],
            timeout: 30000,
            priority: 'P1',
            steps: [
              {
                id: 'cleanup-step',
                name: 'Cleanup Step',
                type: 'navigate',
                url: 'https://example.com'
              }
            ]
          }
        ]
      };

      mockValidator.navigate.mockResolvedValue(undefined);
      mockValidator.close.mockResolvedValue(undefined);
      engine.setReportGenerator(mockReportGenerators[0]);

      await engine.executeSuite(testSuite, createTestContext());

      // Manually trigger cleanup
      await engine.cleanup();

      // Verify cleanup was called
      expect(mockValidator.close).toHaveBeenCalled();
    });
  });

  describe('Configuration Validation', () => {
    it('should validate configuration parameters', async () => {
      const invalidConfigs = [
        { parallelExecution: true, maxConcurrency: 0 }, // Invalid concurrency
        { retryFailedTests: true, maxRetries: -1 }, // Invalid retry count
        { screenshotsOnFailure: true, screenshotPath: '' } // Invalid screenshot path
      ];

      for (const config of invalidConfigs) {
        const engine = new ChromeDevToolsTestExecutionEngine(config);
        // Should not throw, but should use defaults for invalid values
        expect(engine).toBeDefined();
      }
    });
  });
});