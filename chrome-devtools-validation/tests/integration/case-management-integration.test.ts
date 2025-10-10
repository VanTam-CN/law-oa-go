/**
 * 案件管理模块集成测试
 */

import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { ChromeDevToolsTestExecutionEngine } from '../../src/core/test-execution-engine';
import { TestExecutionConfig } from '../../src/types/test-engine-types';
import { CaseManagementTestCases } from '../../src/pages/cases/case-management-test-cases';
import { CaseManagementValidator } from '../../src/pages/cases/case-management-validator';
import { Logger } from '../../src/core/logger';

describe('案件管理模块集成测试', () => {
  let engine: ChromeDevToolsTestExecutionEngine;
  let testCases: CaseManagementTestCases;
  let validator: CaseManagementValidator;
  let logger: Logger;
  let config: TestExecutionConfig;

  const baseUrl = process.env['TEST_BASE_URL'] || 'http://localhost:3000';

  beforeAll(async () => {
    logger = new Logger('CaseManagementIntegrationTest');
    testCases = new CaseManagementTestCases();
    validator = new CaseManagementValidator(logger);

    config = {
      baseUrl,
      timeout: 30000,
      retries: 2,
      parallel: false,
      screenshots: true,
      screenshotsDir: './test-screenshots/case-management',
      reportFormat: 'json',
      headless: true,
      slowMo: 100
    };

    engine = new ChromeDevToolsTestExecutionEngine(config, logger);
    await engine.initialize();
  });

  afterAll(async () => {
    if (engine) {
      await engine.cleanup();
    }
  });

  beforeEach(() => {
    logger.info('开始执行案件管理集成测试');
  });

  describe('案件列表页面测试', () => {
    it('应该能够执行案件列表页面的完整测试流程', async () => {
      const testSuite = testCases.getCaseListTestSuite(baseUrl);

      logger.info('执行案件列表页面测试套件', {
        suiteId: testSuite.id,
        testCaseCount: testSuite.testCases?.length || 0
      });

      const result = await engine.executeSuite(testSuite, {
        baseUrl,
        validator: engine as any,
        sharedData: new Map(),
        hooks: engine as any,
        testData: new Map()
      });

      // 验证测试结果
      expect(result).toBeDefined();
      expect(result.status).toBe('completed');
      expect(result.suites).toHaveLength(1);

      const suiteResult = result.suites[0];
      expect(suiteResult.id).toBe(testSuite.id);
      expect(suiteResult.name).toBe(testSuite.name);

      // 验证测试用例执行情况
      if (suiteResult.testCases) {
        expect(suiteResult.testCases.length).toBeGreaterThan(0);

        // 计算通过率
        const passedTests = suiteResult.testCases.filter(tc => tc.result === 'passed').length;
        const passRate = passedTests / suiteResult.testCases.length;

        logger.info('案件列表页面测试结果', {
          total: suiteResult.testCases.length,
          passed: passedTests,
          passRate: Math.round(passRate * 100) + '%'
        });

        // 至少80%的测试用例应该通过
        expect(passRate).toBeGreaterThan(0.8);
      }

      // 验证性能指标
      if (result.performance) {
        expect(result.performance.duration).toBeGreaterThan(0);
        expect(result.performance.duration).toBeLessThan(120000); // 2分钟内完成
      }
    }, 120000);

    it('应该能够验证案件列表页面功能', async () => {
      const validation = await validator.validateCaseManagementModule(baseUrl);

      expect(validation).toBeDefined();
      expect(validation.details.caseListValidation).toBeDefined();

      const caseListValidation = validation.details.caseListValidation;
      expect(caseListValidation.valid).toBe(true);

      // 验证关键功能
      expect(caseListValidation.details.elementValidation.valid).toBe(true);
      expect(caseListValidation.details.searchTest.success).toBe(true);
      expect(caseListValidation.details.filterTest.success).toBe(true);
    }, 60000);
  });

  describe('案件详情页面测试', () => {
    it('应该能够执行案件详情页面的完整测试流程', async () => {
      const testSuite = testCases.getCaseDetailTestSuite(baseUrl);

      logger.info('执行案件详情页面测试套件', {
        suiteId: testSuite.id,
        testCaseCount: testSuite.testCases?.length || 0
      });

      const result = await engine.executeSuite(testSuite, {
        baseUrl,
        validator: engine as any,
        sharedData: new Map(),
        hooks: engine as any,
        testData: new Map()
      });

      // 验证测试结果
      expect(result).toBeDefined();
      expect(result.status).toBe('completed');
      expect(result.suites).toHaveLength(1);

      const suiteResult = result.suites[0];
      expect(suiteResult.id).toBe(testSuite.id);
      expect(suiteResult.name).toBe(testSuite.name);

      // 验证测试用例执行情况
      if (suiteResult.testCases) {
        expect(suiteResult.testCases.length).toBeGreaterThan(0);

        // 计算通过率
        const passedTests = suiteResult.testCases.filter(tc => tc.result === 'passed').length;
        const passRate = passedTests / suiteResult.testCases.length;

        logger.info('案件详情页面测试结果', {
          total: suiteResult.testCases.length,
          passed: passedTests,
          passRate: Math.round(passRate * 100) + '%'
        });

        // 至少80%的测试用例应该通过
        expect(passRate).toBeGreaterThan(0.8);
      }

      // 验证性能指标
      if (result.performance) {
        expect(result.performance.duration).toBeGreaterThan(0);
        expect(result.performance.duration).toBeLessThan(90000); // 1.5分钟内完成
      }
    }, 90000);

    it('应该能够验证案件详情页面功能', async () => {
      const validation = await validator.validateCaseManagementModule(baseUrl);

      expect(validation).toBeDefined();
      expect(validation.details.caseDetailValidation).toBeDefined();

      const caseDetailValidation = validation.details.caseDetailValidation;
      expect(caseDetailValidation.valid).toBe(true);

      // 验证关键功能
      expect(caseDetailValidation.details.elementValidation.valid).toBe(true);
      expect(caseDetailValidation.details.milestonesTest.success).toBe(true);
      expect(caseDetailValidation.details.documentsTest.success).toBe(true);
      expect(caseDetailValidation.details.timelineTest.success).toBe(true);
    }, 60000);
  });

  describe('案件表单页面测试', () => {
    it('应该能够执行案件表单页面的完整测试流程', async () => {
      const testSuite = testCases.getCaseFormTestSuite(baseUrl);

      logger.info('执行案件表单页面测试套件', {
        suiteId: testSuite.id,
        testCaseCount: testSuite.testCases?.length || 0
      });

      const result = await engine.executeSuite(testSuite, {
        baseUrl,
        validator: engine as any,
        sharedData: new Map(),
        hooks: engine as any,
        testData: new Map()
      });

      // 验证测试结果
      expect(result).toBeDefined();
      expect(result.status).toBe('completed');
      expect(result.suites).toHaveLength(1);

      const suiteResult = result.suites[0];
      expect(suiteResult.id).toBe(testSuite.id);
      expect(suiteResult.name).toBe(testSuite.name);

      // 验证测试用例执行情况
      if (suiteResult.testCases) {
        expect(suiteResult.testCases.length).toBeGreaterThan(0);

        // 计算通过率
        const passedTests = suiteResult.testCases.filter(tc => tc.result === 'passed').length;
        const passRate = passedTests / suiteResult.testCases.length;

        logger.info('案件表单页面测试结果', {
          total: suiteResult.testCases.length,
          passed: passedTests,
          passRate: Math.round(passRate * 100) + '%'
        });

        // 至少80%的测试用例应该通过
        expect(passRate).toBeGreaterThan(0.8);
      }

      // 验证性能指标
      if (result.performance) {
        expect(result.performance.duration).toBeGreaterThan(0);
        expect(result.performance.duration).toBeLessThan(120000); // 2分钟内完成
      }
    }, 120000);

    it('应该能够验证案件表单页面功能', async () => {
      const validation = await validator.validateCaseManagementModule(baseUrl);

      expect(validation).toBeDefined();
      expect(validation.details.caseFormValidation).toBeDefined();

      const caseFormValidation = validation.details.caseFormValidation;
      expect(caseFormValidation.valid).toBe(true);

      // 验证关键功能
      expect(caseFormValidation.details.navigationTest.success).toBe(true);
      expect(caseFormValidation.details.formValidationTest.success).toBe(true);
      expect(caseFormValidation.details.formDataTest.success).toBe(true);
      expect(caseFormValidation.details.tagsTest.success).toBe(true);
    }, 60000);
  });

  describe('端到端案件管理流程测试', () => {
    it('应该能够执行完整的案件创建和管理流程', async () => {
      // 构建端到端测试套件
      const e2eTestSuite = {
        id: 'case-management-e2e',
        name: '案件管理端到端测试',
        description: '测试从案件创建到管理的完整流程',
        baseUrl,
        setup: [
          {
            id: 'login',
            name: '用户登录',
            type: 'navigate',
            url: '/login',
            timeout: 10000
          },
          {
            id: 'perform-login',
            name: '执行登录',
            type: 'input',
            selector: '#username',
            value: 'test-lawyer',
            timeout: 5000
          },
          {
            id: 'enter-password',
            name: '输入密码',
            type: 'input',
            selector: '#password',
            value: 'password123',
            timeout: 5000
          },
          {
            id: 'submit-login',
            name: '提交登录',
            type: 'click',
            selector: '#login-button',
            timeout: 10000
          }
        ],
        testCases: [
          {
            id: 'create-case-workflow',
            name: '创建案件工作流',
            description: '测试案件创建的完整工作流',
            priority: 'high',
            timeout: 60000,
            steps: [
              {
                id: 'navigate-to-cases',
                name: '导航到案件列表',
                type: 'navigate',
                url: '/cases',
                timeout: 10000
              },
              {
                id: 'click-create-case',
                name: '点击创建案件',
                type: 'click',
                selector: '#create-case-button',
                timeout: 5000
              },
              {
                id: 'fill-case-form',
                name: '填写案件表单',
                type: 'input',
                selector: '#case-title',
                value: '端到端测试案件',
                timeout: 5000
              },
              {
                id: 'fill-description',
                name: '填写描述',
                type: 'input',
                selector: '#case-description',
                value: '这是一个端到端测试案件',
                timeout: 5000
              },
              {
                id: 'select-type',
                name: '选择案件类型',
                type: 'select',
                selector: '#case-type',
                value: 'litigation',
                timeout: 5000
              },
              {
                id: 'select-priority',
                name: '选择优先级',
                type: 'select',
                selector: '#case-priority',
                value: 'medium',
                timeout: 5000
              },
              {
                id: 'save-case',
                name: '保存案件',
                type: 'click',
                selector: '#save-case-button',
                timeout: 10000
              },
              {
                id: 'verify-creation',
                name: '验证创建成功',
                type: 'verifyAssertion',
                assertion: {
                  type: 'elementContainsText',
                  selector: '.success-message',
                  expectedText: '案件创建成功'
                },
                timeout: 5000
              }
            ],
            assertions: [
              {
                type: 'elementExists',
                selector: '.case-item',
                description: '应该看到新创建的案件'
              }
            ],
            tags: ['e2e', 'create-case']
          },
          {
            id: 'manage-case-workflow',
            name: '管理案件工作流',
            description: '测试案件管理的完整工作流',
            priority: 'high',
            timeout: 90000,
            steps: [
              {
                id: 'navigate-to-cases',
                name: '导航到案件列表',
                type: 'navigate',
                url: '/cases',
                timeout: 10000
              },
              {
                id: 'search-created-case',
                name: '搜索创建的案件',
                type: 'input',
                selector: '#case-search',
                value: '端到端测试案件',
                timeout: 5000
              },
              {
                id: 'wait-search-results',
                name: '等待搜索结果',
                type: 'wait',
                duration: 2000
              },
              {
                id: 'view-case-details',
                name: '查看案件详情',
                type: 'click',
                selector: '.case-row:first-child .view-case-button',
                timeout: 5000
              },
              {
                id: 'verify-case-details',
                name: '验证案件详情',
                type: 'verifyAssertion',
                assertion: {
                  type: 'elementContainsText',
                  selector: '#case-title',
                  expectedText: '端到端测试案件'
                },
                timeout: 5000
              },
              {
                id: 'add-milestone',
                name: '添加里程碑',
                type: 'click',
                selector: '#add-milestone-button',
                timeout: 5000
              },
              {
                id: 'fill-milestone',
                name: '填写里程碑',
                type: 'input',
                selector: '#milestone-title',
                value: '案件受理',
                timeout: 5000
              },
              {
                id: 'save-milestone',
                name: '保存里程碑',
                type: 'click',
                selector: '#save-milestone-button',
                timeout: 5000
              },
              {
                id: 'verify-milestone',
                name: '验证里程碑',
                type: 'verifyAssertion',
                assertion: {
                  type: 'elementContainsText',
                  selector: '.milestone-item .milestone-title',
                  expectedText: '案件受理'
                },
                timeout: 5000
              }
            ],
            assertions: [
              {
                type: 'elementExists',
                selector: '.milestone-item',
                description: '应该看到新添加的里程碑'
              }
            ],
            tags: ['e2e', 'manage-case']
          }
        ],
        cleanup: [
          {
            id: 'cleanup-test-case',
            name: '清理测试案件',
            type: 'executeScript',
            script: '// 清理测试创建的案件数据'
          },
          {
            id: 'logout',
            name: '用户登出',
            type: 'click',
            selector: '#logout-button',
            timeout: 5000
          }
        ]
      };

      logger.info('执行端到端案件管理测试', {
        suiteId: e2eTestSuite.id,
        testCaseCount: e2eTestSuite.testCases.length
      });

      const result = await engine.executeSuite(e2eTestSuite, {
        baseUrl,
        validator: engine as any,
        sharedData: new Map(),
        hooks: engine as any,
        testData: new Map()
      });

      // 验证测试结果
      expect(result).toBeDefined();
      expect(result.status).toBe('completed');
      expect(result.suites).toHaveLength(1);

      const suiteResult = result.suites[0];
      expect(suiteResult.id).toBe(e2eTestSuite.id);
      expect(suiteResult.name).toBe(e2eTestSuite.name);

      // 验证测试用例执行情况
      if (suiteResult.testCases) {
        expect(suiteResult.testCases.length).toBe(2);

        // 计算通过率
        const passedTests = suiteResult.testCases.filter(tc => tc.result === 'passed').length;
        const passRate = passedTests / suiteResult.testCases.length;

        logger.info('端到端测试结果', {
          total: suiteResult.testCases.length,
          passed: passedTests,
          passRate: Math.round(passRate * 100) + '%'
        });

        // 至少75%的测试用例应该通过（端到端测试可能因环境因素失败）
        expect(passRate).toBeGreaterThan(0.75);
      }

      // 验证性能指标
      if (result.performance) {
        expect(result.performance.duration).toBeGreaterThan(0);
        expect(result.performance.duration).toBeLessThan(180000); // 3分钟内完成
      }
    }, 180000);
  });

  describe('验证报告生成', () => {
    it('应该能够生成完整的验证报告', async () => {
      // 先执行一个测试套件
      const testSuite = testCases.getCaseListTestSuite(baseUrl);
      const executionResult = await engine.executeSuite(testSuite, {
        baseUrl,
        validator: engine as any,
        sharedData: new Map(),
        hooks: engine as any,
        testData: new Map()
      });

      // 执行验证
      const validation = await validator.validateCaseManagementModule(baseUrl, executionResult);

      // 生成报告
      const report = validator.generateValidationReport(validation);

      // 验证报告内容
      expect(report).toBeDefined();
      expect(report.length).toBeGreaterThan(0);
      expect(report.includes('# 案件管理模块验证报告')).toBe(true);
      expect(report.includes('## 总体状态')).toBe(true);
      expect(report.includes('## 详细结果')).toBe(true);

      // 验证报告包含必要的信息
      expect(report.includes('错误数量')).toBe(true);
      expect(report.includes('警告数量')).toBe(true);
      expect(report.includes('案件列表页面')).toBe(true);
      expect(report.includes('案件详情页面')).toBe(true);
      expect(report.includes('案件表单页面')).toBe(true);

      logger.info('验证报告生成成功', {
        reportLength: report.length,
        hasErrors: validation.errors.length > 0,
        hasWarnings: validation.warnings.length > 0
      });

      // 如果有错误或警告，应该包含在报告中
      if (validation.errors.length > 0) {
        expect(report.includes('## 错误详情')).toBe(true);
      }

      if (validation.warnings.length > 0) {
        expect(report.includes('## 警告详情')).toBe(true);
      }

      expect(report.includes('## 建议')).toBe(true);
    }, 60000);
  });

  describe('错误处理和恢复', () => {
    it('应该能够正确处理测试执行错误', async () => {
      // 创建一个会失败的测试套件
      const failingTestSuite = {
        id: 'failing-test-suite',
        name: '失败测试套件',
        description: '测试错误处理机制',
        baseUrl,
        testCases: [
          {
            id: 'failing-test',
            name: '失败的测试用例',
            description: '故意失败的测试用例',
            priority: 'medium',
            timeout: 10000,
            steps: [
              {
                id: 'click-non-existent-element',
                name: '点击不存在的元素',
                type: 'click',
                selector: '#non-existent-element',
                timeout: 5000
              }
            ],
            assertions: [],
            tags: ['error-handling']
          }
        ]
      };

      logger.info('执行错误处理测试');

      const result = await engine.executeSuite(failingTestSuite, {
        baseUrl,
        validator: engine as any,
        sharedData: new Map(),
        hooks: engine as any,
        testData: new Map()
      });

      // 验证测试结果
      expect(result).toBeDefined();
      expect(result.status).toBe('completed');
      expect(result.suites).toHaveLength(1);

      const suiteResult = result.suites[0];
      expect(suiteResult.id).toBe(failingTestSuite.id);
      expect(suiteResult.name).toBe(failingTestSuite.name);

      // 验证失败的测试用例
      if (suiteResult.testCases) {
        expect(suiteResult.testCases).toHaveLength(1);

        const failedTestCase = suiteResult.testCases[0];
        expect(failedTestCase.id).toBe('failing-test');
        expect(failedTestCase.result).toBe('failed');
        expect(failedTestCase.error).toBeDefined();

        logger.info('错误处理测试结果', {
          testCaseId: failedTestCase.id,
          result: failedTestCase.result,
          error: failedTestCase.error
        });
      }

      // 验证引擎状态正常
      expect(engine.getStatus()).toBe('ready');
    }, 30000);
  });

  describe('性能和资源管理', () => {
    it('应该能够正确管理测试资源', async () => {
      const startMemory = process.memoryUsage();

      // 执行多个测试套件
      const testSuites = [
        testCases.getCaseListTestSuite(baseUrl),
        testCases.getCaseDetailTestSuite(baseUrl),
        testCases.getCaseFormTestSuite(baseUrl)
      ];

      const results = [];
      for (const suite of testSuites) {
        const result = await engine.executeSuite(suite, {
          baseUrl,
          validator: engine as any,
          sharedData: new Map(),
          hooks: engine as any,
          testData: new Map()
        });
        results.push(result);
      }

      const endMemory = process.memoryUsage();
      const memoryIncrease = endMemory.heapUsed - startMemory.heapUsed;

      logger.info('性能测试结果', {
        testSuitesExecuted: testSuites.length,
        memoryIncrease: `${Math.round(memoryIncrease / 1024 / 1024)}MB`,
        resultsCount: results.length
      });

      // 验证所有测试都成功执行
      expect(results).toHaveLength(testSuites.length);
      results.forEach(result => {
        expect(result.status).toBe('completed');
        expect(result.suites).toHaveLength(1);
      });

      // 验证内存增长在合理范围内（小于100MB）
      expect(memoryIncrease).toBeLessThan(100 * 1024 * 1024);

      // 验证引擎状态正常
      expect(engine.getStatus()).toBe('ready');
    }, 180000);
  });
});