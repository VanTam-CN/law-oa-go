import {
  JsonReportGenerator,
  HtmlReportGenerator,
  JunitReportGenerator
} from '../../src/core/test-report-generator';
import { TestExecutionResult } from '../../src/types/test-engine-types';
import { readFileSync, existsSync, unlinkSync } from 'fs';
import { join } from 'path';

describe('Report Generation Integration Tests', () => {
  const testOutputDir = join(__dirname, '../test-output');
  let mockResult: TestExecutionResult;

  beforeEach(() => {
    // Create test output directory
    if (!existsSync(testOutputDir)) {
      require('fs').mkdirSync(testOutputDir, { recursive: true });
    }

    mockResult = {
      executionId: 'integration-test-execution-123',
      startTime: new Date('2025-01-01T10:00:00Z'),
      endTime: new Date('2025-01-01T10:02:30Z'),
      duration: 150000,
      config: {
        generateReport: true,
        reportFormat: 'json',
        testEnvironment: 'staging'
      },
      total: 5,
      passed: 3,
      failed: 1,
      skipped: 1,
      pending: 0,
      suites: [
        {
          suiteId: 'suite-1',
          name: 'Authentication Suite',
          description: 'Test authentication functionality',
          status: 'passed',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          testCases: [
            {
              caseId: 'case-1',
              name: 'Successful Login',
              description: 'Test successful user login',
              status: 'passed',
              startTime: new Date('2025-01-01T10:00:00Z'),
              endTime: new Date('2025-01-01T10:00:30Z'),
              duration: 30000,
              steps: [
                {
                  stepId: 'step-1',
                  name: 'Navigate to login page',
                  action: 'navigate',
                  status: 'passed',
                  duration: 10000,
                  startTime: new Date('2025-01-01T10:00:00Z'),
                  endTime: new Date('2025-01-01T10:00:10Z'),
                  logs: [{ timestamp: new Date(), level: 'info', message: 'Navigated to login page' }]
                },
                {
                  stepId: 'step-2',
                  name: 'Enter credentials',
                  action: 'fill',
                  status: 'passed',
                  duration: 15000,
                  startTime: new Date('2025-01-01T10:00:10Z'),
                  endTime: new Date('2025-01-01T10:00:25Z'),
                  logs: [{ timestamp: new Date(), level: 'info', message: 'Entered username and password' }]
                },
                {
                  stepId: 'step-3',
                  name: 'Submit form',
                  action: 'click',
                  status: 'passed',
                  duration: 5000,
                  startTime: new Date('2025-01-01T10:00:25Z'),
                  endTime: new Date('2025-01-01T10:00:30Z'),
                  logs: [{ timestamp: new Date(), level: 'info', message: 'Clicked login button' }]
                }
              ]
            }
          ]
        },
        {
          suiteId: 'suite-2',
          name: 'Failed Test Suite',
          description: 'Test suite with failures',
          status: 'failed',
          startTime: new Date('2025-01-01T10:01:00Z'),
          endTime: new Date('2025-01-01T10:02:30Z'),
          duration: 90000,
          testCases: [
            {
              caseId: 'case-2',
              name: 'Failed Test Case',
              description: 'Test case that fails',
              status: 'failed',
              startTime: new Date('2025-01-01T10:01:00Z'),
              endTime: new Date('2025-01-01T10:01:45Z'),
              duration: 45000,
              error: {
                message: 'Element not found: #submit-button',
                stack: 'Error: Element not found\\n    at TestCase.execute\\n    at TestStepExecutor.execute'
              },
              steps: [
                {
                  stepId: 'step-4',
                  name: 'Failed step',
                  action: 'click',
                  status: 'failed',
                  duration: 20000,
                  startTime: new Date('2025-01-01T10:01:00Z'),
                  endTime: new Date('2025-01-01T10:01:20Z'),
                  error: {
                    message: 'Element not found: #submit-button',
                    stack: 'Error: Element not found\\n    at TestStep.execute'
                  },
                  logs: [
                    { timestamp: new Date(), level: 'info', message: 'Attempted to click submit button' },
                    { timestamp: new Date(), level: 'error', message: 'Element not found' }
                  ]
                },
                {
                  stepId: 'step-5',
                  name: 'Screenshot on failure',
                  action: 'screenshot',
                  status: 'passed',
                  duration: 5000,
                  startTime: new Date('2025-01-01T10:01:20Z'),
                  endTime: new Date('2025-01-01T10:01:25Z'),
                  logs: [{ timestamp: new Date(), level: 'info', message: 'Took screenshot: failure-screenshot.png' }]
                }
              ]
            },
            {
              caseId: 'case-3',
              name: 'Skipped Test Case',
              description: 'Test case that was skipped',
              status: 'skipped',
              startTime: new Date('2025-01-01T10:01:25Z'),
              endTime: new Date('2025-01-01T10:01:25Z'),
              duration: 0,
              steps: []
            }
          ]
        }
      ],
      errors: [
        {
          timestamp: new Date('2025-01-01T10:01:20Z'),
          type: 'validation',
          message: 'Test environment validation failed',
          stack: 'Error: Validation failed\\n    at Validator.validate',
          component: 'TestRunner'
        }
      ],
      metadata: {
        environment: 'integration',
        browser: 'Chrome',
        version: '120.0.0',
        executor: 'ChromeDevToolsValidationEngine',
        platform: process.platform
      }
    };
  });

  afterEach(() => {
    // Clean up test files
    const cleanup = (dir: string) => {
      if (existsSync(dir)) {
        const files = require('fs').readdirSync(dir);
        files.forEach((file: string) => {
          const filePath = join(dir, file);
          unlinkSync(filePath);
        });
      }
    };
    cleanup(testOutputDir);
  });

  describe('JSON Report Generation', () => {
    let generator: JsonReportGenerator;

    beforeEach(() => {
      generator = new JsonReportGenerator();
    });

    it('should generate valid JSON report with comprehensive data', async () => {
      const report = await generator.generateReport(mockResult);

      // Parse and validate JSON
      expect(() => JSON.parse(report)).not.toThrow();
      const parsed = JSON.parse(report);

      // Validate structure
      expect(parsed).toHaveProperty('execution');
      expect(parsed).toHaveProperty('summary');
      expect(parsed).toHaveProperty('suites');
      expect(parsed).toHaveProperty('errors');

      // Validate execution data
      expect(parsed.execution.id).toBe(mockResult.executionId);
      expect(parsed.execution.environment).toBe(mockResult.metadata.environment);
      expect(parsed.execution.browser).toBe(mockResult.metadata.browser);
      expect(parsed.execution.version).toBe(mockResult.metadata.version);
      expect(parsed.execution.executor).toBe(mockResult.metadata.executor);

      // Validate summary
      expect(parsed.summary.total).toBe(mockResult.total);
      expect(parsed.summary.passed).toBe(mockResult.passed);
      expect(parsed.summary.failed).toBe(mockResult.failed);
      expect(parsed.summary.skipped).toBe(mockResult.skipped);
      expect(parsed.summary.successRate).toBe(60); // 3/5 * 100
      expect(parsed.summary.duration).toBe(mockResult.duration);

      // Validate suites
      expect(parsed.suites).toHaveLength(2);
      expect(parsed.suites[0].name).toBe('Authentication Suite');
      expect(parsed.suites[1].name).toBe('Failed Test Suite');

      // Validate test cases
      const authSuite = parsed.suites[0];
      expect(authSuite.testCases).toHaveLength(1);
      expect(authSuite.testCases[0].name).toBe('Successful Login');
      expect(authSuite.testCases[0].status).toBe('passed');
      expect(authSuite.testCases[0].steps).toHaveLength(3);

      // Validate failed test case
      const failedSuite = parsed.suites[1];
      expect(failedSuite.testCases).toHaveLength(2);
      expect(failedSuite.testCases[0].name).toBe('Failed Test Case');
      expect(failedSuite.testCases[0].status).toBe('failed');
      expect(failedSuite.testCases[0].error).toBeDefined();
      expect(failedSuite.testCases[0].error?.message).toBe('Element not found: #submit-button');

      // Validate skipped test case
      expect(failedSuite.testCases[1].name).toBe('Skipped Test Case');
      expect(failedSuite.testCases[1].status).toBe('skipped');

      // Validate errors
      expect(parsed.errors).toHaveLength(1);
      expect(parsed.errors[0].type).toBe('validation');
      expect(parsed.errors[0].message).toBe('Test environment validation failed');
    });

    it('should save JSON report to file', async () => {
      const report = await generator.generateReport(mockResult);
      const filePath = join(testOutputDir, 'test-report.json');

      await generator.saveReport(report, filePath);

      expect(existsSync(filePath)).toBe(true);

      // Read and validate saved file
      const savedContent = readFileSync(filePath, 'utf-8');
      expect(() => JSON.parse(savedContent)).not.toThrow();

      const parsed = JSON.parse(savedContent);
      expect(parsed.execution.id).toBe(mockResult.executionId);
    });

    it('should handle complex nested data structures', async () => {
      const complexResult: TestExecutionResult = {
        ...mockResult,
        suites: [
          {
            ...mockResult.suites[0],
            testCases: [
              {
                ...mockResult.suites[0].testCases[0],
                steps: mockResult.suites[0].testCases[0].steps.map((step: any) => ({
                  ...step,
                  logs: step.logs.map((log: any) => ({
                    timestamp: new Date(),
                    level: 'info',
                    message: log,
                    category: 'test-execution'
                  }))
                }))
              }
            ]
          }
        ]
      };

      const report = await generator.generateReport(complexResult);
      const parsed = JSON.parse(report);

      expect(parsed.suites[0].testCases[0].steps[0].logs).toHaveLength(1);
      expect(parsed.suites[0].testCases[0].steps[0].logs[0]).toHaveProperty('timestamp');
      expect(parsed.suites[0].testCases[0].steps[0].logs[0]).toHaveProperty('level');
      expect(parsed.suites[0].testCases[0].steps[0].logs[0]).toHaveProperty('message');
    });
  });

  describe('HTML Report Generation', () => {
    let generator: HtmlReportGenerator;

    beforeEach(() => {
      generator = new HtmlReportGenerator();
    });

    it('should generate valid HTML report with styling', async () => {
      const report = await generator.generateReport(mockResult);

      // Validate HTML structure
      expect(report).toContain('<!DOCTYPE html>');
      expect(report).toContain('<html lang="zh-CN">');
      expect(report).toContain('<head>');
      expect(report).toContain('<body>');
      expect(report).toContain('</html>');

      // Validate content
      expect(report).toContain('Chrome DevTools Validation Test Report');
      expect(report).toContain('总测试数');
      expect(report).toContain('通过');
      expect(report).toContain('失败');
      expect(report).toContain('跳过');
      expect(report).toContain('5'); // total
      expect(report).toContain('3'); // passed
      expect(report).toContain('1'); // failed
      expect(report).toContain('1'); // skipped
      expect(report).toContain('60%'); // success rate

      // Validate suite information
      expect(report).toContain('Authentication Suite');
      expect(report).toContain('Failed Test Suite');

      // Validate test case information
      expect(report).toContain('Successful Login');
      expect(report).toContain('Failed Test Case');
      expect(report).toContain('Skipped Test Case');

      // Validate CSS styling
      expect(report).toContain('<style>');
      expect(report).toContain('.container');
      expect(report).toContain('.summary-card');
      expect(report).toContain('.suite');
      expect(report).toContain('.test-case');
      expect(report).toContain('.test-step');

      // Validate metadata
      expect(report).toContain('integration');
      expect(report).toContain('Chrome');
      expect(report).toContain('120.0.0');
      expect(report).toContain('ChromeDevToolsValidationEngine');
    });

    it('should save HTML report to file', async () => {
      const report = await generator.generateReport(mockResult);
      const filePath = join(testOutputDir, 'test-report.html');

      await generator.saveReport(report, filePath);

      expect(existsSync(filePath)).toBe(true);

      // Read and validate saved file
      const savedContent = readFileSync(filePath, 'utf-8');
      expect(savedContent).toContain('<!DOCTYPE html>');
      expect(savedContent).toContain('Chrome DevTools Validation Test Report');
    });

    it('should escape HTML special characters properly', async () => {
      const resultWithSpecialChars: TestExecutionResult = {
        ...mockResult,
        suites: [
          {
            ...mockResult.suites[0],
            name: 'Test Suite with <script>alert("XSS")</script>',
            testCases: [
              {
                ...mockResult.suites[0].testCases[0],
                name: 'Test Case with "quotes" & other & special chars',
                caseId: 'case-with-chars',
                steps: [
                  {
                    ...mockResult.suites[0].testCases[0].steps[0],
                    name: 'Step with <div>HTML</div> tags',
                    logs: [{ timestamp: new Date(), level: 'info', message: 'Log with <script>alert("XSS")</script>' }]
                  }
                ]
              }
            ]
          }
        ]
      };

      const report = await generator.generateReport(resultWithSpecialChars);

      // Should contain escaped characters
      expect(report).toContain('&lt;script&gt;alert(&quot;XSS&quot;)&lt;/script&gt;');
      expect(report).toContain('&quot;quotes&quot; &amp; other &amp; special chars');
      expect(report).toContain('&lt;div&gt;HTML&lt;/div&gt;');

      // Should not contain unescaped scripts
      expect(report).not.toContain('<script>alert("XSS")</script>');
    });

    it('should include interactive elements', async () => {
      const report = await generator.generateReport(mockResult);

      // Check for interactive features
      expect(report).toContain('toggleDetails');
      expect(report).toContain('expandCollapse');
      expect(report).toContain('copyToClipboard');
      expect(report).toContain('filterResults');
      expect(report).toContain('sortResults');
    });
  });

  describe('JUnit Report Generation', () => {
    let generator: JunitReportGenerator;

    beforeEach(() => {
      generator = new JunitReportGenerator();
    });

    it('should generate valid XML report', async () => {
      const report = await generator.generateReport(mockResult);

      // Validate XML structure
      expect(report).toContain('<?xml version="1.0" encoding="UTF-8"?>');
      expect(report).toContain('<testsuites>');
      expect(report).toContain('</testsuites>');

      // Validate test suite elements
      expect(report).toContain('<testsuite ');
      expect(report).toContain('name="Authentication Suite"');
      expect(report).toContain('tests="1"');
      expect(report).toContain('failures="0"');
      expect(report).toContain('skipped="0"');
      expect(report).toContain('time="60.000"');
      expect(report).toContain('</testsuite>');

      expect(report).toContain('<testsuite ');
      expect(report).toContain('name="Failed Test Suite"');
      expect(report).toContain('tests="2"');
      expect(report).toContain('failures="1"');
      expect(report).toContain('skipped="1"');
      expect(report).toContain('time="90.000"');
      expect(report).toContain('</testsuite>');

      // Validate test case elements
      expect(report).toContain('<testcase ');
      expect(report).toContain('name="Successful Login"');
      expect(report).toContain('classname="case-1"');
      expect(report).toContain('time="30.000"');
      expect(report).toContain('</testcase>');

      expect(report).toContain('<testcase ');
      expect(report).toContain('name="Failed Test Case"');
      expect(report).toContain('name="Skipped Test Case"');
      expect(report).toContain('</testcase>');

      // Validate failure element
      expect(report).toContain('<failure message="Element not found: #submit-button">');
      expect(report).toContain('Error: Element not found');
      expect(report).toContain('</failure>');

      // Validate skipped element
      expect(report).toContain('<skipped/>');

      // Validate timestamp
      expect(report).toContain('timestamp="2025-01-01T10:00:00.000Z"');
    });

    it('should save XML report to file', async () => {
      const report = await generator.generateReport(mockResult);
      const filePath = join(testOutputDir, 'test-report.xml');

      await generator.saveReport(report, filePath);

      expect(existsSync(filePath)).toBe(true);

      // Read and validate saved file
      const savedContent = readFileSync(filePath, 'utf-8');
      expect(savedContent).toContain('<?xml version="1.0" encoding="UTF-8"?>');
      expect(savedContent).toContain('<testsuites>');
    });

    it('should escape XML special characters', async () => {
      const resultWithSpecialChars: TestExecutionResult = {
        ...mockResult,
        suites: [
          {
            ...mockResult.suites[0],
            name: 'Test Suite with "quotes" & special <chars>',
            testCases: [
              {
                ...mockResult.suites[0].testCases[0],
                name: 'Test Case with <script>alert("XSS")</script>',
                error: {
                  message: 'Error with "quotes" & other & special chars',
                  stack: 'Error stack with "quotes" & other & special chars'
                }
              }
            ]
          }
        ]
      };

      const report = await generator.generateReport(resultWithSpecialChars);

      // Should contain escaped characters
      expect(report).toContain('name="Test Suite with &quot;quotes&quot; &amp; special &lt;chars&gt;"');
      expect(report).toContain('name="Test Case with &lt;script&gt;alert(&quot;XSS&quot;)&lt;/script&gt;"');
      expect(report).toContain('message="Error with &quot;quotes&quot; &amp; other &amp; special chars"');
    });

    it('should handle empty test results', async () => {
      const emptyResult: TestExecutionResult = {
        executionId: 'empty-execution',
        startTime: new Date(),
        endTime: new Date(),
        duration: 0,
        config: mockResult.config,
        total: 0,
        passed: 0,
        failed: 0,
        skipped: 0,
        pending: 0,
        suites: [],
        errors: [],
        metadata: mockResult.metadata
      };

      const report = await generator.generateReport(emptyResult);

      expect(report).toContain('<?xml version="1.0" encoding="UTF-8"?>');
      expect(report).toContain('<testsuites>');
      expect(report).toContain('</testsuites>');
      // Should not contain any testsuite elements
      expect(report).not.toContain('<testsuite ');
    });
  });

  describe('Cross-Format Consistency', () => {
    it('should maintain data consistency across all formats', async () => {
      const jsonGenerator = new JsonReportGenerator();
      const htmlGenerator = new HtmlReportGenerator();
      const junitGenerator = new JunitReportGenerator();

      const jsonReport = await jsonGenerator.generateReport(mockResult);
      const htmlReport = await htmlGenerator.generateReport(mockResult);
      const junitReport = await junitGenerator.generateReport(mockResult);

      // Parse JSON for comparison
      const jsonData = JSON.parse(jsonReport);

      // Verify core data consistency
      expect(jsonData.summary.total).toBe(5);
      expect(htmlReport).toContain('5');
      expect(junitReport).toContain('tests="');

      expect(jsonData.summary.passed).toBe(3);
      expect(htmlReport).toContain('3');
      expect(junitReport).toContain('failures="1"'); // 5-3-1=1 failure

      expect(jsonData.summary.failed).toBe(1);
      expect(htmlReport).toContain('1');
      expect(junitReport).toContain('failures="1"');

      expect(jsonData.summary.skipped).toBe(1);
      expect(htmlReport).toContain('1');
      expect(junitReport).toContain('skipped="1"');

      // Verify suite consistency
      expect(jsonData.suites).toHaveLength(2);
      expect(htmlReport).toContain('Authentication Suite');
      expect(htmlReport).toContain('Failed Test Suite');
      expect(junitReport).toContain('name="Authentication Suite"');
      expect(junitReport).toContain('name="Failed Test Suite"');

      // Verify execution metadata consistency
      expect(jsonData.execution.id).toBe(mockResult.executionId);
      expect(htmlReport).toContain(mockResult.executionId);
      expect(junitReport).toContain('timestamp="2025-01-01T10:00:00.000Z"');
    });
  });

  describe('Error Handling', () => {
    it('should handle malformed test results gracefully', async () => {
      const malformedResult: TestExecutionResult = {
        ...mockResult,
        suites: [
          {
            ...mockResult.suites[0],
            testCases: [
              {
                ...mockResult.suites[0].testCases[0],
                // @ts-ignore - Intentionally malformed for testing
                status: 'invalid-status'
              }
            ]
          }
        ]
      };

      const jsonGenerator = new JsonReportGenerator();
      const htmlGenerator = new HtmlReportGenerator();
      const junitGenerator = new JunitReportGenerator();

      // All generators should handle malformed data gracefully
      await expect(jsonGenerator.generateReport(malformedResult)).resolves.toBeDefined();
      await expect(htmlGenerator.generateReport(malformedResult)).resolves.toBeDefined();
      await expect(junitGenerator.generateReport(malformedResult)).resolves.toBeDefined();
    });

    it('should handle missing optional fields', async () => {
      const minimalResult: TestExecutionResult = {
        executionId: 'minimal-execution',
        startTime: new Date(),
        endTime: new Date(),
        duration: 1000,
        config: mockResult.config,
        total: 1,
        passed: 1,
        failed: 0,
        skipped: 0,
        pending: 0,
        suites: [
          {
            suiteId: 'minimal-suite',
            name: 'Minimal Suite',
            status: 'passed',
            startTime: new Date(),
            endTime: new Date(),
            duration: 1000,
            testCases: [
              {
                caseId: 'minimal-case',
                name: 'Minimal Case',
                status: 'passed',
                startTime: new Date(),
                endTime: new Date(),
                duration: 1000,
                steps: []
              }
            ]
          }
        ],
        errors: [],
        metadata: mockResult.metadata
      };

      const generators = [
        new JsonReportGenerator(),
        new HtmlReportGenerator(),
        new JunitReportGenerator()
      ];

      for (const generator of generators) {
        const report = await generator.generateReport(minimalResult);
        expect(report).toBeDefined();
        expect(report.length).toBeGreaterThan(0);
      }
    });
  });
});