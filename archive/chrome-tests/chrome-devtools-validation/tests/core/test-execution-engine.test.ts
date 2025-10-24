import { TestExecutionEngine } from '../../src/core/test-execution-engine';
import { JsonReportGenerator } from '../../src/core/test-report-generator';
import { MemoryDataProvider } from '../../src/core/test-data-provider';
import { ConsoleLoggingListener } from '../../src/core/test-listeners';
import { ChromeDevToolsValidator } from '../../src/index';
import { TestSuite, TestCase, TestStep } from '../../src/types/test-types';
import { jest } from '@jest/globals';

// Mock ChromeDevToolsValidator
jest.mock('../../src/index');
const MockChromeDevToolsValidator = require('../../src/index').ChromeDevToolsValidator;

describe('TestExecutionEngine', () => {
  let engine: TestExecutionEngine;
  let mockValidator: any;
  let mockListener: any;
  let mockReportGenerator: any;
  let mockDataProvider: any;

  beforeEach(() => {
    // Mock validator
    mockValidator = {
      navigate: jest.fn(),
      click: jest.fn(),
      fill: jest.fn(),
      wait: jest.fn(),
      screenshot: jest.fn(),
      executeScript: jest.fn()
    };

    MockChromeDevToolsValidator.mockImplementation(() => mockValidator);

    // Mock listener
    mockListener = {
      onExecutionStart: jest.fn(),
      onExecutionEnd: jest.fn(),
      onSuiteStart: jest.fn(),
      onSuiteEnd: jest.fn(),
      onTestCaseStart: jest.fn(),
      onTestCaseEnd: jest.fn(),
      onTestStepStart: jest.fn(),
      onTestStepEnd: jest.fn()
    };

    // Mock report generator
    mockReportGenerator = {
      generateReport: jest.fn(),
      saveReport: jest.fn()
    };

    // Mock data provider
    mockDataProvider = {
      getTestData: jest.fn(),
      setTestData: jest.fn(),
      cleanupTestData: jest.fn()
    };

    engine = new TestExecutionEngine();
    engine.addListener(mockListener);
    engine.setReportGenerator(mockReportGenerator);
    engine.setDataProvider(mockDataProvider);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('constructor', () => {
    it('should create engine with default config', () => {
      const defaultEngine = new TestExecutionEngine();
      expect(defaultEngine).toBeInstanceOf(TestExecutionEngine);
    });

    it('should create engine with custom config', () => {
      const customConfig = {
        headless: true,
        parallelExecution: true,
        maxConcurrency: 5
      };
      const customEngine = new TestExecutionEngine(customConfig);
      expect(customEngine).toBeInstanceOf(TestExecutionEngine);
    });
  });

  describe('addListener and removeListener', () => {
    it('should add listener', () => {
      const newListener = { ...mockListener };
      engine.addListener(newListener);
      expect(engine['listeners']).toContain(newListener);
    });

    it('should remove listener', () => {
      engine.addListener(mockListener);
      engine.removeListener(mockListener);
      expect(engine['listeners']).not.toContain(mockListener);
    });
  });

  describe('setReportGenerator', () => {
    it('should set report generator', () => {
      const generator = new JsonReportGenerator();
      engine.setReportGenerator(generator);
      expect(engine['reportGenerator']).toBe(generator);
    });
  });

  describe('setDataProvider', () => {
    it('should set data provider', () => {
      const provider = new MemoryDataProvider();
      engine.setDataProvider(provider);
      expect(engine['dataProvider']).toBe(provider);
    });
  });

  describe('executeSuite', () => {
    const mockSuite: TestSuite = {
      id: 'test-suite',
      name: 'Test Suite',
      description: 'Test suite description',
      testCases: [
        {
          id: 'test-case-1',
          name: 'Test Case 1',
          description: 'Test case description',
          steps: [
            {
              id: 'step-1',
              name: 'Navigate to page',
              action: 'navigate',
              parameters: { url: 'https://example.com' }
            },
            {
              id: 'step-2',
              name: 'Click element',
              action: 'click',
              parameters: { selector: '#button' }
            }
          ]
        }
      ]
    };

    it('should execute test suite successfully', async () => {
      mockValidator.navigate.mockResolvedValue(undefined);
      mockValidator.click.mockResolvedValue(undefined);
      mockReportGenerator.generateReport.mockResolvedValue('{}');
      mockReportGenerator.saveReport.mockResolvedValue(undefined);

      const result = await engine.executeSuite(mockSuite);

      expect(result).toBeDefined();
      expect(result.executionId).toBeDefined();
      expect(result.total).toBe(1);
      expect(result.suites).toHaveLength(1);
      expect(result.suites[0].testCases).toHaveLength(1);

      // Verify validator calls
      expect(mockValidator.navigate).toHaveBeenCalledWith('https://example.com');
      expect(mockValidator.click).toHaveBeenCalledWith('#button');

      // Verify listener calls
      expect(mockListener.onExecutionStart).toHaveBeenCalled();
      expect(mockListener.onExecutionEnd).toHaveBeenCalled();
      expect(mockListener.onSuiteStart).toHaveBeenCalled();
      expect(mockListener.onSuiteEnd).toHaveBeenCalled();
      expect(mockListener.onTestCaseStart).toHaveBeenCalled();
      expect(mockListener.onTestCaseEnd).toHaveBeenCalled();
      expect(mockListener.onTestStepStart).toHaveBeenCalledTimes(2);
      expect(mockListener.onTestStepEnd).toHaveBeenCalledTimes(2);

      // Verify report generation
      expect(mockReportGenerator.generateReport).toHaveBeenCalled();
    });

    it('should handle validator errors', async () => {
      mockValidator.navigate.mockResolvedValue(undefined);
      mockValidator.click.mockRejectedValue(new Error('Click failed'));
      mockReportGenerator.generateReport.mockResolvedValue('{}');
      mockReportGenerator.saveReport.mockResolvedValue(undefined);

      const result = await engine.executeSuite(mockSuite);

      expect(result.failed).toBe(1);
      expect(result.suites[0].testCases[0].status).toBe('failed');
      expect(result.suites[0].testCases[0].steps[1].status).toBe('failed');
    });

    it('should handle empty test suite', async () => {
      const emptySuite: TestSuite = {
        id: 'empty-suite',
        name: 'Empty Suite',
        testCases: []
      };

      mockReportGenerator.generateReport.mockResolvedValue('{}');
      mockReportGenerator.saveReport.mockResolvedValue(undefined);

      const result = await engine.executeSuite(emptySuite);

      expect(result.total).toBe(0);
      expect(result.suites[0].testCases).toHaveLength(0);
    });

    it('should execute with hooks', async () => {
      const mockContext = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: engine['config'],
        validator: mockValidator,
        sharedData: new Map(),
        hooks: {
          beforeAll: jest.fn(),
          afterAll: jest.fn(),
          beforeSuite: jest.fn(),
          afterSuite: jest.fn(),
          beforeTest: jest.fn(),
          afterTest: jest.fn(),
          beforeStep: jest.fn(),
          afterStep: jest.fn()
        }
      };

      mockValidator.navigate.mockResolvedValue(undefined);
      mockValidator.click.mockResolvedValue(undefined);
      mockReportGenerator.generateReport.mockResolvedValue('{}');
      mockReportGenerator.saveReport.mockResolvedValue(undefined);

      await engine.executeSuite(mockSuite, mockContext);

      expect(mockContext.hooks.beforeAll).toHaveBeenCalled();
      expect(mockContext.hooks.afterAll).toHaveBeenCalled();
      expect(mockContext.hooks.beforeSuite).toHaveBeenCalled();
      expect(mockContext.hooks.afterSuite).toHaveBeenCalled();
      expect(mockContext.hooks.beforeTest).toHaveBeenCalled();
      expect(mockContext.hooks.afterTest).toHaveBeenCalled();
      expect(mockContext.hooks.beforeStep).toHaveBeenCalledTimes(2);
      expect(mockContext.hooks.afterStep).toHaveBeenCalledTimes(2);
    });
  });

  describe('executeTestCase', () => {
    const mockTestCase: TestCase = {
      id: 'test-case',
      name: 'Test Case',
      description: 'Test case description',
      steps: [
        {
          id: 'step-1',
          name: 'Navigate to page',
          action: 'navigate',
          parameters: { url: 'https://example.com' }
        }
      ]
    };

    it('should execute single test case', async () => {
      mockValidator.navigate.mockResolvedValue(undefined);
      mockReportGenerator.generateReport.mockResolvedValue('{}');
      mockReportGenerator.saveReport.mockResolvedValue(undefined);

      const context = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: engine['config'],
        validator: mockValidator,
        sharedData: new Map(),
        hooks: {}
      };

      const result = await engine.executeTestCase(mockTestCase, context);

      expect(result).toBeDefined();
      expect(result.total).toBe(1);
      expect(result.passed).toBe(1);
      expect(result.suites[0].testCases[0].status).toBe('passed');
    });
  });

  describe('executeTestStep', () => {
    const mockStep: TestStep = {
      id: 'test-step',
      name: 'Navigate to page',
      action: 'navigate',
      parameters: { url: 'https://example.com' }
    };

    it('should execute single test step', async () => {
      mockValidator.navigate.mockResolvedValue(undefined);

      const context = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: engine['config'],
        validator: mockValidator,
        sharedData: new Map(),
        hooks: {}
      };

      const result = await engine.executeTestStep(mockStep, context);

      expect(result).toBeDefined();
      expect(result.stepId).toBe('test-step');
      expect(result.status).toBe('passed');
      expect(mockValidator.navigate).toHaveBeenCalledWith('https://example.com');
    });

    it('should handle step execution error', async () => {
      mockValidator.navigate.mockRejectedValue(new Error('Navigation failed'));

      const context = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: engine['config'],
        validator: mockValidator,
        sharedData: new Map(),
        hooks: {}
      };

      const result = await engine.executeTestStep(mockStep, context);

      expect(result.status).toBe('failed');
      expect(result.error).toBeDefined();
      expect(result.error?.message).toBe('Navigation failed');
    });

    it('should support different step actions', async () => {
      const testSteps = [
        { action: 'navigate', parameters: { url: 'https://example.com' }, mockFn: 'navigate' },
        { action: 'click', parameters: { selector: '#button' }, mockFn: 'click' },
        { action: 'fill', parameters: { selector: '#input', value: 'test' }, mockFn: 'fill' },
        { action: 'wait', parameters: { timeout: 1000 }, mockFn: 'wait' },
        { action: 'screenshot', parameters: { filename: 'test.png' }, mockFn: 'screenshot' },
        { action: 'executeScript', parameters: { script: 'console.log("test")' }, mockFn: 'executeScript' }
      ];

      const context = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: engine['config'],
        validator: mockValidator,
        sharedData: new Map(),
        hooks: {}
      };

      for (const stepConfig of testSteps) {
        const step: TestStep = {
          id: `step-${stepConfig.action}`,
          name: `Test ${stepConfig.action}`,
          action: stepConfig.action as any,
          parameters: stepConfig.parameters
        };

        mockValidator[stepConfig.mockFn].mockResolvedValue(undefined);

        const result = await engine.executeTestStep(step, context);

        expect(result.status).toBe('passed');
        expect(mockValidator[stepConfig.mockFn]).toHaveBeenCalled();
      }
    });
  });

  describe('error handling', () => {
    it('should handle listener errors gracefully', async () => {
      const errorListener = {
        onExecutionStart: jest.fn().mockRejectedValue(new Error('Listener error')),
        onExecutionEnd: jest.fn(),
        onSuiteStart: jest.fn(),
        onSuiteEnd: jest.fn(),
        onTestCaseStart: jest.fn(),
        onTestCaseEnd: jest.fn(),
        onTestStepStart: jest.fn(),
        onTestStepEnd: jest.fn()
      };

      engine.addListener(errorListener);

      const mockSuite: TestSuite = {
        id: 'test-suite',
        name: 'Test Suite',
        testCases: []
      };

      mockReportGenerator.generateReport.mockResolvedValue('{}');
      mockReportGenerator.saveReport.mockResolvedValue(undefined);

      // Should not throw even if listener fails
      await expect(engine.executeSuite(mockSuite)).resolves.toBeDefined();
    });

    it('should handle report generation errors', async () => {
      const mockSuite: TestSuite = {
        id: 'test-suite',
        name: 'Test Suite',
        testCases: []
      };

      mockReportGenerator.generateReport.mockRejectedValue(new Error('Report generation failed'));

      // Should not throw even if report generation fails
      await expect(engine.executeSuite(mockSuite)).resolves.toBeDefined();
    });
  });

  describe('configuration', () => {
    it('should use custom configuration', async () => {
      const customConfig = {
        parallelExecution: true,
        maxConcurrency: 2,
        retryFailedTests: false,
        screenshotsOnFailure: false
      };

      const customEngine = new TestExecutionEngine(customConfig);
      customEngine.setReportGenerator(mockReportGenerator);

      const mockSuite: TestSuite = {
        id: 'test-suite',
        name: 'Test Suite',
        testCases: [
          {
            id: 'test-case-1',
            name: 'Test Case 1',
            steps: [
              {
                id: 'step-1',
                name: 'Navigate to page',
                action: 'navigate',
                parameters: { url: 'https://example.com' }
              }
            ]
          }
        ]
      };

      mockValidator.navigate.mockResolvedValue(undefined);
      mockReportGenerator.generateReport.mockResolvedValue('{}');
      mockReportGenerator.saveReport.mockResolvedValue(undefined);

      await customEngine.executeSuite(mockSuite);

      expect(customEngine['config']).toMatchObject(customConfig);
    });
  });
});