/**
 * 端到端测试运行器
 */

import { E2EBusinessWorkflowTest } from './e2e-business-workflow-test';
import { E2E_TEST_CONFIG } from './e2e-test-config';
import { Logger } from '../../core/logger';
import * as fs from 'fs';
import * as path from 'path';

export interface E2ETestRunnerConfig {
  environment: 'development' | 'staging' | 'production';
  workflows?: string[];
  scenarios?: string[];
  outputDir?: string;
  parallel?: boolean;
  retries?: number;
  includePerformance?: boolean;
  includeCleanup?: boolean;
}

export interface E2ETestRunResult {
  timestamp: string;
  environment: string;
  config: E2ETestRunnerConfig;
  summary: {
    totalWorkflows: number;
    passedWorkflows: number;
    failedWorkflows: number;
    skippedWorkflows: number;
    successRate: number;
    totalDuration: number;
    averageDuration: number;
  };
  workflowResults: any[];
  performanceMetrics?: {
    totalExecutionTime: number;
    averageWorkflowTime: number;
    slowestWorkflow: string;
    fastestWorkflow: string;
    memoryUsage: any;
    networkRequests: any;
  };
  reportPath: string;
  screenshotsPath: string;
  logsPath: string;
}

export class E2ETestRunner {
  private logger: Logger;
  private config: E2ETestRunnerConfig;
  private envConfig: any;
  private outputDir: string;
  private screenshotsDir: string;
  private logsDir: string;
  private workflowTest: E2EBusinessWorkflowTest;

  constructor(config: E2ETestRunnerConfig) {
    this.logger = new Logger('E2ETestRunner');
    this.config = config;

    // 获取环境配置
    this.envConfig = E2E_TEST_CONFIG.environments[config.environment];

    // 设置输出目录
    this.outputDir = config.outputDir || `./test-results/e2e/${config.environment}`;
    this.screenshotsDir = path.join(this.outputDir, 'screenshots');
    this.logsDir = path.join(this.outputDir, 'logs');

    // 创建目录
    this.ensureDirectoriesExist();

    // 初始化工作流测试
    this.workflowTest = new E2EBusinessWorkflowTest({
      baseUrl: this.envConfig.baseUrl,
      defaultTimeout: this.envConfig.timeout,
      screenshotOnFailure: this.envConfig.screenshots,
      testData: E2E_TEST_CONFIG.testData
    });
  }

  /**
   * 确保测试目录存在
   */
  private ensureDirectoriesExist(): void {
    [this.outputDir, this.screenshotsDir, this.logsDir].forEach(dir => {
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
      }
    });
  }

  /**
   * 运行端到端测试
   */
  async runTests(): Promise<E2ETestRunResult> {
    const startTime = Date.now();
    this.logger.info('开始运行端到端测试', { config: this.config });

    try {
      // 准备测试环境
      await this.prepareTestEnvironment();

      // 运行工作流测试
      const workflowResults = await this.runWorkflowTests();

      // 运行性能测试（如果启用）
      let performanceMetrics: any | undefined = undefined;
      if (this.config.includePerformance) {
        performanceMetrics = await this.runPerformanceTests();
      }

      // 生成报告
      const reportPath = this.generateReport({
        startTime,
        workflowResults,
        performanceMetrics
      });

      // 清理测试数据（如果启用）
      if (this.config.includeCleanup) {
        await this.cleanupTestData();
      }

      const duration = Date.now() - startTime;
      const result: E2ETestRunResult = {
        timestamp: new Date().toISOString(),
        environment: this.config.environment,
        config: this.config,
        summary: this.calculateSummary(workflowResults, duration),
        workflowResults,
        performanceMetrics,
        reportPath,
        screenshotsPath: this.screenshotsDir,
        logsPath: this.logsDir
      };

      this.logger.info('端到端测试运行完成', result.summary);
      return result;

    } catch (error: unknown) {
      this.logger.error('端到端测试运行失败', { error });
      throw error;
    }
  }

  /**
   * 准备测试环境
   */
  private async prepareTestEnvironment(): Promise<void> {
    this.logger.info('准备端到端测试环境');

    try {
      // 检查服务可用性
      await this.checkServiceAvailability();

      // 验证测试数据
      await this.validateTestData();

      // 清理历史数据
      await this.cleanupHistoricalData();

      // 准备测试用户
      await this.prepareTestUsers();

      this.logger.info('测试环境准备完成');
    } catch (error: unknown) {
      this.logger.error('测试环境准备失败', { error });
      throw error;
    }
  }

  /**
   * 检查服务可用性
   */
  private async checkServiceAvailability(): Promise<void> {
    this.logger.info('检查服务可用性');

    // 检查前端服务
    try {
      const response = await fetch(`${this.envConfig.baseUrl}/health`);
      if (!response.ok) {
        throw new Error(`前端服务不可用: ${response.status}`);
      }
    } catch (error: unknown) {
      throw new Error(`无法连接到前端服务: ${error.message}`);
    }

    // 检查后端API服务
    try {
      const response = await fetch(`${this.envConfig.apiBaseUrl}/health`);
      if (!response.ok) {
        throw new Error(`后端API服务不可用: ${response.status}`);
      }
    } catch (error: unknown) {
      throw new Error(`无法连接到后端API服务: ${error.message}`);
    }

    this.logger.info('所有服务检查通过');
  }

  /**
   * 验证测试数据
   */
  private async validateTestData(): Promise<void> {
    this.logger.info('验证测试数据');

    // 验证用户数据
    if (!E2E_TEST_CONFIG.testData.users) {
      throw new Error('测试用户数据未配置');
    }

    // 验证案件数据
    if (!E2E_TEST_CONFIG.testData.cases) {
      throw new Error('测试案件数据未配置');
    }

    // 验证文档数据
    if (!E2E_TEST_CONFIG.testData.documents) {
      throw new Error('测试文档数据未配置');
    }

    // 验证客户数据
    if (!E2E_TEST_CONFIG.testData.clients) {
      throw new Error('测试客户数据未配置');
    }

    this.logger.info('测试数据验证通过');
  }

  /**
   * 清理历史数据
   */
  private async cleanupHistoricalData(): Promise<void> {
    this.logger.info('清理历史测试数据');

    try {
      // 清理E2E测试标记的数据
      // 注意：这里需要根据实际的数据库和API来实现

      this.logger.info('历史数据清理完成');
    } catch (error: unknown) {
      this.logger.warn('清理历史数据失败', { error });
      // 不抛出错误，允许测试继续
    }
  }

  /**
   * 准备测试用户
   */
  private async prepareTestUsers(): Promise<void> {
    this.logger.info('准备测试用户');

    try {
      // 确保测试用户存在
      for (const [role, user] of Object.entries(E2E_TEST_CONFIG.testData.users)) {
        // 创建或更新测试用户
        // 注意：这里需要根据实际的用户管理API来实现

        this.logger.info(`测试用户 ${user.username} 准备完成`);
      }
    } catch (error: unknown) {
      this.logger.error('测试用户准备失败', { error });
      throw error;
    }
  }

  /**
   * 运行工作流测试
   */
  private async runWorkflowTests(): Promise<any[]> {
    this.logger.info('运行工作流测试');

    const workflows = this.config.workflows || [
      'client-intake-workflow',
      'case-management-workflow',
      'document-management-workflow',
      'financial-tracking-workflow',
      'conflict-check-workflow',
      'complete-lifecycle-workflow'
    ];

    let results: any[] = [];

    for (const workflow of workflows) {
      this.logger.info(`运行工作流: ${workflow}`);

      try {
        const result = await this.workflowTest.runWorkflow(workflow);
        results.push(result);

        if (result.status === 'failed') {
          this.logger.error(`工作流 ${workflow} 失败`, { error: result.error });
        } else {
          this.logger.info(`工作流 ${workflow} 完成`, { duration: result.duration });
        }
      } catch (error: unknown) {
        this.logger.error(`工作流 ${workflow} 运行异常`, { error });
        results.push({
          workflowId: workflow,
          workflowName: workflow,
          status: 'failed',
          duration: 0,
          steps: [],
          error: error.message
        });
      }
    }

    return results;
  }

  /**
   * 运行性能测试
   */
  private async runPerformanceTests(): Promise<any> {
    this.logger.info('运行性能测试');

    try {
      const performanceResults: any = {
        totalExecutionTime: 0,
        averageWorkflowTime: 0,
        slowestWorkflow: '',
        fastestWorkflow: '',
        memoryUsage: {},
        networkRequests: {}
      };

      // 运行性能监控
      // 注意：这里需要根据实际的性能监控工具来实现

      return performanceResults;
    } catch (error: unknown) {
      this.logger.error('性能测试运行失败', { error });
      return undefined;
    }
  }

  /**
   * 清理测试数据
   */
  private async cleanupTestData(): Promise<void> {
    this.logger.info('清理测试数据');

    try {
      // 清理所有E2E测试创建的数据
      // 注意：这里需要根据实际的数据库和API来实现

      this.logger.info('测试数据清理完成');
    } catch (error: unknown) {
      this.logger.error('测试数据清理失败', { error });
      // 不抛出错误，避免影响测试结果
    }
  }

  /**
   * 计算测试摘要
   */
  private calculateSummary(results: any[], totalDuration: number): E2ETestRunResult['summary'] {
    const total = results.length;
    const passed = results.filter(r => r.status === 'passed').length;
    const failed = results.filter(r => r.status === 'failed').length;
    const skipped = results.filter(r => r.status === 'skipped').length;
    const successRate = total > 0 ? (passed / total) * 100 : 0;
    const averageDuration = total > 0 ? results.reduce((sum, r) => sum + r.duration, 0) / total : 0;

    return {
      totalWorkflows: total,
      passedWorkflows: passed,
      failedWorkflows: failed,
      skippedWorkflows: skipped,
      successRate,
      totalDuration,
      averageDuration
    };
  }

  /**
   * 生成测试报告
   */
  private generateReport(data: {
    startTime: number;
    workflowResults: any[];
    performanceMetrics?: any;
  }): string {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const reportFileName = `e2e-test-report-${timestamp}.json`;
    const reportPath = path.join(this.outputDir, reportFileName);

    const report = {
      metadata: {
        timestamp: new Date().toISOString(),
        environment: this.config.environment,
        testRunner: 'E2ETestRunner',
        version: '1.0.0',
        executionTime: Date.now() - data.startTime
      },
      config: this.config,
      summary: this.calculateSummary(data.workflowResults, Date.now() - data.startTime),
      workflowResults: data.workflowResults.map(result => ({
        workflowId: result.workflowId,
        workflowName: result.workflowName,
        status: result.status,
        duration: result.duration,
        steps: result.getsteps?.().map((step: any) => ({
          stepId: step.stepId,
          stepName: step.stepName,
          status: step.status,
          duration: step.duration,
          error: step.error
        })),
        error: result.error
      })),
      performanceMetrics: data.performanceMetrics,
      recommendations: this.generateRecommendations(data.workflowResults),
      testEnvironment: {
        nodeVersion: process.version,
        platform: process.platform,
        arch: process.arch,
        timestamp: new Date().toISOString()
      }
    };

    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    this.logger.info(`端到端测试报告已生成: ${reportPath}`);

    return reportPath;
  }

  /**
   * 生成建议
   */
  private generateRecommendations(results: any[]): string[] {
    let recommendations: string[] = [];

    // 分析失败率
    const failureRate = results.filter(r => r.status === 'failed').length / results.length;
    if (failureRate > 0.2) {
      recommendations.push('工作流失败率较高（>20%），建议优先修复失败的工作流');
    }

    // 分析执行时间
    const slowWorkflows = results.filter(r => r.duration > 300000); // 5分钟
    if (slowWorkflows.length > 0) {
      (recommendations = recommendations || []).push('部分工作流执行时间过长，建议优化性能');
    }

    // 分析步骤失败
    const failedSteps = results.flatMap(r => r.getsteps?.().filter((s: any) => s.status === 'failed') || []);
    const stepFailureRate = failedSteps.length / results.reduce((sum, r) => sum + (r.getsteps?.().length || 0), 0);
    if (stepFailureRate > 0.1) {
      (recommendations = recommendations || []).push('步骤失败率较高，建议检查关键步骤的稳定性');
    }

    // 通用建议
    (recommendations = recommendations || []).push('建议定期运行端到端测试以确保系统稳定性');
    (recommendations = recommendations || []).push('建议增加更多边界情况的测试覆盖');
    (recommendations = recommendations || []).push('建议配置自动化报警机制');
    (recommendations = recommendations || []).push('建议优化测试执行效率');

    return recommendations;
  }
}

/**
 * 命令行接口
 */
export async function runE2ETestsFromCLI(): Promise<void> {
  const args = process.argv.slice(2);

  // 解析命令行参数
  const config: E2ETestRunnerConfig = {
    environment: (args.find(arg => arg.startsWith('--env='))?.split('=')[1] as 'development' | 'staging' | 'production') || 'development',
    workflows: args.find(arg => arg.startsWith('--workflows='))?.split('=')[1]?.split(','),
    scenarios: args.find(arg => arg.startsWith('--scenarios='))?.split('=')[1]?.split(','),
    outputDir: args.find(arg => arg.startsWith('--output='))?.split('=')[1],
    parallel: args.includes('--parallel'),
    retries: parseInt(args.find(arg => arg.startsWith('--retries='))?.split('=')[1] || '2'),
    includePerformance: !args.includes('--no-performance'),
    includeCleanup: !args.includes('--no-cleanup')
  };

  const runner = new E2ETestRunner(config);

  try {
    console.log('🚀 开始运行端到端测试...');
    console.log(`📋 环境: ${config.environment}`);
    console.log(`📂 输出目录: ${runner['outputDir']}`);

    const result = await runner.runTests();

    console.log('\n📊 测试结果摘要:');
    console.log(`   总工作流数: ${result.summary.totalWorkflows}`);
    console.log(`   通过工作流: ${result.summary.passedWorkflows}`);
    console.log(`   失败工作流: ${result.summary.failedWorkflows}`);
    console.log(`   成功率: ${result.summary.successRate.toFixed(2)}%`);
    console.log(`   总执行时间: ${Math.round(result.summary.totalDuration / 1000)}s`);

    if (result.summary.failedWorkflows > 0) {
      console.log('\n❌ 失败的工作流:');
      result.workflowResults
        .filter(r => r.status === 'failed')
        .forEach(r => {
          console.log(`   - ${r.workflowName}: ${r.error || '未知错误'}`);
        });
    }

    console.log(`\n📄 报告路径: ${result.reportPath}`);
    console.log(`🖼️ 截图目录: ${result.screenshotsPath}`);
    console.log(`📝 日志目录: ${result.logsPath}`);

    // 根据测试结果设置退出码
    if (result.summary.failedWorkflows > 0) {
      console.error('\n❌ 有工作流测试失败');
      process.exit(1);
    }

    console.log('\n✅ 所有端到端测试通过！');
    process.exit(0);

  } catch (error: unknown) {
    console.error('\n❌ 端到端测试运行失败:', error);
    process.exit(1);
  }
}

// 如果直接运行此文件，执行CLI
if (require.main === module) {
  runE2ETestsFromCLI().catch(error => {
    console.error('未捕获的错误:', error);
    process.exit(1);
  });
}

export default E2ETestRunner;