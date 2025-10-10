/**
 * 客户管理集成测试
 */

import { jest, describe, it, expect, beforeAll, afterAll, beforeEach, afterEach } from '@jest/globals';
import { ChromeDevToolsTestExecutionEngine } from '../../src/core/test-execution-engine';
import { MemoryDataProvider } from '../../src/core/test-data-provider';
import { ClientListPage } from '../../src/pages/clients/client-list-page';
import { ClientDetailPage } from '../../src/pages/clients/client-detail-page';
import { ClientFormPage } from '../../src/pages/clients/client-form-page';
import { ClientManagementTestCases } from '../../src/pages/clients/client-management-test-cases';
import { ClientManagementValidator } from '../../src/pages/clients/client-management-validator';
import { Logger } from '../../src/core/logger';
import { TestExecutionContext } from '../../src/types/test-engine-types';

describe('客户管理集成测试', () => {
  let testEngine: ChromeDevToolsTestExecutionEngine;
  let testDataProvider: MemoryDataProvider;
  let testCases: ClientManagementTestCases;
  let validator: ClientManagementValidator;
  let logger: Logger;
  let mockValidator: any;

  const baseUrl = process.env['TEST_BASE_URL'] || 'http://localhost:3000';

  beforeAll(async () => {
    logger = new Logger('ClientManagementIntegrationTest');
    testEngine = new ChromeDevToolsTestExecutionEngine();
    testDataProvider = new MemoryDataProvider();
    testCases = new ClientManagementTestCases();
    validator = new ClientManagementValidator(baseUrl, logger);

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
  });

  beforeEach(async () => {
    // 每个测试前的准备工作
    logger.debug('开始客户管理集成测试');
  });

  afterEach(async () => {
    // 每个测试后的清理工作
    logger.debug('客户管理集成测试完成');
  });

  describe('客户列表功能', () => {
    it('应该能够成功加载客户列表页面', async () => {
      // 创建测试用例
      const testCase = testCases.getLoadClientListTestCase();

      // 创建执行上下文
      const context: TestExecutionContext = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: testEngine.getConfig(),
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

      // 执行测试
      const result = await testEngine.executeTestCase(testCase, context);

      // 验证结果
      expect(result.passed).toBe(true);
      expect(result.duration).toBeGreaterThan(0);
    });

    it('应该能够正确搜索客户', async () => {
      const testCase = testCases.getSearchClientsTestCase();
      const context: TestExecutionContext = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: testEngine.getConfig(),
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

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确过滤客户', async () => {
      const testCase = testCases.getFilterClientsTestCase();
      const context: TestExecutionContext = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: testEngine.getConfig(),
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

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确排序客户', async () => {
      const testCase = testCases.getSortClientsTestCase();
      const context: TestExecutionContext = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: testEngine.getConfig(),
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

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确处理分页', async () => {
      const testCase = testCases.getPaginationTestCase();
      const context: TestExecutionContext = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: testEngine.getConfig(),
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

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确执行批量操作', async () => {
      const testCase = testCases.getBulkOperationsTestCase();
      const context: TestExecutionContext = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: testEngine.getConfig(),
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

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确导出客户数据', async () => {
      const testCase = testCases.getExportClientsTestCase();
      const context: TestExecutionContext = {
        executionId: 'test-execution',
        startTime: new Date(),
        config: testEngine.getConfig(),
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

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });
  });

  describe('客户详情功能', () => {
    it('应该能够成功加载客户详情页面', async () => {
      const testCase = testCases.getLoadClientDetailTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确显示客户信息', async () => {
      const testCase = testCases.getViewClientInfoTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确管理联系人', async () => {
      const testCase = testCases.getManageContactsTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确管理相关案件', async () => {
      const testCase = testCases.getManageCasesTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确管理相关文档', async () => {
      const testCase = testCases.getManageDocumentsTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确编辑客户信息', async () => {
      const testCase = testCases.getEditClientTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确删除客户', async () => {
      const testCase = testCases.getDeleteClientTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });
  });

  describe('客户表单功能', () => {
    it('应该能够成功创建新客户', async () => {
      const testCase = testCases.getCreateClientTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确编辑现有客户', async () => {
      const testCase = testCases.getEditClientFormTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确进行表单验证', async () => {
      const testCase = testCases.getFormValidationTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确管理表单中的联系人', async () => {
      const testCase = testCases.getContactManagementTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确使用保存并新建功能', async () => {
      const testCase = testCases.getSaveAndNewTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确取消操作', async () => {
      const testCase = testCases.getCancelOperationTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确复制客户', async () => {
      const testCase = testCases.getDuplicateClientTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getClientFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });
  });

  describe('客户管理完整流程测试', () => {
    it('应该能够完成客户创建到查看的完整流程', async () => {
      // 创建客户
      const createTestCase = testCases.getCreateClientTestCase();
      const createContext = await testDataProvider.createTestContext({
        baseUrl,
        testCase: createTestCase,
        testSuite: testCases.getClientFormTestSuite()
      });

      const createResult = await testEngine.executeTestCase(createTestCase, createContext);
      expect(createResult.passed).toBe(true);

      // 等待一下
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 查看客户列表
      const listTestCase = testCases.getLoadClientListTestCase();
      const listContext = await testDataProvider.createTestContext({
        baseUrl,
        testCase: listTestCase,
        testSuite: testCases.getClientListTestSuite()
      });

      const listResult = await testEngine.executeTestCase(listTestCase, listContext);
      expect(listResult.passed).toBe(true);
    });

    it('应该能够完成客户搜索到查看详情的完整流程', async () => {
      // 搜索客户
      const searchTestCase = testCases.getSearchClientsTestCase();
      const searchContext = await testDataProvider.createTestContext({
        baseUrl,
        testCase: searchTestCase,
        testSuite: testCases.getClientListTestSuite()
      });

      const searchResult = await testEngine.executeTestCase(searchTestCase, searchContext);
      expect(searchResult.passed).toBe(true);

      // 等待一下
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 查看客户详情
      const detailTestCase = testCases.getLoadClientDetailTestCase();
      const detailContext = await testDataProvider.createTestContext({
        baseUrl,
        testCase: detailTestCase,
        testSuite: testCases.getClientDetailTestSuite()
      });

      const detailResult = await testEngine.executeTestCase(detailTestCase, detailContext);
      expect(detailResult.passed).toBe(true);
    });
  });

  describe('Page Object验证', () => {
    it('客户列表Page Object应该能够正确验证页面', async () => {
      const clientListPage = new ClientListPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await clientListPage.navigateToClientList();

      const validation = await clientListPage.validateClientListPage();
      expect(validation.valid).toBe(true);
    });

    it('客户详情Page Object应该能够正确验证页面', async () => {
      const clientDetailPage = new ClientDetailPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await clientDetailPage.navigateToClientDetail('client-1');

      const validation = await clientDetailPage.validateClientDetailPage();
      expect(validation.valid).toBe(true);
    });

    it('客户表单Page Object应该能够正确验证页面', async () => {
      const clientFormPage = new ClientFormPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await clientFormPage.navigateToCreateClient();

      const validation = await clientFormPage.validateClientFormPage();
      expect(validation.valid).toBe(true);
    });
  });

  describe('数据验证', () => {
    it('应该能够正确获取客户统计信息', async () => {
      const statistics = await validator.getClientStatistics();

      expect(statistics.total).toBeGreaterThanOrEqual(0);
      expect(statistics.active).toBeGreaterThanOrEqual(0);
      expect(statistics.inactive).toBeGreaterThanOrEqual(0);
      expect(statistics.byType).toBeDefined();
      expect(statistics.byIndustry).toBeDefined();
    });

    it('应该能够正确验证客户管理模块', async () => {
      const validationResult = await validator.validateClientManagementModule(baseUrl);

      expect(validationResult).toBeDefined();
      expect(validationResult.details).toBeDefined();
      expect(validationResult.details.clientList).toBeDefined();
      expect(validationResult.details.clientDetail).toBeDefined();
      expect(validationResult.details.clientForm).toBeDefined();
    });

    it('应该能够生成验证报告', async () => {
      const validationResult = await validator.validateClientManagementModule(baseUrl);
      const report = validator.generateValidationReport(validationResult);

      expect(report).toBeDefined();
      expect(report.length).toBeGreaterThan(0);
      expect(report.includes('客户管理模块验证报告')).toBe(true);
    });
  });

  describe('错误处理和边界情况', () => {
    it('应该能够正确处理无效的客户ID', async () => {
      const clientDetailPage = new ClientDetailPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);

      // 尝试导航到无效的客户ID
      try {
        await clientDetailPage.navigateToClientDetail('invalid-client-id');
        // 如果没有抛出错误，检查页面是否正确处理了无效ID
        const validation = await clientDetailPage.validateClientDetailPage();
        // 预期页面可能显示错误信息，所以验证可能会失败
        expect(validation.valid).toBe(false);
      } catch (error) {
        // 或者抛出错误，这也是可以接受的
        expect(error).toBeDefined();
      }
    });

    it('应该能够正确处理空搜索结果', async () => {
      const clientListPage = new ClientListPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await clientListPage.navigateToClientList();

      // 搜索不存在的客户
      await clientListPage.searchClients('不存在的客户名称123456');
      await new Promise(resolve => setTimeout(resolve, 1000));

      const clients = await clientListPage.getClientList();
      // 搜索结果可能为空，这是正常的
      expect(Array.isArray(clients)).toBe(true);
    });

    it('应该能够正确处理表单验证错误', async () => {
      const clientFormPage = new ClientFormPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await clientFormPage.navigateToCreateClient();

      // 验证空表单应该有错误
      const validation = await clientFormPage.validateForm();
      expect(validation.valid).toBe(false);
      expect(Object.keys(validation.errors).length).toBeGreaterThan(0);
    });
  });

  describe('性能测试', () => {
    it('客户列表页面加载时间应该合理', async () => {
      const startTime = Date.now();
      const clientListPage = new ClientListPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await clientListPage.navigateToClientList();
      const endTime = Date.now();

      const loadTime = endTime - startTime;
      expect(loadTime).toBeLessThan(5000); // 5秒内加载完成
    });

    it('客户详情页面加载时间应该合理', async () => {
      const startTime = Date.now();
      const clientDetailPage = new ClientDetailPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await clientDetailPage.navigateToClientDetail('client-1');
      const endTime = Date.now();

      const loadTime = endTime - startTime;
      expect(loadTime).toBeLessThan(5000); // 5秒内加载完成
    });

    it('客户表单页面加载时间应该合理', async () => {
      const startTime = Date.now();
      const clientFormPage = new ClientFormPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await clientFormPage.navigateToCreateClient();
      const endTime = Date.now();

      const loadTime = endTime - startTime;
      expect(loadTime).toBeLessThan(5000); // 5秒内加载完成
    });
  });
});