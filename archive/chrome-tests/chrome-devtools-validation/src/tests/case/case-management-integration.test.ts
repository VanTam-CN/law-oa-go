/**
 * 案件管理集成测试
 */

import { CaseTestSuite } from '../case/case-test-suite';
import { CASE_TEST_CONFIG } from './case-test-config';
import { CaseValidationUtils } from './case-validation-utils';
import { Logger } from '../../core/logger';

describe('案件管理集成测试', () => {
  let testSuite: CaseTestSuite;
  let validationUtils: CaseValidationUtils;
  let logger: Logger;

  beforeAll(() => {
    logger = new Logger('CaseIntegrationTest');
    testSuite = new CaseTestSuite({
      baseUrl: CASE_TEST_CONFIG.baseUrl,
      defaultTimeout: CASE_TEST_CONFIG.defaultTimeout,
      screenshotOnFailure: true
    }, logger);
    validationUtils = new CaseValidationUtils(logger);
  });

  describe('案件数据验证', () => {
    test('应该验证测试数据的完整性', () => {
      const result = validationUtils.validateTestDataIntegrity();
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test('应该验证有效的案件数据', () => {
      const validCase = CASE_TEST_CONFIG.testCases[0];
      const result = validationUtils.validateCaseData(validCase);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test('应该检测无效的案件数据', () => {
      const invalidCase = {
        ...CASE_TEST_CONFIG.testCases[0],
        title: '', // 空标题
        caseNumber: 'INVALID-FORMAT', // 错误格式
        estimatedValue: -1000 // 负值
      };
      const result = validationUtils.validateCaseData(invalidCase);
      expect(result.isValid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    test('应该验证文档数据', () => {
      const validDocument = CASE_TEST_CONFIG.testDocuments[0];
      const result = validationUtils.validateDocumentData(validDocument);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test('应该验证案件与文档的关联性', () => {
      const testCase = CASE_TEST_CONFIG.testCases[0];
      const testDocuments = CASE_TEST_CONFIG.testDocuments.filter(doc => doc.caseId === testCase.id);
      const result = validationUtils.validateCaseDocumentAssociation(testCase, testDocuments);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });
  });

  describe('搜索功能验证', () => {
    test('应该验证搜索查询', () => {
      const query = '合同';
      const results = CASE_TEST_CONFIG.testCases.filter(c => c.title.includes(query));
      const result = validationUtils.validateSearchFunctionality(query, results, CASE_TEST_CONFIG.testCases);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test('应该检测无效的搜索查询', () => {
      const query = '';
      const results = CASE_TEST_CONFIG.testCases;
      const result = validationUtils.validateSearchFunctionality(query, results, CASE_TEST_CONFIG.testCases);
      expect(result.isValid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });
  });

  describe('测试用例生成', () => {
    test('应该生成随机案件编号', () => {
      const caseNumber1 = validationUtils.generateRandomCaseNumber('CL');
      const caseNumber2 = validationUtils.generateRandomCaseNumber('LD');

      expect(caseNumber1).toMatch(/^CL-\d{4}-\d{3}$/);
      expect(caseNumber2).toMatch(/^LD-\d{4}-\d{3}$/);
      expect(caseNumber1).not.toBe(caseNumber2);
    });

    test('应该生成测试案件数据', () => {
      const testCase = validationUtils.generateTestCase({
        title: '自定义测试案件',
        priority: 'high'
      });

      expect(testCase.title).toBe('自定义测试案件');
      expect(testCase.priority).toBe('high');
      expect(testCase.id).toBeDefined();
      expect(testCase.caseNumber).toBeDefined();
    });

    test('应该生成测试文档数据', () => {
      const testDocument = validationUtils.generateDocument('test-case-001', {
        name: '测试文档',
        documentType: 'contract'
      });

      expect(testDocument.name).toBe('测试文档');
      expect(testDocument.documentType).toBe('contract');
      expect(testDocument.caseId).toBe('test-case-001');
      expect(testDocument.id).toBeDefined();
    });
  });

  describe('批量验证', () => {
    test('应该批量验证案件数据', async () => {
      const testCases = CASE_TEST_CONFIG.testCases.slice(0, 2);
      const results = await validationUtils.batchValidateCases(testCases);

      expect(results).toHaveLength(2);
      expect(results[0].validation.isValid).toBe(true);
      expect(results[1].validation.isValid).toBe(true);
      expect(results[0].suggestions).toBeDefined();
      expect(results[1].suggestions).toBeDefined();
    });

    test('应该批量验证文档数据', async () => {
      const testDocuments = CASE_TEST_CONFIG.testDocuments.slice(0, 2);
      const results = await validationUtils.batchValidateDocuments(testDocuments);

      expect(results).toHaveLength(2);
      expect(results[0].validation.isValid).toBe(true);
      expect(results[1].validation.isValid).toBe(true);
      expect(results[0].suggestions).toBeDefined();
      expect(results[1].suggestions).toBeDefined();
    });
  });

  describe('验证报告生成', () => {
    test('应该生成验证报告', () => {
      const caseValidations = [
        {
          caseData: CASE_TEST_CONFIG.testCases[0],
          validation: validationUtils.validateCaseData(CASE_TEST_CONFIG.testCases[0]),
          suggestions: validationUtils.generateSuggestions(CASE_TEST_CONFIG.testCases[0])
        }
      ];

      const documentValidations = [
        {
          document: CASE_TEST_CONFIG.testDocuments[0],
          validation: validationUtils.validateDocumentData(CASE_TEST_CONFIG.testDocuments[0]),
          suggestions: []
        }
      ];

      const report = validationUtils.generateValidationReport(caseValidations, documentValidations);

      expect(report).toBeDefined();
      expect(typeof report).toBe('string');

      const reportData = JSON.parse(report);
      expect(reportData.summary).toBeDefined();
      expect(reportData.caseValidations).toBeDefined();
      expect(reportData.documentValidations).toBeDefined();
      expect(reportData.recommendations).toBeDefined();
    });
  });

  describe('建议生成', () => {
    test('应该为不完整的案件数据生成建议', () => {
      const incompleteCase = {
        ...CASE_TEST_CONFIG.testCases[0],
        clientEmail: '',
        description: '简短描述',
        tags: [],
        milestones: [],
        budget: 0
      };

      const suggestions = validationUtils.generateSuggestions(incompleteCase);

      expect(suggestions.length).toBeGreaterThan(0);
      expect(suggestions.some(s => s.includes('联系方式'))).toBe(true);
      expect(suggestions.some(s => s.includes('补充更详细的'))).toBe(true);
    });

    test('应该为高优先级未分配案件生成建议', () => {
      const unassignedCase = {
        ...CASE_TEST_CONFIG.testCases[0],
        priority: 'high',
        assignedAttorney: '',
        teamMembers: []
      };

      const suggestions = validationUtils.generateSuggestions(unassignedCase);

      expect(suggestions.length).toBeGreaterThan(0);
      expect(suggestions.some(s => s.includes('尽快分配律师'))).toBe(true);
      expect(suggestions.some(s => s.includes('分配团队成员'))).toBe(true);
    });
  });

  describe('测试数据一致性', () => {
    test('应该验证测试数据的一致性', () => {
      // 验证所有测试案件都有唯一的ID
      const caseIds = CASE_TEST_CONFIG.testCases.map(c => c.id);
      const uniqueCaseIds = new Set(caseIds);
      expect(caseIds.length).toBe(uniqueCaseIds.size);

      // 验证所有测试文档都有唯一的ID
      const documentIds = CASE_TEST_CONFIG.testDocuments.map(d => d.id);
      const uniqueDocumentIds = new Set(documentIds);
      expect(documentIds.length).toBe(uniqueDocumentIds.size);

      // 验证文档与案件的关联性
      const caseIdsInDocs = CASE_TEST_CONFIG.testDocuments.map(d => d.caseId);
      const allCaseIds = CASE_TEST_CONFIG.testCases.map(c => c.id);

      caseIdsInDocs.forEach(docCaseId => {
        expect(allCaseIds).toContain(docCaseId);
      });
    });

    test('应该验证枚举值的合法性', () => {
      const validPriorities = ['low', 'medium', 'high', 'urgent'];
      const validStatuses = ['draft', 'pending', 'active', 'paused', 'completed', 'closed', 'rejected'];
      const validClientTypes = ['individual', 'corporate'];
      const validCaseTypes = Object.keys(CASE_TEST_CONFIG.caseTypes || {});

      CASE_TEST_CONFIG.testCases.forEach(testCase => {
        expect(validPriorities).toContain(testCase.priority);
        expect(validStatuses).toContain(testCase.status);
        expect(validClientTypes).toContain(testCase.clientType);
        expect(validCaseTypes).toContain(testCase.caseType);
      });
    });
  });
});