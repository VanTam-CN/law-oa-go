/**
 * 测试报告生成器实现
 */

import {
  TestExecutionResult,
  TestReportGenerator,
  TestSuiteResult,
  TestCaseResult,
  TestStepResult
} from '../types/test-engine-types';
import { Logger } from '../core/logger';

/**
 * JSON报告生成器
 */
export class JsonReportGenerator implements TestReportGenerator {
  private logger: Logger;

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('JsonReportGenerator');
  }

  override async generateReport(result: TestExecutionResult): Promise<string> {
    try {
      const report = {
        execution: {
          id: result.executionId,
          startTime: result.startTime,
          endTime: result.endTime,
          duration: result.duration,
          environment: result.metadata.environment,
          browser: result.metadata.browser,
          version: result.metadata.version,
          executor: result.metadata.executor
        },
        summary: {
          total: result.total,
          passed: result.passed,
          failed: result.failed,
          skipped: result.skipped,
          pending: result.pending,
          successRate: result.total > 0 ? Math.round((result.passed / result.total) * 100) : 0
        },
        suites: result.suites.map(this.formatSuite.bind(this)),
        errors: result.errors.map(error => ({
          timestamp: error.timestamp,
          type: error.type,
          message: error.message,
          component: error.component
        })),
        generatedAt: new Date().toISOString()
      };

      return JSON.stringify(report, null, 2);
    } catch (error) {
      this.logger.error('生成JSON报告失败', {
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  override async saveReport(report: string, filePath: string): Promise<void> {
    try {
      // 在实际实现中，这里会写入文件系统
      this.logger.info('保存JSON报告', { filePath, length: report.length });
      // await fs.writeFile(filePath, report, 'utf-8');
    } catch (error) {
      this.logger.error('保存JSON报告失败', {
        filePath,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  private formatSuite(suite: TestSuiteResult): any {
    return {
      id: suite.suiteId,
      name: suite.name,
      description: suite.description,
      status: suite.status,
      startTime: suite.startTime,
      endTime: suite.endTime,
      duration: suite.duration,
      summary: {
        total: suite.testCases.length,
        passed: suite.testCases.filter(tc => tc.status === 'passed').length,
        failed: suite.testCases.filter(tc => tc.status === 'failed').length,
        skipped: suite.testCases.filter(tc => tc.status === 'skipped').length,
        pending: suite.testCases.filter(tc => tc.status === 'pending').length
      },
      testCases: suite.testCases.map(this.formatTestCase.bind(this)),
      setupError: suite.setupError,
      teardownError: suite.teardownError
    };
  }

  private formatTestCase(testCase: TestCaseResult): any {
    return {
      id: testCase.caseId,
      name: testCase.name,
      description: testCase.description,
      status: testCase.status,
      startTime: testCase.startTime,
      endTime: testCase.endTime,
      duration: testCase.duration,
      retryCount: testCase.retryCount,
      steps: testCase.steps.map(this.formatTestStep.bind(this)),
      error: testCase.error,
      setupError: testCase.setupError,
      teardownError: testCase.teardownError
    };
  }

  private formatTestStep(step: TestStepResult): any {
    return {
      id: step.stepId,
      name: step.name,
      status: step.status,
      duration: step.duration,
      startTime: step.startTime,
      endTime: step.endTime,
      error: step.error,
      logs: step.logs.map(log => ({
        timestamp: log.timestamp,
        level: log.level,
        message: log.message,
        context: log.context
      }))
    };
  }
}

/**
 * HTML报告生成器
 */
export class HtmlReportGenerator implements TestReportGenerator {
  private logger: Logger;

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('HtmlReportGenerator');
  }

  override async generateReport(result: TestExecutionResult): Promise<string> {
    try {
      const html = this.generateHtmlReport(result);
      return html;
    } catch (error) {
      this.logger.error('生成HTML报告失败', {
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  override async saveReport(report: string, filePath: string): Promise<void> {
    try {
      // 在实际实现中，这里会写入文件系统
      this.logger.info('保存HTML报告', { filePath, length: report.length });
      // await fs.writeFile(filePath, report, 'utf-8');
    } catch (error) {
      this.logger.error('保存HTML报告失败', {
        filePath,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  private generateHtmlReport(result: TestExecutionResult): string {
    const successRate = result.total > 0 ? Math.round((result.passed / result.total) * 100) : 0;

    return `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>测试报告 - ${result.executionId}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            padding: 30px;
        }
        .header {
            border-bottom: 2px solid #e0e0e0;
            padding-bottom: 20px;
            margin-bottom: 30px;
        }
        .title {
            font-size: 28px;
            color: #333;
            margin: 0 0 10px 0;
        }
        .subtitle {
            color: #666;
            margin: 0;
        }
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .summary-card {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
        }
        .summary-value {
            font-size: 32px;
            font-weight: bold;
            margin: 10px 0 5px 0;
        }
        .summary-label {
            color: #666;
            font-size: 14px;
        }
        .success { color: #28a745; }
        .failure { color: #dc3545; }
        .warning { color: #ffc107; }
        .info { color: #17a2b8; }
        .suite {
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            margin-bottom: 20px;
            overflow: hidden;
        }
        .suite-header {
            background: #f8f9fa;
            padding: 15px 20px;
            border-bottom: 1px solid #e0e0e0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .suite-title {
            font-size: 18px;
            font-weight: 600;
            margin: 0;
        }
        .suite-status {
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
        }
        .status-passed { background: #d4edda; color: #155724; }
        .status-failed { background: #f8d7da; color: #721c24; }
        .status-skipped { background: #fff3cd; color: #856404; }
        .status-pending { background: #d1ecf1; color: #0c5460; }
        .test-case {
            border-bottom: 1px solid #f0f0f0;
            padding: 15px 20px;
        }
        .test-case:last-child {
            border-bottom: none;
        }
        .test-case-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        .test-case-title {
            font-weight: 600;
            margin: 0;
        }
        .error-message {
            background: #f8d7da;
            color: #721c24;
            padding: 10px;
            border-radius: 4px;
            margin-top: 10px;
            font-family: monospace;
            font-size: 12px;
        }
        .step {
            margin-left: 20px;
            padding: 8px 0;
            border-left: 2px solid #e0e0e0;
            padding-left: 15px;
        }
        .step-name {
            font-size: 14px;
            margin: 0 0 5px 0;
        }
        .step-duration {
            font-size: 12px;
            color: #666;
        }
        .metadata {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 15px;
            margin-top: 30px;
            font-size: 14px;
            color: #666;
        }
        .progress-bar {
            background: #e0e0e0;
            border-radius: 10px;
            height: 20px;
            overflow: hidden;
            margin: 10px 0;
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #28a745 0%, #20c997 100%);
            transition: width 0.3s ease;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">测试执行报告</h1>
            <p class="subtitle">执行ID: ${result.executionId}</p>
        </div>

        <div class="summary">
            <div class="summary-card">
                <div class="summary-label">总测试数</div>
                <div class="summary-value">${result.total}</div>
            </div>
            <div class="summary-card">
                <div class="summary-label success">通过</div>
                <div class="summary-value success">${result.passed}</div>
            </div>
            <div class="summary-card">
                <div class="summary-label failure">失败</div>
                <div class="summary-value failure">${result.failed}</div>
            </div>
            <div class="summary-card">
                <div class="summary-label warning">跳过</div>
                <div class="summary-value warning">${result.skipped}</div>
            </div>
            <div class="summary-card">
                <div class="summary-label">成功率</div>
                <div class="summary-value ${successRate >= 80 ? 'success' : successRate >= 60 ? 'warning' : 'failure'}">${successRate}%</div>
            </div>
            <div class="summary-card">
                <div class="summary-label info">执行时间</div>
                <div class="summary-value info">${Math.round(result.duration / 1000)}s</div>
            </div>
        </div>

        <div class="progress-bar">
            <div class="progress-fill" style="width: ${successRate}%"></div>
        </div>

        ${result.suites.map(suite => this.generateSuiteHtml(suite)).join('')}

        <div class="metadata">
            <strong>执行信息:</strong><br>
            环境: ${result.metadata.environment} |
            浏览器: ${result.metadata.browser} ${result.metadata.version} |
            执行器: ${result.metadata.executor}<br>
            开始时间: ${result.startTime.toLocaleString()} |
            结束时间: ${result.endTime.toLocaleString()}<br>
            报告生成时间: ${new Date().toLocaleString()}
        </div>
    </div>
</body>
</html>`;
  }

  private generateSuiteHtml(suite: TestSuiteResult): string {
    const caseCount = suite.testCases.length;
    const passedCount = suite.testCases.filter(tc => tc.status === 'passed').length;
    const failedCount = suite.testCases.filter(tc => tc.status === 'failed').length;

    return `
        <div class="suite">
            <div class="suite-header">
                <h3 class="suite-title">${suite.name}</h3>
                <span class="suite-status status-${suite.status}">${this.getStatusText(suite.status)}</span>
            </div>
            <div style="padding: 20px;">
                <div style="margin-bottom: 15px; color: #666; font-size: 14px;">
                    测试用例: ${caseCount} | 通过: ${passedCount} | 失败: ${failedCount}
                </div>
                ${suite.testCases.map(testCase => this.generateTestCaseHtml(testCase)).join('')}
            </div>
        </div>`;
  }

  private generateTestCaseHtml(testCase: TestCaseResult): string {
    return `
        <div class="test-case">
            <div class="test-case-header">
                <h4 class="test-case-title">${testCase.name}</h4>
                <span class="suite-status status-${testCase.status}">${this.getStatusText(testCase.status)}</span>
            </div>
            ${testCase.description ? `<div style="color: #666; margin-bottom: 10px;">${testCase.description}</div>` : ''}
            ${testCase.error ? `<div class="error-message">${testCase.error.message}</div>` : ''}
            ${testCase.steps.length > 0 ? `
                <div style="margin-top: 10px;">
                    <strong>执行步骤:</strong>
                    ${testCase.steps.map(step => this.generateStepHtml(step)).join('')}
                </div>
            ` : ''}
        </div>`;
  }

  private generateStepHtml(step: TestStepResult): string {
    return `
        <div class="step">
            <div class="step-name">${step.name}</div>
            <div class="step-duration">${step.duration}ms - ${this.getStatusText(step.status)}</div>
            ${step.error ? `<div class="error-message">${step.error.message}</div>` : ''}
        </div>`;
  }

  private getStatusText(status: string): string {
    const statusMap: Record<string, string> = {
      'passed': '通过',
      'failed': '失败',
      'skipped': '跳过',
      'pending': '等待'
    };
    return statusMap[status] || status;
  }
}

/**
 * JUnit XML报告生成器
 */
export class JunitReportGenerator implements TestReportGenerator {
  private logger: Logger;

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('JunitReportGenerator');
  }

  override async generateReport(result: TestExecutionResult): Promise<string> {
    try {
      let xml = '<?xml version="1.0" encoding="UTF-8"?>\n';
      xml += '<testsuites>\n';

      for (const suite of result.suites) {
        xml += this.generateSuiteXml(suite);
      }

      xml += '</testsuites>';
      return xml;
    } catch (error) {
      this.logger.error('生成JUnit报告失败', {
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  override async saveReport(report: string, filePath: string): Promise<void> {
    try {
      // 在实际实现中，这里会写入文件系统
      this.logger.info('保存JUnit报告', { filePath, length: report.length });
      // await fs.writeFile(filePath, report, 'utf-8');
    } catch (error) {
      this.logger.error('保存JUnit报告失败', {
        filePath,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  private generateSuiteXml(suite: TestSuiteResult): string {
    const attributes = [
      `name="${this.escapeXml(suite.name)}"`,
      `tests="${suite.testCases.length}"`,
      `failures="${suite.testCases.filter(tc => tc.status === 'failed').length}"`,
      `skipped="${suite.testCases.filter(tc => tc.status === 'skipped').length}"`,
      `time="${(suite.duration / 1000).toFixed(3)}"`,
      `timestamp="${suite.startTime.toISOString()}"`
    ].join(' ');

    let xml = `  <testsuite ${attributes}>\n`;

    for (const testCase of suite.testCases) {
      xml += this.generateTestCaseXml(testCase);
    }

    xml += '  </testsuite>\n';
    return xml;
  }

  private generateTestCaseXml(testCase: TestCaseResult): string {
    const attributes = [
      `name="${this.escapeXml(testCase.name)}"`,
      `classname="${this.escapeXml(testCase.caseId)}"`,
      `time="${(testCase.duration / 1000).toFixed(3)}"`
    ].join(' ');

    let xml = `    <testcase ${attributes}>\n`;

    if (testCase.status === 'failed' && testCase.error) {
      xml += `      <failure message="${this.escapeXml(testCase.error.message)}">\n`;
      if (testCase.error.stack) {
        xml += `        ${this.escapeXml(testCase.error.stack)}\n`;
      }
      xml += '      </failure>\n';
    }

    if (testCase.status === 'skipped') {
      xml += '      <skipped/>\n';
    }

    xml += '    </testcase>\n';
    return xml;
  }

  private escapeXml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }
}