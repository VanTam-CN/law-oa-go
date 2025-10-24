import { JsonReportGenerator, HtmlReportGenerator, JunitReportGenerator } from '../../src/core/test-report-generator';
import { TestExecutionResult } from '../../src/types/test-engine-types';

describe('Test Report Generators', () => {
  let mockResult: TestExecutionResult;

  beforeEach(() => {
    mockResult = {
      executionId: 'test-execution-123',
      startTime: new Date('2025-01-01T10:00:00Z'),
      endTime: new Date('2025-01-01T10:01:00Z'),
      duration: 60000,
      config: {
        generateReport: true,
        reportFormat: 'json',
        testEnvironment: 'development'
      },
      total: 3,
      passed: 2,
      failed: 1,
      skipped: 0,
      pending: 0,
      suites: [
        {
          suiteId: 'suite-1',
          name: 'Login Suite',
          description: 'Test login functionality',
          status: 'failed',
          startTime: new Date('2025-01-01T10:00:00Z'),
          endTime: new Date('2025-01-01T10:01:00Z'),
          duration: 60000,
          testCases: [
            {
              caseId: 'case-1',
              name: 'Successful Login',
              description: 'Test successful login',
              status: 'passed',
              startTime: new Date('2025-01-01T10:00:00Z'),
              endTime: new Date('2025-01-01T10:00:30Z'),
              duration: 30000,
              steps: [
                {
                  stepId: 'step-1',
                  name: 'Navigate to login page',
                  status: 'passed',
                  duration: 10000,
                  startTime: new Date('2025-01-01T10:00:00Z'),
                  endTime: new Date('2025-01-01T10:00:10Z'),
                  logs: []
                }
              ]
            },
            {
              caseId: 'case-2',
              name: 'Failed Login',
              description: 'Test failed login',
              status: 'failed',
              startTime: new Date('2025-01-01T10:00:30Z'),
              endTime: new Date('2025-01-01T10:01:00Z'),
              duration: 30000,
              steps: [
                {
                  stepId: 'step-2',
                  name: 'Enter invalid credentials',
                  status: 'failed',
                  duration: 15000,
                  startTime: new Date('2025-01-01T10:00:30Z'),
                  endTime: new Date('2025-01-01T10:00:45Z'),
                  error: {
                    message: 'Element not found: #error-message',
                    stack: 'Error: Element not found\n    at TestCase.execute'
                  },
                  logs: []
                }
              ],
              error: {
                message: 'Test failed: Element not found',
                stack: 'Error: Test failed\n    at TestCase.execute'
              }
            }
          ]
        }
      ],
      errors: [],
      metadata: {
        environment: 'development',
        browser: 'Chrome',
        version: '120.0.0',
        executor: 'ChromeDevToolsValidationEngine'
      }
    };
  });

  describe('JsonReportGenerator', () => {
    let generator: JsonReportGenerator;

    beforeEach(() => {
      generator = new JsonReportGenerator();
    });

    describe('generateReport', () => {
      it('should generate valid JSON report', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toBeDefined();
        expect(() => JSON.parse(report)).not.toThrow();

        const parsed = JSON.parse(report);
        expect(parsed.execution.id).toBe(mockResult.executionId);
        expect(parsed.summary.total).toBe(mockResult.total);
        expect(parsed.summary.passed).toBe(mockResult.passed);
        expect(parsed.summary.failed).toBe(mockResult.failed);
        expect(parsed.summary.successRate).toBe(67); // Math.round(2/3 * 100)
      });

      it('should include execution metadata', async () => {
        const report = await generator.generateReport(mockResult);
        const parsed = JSON.parse(report);

        expect(parsed.execution.environment).toBe(mockResult.metadata.environment);
        expect(parsed.execution.browser).toBe(mockResult.metadata.browser);
        expect(parsed.execution.version).toBe(mockResult.metadata.version);
        expect(parsed.execution.executor).toBe(mockResult.metadata.executor);
      });

      it('should include suite details', async () => {
        const report = await generator.generateReport(mockResult);
        const parsed = JSON.parse(report);

        expect(parsed.suites).toHaveLength(1);
        expect(parsed.suites[0].name).toBe('Login Suite');
        expect(parsed.suites[0].testCases).toHaveLength(2);
        expect(parsed.suites[0].summary.total).toBe(2);
        expect(parsed.suites[0].summary.passed).toBe(1);
        expect(parsed.suites[0].summary.failed).toBe(1);
      });

      it('should include test case details', async () => {
        const report = await generator.generateReport(mockResult);
        const parsed = JSON.parse(report);

        const testCase = parsed.suites[0].testCases[0];
        expect(testCase.name).toBe('Successful Login');
        expect(testCase.status).toBe('passed');
        expect(testCase.steps).toHaveLength(1);
        expect(testCase.steps[0].name).toBe('Navigate to login page');
      });

      it('should include error information', async () => {
        const report = await generator.generateReport(mockResult);
        const parsed = JSON.parse(report);

        const failedCase = parsed.suites[0].testCases[1];
        expect(failedCase.error).toBeDefined();
        expect(failedCase.error?.message).toBe('Test failed: Element not found');
        expect(failedCase.steps[0].error).toBeDefined();
        expect(failedCase.steps[0].error?.message).toBe('Element not found: #error-message');
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
        const parsed = JSON.parse(report);

        expect(parsed.summary.total).toBe(0);
        expect(parsed.summary.successRate).toBe(0);
        expect(parsed.suites).toHaveLength(0);
      });
    });

    describe('saveReport', () => {
      it('should handle save operation', async () => {
        const report = await generator.generateReport(mockResult);

        // Mock implementation would write to file system
        await expect(generator.saveReport(report, 'test-report.json')).resolves.not.toThrow();
      });
    });
  });

  describe('HtmlReportGenerator', () => {
    let generator: HtmlReportGenerator;

    beforeEach(() => {
      generator = new HtmlReportGenerator();
    });

    describe('generateReport', () => {
      it('should generate valid HTML report', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toBeDefined();
        expect(report).toContain('<!DOCTYPE html>');
        expect(report).toContain('<html lang="zh-CN">');
        expect(report).toContain('<head>');
        expect(report).toContain('<body>');
        expect(report).toContain('</html>');
      });

      it('should include test summary in HTML', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toContain('总测试数');
        expect(report).toContain('通过');
        expect(report).toContain('失败');
        expect(report).toContain('跳过');
        expect(report).toContain('3'); // total
        expect(report).toContain('2'); // passed
        expect(report).toContain('1'); // failed
      });

      it('should include success rate in HTML', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toContain('67%'); // success rate
      });

      it('should include suite information', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toContain('Login Suite');
        expect(report).toContain('Test login functionality');
      });

      it('should include test case information', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toContain('Successful Login');
        expect(report).toContain('Failed Login');
      });

      it('should include CSS styling', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toContain('<style>');
        expect(report).toContain('.container');
        expect(report).toContain('.summary-card');
        expect(report).toContain('.suite');
      });

      it('should include execution metadata', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toContain('development');
        expect(report).toContain('Chrome');
        expect(report).toContain('120.0.0');
        expect(report).toContain('ChromeDevToolsValidationEngine');
      });

      it('should escape HTML special characters', async () => {
        const resultWithSpecialChars: TestExecutionResult = {
          ...mockResult,
          suites: [{
            ...mockResult.suites[0],
            name: 'Test Suite with <script>alert("XSS")</script>',
            testCases: [{
              ...mockResult.suites[0].testCases[0],
              name: 'Test Case with "quotes" & other & special chars',
              caseId: mockResult.suites[0].testCases[0].caseId || 'case-with-chars'
            }]
          }]
        };

        const report = await generator.generateReport(resultWithSpecialChars);

        // Should contain escaped characters
        expect(report).toContain('&lt;script&gt;alert(&quot;XSS&quot;)&lt;/script&gt;');
        expect(report).toContain('&quot;quotes&quot; &amp; other &amp; special chars');
      });
    });

    describe('saveReport', () => {
      it('should handle save operation', async () => {
        const report = await generator.generateReport(mockResult);

        await expect(generator.saveReport(report, 'test-report.html')).resolves.not.toThrow();
      });
    });
  });

  describe('JunitReportGenerator', () => {
    let generator: JunitReportGenerator;

    beforeEach(() => {
      generator = new JunitReportGenerator();
    });

    describe('generateReport', () => {
      it('should generate valid XML report', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toBeDefined();
        expect(report).toContain('<?xml version="1.0" encoding="UTF-8"?>');
        expect(report).toContain('<testsuites>');
        expect(report).toContain('</testsuites>');
      });

      it('should include testsuite element', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toContain('<testsuite ');
        expect(report).toContain('name="Login Suite"');
        expect(report).toContain('tests="2"');
        expect(report).toContain('failures="1"');
        expect(report).toContain('skipped="0"');
        expect(report).toContain('time="60.000"');
        expect(report).toContain('</testsuite>');
      });

      it('should include testcase elements', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toContain('<testcase ');
        expect(report).toContain('name="Successful Login"');
        expect(report).toContain('classname="case-1"');
        expect(report).toContain('time="30.000"');
        expect(report).toContain('</testcase>');
      });

      it('should include failure element for failed tests', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toContain('<failure message="Test failed: Element not found">');
        expect(report).toContain('Error: Test failed');
        expect(report).toContain('</failure>');
      });

      it('should escape XML special characters', async () => {
        const resultWithSpecialChars: TestExecutionResult = {
          ...mockResult,
          suites: [{
            ...mockResult.suites[0],
            name: 'Test Suite with "quotes" & special chars',
            testCases: [{
              ...mockResult.suites[0].testCases[0],
              name: 'Test Case with <script>alert("XSS")</script>',
              error: {
                message: 'Error with "quotes" & other & special chars',
                stack: 'Error stack with "quotes" & other & special chars'
              }
            }]
          }]
        };

        const report = await generator.generateReport(resultWithSpecialChars);

        // Should contain escaped characters
        expect(report).toContain('name="Test Suite with &quot;quotes&quot; &amp; special chars"');
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

      it('should include timestamp', async () => {
        const report = await generator.generateReport(mockResult);

        expect(report).toContain('timestamp="2025-01-01T10:00:00.000Z"');
      });
    });

    describe('saveReport', () => {
      it('should handle save operation', async () => {
        const report = await generator.generateReport(mockResult);

        await expect(generator.saveReport(report, 'test-report.xml')).resolves.not.toThrow();
      });
    });
  });

  describe('Report Generation Edge Cases', () => {
    it('should handle results with errors', async () => {
      const resultWithError: TestExecutionResult = {
        ...mockResult,
        errors: [
          {
            timestamp: new Date(),
            type: 'system',
            message: 'System error occurred',
            stack: 'Error: System error\n    at System.execute',
            component: 'TestRunner'
          }
        ]
      };

      const jsonGenerator = new JsonReportGenerator();
      const report = await jsonGenerator.generateReport(resultWithError);
      const parsed = JSON.parse(report);

      expect(parsed.errors).toHaveLength(1);
      expect(parsed.errors[0].type).toBe('system');
      expect(parsed.errors[0].message).toBe('System error occurred');
    });

    it('should handle results with skipped tests', async () => {
      const resultWithSkipped: TestExecutionResult = {
        ...mockResult,
        total: 4,
        passed: 2,
        failed: 1,
        skipped: 1,
        suites: [
          {
            ...mockResult.suites[0],
            testCases: [
              ...mockResult.suites[0].testCases,
              {
                caseId: 'case-3',
                name: 'Skipped Test',
                description: 'Skipped test case',
                status: 'skipped',
                startTime: new Date(),
                endTime: new Date(),
                duration: 0,
                steps: []
              }
            ]
          }
        ]
      };

      const jsonGenerator = new JsonReportGenerator();
      const report = await jsonGenerator.generateReport(resultWithSkipped);
      const parsed = JSON.parse(report);

      expect(parsed.summary.skipped).toBe(1);
      expect(parsed.summary.successRate).toBe(50); // Math.round(2/4 * 100)
    });
  });
});