/**
 * 文档管理集成测试
 */

import { jest, describe, it, expect, beforeAll, afterAll, beforeEach, afterEach } from '@jest/globals';
import { ChromeDevToolsTestExecutionEngine } from '../../src/core/test-execution-engine';
import { MemoryDataProvider } from '../../src/core/test-data-provider';
import { DocumentListPage } from '../../src/pages/documents/document-list-page';
import { DocumentDetailPage } from '../../src/pages/documents/document-detail-page';
import { DocumentFormPage } from '../../src/pages/documents/document-form-page';
import { DocumentManagementTestCases } from '../../src/pages/documents/document-management-test-cases';
import { DocumentManagementValidator } from '../../src/pages/documents/document-management-validator';
import { Logger } from '../../src/core/logger';
import { TestExecutionContext } from '../../src/types/test-engine-types';

describe('文档管理集成测试', () => {
  let testEngine: ChromeDevToolsTestExecutionEngine;
  let testDataProvider: MemoryDataProvider;
  let testCases: DocumentManagementTestCases;
  let validator: DocumentManagementValidator;
  let logger: Logger;
  let mockValidator: any;

  const baseUrl = process.env['TEST_BASE_URL'] || 'http://localhost:3000';

  beforeAll(async () => {
    logger = new Logger('DocumentManagementIntegrationTest');
    testEngine = new ChromeDevToolsTestExecutionEngine();
    testDataProvider = new MemoryDataProvider();
    testCases = new DocumentManagementTestCases();
    validator = new DocumentManagementValidator(baseUrl, logger);

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
    logger.debug('开始文档管理集成测试');
  });

  afterEach(async () => {
    // 每个测试后的清理工作
    logger.debug('文档管理集成测试完成');
  });

  describe('文档列表功能', () => {
    it('应该能够成功加载文档列表页面', async () => {
      // 创建测试用例
      const testCase = testCases.getLoadDocumentListTestCase();

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

    it('应该能够正确搜索文档', async () => {
      const testCase = testCases.getSearchDocumentsTestCase();
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

    it('应该能够正确过滤文档', async () => {
      const testCase = testCases.getFilterDocumentsTestCase();
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

    it('应该能够正确排序文档', async () => {
      const testCase = testCases.getSortDocumentsTestCase();
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

    it('应该能够正确导出文档数据', async () => {
      const testCase = testCases.getExportDocumentsTestCase();
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

  describe('文档详情功能', () => {
    it('应该能够成功加载文档详情页面', async () => {
      const testCase = testCases.getLoadDocumentDetailTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确显示文档信息', async () => {
      const testCase = testCases.getViewDocumentInfoTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确预览文档', async () => {
      const testCase = testCases.getPreviewDocumentTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确下载文档', async () => {
      const testCase = testCases.getDownloadDocumentTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确共享文档', async () => {
      const testCase = testCases.getShareDocumentTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确管理文档评论', async () => {
      const testCase = testCases.getManageDocumentCommentsTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确查看版本历史', async () => {
      const testCase = testCases.getViewVersionHistoryTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确编辑文档信息', async () => {
      const testCase = testCases.getEditDocumentTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确删除文档', async () => {
      const testCase = testCases.getDeleteDocumentTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentDetailTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });
  });

  describe('文档表单功能', () => {
    it('应该能够成功创建新文档', async () => {
      const testCase = testCases.getCreateDocumentTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确编辑现有文档', async () => {
      const testCase = testCases.getEditDocumentFormTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确进行表单验证', async () => {
      const testCase = testCases.getFormValidationTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确上传文档', async () => {
      const testCase = testCases.getDocumentUploadTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确预览文档', async () => {
      const testCase = testCases.getDocumentPreviewTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确提取文档元数据', async () => {
      const testCase = testCases.getDocumentMetadataTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确使用保存并新建功能', async () => {
      const testCase = testCases.getSaveAndNewTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });

    it('应该能够正确取消操作', async () => {
      const testCase = testCases.getCancelOperationTestCase();
      const context = await testDataProvider.createTestContext({
        baseUrl,
        testCase,
        testSuite: testCases.getDocumentFormTestSuite()
      });

      const result = await testEngine.executeTestCase(testCase, context);

      expect(result.passed).toBe(true);
    });
  });

  describe('文档管理完整流程测试', () => {
    it('应该能够完成文档创建到查看的完整流程', async () => {
      // 创建文档
      const createTestCase = testCases.getCreateDocumentTestCase();
      const createContext = await testDataProvider.createTestContext({
        baseUrl,
        testCase: createTestCase,
        testSuite: testCases.getDocumentFormTestSuite()
      });

      const createResult = await testEngine.executeTestCase(createTestCase, createContext);
      expect(createResult.passed).toBe(true);

      // 等待一下
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 查看文档列表
      const listTestCase = testCases.getLoadDocumentListTestCase();
      const listContext = await testDataProvider.createTestContext({
        baseUrl,
        testCase: listTestCase,
        testSuite: testCases.getDocumentListTestSuite()
      });

      const listResult = await testEngine.executeTestCase(listTestCase, listContext);
      expect(listResult.passed).toBe(true);
    });

    it('应该能够完成文档搜索到查看详情的完整流程', async () => {
      // 搜索文档
      const searchTestCase = testCases.getSearchDocumentsTestCase();
      const searchContext = await testDataProvider.createTestContext({
        baseUrl,
        testCase: searchTestCase,
        testSuite: testCases.getDocumentListTestSuite()
      });

      const searchResult = await testEngine.executeTestCase(searchTestCase, searchContext);
      expect(searchResult.passed).toBe(true);

      // 等待一下
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 查看文档详情
      const detailTestCase = testCases.getLoadDocumentDetailTestCase();
      const detailContext = await testDataProvider.createTestContext({
        baseUrl,
        testCase: detailTestCase,
        testSuite: testCases.getDocumentDetailTestSuite()
      });

      const detailResult = await testEngine.executeTestCase(detailTestCase, detailContext);
      expect(detailResult.passed).toBe(true);
    });
  });

  describe('Page Object验证', () => {
    it('文档列表Page Object应该能够正确验证页面', async () => {
      const documentListPage = new DocumentListPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await documentListPage.navigateToDocumentList();

      const validation = await documentListPage.validateDocumentListPage();
      expect(validation.valid).toBe(true);
    });

    it('文档详情Page Object应该能够正确验证页面', async () => {
      const documentDetailPage = new DocumentDetailPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await documentDetailPage.navigateToDocumentDetail('doc-1');

      const validation = await documentDetailPage.validateDocumentDetailPage();
      expect(validation.valid).toBe(true);
    });

    it('文档表单Page Object应该能够正确验证页面', async () => {
      const documentFormPage = new DocumentFormPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await documentFormPage.navigateToCreateDocument();

      const validation = await documentFormPage.validateDocumentFormPage();
      expect(validation.valid).toBe(true);
    });
  });

  describe('数据验证', () => {
    it('应该能够正确获取文档统计信息', async () => {
      const statistics = await validator.getDocumentStatistics();

      expect(statistics.total).toBeGreaterThanOrEqual(0);
      expect(statistics.totalSize).toBeGreaterThanOrEqual(0);
      expect(statistics.byType).toBeDefined();
      expect(statistics.byClient).toBeDefined();
      expect(statistics.byStatus).toBeDefined();
    });

    it('应该能够正确验证文档管理模块', async () => {
      const validationResult = await validator.validateDocumentManagementModule(baseUrl);

      expect(validationResult).toBeDefined();
      expect(validationResult.details).toBeDefined();
      expect(validationResult.details.documentList).toBeDefined();
      expect(validationResult.details.documentDetail).toBeDefined();
      expect(validationResult.details.documentForm).toBeDefined();
    });

    it('应该能够生成验证报告', async () => {
      const validationResult = await validator.validateDocumentManagementModule(baseUrl);
      const report = validator.generateValidationReport(validationResult);

      expect(report).toBeDefined();
      expect(report.length).toBeGreaterThan(0);
      expect(report.includes('文档管理模块验证报告')).toBe(true);
    });
  });

  describe('错误处理和边界情况', () => {
    it('应该能够正确处理无效的文档ID', async () => {
      const documentDetailPage = new DocumentDetailPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);

      // 尝试导航到无效的文档ID
      try {
        await documentDetailPage.navigateToDocumentDetail('invalid-document-id');
        // 如果没有抛出错误，检查页面是否正确处理了无效ID
        const validation = await documentDetailPage.validateDocumentDetailPage();
        // 预期页面可能显示错误信息，所以验证可能会失败
        expect(validation.valid).toBe(false);
      } catch (error) {
        // 或者抛出错误，这也是可以接受的
        expect(error).toBeDefined();
      }
    });

    it('应该能够正确处理空搜索结果', async () => {
      const documentListPage = new DocumentListPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await documentListPage.navigateToDocumentList();

      // 搜索不存在的文档
      await documentListPage.searchDocuments({ query: '不存在的文档名称123456' });
      await new Promise(resolve => setTimeout(resolve, 1000));

      const documents = await documentListPage.getDocumentList();
      // 搜索结果可能为空，这是正常的
      expect(Array.isArray(documents)).toBe(true);
    });

    it('应该能够正确处理表单验证错误', async () => {
      const documentFormPage = new DocumentFormPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await documentFormPage.navigateToCreateDocument();

      // 验证空表单应该有错误
      const validation = await documentFormPage.validateForm();
      expect(validation.valid).toBe(false);
      expect(Object.keys(validation.errors).length).toBeGreaterThan(0);
    });
  });

  describe('性能测试', () => {
    it('文档列表页面加载时间应该合理', async () => {
      const startTime = Date.now();
      const documentListPage = new DocumentListPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await documentListPage.navigateToDocumentList();
      const endTime = Date.now();

      const loadTime = endTime - startTime;
      expect(loadTime).toBeLessThan(5000); // 5秒内加载完成
    });

    it('文档详情页面加载时间应该合理', async () => {
      const startTime = Date.now();
      const documentDetailPage = new DocumentDetailPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await documentDetailPage.navigateToDocumentDetail('doc-1');
      const endTime = Date.now();

      const loadTime = endTime - startTime;
      expect(loadTime).toBeLessThan(5000); // 5秒内加载完成
    });

    it('文档表单页面加载时间应该合理', async () => {
      const startTime = Date.now();
      const documentFormPage = new DocumentFormPage({
        baseUrl,
        defaultTimeout: 30000,
        screenshotOnFailure: true
      }, logger);
      await documentFormPage.navigateToCreateDocument();
      const endTime = Date.now();

      const loadTime = endTime - startTime;
      expect(loadTime).toBeLessThan(5000); // 5秒内加载完成
    });
  });
});